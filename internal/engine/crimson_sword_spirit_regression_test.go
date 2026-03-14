package engine

import (
	"starcup-engine/internal/model"
	"testing"
)

func TestCrimsonFlash_PhaseEndDamageShouldNotStall(t *testing.T) {
	g := NewGameEngine(noopObserver{})
	if err := g.AddPlayer("p1", "CSS", "crimson_sword_spirit", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := g.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Tokens["css_blood"] = 1
	p1.Heal = 0
	g.State.CurrentTurn = 0
	g.State.Phase = model.PhaseExtraAction
	g.State.Deck = []model.Card{
		{ID: "d1", Name: "补1", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 2},
		{ID: "d2", Name: "补2", Type: model.CardTypeAttack, Element: model.ElementWater, Damage: 2},
		{ID: "d3", Name: "补3", Type: model.CardTypeAttack, Element: model.ElementWind, Damage: 2},
	}

	eventCtx := &model.EventContext{
		Type:       model.EventPhaseEnd,
		SourceID:   p1.ID,
		ActionType: model.ActionAttack,
		AttackInfo: &model.AttackEventInfo{ActionType: string(model.ActionAttack), CounterInitiator: ""},
	}
	ctx := g.buildContext(p1, nil, model.TriggerOnPhaseEnd, eventCtx)
	g.dispatcher.OnTrigger(model.TriggerOnPhaseEnd, ctx)

	if g.State.PendingInterrupt == nil || g.State.PendingInterrupt.Type != model.InterruptResponseSkill {
		t.Fatalf("expected response prompt for css flash, got %+v", g.State.PendingInterrupt)
	}
	if err := g.ConfirmResponseSkill("p1", "css_crimson_flash"); err != nil {
		t.Fatalf("confirm response failed: %v", err)
	}

	if len(g.State.PendingDamageQueue) != 1 {
		t.Fatalf("expected 1 pending damage from crimson flash, got %d", len(g.State.PendingDamageQueue))
	}
	if g.State.PendingDamageQueue[0].Damage != 2 || g.State.PendingDamageQueue[0].TargetID != "p1" {
		t.Fatalf("unexpected pending damage: %+v", g.State.PendingDamageQueue[0])
	}

	// Drive should resolve pending damage and should not return to response phase (would stall).
	g.Drive()

	if g.State.Phase == model.PhaseResponse {
		t.Fatalf("phase stuck in response after crimson flash: return_phase=%s", g.State.ReturnPhase)
	}
	if len(g.State.PendingDamageQueue) != 0 {
		t.Fatalf("expected pending damage drained, got %d", len(g.State.PendingDamageQueue))
	}
	if got := len(p1.Hand); got != 2 {
		t.Fatalf("expected exactly 2 self-damage draw from crimson flash, got hand=%d", got)
	}
}

func TestCrimsonFlash_CombatFlow_DealsExactlyTwoAndKeepsTurnProgressing(t *testing.T) {
	g := NewGameEngine(noopObserver{})
	if err := g.AddPlayer("p1", "CSS", "crimson_sword_spirit", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := g.State.Players["p1"]
	p2 := g.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p2.TurnState = model.NewPlayerTurnState()
	p1.Heal = 0
	p2.Heal = 0
	p1.Hand = []model.Card{
		{ID: "atk1", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 2},
	}
	g.State.CurrentTurn = 0
	g.State.Phase = model.PhaseActionSelection
	g.State.Deck = []model.Card{
		{ID: "d1", Name: "补1", Type: model.CardTypeAttack, Element: model.ElementFire, Damage: 2},
		{ID: "d2", Name: "补2", Type: model.CardTypeAttack, Element: model.ElementWater, Damage: 2},
		{ID: "d3", Name: "补3", Type: model.CardTypeAttack, Element: model.ElementWind, Damage: 2},
		{ID: "d4", Name: "补4", Type: model.CardTypeAttack, Element: model.ElementThunder, Damage: 2},
		{ID: "d5", Name: "补5", Type: model.CardTypeAttack, Element: model.ElementEarth, Damage: 2},
	}

	if err := g.HandleAction(model.PlayerAction{
		PlayerID:  "p1",
		Type:      model.CmdAttack,
		TargetID:  "p2",
		CardIndex: 0,
	}); err != nil {
		t.Fatalf("attack failed: %v", err)
	}
	if err := g.HandleAction(model.PlayerAction{
		PlayerID:  "p2",
		Type:      model.CmdRespond,
		ExtraArgs: []string{"take"},
	}); err != nil {
		t.Fatalf("target take failed: %v", err)
	}

	if g.State.PendingInterrupt == nil || g.State.PendingInterrupt.Type != model.InterruptResponseSkill {
		t.Fatalf("expected response skill prompt after attack end, got %+v", g.State.PendingInterrupt)
	}
	if err := g.ConfirmResponseSkill("p1", "css_crimson_flash"); err != nil {
		t.Fatalf("confirm crimson flash failed: %v", err)
	}

	// 触发后应仅追加一次“对自己2点法术伤害”。
	if len(g.State.PendingDamageQueue) != 1 || g.State.PendingDamageQueue[0].Damage != 2 || g.State.PendingDamageQueue[0].TargetID != "p1" {
		t.Fatalf("unexpected pending damages after crimson flash: %+v", g.State.PendingDamageQueue)
	}

	g.Drive()

	if g.State.Phase == model.PhaseResponse {
		t.Fatalf("phase stuck in response after crimson flash in combat flow")
	}
	if len(g.State.PendingDamageQueue) != 0 {
		t.Fatalf("expected pending damage drained, got %d", len(g.State.PendingDamageQueue))
	}
	if got := len(p1.Hand); got != 2 {
		t.Fatalf("expected CSS hand=2 after self damage draw, got %d", got)
	}
}
