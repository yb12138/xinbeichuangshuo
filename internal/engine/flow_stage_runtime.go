// gameflow: TurnStage/Subflow 与阶段切换辅助。

package engine

import (
	"starcup-engine/internal/engine/core/runtimeutil"
	"starcup-engine/internal/model"
)

func (e *GameEngine) RuntimeStateLabel() string {
	if e == nil || e.State == nil {
		return "turn=<nil> combat=<nil> subflow=<nil>"
	}
	return "turn=" + string(e.State.TurnStage) +
		" combat=" + string(e.State.CombatStage) +
		" subflow=" + string(e.State.Subflow)
}

func (e *GameEngine) IsStartupWindow() bool {
	return e != nil && e.State != nil &&
		e.State.Subflow == model.SubflowNone &&
		e.State.CombatStage == model.CombatStageNone &&
		(e.State.TurnStage == model.TurnStageTurnStart || e.State.TurnStage == model.TurnStageActionStart)
}

func (e *GameEngine) IsActionSelectionWindow() bool {
	return e != nil && e.State != nil &&
		e.State.Subflow == model.SubflowNone &&
		e.State.CombatStage == model.CombatStageNone &&
		e.State.TurnStage == model.TurnStageActionExecution &&
		len(e.State.ActionQueue) == 0
}

func (e *GameEngine) needsActionExecutionActionEndCatchup(player *model.Player) bool {
	return e != nil && e.State != nil &&
		player != nil &&
		e.State.Subflow == model.SubflowNone &&
		e.State.CombatStage == model.CombatStageNone &&
		len(e.State.CombatStack) == 0 &&
		e.State.TurnStage == model.TurnStageActionExecution &&
		len(e.State.ActionQueue) == 0 &&
		player.TurnState.LastActionType != ""
}

func (e *GameEngine) IsBeforeActionWindow() bool {
	return e != nil && e.State != nil &&
		e.State.Subflow == model.SubflowNone &&
		e.State.CombatStage == model.CombatStageNone &&
		e.State.TurnStage == model.TurnStageActionExecution &&
		len(e.State.ActionQueue) > 0
}

func (e *GameEngine) IsCombatInteractionWindow() bool {
	return e != nil && e.State != nil &&
		e.State.Subflow == model.SubflowNone &&
		len(e.State.CombatStack) > 0 &&
		(e.State.CombatStage == model.CombatStageDeclare || e.State.CombatStage == model.CombatStageHitCheck)
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

func (e *GameEngine) EnterDiscardSelection() {
	if e == nil || e.State == nil {
		return
	}
	if !e.HasDiscardSelectionInterrupt() {
		e.Log("[Error] EnterDiscardSelection: 缺少与弃牌子流程匹配的 PendingInterrupt")
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

func (e *GameEngine) IsDiscardSelectionActive() bool {
	return e != nil && e.State != nil &&
		e.State.Subflow == model.SubflowDiscardSelection &&
		e.HasDiscardSelectionInterrupt()
}

func IsDiscardSelectionInterrupt(intr *model.Interrupt) bool {
	if intr == nil {
		return false
	}
	if intr.Type == model.InterruptGiveCards {
		return true
	}
	if intr.Type != model.InterruptChoice {
		return false
	}
	data, ok := intr.Context.(map[string]interface{})
	if !ok || data == nil {
		return false
	}
	if runtimeutil.ToBoolContextValue(data["discard_subflow"]) {
		return true
	}
	choiceType, _ := data["choice_type"].(string)
	return IsDiscardChoiceType(choiceType)
}

func (e *GameEngine) HasDiscardSelectionInterrupt() bool {
	if e == nil || e.State == nil || e.State.PendingInterrupt == nil {
		return false
	}
	return IsDiscardSelectionInterrupt(e.State.PendingInterrupt)
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
	if e.RestoreReturnPoint() {
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

func (e *GameEngine) BuildTimedContext(user *model.Player, target *model.Player, timing model.FlowTiming, eventCtx *model.EventContext) *model.Context {
	ctx := e.BuildContext(user, target, timing, eventCtx)
	ctx.Selections["current_turn_stage"] = e.State.TurnStage
	ctx.Selections["current_combat_stage"] = e.State.CombatStage
	ctx.Selections["current_subflow"] = e.State.Subflow
	return ctx
}
