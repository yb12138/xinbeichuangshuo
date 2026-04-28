// gameflow: 格斗家策略 Hook 声明式注册。

package fighter

import (
	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
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

// responseSkillNormalizeHook 响应技能规范化策略。
// 百式幻龙拳形态下，若蓄力打击和爆裂冲击同时可用，仅保留蓄力打击。
func responseSkillNormalizeHook(rt engineplayer.HookRuntime, ctx engineplayer.TimingHookContext) engineplayer.TimingHookResult {
	skillIDs := ctx.OfferedSkillIDs
	userCtx := ctx.UserCtx

	if len(skillIDs) <= 1 || userCtx == nil || userCtx.User == nil {
		return engineplayer.TimingHookResult{SkillIDs: skillIDs}
	}

	// 条件：必须是攻击宣告阶段、格斗家角色、非反击场景
	if userCtx.Timing != model.TimingOnAttackDeclared ||
		!engineplayer.IsCharacter(userCtx.User, "fighter") ||
		userCtx.EventCtx == nil ||
		userCtx.EventCtx.AttackInfo == nil ||
		userCtx.EventCtx.AttackInfo.CounterInitiator != "" {
		return engineplayer.TimingHookResult{SkillIDs: skillIDs}
	}

	// 检查是否同时拥有蓄力打击和爆裂冲击
	hasCharge := false
	hasBurst := false
	for _, sid := range skillIDs {
		if sid == "fighter_charge_strike" {
			hasCharge = true
		} else if sid == "fighter_burst_crash" {
			hasBurst = true
		}
	}

	if hasCharge && hasBurst {
		return engineplayer.TimingHookResult{Handled: true, SkillIDs: []string{"fighter_charge_strike"}}
	}
	if hasCharge && hasBurst {
		return engineplayer.TimingHookResult{Handled: true, SkillIDs: []string{"fighter_charge_strike"}}
	}
	return engineplayer.TimingHookResult{SkillIDs: skillIDs}
}

// responseSkillAdvanceHook 响应技能推进策略。
// 蓄力一击被跳过时，推进到气绝崩击。
func responseSkillAdvanceHook(rt engineplayer.HookRuntime, ctx engineplayer.TimingHookContext) engineplayer.TimingHookResult {
	skillIDs := ctx.OfferedSkillIDs
	userCtx := ctx.UserCtx

	if len(skillIDs) != 1 || skillIDs[0] != "fighter_charge_strike" {
		return engineplayer.TimingHookResult{}
	}
	if userCtx == nil || userCtx.User == nil {
		return engineplayer.TimingHookResult{}
	}
	if userCtx.Timing != model.TimingOnAttackDeclared ||
		!engineplayer.IsCharacter(userCtx.User, "fighter") ||
		userCtx.EventCtx == nil ||
		userCtx.EventCtx.AttackInfo == nil ||
		userCtx.EventCtx.AttackInfo.CounterInitiator != "" {
		return engineplayer.TimingHookResult{}
	}
	if !rt.IsSkillStillUsable("fighter_burst_crash", userCtx.User, userCtx) {
		return engineplayer.TimingHookResult{}
	}

	return engineplayer.TimingHookResult{
		Handled:  true,
		SkillIDs: []string{"fighter_burst_crash"},
	}
}
