// gameflow: 红莲骑士 Timing Hook 实现。

package crimson_knight

import (
	"fmt"

	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// healResistHook 红莲骑士：仅允许"腥红信仰白名单"中的自伤使用治疗抵御。
// 非自伤 或 未在腥红信仰白名单中的自伤，设置 IgnoreHeal = true。
func healResistHook(rt engineplayer.HookRuntime, ctx engineplayer.TimingHookContext) engineplayer.TimingHookResult {
	if ctx.PendingDamage == nil || ctx.PendingDamage.IgnoreHeal {
		return engineplayer.TimingHookResult{}
	}
	target := rt.GetPlayers()[ctx.TargetID]
	if target == nil || !engineplayer.IsCharacter(target, "crimson_knight") {
		return engineplayer.TimingHookResult{}
	}
	// 自伤 + 腥红信仰白名单：允许治疗抵御
	if target.ID == ctx.PendingDamage.SourceID && ctx.PendingDamage.AllowCrimsonFaithHeal {
		return engineplayer.TimingHookResult{}
	}
	// 非自伤 或 未在白名单中：禁止治疗抵御
	ctx.PendingDamage.IgnoreHeal = true
	return engineplayer.TimingHookResult{}
}

// turnEndHook 回合结束：热血沸腾形态退场。
func turnEndHook(rt engineplayer.HookRuntime, ctx engineplayer.TimingHookContext) engineplayer.TimingHookResult {
	p := rt.GetPlayer(ctx.SourceID)
	if p == nil || !rt.IsCharacter(p, "crimson_knight") {
		return engineplayer.TimingHookResult{}
	}
	if !rt.HasForm(p, model.FormCrimsonKnightHotBlooded) {
		return engineplayer.TimingHookResult{}
	}
	defer rt.PoseChangeGuard()
	engineplayer.ClearForm(p, model.FormCrimsonKnightHotBlooded)
	rt.Heal(p.ID, 2)
	rt.Log(fmt.Sprintf("%s 回合结束脱离 [热血沸腾形态]，获得2点治疗", p.Name))
	return engineplayer.TimingHookResult{}
}

// moraleLossHook 红莲骑士：伤害导致士气下降时，强制进入热血沸腾形态。
func moraleLossHook(rt engineplayer.HookRuntime, ctx engineplayer.TimingHookContext) engineplayer.TimingHookResult {
	if !ctx.FromDamageDraw || ctx.MoraleLoss <= 0 {
		return engineplayer.TimingHookResult{}
	}
	victim := rt.GetPlayer(ctx.TargetID)
	if victim == nil || !engineplayer.IsCharacter(victim, "crimson_knight") {
		return engineplayer.TimingHookResult{}
	}
	if engineplayer.HasForm(victim, model.FormCrimsonKnightHotBlooded) {
		return engineplayer.TimingHookResult{}
	}
	defer rt.PoseChangeGuard()
	engineplayer.SetForm(victim, model.FormCrimsonKnightHotBlooded)
	rt.Log(fmt.Sprintf("%s 的 [热血沸腾] 触发，进入热血沸腾形态", victim.Name))
	return engineplayer.TimingHookResult{}
}
