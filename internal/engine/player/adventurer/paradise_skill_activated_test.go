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

// 冒险家-1：冒险家没有能量时发动天堂，不需要移除能量；星石交给队友。
func TestAdventurerParadise_NoEnergyNoCost(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
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
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	game.State.RedGems = 1
	game.State.RedCrystals = 1

	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdExtract})
	testutils.RequireChoicePrompt(t, game, "p1", "adventurer_extract_paradise_check")
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{0}})
	testutils.RequireChoicePrompt(t, game, "p1", "adventurer_paradise_pick")
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{0}})
	testutils.RequireChoicePrompt(t, game, "p1", "extract")
	testutils.MustHandleAction(t, game, model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{0, 1}})

	if p1.Gem != 0 || p1.Crystal != 0 {
		t.Fatalf("adventurer should stay at 0 energy, got gem=%d crystal=%d", p1.Gem, p1.Crystal)
	}
	if p2.Gem != 1 || p2.Crystal != 1 {
		t.Fatalf("ally should receive 1 gem 1 crystal, got gem=%d crystal=%d", p2.Gem, p2.Crystal)
	}
	if game.State.RedGems != 0 || game.State.RedCrystals != 0 {
		t.Fatalf("camp stones should be consumed, got gems=%d crystals=%d", game.State.RedGems, game.State.RedCrystals)
	}
}

// 冒险家-2：自身能量已满、有可承接的队友时，行动菜单里仍有"提炼"。
func TestAdventurerParadise_FullEnergyMenuOffersExtract(t *testing.T) {
	hasExtract := func(allyFull bool) bool {
		obs := &testutils.CaptureObserver{}
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
		p1.Gem = 3
		if allyFull {
			game.State.Players["p2"].Gem = 3
		}
		game.State.RedGems = 1
		game.Drive()
		var last *model.Prompt
		for _, ev := range obs.Events {
			if ev.Type == model.EventAskInput && ev.Prompt != nil {
				last = ev.Prompt
			}
		}
		if last == nil {
			t.Fatalf("expected action prompt")
		}
		for _, opt := range last.SpecialOptions {
			if opt.ID == "extract" {
				return true
			}
		}
		return false
	}
	if !hasExtract(false) {
		t.Fatalf("full-energy adventurer with eligible ally should be offered extract")
	}
	if hasExtract(true) {
		t.Fatalf("no eligible ally: extract should not be offered at full energy")
	}
}
