// gameflow: 烈焰魔女策略 Hook 声明式注册。

package blaze_witch

import (
	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// attackCardTransformHook 攻击牌变换策略。
func attackCardTransformHook(rt engineplayer.HookRuntime, ctx engineplayer.TimingHookContext) engineplayer.TimingHookResult {
	player := ctx.Player
	card := ctx.CounterCard

	if player == nil {
		return engineplayer.TimingHookResult{Card: *card}
	}

	transformed := ApplyAttackCardTransform(player, *card)
	// 设置 SkipNextHook: true 以短路其他 hook 并返回变换后的卡牌
	return engineplayer.TimingHookResult{Handled: true, SkipNextHook: true, Card: transformed}
}

// 元素变换辅助函数（供 Hook 使用）
func transformToFire(card model.Card) model.Card {
	card.Element = model.ElementFire
	return card
}
