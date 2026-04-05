package engine

import (
	"testing"

	"starcup-engine/internal/model"
)

func TestEnterDiscardSelection_RequiresMatchingPendingInterrupt(t *testing.T) {
	game := NewGameEngine(nil)

	game.enterDiscardSelection()
	if game.State.Subflow != model.SubflowNone {
		t.Fatalf("expected subflow none without pending interrupt, got %s", game.State.Subflow)
	}

	game.State.PendingInterrupt = &model.Interrupt{Type: model.InterruptChoice, PlayerID: "p1"}
	game.enterDiscardSelection()
	if game.State.Subflow != model.SubflowNone {
		t.Fatalf("expected subflow none for non-discard interrupt, got %s", game.State.Subflow)
	}

	game.State.PendingInterrupt = &model.Interrupt{Type: model.InterruptDiscard, PlayerID: "p1"}
	game.enterDiscardSelection()
	if game.State.Subflow != model.SubflowDiscardSelection {
		t.Fatalf("expected discard-selection subflow, got %s", game.State.Subflow)
	}
}

func TestIsDiscardSelectionActive_RequiresLifecycleConsistency(t *testing.T) {
	game := NewGameEngine(nil)
	game.State.Subflow = model.SubflowDiscardSelection

	if game.isDiscardSelectionActive() {
		t.Fatal("expected inactive discard selection without pending interrupt")
	}

	game.State.PendingInterrupt = &model.Interrupt{Type: model.InterruptChoice, PlayerID: "p1"}
	if game.isDiscardSelectionActive() {
		t.Fatal("expected inactive discard selection for non-discard interrupt")
	}

	game.State.PendingInterrupt = &model.Interrupt{Type: model.InterruptDiscard, PlayerID: "p1"}
	if !game.isDiscardSelectionActive() {
		t.Fatal("expected active discard selection with matching discard interrupt")
	}
}

func TestPopInterrupt_ClearsDiscardSelectionSubflow(t *testing.T) {
	game := NewGameEngine(nil)
	game.State.PendingInterrupt = &model.Interrupt{Type: model.InterruptDiscard, PlayerID: "p1"}
	game.enterDiscardSelection()

	game.PopInterrupt()

	if game.State.PendingInterrupt != nil {
		t.Fatalf("expected no pending interrupt, got %+v", game.State.PendingInterrupt)
	}
	if game.State.Subflow != model.SubflowNone {
		t.Fatalf("expected discard subflow cleared after pop, got %s", game.State.Subflow)
	}
}

func TestDriveDiscardSelectionPhase_DoesNotAutoRepairOnMismatch(t *testing.T) {
	game := NewGameEngine(nil)
	game.State.Subflow = model.SubflowDiscardSelection
	game.State.TurnStage = model.TurnStageActionExecution
	game.State.CombatStage = model.CombatStageNone
	game.State.ActionQueue = []model.QueuedAction{{SourceID: "p1", Type: model.ActionAttack}}
	game.State.PendingDamageQueue = []model.PendingDamage{{SourceID: "p1", TargetID: "p2", Damage: 1}}

	outcome := game.driveDiscardSelectionPhase()
	if outcome != driveUnhandled {
		t.Fatalf("expected driveUnhandled on discard-subflow mismatch, got %v", outcome)
	}
	if game.State.Subflow != model.SubflowDiscardSelection {
		t.Fatalf("expected subflow unchanged, got %s", game.State.Subflow)
	}
	if len(game.State.ActionQueue) != 1 {
		t.Fatalf("expected action queue unchanged, got %d", len(game.State.ActionQueue))
	}
	if len(game.State.PendingDamageQueue) != 1 {
		t.Fatalf("expected pending damage unchanged, got %d", len(game.State.PendingDamageQueue))
	}
}
