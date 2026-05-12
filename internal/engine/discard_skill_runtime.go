// gameflow: 弃牌技能与弃牌子流程（如展示/封印联动）。

package engine

import (
	"fmt"

	"starcup-engine/internal/engine/core/runtimeutil"
	"starcup-engine/internal/engine/skill"
	"starcup-engine/internal/model"
)

// ConfirmDiscard 确认执行弃牌（外部命令入口，走 ChoiceEngine 消费语义）。
func (e *GameEngine) ConfirmDiscard(playerID string, indices []int) error {
	if e.State.PendingInterrupt == nil {
		return fmt.Errorf("当前没有待处理的弃牌操作")
	}
	if e.State.PendingInterrupt.PlayerID != "" && e.State.PendingInterrupt.PlayerID != playerID {
		return fmt.Errorf("当前不是你的弃牌回合")
	}
	if e.choiceEngine == nil {
		return fmt.Errorf("选择引擎未初始化")
	}
	data, ok := choiceCtxAsAnyMap(e.State.PendingInterrupt.Context)
	if !ok {
		return fmt.Errorf("中断上下文格式错误")
	}
	ct, _ := data["choice_type"].(string)
	if ct == "" {
		ct = choiceTypeSystemDiscardCards
	}
	result, err := e.choiceEngine.HandleMultiSelectResult(playerID, ct, indices, data)
	if err != nil {
		return err
	}
	if result.ConsumedInterrupt {
		e.PopInterrupt()
		if result.AfterConsume != nil {
			result.AfterConsume(&choiceHostBridge{e: e})
		}
	}
	return nil
}

func hasSkillDiscardID(data map[string]interface{}) bool {
	if data == nil {
		return false
	}
	skillID, _ := data["skill_id"].(string)
	return skillID != ""
}

func (e *GameEngine) handleSkillDiscardSelection(playerID string, indices []int, data map[string]interface{}) error {
	skillID, _ := data["skill_id"].(string)
	if skillID == "" {
		return fmt.Errorf("技能弃牌上下文缺少 skill_id")
	}
	if _, hasCtx := data["user_ctx"]; !hasCtx {
		return e.handleSkillDiscardResume(playerID, skillID, indices, data)
	}
	return e.handleContextSkillDiscardSelection(skillID, indices, data)
}

func (e *GameEngine) handleSkillDiscardResume(playerID, skillID string, indices []int, data map[string]interface{}) error {
	targetIDs := runtimeutil.ParseStringSliceContextValue(data["target_ids"])
	resumePoint := data["resume_phase"]

	e.PopInterrupt()
	if e.State.PendingInterrupt != nil {
		return fmt.Errorf("当前仍有其他待处理的中断")
	}
	// 规则：为发动技能而产生的弃牌中断，处理完必须回到技能声明的恢复点后再继续施放技能。
	e.applyChoiceResumePoint(mustChoiceResumePoint(resumePoint, "resume_phase"))
	return e.UseSkill(playerID, skillID, targetIDs, indices)
}

func (e *GameEngine) handleContextSkillDiscardSelection(skillID string, indices []int, data map[string]interface{}) error {
	minSelect, _ := data["min"].(int)
	maxSelect, _ := data["max"].(int)
	if len(indices) < minSelect {
		return fmt.Errorf("至少需要选择 %d 张牌，你选择了 %d 张", minSelect, len(indices))
	}
	if len(indices) > maxSelect {
		return fmt.Errorf("最多只能选择 %d 张牌，你选择了 %d 张", maxSelect, len(indices))
	}

	userCtx, hasCtx := data["user_ctx"]
	if !hasCtx {
		return fmt.Errorf("技能上下文丢失")
	}
	ctx, ok := userCtx.(*model.Context)
	if !ok {
		return fmt.Errorf("技能上下文格式错误")
	}
	if ctx.Selections == nil {
		ctx.Selections = make(map[string]any)
	}
	ctx.Selections["discard_indices"] = indices
	wasBeforeDraw := ctx.BeforeDrawPhase()

	handler := skills.GetHandler(skillID)
	if handler == nil {
		return fmt.Errorf("技能处理器不存在")
	}

	beforePoses := e.SnapshotPlayerPoses()
	if err := handler.Execute(ctx); err != nil {
		return fmt.Errorf("技能执行失败: %v", err)
	}
	e.DispatchOrientationChanges(beforePoses)

	if discardedCards, ok := ctx.Selections["discardedCards"].([]model.Card); ok {
		e.State.DiscardPile = append(e.State.DiscardPile, discardedCards...)
	}
	if wasBeforeDraw {
		e.resumePendingDraw(ctx)
	}

	if nextSkillIDs, ok := data["remaining_skills"].([]string); ok && len(nextSkillIDs) > 0 {
		playerID := e.State.PendingInterrupt.PlayerID
		e.PopInterrupt()
		e.PushInterrupt(&model.Interrupt{
			Type:     model.InterruptResponseSkill,
			PlayerID: playerID,
			SkillIDs: nextSkillIDs,
			Context:  ctx,
		})
		e.Log("[System] 弃牌技能执行完毕，你还可以选择发动其他技能")
		return nil
	}

	e.PopInterrupt()
	if e.State.PendingInterrupt == nil {
		if missileInterrupt, ok := ctx.Selections["magic_missile_interrupt"].(*model.Interrupt); ok && missileInterrupt != nil {
			if e.resumeMagicMissileAfterResponseSkill(ctx, missileInterrupt) {
				return nil
			}
			e.PushInterrupt(missileInterrupt)
			return nil
		}
		if wasBeforeDraw {
			e.restorePhaseAfterInterruptedDraw(ctx)
			return nil
		}
		e.ResumePhaseAfterSkillDiscardContext(ctx)
	}
	return nil
}

func (e *GameEngine) ResumePhaseAfterSkillDiscardContext(ctx *model.Context) bool {
	if ctx == nil || e.State.PendingInterrupt != nil {
		return false
	}
	if ctx.BeforeDrawPhase() {
		return e.restorePhaseAfterInterruptedDraw(ctx)
	}
	if ctx.ResumeActionEndPhase() {
		// ActionEnd 响应中的弃牌交互完成后，避免 LastActionType 残留触发同一轮 ActionEnd 重入。
		if ctx.User != nil {
			ctx.User.TurnState.LastActionType = ""
			ctx.User.TurnState.LastActionCard = nil
		}
		if point, ok := choiceResumePointValue(ctx.Selections["response_resume_phase"]); ok {
			if e.routePendingDamageWithReturn(point) {
				return true
			}
			e.applyChoiceResumePoint(point)
			return true
		}
		e.enterExtraActionStage()
		return true
	}
	if ctx.ResumeAttackMissPhase() && e.resumePendingAttackMiss(ctx) {
		return true
	}
	if ctx.TurnStartOrStartupWindow() {
		// 启动技能（回合开始触发）中的弃牌后续：应继续当前回合流程。
		e.clearSubflow()
		e.clearCombatStage()
		if !e.routePendingDamageWithReturn(model.TurnStageActionStart) {
			e.setTurnStage(model.TurnStageActionStart)
		}
		return true
	}

	if len(e.State.ActionStack) > 0 {
		e.enterResponseWindow()
	} else if len(e.State.ActionQueue) > 0 {
		e.enterActionExecutionStage()
	} else {
		e.enterTurnEndStage()
	}
	return true
}
