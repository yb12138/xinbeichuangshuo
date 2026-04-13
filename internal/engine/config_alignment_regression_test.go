package engine

import (
	"strings"
	"testing"

	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
)

func TestPendingDamage_PoisonDoesNotConsumeHolyShield(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "A", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "B", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p1.Field = []*model.FieldCard{{
		Card:     model.Card{ID: "shield", Name: "圣盾", Type: model.CardTypeMagic, Element: model.ElementLight},
		OwnerID:  p1.ID,
		SourceID: p1.ID,
		Mode:     model.FieldEffect,
		Effect:   model.EffectShield,
		Hook:     model.FieldHookOnDamaged,
	}}
	game.State.PendingDamageQueue = []model.PendingDamage{{
		SourceID:   "p2",
		TargetID:   "p1",
		Damage:     1,
		DamageType: "poison",
		Card:       &model.Card{ID: "poison", Name: "中毒", Type: model.CardTypeMagic, Element: model.ElementEarth},
	}}
	game.State.CombatStage = model.CombatStageCalcDamage

	for i := 0; i < 8 && len(game.State.PendingDamageQueue) > 0 && game.State.PendingInterrupt == nil; i++ {
		game.processPendingDamages()
	}
	if game.State.PendingInterrupt != nil {
		t.Fatalf("poison damage should not create an interrupt here, got %+v", game.State.PendingInterrupt)
	}
	if got := countFieldEffect(p1, model.EffectShield); got != 1 {
		t.Fatalf("holy shield should remain after poison damage, got %d", got)
	}
}

func TestAngelCleanse_CanPickSpecificBasicEffect(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Angel", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.Deck = rules.InitDeck()
	game.State.TurnStage = model.TurnStageActionExecution

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{{ID: "wind-1", Name: "风牌", Type: model.CardTypeMagic, Element: model.ElementWind}}
	p2.AddFieldCard(&model.FieldCard{Card: model.Card{ID: "weak", Name: "虚弱"}, OwnerID: p2.ID, SourceID: "p3", Mode: model.FieldEffect, Effect: model.EffectWeak, Hook: model.FieldHookOnBeforeAction})
	p2.AddFieldCard(&model.FieldCard{Card: model.Card{ID: "poison", Name: "中毒"}, OwnerID: p2.ID, SourceID: "p3", Mode: model.FieldEffect, Effect: model.EffectPoison, Hook: model.FieldHookOnBeforeAction})

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSkill,
		SkillID:    "angel_cleanse",
		TargetIDs:  []string{"p2"},
		Selections: []int{0},
	}); err != nil {
		t.Fatalf("angel_cleanse should enter basic effect picker, got err=%v", err)
	}
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptChoice {
		t.Fatalf("expected choice interrupt for angel_cleanse, got %+v", game.State.PendingInterrupt)
	}
	ctxData, _ := game.State.PendingInterrupt.Context.(map[string]interface{})
	choiceType, _ := ctxData["choice_type"].(string)
	if choiceType != "basic_effect_pick" {
		t.Fatalf("expected basic_effect_pick interrupt, got %q", choiceType)
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{1},
	}); err != nil {
		t.Fatalf("angel_cleanse choice should succeed, got err=%v", err)
	}
	if got := countFieldEffect(p2, model.EffectWeak); got != 1 {
		t.Fatalf("expected weakness to remain, got %d", got)
	}
	if got := countFieldEffect(p2, model.EffectPoison); got != 0 {
		t.Fatalf("expected poison to be removed, got %d", got)
	}
}

func TestAngelCleanse_NoBasicEffectSkipsRemovalStep(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Angel", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{{ID: "wind-1", Name: "风牌", Type: model.CardTypeMagic, Element: model.ElementWind}}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSkill,
		SkillID:    "angel_cleanse",
		TargetIDs:  []string{"p2"},
		Selections: []int{0},
	}); err != nil {
		t.Fatalf("angel_cleanse should succeed when no basic effect exists, got err=%v", err)
	}
	if game.State.PendingInterrupt != nil {
		t.Fatalf("expected no follow-up choice interrupt when no basic effect exists, got %+v", game.State.PendingInterrupt)
	}
	if len(p1.Hand) != 0 {
		t.Fatalf("wind cleanse should still consume discard, hand=%d", len(p1.Hand))
	}
	if len(game.State.DiscardPile) != 1 {
		t.Fatalf("wind cleanse should place discard into pile, discard=%d", len(game.State.DiscardPile))
	}
}

func TestAngelBlessing_RejectsDuplicateTwoTargetSelection(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Angel", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Target", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution

	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{{ID: "water-1", Name: "水牌", Type: model.CardTypeMagic, Element: model.ElementWater}}

	err := game.HandleAction(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSkill,
		SkillID:    "angel_blessing",
		TargetIDs:  []string{"p2", "p2"},
		Selections: []int{0},
	})
	if err == nil || !strings.Contains(err.Error(), "不能重复选择同一角色") {
		t.Fatalf("expected angel_blessing duplicate target error, got err=%v", err)
	}
	if len(p1.Hand) != 1 {
		t.Fatalf("angel blessing should not consume discard on duplicate targets, hand=%d", len(p1.Hand))
	}
	if len(game.State.DiscardPile) != 0 {
		t.Fatalf("angel blessing should not discard any card on duplicate targets, discard=%d", len(game.State.DiscardPile))
	}
}

func TestAngelBond_IgnoresSystemBuffRemoval(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Angel", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p1.AddFieldCard(&model.FieldCard{
		Card:     model.Card{ID: "poison", Name: "中毒"},
		OwnerID:  p1.ID,
		SourceID: "p2",
		Mode:     model.FieldEffect,
		Effect:   model.EffectPoison,
		Hook:     model.FieldHookOnBeforeAction,
	})

	if !game.RemoveFieldCardBy("p1", model.EffectPoison, "") {
		t.Fatalf("expected poison to be removed")
	}
	if game.State.PendingInterrupt != nil {
		t.Fatalf("system buff removal should not dispatch angel bond, got %+v", game.State.PendingInterrupt)
	}
}

func TestAngelBond_RunsOnAngelWallShieldPlacement(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Angel", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Friend", "saintess", model.RedCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.Heal = 1
	p2.MaxHeal = 5
	p1.Hand = []model.Card{{
		ID:              "wall-1",
		Name:            "天使之墙",
		Type:            model.CardTypeMagic,
		Element:         model.ElementLight,
		ExclusiveChar1:  "angel",
		ExclusiveSkill1: "天使之墙",
	}}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSkill,
		SkillID:    "angel_wall",
		TargetIDs:  []string{"p2"},
		Selections: []int{0},
	}); err != nil {
		t.Fatalf("angel_wall should succeed, got err=%v", err)
	}
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptChoice {
		t.Fatalf("expected angel bond choice after angel_wall, got %+v", game.State.PendingInterrupt)
	}
	if err := game.HandleAction(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{1},
	}); err != nil {
		t.Fatalf("angel bond choice after angel_wall should succeed, got err=%v", err)
	}
	if p2.Heal != 2 {
		t.Fatalf("expected angel wall shield placement to dispatch angel bond heal, got %d", p2.Heal)
	}
}
