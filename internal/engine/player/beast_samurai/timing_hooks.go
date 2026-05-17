// gameflow: 兽灵武士 Timing Hook 实现。

package beast_samurai

import (
	"fmt"
	"strings"

	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// damageCalculateHook 兽灵武士·御魂流居合形态被动增伤：横置目标主动攻击伤害 +1。
func damageCalculateHook(rt player.HookRuntime, ctx player.TimingHookContext) player.TimingHookResult {
	p := rt.GetPlayer(ctx.SourceID)
	if p == nil || !rt.IsCharacter(p, "beast_samurai") {
		return player.TimingHookResult{}
	}
	if ctx.ActionType != model.ActionAttack || ctx.CounterInitiator != "" {
		return player.TimingHookResult{}
	}
	if !InIaijutsuForm(p) {
		return player.TimingHookResult{}
	}
	target := rt.GetPlayers()[ctx.TargetID]
	if target == nil {
		return player.TimingHookResult{}
	}
	if rt.GetPlayerOrientation(target) != model.OrientationTapped {
		return player.TimingHookResult{}
	}
	rt.Log(fmt.Sprintf("[Passive] %s 的 [御魂流居合形态·横置目标增伤] 生效，本次主动攻击伤害 +1", p.Name))
	return player.TimingHookResult{DamageDelta: 1}
}

// attackGatingHook applies beast samurai one-strike gating: ignore holy shield + force hit on technique faction.
func attackGatingHook(rt player.HookRuntime, ctx player.TimingHookContext) player.TimingHookResult {
	p := rt.GetPlayer(ctx.SourceID)
	if p == nil || p.TurnState.UsedSkillCounts["bs_one_strike_armed"] <= 0 {
		return player.TimingHookResult{}
	}
	if ctx.AttackInfo == nil {
		return player.TimingHookResult{}
	}
	p.TurnState.UsedSkillCounts["bs_one_strike_armed"] = 0
	ctx.AttackInfo.SetInterceptTag(model.CombatInterceptIgnoreHolyShield)
	ctx.AttackInfo.SetInterceptTag(model.CombatInterceptIgnoreTargetHoly)
	if ctx.Card != nil && strings.TrimSpace(ctx.Card.Faction) == "技" {
		ctx.AttackInfo.SetInterceptTag(model.CombatInterceptForceHit)
	}
	rt.Log(fmt.Sprintf("%s 的 [一击无念·下次攻击劫持] 生效：本次主动攻击无视圣盾、不可用圣光抵挡%s", p.Name, func() string {
		if ctx.Card != nil && strings.TrimSpace(ctx.Card.Faction) == "技" {
			return "，且技命格攻击强制命中"
		}
		return ""
	}()))
	return player.TimingHookResult{}
}

func postAttackHitHook(rt player.HookRuntime, ctx player.TimingHookContext) player.TimingHookResult {
	p := rt.GetPlayer(ctx.SourceID)
	if p == nil {
		return player.TimingHookResult{}
	}
	if ctx.IsCounter || InIaijutsuForm(p) {
		return player.TimingHookResult{}
	}
	before := BeastSoul(p)
	after := AddBeastSoul(p, 1, false)
	if after > before {
		rt.Log(fmt.Sprintf("%s 的 [兽魂意念] 生效：普通形态主动攻击命中，兽魂+1（当前%d）", p.Name, after))
	}
	return player.TimingHookResult{}
}

// attackStateResetHook resets beast samurai attack-related tokens when a new attack is declared.
func attackStateResetHook(rt player.HookRuntime, ctx player.TimingHookContext) player.TimingHookResult {
	p := rt.GetPlayer(ctx.SourceID)
	if p == nil {
		return player.TimingHookResult{}
	}
	ClearAttackTokens(p)
	return player.TimingHookResult{}
}

// attackMissHook 攻击未命中后清除攻击令牌。
func attackMissHook(rt player.HookRuntime, ctx player.TimingHookContext) player.TimingHookResult {
	p := rt.GetPlayer(ctx.SourceID)
	if p == nil {
		return player.TimingHookResult{}
	}
	ClearAttackTokens(p)
	return player.TimingHookResult{}
}

// postDamageResolvedHook 伤害结算完成后：清除攻击指示物 + 居合形态退场。
func postDamageResolvedHook(rt player.HookRuntime, ctx player.TimingHookContext) player.TimingHookResult {
	source := rt.GetPlayer(ctx.SourceID)
	if source == nil {
		return player.TimingHookResult{}
	}
	// Clear attack tokens for attack damage
	if strings.EqualFold(string(ctx.DamageType), string(model.AttackDamage)) {
		ClearAttackTokens(source)
	}
	// Leave iaijutsu form on damage dealt
	if ctx.Damage > 0 && InIaijutsuForm(source) {
		defer rt.PoseChangeGuard()
		if LeaveIaijutsuForm(source) {
			rt.Log(fmt.Sprintf("%s 的 [御魂流居合形态·造成伤害退场] 生效：转正并脱离御魂流居合形态", source.Name))
		}
	}
	return player.TimingHookResult{}
}

// turnEndHook 回合结束：居合形态扣魂 + 兽魂归零退场 + 状态清理。
func turnEndHook(rt player.HookRuntime, ctx player.TimingHookContext) player.TimingHookResult {
	p := rt.GetPlayer(ctx.SourceID)
	if p == nil {
		return player.TimingHookResult{}
	}
	player.EnsurePlayerTokensMap(p)
	if InIaijutsuForm(p) && BeastSoul(p) >= 1 {
		current := BeastSoul(p)
		consume := 1
		if consume > current {
			consume = current
		}
		AddBeastSoul(p, -consume, true)
		AddZanshin(p, consume)
		rt.Log(fmt.Sprintf("%s 的 [御魂流居合形态·回合结束扣魂] 生效：兽魂-1，残心+1", p.Name))
	}
	if InIaijutsuForm(p) && BeastSoul(p) == 0 {
		defer rt.PoseChangeGuard()
		if player.ClearForm(p, model.FormBeastSamuraiIaijutsu) {
			rt.Log(fmt.Sprintf("%s 的 [御魂流居合形态·兽魂归零退场] 生效：转正并脱离御魂流居合形态", p.Name))
		}
	}
	ClearAttackTokens(p)
	return player.TimingHookResult{}
}

// turnEndFinalHook 在额外行动耗尽后的回合结束点清理一击无念挂载（避免在 PendingActions 额外攻击发放前清掉 armed）。
func turnEndFinalHook(rt player.HookRuntime, ctx player.TimingHookContext) player.TimingHookResult {
	p := rt.GetPlayer(ctx.SourceID)
	if p == nil {
		return player.TimingHookResult{}
	}
	if p.TurnState.UsedSkillCounts["bs_one_strike_armed"] > 0 {
		p.TurnState.UsedSkillCounts["bs_one_strike_armed"] = 0
		rt.Log(fmt.Sprintf("%s 的 [一击无念·挂载过期] 生效：本回合未消耗的下次攻击劫持已移除", p.Name))
	}
	return player.TimingHookResult{}
}
