// gameflow: 瘟疫法师 Timing Hook 实现。

package plague_mage

import (
	"fmt"

	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// healResistHook 瘟疫法师圣渎：攻击伤害不可用治疗抵挡，法术伤害可以。
func healResistHook(rt player.HookRuntime, ctx player.TimingHookContext) player.TimingHookResult {
	if ctx.PendingDamage == nil || ctx.PendingDamage.IgnoreHeal {
		return player.TimingHookResult{}
	}
	target := rt.GetPlayers()[ctx.TargetID]
	if target == nil || !player.IsCharacter(target, "plague_mage") {
		return player.TimingHookResult{}
	}
	if ctx.PendingDamage.DamageType == model.AttackDamage {
		ctx.PendingDamage.IgnoreHeal = true
	}
	return player.TimingHookResult{}
}

func moraleLossAppliedHook(rt player.HookRuntime, ctx player.TimingHookContext) player.TimingHookResult {
	if ctx.MoraleLoss <= 0 || ctx.SourceSkillID != "plague_outbreak" || ctx.SourceID == "" {
		return player.TimingHookResult{}
	}
	source := rt.GetPlayer(ctx.SourceID)
	if source == nil || !rt.IsCharacter(source, "plague_mage") {
		return player.TimingHookResult{}
	}
	source.TurnState.UsedSkillCounts["plague_outbreak_morale_drop"] = 1
	return player.TimingHookResult{}
}

// turnEndHook 回合结束：瘟疫爆发治疗奖励。
func turnEndHook(rt player.HookRuntime, ctx player.TimingHookContext) player.TimingHookResult {
	p := rt.GetPlayer(ctx.SourceID)
	if p == nil || !rt.IsCharacter(p, "plague_mage") {
		return player.TimingHookResult{}
	}
	if p.TurnState.UsedSkillCounts["plague_outbreak_morale_drop"] <= 0 {
		return player.TimingHookResult{}
	}
	p.TurnState.UsedSkillCounts["plague_outbreak_morale_drop"] = 0
	rt.Heal(p.ID, 1)
	rt.Log(fmt.Sprintf("%s 的 [瘟疫] 回合结束奖励生效：+1治疗", p.Name))
	return player.TimingHookResult{}
}
