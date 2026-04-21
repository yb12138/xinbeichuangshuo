// gameflow: 剑帝 Timing Hook 实现。

package sword_emperor

import (
	"fmt"

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
