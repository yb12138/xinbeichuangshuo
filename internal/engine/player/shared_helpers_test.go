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

func TestAddPlayerEnergyCapped(t *testing.T) {
	t.Run("clips gem gain to remaining room", func(t *testing.T) {
		p := &model.Player{Gem: 2, Crystal: 1}
		got := engineplayer.AddPlayerGemCapped(p, 2, 4)
		if got != 1 {
			t.Fatalf("expected actual gem gain=1, got %d", got)
		}
		if p.Gem != 3 || p.Crystal != 1 {
			t.Fatalf("expected total energy capped at 4 with gem=3 crystal=1, got gem=%d crystal=%d", p.Gem, p.Crystal)
		}
	})

	t.Run("does not add crystal when energy is full", func(t *testing.T) {
		p := &model.Player{Gem: 1, Crystal: 2}
		got := engineplayer.AddPlayerCrystalCapped(p, 1, 3)
		if got != 0 {
			t.Fatalf("expected actual crystal gain=0, got %d", got)
		}
		if p.Gem != 1 || p.Crystal != 2 {
			t.Fatalf("expected energy unchanged at cap, got gem=%d crystal=%d", p.Gem, p.Crystal)
		}
	})
}
