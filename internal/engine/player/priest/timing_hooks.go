// gameflow: 牧师 Timing Hook 实现。

package priest

import (
	"starcup-engine/internal/engine/player"
)

// healCapHook 牧师治疗上限：治疗抵伤额度上限为1。
// 当 maxHeal > 1 时，返回 HealCapDelta 使得实际上限为 1。
func healCapHook(rt player.HookRuntime, ctx player.TimingHookContext) player.TimingHookResult {
	target := rt.GetPlayer(ctx.TargetID)
	if target == nil || !rt.IsCharacter(target, "priest") || ctx.HealCap <= 1 {
		return player.TimingHookResult{}
	}
	// 返回上限修正值：将上限限制为1
	return player.TimingHookResult{HealCapDelta: 1 - ctx.HealCap}
}
