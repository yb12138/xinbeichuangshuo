// gameflow: 技能能否在当阶段使用（启动技/主动技/响应禁止手动等）。

package engine

import (
	"fmt"

	"starcup-engine/internal/model"
)

// 技能发动流程（校验阶段）：
// 1) 交互弃牌检查 -> 2) 技能窗口合法性 -> 3) 目标合法性与场牌落点合法性。
func (e *GameEngine) maybeRequestSkillDiscardSelection(use *skillUseRequest) (bool, error) {
	if use.requiredDiscards <= 0 || len(use.discardIndices) > 0 {
		return false, nil
	}
	if len(use.player.Hand) < use.requiredDiscards {
		return false, fmt.Errorf("手牌不足：发动 [%s] 需要弃置 %d 张牌", use.skillDef.Title, use.requiredDiscards)
	}

	e.State.PendingInterrupt = newDiscardChoiceInterrupt(use.player.ID, map[string]interface{}{
		"discard_count": use.requiredDiscards,
		"skill_id":      use.skillID,
		"target_ids":    use.targetIDs,
		"resume_phase":  e.currentChoiceResumePoint(),
	})
	e.State.PendingInterrupt.SkillIDs = []string{use.skillID}
	e.syncGamePhaseWithInterrupt(e.State.PendingInterrupt)
	e.Log(fmt.Sprintf("%s 请选择用于发动 [%s] 的卡牌", use.player.Name, use.skillDef.Title))
	return true, nil
}

func (e *GameEngine) validateSkillDiscardSelection(use *skillUseRequest) error {
	seen := map[int]bool{}
	for _, idx := range use.discardIndices {
		if seen[idx] {
			return fmt.Errorf("不能重复选择同一张牌")
		}
		seen[idx] = true
	}

	if use.requiredDiscards > 0 && len(use.discardIndices) != use.requiredDiscards {
		return fmt.Errorf("技能需要弃 %d 张牌，你选择了 %d 张", use.requiredDiscards, len(use.discardIndices))
	}

	discardedCards := make([]model.Card, 0, len(use.discardIndices))
	for _, idx := range use.discardIndices {
		if idx < 0 || idx >= len(use.player.Hand) {
			return fmt.Errorf("弃牌索引越界: %d", idx)
		}

		card := use.player.Hand[idx]
		effectiveElement := card.Element
		if use.skillDef.DiscardElement != "" {
			for _, entry := range roleRegistry.Entries() {
				if entry.AttackElementResolver != nil {
					effectiveElement = entry.AttackElementResolver(use.player, card)
					break
				}
			}
		}
		if use.skillDef.DiscardElement != "" && effectiveElement != use.skillDef.DiscardElement {
			return fmt.Errorf("弃牌 %s 不符合元素要求", card.Name)
		}
		if use.skillDef.DiscardType != "" && card.Type != use.skillDef.DiscardType {
			return fmt.Errorf("弃牌 %s 不符合卡牌类型要求", card.Name)
		}
		if use.skillDef.DiscardFate != "" && card.Faction != use.skillDef.DiscardFate {
			return fmt.Errorf("弃牌 %s 不符合命格要求", card.Name)
		}
		if use.skillDef.RequireExclusive && !card.MatchExclusive(use.player.Character.ID, use.skillDef.Title) {
			return fmt.Errorf("弃牌 %s 不是该技能对应的独有牌", card.Name)
		}
		discardedCards = append(discardedCards, card)
	}
	use.discardedCards = discardedCards

	if use.policy.ValidateDiscardedCards != nil {
		if err := use.policy.ValidateDiscardedCards(use.policyContext()); err != nil {
			return err
		}
	}

	if use.skillDef.RequireExclusive && use.skillDef.CostDiscards <= 0 && len(use.discardedCards) == 0 {
		if use.player.Character == nil || use.player.Character.ID == "" {
			return fmt.Errorf("角色信息缺失，无法校验独有牌")
		}
		if use.policy.ManualExclusiveCard {
			if !use.player.HasExclusiveCard(use.player.Character.ID, use.skillDef.Title) {
				return fmt.Errorf("未找到技能 [%s] 对应的专属技能卡", use.skillDef.Title)
			}
		} else {
			card, ok := use.player.ConsumeExclusiveCard(use.player.Character.ID, use.skillDef.Title)
			if !ok {
				return fmt.Errorf("未找到技能 [%s] 对应的专属技能卡", use.skillDef.Title)
			}
			use.consumedExclusiveCard = &card
		}
	}

	return nil
}

