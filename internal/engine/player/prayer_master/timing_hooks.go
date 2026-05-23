// gameflow: 祈祷师 Timing Hook 实现。

package prayer_master

import (
	"fmt"

	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// attackRuneGainHook 祈祷形态下主动攻击宣告时：+2 祈祷符文（上限3）。
// 这是祈祷师的被动规则，不作为角色技能表中的可触发技能注册。
func attackRuneGainHook(rt player.HookRuntime, ctx player.TimingHookContext) player.TimingHookResult {
	attacker := ctx.Attacker
	if attacker == nil && ctx.SourceID != "" {
		attacker = rt.GetPlayer(ctx.SourceID)
	}
	if attacker == nil || !rt.IsCharacter(attacker, "prayer_master") {
		return player.TimingHookResult{}
	}
	if !rt.HasForm(attacker, model.FormPrayerMasterPrayer) {
		return player.TimingHookResult{}
	}
	if ctx.CounterInitiator != "" {
		return player.TimingHookResult{}
	}
	if ctx.AttackInfo != nil && ctx.AttackInfo.CounterInitiator != "" {
		return player.TimingHookResult{}
	}
	if ctx.UserCtx != nil && ctx.UserCtx.EventCtx != nil && ctx.UserCtx.EventCtx.AttackInfo != nil &&
		ctx.UserCtx.EventCtx.AttackInfo.CounterInitiator != "" {
		return player.TimingHookResult{}
	}

	v := player.AddToken(attacker, "prayer_rune", 2, 3)
	rt.Log(fmt.Sprintf("%s 的 [祈祷符文] 触发，祈祷符文=%d", attacker.Name, v))
	return player.TimingHookResult{}
}

// postActionEndHook 攻击/法术行动结束后：迅捷赐福后续选择。
func postActionEndHook(rt player.HookRuntime, ctx player.TimingHookContext) player.TimingHookResult {
	p := rt.GetPlayer(ctx.SourceID)
	if p == nil {
		return player.TimingHookResult{}
	}
	if player.GetFieldEffectCard(p, model.EffectSwiftBlessing) == nil {
		return player.TimingHookResult{}
	}
	rt.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: p.ID,
		Context: map[string]interface{}{
			"choice_type": "prayer_swift_blessing_response",
			"user_id":     p.ID,
			"action_type": string(ctx.ActionType),
		},
	})
	return player.TimingHookResult{Interrupted: true}
}

// postAttackHitHook 攻击命中后：力量赐福后续选择。
func postAttackHitHook(rt player.HookRuntime, ctx player.TimingHookContext) player.TimingHookResult {
	p := rt.GetPlayer(ctx.SourceID)
	if p == nil {
		return player.TimingHookResult{}
	}
	if player.GetFieldEffectCard(p, model.EffectPowerBlessing) == nil {
		return player.TimingHookResult{}
	}
	rt.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: p.ID,
		Context: map[string]interface{}{
			"choice_type": "prayer_power_blessing_response",
			"user_id":     p.ID,
			"source_id":   ctx.SourceID,
			"target_id":   ctx.TargetID,
		},
	})
	return player.TimingHookResult{Interrupted: true}
}
