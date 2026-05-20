package player_test

import (
	"testing"

	"starcup-engine/internal/engine"
	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
	"starcup-engine/internal/testutils"
)

func TestAdvancePromptFlowRuntimeChoiceUsesStepSpecChoiceType(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "Player", "fighter", model.RedCamp); err != nil {
		t.Fatal(err)
	}

	flowRT := model.MustNewPromptFlowRuntime("sample", []model.PromptFlowStepSpec{
		{ID: "cards", ChoiceType: "sample_cards", CancelPolicy: model.CancelPolicyAbort},
		{ID: "target", ChoiceType: "sample_target", CancelPolicy: model.CancelPolicyBack},
	})
	flow := flowRT.Begin()
	ctxData := map[string]interface{}{
		"choice_type":              "sample_cards",
		model.PromptFlowContextKey: flow,
	}
	game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: "p1",
		Context:  ctxData,
	})

	if err := engineplayer.AdvancePromptFlowRuntimeChoice(engine.NewRoleChoiceRuntime(game), ctxData, flowRT, flow, "target"); err != nil {
		t.Fatal(err)
	}

	if got := ctxData["choice_type"]; got != "sample_target" {
		t.Fatalf("expected helper to derive choice_type from step spec, got %v", got)
	}
	if flow.StepID != "target" {
		t.Fatalf("expected flow to move to target step, got %q", flow.StepID)
	}
	pending, ok := game.State.PendingInterrupt.Context.(map[string]interface{})
	if !ok {
		t.Fatalf("expected pending interrupt context map, got %T", game.State.PendingInterrupt.Context)
	}
	if got := pending["choice_type"]; got != "sample_target" {
		t.Fatalf("expected pending context to receive derived choice_type, got %v", got)
	}
}
