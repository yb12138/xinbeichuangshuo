package engine

import (
	"testing"

	"starcup-engine/internal/model"
)

func TestBerserkerTear_CanTimingOnCounterAttackHit(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Berserker", "berserker", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.TurnState = model.NewPlayerTurnState()
	p1.Gem = 1

	damage := 2
	hitCtx := game.buildContext(p1, p2, model.TimingOnHitCheck, &model.EventContext{
		Type:      model.EventAttack,
		SourceID:  p1.ID,
		TargetID:  p2.ID,
		DamageVal: &damage,
		Card: &model.Card{
			ID:      "counter-atk",
			Name:    "火焰斩",
			Type:    model.CardTypeAttack,
			Element: model.ElementFire,
			Damage:  2,
		},
		AttackInfo: &model.AttackEventInfo{
			ActionType:       string(model.ActionAttack),
			IsHit:            true,
			CounterInitiator: p1.ID,
		},
	})

	game.dispatcher.OnTiming(hitCtx.Timing, hitCtx)
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptResponseSkill {
		t.Fatalf("expected tear response interrupt on counter hit, got %+v", game.State.PendingInterrupt)
	}
	if !interruptHasSkillID(game.State.PendingInterrupt, "berserker_tear") {
		t.Fatalf("expected berserker_tear in response skill list, got %+v", game.State.PendingInterrupt.SkillIDs)
	}

	if err := game.ConfirmResponseSkill("p1", "berserker_tear"); err != nil {
		t.Fatalf("confirm tear on counter hit failed: %v", err)
	}
	if damage != 4 {
		t.Fatalf("expected tear to add 2 damage on counter hit, got %d", damage)
	}
	if p1.Gem != 0 {
		t.Fatalf("expected tear to consume 1 gem, got %d", p1.Gem)
	}
}

func TestBloodBlade_RunsOnHitCheckForActiveUniqueAttack(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Berserker", "berserker", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()
	p2.Hand = []model.Card{
		{ID: "h1", Name: "牌1", Type: model.CardTypeMagic, Element: model.ElementWater},
		{ID: "h2", Name: "牌2", Type: model.CardTypeMagic, Element: model.ElementFire},
	}

	damage := 2
	hitCtx := game.buildContext(p1, p2, model.TimingOnHitCheck, &model.EventContext{
		Type:      model.EventAttack,
		SourceID:  p1.ID,
		TargetID:  p2.ID,
		DamageVal: &damage,
		Card: &model.Card{
			ID:              "blood-blade-card",
			Name:            "血影狂刀",
			Type:            model.CardTypeAttack,
			Element:         model.ElementDark,
			Damage:          2,
			ExclusiveChar1:  "berserker",
			ExclusiveSkill1: "血影狂刀",
		},
		AttackInfo: &model.AttackEventInfo{
			ActionType:       string(model.ActionAttack),
			IsHit:            true,
			CounterInitiator: "",
		},
	})

	game.dispatcher.OnTiming(hitCtx.Timing, hitCtx)
	if damage != 4 {
		t.Fatalf("expected blood_blade to add 2 damage when target hand=2, got %d", damage)
	}
	if game.State.PendingInterrupt != nil {
		t.Fatalf("blood_blade should resolve silently, got %+v", game.State.PendingInterrupt)
	}
}
