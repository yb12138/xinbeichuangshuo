// gameflow: 月神 Timing Hook 实现。

package moon

import (
	"fmt"

	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// attackGatingHook applies moon goddess next attack no-counter gating.
func attackGatingHook(rt engineplayer.HookRuntime, ctx engineplayer.TimingHookContext) engineplayer.TimingHookResult {
	p := rt.GetPlayer(ctx.SourceID)
	if p == nil || p.TurnState.UsedSkillCounts["mg_next_attack_no_counter"] <= 0 {
		return engineplayer.TimingHookResult{}
	}
	if ctx.AttackInfo == nil {
		return engineplayer.TimingHookResult{}
	}
	ctx.AttackInfo.SetInterceptTag(model.CombatInterceptUnrespondable)
	p.TurnState.UsedSkillCounts["mg_next_attack_no_counter"]--
	if p.TurnState.UsedSkillCounts["mg_next_attack_no_counter"] < 0 {
		p.TurnState.UsedSkillCounts["mg_next_attack_no_counter"] = 0
	}
	return engineplayer.TimingHookResult{}
}

// postDamageResolvedHook 伤害结算完成后：月渎触发检查。
func postDamageResolvedHook(rt engineplayer.HookRuntime, ctx engineplayer.TimingHookContext) engineplayer.TimingHookResult {
	pd := ctx.PendingDamage
	if pd == nil || pd.Damage <= 0 {
		return engineplayer.TimingHookResult{}
	}
	if !rt.IsMagicDamageType(pd.DamageType) {
		return engineplayer.TimingHookResult{}
	}
	source := rt.LookupPlayer(pd.SourceID)
	if source == nil || !engineplayer.IsCharacter(source, "moon_goddess") {
		return engineplayer.TimingHookResult{}
	}
	currentTurnSource := rt.CurrentTurnPlayerID() == source.ID
	if !source.IsActive && !currentTurnSource {
		return engineplayer.TimingHookResult{}
	}
	target := rt.LookupPlayer(pd.TargetID)
	if target == nil || target.Camp == source.Camp {
		return engineplayer.TimingHookResult{}
	}
	ensureTokens(source)
	if source.TurnState.UsedSkillCounts["mg_blasphemy"] > 0 {
		return engineplayer.TimingHookResult{}
	}
	if source.TurnState.SkillFlowState["mg_blasphemy_pending"] > 0 {
		return engineplayer.TimingHookResult{}
	}
	if source.Heal <= 0 {
		return engineplayer.TimingHookResult{}
	}
	rt.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: source.ID,
		Context: map[string]interface{}{
			"choice_type":            "mg_blasphemy_target",
			"user_id":                source.ID,
			"target_ids":             []string{target.ID},
			"source_id":              pd.SourceID,
			"context_pending_damage": pd,
		},
	})
	source.TurnState.SkillFlowState["mg_blasphemy_pending"] = 1
	rt.Log(fmt.Sprintf("%s 的 [月渎] 可触发：请选择目标（或跳过）", source.Name))
	return engineplayer.TimingHookResult{Interrupted: true}
}
