package engine

import (
	"testing"

	"starcup-engine/internal/model"
)

func interruptHasSkillID(intr *model.Interrupt, skillID string) bool {
	if intr == nil {
		return false
	}
	for _, sid := range intr.SkillIDs {
		if sid == skillID {
			return true
		}
	}
	return false
}

func TestAngelSong_TriggersAsTurnStartResponseAndResumesActionSelection(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Angel", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionStart

	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()
	p1.Crystal = 1
	p2.AddFieldCard(&model.FieldCard{
		Card:     model.Card{ID: "weak-1", Name: "虚弱", Type: model.CardTypeMagic, Element: model.ElementWind},
		OwnerID:  p2.ID,
		SourceID: "p3",
		Mode:     model.FieldEffect,
		Effect:   model.EffectWeak,
		Trigger:  model.EffectTriggerOnBeforeAction,
	})

	game.Drive()

	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptResponseSkill {
		t.Fatalf("expected angel_song response interrupt at startup, got %+v", game.State.PendingInterrupt)
	}
	if !interruptHasSkillID(game.State.PendingInterrupt, "angel_song") {
		t.Fatalf("expected angel_song in pending skills, got %+v", game.State.PendingInterrupt.SkillIDs)
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0},
	}); err != nil {
		t.Fatalf("confirm angel_song failed: %v", err)
	}
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptChoice {
		t.Fatalf("expected basic effect choice after angel_song, got %+v", game.State.PendingInterrupt)
	}
	ctxData, _ := game.State.PendingInterrupt.Context.(map[string]interface{})
	if got, _ := ctxData["choice_type"].(string); got != "basic_effect_pick" {
		t.Fatalf("expected basic_effect_pick, got %q", got)
	}
	if p1.Crystal != 0 {
		t.Fatalf("angel_song should consume 1 crystal before selection, got %d", p1.Crystal)
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0},
	}); err != nil {
		t.Fatalf("resolve angel_song basic effect pick failed: %v", err)
	}
	if got := countFieldEffect(p2, model.EffectWeak); got != 0 {
		t.Fatalf("expected weakness removed by angel_song, got %d", got)
	}
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptChoice {
		t.Fatalf("expected angel bond follow-up choice after angel_song removal, got %+v", game.State.PendingInterrupt)
	}
	ctxData, _ = game.State.PendingInterrupt.Context.(map[string]interface{})
	if got, _ := ctxData["choice_type"].(string); got != "angel_bond_heal_target" {
		t.Fatalf("expected angel_bond_heal_target, got %q", got)
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0},
	}); err != nil {
		t.Fatalf("resolve angel bond after angel_song failed: %v", err)
	}
	if game.State.PendingInterrupt != nil {
		t.Fatalf("expected no pending interrupt after angel bond resolution, got %+v", game.State.PendingInterrupt)
	}
	if game.State.TurnStage != model.TurnStageActionExecution || game.State.CombatStage != model.CombatStageNone || game.State.Subflow != model.SubflowNone {
		t.Fatalf("expected turn to continue to action execution window, got turn=%s combat=%s subflow=%s", game.State.TurnStage, game.State.CombatStage, game.State.Subflow)
	}
}

func TestGodProtection_PromptsForXAndPartiallyMitigatesMoraleLoss(t *testing.T) {
	game := NewGameEngine(noopObserver{})
	if err := game.AddPlayer("p1", "Angel", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.RedMorale = 10

	p1 := game.State.Players["p1"]
	p1.TurnState = model.NewPlayerTurnState()
	p1.Crystal = 3

	moraleLoss := 3
	lossCtx := game.buildContext(p1, nil, model.TriggerBeforeMoraleLoss, &model.EventContext{
		Type:      model.EventDamage,
		DamageVal: &moraleLoss,
	})
	lossCtx.Flags["IsMagicDamage"] = true
	lossCtx.Selections = map[string]any{
		"morale_loss_pending":              true,
		"morale_loss_value":                3,
		"is_magic":                         true,
		"from_damage_draw":                 false,
		"overflow_morale_loss_fixed":       0,
		"discarded_cards":                  []model.Card{},
		"victim_id":                        p1.ID,
		"discard_player_id":                p1.ID,
		"morale_loss_stay_in_turn":         false,
		"morale_loss_is_damage_resolution": false,
	}

	game.dispatcher.OnTrigger(model.TriggerBeforeMoraleLoss, lossCtx)
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptResponseSkill {
		t.Fatalf("expected god_protection response interrupt, got %+v", game.State.PendingInterrupt)
	}
	if !interruptHasSkillID(game.State.PendingInterrupt, "god_protection") {
		t.Fatalf("expected god_protection in pending skills, got %+v", game.State.PendingInterrupt.SkillIDs)
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{0},
	}); err != nil {
		t.Fatalf("confirm god_protection failed: %v", err)
	}
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptChoice {
		t.Fatalf("expected X choice after god_protection confirmation, got %+v", game.State.PendingInterrupt)
	}
	ctxData, _ := game.State.PendingInterrupt.Context.(map[string]interface{})
	if got, _ := ctxData["choice_type"].(string); got != "god_protection_x" {
		t.Fatalf("expected god_protection_x, got %q", got)
	}
	if moraleLoss != 3 || p1.Crystal != 3 {
		t.Fatalf("god_protection should wait for X selection before consuming resources, moraleLoss=%d crystal=%d", moraleLoss, p1.Crystal)
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:   "p1",
		Type:       model.CmdSelect,
		Selections: []int{1},
	}); err != nil {
		t.Fatalf("resolve god_protection X failed: %v", err)
	}
	if moraleLoss != 1 {
		t.Fatalf("expected morale loss reduced to 1 after choosing X=2, got %d", moraleLoss)
	}
	if p1.Crystal != 1 {
		t.Fatalf("expected 2 crystal consumed, got crystal=%d", p1.Crystal)
	}
	if game.State.RedMorale != 9 {
		t.Fatalf("expected red morale 10 -> 9 after partial mitigation, got %d", game.State.RedMorale)
	}
	if game.State.PendingInterrupt != nil {
		t.Fatalf("expected no pending interrupt after god_protection resolution, got %+v", game.State.PendingInterrupt)
	}
}
