package engine_test

import (
	"testing"

	"starcup-engine/internal/engine"
	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
	"starcup-engine/internal/testutils"
)

func TestMagicMissileFusionOfferedInMissilePanel(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	game.State.Deck = rules.InitDeck()
	if err := game.AddPlayer("p1", "Caster", "berserker", model.RedCamp); err != nil {
		t.Fatalf("add p1 failed: %v", err)
	}
	if err := game.AddPlayer("p2", "Responder", "magical_girl", model.BlueCamp); err != nil {
		t.Fatalf("add p2 failed: %v", err)
	}

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution
	game.State.PlayerOrder = []string{"p1", "p2"}
	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{
		{ID: "magic-missile", Name: "魔弹", Type: model.CardTypeMagic, Element: model.ElementWater, Damage: 2},
	}
	p2.Hand = []model.Card{
		{ID: "fusion-card", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire},
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdMagic,
		CardID:   testutils.PlayableCardID(t, game, "p1", 0),
	}); err != nil {
		t.Fatalf("cast magic missile failed: %v", err)
	}

	// 魔弹融合已并入魔弹响应面板：不再先弹响应技能窗口，融合牌直接作为"传递"牌给出。
	intr := game.State.PendingInterrupt
	if intr == nil || intr.Type != model.InterruptMagicMissile || intr.PlayerID != "p2" {
		t.Fatalf("expected magic missile prompt on p2, got %+v", intr)
	}
	prompt := game.GetCurrentPrompt()
	found := false
	for _, id := range prompt.CounterCardIDs {
		if id == "fusion-card" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected fusion card offered as counter, got %v", prompt.CounterCardIDs)
	}
}
