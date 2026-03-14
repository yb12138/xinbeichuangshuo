package engine

import (
	"testing"

	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
)

// 回归：祈祷师进入祈祷形态后应持续到对局结束，不会在回合结束时自动退出。
func TestPrayerForm_PersistsAfterTurnEnd(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Prayer", "prayer_master", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.Deck = rules.InitDeck()
	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseTurnEnd

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.Tokens["prayer_form"] = 1
	p1.Tokens["prayer_rune"] = 3

	game.Drive()

	if got := p1.Tokens["prayer_form"]; got != 1 {
		t.Fatalf("expected prayer_form remain 1 after turn end, got %d", got)
	}
	if got := p1.Tokens["prayer_rune"]; got != 3 {
		t.Fatalf("expected prayer_rune remain 3 after turn end, got %d", got)
	}
}
