// gameflow: 英雄策略 Hook 声明式注册。

package hero

import (
	engineplayer "starcup-engine/internal/engine/player"
)

// beforeActionOptionHook 行动选项策略。
func beforeActionOptionHook(rt engineplayer.HookRuntime, ctx engineplayer.TimingHookContext) engineplayer.TimingHookResult {
	player := ctx.Player
	modifier := ctx.OptionModifier
	choiceRt := ctx.ChoiceRuntime

	if player == nil || modifier == nil || choiceRt == nil {
		return engineplayer.TimingHookResult{}
	}

	TauntOptionPolicy(choiceRt, player, modifier, rt.HasPlayableAttackCard)
	return engineplayer.TimingHookResult{Handled: true}
}

// beforeActionValidationHook 行动验证策略。
func beforeActionValidationHook(rt engineplayer.HookRuntime, ctx engineplayer.TimingHookContext) engineplayer.TimingHookResult {
	player := ctx.Player
	modifier := ctx.ValidationModifier
	choiceRt := ctx.ChoiceRuntime

	if player == nil || modifier == nil || choiceRt == nil {
		return engineplayer.TimingHookResult{}
	}

	TauntValidationPolicy(choiceRt, player, modifier, rt.HasPlayableAttackCard)
	return engineplayer.TimingHookResult{Handled: true}
}
