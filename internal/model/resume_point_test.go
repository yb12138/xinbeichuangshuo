package model

import "testing"

func TestParseResumePointTurnStage_StrictPrefixOnly(t *testing.T) {
	if got := ParseResumePointTurnStage("BeforeAction"); got != "" {
		t.Fatalf("expected bare turn stage rejected, got %q", got)
	}
	if got := ParseResumePointTurnStage("turn:BeforeAction"); got != TurnStageBeforeAction {
		t.Fatalf("expected prefixed turn stage accepted, got %q", got)
	}
}

func TestParseResumePointCombatStage_StrictPrefixOnly(t *testing.T) {
	if got := ParseResumePointCombatStage("CombatCalcDamage"); got != CombatStageNone {
		t.Fatalf("expected bare combat stage rejected, got %q", got)
	}
	if got := ParseResumePointCombatStage("combat:CombatCalcDamage"); got != CombatStageCalcDamage {
		t.Fatalf("expected prefixed combat stage accepted, got %q", got)
	}
}

func TestParseResumePointSubflow_StrictPrefixOnly(t *testing.T) {
	if got := ParseResumePointSubflow("Response"); got != SubflowNone {
		t.Fatalf("expected bare subflow rejected, got %q", got)
	}
	if got := ParseResumePointSubflow("subflow:Response"); got != SubflowResponse {
		t.Fatalf("expected prefixed subflow accepted, got %q", got)
	}
}
