// gameflow: Timing Hook 注册与分派基础设施。

package engine

import (
	"sort"

	engineplayer "starcup-engine/internal/engine/player"
)

type roleTimingHookEntry struct {
	RoleID   string
	Priority int
	Hook     engineplayer.TimingHookFunc
}

// mountRoleTimingHooks 从 RoleRegistry 收集 TimingHookSpecs，按 timing 分组并排序。
func mountRoleTimingHooks() map[engineplayer.TimingPoint][]roleTimingHookEntry {
	hooks := make(map[engineplayer.TimingPoint][]roleTimingHookEntry)
	for _, entry := range roleRegistry.Entries() {
		for _, spec := range entry.TimingHookSpecs {
			s := spec
			hooks[s.Timing] = append(hooks[s.Timing], roleTimingHookEntry{
				RoleID:   entry.ID,
				Priority: s.Priority,
				Hook:     s.Hook,
			})
		}
	}
	for timing := range hooks {
		sort.Slice(hooks[timing], func(i, j int) bool {
			return hooks[timing][i].Priority < hooks[timing][j].Priority
		})
	}
	return hooks
}

// dispatchRoleTimingHook 按 timing 点分派 Hook，首个返回 Interrupted 即停止。
func (e *GameEngine) dispatchRoleTimingHook(
	timing engineplayer.TimingPoint,
	ctx engineplayer.TimingHookContext,
) bool {
	hooks := e.roleTimingHooks[timing]
	rt := newHookRuntime(e)
	for _, entry := range hooks {
		result := entry.Hook(rt, ctx)
		if result.Interrupted {
			return true
		}
	}
	return false
}
