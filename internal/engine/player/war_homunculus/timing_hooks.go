// gameflow: 英灵人形 Timing Hook 实现。

package war_homunculus

import (
	"fmt"

	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// turnEndHook 回合结束：符文改造退场 + 手牌上限检查。
func turnEndHook(rt engineplayer.HookRuntime, ctx engineplayer.TimingHookContext) engineplayer.TimingHookResult {
	p := rt.GetPlayer(ctx.SourceID)
	if p == nil || !rt.IsCharacter(p, "war_homunculus") {
		return engineplayer.TimingHookResult{}
	}
	if !rt.HasForm(p, model.FormWarHomunculusBurst) {
		return engineplayer.TimingHookResult{}
	}
	rt.ClearForm(p, model.FormWarHomunculusBurst)
	rt.Log(fmt.Sprintf("%s 的 [符文改造] 效果结束，脱离蓄势迸发形态", p.Name))
	rt.CheckHandLimit(p)
	if rt.GetPendingInterrupt() != nil {
		return engineplayer.TimingHookResult{Interrupted: true}
	}
	return engineplayer.TimingHookResult{}
}
