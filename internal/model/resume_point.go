package model

import "strings"

const (
	ResumePointTurnStagePrefix   = "turn:"
	ResumePointCombatStagePrefix = "combat:"
	ResumePointSubflowPrefix     = "subflow:"
)

func NormalizeResumePoint(raw interface{}) string {
	switch value := raw.(type) {
	case string:
		return strings.TrimSpace(value)
	case TurnStage:
		if value == "" {
			return ""
		}
		return ResumePointTurnStagePrefix + string(value)
	case CombatStage:
		if value == "" {
			return ""
		}
		return ResumePointCombatStagePrefix + string(value)
	case Subflow:
		if value == "" {
			return ""
		}
		return ResumePointSubflowPrefix + string(value)
	default:
		return ""
	}
}

func IsKnownTurnStage(stage TurnStage) bool {
	switch stage {
	case TurnStageTurnBeforeStart,
		TurnStageTurnStart,
		TurnStageBeforeAction,
		TurnStageActionStart,
		TurnStageActionExecution,
		TurnStageActionEnd,
		TurnStageExtraAction,
		TurnStageTurnEnd:
		return true
	default:
		return false
	}
}

func IsKnownCombatStage(stage CombatStage) bool {
	switch stage {
	case CombatStageNone,
		CombatStageDeclare,
		CombatStageHitCheck,
		CombatStageCalcDamage,
		CombatStageHeal,
		CombatStageApply,
		CombatStageDraw:
		return true
	default:
		return false
	}
}

func IsKnownSubflow(subflow Subflow) bool {
	switch subflow {
	case SubflowNone,
		SubflowResponse,
		SubflowDiscardSelection:
		return true
	default:
		return false
	}
}

func ParseResumePointTurnStage(raw interface{}) TurnStage {
	point := NormalizeResumePoint(raw)
	switch {
	case strings.HasPrefix(point, ResumePointTurnStagePrefix):
		stage := TurnStage(strings.TrimPrefix(point, ResumePointTurnStagePrefix))
		if IsKnownTurnStage(stage) {
			return stage
		}
	case point != "":
		stage := TurnStage(point)
		if IsKnownTurnStage(stage) {
			return stage
		}
	}
	return ""
}

func ParseResumePointCombatStage(raw interface{}) CombatStage {
	point := NormalizeResumePoint(raw)
	switch {
	case strings.HasPrefix(point, ResumePointCombatStagePrefix):
		stage := CombatStage(strings.TrimPrefix(point, ResumePointCombatStagePrefix))
		if IsKnownCombatStage(stage) {
			return stage
		}
	case point != "":
		stage := CombatStage(point)
		if IsKnownCombatStage(stage) {
			return stage
		}
	}
	return CombatStageNone
}

func ParseResumePointSubflow(raw interface{}) Subflow {
	point := NormalizeResumePoint(raw)
	switch {
	case strings.HasPrefix(point, ResumePointSubflowPrefix):
		subflow := Subflow(strings.TrimPrefix(point, ResumePointSubflowPrefix))
		if IsKnownSubflow(subflow) {
			return subflow
		}
	case point != "":
		subflow := Subflow(point)
		if IsKnownSubflow(subflow) {
			return subflow
		}
	}
	return SubflowNone
}
