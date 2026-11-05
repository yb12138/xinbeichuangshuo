package crimson_sword_spirit_test

import (
	"testing"

	"starcup-engine/internal/engine"
	"starcup-engine/internal/model"
	"starcup-engine/internal/testutils"
)

func setupCrimsonDance(t *testing.T) (*engine.GameEngine, *model.Player) {
	t.Helper()
	g := engine.NewGameEngine(testutils.NoopObserver{})
	if err := g.AddPlayer("p1", "CSS", "crimson_sword_spirit", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	p1 := g.State.Players["p1"]
	p1.TurnState = model.NewPlayerTurnState()
	p1.ExclusiveCards = append(p1.ExclusiveCards, model.Card{
		ID:              "css_test_courtyard",
		Name:            "血蔷薇庭院",
		Type:            model.CardTypeMagic,
		Element:         model.ElementDark,
		ExclusiveChar1:  p1.Character.ID,
		ExclusiveSkill1: "血蔷薇庭院",
	})
	return g, p1
}

// 血色剑灵-1：宝石分支本次 +2 最高到 4，但鲜血上限仍为 3，之后血色荆棘不能再往上加。
func TestCrimsonDance_GemBranchOnlyTemporarilyExceedsCap(t *testing.T) {
	g, p1 := setupCrimsonDance(t)
	p1.Gem = 1
	p1.Crystal = 0
	p1.Tokens["css_blood"] = 3
	g.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: "p1",
		Context: map[string]interface{}{
			"choice_type": "css_dance_mode",
			"user_id":     "p1",
			"can_crystal": false,
			"can_gem":     true,
		},
	})
	if err := g.HandleInterruptAction(model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{0}}); err != nil {
		t.Fatalf("dance gem branch failed: %v", err)
	}
	if got := p1.Tokens["css_blood"]; got != 4 {
		t.Fatalf("expected blood 4 after gem branch, got %d", got)
	}
	if got := p1.Tokens["css_blood_cap"]; got != 3 {
		t.Fatalf("expected blood cap stays 3, got %d", got)
	}
	if !hasFieldEffect(p1, model.EffectRoseCourtyard) {
		t.Fatalf("expected courtyard placed")
	}

	g.RunTurnEndTimingStageHooks(p1, engine.TimingTurnEndFinal)
	if got := p1.Tokens["css_blood"]; got != 4 {
		t.Fatalf("expected blood not trimmed at turn end, got %d", got)
	}
}

// 血色剑灵-2：扣费失败时不放置庭院，专属卡仍在。
func TestCrimsonDance_PaymentFailureDoesNotPlaceCourtyard(t *testing.T) {
	g, p1 := setupCrimsonDance(t)
	p1.Gem = 0
	p1.Crystal = 0
	g.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: "p1",
		Context: map[string]interface{}{
			"choice_type": "css_dance_mode",
			"user_id":     "p1",
			"can_crystal": true,
			"can_gem":     false,
		},
	})
	if err := g.HandleInterruptAction(model.PlayerAction{PlayerID: "p1", Type: model.CmdSelect, Selections: []int{0}}); err == nil {
		t.Fatalf("expected dance to fail without crystal")
	}
	if hasFieldEffect(p1, model.EffectRoseCourtyard) {
		t.Fatalf("courtyard must not be placed when payment fails")
	}
	if !p1.HasExclusiveCard(p1.Character.ID, "血蔷薇庭院") {
		t.Fatalf("courtyard card must remain in exclusive zone")
	}
}
