// gameflow: 仲裁者策略 Hook 声明式注册。

package arbiter

import (
	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// beforeActionOptionHook 行动选项策略。
func beforeActionOptionHook(rt engineplayer.HookRuntime, ctx engineplayer.TimingHookContext) engineplayer.TimingHookResult {
	player := ctx.Player
	modifier := ctx.OptionModifier
	crt := ctx.ChoiceRuntime

	if player == nil || modifier == nil || crt == nil {
		return engineplayer.TimingHookResult{}
	}

	ForcedDoomsdayOptionPolicy(crt, player, modifier)
	return engineplayer.TimingHookResult{Handled: true}
}

// beforeActionValidationHook 行动验证策略。
func beforeActionValidationHook(rt engineplayer.HookRuntime, ctx engineplayer.TimingHookContext) engineplayer.TimingHookResult {
	player := ctx.Player
	modifier := ctx.ValidationModifier
	crt := ctx.ChoiceRuntime

	if player == nil || modifier == nil || crt == nil {
		return engineplayer.TimingHookResult{}
	}

	ForcedDoomsdayValidationPolicy(crt, player, modifier)
	return engineplayer.TimingHookResult{Handled: true}
}

// skillPostCleanupHook 技能后置清理策略。
func skillPostCleanupHook(rt engineplayer.HookRuntime, ctx engineplayer.TimingHookContext) engineplayer.TimingHookResult {
	skillID := ctx.SkillID
	player := ctx.Player

	if skillID != "arbiter_doomsday" || player == nil {
		return engineplayer.TimingHookResult{}
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
	}
	return engineplayer.TimingHookResult{Handled: true}
}
