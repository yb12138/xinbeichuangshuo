package valkyrie_test

import (
	"testing"

	"starcup-engine/internal/engine"
	playerpkg "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
	"starcup-engine/internal/testutils"
)

func newMilitaryGloryGame(t *testing.T, obs model.GameObserver) *engine.GameEngine {
	t.Helper()
	game := engine.NewGameEngine(obs)
	if err := game.AddPlayer("p1", "Valkyrie", "valkyrie", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Ally", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageTurnStart
	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	game.State.Players["p2"].TurnState = model.NewPlayerTurnState()
	playerpkg.SetForm(p1, model.FormValkyrieHeroic)
	game.State.RedCrystals = 2
	game.Drive()
	testutils.RequireChoicePrompt(t, game, "p1", "valkyrie_military_glory_mode")
	return game
}

func TestValkyrie_MilitaryGlory_SelfBranchEmitsSkillActivated(t *testing.T) {
	obs := &testutils.CaptureObserver{}
	game := newMilitaryGloryGame(t, obs)
	if n := len(obs.SkillActivated("valkyrie_military_glory")); n != 0 {
		t.Fatalf("mode prompt should not emit, got %d", n)
	}
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{0}})
	testutils.RequireSingleSkillActivation(t, obs, "valkyrie_military_glory", "p1", "p1")
}

func TestValkyrie_MilitaryGlory_TargetBranchEmitsSkillActivated(t *testing.T) {
	obs := &testutils.CaptureObserver{}
	game := newMilitaryGloryGame(t, obs)
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{1}})
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{0}})
	testutils.RequireChoicePrompt(t, game, "p1", "valkyrie_military_glory_target")
	if n := len(obs.SkillActivated("valkyrie_military_glory")); n != 0 {
		t.Fatalf("should not emit before target chosen, got %d", n)
	}
	ctxData, _ := game.State.PendingInterrupt.Context.(map[string]interface{})
	targetIDs, _ := ctxData["target_ids"].([]string)
	sel := -1
	for i, id := range targetIDs {
		if id == "p2" {
			sel = i
		}
	}
	if sel < 0 {
		t.Fatalf("p2 not in target list %v", targetIDs)
	}
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{sel}})
	testutils.RequireSingleSkillActivation(t, obs, "valkyrie_military_glory", "p1", "p2")
}
