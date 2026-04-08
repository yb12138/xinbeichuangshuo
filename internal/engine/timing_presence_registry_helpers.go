package engine

// presenceHookEntry 描述“角色在场 -> 挂载规则函数”的一条声明式配置。
// requireAny 为空表示该规则属于阶段通用规则，始终生效。
type presenceHookEntry[T any] struct {
	requireAny []string
	hook       T
}

func requireAny(roleIDs ...string) []string {
	return roleIDs
}

func buildPresenceHooks[T any](present map[string]bool, entries []presenceHookEntry[T]) []T {
	var hooks []T
	for _, entry := range entries {
		if !matchPresenceAny(present, entry.requireAny) {
			continue
		}
		hooks = append(hooks, entry.hook)
	}
	return hooks
}

func matchPresenceAny(present map[string]bool, roleIDs []string) bool {
	if len(roleIDs) == 0 {
		return true
	}
	for _, roleID := range roleIDs {
		if present[roleID] {
			return true
		}
	}
	return false
}
