package engine

import (
	"testing"

	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
)

// 回归：提炼选择弹窗取消后，应回到行动选择阶段而不是报错中断。
func TestExtractCancel_ReturnsToActionSelection(t *testing.T) {
	g := NewGameEngine(noopObserver{})
	if err := g.AddPlayer("p1", "Tester", "berserker", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	g.State.CurrentTurn = 0
	g.State.Deck = rules.InitDeck()
	g.State.Phase = model.PhaseActionSelection

	p1 := g.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{
		{ID: "atk-1", Name: "火斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
	}
	p1.Gem = 0
	p1.Crystal = 0

	g.State.RedGems = 0
	g.State.RedCrystals = 2

	mustDo(t, g, model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdExtract,
	})

	if g.State.PendingInterrupt == nil || g.State.PendingInterrupt.Type != model.InterruptChoice {
		t.Fatalf("expected extract choice interrupt, got %+v", g.State.PendingInterrupt)
	}
	ctx, ok := g.State.PendingInterrupt.Context.(map[string]interface{})
	if !ok {
		t.Fatalf("expected extract interrupt context map")
	}
	if ct, _ := ctx["choice_type"].(string); ct != "extract" {
		t.Fatalf("expected choice_type=extract, got %q", ct)
	}

	mustDo(t, g, model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdCancel,
	})

	if g.State.PendingInterrupt != nil {
		t.Fatalf("expected no pending interrupt after cancel, got %+v", g.State.PendingInterrupt)
	}
	if g.State.Phase != model.PhaseActionSelection {
		t.Fatalf("expected phase ActionSelection after cancel, got %s", g.State.Phase)
	}
	if g.State.CurrentTurn != 0 || g.State.PlayerOrder[g.State.CurrentTurn] != "p1" {
		t.Fatalf("expected turn to stay on p1, got turn=%d player=%s", g.State.CurrentTurn, g.State.PlayerOrder[g.State.CurrentTurn])
	}
	if g.State.RedCrystals != 2 || p1.Crystal != 0 {
		t.Fatalf("expected no extraction applied on cancel (red crystals=2, p1 crystal=0), got red=%d p1=%d", g.State.RedCrystals, p1.Crystal)
	}
}
