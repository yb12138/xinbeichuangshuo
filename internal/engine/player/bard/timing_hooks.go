// gameflow: 吟游诗人 Timing Hook 实现。

package bard

import (
	"fmt"

	engineplayer "starcup-engine/internal/engine/player"
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
// 当永恒乐章持有者的回合开始时，向持有者推送响应询问（持有者决定是否发动）。
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

	// 检查技能发动条件：持有者是否可以发动
	holder := currentPlayer
	bardID := bard.ID
	// 获取敌人列表
	targetIDs := make([]string, 0)
	for _, p := range rt.GetPlayers() {
		if p != nil && p.Camp != holder.Camp {
			targetIDs = append(targetIDs, p.ID)
		}
	}
	canDamage := len(targetIDs) >= 2
	canDiscard := len(holder.Hand) >= 2
	if !canDamage && !canDiscard {
		return engineplayer.TimingHookResult{}
	}

	// 向永恒乐章持有者推送确认弹窗（是否发动）
	rt.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: currentPlayer.ID,
		Context: map[string]interface{}{
			"choice_type": "bd_rousing_confirm",
			"user_id":     holder.ID,
			"bard_id":     bardID,
			"target_ids":  targetIDs,
		},
	})
	rt.Log(fmt.Sprintf("%s 持有永恒乐章，回合开始时满足 [激昂狂想曲] 的发动条件，询问是否发动", currentPlayer.Name))
	return engineplayer.TimingHookResult{Interrupted: true}
}

// turnEndVictoryHook 回合结束时胜利交响诗触发检查。
// 当永恒乐章持有者的回合结束时，向持有者推送响应询问。
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

	// 检查技能发动条件：胜利交响诗总是可以发动（分支2不需要阵营资源）
	holder := currentPlayer
	bardID := bard.ID

	// 向永恒乐章持有者推送确认弹窗（是否发动）
	rt.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: currentPlayer.ID,
		Context: map[string]interface{}{
			"choice_type": "bd_victory_confirm",
			"user_id":     holder.ID,
			"bard_id":     bardID,
		},
	})
	rt.Log(fmt.Sprintf("%s 持有永恒乐章，回合结束时满足 [胜利交响诗] 的发动条件，询问是否发动", currentPlayer.Name))
	return engineplayer.TimingHookResult{Interrupted: true}
}

// postDamageResolvedHook 伤害结算完成后：记录全队法术伤害目标（供回合末沉沦协奏曲检查）。
// 仅做追踪，不触发技能。触发逻辑在 turnEndDescentHook 中。
func postDamageResolvedHook(rt engineplayer.HookRuntime, ctx engineplayer.TimingHookContext) engineplayer.TimingHookResult {
	if ctx.Damage <= 0 || !rt.IsMagicDamageType(ctx.DamageType) {
		return engineplayer.TimingHookResult{}
	}
	source := rt.GetPlayers()[ctx.SourceID]
	target := rt.GetPlayers()[ctx.TargetID]
	if source == nil || target == nil || source.Camp == target.Camp {
		return engineplayer.TimingHookResult{}
	}

	// 找到吟游诗人，将队友的法术伤害目标记录到诗人名下
	bard := findBardPlayer(rt)
	if bard == nil || bard.Camp != source.Camp {
		return engineplayer.TimingHookResult{}
	}

	rt.RecordMagicDamageTarget(bard.ID, target.ID)
	return engineplayer.TimingHookResult{}
}

// turnEndDescentHook 回合结束时检查沉沦协奏曲触发条件。
// 当全队在本回合对至少2名不同敌方目标造成过法术伤害时触发。
func turnEndDescentHook(rt engineplayer.HookRuntime, ctx engineplayer.TimingHookContext) engineplayer.TimingHookResult {
	bard := findBardPlayer(rt)
	if bard == nil {
		return engineplayer.TimingHookResult{}
	}

	if rt.MagicDamageTargetCount(bard.ID) < 2 {
		return engineplayer.TimingHookResult{}
	}
	if InEternalPrisonerForm(bard) || bard.TurnState.UsedSkillCounts["bd_descent"] > 0 {
		return engineplayer.TimingHookResult{}
	}
	if engineplayer.MaxSameElementCount(bard) < 2 {
		return engineplayer.TimingHookResult{}
	}

	rt.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: bard.ID,
		Context: map[string]interface{}{
			"choice_type": "bd_descent_element",
			"user_id":     bard.ID,
		},
	})
	return engineplayer.TimingHookResult{Interrupted: true}
}
