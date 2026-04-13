package engine

import (
	"testing"

	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
)

func TestSaintessFrostPrayer_DefendChoiceDoesNotReopenActionSelection(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Attacker", "berserker", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Saintess", "saintess", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Ally", "angel", model.BlueCamp); err != nil {
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
	p1.Hand = []model.Card{
		{ID: "atk-1", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 2},
		{ID: "atk-2", Name: "地裂斩", Type: model.CardTypeAttack, Element: model.ElementEarth, Damage: 2},
	}
	p2.Hand = []model.Card{
		{ID: "holy-light", Name: "圣光", Type: model.CardTypeMagic, Element: model.ElementLight},
	}

	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdAttack,
		TargetID:  "p2",
		CardIndex: 0,
	})
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		ExtraArgs: []string{"defend"},
		CardIndex: 0,
	})

	ctx := requireChoiceContext(t, game, "p2", "frost_prayer_target")
	targetSelection := choiceIndexForTarget(t, ctx, "p2")
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p2",
		Type:       model.CmdSelect,
		Selections: []int{targetSelection},
	})

	if len(p1.Hand) != 1 {
		t.Fatalf("expected attacker to keep one unplayed card after first action, got %d", len(p1.Hand))
	}
	if err := game.HandleAction(model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdAttack,
		TargetID:  "p2",
		CardIndex: 0,
	}); err == nil {
		t.Fatalf("expected second immediate attack to be rejected after frost-prayer resume, but got nil")
	}
	if len(p1.Hand) != 1 {
		t.Fatalf("expected failed second attack not to consume card, got %d cards", len(p1.Hand))
	}
}
