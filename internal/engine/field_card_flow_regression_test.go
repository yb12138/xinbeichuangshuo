package engine

import (
	"testing"

	soulsorcererpkg "starcup-engine/internal/engine/player/soul_sorcerer"
	"starcup-engine/internal/model"
)

func TestPlaceSoulLink_RejectsRebinding(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Soul", "soul_sorcerer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Ally", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	card := makeStarterSoulLinkCard(p1)
	if err := soulsorcererpkg.PlaceSoulLink(newRoleChoiceRuntime(game), p1, p2, card); err != nil {
		t.Fatalf("place soul link failed: %v", err)
	}

	second := makeStarterSoulLinkCard(p1)
	if err := soulsorcererpkg.PlaceSoulLink(newRoleChoiceRuntime(game), p1, p2, second); err == nil {
		t.Fatalf("expected rebinding soul link to fail")
	}
	holder, fc := soulsorcererpkg.FindSoulLink(newRoleChoiceRuntime(game), p1)
	if holder == nil || fc == nil || holder.ID != p2.ID {
		t.Fatalf("expected original soul link remain on p2, holder=%+v card=%+v", holder, fc)
	}
	if len(game.State.DiscardPile) != 0 {
		t.Fatalf("expected no discard during failed rebinding, discard=%+v", game.State.DiscardPile)
	}
}
