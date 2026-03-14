package engine

import (
	"testing"

	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
)

func countFieldEffects(p *model.Player, effect model.EffectType) int {
	if p == nil {
		return 0
	}
	count := 0
	for _, fc := range p.Field {
		if fc != nil && fc.Mode == model.FieldEffect && fc.Effect == effect {
			count++
		}
	}
	return count
}

func TestDebugCheat_EffectSetCount(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "A", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "B", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	if p1 == nil {
		t.Fatal("p1 not found")
	}
	game.State.Deck = rules.InitDeck()

	if err := game.handleCheat(model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdCheat,
		TargetID: "effect",
		ExtraArgs: []string{
			"p1", "Shield", "2",
		},
	}); err != nil {
		t.Fatalf("set shield effect failed: %v", err)
	}
	if got := countFieldEffects(p1, model.EffectShield); got != 2 {
		t.Fatalf("expected shield count=2, got=%d", got)
	}

	if err := game.handleCheat(model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdCheat,
		TargetID: "effect",
		ExtraArgs: []string{
			"p1", "Shield", "0",
		},
	}); err != nil {
		t.Fatalf("clear shield effect failed: %v", err)
	}
	if got := countFieldEffects(p1, model.EffectShield); got != 0 {
		t.Fatalf("expected shield count=0 after clear, got=%d", got)
	}
}

func TestDebugCheat_CardFiltersAndExclusive(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "A", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "B", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	p1 := game.State.Players["p1"]
	if p1 == nil {
		t.Fatal("p1 not found")
	}
	game.State.Deck = rules.InitDeck()
	p1.Hand = nil

	if err := game.handleCheat(model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdCheat,
		TargetID: "card_element",
		ExtraArgs: []string{
			"p1", "Fire", "2",
		},
	}); err != nil {
		t.Fatalf("card_element failed: %v", err)
	}
	if len(p1.Hand) != 2 {
		t.Fatalf("expected 2 cards after card_element, got=%d", len(p1.Hand))
	}
	for i, c := range p1.Hand {
		if c.Element != model.ElementFire {
			t.Fatalf("card %d expected Fire element, got=%s", i, c.Element)
		}
	}

	p1.Hand = nil
	if err := game.handleCheat(model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdCheat,
		TargetID: "card_faction",
		ExtraArgs: []string{
			"p1", "圣", "2",
		},
	}); err != nil {
		t.Fatalf("card_faction failed: %v", err)
	}
	if len(p1.Hand) != 2 {
		t.Fatalf("expected 2 cards after card_faction, got=%d", len(p1.Hand))
	}
	for i, c := range p1.Hand {
		if c.Faction != "圣" {
			t.Fatalf("card %d expected faction=圣, got=%s", i, c.Faction)
		}
	}

	p1.Hand = nil
	if err := game.handleCheat(model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdCheat,
		TargetID: "card_magic",
		ExtraArgs: []string{
			"p1", "魔弹", "1",
		},
	}); err != nil {
		t.Fatalf("card_magic failed: %v", err)
	}
	if len(p1.Hand) != 1 {
		t.Fatalf("expected 1 card after card_magic, got=%d", len(p1.Hand))
	}
	if p1.Hand[0].Type != model.CardTypeMagic || p1.Hand[0].Name != "魔弹" {
		t.Fatalf("expected magic card 魔弹, got type=%s name=%s", p1.Hand[0].Type, p1.Hand[0].Name)
	}

	char := game.debugFindCharacter("blade_master")
	if char == nil || len(char.Skills) == 0 {
		t.Fatal("blade_master character or skills not found")
	}
	var skill model.SkillDefinition
	foundSkill := false
	for _, s := range char.Skills {
		if s.ID == "gale_skill" {
			skill = s
			foundSkill = true
			break
		}
	}
	if !foundSkill {
		t.Fatal("blade_master.gale_skill not found")
	}
	p1.Hand = nil
	available := 0
	for _, c := range game.State.Deck {
		if c.MatchExclusive(char.Name, skill.Title) {
			available++
		}
	}
	if available == 0 {
		t.Fatalf("precondition failed: no deck cards match [%s·%s]", char.Name, skill.Title)
	}
	beforeStock := len(game.State.Deck) + len(game.State.DiscardPile)
	if err := game.handleCheat(model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdCheat,
		TargetID: "card_exclusive",
		ExtraArgs: []string{
			"p1", char.ID, skill.ID, "1",
		},
	}); err != nil {
		t.Fatalf("card_exclusive failed: %v", err)
	}
	if len(p1.Hand) != 1 {
		t.Fatalf("expected 1 card after card_exclusive, got=%d", len(p1.Hand))
	}
	if !p1.Hand[0].MatchExclusive(char.Name, skill.Title) {
		t.Fatalf("unexpected exclusive mark: char1=%s skill1=%s char2=%s skill2=%s",
			p1.Hand[0].ExclusiveChar1, p1.Hand[0].ExclusiveSkill1, p1.Hand[0].ExclusiveChar2, p1.Hand[0].ExclusiveSkill2)
	}
	if p1.Hand[0].Description == "调试专属牌" {
		t.Fatalf("expected card_exclusive to draw real card from stock, got synthetic debug card")
	}
	afterStock := len(game.State.Deck) + len(game.State.DiscardPile)
	if afterStock != beforeStock-1 {
		t.Fatalf("expected deck/discard stock reduce by 1 after card_exclusive, before=%d after=%d", beforeStock, afterStock)
	}
}
