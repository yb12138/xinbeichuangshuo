package engine

import (
	"testing"

	skillhandlers "starcup-engine/internal/engine/skill"
	"starcup-engine/internal/model"
)

const (
	testRuleTimingNoopID      = "test_rule_timing_noop"
	testRuleTimingInterruptID = "test_rule_timing_interrupt"
)

type ruleTimingNoopRecorder struct{}

type ruleTimingNoopObserver struct{}

func (ruleTimingNoopObserver) OnGameEvent(model.GameEvent) {}

func (ruleTimingNoopRecorder) CanUse(ctx *model.Context) bool {
	return ctx != nil && ctx.Selections["rulebook_timing"] == model.TimingMagicResolve
}

func (ruleTimingNoopRecorder) Execute(ctx *model.Context) error { return nil }

type ruleTimingInterruptRecorder struct{}

func (ruleTimingInterruptRecorder) CanUse(ctx *model.Context) bool {
	return ctx != nil && ctx.Selections["rulebook_timing"] == model.TimingMagicResolve
}

func (ruleTimingInterruptRecorder) Execute(ctx *model.Context) error {
	ctx.Game.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: ctx.User.ID,
		Context: map[string]interface{}{
			"choice_type": "test_rule_timing_interrupt",
		},
	})
	return nil
}

func init() {
	skillhandlers.Register(testRuleTimingNoopID, ruleTimingNoopRecorder{})
	skillhandlers.Register(testRuleTimingInterruptID, ruleTimingInterruptRecorder{})
}

func TestDispatchRuleTimingDoesNotPauseForExistingPendingInterrupt(t *testing.T) {
	game, player := newRuleTimingDispatchTestGame(t, testRuleTimingNoopID)
	game.PushInterrupt(&model.Interrupt{Type: model.InterruptChoice, PlayerID: player.ID})
	revisionBefore := game.State.InterruptRevision

	result := game.dispatchRuleTiming(ruleTimingDispatchInput{
		Timing: model.TimingMagicResolve,
		User:   player,
	})

	if !result.Dispatched {
		t.Fatalf("expected timing to dispatch")
	}
	if result.PendingChanged || result.Interrupted || result.QueuedInterrupt {
		t.Fatalf("expected existing pending interrupt not to count as new pause, got %+v", result)
	}
	if game.State.InterruptRevision != revisionBefore {
		t.Fatalf("interrupt revision changed from %d to %d", revisionBefore, game.State.InterruptRevision)
	}
}

func TestDispatchRuleTimingReportsNewInterrupt(t *testing.T) {
	game, player := newRuleTimingDispatchTestGame(t, testRuleTimingInterruptID)

	result := game.dispatchRuleTiming(ruleTimingDispatchInput{
		Timing: model.TimingMagicResolve,
		User:   player,
	})

	if !result.Dispatched {
		t.Fatalf("expected timing to dispatch")
	}
	if !result.PendingChanged || !result.Interrupted {
		t.Fatalf("expected new pending interrupt, got %+v", result)
	}
	if result.QueuedInterrupt {
		t.Fatalf("first interrupt should become pending, not queued: %+v", result)
	}
	if game.State.PendingInterrupt == nil {
		t.Fatalf("expected pending interrupt")
	}
}

func TestDispatchRuleTimingReportsQueuedInterrupt(t *testing.T) {
	game, player := newRuleTimingDispatchTestGame(t, testRuleTimingInterruptID)
	game.PushInterrupt(&model.Interrupt{Type: model.InterruptChoice, PlayerID: player.ID})

	result := game.dispatchRuleTiming(ruleTimingDispatchInput{
		Timing: model.TimingMagicResolve,
		User:   player,
	})

	if !result.Dispatched {
		t.Fatalf("expected timing to dispatch")
	}
	if !result.PendingChanged || !result.Interrupted || !result.QueuedInterrupt {
		t.Fatalf("expected queued interrupt to be reported, got %+v", result)
	}
	if len(game.State.InterruptQueue) != 1 {
		t.Fatalf("expected one queued interrupt, got %d", len(game.State.InterruptQueue))
	}
}

func TestInterruptRevisionChangesOnSetEnqueueAndPop(t *testing.T) {
	game, player := newRuleTimingDispatchTestGame(t, testRuleTimingNoopID)
	initial := game.State.InterruptRevision

	game.PushInterrupt(&model.Interrupt{Type: model.InterruptChoice, PlayerID: player.ID})
	afterPending := game.State.InterruptRevision
	if afterPending <= initial {
		t.Fatalf("push pending should increment revision: before=%d after=%d", initial, afterPending)
	}

	game.PushInterrupt(&model.Interrupt{Type: model.InterruptChoice, PlayerID: player.ID})
	afterQueue := game.State.InterruptRevision
	if afterQueue <= afterPending {
		t.Fatalf("queue interrupt should increment revision: before=%d after=%d", afterPending, afterQueue)
	}

	game.PopInterrupt()
	afterPop := game.State.InterruptRevision
	if afterPop <= afterQueue {
		t.Fatalf("pop interrupt should increment revision: before=%d after=%d", afterQueue, afterPop)
	}
}

func newRuleTimingDispatchTestGame(t *testing.T, skillID string) (*GameEngine, *model.Player) {
	t.Helper()
	game := NewGameEngine(ruleTimingNoopObserver{})
	if err := game.AddPlayer("p1", "Tester", "berserker", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	player := game.State.Players["p1"]
	player.Character.Skills = append(player.Character.Skills, model.SkillDefinition{
		ID:           skillID,
		Title:        skillID,
		Type:         model.SkillTypePassive,
		ResponseType: model.ResponseSilent,
		LogicHandler: skillID,
		Timings:      []model.FlowTiming{model.TimingMagicResolve},
	})
	return game, player
}
