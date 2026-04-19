// gameflow: 魔枪 Timing Hook 实现。

package magic_lancer

import (
	"fmt"

	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// postDamageResolvedHook 幻影星尘自伤后结算。
func postDamageResolvedHook(rt player.HookRuntime, ctx player.TimingHookContext) player.TimingHookResult {
	source := rt.GetPlayer(ctx.SourceID)
	if source == nil {
		return player.TimingHookResult{}
	}
	if source.TurnState.SkillFlowState == nil || source.TurnState.SkillFlowState["ml_stardust_pending"] <= 0 {
		return player.TimingHookResult{}
	}
	// Only trigger on self-damage
	if ctx.SourceID != source.ID || ctx.TargetID != source.ID {
		return player.TimingHookResult{}
	}
	// If waiting for pending discard, defer
	if rt.HasPendingDiscardFor(source.ID) {
		source.TurnState.SkillFlowState["ml_stardust_wait_discard"] = 1
		return player.TimingHookResult{}
	}
	before := source.TurnState.SkillFlowState["ml_stardust_morale_before"]
	current := rt.CampMorale(source.Camp)
	source.TurnState.SkillFlowState["ml_stardust_pending"] = 0
	source.TurnState.SkillFlowState["ml_stardust_wait_discard"] = 0
	source.TurnState.SkillFlowState["ml_stardust_morale_before"] = 0
	// Leave phantom form if applicable
	if player.HasForm(source, model.FormMagicLancerPhantom) {
		beforePoses := rt.SnapshotPlayerPoses()
		player.ClearForm(source, model.FormMagicLancerPhantom)
		rt.Log(fmt.Sprintf("%s 的 [幻影星尘] 结算完成，脱离幻影形态并转正", source.Name))
		rt.DispatchOrientationChanges(beforePoses)
	}
	if before > 0 && current < before {
		rt.Log(fmt.Sprintf("%s 的 [幻影星尘] 未触发后续伤害：本次自伤导致己方士气下降", source.Name))
		return player.TimingHookResult{}
	}
	// Find enemy targets
	targetIDs := make([]string, 0)
	for _, pid := range rt.PlayerOrder() {
		if p := rt.GetPlayer(pid); p != nil && p.Camp != source.Camp {
			targetIDs = append(targetIDs, pid)
		}
	}
	lockedOrder := source.TurnState.SkillFlowState["ml_stardust_locked_target_order"]
	source.TurnState.SkillFlowState["ml_stardust_locked_target_order"] = 0
	if len(targetIDs) == 0 {
		return player.TimingHookResult{}
	}
	if lockedOrder > 0 && lockedOrder <= len(rt.PlayerOrder()) {
		lockedID := rt.PlayerOrder()[lockedOrder-1]
		for _, tid := range targetIDs {
			if tid != lockedID {
				continue
			}
			rt.AddPendingDamage(model.PendingDamage{
				SourceID:   source.ID,
				TargetID:   lockedID,
				Damage:     2,
				DamageType: model.MagicDamage,
			})
			if target := rt.GetPlayer(lockedID); target != nil {
				rt.Log(fmt.Sprintf("%s 的 [幻影星尘] 生效：对 %s 造成2点法术伤害", source.Name, target.Name))
			}
			return player.TimingHookResult{}
		}
	}
	rt.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: source.ID,
		Context: map[string]interface{}{
			"choice_type": "ml_stardust_target",
			"user_id":     source.ID,
			"target_ids":  targetIDs,
		},
	})
	return player.TimingHookResult{Interrupted: true}
}
