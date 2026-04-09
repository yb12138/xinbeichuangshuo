package engine

import (
	"fmt"

	"starcup-engine/internal/model"
)

func (e *GameEngine) runtimeStateLabel() string {
	if e == nil || e.State == nil {
		return "turn=<nil> combat=<nil> subflow=<nil>"
	}
	return "turn=" + string(e.State.TurnStage) +
		" combat=" + string(e.State.CombatStage) +
		" subflow=" + string(e.State.Subflow)
}

func (e *GameEngine) isStartupWindow() bool {
	return e != nil && e.State != nil &&
		e.State.Subflow == model.SubflowNone &&
		e.State.CombatStage == model.CombatStageNone &&
		(e.State.TurnStage == model.TurnStageTurnStart || e.State.TurnStage == model.TurnStageActionStart)
}

func (e *GameEngine) isActionSelectionWindow() bool {
	return e != nil && e.State != nil &&
		e.State.Subflow == model.SubflowNone &&
		e.State.CombatStage == model.CombatStageNone &&
		e.State.TurnStage == model.TurnStageActionExecution &&
		len(e.State.ActionQueue) == 0
}

func (e *GameEngine) isBeforeActionWindow() bool {
	return e != nil && e.State != nil &&
		e.State.Subflow == model.SubflowNone &&
		e.State.CombatStage == model.CombatStageNone &&
		e.State.TurnStage == model.TurnStageActionExecution &&
		len(e.State.ActionQueue) > 0
}

func (e *GameEngine) isCombatInteractionWindow() bool {
	return e != nil && e.State != nil &&
		e.State.Subflow == model.SubflowNone &&
		len(e.State.CombatStack) > 0 &&
		(e.State.CombatStage == model.CombatStageDeclare || e.State.CombatStage == model.CombatStageHitCheck)
}

func (e *GameEngine) isTurnEndWindow() bool {
	return e != nil && e.State != nil &&
		e.State.Subflow == model.SubflowNone &&
		e.State.CombatStage == model.CombatStageNone &&
		e.State.TurnStage == model.TurnStageTurnEnd
}

func (e *GameEngine) enterDamageResolution(returnTo interface{}) {
	if e == nil || e.State == nil {
		return
	}
	if returnTo != nil {
		e.setReturnPoint(returnTo)
	}
	if e.State.CombatStage == model.CombatStageNone || e.State.CombatStage == model.CombatStageDeclare || e.State.CombatStage == model.CombatStageHitCheck {
		e.State.CombatStage = model.CombatStageCalcDamage
	}
}

func (e *GameEngine) hasReturnPoint() bool {
	return e != nil && e.State != nil &&
		(e.State.ReturnTurnStage != "" ||
			e.State.ReturnCombatStage != model.CombatStageNone ||
			e.State.ReturnSubflow != model.SubflowNone)
}

func (e *GameEngine) ensureReturnPoint(raw interface{}) bool {
	if e == nil || e.State == nil {
		return false
	}
	if e.hasReturnPoint() {
		return true
	}
	return e.setReturnPoint(raw)
}

func (e *GameEngine) routePendingDamageWithDefaultReturn(defaultReturn interface{}) bool {
	if e == nil || e.State == nil || len(e.State.PendingDamageQueue) == 0 {
		return false
	}
	e.ensureReturnPoint(defaultReturn)
	e.enterDamageResolution(nil)
	return true
}

func (e *GameEngine) routePendingDamageWithReturn(returnTo interface{}) bool {
	if e == nil || e.State == nil || len(e.State.PendingDamageQueue) == 0 {
		return false
	}
	e.setReturnPoint(returnTo)
	e.enterDamageResolution(nil)
	return true
}

func (e *GameEngine) routePendingDamageOr(defaultReturn interface{}, onNoPending func()) bool {
	if e.routePendingDamageWithDefaultReturn(defaultReturn) {
		return true
	}
	if onNoPending != nil {
		onNoPending()
	}
	return false
}

func (e *GameEngine) enterDiscardSelection() {
	if e == nil || e.State == nil {
		return
	}
	if !e.hasDiscardSelectionInterrupt() {
		e.Log("[Error] enterDiscardSelection: 缺少与弃牌子流程匹配的 PendingInterrupt")
		return
	}
	e.State.Subflow = model.SubflowDiscardSelection
}

func (e *GameEngine) enterResponseWindow() {
	if e == nil || e.State == nil {
		return
	}
	e.State.Subflow = model.SubflowResponse
}

func (e *GameEngine) enterActionExecutionStage() {
	if e == nil || e.State == nil {
		return
	}
	e.clearSubflow()
	e.clearCombatStage()
	e.setTurnStage(model.TurnStageActionExecution)
}

func (e *GameEngine) enterActionEndStage() {
	if e == nil || e.State == nil {
		return
	}
	e.clearSubflow()
	e.clearCombatStage()
	e.setTurnStage(model.TurnStageActionEnd)
}

func (e *GameEngine) enterExtraActionStage() {
	if e == nil || e.State == nil {
		return
	}
	e.clearSubflow()
	e.clearCombatStage()
	e.setTurnStage(model.TurnStageExtraAction)
}

