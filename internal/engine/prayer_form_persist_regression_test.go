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
	game.State.TurnStage = model.TurnStageTurnEnd

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.Form = model.FormPrayerMasterPrayer
	p1.Tokens["prayer_rune"] = 3

	game.Drive()

	if got := p1.Form; got != model.FormPrayerMasterPrayer {
		t.Fatalf("expected prayer form remain %q after turn end, got %q", model.FormPrayerMasterPrayer, got)
	}
	if got := p1.Tokens["prayer_rune"]; got != 3 {
		t.Fatalf("expected prayer_rune remain 3 after turn end, got %d", got)
	}
	if got := game.GetMaxHand(p1); got != 5 {
		t.Fatalf("expected prayer form fixed max hand=5, got %d", got)
	}
}
