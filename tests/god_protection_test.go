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
	angel.Crystal = 2

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

	if loss != 1 {
		t.Fatalf("expected morale loss 1 after mitigation, got %d", loss)
	}
	if angel.Crystal != 0 {
		t.Fatalf("expected angel crystal 0 after mitigation, got %d", angel.Crystal)
	}
}
