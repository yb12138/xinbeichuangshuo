// gameflow: 兽武士策略 Hook 声明式注册。

package beast_samurai

import (
	engineplayer "starcup-engine/internal/engine/player"
)

// PolicySpecs 导出兽武士策略声明。
func PolicySpecs() []engineplayer.PolicySpec {
	return []engineplayer.PolicySpec{
		// 响应技能增强策略（野兽形态）
		{Type: engineplayer.PolicyResponseSkillAugment, Priority: 100, Hook: responseSkillAugmentHook},
	}
}

// responseSkillAugmentHook 响应技能增强策略。
func responseSkillAugmentHook(host engineplayer.PolicyHost, ctx engineplayer.PolicyHookContext) engineplayer.PolicyHookResult {
	skillIDs := ctx.SkillIDs
	userCtx := ctx.UserCtx

	if len(skillIDs) == 0 || userCtx == nil {
		return engineplayer.PolicyHookResult{SkillIDs: skillIDs}
	}

	augmented := host.ApplyBeastSamuraiResponseSkillAugment(skillIDs, userCtx)
	return engineplayer.PolicyHookResult{Handled: true, SkillIDs: augmented}
}
