// gameflow: 特殊行动（如提炼、合成）与阶段互动。

package engine

import (
	"errors"
	"fmt"

	playerpkg "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

func (e *GameEngine) executeSpecialActionWithRuntime(player *model.Player, actionType model.ActionType) error {
	handled, err := e.applyTimingActionStartExecuteSpecialActionOverride(player, actionType)
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
		return e.HandleExtract(p)
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

	// 检查是否有改写购买战绩区奖励的被动技能（如冒险者的地下法则）
	if entry := roleRegistry.Entry(p.Character.ID); entry.SpecialActionHook.BuyRewardOverride != nil {
		var campStones int
		if p.Camp == model.RedCamp {
			campStones = e.State.RedGems + e.State.RedCrystals
		} else {
			campStones = e.State.BlueGems + e.State.BlueCrystals
		}
		result := entry.SpecialActionHook.BuyRewardOverride(p, campStones, 5)
		if result.Handled {
			if result.AddGems > 0 {
				e.ModifyGem(string(p.Camp), result.AddGems)
			}
			if result.AddCrystals > 0 {
				e.ModifyCrystal(string(p.Camp), result.AddCrystals)
			}
			if result.LogMessage != "" {
				e.Log(result.LogMessage)
			}
			return nil
		}
	}

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
		e.AddCampCup(model.RedCamp)
		e.Log(fmt.Sprintf("[Action] %s 合成星杯！红方星杯+1，蓝方士气-1", p.Name))
		e.State.BlueMorale--
	} else {
		e.AddCampCup(model.BlueCamp)
		e.Log(fmt.Sprintf("[Action] %s 合成星杯！蓝方星杯+1，红方士气-1", p.Name))
		e.State.RedMorale--
	}
	e.checkGameEnd()
	return nil
}

func (e *GameEngine) HandleExtract(p *model.Player) error {
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
	if energyRoom <= 0 {
		return errors.New("能量已达上限，无法提炼")
	}

	var opts []interface{}
	for i := 0; i < availableGems; i++ {
		opts = append(opts, map[string]interface{}{"type": "gem"})
	}
	for i := 0; i < availableCrystals; i++ {
		opts = append(opts, map[string]interface{}{"type": "crystal"})
	}

	maxSelect := 2
	if energyRoom < maxSelect {
		maxSelect = energyRoom
	}
	if totalAvailable < maxSelect {
		maxSelect = totalAvailable
	}

	e.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: p.ID,
		Context: map[string]interface{}{
			"choice_type":       "extract",
			"extract_options":   opts,
			"extract_min":       1,
			"extract_max":       maxSelect,
			"extract_self_room": energyRoom,
		},
	})
	e.Log(fmt.Sprintf("[Action] %s 提炼：战绩区有 %d 宝石 %d 水晶，请选择 1-%d 个提炼", p.Name, availableGems, availableCrystals, maxSelect))
	return nil
}

func (e *GameEngine) getPlayerEnergyCap(player *model.Player) int {
	if player == nil {
		return 3
	}
	cap := 3
	if player.Character != nil {
		entry := roleRegistry.Entry(player.Character.ID)
		if entry.EnergyCapRule != nil {
			cap = entry.EnergyCapRule.ModifierCap(player, cap)
		}
	}
	return cap
}

// GetPlayerEnergyCap 公开方法，供 IGameEngine 接口使用。
func (e *GameEngine) GetPlayerEnergyCap(player *model.Player) int {
	return e.getPlayerEnergyCap(player)
}

// StartExtractForPlayer 为指定玩家启动提炼流程（由角色包调用）。
func (e *GameEngine) StartExtractForPlayer(playerID string) error {
	p, ok := e.State.Players[playerID]
	if !ok || p == nil {
		return fmt.Errorf("玩家不存在: %s", playerID)
	}
	return e.HandleExtract(p)
}

func (e *GameEngine) runPostSpecialActionRuntime(player *model.Player, actionType model.ActionType) {
	e.runTimingActionEndSpecialActionPost(player, actionType)
}

// applyTimingActionStartExecuteSpecialActionOverride 在执行特殊行动前应用覆盖策略。
func (e *GameEngine) applyTimingActionStartExecuteSpecialActionOverride(player *model.Player, actionType model.ActionType) (bool, error) {
	ctx := playerpkg.TimingHookContext{
		Player:     player,
		ActionType: actionType,
	}
	result := e.dispatchRoleTimingHook(playerpkg.TimingOnSpecialActionOverride, ctx)
	if result.ValidationError != nil {
		return false, result.ValidationError
	}
	return result.Handled, nil
}

// runTimingActionEndSpecialActionPost 在特殊行动完成后执行后置规则。
func (e *GameEngine) runTimingActionEndSpecialActionPost(player *model.Player, actionType model.ActionType) {
	ctx := playerpkg.TimingHookContext{
		Player:     player,
		ActionType: actionType,
	}
	e.dispatchAllRoleTimingHooks(playerpkg.TimingOnSpecialActionPost, ctx)
}
