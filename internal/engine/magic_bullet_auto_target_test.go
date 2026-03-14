package engine

import (
	"testing"

	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
)

type noopMagicBulletObserver struct{}

func (noopMagicBulletObserver) OnGameEvent(event model.GameEvent) {}

func buildMagicActionGame(t *testing.T) *GameEngine {
	t.Helper()

	game := NewGameEngine(noopMagicBulletObserver{})
	game.State.Deck = rules.InitDeck()

	if err := game.AddPlayer("p1", "P1", "berserker", model.RedCamp); err != nil {
		t.Fatalf("add p1 failed: %v", err)
	}
	if err := game.AddPlayer("p2", "P2", "angel", model.BlueCamp); err != nil {
		t.Fatalf("add p2 failed: %v", err)
	}

	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection
	game.State.PlayerOrder = []string{"p1", "p2"}

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()

	return game
}

func TestMagicBullet_AllowsMagicWithoutExplicitTarget(t *testing.T) {
	game := buildMagicActionGame(t)
	p1 := game.State.Players["p1"]

	p1.Hand = []model.Card{
		{ID: "mb", Name: "魔弹", Type: model.CardTypeMagic, Element: model.ElementWater, Damage: 2},
	}

	err := game.HandleAction(model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdMagic,
		CardIndex: 0,
	})
	if err != nil {
		t.Fatalf("magic bullet without target should succeed, got: %v", err)
	}

	if game.State.PendingInterrupt == nil {
		t.Fatalf("expected magic missile interrupt, got nil")
	}
	if game.State.PendingInterrupt.Type != model.InterruptMagicMissile {
		t.Fatalf("expected interrupt type %s, got %s", model.InterruptMagicMissile, game.State.PendingInterrupt.Type)
	}
	if game.State.PendingInterrupt.PlayerID != "p2" {
		t.Fatalf("expected first bullet target p2, got %s", game.State.PendingInterrupt.PlayerID)
	}
}

func TestMagicWithoutTarget_StillRequiresTargetForNonMagicBullet(t *testing.T) {
	game := buildMagicActionGame(t)
	p1 := game.State.Players["p1"]

	p1.Hand = []model.Card{
		{ID: "shield", Name: "圣盾", Type: model.CardTypeMagic, Element: model.ElementLight, Damage: 0},
	}

	err := game.HandleAction(model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdMagic,
		CardIndex: 0,
	})
	if err == nil {
		t.Fatalf("non-magic-bullet without target should fail")
	}
}
