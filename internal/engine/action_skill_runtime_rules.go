// gameflow: 行动阶段使用技能的运行时规则。

package engine

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

func (e *GameEngine) hasUsableActionSkillForExtraMagic(playerObj *model.Player) bool {
	if playerObj == nil || playerObj.Character == nil {
		return false
	}

	for _, skillDef := range playerObj.Character.Skills {
		if skillDef.Type != model.SkillTypeAction {
			continue
		}
		if e.isActionSkillUsableForExtraMagic(playerObj, skillDef) {
			return true
		}
	}
	return false
}

func (e *GameEngine) isActionSkillUsableForExtraMagic(playerObj *model.Player, skillDef model.SkillDefinition) bool {
	// 回合限定：本回合已用过则不可再用。
	if model.ContainsSkillTag(skillDef.Tags, model.TagTurnLimit) && playerObj.TurnState.UsedSkillCounts[skillDef.ID] > 0 {
		return false
	}
	// 资源校验（宝石/水晶）。
	if !canPaySkillEnergyCost(playerObj, skillDef.CostGem, skillDef.CostCrystal) {
		return false
	}
	// 独有技：需拥有对应独有牌（手牌或专属卡区）。
	if skillDef.RequireExclusive && !playerObj.HasExclusiveCard(playerObj.Character.ID, skillDef.Title) {
		return false
	}
	// 弃牌成本可达成性。
	if !e.canSatisfyActionSkillDiscardRequirement(playerObj, skillDef) {
		return false
	}
	// 目标可达成性（仅做最小目标数校验）。
	if !e.hasActionSkillValidTarget(playerObj, skillDef) {
		return false
	}

	// 角色特定技能可用性检查：通过注册表调用角色包的检查器。
	if playerObj.Character != nil {
		checker := roleRegistry.SkillUsabilityChecker(playerObj.Character.ID, skillDef.ID)
		if checker != nil {
			if !checker(e, playerObj, skillDef) {
				return false
			}
		}
	}

	return true
}

func (e *GameEngine) canSatisfyActionSkillDiscardRequirement(playerObj *model.Player, skillDef model.SkillDefinition) bool {
	requiredDiscards := skillDef.CostDiscards
	if requiredDiscards <= 0 {
		return true
	}

	matched := 0
	for _, card := range playerObj.Hand {
		effectiveElement := card.Element
		// 如果需要特定元素弃牌，检查是否需要通过角色变换（如苍炎魔女火焰形态）。
		if skillDef.DiscardElement != "" && playerObj.Character != nil {
			transformFn := roleRegistry.AttackCardElementTransform(playerObj.Character.ID)
			if transformFn != nil {
				effectiveElement = transformFn(playerObj, card)
			}
		}
		if skillDef.DiscardElement != "" && effectiveElement != skillDef.DiscardElement {
			continue
		}
		if skillDef.DiscardType != "" && card.Type != skillDef.DiscardType {
			continue
		}
		if skillDef.DiscardFate != "" && card.Faction != skillDef.DiscardFate {
			continue
		}
		if skillDef.RequireExclusive && !card.MatchExclusive(playerObj.Character.ID, skillDef.Title) {
			continue
		}
		matched++
		if matched >= requiredDiscards {
			return true
		}
	}
	return false
}

func (e *GameEngine) hasActionSkillValidTarget(playerObj *model.Player, skillDef model.SkillDefinition) bool {
	switch skillDef.TargetType {
	case model.TargetNone, model.TargetSelf:
		return true
	}

	minTargets := skillDef.MinTargets
	if minTargets <= 0 {
		if skillDef.TargetType >= model.TargetEnemy {
			minTargets = 1
		} else {
			minTargets = 0
		}
	}
	if minTargets <= 0 {
		return true
	}

	candidates := 0
	for _, playerID := range e.State.PlayerOrder {
		target := e.State.Players[playerID]
		if target == nil {
			continue
		}
		switch skillDef.TargetType {
		case model.TargetEnemy:
			if target.Camp != playerObj.Camp {
				candidates++
			}
		case model.TargetAlly:
			if target.Camp == playerObj.Camp && target.ID != playerObj.ID {
				candidates++
			}
		case model.TargetAllySelf:
			if target.Camp == playerObj.Camp {
				candidates++
			}
		case model.TargetAny, model.TargetSpecific:
			candidates++
		}
		if candidates >= minTargets {
			return true
		}
	}
	return false
}

// SkillUsabilityCheckerEngine 接口实现方法。
func (e *GameEngine) LookupPlayer(playerID string) *model.Player {
	if e == nil || e.State == nil {
		return nil
	}
	return e.State.Players[playerID]
}

func (e *GameEngine) PlayerOrder() []string {
	if e == nil || e.State == nil {
		return nil
	}
	return e.State.PlayerOrder
}

func (e *GameEngine) HasForm(p *model.Player, form string) bool {
	return player.HasForm(p, form)
}

func (e *GameEngine) IsCharacter(p *model.Player, roleID string) bool {
	return player.IsCharacter(p, roleID)
}

func (e *GameEngine) GetToken(p *model.Player, key string) int {
	if p == nil || p.Tokens == nil {
		return 0
	}
	return p.Tokens[key]
}

func (e *GameEngine) CountCoverCardsByEffectAndElement(p *model.Player, effect model.EffectType, element model.Element) int {
	return e.countCoverCardsByEffectAndElement(p, effect, element)
}
