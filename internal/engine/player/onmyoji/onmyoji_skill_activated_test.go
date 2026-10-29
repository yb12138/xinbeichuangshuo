package onmyoji_test

import (
	"reflect"
	"testing"

	"starcup-engine/internal/engine"
	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
	"starcup-engine/internal/testutils"
)

func TestOnmyojiDarkRitual_EmitsSkillActivatedWithTarget(t *testing.T) {
	obs := &testutils.CaptureObserver{}
	game := engine.NewGameEngine(obs)
	if err := game.AddPlayer("p1", "Onmyoji", "onmyoji", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Tokens["onmyoji_ghost_fire"] = 3
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageTurnEnd

	game.Drive()
	if got := choiceTypeOf(game.State.PendingInterrupt); got != "onmyoji_dark_ritual_target" {
		t.Fatalf("expected onmyoji_dark_ritual_target prompt, got %s", got)
	}
	if got := len(obs.SkillActivated("onmyoji_dark_ritual")); got != 0 {
		t.Fatalf("expected no dark ritual event before target choice, got %d", got)
	}
	if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p1", Selections: []int{0}}); err != nil {
		t.Fatalf("choose dark ritual target failed: %v", err)
	}

	events := obs.SkillActivated("onmyoji_dark_ritual")
	if len(events) != 1 {
		t.Fatalf("expected exactly one dark ritual skill_activated, got %d", len(events))
	}
	if events[0].PlayerID != "p1" || !reflect.DeepEqual(events[0].TargetIDs, []string{"p2"}) {
		t.Fatalf("unexpected payload: %+v", events[0])
	}
}

func TestOnmyojiYinYangShift_EmitsSkillActivatedTargetingAttacker(t *testing.T) {
	obs := &testutils.CaptureObserver{}
	game := engine.NewGameEngine(obs)
	if err := game.AddPlayer("p1", "Attacker", "berserker", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Onmyoji", "onmyoji", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "RedAlly", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	game.State.CurrentTurn = 0
	game.State.Deck = rules.InitDeck()
	game.State.TurnStage = model.TurnStageActionExecution
	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()
	game.State.Players["p3"].TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{
		{ID: "atk1", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Faction: "咏", Damage: 2},
	}
	p2.Hand = []model.Card{
		{ID: "55", Name: "水涟斩", Type: model.CardTypeAttack, Element: model.ElementWater, Faction: "咏", Damage: 2},
	}

	testutils.MustHandleAction(t, game, model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdAttack,
		TargetID: "p2",
		CardID:   testutils.PlayableCardID(t, game, "p1", 0),
	})
	if got := choiceTypeOf(game.State.PendingInterrupt); got != "onmyoji_yinyang_confirm" {
		t.Fatalf("expected onmyoji_yinyang_confirm prompt, got %s", got)
	}
	testutils.MustHandleAction(t, game, model.PlayerAction{Type: model.CmdSelect, PlayerID: "p2", Selections: []int{0}})
	testutils.MustHandleAction(t, game, model.PlayerAction{Type: model.CmdSelect, PlayerID: "p2", CardIDs: []string{"55"}})
	if got := len(obs.SkillActivated("onmyoji_yinyang_shift")); got != 0 {
		t.Fatalf("expected no yinyang event before counter target choice, got %d", got)
	}
	testutils.MustHandleAction(t, game, model.PlayerAction{Type: model.CmdSelect, PlayerID: "p2", Selections: []int{0}})

	events := obs.SkillActivated("onmyoji_yinyang_shift")
	if len(events) != 1 {
		t.Fatalf("expected exactly one yinyang shift skill_activated, got %d", len(events))
	}
	if events[0].PlayerID != "p2" || !reflect.DeepEqual(events[0].TargetIDs, []string{"p1"}) {
		t.Fatalf("unexpected payload: %+v", events[0])
	}
}
