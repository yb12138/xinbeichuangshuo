// gameflow: Timing Hook 注册与分派基础设施。

package engine

import (
	"sort"

	engineplayer "starcup-engine/internal/engine/player"
)

type roleTimingHookEntry struct {
	RoleID     string
	Priority   int
	Hook       engineplayer.TimingHookFunc
	RoleFilter *string // nil=按SourceID角色过滤, "none"=不过滤
}

// mountRoleTimingHooks 从 RoleRegistry 收集 TimingHookSpecs，按 timing 分组并排序。
func mountRoleTimingHooks() map[engineplayer.TimingPoint][]roleTimingHookEntry {
	hooks := make(map[engineplayer.TimingPoint][]roleTimingHookEntry)
	for _, entry := range roleRegistry.Entries() {
		for _, spec := range entry.TimingHookSpecs {
			s := spec
			hooks[s.Timing] = append(hooks[s.Timing], roleTimingHookEntry{
				RoleID:     entry.ID,
				Priority:   s.Priority,
				Hook:       s.Hook,
				RoleFilter: s.RoleFilter,
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

// hookMatchesRole 判断 hook 是否应对当前上下文执行。
// 默认（RoleFilter==nil）：当上下文中任一相关玩家的 Character.ID 匹配 hook 的 RoleID 时执行。
// 检查顺序：SourceID → TargetID → ctx.Player
// 若三者均未设置，则放行（无法判定角色，保守执行）。
// RoleFilter=="none"：不按角色过滤，全局执行（跨角色 hook）。
func hookMatchesRole(entry roleTimingHookEntry, e *GameEngine, ctx engineplayer.TimingHookContext) bool {
	if entry.RoleFilter != nil && *entry.RoleFilter == engineplayer.HookRoleNone {
		return true
	}
	// Check SourceID and TargetID
	for _, pid := range []string{ctx.SourceID, ctx.TargetID} {
		if pid == "" {
			continue
		}
		p := e.State.Players[pid]
		if p != nil && p.Character != nil && p.Character.ID == entry.RoleID {
			return true
		}
	}
	// Check ctx.Player (used by policy hooks like BeforeActionOption/Validation)
	if ctx.Player != nil && ctx.Player.Character != nil && ctx.Player.Character.ID == entry.RoleID {
		return true
	}
	// No player context available — conservatively allow the hook to run
	if ctx.SourceID == "" && ctx.TargetID == "" && ctx.Player == nil {
		return true
	}
	return false
}

// dispatchRoleTimingHook 按 timing 点分派 Hook，首个返回 Interrupted/Blocked/SkipNextHook/ValidationError 即停止。
func (e *GameEngine) dispatchRoleTimingHook(
	timing engineplayer.TimingPoint,
	ctx engineplayer.TimingHookContext,
) engineplayer.TimingHookResult {
	hooks := e.roleTimingHooks[timing]
	rt := newHookRuntime(e)
	for _, entry := range hooks {
		if !hookMatchesRole(entry, e, ctx) {
			continue
		}
		result := entry.Hook(rt, ctx)
		if result.Handled || result.Interrupted || result.Blocked || result.SkipNextHook || result.ValidationError != nil {
			return result
		}
	}
	return engineplayer.TimingHookResult{}
}

// dispatchAllRoleTimingHooks 按 timing 点分派 Hook，全部执行（不短路），收集所有结果。
func (e *GameEngine) dispatchAllRoleTimingHooks(
	timing engineplayer.TimingPoint,
	ctx engineplayer.TimingHookContext,
) engineplayer.TimingHookResult {
	hooks := e.roleTimingHooks[timing]
	rt := newHookRuntime(e)
	aggregated := engineplayer.TimingHookResult{}
	for _, entry := range hooks {
		if !hookMatchesRole(entry, e, ctx) {
			continue
		}
		result := entry.Hook(rt, ctx)
		if result.Interrupted {
			aggregated.Interrupted = true
		}
		if result.Blocked {
			aggregated.Blocked = true
			if aggregated.BlockReason == "" && result.BlockReason != "" {
				aggregated.BlockReason = result.BlockReason
			}
		}
		aggregated.DamageDelta += result.DamageDelta
		aggregated.HealCapDelta += result.HealCapDelta
		if result.ValidationError != nil && aggregated.ValidationError == nil {
			aggregated.ValidationError = result.ValidationError
		}
		if result.Handled {
			aggregated.Handled = true
		}
		if result.UseFaction {
			aggregated.UseFaction = true
		}
		if result.CounterAllowed {
			aggregated.CounterAllowed = true
		}
		if result.CounterCard != nil {
			aggregated.CounterCard = result.CounterCard
		}
		if result.Card.Name != "" {
			aggregated.Card = result.Card
		}
		if len(result.SkillIDs) > 0 {
			aggregated.SkillIDs = append(aggregated.SkillIDs, result.SkillIDs...)
		}
		if result.PlayerAction.Type != "" {
			aggregated.PlayerAction = result.PlayerAction
		}
	}
	return aggregated
}
