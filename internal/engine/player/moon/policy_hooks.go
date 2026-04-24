// gameflow: 月神策略 Hook 声明式注册。

package moon

import (
	engineplayer "starcup-engine/internal/engine/player"
)

// PolicySpecs 导出月神策略声明。
func PolicySpecs() []engineplayer.PolicySpec {
	return []engineplayer.PolicySpec{
		// 攻击宣言中断策略（美杜莎之眼）
		{Type: engineplayer.PolicyAttackDeclaredInterrupt, Priority: 100, Hook: medusaInterruptHook},
	}
}

// medusaInterruptHook 美杜莎之眼攻击中断策略。
func medusaInterruptHook(host engineplayer.PolicyHost, ctx engineplayer.PolicyHookContext) engineplayer.PolicyHookResult {
	attacker := ctx.Attacker
	target := ctx.Target
	action := ctx.Action
	userCtx := ctx.UserCtx

	if attacker == nil || target == nil || action == nil || userCtx == nil {
		return engineplayer.PolicyHookResult{}
	}

	if host.ApplyMoonMedusaInterrupt(attacker, target, action, userCtx) {
		return engineplayer.PolicyHookResult{Handled: true, Stop: true}
	}
	return engineplayer.PolicyHookResult{}
}
