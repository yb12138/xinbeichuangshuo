// gameflow: 蝶舞者 Timing Hook 实现。

package butterfly_dancer

import (
	"fmt"

	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

type butterflyDamageHookRuntime struct {
	player.HookRuntime
}

func (r butterflyDamageHookRuntime) GetPlayerOrder() []string {
	if r.HookRuntime == nil {
		return nil
	}
	return r.HookRuntime.GetPlayerOrder()
}

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

// damageResponseHook 处理蝶舞者时间轴⑤：治疗抵御后、承受伤害前的法术伤害响应。
func damageResponseHook(rt player.HookRuntime, ctx player.TimingHookContext) player.TimingHookResult {
	pd := ctx.PendingDamage
	if pd == nil || pd.Damage <= 0 {
		return player.TimingHookResult{}
	}
	sourceName := ""
	if src := rt.GetPlayer(pd.SourceID); src != nil {
		sourceName = src.Name
	}
	targetName := ""
	if tgt := rt.GetPlayer(pd.TargetID); tgt != nil {
		targetName = tgt.Name
	}
	damageAmount := pd.Damage

	if pd.DamageType != model.MagicDamage {
		return player.TimingHookResult{}
	}
	// 毒粉：伤害为1时，蝶舞者可移除1个茧令伤害+1。该检查仅在⑤窗口执行一次，
	// 避免朝圣在⑥把2点伤害降为1点后又倒回触发毒粉。
	if pd.Damage == 1 && !pd.HasCheck(model.PendingDamageCheckBeforeApplyPoison) {
		pd.SetCheck(model.PendingDamageCheckBeforeApplyPoison, true)
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
					"source_name":    sourceName,
					"target_id":      pd.TargetID,
					"target_name":    targetName,
					"damage_index":   0,
					"damage_amount":  damageAmount,
					"cocoon_indices": indices,
				},
			})
			rt.Log(fmt.Sprintf("%s 的 [毒粉] 可触发：是否移除1个茧令该次法术伤害+1", user.Name))
			return player.TimingHookResult{Interrupted: true}
		}
	}

	if queueButterflyMirrorResponse(butterflyDamageHookRuntime{HookRuntime: rt}, pd, sourceName, targetName) {
		return player.TimingHookResult{Interrupted: true}
	}

	markButterflyMagicResponseWindowClosed(pd)
	return player.TimingHookResult{}
}

func pilgrimageBeforeApplyHook(rt player.HookRuntime, ctx player.TimingHookContext) player.TimingHookResult {
	pd := ctx.PendingDamage
	if pd == nil || pd.Damage <= 0 {
		return player.TimingHookResult{}
	}
	return triggerPilgrimageBeforeApply(rt, pd, ctx.TargetID)
}

func triggerPilgrimageBeforeApply(rt player.HookRuntime, pd *model.PendingDamage, targetID string) player.TimingHookResult {
	// 朝圣：⑥ 承受伤害时，移除1个茧抵御1点伤害。
	if pd.HasCheck(model.PendingDamageCheckBeforeApplyDefend) {
		return player.TimingHookResult{}
	}
	pd.SetCheck(model.PendingDamageCheckBeforeApplyDefend, true)

	target := rt.GetPlayer(targetID)
	if target != nil && rt.IsCharacter(target, "butterfly_dancer") && CocoonCount(target) > 0 {
		indices := CocoonFieldIndices(target)
		if len(indices) > 0 {
			rt.PushInterrupt(&model.Interrupt{
				Type:     model.InterruptChoice,
				PlayerID: target.ID,
				Context: map[string]interface{}{
					"choice_type":    "bt_pilgrimage_confirm",
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
	return player.TimingHookResult{}
}
