package engine

import (
	"testing"

	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
)

func TestAngelBond_OnlyTriggersWhenAngelIsRemovalSource(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Angel", "angel", model.RedCamp); err != nil {
		t.Fatalf("add p1 failed: %v", err)
	}
	if err := game.AddPlayer("p2", "Friend", "berserker", model.RedCamp); err != nil {
		t.Fatalf("add p2 failed: %v", err)
	}
	if err := game.AddPlayer("p3", "Enemy", "archer", model.BlueCamp); err != nil {
		t.Fatalf("add p3 failed: %v", err)
	}

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	game.State.Phase = model.PhaseActionSelection

	// 非天使移除基础效果：不应触发天使羁绊。
	p2.AddFieldCard(&model.FieldCard{Mode: model.FieldEffect, Effect: model.EffectWeak, SourceID: "p3"})
	game.RemoveFieldCardBy("p2", model.EffectWeak, "p3")
	if game.State.PendingInterrupt != nil {
		t.Fatalf("angel bond should not trigger when remover is not angel, got: %+v", game.State.PendingInterrupt)
	}

	// 天使本人移除基础效果：应触发天使羁绊选择框。
	p2.AddFieldCard(&model.FieldCard{Mode: model.FieldEffect, Effect: model.EffectWeak, SourceID: "p3"})
	game.RemoveFieldCardBy("p2", model.EffectWeak, "p1")
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptChoice {
		t.Fatalf("angel bond should trigger when angel removes basic effect, got: %+v", game.State.PendingInterrupt)
	}
}

func TestBloodRoar_ForcedHitIgnoresShield(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Berserker", "berserker", model.RedCamp); err != nil {
		t.Fatalf("add p1 failed: %v", err)
	}
	if err := game.AddPlayer("p2", "Target", "angel", model.BlueCamp); err != nil {
		t.Fatalf("add p2 failed: %v", err)
	}

	game.State.Deck = rules.InitDeck()
	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()
	p2.Heal = 2 // 满足血腥咆哮触发条件
	p2.AddFieldCard(&model.FieldCard{
		Card: model.Card{
			ID:      "shield-fc",
			Name:    "圣盾",
			Type:    model.CardTypeMagic,
			Element: model.ElementLight,
		},
		OwnerID:  p2.ID,
		SourceID: p2.ID,
		Mode:     model.FieldEffect,
		Effect:   model.EffectShield,
		Trigger:  model.EffectTriggerOnDamaged,
	})
	p1.Hand = []model.Card{
		{
			ID:              "blood-roar-card",
			Name:            "血腥咆哮",
			Type:            model.CardTypeAttack,
			Element:         model.ElementFire,
			Damage:          2,
			ExclusiveChar1:  "狂战士",
			ExclusiveSkill1: "血腥咆哮",
		},
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdAttack,
		TargetID:  "p2",
		CardIndex: 0,
	}); err != nil {
		t.Fatalf("attack failed: %v", err)
	}

	// 血腥咆哮应无视圣盾并造成伤害（狂化基础+1 => 至少摸3张）。
	if got := len(p2.Hand); got < 3 {
		t.Fatalf("expected blood roar damage to land despite shield, target hand=%d", got)
	}
	if !p2.HasFieldEffect(model.EffectShield) {
		t.Fatalf("shield should remain because blood roar ignores shield instead of consuming it")
	}
}

func TestSealBreak_SelectSpecificBasicEffectAndTakeCard(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Sealer", "sealer", model.RedCamp); err != nil {
		t.Fatalf("add p1 failed: %v", err)
	}
	if err := game.AddPlayer("p2", "Target", "angel", model.BlueCamp); err != nil {
		t.Fatalf("add p2 failed: %v", err)
	}

	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.Crystal = 1
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()

	p2.AddFieldCard(&model.FieldCard{
		Card: model.Card{
			ID:      "shield-card",
			Name:    "圣盾",
			Type:    model.CardTypeMagic,
			Element: model.ElementLight,
		},
		OwnerID:  p2.ID,
		SourceID: "p3",
		Mode:     model.FieldEffect,
		Effect:   model.EffectShield,
		Trigger:  model.EffectTriggerOnDamaged,
	})
	p2.AddFieldCard(&model.FieldCard{
		Card: model.Card{
			ID:              "seal-fire-card",
			Name:            "火焰斩",
			Type:            model.CardTypeAttack,
			Element:         model.ElementFire,
			ExclusiveChar1:  "封印师",
			ExclusiveSkill1: "火之封印",
		},
		OwnerID:  p2.ID,
		SourceID: "p3",
		Mode:     model.FieldEffect,
		Effect:   model.EffectSealFire,
		Trigger:  model.EffectTriggerOnAttack,
	})

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdSkill,
		SkillID:   "seal_break",
		TargetIDs: []string{"p2"},
	}); err != nil {
		t.Fatalf("seal_break failed: %v", err)
	}

	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptChoice {
		t.Fatalf("expected seal_break effect choice interrupt, got: %+v", game.State.PendingInterrupt)
	}

	// 选择第二张基础效果（火之封印）回收。
	if err := game.HandleAction(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{1},
	}); err != nil {
		t.Fatalf("seal_break choice failed: %v", err)
	}

	if len(p1.Hand) != 1 {
		t.Fatalf("expected sealer to gain 1 card, got %d", len(p1.Hand))
	}
	if p1.Hand[0].ID != "seal-fire-card" {
		t.Fatalf("expected taken card to be selected field card, got %+v", p1.Hand[0])
	}
	if len(p2.Field) != 1 || p2.Field[0].Effect != model.EffectShield {
		t.Fatalf("expected only shield to remain on target field, got %+v", p2.Field)
	}
}