func (e *GameEngine) enterTurnEndStage() {
	if e == nil || e.State == nil {
		return
	}
	e.clearSubflow()
	e.clearCombatStage()
	e.setTurnStage(model.TurnStageTurnEnd)
}

func (e *GameEngine) setTurnStage(stage model.TurnStage) {
	if e == nil || e.State == nil {
		return
	}
	e.State.TurnStage = stage
}

func (e *GameEngine) setCombatStage(stage model.CombatStage) {
	if e == nil || e.State == nil {
		return
	}
	e.State.CombatStage = stage
}

func (e *GameEngine) setSubflow(subflow model.Subflow) {
	if e == nil || e.State == nil {
		return
	}
	e.State.Subflow = subflow
}

func (e *GameEngine) clearSubflow() {
	e.setSubflow(model.SubflowNone)
}

func (e *GameEngine) setGameOver(over bool) {
	if e == nil || e.State == nil {
		return
	}
	e.State.GameOver = over
}

func (e *GameEngine) isDamageResolutionActive() bool {
	return e != nil && e.State != nil &&
		(e.State.CombatStage == model.CombatStageCalcDamage ||
			e.State.CombatStage == model.CombatStageHeal ||
			e.State.CombatStage == model.CombatStageApply ||
			e.State.CombatStage == model.CombatStageDraw)
}

func (e *GameEngine) isDiscardSelectionActive() bool {
	return e != nil && e.State != nil &&
		e.State.Subflow == model.SubflowDiscardSelection &&
		e.hasDiscardSelectionInterrupt()
}

func isDiscardSelectionInterruptType(interruptType model.InterruptType) bool {
	return interruptType == model.InterruptDiscard || interruptType == model.InterruptGiveCards
}

func (e *GameEngine) hasDiscardSelectionInterrupt() bool {
	if e == nil || e.State == nil || e.State.PendingInterrupt == nil {
		return false
	}
	return isDiscardSelectionInterruptType(e.State.PendingInterrupt.Type)
}

func (e *GameEngine) isResponseWindowActive() bool {
	return e != nil && e.State != nil && e.State.Subflow == model.SubflowResponse
}

func (e *GameEngine) driveResponseRecoveryPhase() driveOutcome {
	if e == nil || e.State == nil || e.State.PendingInterrupt != nil || e.State.Subflow != model.SubflowResponse {
		return driveUnhandled
	}
	if len(e.State.PendingDamageQueue) > 0 {
		e.enterDamageResolution(nil)
		return driveContinueLoop
	}
	if e.restoreReturnPoint() {
		return driveContinueLoop
	}
	e.clearSubflow()
	if len(e.State.CombatStack) > 0 && e.State.CombatStage == model.CombatStageNone {
		e.setCombatStage(model.CombatStageHitCheck)
	}
	return driveContinueLoop
}

func (e *GameEngine) clearCombatStage() {
	e.setCombatStage(model.CombatStageNone)
}

func (e *GameEngine) defaultTimingForTrigger(trigger model.TriggerType) model.TriggerTiming {
	switch trigger {
	case model.TriggerNone:
		return model.TimingActive
	case model.TriggerOnTurnStart:
		return model.TimingOnTurnStart
	case model.TriggerOnBuffPhase:
		return model.TimingOnBeforeAction
	case model.TriggerOnAttackStart:
		return model.TimingOnAttackDeclared
	case model.TriggerOnAttackHit, model.TriggerOnAttackMiss:
		return model.TimingOnHitCheck
	case model.TriggerModifyDamage:
		return model.TimingOnDamageCalculated
	case model.TriggerOnDamageTaken:
		return model.TimingOnDamageTaken
	case model.TriggerOnPhaseEnd:
		return model.TimingOnActionEnd
	case model.TriggerOnCardUsed, model.TriggerOnCardRevealed:
		return model.TimingOnCardPlayedOrRevealed
	case model.TriggerOnBuffAdded, model.TriggerOnBuffRemoved:
		return model.TimingOnFieldMarkChanged
	case model.TriggerBeforeDraw:
		return model.TimingBeforeCardDrawn
	case model.TriggerAfterDraw:
		return model.TimingOnCardDrawn
	case model.TriggerBeforeMoraleLoss:
		return model.TimingBeforeMoraleLoss
	case model.TriggerOnOrientationChanged:
		return model.TimingOnOrientationChanged
	default:
		panic(fmt.Sprintf("unmapped trigger timing: %d", trigger))
	}
}

func (e *GameEngine) buildTimedContext(user *model.Player, target *model.Player, trigger model.TriggerType, timing model.TriggerTiming, eventCtx *model.EventContext) *model.Context {
	ctx := e.buildContext(user, target, trigger, eventCtx)
	ctx.Timing = timing
	ctx.Selections["current_turn_stage"] = e.State.TurnStage
	ctx.Selections["current_combat_stage"] = e.State.CombatStage
	ctx.Selections["current_subflow"] = e.State.Subflow
	return ctx
}
