// gameflow: 吟游诗人 Timing Hook 实现。

package bard

import (
	"fmt"

	engineplayer "starcup-engine/internal/engine/player"
	skills "starcup-engine/internal/engine/skill"
	"starcup-engine/internal/model"
)

// turnStartRousingHook 回合开始时激昂狂想曲触发检查。
func turnStartRousingHook(rt engineplayer.HookRuntime, ctx engineplayer.TimingHookContext) engineplayer.TimingHookResult {
	player := rt.GetPlayer(ctx.SourceID)
	if player == nil || !engineplayer.IsCharacter(player, "bard") {
		return engineplayer.TimingHookResult{}
	}
	if player.TurnState.UsedSkillCounts["bd_rousing_prompted"] > 0 {
		return engineplayer.TimingHookResult{}
	}
	player.TurnState.UsedSkillCounts["bd_rousing_prompted"] = 1

	crt := rt.AsChoiceRuntime()
	if crt == nil {
		return engineplayer.TimingHookResult{}
	}
	if EternalHolderID(crt, player) == "" {
		return engineplayer.TimingHookResult{}
	}

	ctx2 := responseContext(crt, player, "turn_start", model.TurnStageActionStart)
	handler := skills.GetHandler("bd_rousing_rhapsody")
	if handler == nil || !handler.CanUse(ctx2) {
		return engineplayer.TimingHookResult{}
	}

	rt.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptResponseSkill,
		PlayerID: player.ID,
		SkillIDs: []string{"bd_rousing_rhapsody"},
		Context:  ctx2,
	})
	rt.Log(fmt.Sprintf("%s 在回合开始时满足 [激昂狂想曲] 的发动条件", player.Name))
	return engineplayer.TimingHookResult{Interrupted: true}
}

// turnEndVictoryHook 回合结束时胜利交响诗触发检查。
func turnEndVictoryHook(rt engineplayer.HookRuntime, ctx engineplayer.TimingHookContext) engineplayer.TimingHookResult {
	player := rt.GetPlayer(ctx.SourceID)
	if player == nil || !engineplayer.IsCharacter(player, "bard") {
		return engineplayer.TimingHookResult{}
	}
	if player.TurnState.UsedSkillCounts["bd_victory_prompted"] > 0 {
		return engineplayer.TimingHookResult{}
	}
	player.TurnState.UsedSkillCounts["bd_victory_prompted"] = 1

	crt := rt.AsChoiceRuntime()
	if crt == nil {
		return engineplayer.TimingHookResult{}
	}
	if EternalHolderID(crt, player) == "" {
		return engineplayer.TimingHookResult{}
	}

	ctx2 := responseContext(crt, player, "turn_end", model.TurnStageTurnEnd)
	handler := skills.GetHandler("bd_victory_symphony")
	if handler == nil || !handler.CanUse(ctx2) {
		return engineplayer.TimingHookResult{}
	}

	rt.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptResponseSkill,
		PlayerID: player.ID,
		SkillIDs: []string{"bd_victory_symphony"},
		Context:  ctx2,
	})
	rt.Log(fmt.Sprintf("%s 在回合结束时满足 [胜利交响诗] 的发动条件", player.Name))
	return engineplayer.TimingHookResult{Interrupted: true}
}

// postDamageResolvedHook 伤害结算完成后：沉沦协奏曲触发检查。
func postDamageResolvedHook(rt engineplayer.HookRuntime, ctx engineplayer.TimingHookContext) engineplayer.TimingHookResult {
	if ctx.Damage <= 0 || !rt.IsMagicDamageType(ctx.DamageType) {
		return engineplayer.TimingHookResult{}
	}
	source := rt.LookupPlayer(ctx.SourceID)
	target := rt.LookupPlayer(ctx.TargetID)
	if source == nil || target == nil || source.Camp == target.Camp {
		return engineplayer.TimingHookResult{}
	}
	if !engineplayer.IsCharacter(source, "bard") || !source.IsActive {
		return engineplayer.TimingHookResult{}
	}

	rt.RecordMagicDamageTarget(source.ID, target.ID)
	if rt.MagicDamageTargetCount(source.ID) < 2 {
		return engineplayer.TimingHookResult{}
	}
	if InEternalPrisonerForm(source) || source.TurnState.UsedSkillCounts["bd_descent"] > 0 {
		return engineplayer.TimingHookResult{}
	}
	if maxSameElementCount(source) < 2 {
		return engineplayer.TimingHookResult{}
	}
	rt.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: source.ID,
		Context: map[string]interface{}{
			"choice_type": "bd_descent_element",
			"user_id":     source.ID,
		},
	})
	return engineplayer.TimingHookResult{Interrupted: true}
}
