// gameflow: 格斗家策略 Hook 声明式注册。

package fighter

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

	HundredDragonOptionPolicy(choiceRt, player, modifier)
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

	HundredDragonValidationPolicy(choiceRt, player, modifier)
	return engineplayer.TimingHookResult{Handled: true}
}

// 注：原 responseSkillNormalizeHook / responseSkillAdvanceHook 把蓄力一击与气绝崩击
// 拆成两段响应（先弹蓄力 → 跳过 → 再弹气绝），与产品需求「单次面板三选一（蓄力 /
// 气绝 / 跳过）」不一致。已删除，统一交由 buildResponseSkillPrompt 一次性渲染多技能选择。
