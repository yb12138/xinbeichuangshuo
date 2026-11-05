// gameflow: 魔枪 Timing Hook 实现。

package magic_lancer

import (
	"fmt"

	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// damageCalculateHook 魔枪被动增伤：暗之解放 +充盈。
func damageCalculateHook(rt player.HookRuntime, ctx player.TimingHookContext) player.TimingHookResult {
	p := rt.GetPlayer(ctx.SourceID)
	if p == nil || !rt.IsCharacter(p, "magic_lancer") {
		return player.TimingHookResult{}
	}
	if ctx.ActionType != model.ActionAttack || ctx.CounterInitiator != "" {
		return player.TimingHookResult{}
	}
	action := model.Action{Type: ctx.ActionType, Card: ctx.Card}
	delta := 0
	if darkBonus := player.GetSkillFlowState(p, darkReleaseArmedKey); darkBonus > 0 {
		player.SetSkillFlowState(p, darkReleaseArmedKey, 0)
		delta += darkBonus
		rt.Log(fmt.Sprintf("[Passive] %s 的 [暗之解放] 生效，本次主动攻击伤害 +1", p.Name))
	}
	if additional := rt.ConsumeAttackDamageRuleBonus(p, "ml_fullness_next_attack_bonus", action); additional > 0 {
		delta += additional
		rt.Log(fmt.Sprintf("[Passive] %s 的 [充盈] 生效，本次主动攻击伤害 +%d", p.Name, additional))
	}
	if delta != 0 {
		return player.TimingHookResult{DamageDelta: delta}
	}
	return player.TimingHookResult{}
}

// darkReleaseArmedKey 暗之解放 +1 已被本回合第一次主动攻击领取，等待该次攻击的伤害计算。
const darkReleaseArmedKey = "ml_dark_release_armed_bonus"

// darkReleaseAttackDeclareHook 主动攻击宣言：暗之解放的 +1 只属于进入幻影形态那个回合的第一次主动攻击。
// 宣言时就消耗规则修正，不论之后命中与否；上一次攻击未用掉的 +1 在下一次宣言时作废。
func darkReleaseAttackDeclareHook(rt player.HookRuntime, ctx player.TimingHookContext) player.TimingHookResult {
	p := rt.GetPlayer(ctx.SourceID)
	if p == nil || !rt.IsCharacter(p, "magic_lancer") {
		return player.TimingHookResult{}
	}
	player.SetSkillFlowState(p, darkReleaseArmedKey, 0)
	bonus := rt.ConsumeAttackDamageRuleBonus(p, "ml_dark_release_next_attack_bonus", model.Action{Type: model.ActionAttack})
	if bonus > 0 {
		player.SetSkillFlowState(p, darkReleaseArmedKey, bonus)
	}
	return player.TimingHookResult{}
}

// darkReleaseAttackMissHook 第一次主动攻击未命中：已领取的 +1 随之作废。
func darkReleaseAttackMissHook(rt player.HookRuntime, ctx player.TimingHookContext) player.TimingHookResult {
	p := rt.GetPlayer(ctx.SourceID)
	if p == nil || !rt.IsCharacter(p, "magic_lancer") || ctx.IsCounter {
		return player.TimingHookResult{}
	}
	player.SetSkillFlowState(p, darkReleaseArmedKey, 0)
	return player.TimingHookResult{}
}

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
		// 自伤摸牌导致爆牌：在引擎推出的弃牌中断上挂续接标记，弃牌（含士气结算）完成后
		// 由 FlowContinuationAfterDiscard -> handleMagicLancerAfterDiscard 继续结算幻影星尘。
		attachStardustDiscardContinuation(rt.PendingInterrupt(), source.ID)
		return player.TimingHookResult{}
	}
	before := source.TurnState.SkillFlowState["ml_stardust_morale_before"]
	var current int
	if source.Camp == model.RedCamp {
		current = rt.GetRedMorale()
	} else {
		current = rt.GetBlueMorale()
	}
	source.TurnState.SkillFlowState["ml_stardust_pending"] = 0
	source.TurnState.SkillFlowState["ml_stardust_wait_discard"] = 0
	source.TurnState.SkillFlowState["ml_stardust_morale_before"] = 0
	// Leave phantom form if applicable
	if player.HasForm(source, model.FormMagicLancerPhantom) {
		defer rt.PoseChangeGuard()
		player.ClearForm(source, model.FormMagicLancerPhantom)
		rt.Log(fmt.Sprintf("%s 的 [幻影星尘] 结算完成，脱离幻影形态并转正", source.Name))
	}
	if before > 0 && current < before {
		rt.Log(fmt.Sprintf("%s 的 [幻影星尘] 未触发后续伤害：本次自伤导致己方士气下降", source.Name))
		return player.TimingHookResult{}
	}
	// Find enemy targets
	targetIDs := make([]string, 0)
	for _, pid := range rt.GetPlayerOrder() {
		if p := rt.GetPlayers()[pid]; p != nil && p.Camp != source.Camp {
			targetIDs = append(targetIDs, pid)
		}
	}
	lockedOrder := source.TurnState.SkillFlowState["ml_stardust_locked_target_order"]
	source.TurnState.SkillFlowState["ml_stardust_locked_target_order"] = 0
	if len(targetIDs) == 0 {
		return player.TimingHookResult{}
	}
	if lockedOrder > 0 && lockedOrder <= len(rt.GetPlayerOrder()) {
		lockedID := rt.GetPlayerOrder()[lockedOrder-1]
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

// defendValidationHook 黑暗束缚：防御验证 Hook（通过 TimingHookSpec 调用）。
func defendValidationHook(rt player.HookRuntime, ctx player.TimingHookContext) player.TimingHookResult {
	if ctx.Player == nil || !rt.IsCharacter(ctx.Player, "magic_lancer") {
		return player.TimingHookResult{}
	}
	return player.TimingHookResult{
		ValidationError: fmt.Errorf("魔枪受[黑暗束缚]影响，不能使用法术牌防御"),
	}
}

// magicMissileDefendHook 黑暗束缚：魔弹链防御验证 Hook。
func magicMissileDefendHook(rt player.HookRuntime, ctx player.TimingHookContext) player.TimingHookResult {
	if ctx.Player == nil || !rt.IsCharacter(ctx.Player, "magic_lancer") {
		return player.TimingHookResult{}
	}
	return player.TimingHookResult{
		ValidationError: fmt.Errorf("魔枪受[黑暗束缚]影响，不能使用法术牌防御"),
	}
}

// magicMissileCounterHook 黑暗束缚：魔弹链反击验证 Hook。
func magicMissileCounterHook(rt player.HookRuntime, ctx player.TimingHookContext) player.TimingHookResult {
	if ctx.Player == nil || !rt.IsCharacter(ctx.Player, "magic_lancer") {
		return player.TimingHookResult{}
	}
	return player.TimingHookResult{
		ValidationError: fmt.Errorf("魔枪受[黑暗束缚]影响，不能使用法术牌"),
	}
}

// attachStardustDiscardContinuation 给待处理的弃牌中断挂上魔枪的续接字段。
func attachStardustDiscardContinuation(intr *model.Interrupt, playerID string) {
	if intr == nil {
		return
	}
	data, ok := intr.Context.(map[string]interface{})
	if !ok || data == nil {
		return
	}
	if existing, _ := data["flow_continuation_role_id"].(string); existing != "" {
		return
	}
	data["flow_continuation_role_id"] = "magic_lancer"
	data["flow_continuation_player_id"] = playerID
	data["flow_continuation_skill_id"] = "ml_phantom_stardust"
}
