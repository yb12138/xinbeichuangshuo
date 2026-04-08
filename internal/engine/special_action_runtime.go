package engine

import (
	"errors"
	"fmt"
	"sort"

	"starcup-engine/internal/model"
)

type specialActionOverridePolicy func(e *GameEngine, player *model.Player, actionType model.ActionType) (bool, error)
type specialActionPostHook func(e *GameEngine, player *model.Player, actionType model.ActionType)

func (e *GameEngine) executeSpecialActionWithRuntime(player *model.Player, actionType model.ActionType) error {
	handled, err := e.applyTimingBeforeActionExecuteSpecialActionOverride(player, actionType)
	if err != nil {
		return err
	}
	if handled {
		return nil
	}
	return e.executeSpecialAction(player, actionType)
}

func (e *GameEngine) executeSpecialAction(p *model.Player, actType model.ActionType) error {
	switch actType {
	case model.ActionBuy:
		return e.handleBuy(p)
	case model.ActionSynthesize:
		return e.handleSynthesize(p)
	case model.ActionExtract:
		return e.handleExtract(p)
	default:
		return fmt.Errorf("未知的特殊行动类型: %s", actType)
	}
}

func (e *GameEngine) handleBuy(p *model.Player) error {
	maxHand := e.GetMaxHand(p)
	if len(p.Hand)+3 > maxHand {
		return fmt.Errorf("购买后手牌将超过上限(%d+3>%d)，无法购买", len(p.Hand), maxHand)
	}

	e.drawForAction(p, 3)
	const maxStones = 5
	var stones int
	if p.Camp == model.RedCamp {
		stones = e.State.RedGems + e.State.RedCrystals
	} else {
		stones = e.State.BlueGems + e.State.BlueCrystals
	}
	if stones >= maxStones {
		e.Log(fmt.Sprintf("[Action] %s 购买：摸3牌，战绩区已满不加星石", p.Name))
		return nil
	}
	if stones == 4 {
		e.PushInterrupt(&model.Interrupt{
			Type:     model.InterruptChoice,
			PlayerID: p.ID,
			Context: map[string]interface{}{
				"choice_type": "buy_resource",
				"camp":        string(p.Camp),
			},
		})
		e.Log(fmt.Sprintf("[Action] %s 购买：摸3牌，战绩区4星石，请选择添加宝石或水晶", p.Name))
		return nil
	}
	if p.Camp == model.RedCamp {
		if e.State.RedGems+e.State.RedCrystals < maxStones {
			e.State.RedGems++
		}
		if e.State.RedGems+e.State.RedCrystals < maxStones {
			e.State.RedCrystals++
		}
	} else {
		if e.State.BlueGems+e.State.BlueCrystals < maxStones {
			e.State.BlueGems++
		}
		if e.State.BlueGems+e.State.BlueCrystals < maxStones {
			e.State.BlueCrystals++
		}
	}
	e.Log(fmt.Sprintf("[Action] %s 购买：摸3牌，战绩区+1宝石+1水晶", p.Name))
	return nil
}

func (e *GameEngine) handleSynthesize(p *model.Player) error {
	maxHand := e.GetMaxHand(p)
	if len(p.Hand)+3 > maxHand {
		return fmt.Errorf("合成后手牌将超过上限(%d+3>%d)，无法合成", len(p.Hand), maxHand)
	}

	var totalStones int
	if p.Camp == model.RedCamp {
		totalStones = e.State.RedGems + e.State.RedCrystals
	} else {
		totalStones = e.State.BlueGems + e.State.BlueCrystals
	}
	if totalStones < 3 {
		return errors.New("战绩区星石不足3个，无法合成")
	}
	e.drawForAction(p, 3)
	cost := 3
	if p.Camp == model.RedCamp {
		if e.State.RedGems >= cost {
			e.State.RedGems -= cost
		} else {
			cost -= e.State.RedGems
			e.State.RedGems = 0
			e.State.RedCrystals -= cost
		}
	} else {
		if e.State.BlueGems >= cost {
			e.State.BlueGems -= cost
		} else {
			cost -= e.State.BlueGems
			e.State.BlueGems = 0
			e.State.BlueCrystals -= cost
		}
	}
	if p.Camp == model.RedCamp {
		e.addCampCup(model.RedCamp)
		e.Log(fmt.Sprintf("[Action] %s 合成星杯！红方星杯+1，蓝方士气-1", p.Name))
		e.State.BlueMorale--
	} else {
		e.addCampCup(model.BlueCamp)
		e.Log(fmt.Sprintf("[Action] %s 合成星杯！蓝方星杯+1，红方士气-1", p.Name))
		e.State.RedMorale--
	}
	e.checkGameEnd()
	return nil
}

