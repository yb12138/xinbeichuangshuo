// gameflow: 烈焰魔女策略 Hook 声明式注册。

package blaze_witch

import (
	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// PolicySpecs 导出烈焰魔女策略声明。
func PolicySpecs() []engineplayer.PolicySpec {
	return []engineplayer.PolicySpec{
		// 攻击牌变换策略（烈焰形态变换）
		{Type: engineplayer.PolicyAttackCardTransform, Priority: 100, Hook: attackCardTransformHook},
	}
}

// attackCardTransformHook 攻击牌变换策略。
func attackCardTransformHook(host engineplayer.PolicyHost, ctx engineplayer.PolicyHookContext) engineplayer.PolicyHookResult {
	player := ctx.Player
	card := ctx.CounterCard

	if player == nil {
		return engineplayer.PolicyHookResult{Card: card}
	}

	transformed := host.ApplyBlazeWitchAttackCardTransform(player, card)
	// 设置 Stop: true 以短路其他 hook 并返回变换后的卡牌
	return engineplayer.PolicyHookResult{Handled: true, Stop: true, Card: transformed}
}

// 元素变换辅助函数（供 Hook 使用）
func transformToFire(card model.Card) model.Card {
	card.Element = model.ElementFire
	return card
}
