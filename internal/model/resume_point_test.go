package model

import "testing"

func mustPanic(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	fn()
}

func TestParseResumePointTurnStage_StrictEnumOnly(t *testing.T) {
	mustPanic(t, func() { ParseResumePointTurnStage("BeforeAction") })
	mustPanic(t, func() { ParseResumePointTurnStage("turn:BeforeAction") })
	if got := ParseResumePointTurnStage(TurnStageBeforeAction); got != TurnStageBeforeAction {
		t.Fatalf("expected typed turn stage accepted, got %q", got)
	}
}

func TestParseResumePointCombatStage_StrictEnumOnly(t *testing.T) {
	mustPanic(t, func() { ParseResumePointCombatStage("CombatCalcDamage") })
	mustPanic(t, func() { ParseResumePointCombatStage("combat:CombatCalcDamage") })
	if got := ParseResumePointCombatStage(CombatStageCalcDamage); got != CombatStageCalcDamage {
		t.Fatalf("expected typed combat stage accepted, got %q", got)
	}
}

func TestParseResumePointSubflow_StrictEnumOnly(t *testing.T) {
	mustPanic(t, func() { ParseResumePointSubflow("Response") })
	mustPanic(t, func() { ParseResumePointSubflow("subflow:Response") })
	if got := ParseResumePointSubflow(SubflowResponse); got != SubflowResponse {
		t.Fatalf("expected typed subflow accepted, got %q", got)
	}
}
