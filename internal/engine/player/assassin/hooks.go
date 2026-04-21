// gameflow: 暗杀者 Timing Hook 实现。

package assassin

import (
	"fmt"

	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// damageCalculateHook 暗杀者·潜行被动增伤：潜行形态下主动攻击额外 +剩余能量。
func damageCalculateHook(rt engineplayer.HookRuntime, ctx engineplayer.TimingHookContext) engineplayer.TimingHookResult {
	p := rt.GetPlayer(ctx.SourceID)
	if p == nil || !rt.IsCharacter(p, "assassin") {
		return engineplayer.TimingHookResult{}
	}
	if ctx.ActionType != model.ActionAttack || ctx.CounterInitiator != "" {
		return engineplayer.TimingHookResult{}
	}
	if !InStealthForm(p) {
		return engineplayer.TimingHookResult{}
	}
	extra := p.Gem + p.Crystal
	if extra <= 0 {
		return engineplayer.TimingHookResult{}
	}
	rt.Log(fmt.Sprintf("[Passive] %s 处于[潜行]，本次主动攻击伤害额外+%d（剩余能量）", p.Name, extra))
	return engineplayer.TimingHookResult{DamageDelta: extra}
}

// attackGatingHook applies assassin stealth no-counter gating on attack.
func attackGatingHook(rt engineplayer.HookRuntime, ctx engineplayer.TimingHookContext) engineplayer.TimingHookResult {
	p := rt.GetPlayer(ctx.SourceID)
	if p == nil || !InStealthForm(p) {
		return engineplayer.TimingHookResult{}
	}
	if ctx.AttackInfo == nil {
		return engineplayer.TimingHookResult{}
	}
	ctx.AttackInfo.SetInterceptTag(model.CombatInterceptUnrespondable)
	rt.Log(fmt.Sprintf("[Skill] %s 处于[潜行]：本次主动攻击无法应战", p.Name))
	return engineplayer.TimingHookResult{}
}

// beforeActionStealthReleaseHook 行动前脱离潜行形态。
func beforeActionStealthReleaseHook(rt engineplayer.HookRuntime, ctx engineplayer.TimingHookContext) engineplayer.TimingHookResult {
	p := rt.GetPlayer(ctx.SourceID)
	if p == nil || !engineplayer.IsCharacter(p, "assassin") {
		return engineplayer.TimingHookResult{}
	}
	if rt.HasUsedActionSkill(p) || !rt.HasForm(p, model.FormAssassinStealth) {
		return engineplayer.TimingHookResult{}
	}
	rt.ReleaseAssassinStealth(p)
	return engineplayer.TimingHookResult{}
}
