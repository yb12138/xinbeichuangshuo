// gameflow: 冒险者策略 Hook 声明式注册。

package adventurer

import (
	engineplayer "starcup-engine/internal/engine/player"
)

// PolicySpecs 导出冒险者策略声明。
func PolicySpecs() []engineplayer.PolicySpec {
	return []engineplayer.PolicySpec{
		// 特殊行动覆盖策略（地下法则）
		{Type: engineplayer.PolicySpecialActionOverride, Priority: 100, Hook: undergroundLawOverrideHook},
	}
}

// undergroundLawOverrideHook 地下法则特殊行动覆盖策略。
func undergroundLawOverrideHook(host engineplayer.PolicyHost, ctx engineplayer.PolicyHookContext) engineplayer.PolicyHookResult {
	player := ctx.Player
	action := ctx.PlayerAction

	if player == nil {
		return engineplayer.PolicyHookResult{}
	}

	newAction, handled := host.ApplyAdventurerUndergroundLawOverride(player, action)
	if handled {
		return engineplayer.PolicyHookResult{Handled: true, Stop: true, PlayerAction: newAction}
	}
	return engineplayer.PolicyHookResult{}
}
