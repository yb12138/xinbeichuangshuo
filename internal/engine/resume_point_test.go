package engine

import (
	"testing"

	"starcup-engine/internal/model"
)

func TestParseChoiceResumeTurnStage_BareBeforeActionStaysExactStage(t *testing.T) {
	if got := parseChoiceResumeTurnStage("BeforeAction"); got != model.TurnStageBeforeAction {
		t.Fatalf("expected bare BeforeAction to stay exact turn stage, got %s", got)
	}
}

func TestParseChoiceResumeTurnStage_ExplicitStageRoundTrips(t *testing.T) {
	point := model.NormalizeResumePoint(model.TurnStageBeforeAction)
	if got := parseChoiceResumeTurnStage(point); got != model.TurnStageBeforeAction {
		t.Fatalf("expected explicit stage BeforeAction to round-trip, got %s", got)
	}
}

func TestParseChoiceResumeTurnStage_ActionStartRoundTrips(t *testing.T) {
	point := model.NormalizeResumePoint(model.TurnStageActionStart)
	if got := parseChoiceResumeTurnStage(point); got != model.TurnStageActionStart {
		t.Fatalf("expected explicit stage ActionStart to round-trip, got %s", got)
	}
}

func TestParseChoiceResumeCombatStage_RoundTrips(t *testing.T) {
	point := model.NormalizeResumePoint(model.CombatStageCalcDamage)
	if got := parseChoiceResumeCombatStage(point); got != model.CombatStageCalcDamage {
		t.Fatalf("expected explicit combat stage CalcDamage to round-trip, got %s", got)
	}
}

func TestParseChoiceResumeSubflow_RoundTrips(t *testing.T) {
	point := model.NormalizeResumePoint(model.SubflowDiscardSelection)
	if got := parseChoiceResumeSubflow(point); got != model.SubflowDiscardSelection {
		t.Fatalf("expected explicit discard-selection subflow to round-trip, got %s", got)
	}
}
