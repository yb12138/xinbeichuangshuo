package model

import "testing"

func TestPromptFlowStateTracksHistoryAndAccumulatedSelections(t *testing.T) {
	flow := NewPromptFlowState("death_touch", "element")
	flow.PutSelection("element", PromptFlowSelection{Element: "Fire", OptionIndexes: []int{2}})
	flow.Advance("x")
	flow.PutSelection("x", PromptFlowSelection{Count: 3, OptionIndexes: []int{1}})
	flow.Advance("cards")

	if flow.FlowID != "death_touch" || flow.StepID != "cards" {
		t.Fatalf("unexpected flow position: %+v", flow)
	}
	if len(flow.History) != 2 || flow.History[0] != "element" || flow.History[1] != "x" {
		t.Fatalf("unexpected flow history: %+v", flow.History)
	}
	if got := flow.Selection("element").Element; got != "Fire" {
		t.Fatalf("expected element Fire, got %q", got)
	}
	if got := flow.Selection("x").Count; got != 3 {
		t.Fatalf("expected x count 3, got %d", got)
	}
}

func TestPromptFlowContextHelpers(t *testing.T) {
	ctx := map[string]interface{}{}
	flow := NewPromptFlowState("flow", "step")

	SetPromptFlowContext(ctx, flow)

	if got := PromptFlowFromContext(ctx); got != flow {
		t.Fatalf("expected context flow pointer")
	}
}
