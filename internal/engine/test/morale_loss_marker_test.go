package engine_test

import (
	"testing"

	"starcup-engine/internal/engine"
	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
	"starcup-engine/internal/testutils"
)

// 士气下降要带原因（谁爆牌、谁造成的伤害），前端才能在战斗链里写清楚，而不是只有一行数字。
func TestMoraleLoss_PublishesCauseMarker(t *testing.T) {
	obs := &testutils.CaptureObserver{}
	game := engine.NewGameEngine(obs)
	_ = game.AddPlayer("p1", "Attacker", "berserker", model.BlueCamp)
	_ = game.AddPlayer("p2", "Victim", "angel", model.RedCamp)
	game.State.Deck = rules.InitDeck()
	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution
	for _, id := range []string{"p1", "p2"} {
		game.State.Players[id].TurnState = model.NewPlayerTurnState()
	}
	game.State.Players["p1"].IsActive = true
	victim := game.State.Players["p2"]
	victim.Hand = nil
	for i := 0; i < 6; i++ {
		victim.Hand = append(victim.Hand, model.Card{ID: string(rune('a' + i)), Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 2})
	}

	game.AddPendingDamage(model.PendingDamage{SourceID: "p1", TargetID: "p2", Damage: 1, DamageType: model.MagicAttack})
	game.ProcessPendingDamages()
	// 爆牌弃 1 张
	if err := game.HandleInterruptAction(model.PlayerAction{Type: model.CmdSelect, PlayerID: "p2", Selections: []int{0}}); err != nil {
		t.Fatalf("discard failed: %v", err)
	}
	if game.State.RedMorale != 14 {
		t.Fatalf("expected red morale 14, got %d", game.State.RedMorale)
	}

	var marker *model.TimelineMarkerPayload
	for _, ev := range obs.Events {
		if ev.Type == model.EventTimelineMarker && ev.TimelineMarker != nil && ev.TimelineMarker.NarrativeKind == "morale_loss" {
			marker = ev.TimelineMarker
		}
	}
	if marker == nil {
		t.Fatalf("expected a morale_loss timeline marker")
	}
	if marker.PlayerID != "p2" {
		t.Fatalf("marker should name the overflowing player, got %q", marker.PlayerID)
	}
	if len(marker.TargetIDs) < 2 || marker.TargetIDs[1] != "p1" {
		t.Fatalf("marker should carry the damage source, got %v", marker.TargetIDs)
	}
}
