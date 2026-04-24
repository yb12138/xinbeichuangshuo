// gameflow: 格斗家策略 Hook 声明式注册。

package fighter

import (
	engineplayer "starcup-engine/internal/engine/player"
)

// PolicySpecs 导出格斗家策略声明。
func PolicySpecs() []engineplayer.PolicySpec {
	return []engineplayer.PolicySpec{
		// 行动选项策略（百式幻龙拳）
		{Type: engineplayer.PolicyBeforeActionOption, Priority: 300, Hook: beforeActionOptionHook},
		// 行动验证策略（百式幻龙拳）
		{Type: engineplayer.PolicyBeforeActionValidation, Priority: 300, Hook: beforeActionValidationHook},
		// 响应技能规范化策略（百式幻龙拳形态）
		{Type: engineplayer.PolicyResponseSkillNormalize, Priority: 100, Hook: responseSkillNormalizeHook},
	}
}

// beforeActionOptionHook 行动选项策略。
func beforeActionOptionHook(host engineplayer.PolicyHost, ctx engineplayer.PolicyHookContext) engineplayer.PolicyHookResult {
	player := ctx.Player
	modifier := ctx.OptionModifier
	rt := ctx.ChoiceRuntime

	if player == nil || modifier == nil || rt == nil {
		return engineplayer.PolicyHookResult{}
	}

	HundredDragonOptionPolicy(rt, player, modifier)
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

	HundredDragonValidationPolicy(rt, player, modifier)
	return engineplayer.PolicyHookResult{Handled: true}
}

// responseSkillNormalizeHook 响应技能规范化策略。
func responseSkillNormalizeHook(host engineplayer.PolicyHost, ctx engineplayer.PolicyHookContext) engineplayer.PolicyHookResult {
	skillIDs := ctx.SkillIDs
	userCtx := ctx.UserCtx

	if len(skillIDs) == 0 || userCtx == nil {
		return engineplayer.PolicyHookResult{SkillIDs: skillIDs}
	}

	normalized := host.ApplyFighterResponseSkillNormalize(skillIDs, userCtx)
	return engineplayer.PolicyHookResult{Handled: true, SkillIDs: normalized}
}