func (e *GameEngine) handleExtract(p *model.Player) error {
	e.clearAdventurerExtractState(p)

	currentEnergy := p.Gem + p.Crystal
	maxEnergy := e.getPlayerEnergyCap(p)

	var availableGems, availableCrystals int
	if p.Camp == model.RedCamp {
		availableGems = e.State.RedGems
		availableCrystals = e.State.RedCrystals
	} else {
		availableGems = e.State.BlueGems
		availableCrystals = e.State.BlueCrystals
	}

	totalAvailable := availableGems + availableCrystals
	if totalAvailable == 0 {
		return errors.New("阵营资源池中没有可提取的资源")
	}

	energyRoom := maxEnergy - currentEnergy
	allowParadise := e.playerHasSkill(p, "adventurer_paradise")
	maxAllyRoom := 0
	if allowParadise {
		maxAllyRoom = e.maxAllyEnergyRoom(p)
	}
	maxRecipientRoom := energyRoom
	if maxAllyRoom > maxRecipientRoom {
		maxRecipientRoom = maxAllyRoom
	}
	if maxRecipientRoom <= 0 {
		return errors.New("能量已达上限，且没有可承接提炼能量的队友")
	}

	var opts []interface{}
	for i := 0; i < availableGems; i++ {
		opts = append(opts, map[string]interface{}{"type": "gem"})
	}
	for i := 0; i < availableCrystals; i++ {
		opts = append(opts, map[string]interface{}{"type": "crystal"})
	}

	maxSelect := 2
	if maxRecipientRoom < maxSelect {
		maxSelect = maxRecipientRoom
	}
	if totalAvailable < maxSelect {
		maxSelect = totalAvailable
	}

	e.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: p.ID,
		Context: map[string]interface{}{
			"choice_type":            "extract",
			"extract_options":        opts,
			"extract_min":            1,
			"extract_max":            maxSelect,
			"extract_self_room":      energyRoom,
			"extract_max_ally_room":  maxAllyRoom,
			"extract_allow_paradise": allowParadise,
		},
	})
	if allowParadise && maxAllyRoom > 0 {
		e.Log(fmt.Sprintf("[Action] %s 提炼：战绩区有 %d 红宝石 %d 蓝水晶，请选择 1-%d 个提炼（可通过冒险者天堂转移给队友）", p.Name, availableGems, availableCrystals, maxSelect))
	} else {
		e.Log(fmt.Sprintf("[Action] %s 提炼：战绩区有 %d 红宝石 %d 蓝水晶，请选择 1-%d 个提炼", p.Name, availableGems, availableCrystals, maxSelect))
	}
	return nil
}

func (e *GameEngine) playerHasSkill(p *model.Player, skillID string) bool {
	if p == nil || p.Character == nil {
		return false
	}
	for _, s := range p.Character.Skills {
		if s.ID == skillID {
			return true
		}
	}
	return false
}

func (e *GameEngine) getPlayerEnergyCap(player *model.Player) int {
	if player == nil {
		return 3
	}
	cap := 3
	if e.isSage(player) {
		cap++
	}
	return cap
}

