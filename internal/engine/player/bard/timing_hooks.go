// gameflow: 吟游诗人 Timing Hook 实现。

package bard

import (
	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// postDamageResolvedHook 伤害结算完成后：沉沦协奏曲触发检查。
func postDamageResolvedHook(rt engineplayer.HookRuntime, ctx engineplayer.TimingHookContext) engineplayer.TimingHookResult {
	if ctx.Damage <= 0 || !rt.IsMagicDamageType(ctx.DamageType) {
		return engineplayer.TimingHookResult{}
	}
	source := rt.LookupPlayer(ctx.SourceID)
	target := rt.LookupPlayer(ctx.TargetID)
	if source == nil || target == nil || source.Camp == target.Camp {
		return engineplayer.TimingHookResult{}
	}
	if !engineplayer.IsCharacter(source, "bard") || !source.IsActive {
		return engineplayer.TimingHookResult{}
	}

	rt.RecordMagicDamageTarget(source.ID, target.ID)
	if rt.MagicDamageTargetCount(source.ID) < 2 {
		return engineplayer.TimingHookResult{}
	}
	if InEternalPrisonerForm(source) || source.TurnState.UsedSkillCounts["bd_descent"] > 0 {
		return engineplayer.TimingHookResult{}
	}
	if maxSameElementCount(source) < 2 {
		return engineplayer.TimingHookResult{}
	}
	rt.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: source.ID,
		Context: map[string]interface{}{
			"choice_type": "bd_descent_element",
			"user_id":     source.ID,
		},
	})
	return engineplayer.TimingHookResult{Interrupted: true}
}
