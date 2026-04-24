// gameflow: 英雄策略 Hook 声明式注册。

package hero

import (
	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// PolicySpecs 导出英雄策略声明。
func PolicySpecs() []engineplayer.PolicySpec {
	return []engineplayer.PolicySpec{
		// 行动选项策略（嘲讽）
		{Type: engineplayer.PolicyBeforeActionOption, Priority: 200, Hook: beforeActionOptionHook},
		// 行动验证策略（嘲讽）
		{Type: engineplayer.PolicyBeforeActionValidation, Priority: 200, Hook: beforeActionValidationHook},
	}
}

// hasPlayableAttackCard 检查是否有可用的攻击牌。
func hasPlayableAttackCard(p *model.Player) bool {
	for _, card := range p.Hand {
		if card.Type == model.CardTypeAttack {
			return true
		}
	}
	return false
}

// beforeActionOptionHook 行动选项策略。
func beforeActionOptionHook(host engineplayer.PolicyHost, ctx engineplayer.PolicyHookContext) engineplayer.PolicyHookResult {
	player := ctx.Player
	modifier := ctx.OptionModifier
	rt := ctx.ChoiceRuntime

	if player == nil || modifier == nil || rt == nil {
		return engineplayer.PolicyHookResult{}
	}

	TauntOptionPolicy(rt, player, modifier, hasPlayableAttackCard)
	return engineplayer.PolicyHookResult{Handled: true}
}

// beforeActionValidationHook 行动验证策略。
func beforeActionValidationHook(host engineplayer.PolicyHost, ctx engineplayer.PolicyHookContext) engineplayer.PolicyHookResult {
	player := ctx.Player
	modifier := ctx.ValidationModifier
	rt := ctx.ChoiceRuntime

	if player == nil || modifier == nil || rt == nil {
		return engineplayer.PolicyHookResult{}
	}

	TauntValidationPolicy(rt, player, modifier, hasPlayableAttackCard)
	return engineplayer.PolicyHookResult{Handled: true}
}
