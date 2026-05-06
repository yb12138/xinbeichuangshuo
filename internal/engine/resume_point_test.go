package engine

import (
	"testing"

	"starcup-engine/internal/model"
)

func mustPanic(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	fn()
}

func TestParseChoiceResumeTurnStage_BareBeforeActionStaysExactStage(t *testing.T) {
	mustPanic(t, func() { parseChoiceResumeTurnStage("BeforeAction") })
}

func TestParseChoiceResumeTurnStage_ExplicitStageRoundTrips(t *testing.T) {
	if got := parseChoiceResumeTurnStage(model.TurnStageBeforeAction); got != model.TurnStageBeforeAction {
		t.Fatalf("expected explicit stage BeforeAction to round-trip, got %s", got)
	}
}

func TestParseChoiceResumeTurnStage_ActionStartRoundTrips(t *testing.T) {
	if got := parseChoiceResumeTurnStage(model.TurnStageActionStart); got != model.TurnStageActionStart {
		t.Fatalf("expected explicit stage ActionStart to round-trip, got %s", got)
	}
}

func TestParseChoiceResumeCombatStage_RoundTrips(t *testing.T) {
	if got := parseChoiceResumeCombatStage(model.CombatStageCalcDamage); got != model.CombatStageCalcDamage {
		t.Fatalf("expected explicit combat stage CalcDamage to round-trip, got %s", got)
	}
}

func TestParseChoiceResumeSubflow_RoundTrips(t *testing.T) {
	if got := parseChoiceResumeSubflow(model.SubflowDiscardSelection); got != model.SubflowDiscardSelection {
		t.Fatalf("expected explicit discard-selection subflow to round-trip, got %s", got)
	}
}

func TestParseChoiceResumeCombatStage_BareValueRejected(t *testing.T) {
	mustPanic(t, func() { parseChoiceResumeCombatStage("CombatCalcDamage") })
}

func TestParseChoiceResumeSubflow_BareValueRejected(t *testing.T) {
	mustPanic(t, func() { parseChoiceResumeSubflow("Response") })
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
