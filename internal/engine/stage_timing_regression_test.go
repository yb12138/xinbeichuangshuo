package engine

import (
	"testing"

	"starcup-engine/internal/model"
)

func TestStartupSkill_WindowSeparatedFromTurnStartTiming(t *testing.T) {
	game := NewGameEngine(nil)
	if err := game.AddPlayer("p1", "Assassin", "assassin", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	game.State.CurrentTurn = 0
	game.State.TurnStage = model.TurnStageActionStart
	p1 := game.State.Players["p1"]
	p1.IsActive = true
	p1.TurnState = model.NewPlayerTurnState()
	p1.Gem = 1

	eventCtx := &model.EventContext{
		Type:     model.EventTurnStart,
		SourceID: "p1",
	}

	turnStartCtx := game.buildTimedContext(p1, nil, model.TimingOnTurnStart, eventCtx)
	game.dispatcher.OnTiming(turnStartCtx.Timing, turnStartCtx)
	if game.State.PendingInterrupt != nil {
		t.Fatalf("expected no startup interrupt on turn-start timing, got %+v", game.State.PendingInterrupt)
	}

	startupCtx := game.buildTimedContext(p1, nil, model.TimingStartup, eventCtx)
	game.dispatcher.OnTiming(startupCtx.Timing, startupCtx)
	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptStartupSkill {
		t.Fatalf("expected startup interrupt on startup timing, got %+v", game.State.PendingInterrupt)
	}
	if len(game.State.PendingInterrupt.SkillIDs) != 1 || game.State.PendingInterrupt.SkillIDs[0] != "stealth" {
		t.Fatalf("expected startup skill stealth, got %+v", game.State.PendingInterrupt.SkillIDs)
	}
}
