package engine

import (
	"testing"

	"starcup-engine/internal/model"
)

func promptFlowProgressGame(t *testing.T, flow *model.PromptFlowState) *GameEngine {
	t.Helper()
	game := NewGameEngine(nil)
	if err := game.AddPlayer("p1", "A", "butterfly_dancer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	ctx := map[string]interface{}{}
	if flow != nil {
		ctx[model.PromptFlowContextKey] = flow
	}
	game.State.PendingInterrupt = &model.Interrupt{Type: model.InterruptChoice, PlayerID: "p1", Context: ctx}
	return game
}

func testPrompt() *model.Prompt {
	return &model.Prompt{
		PlayerID:     "p1",
		Message:      "test",
		Presentation: &model.PromptPresentation{Kind: model.PresentationBranchSelect, Layout: "overlay"},
	}
}

func TestDecoratePromptForClient_AnnotatesMultiStepFlowProgress(t *testing.T) {
	rt := model.MustNewPromptFlowRuntime("test_flow", []model.PromptFlowStepSpec{
		{ID: "branch", ChoiceType: "a"},
		{ID: "target", ChoiceType: "b"},
		{ID: "confirm", ChoiceType: "c"},
	})
	flow := rt.Begin()
	game := promptFlowProgressGame(t, flow)
	original := testPrompt()

	got := game.decoratePromptForClient(original)
	if got.Presentation.StepIndex != 1 || got.Presentation.TotalSteps != 3 {
		t.Fatalf("first step: got %d/%d, want 1/3", got.Presentation.StepIndex, got.Presentation.TotalSteps)
	}
	if original.Presentation.StepIndex != 0 {
		t.Fatalf("decorate must not mutate the engine's prompt, got StepIndex=%d", original.Presentation.StepIndex)
	}

	if err := rt.MoveTo(flow, "target"); err != nil {
		t.Fatal(err)
	}
	got = game.decoratePromptForClient(testPrompt())
	if got.Presentation.StepIndex != 2 || got.Presentation.TotalSteps != 3 {
		t.Fatalf("second step: got %d/%d, want 2/3", got.Presentation.StepIndex, got.Presentation.TotalSteps)
	}
}

func TestDecoratePromptForClient_LeavesSingleStepPromptsUnannotated(t *testing.T) {
	for name, flow := range map[string]*model.PromptFlowState{
		"no flow":     nil,
		"single step": model.MustNewPromptFlowRuntime("one", []model.PromptFlowStepSpec{{ID: "only", ChoiceType: "x"}}).Begin(),
	} {
		got := promptFlowProgressGame(t, flow).decoratePromptForClient(testPrompt())
		if got.Presentation.StepIndex != 0 || got.Presentation.TotalSteps != 0 {
			t.Fatalf("%s: expected no progress, got %d/%d", name, got.Presentation.StepIndex, got.Presentation.TotalSteps)
		}
	}
}