func (e *GameEngine) maxAllyEnergyRoom(p *model.Player) int {
	if p == nil {
		return 0
	}
	maxRoom := 0
	for _, ally := range e.State.Players {
		if ally == nil || ally.Camp != p.Camp || ally.ID == p.ID {
			continue
		}
		maxEnergy := e.getPlayerEnergyCap(ally)
		room := maxEnergy - (ally.Gem + ally.Crystal)
		if room > maxRoom {
			maxRoom = room
		}
	}
	return maxRoom
}

func (e *GameEngine) adventurerParadiseEligibleAllies(user *model.Player, transferTotal int) []string {
	if user == nil || transferTotal <= 0 {
		return nil
	}
	var allyIDs []string
	for _, ally := range e.State.Players {
		if ally == nil || ally.Camp != user.Camp || ally.ID == user.ID {
			continue
		}
		room := e.getPlayerEnergyCap(ally) - (ally.Gem + ally.Crystal)
		if room >= transferTotal {
			allyIDs = append(allyIDs, ally.ID)
		}
	}
	sort.Strings(allyIDs)
	return allyIDs
}

func (e *GameEngine) clearAdventurerExtractState(p *model.Player) {
	if p == nil {
		return
	}
	if p.TurnState.SkillFlowState == nil {
		p.TurnState.SkillFlowState = make(map[string]int)
	}
	p.TurnState.SkillFlowState["adventurer_extract_last_gem"] = 0
	p.TurnState.SkillFlowState["adventurer_extract_last_crystal"] = 0
	p.TurnState.SkillFlowState["adventurer_extract_requires_paradise"] = 0
}

func (e *GameEngine) recordAdventurerExtractResult(p *model.Player, gem, crystal int, requiresParadise bool) {
	if p == nil {
		return
	}
	if p.TurnState.SkillFlowState == nil {
		p.TurnState.SkillFlowState = make(map[string]int)
	}
	p.TurnState.SkillFlowState["adventurer_extract_last_gem"] = gem
	p.TurnState.SkillFlowState["adventurer_extract_last_crystal"] = crystal
	if requiresParadise {
		p.TurnState.SkillFlowState["adventurer_extract_requires_paradise"] = 1
	} else {
		p.TurnState.SkillFlowState["adventurer_extract_requires_paradise"] = 0
	}
}

func (e *GameEngine) runPostSpecialActionRuntime(player *model.Player, actionType model.ActionType) {
	e.runTimingOnActionEndSpecialActionPost(player, actionType)
}

// applyTimingBeforeActionExecuteSpecialActionOverride 在执行特殊行动前应用覆盖策略。
func (e *GameEngine) applyTimingBeforeActionExecuteSpecialActionOverride(player *model.Player, actionType model.ActionType) (bool, error) {
	for _, policy := range e.specialActionOverridePolicies {
		handled, err := policy(e, player, actionType)
		if err != nil {
			return false, err
		}
		if handled {
			return true, nil
		}
	}
	return false, nil
}

// runTimingOnActionEndSpecialActionPost 在特殊行动完成后执行后置规则。
func (e *GameEngine) runTimingOnActionEndSpecialActionPost(player *model.Player, actionType model.ActionType) {
	for _, hook := range e.specialActionPostHooks {
		hook(e, player, actionType)
	}
}

func specialActionAdventurerUndergroundLawOverride(e *GameEngine, player *model.Player, actionType model.ActionType) (bool, error) {
	if e == nil || player == nil || actionType != model.ActionBuy || !e.playerHasSkill(player, "adventurer_underground_law") {
		return false, nil
	}
	e.resolveAdventurerUndergroundLaw(player)
	return true, nil
}

func specialActionHolyBowHolyGloryExitHook(e *GameEngine, player *model.Player, _ model.ActionType) {
	if e == nil || player == nil || !e.isHolyBow(player) || !hasHolyBowHolyGloryForm(player) {
		return
	}
	beforePoses := e.snapshotPlayerPoses()
	leaveHolyBowHolyGloryForm(player)
	e.Heal(player.ID, 1)
	e.Log(fmt.Sprintf("%s 在圣煌形态下执行特殊行动，脱离圣煌形态并获得1点治疗", player.Name))
	e.dispatchOrientationChanges(beforePoses)
}
