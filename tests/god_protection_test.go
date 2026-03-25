package tests

import (
	"reflect"
	"testing"
	"unsafe"

	"starcup-engine/internal/engine"
	"starcup-engine/internal/model"
)

func getDispatcher(e *engine.GameEngine) *engine.SkillDispatcher {
	value := reflect.ValueOf(e).Elem().FieldByName("dispatcher")
	return reflect.NewAt(value.Type(), unsafe.Pointer(value.UnsafeAddr())).Elem().Interface().(*engine.SkillDispatcher)
}

func TestGodProtectionMitigatesMoraleLossFromMagicDamage(t *testing.T) {
	game := engine.NewGameEngine(nil)
	if err := game.AddPlayer("p1", "Angel", "angel", model.RedCamp); err != nil {
		t.Fatalf("add angel player: %v", err)
	}
	if err := game.AddPlayer("p2", "Berserker", "berserker", model.RedCamp); err != nil {
		t.Fatalf("add victim player: %v", err)
	}

	angel := game.State.Players["p1"]
	victim := game.State.Players["p2"]
	angel.TurnState = model.NewPlayerTurnState()
	angel.Crystal = 2
	game.State.RedMorale = 10

	loss := 3
	ctx := &model.Context{
		Game: game,
		User: victim,
		TriggerCtx: &model.EventContext{
			Type:      model.EventDamage,
			DamageVal: &loss,
		},
		Flags: map[string]bool{
			"IsMagicDamage": true,
		},
		Selections: map[string]interface{}{
			"morale_loss_pending":              true,
			"morale_loss_value":                3,
			"is_magic":                         true,
			"from_damage_draw":                 false,
			"overflow_morale_loss_fixed":       0,
			"discarded_cards":                  []model.Card{},
			"victim_id":                        victim.ID,
			"discard_player_id":                victim.ID,
			"morale_loss_stay_in_turn":         false,
			"morale_loss_is_damage_resolution": false,
		},
	}

	getDispatcher(game).OnTrigger(model.TriggerBeforeMoraleLoss, ctx)
	if game.State.PendingInterrupt == nil {
		t.Fatalf("expected response interrupt")
	}
	if game.State.PendingInterrupt.PlayerID != angel.ID {
		t.Fatalf("expected interrupt for angel player, got %s", game.State.PendingInterrupt.PlayerID)
	}

	if err := game.ConfirmResponseSkill(angel.ID, "god_protection"); err != nil {
		t.Fatalf("confirm response skill: %v", err)
	}
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptChoice {
		t.Fatalf("expected X-choice interrupt after confirm, got %+v", game.State.PendingInterrupt)
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID:   angel.ID,
		Type:       model.CmdSelect,
		Selections: []int{1}, // X=2
	}); err != nil {
		t.Fatalf("resolve god protection X failed: %v", err)
	}

	if loss != 1 {
		t.Fatalf("expected morale loss 1 after mitigation, got %d", loss)
	}
	if angel.Crystal != 0 {
		t.Fatalf("expected angel crystal 0 after mitigation, got %d", angel.Crystal)
	}
	if game.State.RedMorale != 9 {
		t.Fatalf("expected red morale drop to 9 after mitigation, got %d", game.State.RedMorale)
	}
}
