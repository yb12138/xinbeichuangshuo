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

func TestPromptFlowRuntimeRecordsTransitionsAndBack(t *testing.T) {
	rt, err := NewPromptFlowRuntime("sample", []PromptFlowStepSpec{
		{ID: "cards", ChoiceType: "sample_cards", CancelPolicy: CancelPolicyAbort},
		{ID: "target", ChoiceType: "sample_target", CancelPolicy: CancelPolicyBack},
	})
	if err != nil {
		t.Fatalf("runtime init failed: %v", err)
	}

	flow := rt.Begin()
	if flow == nil || flow.StepID != "cards" {
		t.Fatalf("unexpected initial flow: %+v", flow)
	}

	if err := rt.RecordAndMove(flow, PromptFlowSelection{CardIDs: []string{"c1"}}, "target"); err != nil {
		t.Fatalf("transition failed: %v", err)
	}
	if flow.StepID != "target" || flow.Selection("cards").CardIDs[0] != "c1" {
		t.Fatalf("unexpected flow after transition: %+v", flow)
	}

	result, err := rt.HandleCancel(flow)
	if err != nil {
		t.Fatalf("back cancel failed: %v", err)
	}
	if result.Action != PromptFlowCancelBack || result.StepID != "cards" || flow.StepID != "cards" {
		t.Fatalf("unexpected back result=%+v flow=%+v", result, flow)
	}
	if step, ok := flow.CurrentStepSpec(); !ok || step.ChoiceType != "sample_cards" {
		t.Fatalf("expected copied step spec on state, got step=%+v ok=%v", step, ok)
	}
}

func TestPromptFlowRuntimeBeginAtStartsWithoutSyntheticHistory(t *testing.T) {
	rt, err := NewPromptFlowRuntime("sample", []PromptFlowStepSpec{
		{ID: "mode"},
		{ID: "cards"},
	})
	if err != nil {
		t.Fatalf("runtime init failed: %v", err)
	}

	flow, err := rt.BeginAt("cards")
	if err != nil {
		t.Fatalf("begin at failed: %v", err)
	}
	if flow.FlowID != "sample" || flow.StepID != "cards" || len(flow.History) != 0 {
		t.Fatalf("unexpected flow from BeginAt: %+v", flow)
	}
	if _, err := rt.BeginAt("unknown"); err == nil {
		t.Fatalf("expected unknown step error")
	}
}

func TestPromptFlowRuntimeCancelPolicies(t *testing.T) {
	rt, err := NewPromptFlowRuntime("sample", []PromptFlowStepSpec{
		{ID: "deny", CancelPolicy: CancelPolicyDeny},
		{ID: "abort", CancelPolicy: CancelPolicyAbort},
		{ID: "decline", CancelPolicy: CancelPolicyDecline},
	})
	if err != nil {
		t.Fatalf("runtime init failed: %v", err)
	}

	deny := NewPromptFlowState("sample", "deny")
	if result, err := rt.HandleCancel(deny); err == nil || result.Action != PromptFlowCancelDenied {
		t.Fatalf("expected denied cancel, got result=%+v err=%v", result, err)
	}

	abort := NewPromptFlowState("sample", "abort")
	abort.PutSelection("abort", PromptFlowSelection{Count: 1})
	result, err := rt.HandleCancel(abort)
	if err != nil {
		t.Fatalf("abort cancel failed: %v", err)
	}
	if result.Action != PromptFlowCancelAborted || abort.StepID != "" || len(abort.AccumulatedSelections) != 0 {
		t.Fatalf("unexpected abort result=%+v flow=%+v", result, abort)
	}

	decline := NewPromptFlowState("sample", "decline")
	result, err = rt.HandleCancel(decline)
	if err != nil {
		t.Fatalf("decline cancel failed: %v", err)
	}
	if result.Action != PromptFlowCancelDeclined {
		t.Fatalf("unexpected decline result=%+v", result)
	}
	if _, ok := decline.AccumulatedSelections["decline"]; !ok {
		t.Fatalf("decline should record an explicit empty selection")
	}
}
