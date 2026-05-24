package server

import (
	"testing"

	"starcup-engine/internal/engine"
	"starcup-engine/internal/model"
)

func TestPublicTimelineStateDeltaDiffsVisibleState(t *testing.T) {
	room := NewRoom("DELTA")
	room.Engine = engine.NewGameEngine(room)
	if err := room.Engine.AddPlayer("p1", "Alice", "berserker", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := room.Engine.AddPlayer("p2", "Bob", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	room.resetPublicTimelineSnapshot()

	state := room.Engine.State
	state.RedMorale = 13
	state.RedGems = 2
	state.DiscardPile = append(state.DiscardPile, model.Card{ID: "discarded", Name: "弃牌", Type: model.CardTypeAttack})
	state.Players["p1"].Gem = 1
	state.Players["p1"].Hand = append(state.Players["p1"].Hand, model.Card{ID: "h1", Name: "手牌", Type: model.CardTypeMagic})
	state.Players["p1"].Field = append(state.Players["p1"].Field, &model.FieldCard{
		Card:     model.Card{ID: "secret", Name: "秘密牌", Type: model.CardTypeMagic},
		OwnerID:  "p1",
		SourceID: "p1",
		Mode:     model.FieldCover,
		Effect:   model.EffectMagicBowCharge,
	})

	next := room.capturePublicTimelineSnapshot()
	deltas := diffPublicTimelineSnapshots(room.publicTimelineSnapshot, next, "test")

	assertDelta := func(deltaType string) {
		t.Helper()
		for _, delta := range deltas {
			if delta.Type == deltaType {
				return
			}
		}
		t.Fatalf("expected delta %s in %+v", deltaType, deltas)
	}
	assertDelta("morale")
	assertDelta("team_gem")
	assertDelta("discard_count")
	assertDelta("player_gem")
	assertDelta("hand_count")
	assertDelta("field_card_added")

	for _, delta := range deltas {
		if delta.Type == "field_card_added" {
			if delta.FieldCard == nil {
				t.Fatalf("expected masked field card")
			}
			if delta.FieldCard.Card.Name == "秘密牌" || delta.FieldCard.Card.ID == "secret" {
				t.Fatalf("field delta leaked cover card front: %+v", delta.FieldCard.Card)
			}
		}
	}
}
