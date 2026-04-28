package magic_lancer

import (
	"fmt"

	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

func ResolveStardustAfterSelf(rt engineplayer.ChoiceRuntime, user *model.Player) bool {
	if user == nil || !engineplayer.IsCharacter(user, "magic_lancer") {
		return false
	}
	if user.TurnState.SkillFlowState == nil || user.TurnState.SkillFlowState["ml_stardust_pending"] <= 0 {
		return false
	}

	if rt.PendingDiscardVictimID() == user.ID {
		user.TurnState.SkillFlowState["ml_stardust_wait_discard"] = 1
		return false
	}

	before := user.TurnState.SkillFlowState["ml_stardust_morale_before"]
	var current int
	if user.Camp == model.RedCamp {
		current = rt.GetRedMorale()
	} else {
		current = rt.GetBlueMorale()
	}
	user.TurnState.SkillFlowState["ml_stardust_pending"] = 0
	user.TurnState.SkillFlowState["ml_stardust_wait_discard"] = 0
	user.TurnState.SkillFlowState["ml_stardust_morale_before"] = 0

	if engineplayer.HasForm(user, model.FormMagicLancerPhantom) {
		defer rt.PoseChangeGuard()
		engineplayer.ClearForm(user, model.FormMagicLancerPhantom)
		rt.Log(fmt.Sprintf("%s 的 [幻影星尘] 结算完成，脱离幻影形态并转正", user.Name))
	}

	if before > 0 && current < before {
		rt.Log(fmt.Sprintf("%s 的 [幻影星尘] 未触发后续伤害：本次自伤导致己方士气下降", user.Name))
		return false
	}

	targetIDs := make([]string, 0, len(rt.GetPlayerOrder()))
	for _, pid := range rt.GetPlayerOrder() {
		if p := rt.GetPlayers()[pid]; p != nil && p.Camp != user.Camp {
			targetIDs = append(targetIDs, pid)
		}
	}
	lockedOrder := user.TurnState.SkillFlowState["ml_stardust_locked_target_order"]
	user.TurnState.SkillFlowState["ml_stardust_locked_target_order"] = 0
	if len(targetIDs) == 0 {
		return false
	}
	if lockedOrder > 0 && lockedOrder <= len(rt.GetPlayerOrder()) {
		lockedID := rt.GetPlayerOrder()[lockedOrder-1]
		for _, tid := range targetIDs {
			if tid != lockedID {
				continue
			}
			rt.AddPendingDamage(model.PendingDamage{
				SourceID:   user.ID,
				TargetID:   lockedID,
				Damage:     2,
				DamageType: model.MagicDamage,
			})
			if target := rt.GetPlayers()[lockedID]; target != nil {
				rt.Log(fmt.Sprintf("%s 的 [幻影星尘] 生效：对 %s 造成2点法术伤害", user.Name, target.Name))
			}
			return false
		}
	}

	rt.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: user.ID,
		Context: map[string]interface{}{
			"choice_type": "ml_stardust_target",
			"user_id":     user.ID,
			"target_ids":  targetIDs,
		},
	})
	return false
}
