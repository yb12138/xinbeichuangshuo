// gameflow: 格斗家 Timing Hook 实现。

package fighter

import (
	"fmt"

	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// damageCalculateHook 格斗家被动增伤：蓄力一击 + 百式幻龙拳。
func damageCalculateHook(rt engineplayer.HookRuntime, ctx engineplayer.TimingHookContext) engineplayer.TimingHookResult {
	p := rt.GetPlayer(ctx.SourceID)
	if p == nil || !rt.IsCharacter(p, "fighter") {
		return engineplayer.TimingHookResult{}
	}
	delta := 0
	action := model.Action{Type: ctx.ActionType, Card: ctx.Card}
	// 蓄力一击：主动攻击 +1（百式幻龙拳形态下不可叠加）
	if ctx.ActionType == model.ActionAttack && ctx.CounterInitiator == "" && !InHundredDragonForm(p) && rt.ConsumeAttackDamageRuleBonus(p, "fighter_charge_attack_bonus", action) > 0 {
		delta += 1
		p.TurnState.SkillFlowState["fighter_charge_pending"] = 0
		rt.Log(fmt.Sprintf("[Passive] %s 的 [蓄力一击] 生效，本次主动攻击伤害 +1", p.Name))
	}
	// 百式幻龙拳：主动攻击 +2，应战攻击 +1
	if InHundredDragonForm(p) {
		if ctx.ActionType == model.ActionAttack && ctx.CounterInitiator == "" {
			delta += 2
			rt.Log(fmt.Sprintf("[Passive] %s 的 [百式幻龙拳] 生效，本次主动攻击伤害 +2", p.Name))
		} else if ctx.ActionType == model.ActionAttack && ctx.CounterInitiator != "" {
			delta += 1
			rt.Log(fmt.Sprintf("[Passive] %s 的 [百式幻龙拳] 生效，本次应战攻击伤害 +1", p.Name))
		}
	}
	if delta != 0 {
		return engineplayer.TimingHookResult{DamageDelta: delta}
	}
	return engineplayer.TimingHookResult{}
}

// attackStateResetHook resets fighter attack-related state when a new attack is declared.
func attackStateResetHook(rt engineplayer.HookRuntime, ctx engineplayer.TimingHookContext) engineplayer.TimingHookResult {
	player := rt.GetPlayer(ctx.SourceID)
	if player == nil {
		return engineplayer.TimingHookResult{}
	}
	player.TurnState.SkillFlowState["fighter_attack_start_skill_lock"] = 0
	return engineplayer.TimingHookResult{}
}

// attackGatingHook applies fighter qi burst no-counter gating on attack.
func attackGatingHook(rt engineplayer.HookRuntime, ctx engineplayer.TimingHookContext) engineplayer.TimingHookResult {
	p := rt.GetPlayer(ctx.SourceID)
	if p == nil || p.TurnState.SkillFlowState["fighter_qiburst_force_no_counter"] <= 0 {
		return engineplayer.TimingHookResult{}
	}
	if ctx.AttackInfo == nil {
		return engineplayer.TimingHookResult{}
	}
	ctx.AttackInfo.SetInterceptTag(model.CombatInterceptUnrespondable)
	p.TurnState.SkillFlowState["fighter_qiburst_force_no_counter"] = 0
	return engineplayer.TimingHookResult{}
}

// pendingDamageInitHook 格斗家·蓄力一击未命中标记：攻击宣言时若蓄力挂起，
// 在 PendingDamage 上设置 FighterChargeMissArmed 标记。
func pendingDamageInitHook(rt engineplayer.HookRuntime, ctx engineplayer.TimingHookContext) engineplayer.TimingHookResult {
	p := rt.GetPlayer(ctx.SourceID)
	if p == nil || !rt.IsCharacter(p, "fighter") {
		return engineplayer.TimingHookResult{}
	}
	if p.TurnState.SkillFlowState["fighter_charge_pending"] > 0 {
		ctx.PendingDamage.SetCheck(model.PendingDamageCheckFighterChargeMissArmed, true)
	}
	return engineplayer.TimingHookResult{}
}

// turnEndHook 回合结束：百式幻龙拳退场并转正。
func turnEndHook(rt engineplayer.HookRuntime, ctx engineplayer.TimingHookContext) engineplayer.TimingHookResult {
	p := rt.GetPlayer(ctx.SourceID)
	if p == nil || !rt.IsCharacter(p, "fighter") {
		return engineplayer.TimingHookResult{}
	}
	if rt.HasForm(p, model.FormFighterHundredDragon) {
		defer rt.PoseChangeGuard()
		active := rt.HasForm(p, model.FormFighterHundredDragon) || engineplayer.GetSkillFlowState(p, "fighter_hundred_dragon_target_order") > 0
		engineplayer.ClearForm(p, model.FormFighterHundredDragon)
		engineplayer.SetSkillFlowState(p, "fighter_hundred_dragon_target_order", 0)
		if active {
			rt.Log(fmt.Sprintf("%s 的 [百式幻龙拳] 在本行动阶段结束时退场并转正", p.Name))
		}
	}
	return engineplayer.TimingHookResult{}
}
