// gameflow: 兽武士策略 Hook 声明式注册。

package beast_samurai

import (
	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// responseSkillAugmentHook 响应技能增强策略。
// 野兽形态下残心满时，追加一击无心到响应技能列表。
func responseSkillAugmentHook(rt engineplayer.HookRuntime, ctx engineplayer.TimingHookContext) engineplayer.TimingHookResult {
	skillIDs := ctx.OfferedSkillIDs
	userCtx := ctx.UserCtx

	if len(skillIDs) == 0 || userCtx == nil {
		return engineplayer.TimingHookResult{SkillIDs: skillIDs}
	}

	player := ctx.Player
	if player == nil {
		return engineplayer.TimingHookResult{SkillIDs: skillIDs}
	}

	// 条件1：必须是在行动结束后的攻击事件
	if userCtx.Timing != model.TimingOnActionEnd || userCtx.EventCtx == nil || userCtx.EventCtx.ActionType != model.ActionAttack {
		return engineplayer.TimingHookResult{SkillIDs: skillIDs}
	}

	// 条件2：必须是兽武士角色
	if !engineplayer.IsCharacter(player, "beast_samurai") {
		return engineplayer.TimingHookResult{SkillIDs: skillIDs}
	}

	// 条件3：技能列表中未包含一击无心（已存在则跳过）
	const targetSkillID = "bs_one_strike_no_thought"
	for _, sid := range skillIDs {
		if sid == targetSkillID {
			return engineplayer.TimingHookResult{SkillIDs: skillIDs}
		}
	}

	// 条件4：残心指示物必须达到上限
	if Zanshin(player) < ZanshinCap {
		return engineplayer.TimingHookResult{SkillIDs: skillIDs}
	}

	// 条件5：角色技能定义中存在一击无心
	var skillDef *model.SkillDefinition
	if player.Character != nil {
		for i := range player.Character.Skills {
			if player.Character.Skills[i].ID == targetSkillID {
				skillDef = &player.Character.Skills[i]
				break
			}
		}
	}
	if skillDef == nil {
		return engineplayer.TimingHookResult{SkillIDs: skillIDs}
	}

	// 条件6：技能仍然可用
	if !rt.IsSkillStillUsable(targetSkillID, player, userCtx) {
		return engineplayer.TimingHookResult{SkillIDs: skillIDs}
	}

	augmented := append(skillIDs, targetSkillID)
	return engineplayer.TimingHookResult{Handled: true, SkillIDs: augmented}
}
