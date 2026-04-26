// gameflow: 冒险者策略 Hook 声明式注册。

package adventurer

import (
	engineplayer "starcup-engine/internal/engine/player"
)

// undergroundLawOverrideHook 地下法则特殊行动覆盖策略。
// 保持与原 PolicySpec stub 一致：不覆盖，让正常购买流程（handleBuy → hasBuyRewriteSkill → executeBuyRewrite）继续。
func undergroundLawOverrideHook(rt engineplayer.HookRuntime, ctx engineplayer.TimingHookContext) engineplayer.TimingHookResult {
	return engineplayer.TimingHookResult{}
}
