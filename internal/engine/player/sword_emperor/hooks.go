// gameflow: 剑帝 Timing Hook 实现。

package sword_emperor

import (
	"fmt"
	"strings"

	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// damageCalculateHook 剑帝·恶魔之魂被动增伤：本次主动攻击伤害 +1。
func damageCalculateHook(rt engineplayer.HookRuntime, ctx engineplayer.TimingHookContext) engineplayer.TimingHookResult {
	p := rt.GetPlayer(ctx.SourceID)
	if p == nil || !rt.IsCharacter(p, "sword_emperor") {
		return engineplayer.TimingHookResult{}
	}
	if ctx.ActionType != model.ActionAttack || ctx.CounterInitiator != "" {
		return engineplayer.TimingHookResult{}
	}
	action := model.Action{Type: ctx.ActionType, Card: ctx.Card}
	if rt.ConsumeAttackDamageRuleBonus(p, "se_demon_soul_attack_bonus", action) <= 0 {
		return engineplayer.TimingHookResult{}
	}
	rt.Log(fmt.Sprintf("[Passive] %s 的 [恶魔之魂] 生效：本次主动攻击伤害 +1", p.Name))
	return engineplayer.TimingHookResult{DamageDelta: 1}
}

// attackStateResetHook resets sword emperor attack-related flags when a new attack is declared.
func attackStateResetHook(rt engineplayer.HookRuntime, ctx engineplayer.TimingHookContext) engineplayer.TimingHookResult {
	player := rt.GetPlayer(ctx.SourceID)
	if player == nil {
		return engineplayer.TimingHookResult{}
	}
	player.TurnState.UsedSkillCounts["se_guard_disabled_current_attack"] = 0
	player.TurnState.UsedSkillCounts["se_angel_soul_armed"] = 0
	player.TurnState.UsedSkillCounts["se_demon_soul_armed"] = 0
	return engineplayer.TimingHookResult{}
}

// damageAfterTakenHook 剑帝命中后置：承伤触发后、治疗抵伤前执行命中分支。
// 使用 AttackPostHitEffectsDone 标记确保同一次伤害只处理一次。
func damageAfterTakenHook(rt engineplayer.HookRuntime, ctx engineplayer.TimingHookContext) engineplayer.TimingHookResult {
	pd := ctx.PendingDamage
	if pd == nil || pd.AttackPostHitEffectsDone || pd.HasCheck(model.PendingDamageCheckAttackMissResolved) || !strings.EqualFold(string(pd.DamageType), string(model.AttackDamage)) {
		return engineplayer.TimingHookResult{}
	}
	attacker := rt.GetPlayer(ctx.SourceID)
	if attacker == nil || !rt.IsCharacter(attacker, "sword_emperor") {
		return engineplayer.TimingHookResult{}
	}
	// 天使之魂命中分支：治疗+2
	if attacker.TurnState.UsedSkillCounts["se_angel_soul_armed"] > 0 {
		rt.Heal(attacker.ID, 2)
		rt.Log(fmt.Sprintf("%s 的 [天使之魂] 命中分支生效：治疗+2", attacker.Name))
	}
	ClearCombatTokens(attacker)
	pd.AttackPostHitEffectsDone = true
	return engineplayer.TimingHookResult{}
}
