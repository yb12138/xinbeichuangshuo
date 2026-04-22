// gameflow: 勇者 Timing Hook 实现。

package hero

import (
	"fmt"

	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// damageCalculateHook 勇者·怒吼被动增伤：本次主动攻击伤害 +2。
func damageCalculateHook(rt player.HookRuntime, ctx player.TimingHookContext) player.TimingHookResult {
	p := rt.GetPlayer(ctx.SourceID)
	if p == nil || !rt.IsCharacter(p, "hero") {
		return player.TimingHookResult{}
	}
	if ctx.ActionType != model.ActionAttack || ctx.CounterInitiator != "" {
		return player.TimingHookResult{}
	}
	action := model.Action{Type: ctx.ActionType, Card: ctx.Card}
	if rt.ConsumeAttackDamageRuleBonus(p, "hero_roar_attack_bonus", action) <= 0 {
		return player.TimingHookResult{}
	}
	p.TurnState.UsedSkillCounts["hero_roar_active"] = 0
	rt.Log(fmt.Sprintf("[Passive] %s 的 [怒吼] 生效，本次主动攻击伤害 +2", p.Name))
	return player.TimingHookResult{DamageDelta: 2}
}

// attackGatingHook applies hero calm force no-counter gating on attack.
func attackGatingHook(rt player.HookRuntime, ctx player.TimingHookContext) player.TimingHookResult {
	p := rt.GetPlayer(ctx.SourceID)
	if p == nil || p.TurnState.UsedSkillCounts["hero_calm_force_no_counter"] <= 0 {
		return player.TimingHookResult{}
	}
	if ctx.AttackInfo == nil {
		return player.TimingHookResult{}
	}
	ctx.AttackInfo.SetInterceptTag(model.CombatInterceptUnrespondable)
	p.TurnState.UsedSkillCounts["hero_calm_force_no_counter"] = 0
	return player.TimingHookResult{}
}

// postActionEndHook 攻击行动结束后：明镜止水水晶+1。
func postActionEndHook(rt player.HookRuntime, ctx player.TimingHookContext) player.TimingHookResult {
	p := rt.GetPlayer(ctx.SourceID)
	if p == nil {
		return player.TimingHookResult{}
	}
	if ctx.ActionType != model.ActionAttack {
		return player.TimingHookResult{}
	}
	player.EnsurePlayerSkillFlowState(p)
	if p.TurnState.SkillFlowState["hero_calm_end_crystal_pending"] <= 0 {
		return player.TimingHookResult{}
	}
	p.TurnState.SkillFlowState["hero_calm_end_crystal_pending"]--
	if p.TurnState.SkillFlowState["hero_calm_end_crystal_pending"] < 0 {
		p.TurnState.SkillFlowState["hero_calm_end_crystal_pending"] = 0
	}
	capV := rt.GetPlayerEnergyCap(p)
	if p.Gem+p.Crystal < capV {
		p.Crystal++
		rt.Log(fmt.Sprintf("%s 的 [明镜止水] 结算：水晶+1", p.Name))
	} else {
		rt.Log(fmt.Sprintf("%s 的 [明镜止水] 结算：能量已满，水晶未增加", p.Name))
	}
	return player.TimingHookResult{}
}

// pendingDamageInitHook 勇者·怒吼未命中标记：攻击宣言时若怒吼已发动，
// 在 PendingDamage 上设置 HeroRoarMissArmed 标记。
func pendingDamageInitHook(rt player.HookRuntime, ctx player.TimingHookContext) player.TimingHookResult {
	p := rt.GetPlayer(ctx.SourceID)
	if p == nil || !rt.IsCharacter(p, "hero") {
		return player.TimingHookResult{}
	}
	if p.TurnState.UsedSkillCounts["hero_roar_active"] > 0 {
		ctx.PendingDamage.SetCheck(model.PendingDamageCheckHeroRoarMissArmed, true)
	}
	return player.TimingHookResult{}
}

// turnStartExhaustionReleaseHook 回合开始前检查精疲力竭结束。
func turnStartExhaustionReleaseHook(rt player.HookRuntime, ctx player.TimingHookContext) player.TimingHookResult {
	p := rt.GetPlayer(ctx.SourceID)
	if p == nil || !rt.IsCharacter(p, "hero") || p.TurnState.HasUsedActionSkill || !InExhaustionForm(p) {
		return player.TimingHookResult{}
	}
	player.EnsurePlayerSkillFlowState(p)
	if p.TurnState.SkillFlowState["hero_exhaustion_release_pending"] <= 0 {
		return player.TimingHookResult{}
	}
	beforePoses := rt.SnapshotPlayerPoses()
	LeaveExhaustionForm(p)
	p.TurnState.SkillFlowState["hero_exhaustion_release_pending"] = 0
	rt.Log(fmt.Sprintf("%s 的 [精疲力竭] 结束：转正，手牌上限恢复，并对自己造成3点法术伤害", p.Name))
	rt.DispatchOrientationChanges(beforePoses)
	rt.AddPendingDamage(model.PendingDamage{
		SourceID:   p.ID,
		TargetID:   p.ID,
		Damage:     3,
		DamageType: model.MagicAttack,
	})
	rt.EnterDamageResolution(model.TurnStageTurnStart)
	return player.TimingHookResult{Interrupted: true}
}

// turnStartTauntStartupHook 回合开始时挑衅约束初始化。
func turnStartTauntStartupHook(rt player.HookRuntime, ctx player.TimingHookContext) player.TimingHookResult {
	p := rt.GetPlayer(ctx.SourceID)
	if p == nil {
		return player.TimingHookResult{}
	}
	p.TurnState.UsedSkillCounts["hero_taunt_active_turn"] = 0
	if p.TurnState.UsedSkillCounts["arbiter_forced_doomsday_pending"] > 0 {
		return player.TimingHookResult{}
	}
	tauntCard := player.GetFieldEffectCard(p, model.EffectHeroTaunt)
	if tauntCard == nil {
		ConsumeTauntRestrictionInternal(rt, p)
		return player.TimingHookResult{}
	}
	src := rt.LookupPlayer(tauntCard.SourceID)
	if src == nil || src.Camp == p.Camp {
		ConsumeTauntRestrictionInternal(rt, p)
		return player.TimingHookResult{}
	}
	p.TurnState.UsedSkillCounts["hero_taunt_active_turn"] = 1
	rt.Log(fmt.Sprintf("[Taunt] %s 在本行动阶段受到 %s 的【挑衅】影响：必须且只能主动攻击该勇者，或选择跳过行动并移除此牌", p.Name, model.GetPlayerDisplayName(src)))
	return player.TimingHookResult{}
}

// ConsumeTauntRestrictionInternal 消耗挑衅约束（移除场牌+重置状态）。
func ConsumeTauntRestrictionInternal(rt player.HookRuntime, p *model.Player) {
	if p == nil {
		return
	}
	p.TurnState.UsedSkillCounts["hero_taunt_active_turn"] = 0
	rt.RemoveFieldCard(p.ID, model.EffectHeroTaunt)
}
