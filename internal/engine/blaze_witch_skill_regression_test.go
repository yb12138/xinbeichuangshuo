package engine

import (
	"fmt"
	"testing"

	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
)

func makeBlazeWitchTestCards(n int) []model.Card {
	elements := []model.Element{
		model.ElementFire,
		model.ElementWater,
		model.ElementWind,
		model.ElementThunder,
		model.ElementEarth,
		model.ElementDark,
		model.ElementLight,
	}
	cards := make([]model.Card, 0, n)
	for i := 0; i < n; i++ {
		cardType := model.CardTypeAttack
		if i%2 == 0 {
			cardType = model.CardTypeMagic
		}
		cards = append(cards, model.Card{
			ID:      fmt.Sprintf("bw_test_%d", i),
			Name:    fmt.Sprintf("测试牌%d", i+1),
			Type:    cardType,
			Element: elements[i%len(elements)],
			Faction: "血",
			Damage:  2,
		})
	}
	return cards
}

func TestBlazeWitchPainLink_ConsumesCrystalOnceAndQueuesDiscardToThree(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Blaze", "blaze_witch", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Crystal = 1
	p1.Hand = makeBlazeWitchTestCards(5)

	game.State.Deck = rules.InitDeck()
	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection

	if err := game.UseSkill("p1", "bw_pain_link", []string{"p2"}, nil); err != nil {
		t.Fatalf("use pain link failed: %v", err)
	}
	if p1.Crystal != 0 || p1.Gem != 0 {
		t.Fatalf("expected consume exactly 1 crystal-like resource, got crystal=%d gem=%d", p1.Crystal, p1.Gem)
	}
	if got := len(game.State.PendingDamageQueue); got != 2 {
		t.Fatalf("expected 2 pending damages, got %d", got)
	}

	game.State.Phase = model.PhasePendingDamageResolution
	for i := 0; i < 16 && game.State.PendingInterrupt == nil && len(game.State.PendingDamageQueue) > 0; i++ {
		game.processPendingDamages()
	}

	if game.State.PendingInterrupt == nil {
		t.Fatalf("expected discard interrupt from pain link")
	}
	if game.State.PendingInterrupt.Type != model.InterruptDiscard || game.State.PendingInterrupt.PlayerID != "p1" {
		t.Fatalf("unexpected interrupt: %+v", game.State.PendingInterrupt)
	}
	data, _ := game.State.PendingInterrupt.Context.(map[string]interface{})
	discardCount, _ := data["discard_count"].(int)
	if discardCount != len(p1.Hand)-3 {
		t.Fatalf("expected discard_count=len(hand)-3, got discard=%d hand=%d", discardCount, len(p1.Hand))
	}
	if discardCount <= 0 {
		t.Fatalf("expected positive discard count after pain link, got %d", discardCount)
	}
}

func TestBlazeWitchHeavenfireCleave_AllowsNonFireAttackDiscardInFlameForm(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Blaze", "blaze_witch", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Tokens["bw_flame_form"] = 1
	p1.Tokens["bw_rebirth"] = 1
	p1.Hand = []model.Card{
		{ID: "a1", Name: "风神斩", Type: model.CardTypeAttack, Element: model.ElementWind, Faction: "血", Damage: 2},
		{ID: "a2", Name: "雷光斩", Type: model.CardTypeAttack, Element: model.ElementThunder, Faction: "血", Damage: 2},
	}

	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection

	if err := game.UseSkill("p1", "bw_heavenfire_cleave", []string{"p2"}, []int{0, 1}); err != nil {
		t.Fatalf("heavenfire cleave should accept transformed fire discards in flame form, got: %v", err)
	}
	if got := p1.Tokens["bw_rebirth"]; got != 1 {
		t.Fatalf("expected rebirth not consumed in flame form, got %d", got)
	}
	if got := len(game.State.PendingDamageQueue); got != 2 {
		t.Fatalf("expected 2 pending damages, got %d", got)
	}
	if game.State.PendingDamageQueue[0].Damage != 3 || game.State.PendingDamageQueue[1].Damage != 3 {
		t.Fatalf("expected heavenfire base damage 3 when morale not behind, got %+v", game.State.PendingDamageQueue)
	}
}

func TestBlazeWitchRebirthClock_IncreasesOnMagicMoraleLossWithCap(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Blaze", "blaze_witch", model.RedCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p1.Hand = makeBlazeWitchTestCards(8)

	damageOverflowCtx := &model.Context{
		Flags: map[string]bool{
			"FromDamageDraw": true,
			"IsMagicDamage":  true,
		},
	}
	game.checkHandLimit(p1, damageOverflowCtx)
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptDiscard {
		t.Fatalf("expected discard interrupt, got %+v", game.State.PendingInterrupt)
	}
	if err := game.ConfirmDiscard("p1", []int{0, 1}); err != nil {
		t.Fatalf("confirm discard failed: %v", err)
	}
	if got := p1.Tokens["bw_rebirth"]; got != 1 {
		t.Fatalf("expected rebirth +1 after magic morale loss, got %d", got)
	}

	p1.Tokens["bw_rebirth"] = 4
	p1.Hand = makeBlazeWitchTestCards(8)
	game.checkHandLimit(p1, damageOverflowCtx)
	if err := game.ConfirmDiscard("p1", []int{0, 1}); err != nil {
		t.Fatalf("confirm discard at cap failed: %v", err)
	}
	if got := p1.Tokens["bw_rebirth"]; got != 4 {
		t.Fatalf("expected rebirth capped at 4, got %d", got)
	}
}

func TestBlazeWitchFlameForm_ReleasesAtStartup(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Blaze", "blaze_witch", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Tokens["bw_flame_form"] = 1
	p1.Tokens["bw_flame_release_pending"] = 1
	p1.Hand = makeBlazeWitchTestCards(5)

	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseStartup

	game.Drive()

	if got := p1.Tokens["bw_flame_form"]; got != 0 {
		t.Fatalf("expected flame form released at startup, got %d", got)
	}
	if got := p1.Tokens["bw_flame_release_pending"]; got != 0 {
		t.Fatalf("expected flame release flag cleared, got %d", got)
	}
}

func TestBlazeWitchGetMaxHand_DynamicByRebirthInFlameForm(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Blaze", "blaze_witch", model.RedCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	if got := game.GetMaxHand(p1); got != 6 {
		t.Fatalf("expected base max hand 6, got %d", got)
	}

	p1.Tokens["bw_flame_form"] = 1
	p1.Tokens["bw_rebirth"] = 0
	if got := game.GetMaxHand(p1); got != 4 {
		t.Fatalf("expected max hand 4 when rebirth=0 in flame form, got %d", got)
	}
	p1.Tokens["bw_rebirth"] = 1
	if got := game.GetMaxHand(p1); got != 5 {
		t.Fatalf("expected max hand 5 when rebirth=1 in flame form, got %d", got)
	}
	p1.Tokens["bw_rebirth"] = 3
	if got := game.GetMaxHand(p1); got != 7 {
		t.Fatalf("expected max hand 7 when rebirth=3 in flame form, got %d", got)
	}
}
