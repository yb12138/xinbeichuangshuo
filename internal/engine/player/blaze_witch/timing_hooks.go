// gameflow: 烈焰魔女 Timing Hook 实现。

package blaze_witch

import (
	"fmt"

	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// beforeActionFlameReleaseHook 行动前脱离烈焰形态。
func beforeActionFlameReleaseHook(rt player.HookRuntime, ctx player.TimingHookContext) player.TimingHookResult {
	p := rt.GetPlayer(ctx.SourceID)
	if p == nil || !player.IsCharacter(p, "blaze_witch") {
		return player.TimingHookResult{}
	}
	if rt.HasUsedActionSkill(p) || !rt.HasForm(p, model.FormBlazeWitchFlame) {
		return player.TimingHookResult{}
	}
	player.EnsurePlayerSkillFlowState(p)
	if p.TurnState.SkillFlowState["bw_flame_release_pending"] <= 0 {
		return player.TimingHookResult{}
	}
	beforePoses := rt.SnapshotPlayerPoses()
	rt.ClearForm(p, model.FormBlazeWitchFlame)
	p.TurnState.SkillFlowState["bw_flame_release_pending"] = 0
	rt.Log(fmt.Sprintf("%s 脱离烈焰形态并转正", p.Name))
	rt.DispatchOrientationChanges(beforePoses)
	return player.TimingHookResult{}
}

// postDamageResolvedHook 痛苦链接弃牌判定。
func postDamageResolvedHook(rt player.HookRuntime, ctx player.TimingHookContext) player.TimingHookResult {
	source := rt.GetPlayer(ctx.SourceID)
	if source == nil || source.TurnState.SkillFlowState == nil {
		return player.TimingHookResult{}
	}
	if source.TurnState.SkillFlowState["bw_pain_link_pending_discard"] <= 0 {
		return player.TimingHookResult{}
	}
	if source.TurnState.SkillFlowState["bw_pain_link_pending_hits"] > 0 {
		source.TurnState.SkillFlowState["bw_pain_link_pending_hits"]--
	}
	if source.TurnState.SkillFlowState["bw_pain_link_pending_hits"] > 0 {
		return player.TimingHookResult{}
	}
	source.TurnState.SkillFlowState["bw_pain_link_pending_hits"] = 0
	source.TurnState.SkillFlowState["bw_pain_link_pending_discard"] = 0
	if len(source.Hand) > 3 {
		rt.PushDiscardChoiceInterrupt(source.ID, map[string]interface{}{
			"discard_count": len(source.Hand) - 3,
			"stay_in_turn":  true,
			"prompt":        "【痛苦链接】请弃牌至3张手牌：",
		})
		return player.TimingHookResult{Interrupted: true}
	}
	return player.TimingHookResult{}
}

// afterApplyHook 伤害应用后重置替身人偶和魔力反转锁定。
func afterApplyHook(rt player.HookRuntime, ctx player.TimingHookContext) player.TimingHookResult {
	target := rt.LookupPlayer(ctx.TargetID)
	if target == nil {
		return player.TimingHookResult{}
	}
	player.EnsurePlayerSkillFlowState(target)
	target.TurnState.SkillFlowState["bw_substitute_lock"] = 0
	target.TurnState.SkillFlowState["bw_mana_inversion_lock"] = 0
	return player.TimingHookResult{}
}
