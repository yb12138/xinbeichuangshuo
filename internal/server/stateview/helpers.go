package stateview

import "starcup-engine/internal/model"

func HasRuleModifierWithModifierID(p *model.Player, modifierID string) bool {
	if p == nil || modifierID == "" || len(p.ActiveRuleModifiers) == 0 {
		return false
	}
	for _, modifier := range p.ActiveRuleModifiers {
		if modifier != nil && modifier.ModifierID == modifierID {
			return true
		}
	}
	return false
}

func CombatPolicyAttackBonusByModifierID(p *model.Player, modifierID string) int {
	if p == nil || modifierID == "" || len(p.ActiveRuleModifiers) == 0 {
		return 0
	}
	total := 0
	for _, modifier := range p.ActiveRuleModifiers {
		if modifier == nil || modifier.ModifierID != modifierID || modifier.Domain != model.RuleModifierDomainCombatPolicy || modifier.CombatPolicyPayload == nil {
			continue
		}
		total += modifier.CombatPolicyPayload.AttackDamageBonus
	}
	return total
}

func CountMagicBowChargesByElement(p *model.Player, element model.Element) int {
	if p == nil {
		return 0
	}
	count := 0
	for _, fc := range p.Field {
		if fc == nil || fc.Mode != model.FieldCover || fc.Effect != model.EffectMagicBowCharge {
			continue
		}
		if element != "" && fc.Card.Element != element {
			continue
		}
		count++
	}
	return count
}

func CountMagicBowCharges(p *model.Player) int {
	return CountMagicBowChargesByElement(p, "")
}

func CountSpiritCasterPowers(p *model.Player) int {
	if p == nil {
		return 0
	}
	count := 0
	for _, fc := range p.Field {
		if fc == nil || fc.Mode != model.FieldCover || fc.Effect != model.EffectSpiritCasterPower {
			continue
		}
		count++
	}
	return count
}

func CountMoonDarkMoons(p *model.Player) int {
	if p == nil {
		return 0
	}
	count := 0
	for _, fc := range p.Field {
		if fc == nil || fc.Mode != model.FieldCover || fc.Effect != model.EffectMoonDarkMoon {
			continue
		}
		count++
	}
	return count
}

func CountButterflyCocoons(p *model.Player) int {
	if p == nil {
		return 0
	}
	count := 0
	for _, fc := range p.Field {
		if fc == nil || fc.Mode != model.FieldCover || fc.Effect != model.EffectButterflyCocoon {
			continue
		}
		count++
	}
	return count
}

func CountElfBlessings(p *model.Player) int {
	if p == nil {
		return 0
	}
	count := 0
	for _, fc := range p.Field {
		if fc == nil || fc.Mode != model.FieldCover || fc.Effect != model.EffectElfBlessing {
			continue
		}
		count++
	}
	return count
}

func CountBloodSharedLifeAsSource(state *model.GameState, sourceID string) int {
	if state == nil || sourceID == "" {
		return 0
	}
	count := 0
	for _, p := range state.Players {
		if p == nil {
			continue
		}
		for _, fc := range p.Field {
			if fc == nil || fc.Mode != model.FieldEffect || fc.Effect != model.EffectBloodSharedLife {
				continue
			}
			if fc.SourceID == sourceID {
				count++
			}
		}
	}
	return count
}

func CountBloodSharedLifeAsHolder(player *model.Player) int {
	if player == nil {
		return 0
	}
	count := 0
	for _, fc := range player.Field {
		if fc == nil || fc.Mode != model.FieldEffect || fc.Effect != model.EffectBloodSharedLife {
			continue
		}
		count++
	}
	return count
}

func BuildMaskedFieldForViewer(owner *model.Player, viewerID string) []*model.FieldCard {
	if owner == nil || len(owner.Field) == 0 {
		return nil
	}
	out := make([]*model.FieldCard, 0, len(owner.Field))
	for _, fc := range owner.Field {
		if fc == nil {
			continue
		}
		clone := *fc
		if owner.ID != viewerID && clone.Mode == model.FieldCover &&
			(clone.Effect == model.EffectMagicBowCharge || clone.Effect == model.EffectSpiritCasterPower || clone.Effect == model.EffectMoonDarkMoon || clone.Effect == model.EffectButterflyCocoon || clone.Effect == model.EffectElfBlessing) {
			maskedName := "盖牌"
			if clone.Effect == model.EffectMagicBowCharge {
				maskedName = "充能"
			} else if clone.Effect == model.EffectSpiritCasterPower {
				maskedName = "妖力"
			} else if clone.Effect == model.EffectMoonDarkMoon {
				maskedName = "暗月"
			} else if clone.Effect == model.EffectButterflyCocoon {
				maskedName = "茧"
			} else if clone.Effect == model.EffectElfBlessing {
				maskedName = "祝福"
			}
			clone.Card = model.Card{
				ID:          clone.Card.ID,
				Name:        maskedName,
				Type:        clone.Card.Type,
				Description: "盖牌（内容对他人不可见）",
			}
		}
		out = append(out, &clone)
	}
	return out
}

// IsSkillBlockedBySkillGate 检查玩家是否有SkillGate规则锁定指定技能。
// 前端可通过此函数判断技能按钮是否应该变灰。
func IsSkillBlockedBySkillGate(p *model.Player, skillID string) bool {
	if p == nil || skillID == "" || len(p.ActiveRuleModifiers) == 0 {
		return false
	}
	for _, modifier := range p.ActiveRuleModifiers {
		if modifier == nil || modifier.Domain != model.RuleModifierDomainSkillGate || modifier.SkillGatePayload == nil {
			continue
		}
		if modifier.SkillGatePayload.Mode != model.SkillGateDisallowList {
			continue
		}
		for _, blockedID := range modifier.SkillGatePayload.SkillIDs {
			if blockedID == skillID {
				return true
			}
		}
	}
	return false
}

// GetBlockedSkillIDsBySkillGate 获取玩家当前所有被SkillGate锁定的技能ID列表。
// 前端可通过此列表批量判断哪些技能按钮需要变灰。
func GetBlockedSkillIDsBySkillGate(p *model.Player) []string {
	if p == nil || len(p.ActiveRuleModifiers) == 0 {
		return nil
	}
	blockedIDs := make([]string, 0)
	for _, modifier := range p.ActiveRuleModifiers {
		if modifier == nil || modifier.Domain != model.RuleModifierDomainSkillGate || modifier.SkillGatePayload == nil {
			continue
		}
		if modifier.SkillGatePayload.Mode != model.SkillGateDisallowList {
			continue
		}
		blockedIDs = append(blockedIDs, modifier.SkillGatePayload.SkillIDs...)
	}
	return blockedIDs
}
