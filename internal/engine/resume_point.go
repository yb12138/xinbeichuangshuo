package engine

import (
	"fmt"

	"starcup-engine/internal/model"
)

func normalizeChoiceResumePoint(raw interface{}) string {
	return model.NormalizeResumePoint(raw)
}

func parseChoiceResumeTurnStage(raw interface{}) model.TurnStage {
	return model.ParseResumePointTurnStage(raw)
}

func parseChoiceResumeCombatStage(raw interface{}) model.CombatStage {
	return model.ParseResumePointCombatStage(raw)
}

func parseChoiceResumeSubflow(raw interface{}) model.Subflow {
	return model.ParseResumePointSubflow(raw)
}

func (e *GameEngine) currentTurnPlayer() *model.Player {
	if e == nil || e.State == nil || len(e.State.PlayerOrder) == 0 {
		return nil
	}
	if e.State.CurrentTurn < 0 || e.State.CurrentTurn >= len(e.State.PlayerOrder) {
		return nil
	}
	return e.State.Players[e.State.PlayerOrder[e.State.CurrentTurn]]
}

func (e *GameEngine) currentChoiceResumePoint() string {
	if e == nil || e.State == nil {
		panic("currentChoiceResumePoint: engine/state is nil")
	}
	if e.State.Subflow != model.SubflowNone {
		if !model.IsKnownSubflow(e.State.Subflow) {
			panic(fmt.Sprintf("currentChoiceResumePoint: unknown subflow %q", e.State.Subflow))
		}
		return normalizeChoiceResumePoint(e.State.Subflow)
	}
	if e.State.CombatStage != model.CombatStageNone {
		if !model.IsKnownCombatStage(e.State.CombatStage) {
			panic(fmt.Sprintf("currentChoiceResumePoint: unknown combat stage %q", e.State.CombatStage))
		}
		return normalizeChoiceResumePoint(e.State.CombatStage)
	}
	if e.State.TurnStage == "" {
		return ""
	}
	if !model.IsKnownTurnStage(e.State.TurnStage) {
		panic(fmt.Sprintf("currentChoiceResumePoint: unknown turn stage %q", e.State.TurnStage))
	}
	return normalizeChoiceResumePoint(e.State.TurnStage)
}

func (e *GameEngine) applyChoiceResumePoint(raw interface{}) bool {
	if e == nil || e.State == nil {
		return false
	}
	stage := parseChoiceResumeTurnStage(raw)
	combat := parseChoiceResumeCombatStage(raw)
	subflow := parseChoiceResumeSubflow(raw)
	if stage == "" && combat == model.CombatStageNone && subflow == model.SubflowNone {
		return false
	}

	e.State.ReturnTurnStage = ""
	e.State.ReturnCombatStage = model.CombatStageNone
	e.State.ReturnSubflow = model.SubflowNone

	if subflow != model.SubflowNone {
		e.State.Subflow = subflow
	} else {
		e.State.Subflow = model.SubflowNone
	}
	if combat != model.CombatStageNone {
		e.State.CombatStage = combat
	}
	if stage != "" {
		e.State.TurnStage = stage
	}
	return true
}

func (e *GameEngine) setReturnPoint(raw interface{}) bool {
	if e == nil || e.State == nil {
		return false
	}
	stage := parseChoiceResumeTurnStage(raw)
	combat := parseChoiceResumeCombatStage(raw)
	subflow := parseChoiceResumeSubflow(raw)
	if stage == "" && combat == model.CombatStageNone && subflow == model.SubflowNone {
		return false
	}
	e.State.ReturnTurnStage = stage
	e.State.ReturnCombatStage = combat
	e.State.ReturnSubflow = subflow
	return true
}

func (e *GameEngine) restoreReturnPoint() bool {
	if e == nil || e.State == nil {
		return false
	}
	if e.State.ReturnTurnStage == "" && e.State.ReturnCombatStage == model.CombatStageNone && e.State.ReturnSubflow == model.SubflowNone {
		return false
	}
	e.State.TurnStage = e.State.ReturnTurnStage
	e.State.CombatStage = e.State.ReturnCombatStage
	e.State.Subflow = e.State.ReturnSubflow
	e.State.ReturnTurnStage = ""
	e.State.ReturnCombatStage = model.CombatStageNone
	e.State.ReturnSubflow = model.SubflowNone
	return true
}
