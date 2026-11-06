package blade_master_test

import (
	"testing"

	"starcup-engine/internal/engine"
	skills "starcup-engine/internal/engine/skill"
	"starcup-engine/internal/model"
	"starcup-engine/internal/testutils"
)

// 风之剑圣-1：疾风技应战时不触发。
func TestBladeMaster_GaleSkill_NotOnCounter(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "BladeMaster", "blade_master", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	h := skills.GetHandler("gale_skill")
	if h == nil {
		t.Fatalf("gale_skill handler missing")
	}
	p1 := game.State.Players["p1"]
	counter := &model.Context{User: p1, EventCtx: &model.EventContext{AttackInfo: &model.AttackEventInfo{ActionType: string(model.ActionAttack), CounterInitiator: "p1"}}}
	if h.CanUse(counter) {
		t.Fatalf("gale skill must not trigger on counter")
	}
	active := &model.Context{User: p1, EventCtx: &model.EventContext{AttackInfo: &model.AttackEventInfo{ActionType: string(model.ActionAttack)}}}
	if !h.CanUse(active) {
		t.Fatalf("gale skill should trigger on active attack")
	}
}

// 风之剑圣-2：第3次攻击之后的法术行动结束不再弹出圣剑摸X弃X。
func TestBladeMaster_HolySwordDraw_NotOnMagicActionEnd(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "BladeMaster", "blade_master", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	p1 := game.State.Players["p1"]
	p1.TurnState = model.NewPlayerTurnState()
	p1.TurnState.AttackCount = 3
	if game.HandlePostActionEndEffects(p1, model.ActionMagic) {
		t.Fatalf("holy sword draw must not trigger on magic action end, got %+v", game.State.PendingInterrupt)
	}
	if !game.HandlePostActionEndEffects(p1, model.ActionAttack) {
		t.Fatalf("expected holy sword draw on 3rd attack action end")
	}
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptHolySwordDraw {
		t.Fatalf("expected holy sword draw interrupt, got %+v", game.State.PendingInterrupt)
	}
}
