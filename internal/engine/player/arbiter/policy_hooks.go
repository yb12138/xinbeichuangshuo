// gameflow: 仲裁者策略 Hook 声明式注册。

package arbiter

import (
	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// PolicySpecs 导出仲裁者策略声明。
func PolicySpecs() []engineplayer.PolicySpec {
	return []engineplayer.PolicySpec{
		// 行动选项策略（强制末日审判）
		{Type: engineplayer.PolicyBeforeActionOption, Priority: 100, Hook: beforeActionOptionHook},
		// 行动验证策略（强制末日审判）
		{Type: engineplayer.PolicyBeforeActionValidation, Priority: 100, Hook: beforeActionValidationHook},
		// 技能后置清理策略（末日审判清理）
		{Type: engineplayer.PolicySkillPost, Priority: 100, Hook: skillPostCleanupHook},
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

	ForcedDoomsdayOptionPolicy(rt, player, modifier)
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

	ForcedDoomsdayValidationPolicy(rt, player, modifier)
	return engineplayer.PolicyHookResult{Handled: true}
}

// skillPostCleanupHook 技能后置清理策略。
func skillPostCleanupHook(host engineplayer.PolicyHost, ctx engineplayer.PolicyHookContext) engineplayer.PolicyHookResult {
	skillID := ctx.SkillID
	player := ctx.Player
	rt := ctx.ChoiceRuntime

	if skillID != "arbiter_doomsday" || player == nil {
		return engineplayer.PolicyHookResult{}
	}

	// 清理末日审判状态
	if player.Tokens != nil {
		player.Tokens["arbiter_forced_doomsday"] = 0
	}
	// 清理强制末日审判 pending 状态
	if player.TurnState.UsedSkillCounts["arbiter_forced_doomsday_pending"] > 0 {
		player.TurnState.UsedSkillCounts["arbiter_forced_doomsday_pending"] = 0
		player.TurnState.UsedSkillCounts["arbiter_forced_doomsday_done_turn"] = 1
		// 清理挑衅效果（强制末日审判不受挑衅约束）
		player.TurnState.UsedSkillCounts["hero_taunt_active_turn"] = 0
		// 直接从场牌中移除 EffectHeroTaunt
		newField := make([]*model.FieldCard, 0, len(player.Field))
		for _, fc := range player.Field {
			if fc == nil || fc.Mode != model.FieldEffect || fc.Effect != model.EffectHeroTaunt {
				newField = append(newField, fc)
			}
		}
		player.Field = newField
		_ = rt // rt 不需要使用（已经在上面直接清理了状态）
	}
	return engineplayer.PolicyHookResult{Handled: true}
}
