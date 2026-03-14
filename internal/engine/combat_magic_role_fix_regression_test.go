package engine

import (
	"testing"

	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
)

func TestAssassinStealthAttack_NoCounterAndBonusDamage(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Assassin", "assassin", model.RedCamp); err != nil {
		t.Fatalf("add p1 failed: %v", err)
	}
	if err := game.AddPlayer("p2", "Target", "berserker", model.BlueCamp); err != nil {
		t.Fatalf("add p2 failed: %v", err)
	}

	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection
	game.State.Deck = rules.InitDeck()
	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()
	p1.Gem = 1
	p1.Crystal = 2 // X=3
	p1.AddFieldCard(&model.FieldCard{
		Card: model.Card{
			ID:   "stealth-fc",
			Name: "潜行",
			Type: model.CardTypeMagic,
		},
		OwnerID:  p1.ID,
		SourceID: p1.ID,
		Mode:     model.FieldEffect,
		Effect:   model.EffectStealth,
		Trigger:  model.EffectTriggerManual,
	})
	p1.Hand = []model.Card{
		{ID: "atk1", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 1},
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdAttack,
		TargetID:  "p2",
		CardIndex: 0,
	}); err != nil {
		t.Fatalf("attack failed: %v", err)
	}

	if len(game.State.CombatStack) != 1 {
		t.Fatalf("expected combat stack size 1, got %d", len(game.State.CombatStack))
	}
	req := game.State.CombatStack[0]
	if req.CanBeResponded {
		t.Fatalf("stealth attack should be non-counterable")
	}

	// 目标承受伤害，期望伤害=基础1 + 剩余能量3 = 4
	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p2",
		Type:     model.CmdRespond,
		ExtraArgs: []string{
			"take",
		},
	}); err != nil {
		t.Fatalf("respond take failed: %v", err)
	}

	if got := len(p2.Hand); got != 4 {
		t.Fatalf("expected target draw 4 cards from stealth bonus attack, got %d", got)
	}
}

func TestSaintessFrostPrayer_PromptTargetAndHeal(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Saintess", "saintess", model.RedCamp); err != nil {
		t.Fatalf("add p1 failed: %v", err)
	}
	if err := game.AddPlayer("p2", "Ally", "angel", model.RedCamp); err != nil {
		t.Fatalf("add p2 failed: %v", err)
	}
	if err := game.AddPlayer("p3", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatalf("add p3 failed: %v", err)
	}

	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection
	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{
		{ID: "m1", Name: "水之印", Type: model.CardTypeMagic, Element: model.ElementWater, Damage: 0},
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdMagic,
		TargetID:  "p3",
		CardIndex: 0,
	}); err != nil {
		t.Fatalf("magic failed: %v", err)
	}

	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptChoice {
		t.Fatalf("expected frost prayer choice interrupt, got %+v", game.State.PendingInterrupt)
	}
	ctx, _ := game.State.PendingInterrupt.Context.(map[string]interface{})
	if ctx == nil || ctx["choice_type"] != "frost_prayer_target" {
		t.Fatalf("expected frost_prayer_target choice, got %+v", ctx)
	}

	targetIDs := make([]string, 0)
	if arr, ok := ctx["target_ids"].([]string); ok {
		targetIDs = append(targetIDs, arr...)
	} else if arr, ok := ctx["target_ids"].([]interface{}); ok {
		for _, v := range arr {
			if s, ok := v.(string); ok {
				targetIDs = append(targetIDs, s)
			}
		}
	}
	selectIdx := -1
	for i, tid := range targetIDs {
		if tid == "p2" {
			selectIdx = i
			break
		}
	}
	if selectIdx < 0 {
		t.Fatalf("ally target p2 not found in frost prayer options: %+v", targetIDs)
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{selectIdx},
	}); err != nil {
		t.Fatalf("select frost prayer target failed: %v", err)
	}

	if p2.Heal != 1 {
		t.Fatalf("expected ally heal +1 from frost prayer, got %d", p2.Heal)
	}
}

func TestMagicalGirlMagicBulletFusion_ActionSkill(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Girl", "magical_girl", model.RedCamp); err != nil {
		t.Fatalf("add p1 failed: %v", err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatalf("add p2 failed: %v", err)
	}

	game.State.CurrentTurn = 0
	game.State.Phase = model.PhaseActionSelection
	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{
		{ID: "f1", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 2},
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSkill,
		SkillID:    "magic_bullet_fusion",
		Selections: []int{0},
	}); err != nil {
		t.Fatalf("magic_bullet_fusion failed: %v", err)
	}

	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptMagicBulletDirection {
		t.Fatalf("expected magic bullet direction prompt after fusion, got %+v", game.State.PendingInterrupt)
	}
}
