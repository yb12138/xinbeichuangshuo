package engine

import (
	"strings"
	"testing"

	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
)

func TestOnmyojiShikigamiDescend_RequiresSameFactionDiscards(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Onmyoji", "onmyoji", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{
		{ID: "c1", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Faction: "咏", Damage: 2},
		{ID: "c2", Name: "圣光", Type: model.CardTypeMagic, Element: model.ElementLight, Faction: "圣", Damage: 0},
		{ID: "c3", Name: "风神斩", Type: model.CardTypeAttack, Element: model.ElementWind, Faction: "咏", Damage: 2},
	}

	err := game.UseSkill("p1", "onmyoji_shikigami_descend", nil, []int{0, 1})
	if err == nil || !strings.Contains(err.Error(), "命格相同") {
		t.Fatalf("expected same-faction discard error, got %v", err)
	}

	if err := game.UseSkill("p1", "onmyoji_shikigami_descend", nil, []int{0, 2}); err != nil {
		t.Fatalf("use skill with same-faction discards failed: %v", err)
	}

	if got := p1.Tokens["onmyoji_form"]; got != 1 {
		t.Fatalf("expected onmyoji_form=1, got %d", got)
	}
	if got := p1.Tokens["onmyoji_ghost_fire"]; got != 1 {
		t.Fatalf("expected onmyoji_ghost_fire=1, got %d", got)
	}
	if got := len(p1.Hand); got != 1 {
		t.Fatalf("expected hand size 1 after discarding 2, got %d", got)
	}
	if got := len(game.State.DiscardPile); got != 2 {
		t.Fatalf("expected discard pile size 2, got %d", got)
	}
	if len(p1.TurnState.PendingActions) == 0 || p1.TurnState.PendingActions[0].MustType != "Attack" {
		t.Fatalf("expected extra attack action from shikigami descend, got %+v", p1.TurnState.PendingActions)
	}
}

func TestOnmyojiYinYangShift_InShikigamiForm(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Attacker", "berserker", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Onmyoji", "onmyoji", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "RedAlly", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.Deck = rules.InitDeck()
	game.State.Phase = model.PhaseActionSelection

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p3 := game.State.Players["p3"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()
	p3.TurnState = model.NewPlayerTurnState()
	p2.Tokens["onmyoji_form"] = 1
	p2.Tokens["onmyoji_ghost_fire"] = 1

	p1.Hand = []model.Card{
		{ID: "atk1", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Faction: "咏", Damage: 2},
	}
	p2.Hand = []model.Card{
		{ID: "atk2", Name: "水涟斩", Type: model.CardTypeAttack, Element: model.ElementWater, Faction: "咏", Damage: 2},
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdAttack,
		TargetID:  "p2",
		CardIndex: 0,
	}); err != nil {
		t.Fatalf("attack failed: %v", err)
	}

	if game.State.PendingInterrupt == nil || choiceTypeOf(game.State.PendingInterrupt) != "onmyoji_yinyang_confirm" {
		t.Fatalf("expected onmyoji_yinyang_confirm prompt, got %+v", game.State.PendingInterrupt)
	}
	if err := game.handleWeakChoiceInput("p2", 0); err != nil {
		t.Fatalf("confirm yinyang failed: %v", err)
	}
	if got := choiceTypeOf(game.State.PendingInterrupt); got != "onmyoji_yinyang_card" {
		t.Fatalf("expected onmyoji_yinyang_card prompt, got %s", got)
	}
	if err := game.handleWeakChoiceInput("p2", 0); err != nil {
		t.Fatalf("choose yinyang card failed: %v", err)
	}
	if got := choiceTypeOf(game.State.PendingInterrupt); got != "onmyoji_yinyang_counter_target" {
		t.Fatalf("expected onmyoji_yinyang_counter_target prompt, got %s", got)
	}
	if err := game.handleWeakChoiceInput("p2", 0); err != nil {
		t.Fatalf("choose yinyang counter target failed: %v", err)
	}

	if got := p2.Tokens["onmyoji_ghost_fire"]; got != 3 {
		t.Fatalf("expected ghost fire=3 after 阴阳转换+式神转换, got %d", got)
	}
	if got := p2.Tokens["onmyoji_form"]; got != 0 {
		t.Fatalf("expected leave shikigami form, got onmyoji_form=%d", got)
	}
	if got := len(p2.Hand); got != 1 {
		t.Fatalf("expected hand size 1 (counter consume 1 then draw 1), got %d", got)
	}

	if len(game.State.CombatStack) == 0 {
		t.Fatalf("expected reflected combat request in stack")
	}
	top := game.State.CombatStack[len(game.State.CombatStack)-1]
	if top.AttackerID != "p2" || top.TargetID != "p3" {
		t.Fatalf("expected reflected combat p2->p3, got %+v", top)
	}
	if top.Card == nil {
		t.Fatalf("expected reflected combat card not nil")
	}
	if top.Card.Damage != 3 {
		t.Fatalf("expected reflected damage = ghost fire(3), got %d", top.Card.Damage)
	}
	if top.Card.Element != model.ElementWater {
		t.Fatalf("expected reflected element converted to card element Water, got %s", top.Card.Element)
	}
}