func (e *GameEngine) validateSkillActivation(use *skillUseRequest) error {
	switch use.skillDef.Type {
	case model.SkillTypeStartup:
		if !e.isStartupWindow() {
			return fmt.Errorf("startup skills can only be used during the startup skill window")
		}
	case model.SkillTypeAction:
		if !e.isActionSelectionWindow() {
			return fmt.Errorf("action skills can only be used during action phase")
		}
	case model.SkillTypeResponse:
		return fmt.Errorf("response skills are activated automatically")
	case model.SkillTypePassive:
		return fmt.Errorf("passive skills are activated automatically")
	}

	if use.player.TurnState.CurrentExtraAction == "Attack" {
		return fmt.Errorf("当前是额外攻击行动，不能使用技能，只能发起攻击")
	}
	if model.ContainsSkillTag(use.skillDef.Tags, model.TagTurnLimit) && use.player.TurnState.UsedSkillCounts[use.skillID] > 0 {
		return fmt.Errorf("skill %s can only be used once per turn", use.skillID)
	}
	return nil
}

func (e *GameEngine) consumeSkillCoverCost(use *skillUseRequest) error {
	if use.skillDef.CostCoverCards <= 0 {
		return nil
	}
	coverCards, err := use.player.ConsumeCoverCards(use.skillDef.CostCoverCards)
	if err != nil {
		return fmt.Errorf("盖牌消耗失败: %v", err)
	}
	e.State.DiscardPile = append(e.State.DiscardPile, coverCards...)
	e.Log(fmt.Sprintf("%s 消耗了 %d 张盖牌作为技能消耗", use.player.Name, use.skillDef.CostCoverCards))
	return nil
}

func (e *GameEngine) ensureSkillEnergyCost(use *skillUseRequest) error {
	if canPaySkillEnergyCost(use.player, use.skillDef.CostGem, use.skillDef.CostCrystal) {
		return nil
	}
	return fmt.Errorf(
		"资源不足: 需要 宝石%d/水晶%d，当前 宝石%d/水晶%d（红宝石可替代水晶）",
		use.skillDef.CostGem, use.skillDef.CostCrystal, use.player.Gem, use.player.Crystal,
	)
}

func (e *GameEngine) resolveSkillTargets(use *skillUseRequest) error {
	use.actualTargets = nil
	use.target = nil

	if use.skillDef.TargetType == model.TargetNone {
		return validateTargetRules(use.policy.TargetRules, use)
	}

	actualTargets := make([]*model.Player, 0, len(use.targetIDs))
	for _, id := range use.targetIDs {
		target := e.State.Players[id]
		if target == nil {
			return fmt.Errorf("target player %s not found", id)
		}
		actualTargets = append(actualTargets, target)
	}
	use.actualTargets = actualTargets
	if len(actualTargets) == 1 {
		use.target = actualTargets[0]
	}

	minTargets, maxTargets, countErr := effectiveTargetCountRange(use)
	if (maxTargets > 0 && len(actualTargets) > maxTargets) || len(actualTargets) < minTargets {
		return formatTargetCountError(len(actualTargets), minTargets, maxTargets, countErr)
	}

	if err := validateTargetTypeConstraints(use); err != nil {
		return err
	}
	return validateTargetRules(use.policy.TargetRules, use)
}

func (e *GameEngine) validateSkillFieldPlacement(use *skillUseRequest) error {
	if !use.skillDef.PlaceCard || len(use.actualTargets) == 0 {
		return nil
	}
	fieldTarget := use.actualTargets[0]
	if use.skillDef.PlaceMode == model.FieldEffect && model.IsBasicEffect(string(use.skillDef.PlaceEffect)) {
		for _, fc := range fieldTarget.Field {
			if fc == nil || fc.Mode != model.FieldEffect {
				continue
			}
			if fc.Effect == use.skillDef.PlaceEffect {
				return fmt.Errorf("%s 面前已有同种基础效果，不可重复放置", fieldTarget.Name)
			}
		}
	}
	return nil
}
