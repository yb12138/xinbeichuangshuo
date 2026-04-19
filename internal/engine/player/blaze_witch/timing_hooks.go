// gameflow: 烈焰魔女 Timing Hook 实现。

package blaze_witch

import (
	"starcup-engine/internal/engine/player"
)

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
