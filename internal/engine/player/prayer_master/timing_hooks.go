// gameflow: 祈祷师 Timing Hook 实现。

package prayer_master

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

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
			"choice_type": "prayer_swift_blessing_followup",
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
			"choice_type": "prayer_power_blessing_followup",
			"user_id":     p.ID,
			"source_id":   ctx.SourceID,
			"target_id":   ctx.TargetID,
		},
	})
	return player.TimingHookResult{Interrupted: true}
}
