package engine

import (
	"strings"
	"testing"

	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
)

func elementalistExclusiveCard(owner *model.Player, skillTitle string, element model.Element) model.Card {
	charName := "元素师"
	faction := "咏"
	if owner != nil && owner.Character != nil {
		charName = owner.Character.Name
		faction = owner.Character.Faction
	}
	return model.Card{
		ID:              "elem-exclusive-" + skillTitle,
		Name:            skillTitle,
		Type:            model.CardTypeMagic,
		Element:         element,
		Faction:         faction,
		Damage:          0,
		Description:     "元素师独有技测试卡",
		ExclusiveChar1:  charName,
		ExclusiveSkill1: skillTitle,
	}
}

func TestElementalistFreeze_RequiresTwoTargets(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Elem", "elementalist", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{
		elementalistExclusiveCard(p1, "冰冻", model.ElementFire),
	}
	p2.Hand = nil
	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection
	game.State.Deck = rules.InitDeck()

	err := game.HandleAction(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSkill,
		SkillID:    "elementalist_freeze",
		TargetIDs:  []string{"p2"},
		Selections: []int{0},
	})
	if err == nil || !strings.Contains(err.Error(), "最少需要指定 2 个目标") {
		t.Fatalf("expected freeze single-target rejection, got err=%v", err)
	}

	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSkill,
		SkillID:    "elementalist_freeze",
		TargetIDs:  []string{"p2", "p1"},
		Selections: []int{0},
	})

	if got := p1.Heal; got != 1 {
		t.Fatalf("expected freeze heal target gain 1 heal, got %d", got)
	}
	if got := len(p2.Hand); got != 1 {
		t.Fatalf("expected freeze deal 1 damage (draw 1), got hand=%d", got)
	}
}

func TestElementalistMoonlight_ConsumesGemAndRequiresGem(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Elem", "elementalist", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.Hand = nil
	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection
	game.State.Deck = rules.InitDeck()

	p1.Gem = 0
	p1.Crystal = 3
	err := game.HandleAction(model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdSkill,
		SkillID:   "elementalist_moonlight",
		TargetIDs: []string{"p2"},
	})
	if err == nil || !strings.Contains(err.Error(), "资源不足") {
		t.Fatalf("expected moonlight require gem, got err=%v", err)
	}

	p1.Gem = 1
	p1.Crystal = 2
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdSkill,
		SkillID:   "elementalist_moonlight",
		TargetIDs: []string{"p2"},
	})

	if got := p1.Gem; got != 0 {
		t.Fatalf("expected moonlight consume 1 gem, got %d", got)
	}
	if got := len(p2.Hand); got != 3 {
		t.Fatalf("expected moonlight damage=3 after paying gem (remaining energy 2), got hand=%d", got)
	}
}

func TestElementalistIgnite_RequiresThreeElement(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Elem", "elementalist", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Tokens["element"] = 2
	p2.Hand = nil
	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection
	game.State.Deck = rules.InitDeck()

	err := game.HandleAction(model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdSkill,
		SkillID:   "elementalist_ignite",
		TargetIDs: []string{"p2"},
	})
	if err == nil || !strings.Contains(err.Error(), "元素不足") {
		t.Fatalf("expected ignite reject when element<3, got err=%v", err)
	}

	p1.Tokens["element"] = 3
	mustHandleAction(t, game, model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdSkill,
		SkillID:   "elementalist_ignite",
		TargetIDs: []string{"p2"},
	})

	if got := p1.Tokens["element"]; got != 0 {
		t.Fatalf("expected ignite consume 3 element, got %d", got)
	}
	if got := len(p2.Hand); got != 2 {
		t.Fatalf("expected ignite deal 2 damage (draw 2), got hand=%d", got)
	}
	if p1.TurnState.CurrentExtraAction != "Magic" && len(p1.TurnState.PendingActions) == 0 {
		t.Fatalf("expected ignite grant extra magic action, current=%q pending=%d", p1.TurnState.CurrentExtraAction, len(p1.TurnState.PendingActions))
	}
}
