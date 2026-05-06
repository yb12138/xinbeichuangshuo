// gameflow: 月神策略 Hook 声明式注册。

package moon

import (
	engineplayer "starcup-engine/internal/engine/player"
)

// medusaInterruptHook 美杜莎之眼攻击中断策略。
func medusaInterruptHook(rt engineplayer.HookRuntime, ctx engineplayer.TimingHookContext) engineplayer.TimingHookResult {
	attacker := ctx.Attacker
	target := ctx.Target
	action := ctx.Action
	userCtx := ctx.UserCtx

	if attacker == nil || target == nil || action == nil || userCtx == nil {
		return engineplayer.TimingHookResult{}
	}

	sourceSkill := action.SourceSkill
	attackCard := action.Card

	crt := rt.AsChoiceRuntime()
	if MaybeMedusa(crt, attacker, target, sourceSkill, attackCard, userCtx) {
		return engineplayer.TimingHookResult{Handled: true, Interrupted: true}
	}
	return engineplayer.TimingHookResult{}
}
