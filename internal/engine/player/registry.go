// gameflow: 玩家角色入口注册表。

package player

import (
	"sort"

	"starcup-engine/internal/model"
)

// RoleRegistry 统一管理所有角色入口定义。
type RoleRegistry struct {
	entries map[string]RoleEntry
}

// NewRoleRegistry 创建角色入口注册表。
func NewRoleRegistry() *RoleRegistry {
	return &RoleRegistry{
		entries: map[string]RoleEntry{},
	}
}

// Register 注册角色入口。
func (r *RoleRegistry) Register(entry RoleEntry) {
	if r == nil || entry.ID == "" {
		return
	}
	if existing, ok := r.entries[entry.ID]; ok {
		r.entries[entry.ID] = mergeRoleEntry(existing, entry)
		return
	}
	r.entries[entry.ID] = entry
}

// Entry 返回指定角色的入口定义。
func (r *RoleRegistry) Entry(roleID string) RoleEntry {
	if r == nil || roleID == "" {
		return RoleEntry{}
	}
	if entry, ok := r.entries[roleID]; ok {
		return entry
	}
	return RoleEntry{}
}

// HandLimitRule 返回角色手牌上限规则。
func (r *RoleRegistry) HandLimitRule(roleID string) HandLimitRule {
	return r.Entry(roleID).HandLimitRule()
}

// Entries 返回按角色 ID 排序后的入口定义列表。
func (r *RoleRegistry) Entries() []RoleEntry {
	if r == nil || len(r.entries) == 0 {
		return nil
	}
	ids := make([]string, 0, len(r.entries))
	for id := range r.entries {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	entries := make([]RoleEntry, 0, len(ids))
	for _, id := range ids {
		entries = append(entries, r.entries[id])
	}
	return entries
}

// SkillUsabilityChecker 返回指定角色和技能的可用性检查器。
func (r *RoleRegistry) SkillUsabilityChecker(roleID, skillID string) SkillUsabilityChecker {
	if r == nil || roleID == "" || skillID == "" {
		return nil
	}
	entry := r.Entry(roleID)
	if entry.SkillUsabilityCheckers == nil {
		return nil
	}
	return entry.SkillUsabilityCheckers[skillID]
}

// AttackCardElementTransform 返回指定角色的攻击牌元素变换函数。
func (r *RoleRegistry) AttackCardElementTransform(roleID string) func(player *model.Player, card model.Card) model.Element {
	if r == nil || roleID == "" {
		return nil
	}
	entry := r.Entry(roleID)
	return entry.AttackCardElementTransform
}

// CannotActChecker 返回指定角色的无法行动判断函数。
func (r *RoleRegistry) CannotActChecker(roleID string) CannotActChecker {
	if r == nil || roleID == "" {
		return nil
	}
	entry := r.Entry(roleID)
	return entry.CannotActChecker
}

func mergeRoleEntry(base RoleEntry, overlay RoleEntry) RoleEntry {
	merged := base
	if merged.ID == "" {
		merged.ID = overlay.ID
	}
	if overlay.Defaults != nil {
		merged.Defaults = overlay.Defaults
	}
	if overlay.HandLimit != nil {
		merged.HandLimit = overlay.HandLimit
	}
	if overlay.TargetFilter != nil {
		merged.TargetFilter = overlay.TargetFilter
	}
	if overlay.EnergyCapRule != nil {
		merged.EnergyCapRule = overlay.EnergyCapRule
	}
	if overlay.MaxHeal != nil {
		merged.MaxHeal = overlay.MaxHeal
	}
	if overlay.Choices != nil {
		merged.Choices = overlay.Choices
	}
	if len(overlay.Skills) > 0 {
		merged.Skills = append(append([]SkillEntry{}, merged.Skills...), overlay.Skills...)
	}
	if len(overlay.ChoiceSpecs) > 0 {
		merged.ChoiceSpecs = append(append([]ChoiceSpec{}, merged.ChoiceSpecs...), overlay.ChoiceSpecs...)
	}
	merged.ChoiceRouteSpecs = mergeMap(merged.ChoiceRouteSpecs, overlay.ChoiceRouteSpecs)
	merged.FlowContinuationHandlers = mergeMap(merged.FlowContinuationHandlers, overlay.FlowContinuationHandlers)
	merged.SkillUsabilityCheckers = mergeMap(merged.SkillUsabilityCheckers, overlay.SkillUsabilityCheckers)
	if overlay.AttackCardElementTransform != nil {
		merged.AttackCardElementTransform = overlay.AttackCardElementTransform
	}
	if overlay.AttackElementResolver != nil {
		merged.AttackElementResolver = overlay.AttackElementResolver
	}
	if overlay.CannotActChecker != nil {
		merged.CannotActChecker = overlay.CannotActChecker
	}
	if overlay.HandLimitModifier != nil {
		merged.HandLimitModifier = overlay.HandLimitModifier
	}
	if len(overlay.TimingHookSpecs) > 0 {
		merged.TimingHookSpecs = append(append([]TimingHookSpec{}, merged.TimingHookSpecs...), overlay.TimingHookSpecs...)
	}
	if overlay.SpecialActionHook.BuyRewardOverride != nil {
		merged.SpecialActionHook = overlay.SpecialActionHook
	}
	return merged
}

// mergeMap 泛型合并两个 map。
func mergeMap[K comparable, V any](base, overlay map[K]V) map[K]V {
	if len(base) == 0 && len(overlay) == 0 {
		return nil
	}
	out := make(map[K]V, len(base)+len(overlay))
	for key, val := range base {
		out[key] = val
	}
	for key, val := range overlay {
		out[key] = val
	}
	return out
}
