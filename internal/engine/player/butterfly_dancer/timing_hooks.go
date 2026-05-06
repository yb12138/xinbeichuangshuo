// gameflow: 蝶舞者 Timing Hook 实现。

package butterfly_dancer

import (
	"fmt"

	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// witherExpiryHook 回合开始前检查凋零效果到期。
func witherExpiryHook(rt player.HookRuntime, ctx player.TimingHookContext) player.TimingHookResult {
	playerObj := rt.GetPlayer(ctx.TargetID)
	if playerObj == nil || !rt.IsCharacter(playerObj, "butterfly_dancer") {
		return player.TimingHookResult{}
	}
	player.EnsurePlayerSkillFlowState(playerObj)
	if playerObj.TurnState.SkillFlowState["bt_wither_active"] <= 0 {
		return player.TimingHookResult{}
	}
	playerObj.TurnState.SkillFlowState["bt_wither_active"] = 0
	rt.Log(fmt.Sprintf("%s 的 [凋零] 效果到期：对方士气下限保护已解除", playerObj.Name))
	return player.TimingHookResult{}
}

// damageBeforeApplyHook 蝶舞者承伤前响应：治疗抵伤处理后、正式扣血前的统一插入点。
func damageBeforeApplyHook(rt player.HookRuntime, ctx player.TimingHookContext) player.TimingHookResult {
	pd := ctx.PendingDamage
	if pd == nil || pd.Damage <= 0 {
		return player.TimingHookResult{}
	}

	// 朝圣：移除1个茧抵御1点伤害
	if !pd.HasCheck(model.PendingDamageCheckBeforeApplyDefend) {
		pd.SetCheck(model.PendingDamageCheckBeforeApplyDefend, true)
		target := rt.GetPlayer(ctx.TargetID)
		if target != nil && rt.IsCharacter(target, "butterfly_dancer") && CocoonCount(target) > 0 {
			indices := CocoonFieldIndices(target)
			if len(indices) > 0 {
				rt.PushInterrupt(&model.Interrupt{
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
				rt.Log(fmt.Sprintf("%s 的 [朝圣] 可触发：是否移除1个茧抵御1点伤害", target.Name))
				return player.TimingHookResult{Interrupted: true}
			}
		}
	}

	// 毒粉/镜花水月：法术伤害响应
	if pd.DamageType != model.MagicDamage {
		return player.TimingHookResult{}
	}
	if pd.HasCheck(model.PendingDamageCheckBeforeApplyResponse) {
		return player.TimingHookResult{}
	}
	pd.SetCheck(model.PendingDamageCheckBeforeApplyResponse, true)

	// 毒粉：伤害为1时，蝶舞者可移除1个茧令伤害+1
	if pd.Damage == 1 {
		for _, pid := range rt.GetPlayerOrder() {
			user := rt.GetPlayer(pid)
			if user == nil || !rt.IsCharacter(user, "butterfly_dancer") || CocoonCount(user) <= 0 {
				continue
			}
			indices := CocoonFieldIndices(user)
			if len(indices) == 0 {
				continue
			}
			rt.PushInterrupt(&model.Interrupt{
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
			rt.Log(fmt.Sprintf("%s 的 [毒粉] 可触发：是否移除1个茧令该次法术伤害+1", user.Name))
			return player.TimingHookResult{Interrupted: true}
		}
		return player.TimingHookResult{}
	}

	// 镜花水月：伤害为2时，蝶舞者可移除2张同系茧改写伤害来源
	if pd.Damage == 2 {
		for _, pid := range rt.GetPlayerOrder() {
			user := rt.GetPlayer(pid)
			if user == nil || !rt.IsCharacter(user, "butterfly_dancer") || CocoonCount(user) < 2 {
				continue
			}
			defs, labels := mirrorPairDefs(user)
			if len(defs) == 0 {
				continue
			}
			rt.PushInterrupt(&model.Interrupt{
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
			rt.Log(fmt.Sprintf("%s 的 [镜花水月] 可触发：是否移除2张同系茧改写本次伤害来源", user.Name))
			return player.TimingHookResult{Interrupted: true}
		}
	}
	return player.TimingHookResult{}
}
