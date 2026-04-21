// gameflow: 红莲骑士 Timing Hook 实现。

package crimson_knight

import (
	"starcup-engine/internal/engine/player"
)

// healResistHook 红莲骑士：仅允许"腥红信仰白名单"中的自伤使用治疗抵御。
// 非自伤 或 未在腥红信仰白名单中的自伤，设置 IgnoreHeal = true。
func healResistHook(rt player.HookRuntime, ctx player.TimingHookContext) player.TimingHookResult {
	if ctx.PendingDamage == nil || ctx.PendingDamage.IgnoreHeal {
		return player.TimingHookResult{}
	}
	target := rt.LookupPlayer(ctx.TargetID)
	if target == nil || !player.IsCharacter(target, "crimson_knight") {
		return player.TimingHookResult{}
	}
	// 自伤 + 腥红信仰白名单：允许治疗抵御
	if target.ID == ctx.PendingDamage.SourceID && ctx.PendingDamage.AllowCrimsonFaithHeal {
		return player.TimingHookResult{}
	}
	// 非自伤 或 未在白名单中：禁止治疗抵御
	ctx.PendingDamage.IgnoreHeal = true
	return player.TimingHookResult{}
}

// turnEndHook 回合结束：热血沸腾形态退场。
func turnEndHook(rt player.HookRuntime, ctx player.TimingHookContext) player.TimingHookResult {
	p := rt.GetPlayer(ctx.SourceID)
	if p == nil || !rt.IsCharacter(p, "crimson_knight") {
		return player.TimingHookResult{}
	}
	rt.ResolveCrimsonKnightHotFormTurnEnd(p)
	return player.TimingHookResult{}
}
