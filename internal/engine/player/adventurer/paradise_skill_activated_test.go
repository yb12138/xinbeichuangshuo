package adventurer_test

import (
	"testing"

	"starcup-engine/internal/engine"
	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
	"starcup-engine/internal/testutils"
)

func newParadiseGame(t *testing.T, obs model.GameObserver) *engine.GameEngine {
	t.Helper()
	game := engine.NewGameEngine(obs)
	if err := game.AddPlayer("p1", "Adventurer", "adventurer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Ally", "berserker", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	game.State.CurrentTurn = 0
	game.State.Deck = rules.InitDeck()
	game.State.TurnStage = model.TurnStageActionExecution
	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Gem = 2
	game.State.RedGems = 1
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdExtract})
	testutils.RequireChoicePrompt(t, game, "p1", "adventurer_extract_paradise_check")
	return game
}

func TestAdventurerParadise_AllyPickEmitsSkillActivated(t *testing.T) {
	obs := &testutils.CaptureObserver{}
	game := newParadiseGame(t, obs)
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{0}})
	testutils.RequireChoicePrompt(t, game, "p1", "adventurer_paradise_pick")
	if n := len(obs.SkillActivated("adventurer_paradise")); n != 0 {
		t.Fatalf("should not emit before ally picked, got %d", n)
	}
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{0}})
	testutils.RequireSingleSkillActivation(t, obs, "adventurer_paradise", "p1", "p2")
}

func TestAdventurerParadise_DeclineDoesNotEmit(t *testing.T) {
	obs := &testutils.CaptureObserver{}
	game := newParadiseGame(t, obs)
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{1}})
	if n := len(obs.SkillActivated("adventurer_paradise")); n != 0 {
		t.Fatalf("declined paradise should not emit, got %d", n)
	}
}
