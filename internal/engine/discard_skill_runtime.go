package engine

import (
	"fmt"
	"starcup-engine/internal/engine/core/runtimeutil"

	"starcup-engine/internal/engine/skill"
	"starcup-engine/internal/model"
)

// ConfirmDiscard 确认执行弃牌。
func (e *GameEngine) ConfirmDiscard(playerID string, indices []int) error {
	if e.State.PendingInterrupt == nil || e.State.PendingInterrupt.Type != model.InterruptDiscard {
		return fmt.Errorf("当前没有待处理的弃牌操作")
	}

	data, _ := e.State.PendingInterrupt.Context.(map[string]interface{})
	skillID, hasSkillID := data["skill_id"].(string)

	if handled, err := e.handleBeastSamuraiDiscardInput(playerID, indices); handled || err != nil {
		return err
	}

	if hasSkillID && skillID != "" {
		return e.handleSkillDiscardSelection(playerID, indices, data)
	}

	return e.handleDiscardSelection(playerID, indices, data)
}

func (e *GameEngine) handleSkillDiscardSelection(playerID string, indices []int, data map[string]interface{}) error {
	skillID, _ := data["skill_id"].(string)
	if skillID == "" {
		return fmt.Errorf("技能弃牌上下文缺少 skill_id")
	}
	if _, hasCtx := data["user_ctx"]; !hasCtx {
		return e.handleDeferredSkillDiscardSelection(playerID, skillID, indices, data)
	}
	return e.handleContextSkillDiscardSelection(skillID, indices, data)
}

func (e *GameEngine) handleDeferredSkillDiscardSelection(playerID, skillID string, indices []int, data map[string]interface{}) error {
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

	handler := skills.GetHandler(skillID)
	if handler == nil {
		return fmt.Errorf("技能处理器不存在")
	}

	beforePoses := e.snapshotPlayerPoses()
	if err := handler.Execute(ctx); err != nil {
		return fmt.Errorf("技能执行失败: %v", err)
	}
	e.dispatchOrientationChanges(beforePoses)

	if discardedCards, ok := ctx.Selections["discardedCards"].([]model.Card); ok {
		e.State.DiscardPile = append(e.State.DiscardPile, discardedCards...)
	}
	if ctx.Trigger == model.TriggerBeforeDraw {
		e.resumePendingDraw(ctx)
	}

	if nextSkillIDs, ok := data["remaining_skills"].([]string); ok && len(nextSkillIDs) > 0 {
		e.State.PendingInterrupt.Type = model.InterruptResponseSkill
		e.State.PendingInterrupt.SkillIDs = nextSkillIDs
		e.State.PendingInterrupt.Context = ctx
		e.Log("[System] 弃牌技能执行完毕，你还可以选择发动其他技能")
		e.enterResponseWindow()
		return nil
	}

	e.PopInterrupt()
	if e.State.PendingInterrupt == nil {
		e.resumePhaseAfterSkillDiscardContext(ctx)
	}
	return nil
}

func (e *GameEngine) resumePhaseAfterSkillDiscardContext(ctx *model.Context) bool {
	if ctx == nil || e.State.PendingInterrupt != nil {
		return false
	}
	if ctx.Trigger == model.TriggerBeforeDraw {
		return e.restorePhaseAfterInterruptedDraw(ctx)
	}
	if ctx.Trigger == model.TriggerOnAttackMiss && e.resumePendingAttackMiss(ctx) {
		return true
	}
	if ctx.Trigger == model.TriggerOnTurnStart {
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
