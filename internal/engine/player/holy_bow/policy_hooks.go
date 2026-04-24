// gameflow: 圣弓策略 Hook 声明式注册。

package holy_bow

import (
	engineplayer "starcup-engine/internal/engine/player"
)

// PolicySpecs 导出圣弓策略声明。
func PolicySpecs() []engineplayer.PolicySpec {
	return []engineplayer.PolicySpec{
		// 特殊行动后置策略（圣光荣耀退出）
		{Type: engineplayer.PolicySpecialActionPost, Priority: 100, Hook: holyGloryExitHook},
	}
}

// holyGloryExitHook 圣光荣耀退出策略。
func holyGloryExitHook(host engineplayer.PolicyHost, ctx engineplayer.PolicyHookContext) engineplayer.PolicyHookResult {
	player := ctx.Player
	if player == nil {
		return engineplayer.PolicyHookResult{}
	}

	host.ApplyHolyBowHolyGloryExitHook(player, ctx.ActionType)
	return engineplayer.PolicyHookResult{Handled: true}
}
