package engine_test

import (
	"testing"

	"starcup-engine/internal/engine"
	"starcup-engine/internal/model"
	"starcup-engine/internal/rules"
	"starcup-engine/internal/testutils"
)

func TestMagicMissileResponseSkillUsesDedicatedTimingContext(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	game.State.Deck = rules.InitDeck()
	if err := game.AddPlayer("p1", "Caster", "berserker", model.RedCamp); err != nil {
		t.Fatalf("add p1 failed: %v", err)
	}
	if err := game.AddPlayer("p2", "Responder", "magical_girl", model.BlueCamp); err != nil {
		t.Fatalf("add p2 failed: %v", err)
	}

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionExecution
	game.State.PlayerOrder = []string{"p1", "p2"}
	p1 := game.State.Players["p1"]
	p2 := game.State.Players["p2"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Hand = []model.Card{
		{ID: "magic-missile", Name: "魔弹", Type: model.CardTypeMagic, Element: model.ElementWater, Damage: 2},
	}
	p2.Hand = []model.Card{
		{ID: "fusion-card", Name: "火焰斩", Type: model.CardTypeAttack, Element: model.ElementFire},
	}

	if err := game.HandleAction(model.PlayerAction{
		PlayerID: "p1",
		Type:     model.CmdMagic,
		CardID:   testutils.PlayableCardID(t, game, "p1", 0),
	}); err != nil {
		t.Fatalf("cast magic missile failed: %v", err)
	}

	intr := game.State.PendingInterrupt
	if intr == nil || intr.Type != model.InterruptResponseSkill {
		t.Fatalf("expected response skill interrupt, got %+v", intr)
	}
	ctx, ok := intr.Context.(*model.Context)
	if !ok || ctx == nil {
		t.Fatalf("expected model.Context on response skill interrupt, got %T", intr.Context)
	}
	if ctx.Timing != model.TimingMagicMissileResponseSkill {
		t.Fatalf("response skill timing = %q, want %q", ctx.Timing, model.TimingMagicMissileResponseSkill)
	}
	if !ctx.MagicMissileResponseSkillPhase() {
		t.Fatalf("magic missile response skill helper should match timing %q", ctx.Timing)
	}
	if ctx.AttackResponsePhase() {
		t.Fatalf("magic missile response skill context must not be treated as attack response")
	}
}
