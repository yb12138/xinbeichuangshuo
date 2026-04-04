package engine

import (
	"testing"

	"starcup-engine/internal/model"
)

func TestParseChoiceResumeTurnStage_BareBeforeActionStaysExactStage(t *testing.T) {
	if got := parseChoiceResumeTurnStage("BeforeAction"); got != "" {
		t.Fatalf("expected bare BeforeAction rejected in strict mode, got %s", got)
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

func TestParseChoiceResumeCombatStage_BareValueRejected(t *testing.T) {
	if got := parseChoiceResumeCombatStage("CombatCalcDamage"); got != model.CombatStageNone {
		t.Fatalf("expected bare combat stage rejected in strict mode, got %s", got)
	}
}

func TestParseChoiceResumeSubflow_BareValueRejected(t *testing.T) {
	if got := parseChoiceResumeSubflow("Response"); got != model.SubflowNone {
		t.Fatalf("expected bare subflow rejected in strict mode, got %s", got)
	}
}

func TestCurrentChoiceResumePoint_InvalidTurnStagePanics(t *testing.T) {
	game := NewGameEngine(nil)
	game.State.TurnStage = model.TurnStage("InvalidStage")
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic for invalid turn stage")
		}
	}()
	_ = game.currentChoiceResumePoint()
}

func TestCurrentChoiceResumePoint_NilEnginePanics(t *testing.T) {
	var game *GameEngine
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic for nil engine")
		}
	}()
	_ = game.currentChoiceResumePoint()
}
