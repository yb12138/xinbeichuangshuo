// gameflow: 魔枪兵（Magic Lancer）延迟星尘结算。
package engine

import (
	"fmt"

	"starcup-engine/internal/model"
)

func (e *GameEngine) resolveMagicLancerStardustAfterSelf(user *model.Player) bool {
	if user == nil || !e.isMagicLancer(user) {
		return false
	}
	if user.TurnState.SkillFlowState == nil || user.TurnState.SkillFlowState["ml_stardust_pending"] <= 0 {
		return false
	}

	// 若还在等待本次自伤导致的爆牌弃牌，则延后到 ConfirmDiscard 再判定。
	if e.pendingDiscardVictimID() == user.ID {
		user.TurnState.SkillFlowState["ml_stardust_wait_discard"] = 1
		return false
	}

	before := user.TurnState.SkillFlowState["ml_stardust_morale_before"]
	current := e.campMorale(user.Camp)
	user.TurnState.SkillFlowState["ml_stardust_pending"] = 0
	user.TurnState.SkillFlowState["ml_stardust_wait_discard"] = 0
	user.TurnState.SkillFlowState["ml_stardust_morale_before"] = 0

	if hasMagicLancerPhantomForm(user) {
		beforePoses := e.snapshotPlayerPoses()
		leaveMagicLancerPhantomForm(user)
		e.Log(fmt.Sprintf("%s 的 [幻影星尘] 结算完成，脱离幻影形态并转正", user.Name))
		e.dispatchOrientationChanges(beforePoses)
	}

	if before > 0 && current < before {
		e.Log(fmt.Sprintf("%s 的 [幻影星尘] 未触发后续伤害：本次自伤导致己方士气下降", user.Name))
		return false
	}

	targetIDs := make([]string, 0, len(e.State.PlayerOrder))
	for _, pid := range e.State.PlayerOrder {
		if p := e.State.Players[pid]; p != nil && p.Camp != user.Camp {
			targetIDs = append(targetIDs, pid)
		}
	}
	lockedOrder := user.TurnState.SkillFlowState["ml_stardust_locked_target_order"]
	user.TurnState.SkillFlowState["ml_stardust_locked_target_order"] = 0
	if len(targetIDs) == 0 {
		return false
	}
	if lockedOrder > 0 && lockedOrder <= len(e.State.PlayerOrder) {
		lockedID := e.State.PlayerOrder[lockedOrder-1]
		for _, tid := range targetIDs {
			if tid != lockedID {
				continue
			}
			e.AddPendingDamage(model.PendingDamage{
				SourceID:   user.ID,
				TargetID:   lockedID,
				Damage:     2,
				DamageType: model.MagicDamage,
			})
			if target := e.State.Players[lockedID]; target != nil {
				e.Log(fmt.Sprintf("%s 的 [幻影星尘] 生效：对 %s 造成2点法术伤害", user.Name, target.Name))
			}
			return false
		}
	}

	e.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: user.ID,
		Context: map[string]interface{}{
			"choice_type": "ml_stardust_target",
			"user_id":     user.ID,
			"target_ids":  targetIDs,
		},
	})
	return true
}
