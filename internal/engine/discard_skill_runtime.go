package engine

import (
	"fmt"

	"starcup-engine/internal/engine/skills"
	"starcup-engine/internal/model"
)

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
	targetIDs := parseStringSliceContextValue(data["target_ids"])
	resumePoint := data["resume_phase"]

	e.PopInterrupt()
	if e.State.PendingInterrupt != nil {
		return fmt.Errorf("当前仍有其他待处理的中断")
	}
	if !e.applyChoiceResumePoint(resumePoint) {
		e.enterActionExecutionStage()
	}
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
		if len(e.State.PendingDamageQueue) > 0 {
			e.setReturnPoint(model.TurnStageActionStart)
			e.enterDamageResolution(nil)
		} else {
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
