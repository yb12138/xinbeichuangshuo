// gameflow: 蝶舞者伤害传递/蛹相关辅助。

package engine

import (
	"fmt"

	"starcup-engine/internal/model"
)

func (e *GameEngine) maybeButterflyDamageResponses(pd *model.PendingDamage) bool {
	if pd == nil || pd.Damage <= 0 {
		return false
	}
	// 朝圣：承伤者若为蝶舞者，可在每次伤害中询问一次。
	if !pd.HasCheck(model.PendingDamageCheckBeforeApplyDefend) {
		pd.SetCheck(model.PendingDamageCheckBeforeApplyDefend, true)
		target := e.State.Players[pd.TargetID]
		if target != nil && e.isButterflyDancer(target) && butterflyCocoonCount(target) > 0 {
			indices := butterflyCocoonFieldIndices(target)
			if len(indices) > 0 {
				e.PushInterrupt(&model.Interrupt{
					Type:     model.InterruptChoice,
					PlayerID: target.ID,
					Context: map[string]interface{}{
						"choice_type":    "bt_pilgrimage_pick",
						"user_id":        target.ID,
						"source_id":      pd.SourceID,
						"target_id":      pd.TargetID,
						"damage_index":   0,
						"cocoon_indices": indices,
					},
				})
				e.Log(fmt.Sprintf("%s 的 [朝圣] 可触发：是否移除1个茧抵御1点伤害", target.Name))
				return true
			}
		}
	}
	// 毒粉/镜花水月仅作用于法术伤害。
	if pd.DamageType != model.MagicDamage {
		return false
	}

	// 毒粉/镜花水月：按“实际法术伤害”值检查，仅询问一次。
	if pd.HasCheck(model.PendingDamageCheckBeforeApplyResponse) {
		return false
	}
	pd.SetCheck(model.PendingDamageCheckBeforeApplyResponse, true)

	if pd.Damage == 1 {
		for _, pid := range e.State.PlayerOrder {
			user := e.State.Players[pid]
			if user == nil || !e.isButterflyDancer(user) || butterflyCocoonCount(user) <= 0 {
				continue
			}
			indices := butterflyCocoonFieldIndices(user)
			if len(indices) == 0 {
				continue
			}
			e.PushInterrupt(&model.Interrupt{
				Type:     model.InterruptChoice,
				PlayerID: user.ID,
				Context: map[string]interface{}{
					"choice_type":    "bt_poison_pick",
					"user_id":        user.ID,
					"source_id":      pd.SourceID,
					"target_id":      pd.TargetID,
					"damage_index":   0,
					"cocoon_indices": indices,
				},
			})
			e.Log(fmt.Sprintf("%s 的 [毒粉] 可触发：是否移除1个茧令该次法术伤害+1", user.Name))
			return true
		}
		return false
	}

	if pd.Damage == 2 {
		for _, pid := range e.State.PlayerOrder {
			user := e.State.Players[pid]
			if user == nil || !e.isButterflyDancer(user) || butterflyCocoonCount(user) < 2 {
				continue
			}
			defs, labels := butterflyMirrorPairDefs(user)
			if len(defs) == 0 {
				continue
			}
			e.PushInterrupt(&model.Interrupt{
				Type:     model.InterruptChoice,
				PlayerID: user.ID,
				Context: map[string]interface{}{
					"choice_type":  "bt_mirror_pair",
					"user_id":      user.ID,
					"source_id":    pd.SourceID,
					"target_id":    pd.TargetID,
					"damage_index": 0,
					"pair_defs":    defs,
					"pair_labels":  labels,
				},
			})
			e.Log(fmt.Sprintf("%s 的 [镜花水月] 可触发：是否移除2张同系茧改写本次伤害来源", user.Name))
			return true
		}
	}
	return false
}
