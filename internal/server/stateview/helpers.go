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

func CountSwordEmperorSwordSouls(p *model.Player) int {
	if p == nil {
		return 0
	}
	count := 0
	for _, fc := range p.Field {
		if fc == nil || fc.Mode != model.FieldCover || fc.Effect != model.EffectSwordSoul {
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
			(clone.Effect == model.EffectMagicBowCharge || clone.Effect == model.EffectSpiritCasterPower || clone.Effect == model.EffectMoonDarkMoon || clone.Effect == model.EffectButterflyCocoon || clone.Effect == model.EffectElfBlessing || clone.Effect == model.EffectSwordSoul) {
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
			} else if clone.Effect == model.EffectSwordSoul {
				maskedName = "剑魂"
			}
			maskedID := clone.Card.ID
			maskedType := clone.Card.Type
			if clone.Effect == model.EffectButterflyCocoon {
				maskedID = ""
				maskedType = ""
			}
			clone.Card = model.Card{
				ID:          maskedID,
				Name:        maskedName,
				Type:        maskedType,
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

// PublicFlowState 声明哪些 TurnState.SkillFlowState 键属于“持续中的公开效果”（凋零保护、额外回合、
// 下一击加成……），需要下发给所有玩家并在座位上展示。其余流程记账键保持隐藏。
// value 为指示物键名（前端 TOKEN_DISPLAY / battleMarks 按此键配置名称与说明）。
var PublicFlowState = map[string]string{
	"bt_wither_active":      "bt_wither_active",      // 凋零：对方阵营士气最低为1，到蝶舞者下回合开始
	"mg_extra_turn_pending": "mg_extra_turn_pending", // 苍白之月①：本回合结束后再进行一个回合
}

// publicUsedSkillCounts 记在 TurnState.UsedSkillCounts 里、但实际是“本回合持续效果”的计数。
var publicUsedSkillCounts = map[string]string{
	"mg_next_attack_no_counter": "mg_next_attack_no_counter", // 苍白之月①：下次主动攻击不可应战
}

// publicNextAttackBonuses 跨行动持续到回合结束的“下次主动攻击 +N”（规则修正 ID）。
// 只在一次攻击内有效的加成（怒吼、蓄力一击、火之矢……）在攻击结算时已体现，不单独展示。
var publicNextAttackBonuses = []string{
	"ml_dark_release_next_attack_bonus",
	"ml_fullness_next_attack_bonus",
}

// PublicIndicators 返回玩家对所有人公开的派生指示物：盖牌计数、规则修正与公开的技能流程状态。
// 同步状态（SyncState）与时间线快照共用这一份，保证两边一致。
func PublicIndicators(state *model.GameState, p *model.Player) map[string]int {
	indicators := map[string]int{}
	if p == nil {
		return indicators
	}
	set := func(key string, value int) {
		if value > 0 {
			indicators[key] = value
		}
	}
	set("elf_blessing_count", CountElfBlessings(p))
	set("mb_charge_count", CountMagicBowCharges(p))
	set("sc_power_count", CountSpiritCasterPowers(p))
	set("mg_dark_moon_count", CountMoonDarkMoons(p))
	set("bt_cocoon_count", CountButterflyCocoons(p))
	set("bp_shared_life_active", CountBloodSharedLifeAsSource(state, p.ID))
	set("bp_shared_life_bound", CountBloodSharedLifeAsHolder(p))
	set("se_sword_soul_count", CountSwordEmperorSwordSouls(p))
	for _, modifierID := range publicNextAttackBonuses {
		set(modifierID, CombatPolicyAttackBonusByModifierID(p, modifierID))
	}
	if HasRuleModifierWithModifierID(p, "ml_dark_release_lock_turn") {
		indicators["ml_dark_release_lock_turn"] = 1
	}
	for flowKey, indicatorKey := range PublicFlowState {
		set(indicatorKey, p.TurnState.SkillFlowState[flowKey])
	}
	for countKey, indicatorKey := range publicUsedSkillCounts {
		set(indicatorKey, p.TurnState.UsedSkillCounts[countKey])
	}
	return indicators
}

// DerivedIndicatorKeys 是 PublicIndicators 里由引擎状态派生的键：Tokens 里若有同名镜像需删除，避免重复展示。
var DerivedIndicatorKeys = []string{
	"elf_blessing_count",
	"mb_charge_count",
	"sc_power_count",
	"mg_dark_moon_count",
	"bt_cocoon_count",
	"bp_shared_life_active",
	"bp_shared_life_bound",
	"css_blood_cap",
	"ml_dark_release_next_attack_bonus",
	"ml_fullness_next_attack_bonus",
	"ml_dark_release_lock_turn",
	"se_sword_soul_count",
}
