package engine

import (
	"testing"

	"starcup-engine/internal/model"
)

// 回合结束时的被动（圣弓·自动填充等）不能沿用上一个行动的 ActionID，否则会和该行动并到同一条战斗链。
func TestNarrativeTrace_TurnEndStartsNewAction(t *testing.T) {
	game := NewGameEngine(nil)
	_ = game.AddPlayer("p1", "HolyBow", "holy_bow", model.RedCamp)
	_ = game.AddPlayer("p2", "Enemy", "angel", model.BlueCamp)
	game.State.CurrentTurn = 0

	magicID := game.beginNarrativeAction("magic", "p1").ActionID
	if magicID == "" {
		t.Fatal("expected magic action id")
	}
	game.enterTurnEndStage()
	turnEnd := game.currentNarrativeTrace()
	if turnEnd.ActionID == "" || turnEnd.ActionID == magicID {
		t.Fatalf("turn end should open a new narrative action, got %q (magic %q)", turnEnd.ActionID, magicID)
	}
}
