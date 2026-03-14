package engine

import (
	"testing"

	"starcup-engine/internal/model"
)

func TestBerserkerAttackSealer_TakeDamageResolves(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Berserker", "berserker", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Sealer", "sealer", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Gem = 0

	// 手牌 6 张，打出 1 张后剩 5，满足狂化额外+1 条件
	p1.Hand = []model.Card{
		{ID: "a1", Name: "雷光斩", Type: model.CardTypeAttack, Element: model.ElementThunder, Damage: 2},
		{ID: "f1", Name: "填充1", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 2},
		{ID: "f2", Name: "填充2", Type: model.CardTypeAttack, Element: model.ElementWater, Damage: 2},
		{ID: "f3", Name: "填充3", Type: model.CardTypeAttack, Element: model.ElementWind, Damage: 2},
		{ID: "f4", Name: "填充4", Type: model.CardTypeAttack, Element: model.ElementEarth, Damage: 2},
		{ID: "f5", Name: "填充5", Type: model.CardTypeAttack, Element: model.ElementThunder, Damage: 2},
	}

	// 目标初始手牌为空，便于验证摸牌数量
	p2.Hand = nil
	game.State.Deck = []model.Card{
		{ID: "d1", Name: "补牌1", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 2},
		{ID: "d2", Name: "补牌2", Type: model.CardTypeAttack, Element: model.ElementWater, Damage: 2},
		{ID: "d3", Name: "补牌3", Type: model.CardTypeAttack, Element: model.ElementThunder, Damage: 2},
		{ID: "d4", Name: "补牌4", Type: model.CardTypeAttack, Element: model.ElementWind, Damage: 2},
	}

	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection

	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdAttack,
		TargetID:  "p2",
		CardIndex: 0,
	})

	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		ExtraArgs: []string{"take"},
	})

	if game.State.PendingInterrupt != nil {
		t.Fatalf("expected no pending interrupt after take, got %v", game.State.PendingInterrupt.Type)
	}
	if got := len(game.State.PendingDamageQueue); got != 0 {
		t.Fatalf("expected pending damage drained, got %d", got)
	}
	if got := len(p2.Hand); got != 4 {
		t.Fatalf("expected sealer draw=4 (雷光斩2 + 狂化2), got %d", got)
	}
}
