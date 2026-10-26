package hero_test

import (
	"testing"

	"starcup-engine/internal/engine"
	"starcup-engine/internal/model"
	"starcup-engine/internal/testutils"
)

func TestHeroExhaustion_ReleaseEmitsSkillActivated(t *testing.T) {
	obs := &testutils.CaptureObserver{}
	game := engine.NewGameEngine(obs)
	if err := game.AddPlayer("p1", "Hero", "hero", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Target", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Form = model.FormHeroExhaustion
	p1.TurnState.SkillFlowState["hero_exhaustion_release_pending"] = 1
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionStart

	game.Drive()

	testutils.RequireSingleSkillActivation(t, obs, "hero_exhaustion", "p1", "p1")
}

func TestHeroExhaustion_NoReleasePendingDoesNotEmit(t *testing.T) {
	obs := &testutils.CaptureObserver{}
	game := engine.NewGameEngine(obs)
	if err := game.AddPlayer("p1", "Hero", "hero", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Target", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Form = model.FormHeroExhaustion
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionStart

	game.Drive()

	if n := len(obs.SkillActivated("hero_exhaustion")); n != 0 {
		t.Fatalf("exhaustion still active should not emit, got %d", n)
	}
}
