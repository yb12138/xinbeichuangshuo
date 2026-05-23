package engine_test

import (
	"testing"

	"starcup-engine/internal/engine"
	"starcup-engine/internal/model"
	"starcup-engine/internal/testutils"
)

func TestPromptFlowCancelBackUsesRuntimeStepSpec(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Sword", "sword_emperor", model.RedCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	flow := model.NewPromptFlowState("se_sword_qi_slash", "target")
	flow.History = []string{"x"}
	flow.StepSpecs = map[string]model.PromptFlowStepSpec{
		"x":      {ID: "x", ChoiceType: "se_sword_qi_slash_x", CancelPolicy: model.CancelPolicyBack},
		"target": {ID: "target", ChoiceType: "se_sword_qi_slash_target", CancelPolicy: model.CancelPolicyBack},
	}
	game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: "p1",
		Context: map[string]interface{}{
			"choice_type":              "se_sword_qi_slash_target",
			"max_x":                    1,
			model.PromptFlowContextKey: flow,
		},
	})

	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdCancel,
	})

	ctxData := testutils.RequireChoiceContext(t, game, "p1", "se_sword_qi_slash_x")
	gotFlow := testutils.RequirePromptFlow(t, ctxData, "se_sword_qi_slash", "x")
	if len(gotFlow.History) != 0 {
		t.Fatalf("back should pop flow history, got %+v", gotFlow.History)
	}
	if p1 == nil {
		t.Fatalf("player should still exist after back")
	}
}

func TestPromptFlowCancelAbortClearsInterrupt(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Sword", "sword_emperor", model.RedCamp); err != nil {
		t.Fatal(err)
	}

	flow := model.NewPromptFlowState("se_sword_qi_slash", "target")
	flow.History = []string{"x"}
	flow.StepSpecs = map[string]model.PromptFlowStepSpec{
		"x":      {ID: "x", ChoiceType: "se_sword_qi_slash_x", CancelPolicy: model.CancelPolicyBack},
		"target": {ID: "target", ChoiceType: "se_sword_qi_slash_target", CancelPolicy: model.CancelPolicyAbort},
	}
	game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: "p1",
		Context: map[string]interface{}{
			"choice_type":              "se_sword_qi_slash_target",
			model.PromptFlowContextKey: flow,
		},
	})
	testutils.RequireChoicePrompt(t, game, "p1", "se_sword_qi_slash_target")

	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdCancel,
	})

	if game.State.PendingInterrupt != nil {
		t.Fatalf("abort should clear prompt, got %+v", game.State.PendingInterrupt)
	}
}
