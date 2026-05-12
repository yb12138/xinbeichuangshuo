// gameflow: 吟游诗人 Timing Hook 实现。

package bard

import (
	"fmt"

	engineplayer "starcup-engine/internal/engine/player"
	skills "starcup-engine/internal/engine/skill"
	"starcup-engine/internal/model"
)

// findBardPlayer 找到吟游诗人玩家。
func findBardPlayer(rt engineplayer.HookRuntime) *model.Player {
	for _, p := range rt.AllPlayers() {
		if p != nil && engineplayer.IsCharacter(p, "bard") {
			return p
		}
	}
	return nil
}

// turnStartRousingHook 回合开始时激昂狂想曲触发检查。
// 当永恒乐章持有者的回合开始时，向吟游诗人推送响应询问。
// 优先级 200：低于血祭司流血（100），确保流血效果先结算。
func turnStartRousingHook(rt engineplayer.HookRuntime, ctx engineplayer.TimingHookContext) engineplayer.TimingHookResult {
	// 找到吟游诗人
	bard := findBardPlayer(rt)
	if bard == nil {
		return engineplayer.TimingHookResult{}
	}

	// 获取当前回合玩家（永恒乐章可能的持有者）
	currentPlayer := rt.GetPlayer(ctx.SourceID)
	if currentPlayer == nil {
		return engineplayer.TimingHookResult{}
	}

	// 检查永恒乐章是否在当前回合玩家身上
	crt := rt.AsChoiceRuntime()
	if crt == nil {
		return engineplayer.TimingHookResult{}
	}
	holderID := EternalHolderID(crt, bard)
	if holderID == "" || holderID != currentPlayer.ID {
		return engineplayer.TimingHookResult{}
	}

	// 如果有待结算的延迟伤害（如血祭司的流血效果），先让伤害结算完成，
	// 下次回到 TurnStart 阶段时再触发激昂狂想曲。
	// 注意：此处不可设置 bd_rousing_prompted 标记，否则下次无法再触发。
	if len(rt.GetPendingDamageQueue()) > 0 {
		return engineplayer.TimingHookResult{}
	}

	// 防止重复触发
	if currentPlayer.TurnState.UsedSkillCounts["bd_rousing_prompted"] > 0 {
		return engineplayer.TimingHookResult{}
	}
	currentPlayer.TurnState.UsedSkillCounts["bd_rousing_prompted"] = 1

	// 检查技能发动条件：ctx.User 设为吟游诗人（技能主人），而非持有者
	ctx2 := responseContext(crt, bard, "turn_start", model.TurnStageActionStart)
	handler := skills.GetHandler("bd_rousing_rhapsody")
	if handler == nil || !handler.CanUse(ctx2) {
		return engineplayer.TimingHookResult{}
	}

	// 向吟游诗人推送响应询问（PlayerID 是吟游诗人）
	rt.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptResponseSkill,
		PlayerID: bard.ID,
		SkillIDs: []string{"bd_rousing_rhapsody"},
		Context:  ctx2,
	})
	rt.Log(fmt.Sprintf("%s 持有永恒乐章，回合开始时满足 [激昂狂想曲] 的发动条件，询问 %s 是否发动", currentPlayer.Name, bard.Name))
	return engineplayer.TimingHookResult{Interrupted: true}
}

// turnEndVictoryHook 回合结束时胜利交响诗触发检查。
// 当永恒乐章持有者的回合结束时，向吟游诗人推送响应询问。
func turnEndVictoryHook(rt engineplayer.HookRuntime, ctx engineplayer.TimingHookContext) engineplayer.TimingHookResult {
	// 找到吟游诗人
	bard := findBardPlayer(rt)
	if bard == nil {
		return engineplayer.TimingHookResult{}
	}

	// 获取当前回合玩家（永恒乐章可能的持有者）
	currentPlayer := rt.GetPlayer(ctx.SourceID)
	if currentPlayer == nil {
		return engineplayer.TimingHookResult{}
	}

	// 检查永恒乐章是否在当前回合玩家身上
	crt := rt.AsChoiceRuntime()
	if crt == nil {
		return engineplayer.TimingHookResult{}
	}
	holderID := EternalHolderID(crt, bard)
	if holderID == "" || holderID != currentPlayer.ID {
		return engineplayer.TimingHookResult{}
	}

	// 防止重复触发
	if currentPlayer.TurnState.UsedSkillCounts["bd_victory_prompted"] > 0 {
		return engineplayer.TimingHookResult{}
	}
	currentPlayer.TurnState.UsedSkillCounts["bd_victory_prompted"] = 1

	// 检查技能发动条件：ctx.User 设为吟游诗人（技能主人），而非持有者
	ctx2 := responseContext(crt, bard, "turn_end", model.TurnStageTurnEnd)
	handler := skills.GetHandler("bd_victory_symphony")
	if handler == nil || !handler.CanUse(ctx2) {
		return engineplayer.TimingHookResult{}
	}

	// 向吟游诗人推送响应询问（PlayerID 是吟游诗人）
	rt.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptResponseSkill,
		PlayerID: bard.ID,
		SkillIDs: []string{"bd_victory_symphony"},
		Context:  ctx2,
	})
	rt.Log(fmt.Sprintf("%s 持有永恒乐章，回合结束时满足 [胜利交响诗] 的发动条件，询问 %s 是否发动", currentPlayer.Name, bard.Name))
	return engineplayer.TimingHookResult{Interrupted: true}
}

// postDamageResolvedHook 伤害结算完成后：沉沦协奏曲触发检查。
func postDamageResolvedHook(rt engineplayer.HookRuntime, ctx engineplayer.TimingHookContext) engineplayer.TimingHookResult {
	if ctx.Damage <= 0 || !rt.IsMagicDamageType(ctx.DamageType) {
		return engineplayer.TimingHookResult{}
	}
	source := rt.GetPlayers()[ctx.SourceID]
	target := rt.GetPlayers()[ctx.TargetID]
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
	if engineplayer.MaxSameElementCount(source) < 2 {
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
