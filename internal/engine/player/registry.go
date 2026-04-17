// gameflow: 玩家角色入口注册表。

package player

import (
	"sort"
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
	if overlay.MaxHeal != nil {
		merged.MaxHeal = overlay.MaxHeal
	}
	if overlay.Choices != nil {
		merged.Choices = overlay.Choices
	}
	if len(overlay.Skills) > 0 {
		merged.Skills = append(append([]SkillEntry{}, merged.Skills...), overlay.Skills...)
	}
	merged.ChoiceRouteSpecs = mergeMap(merged.ChoiceRouteSpecs, overlay.ChoiceRouteSpecs)
	merged.FollowupSpecs = mergeMap(merged.FollowupSpecs, overlay.FollowupSpecs)
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
