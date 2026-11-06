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

// 同一玩家下一轮的行动不能拿到上一轮的 ActionID：服务端按 ActionID 聚合时间线，
// 重复的 ID 会把上一轮的结算（如"月之女神 承受"）并进这一手的战斗链。
func TestNarrativeTrace_ActionIDUniqueAcrossRounds(t *testing.T) {
	game := NewGameEngine(nil)
	_ = game.AddPlayer("p1", "A", "berserker", model.RedCamp)
	_ = game.AddPlayer("p2", "B", "angel", model.BlueCamp)
	if err := game.StartGame(); err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	first := game.State.PlayerOrder[game.State.CurrentTurn]
	for round := 0; round < 2; round++ {
		for range game.State.PlayerOrder {
			pid := game.State.PlayerOrder[game.State.CurrentTurn]
			id := game.beginNarrativeAction("attack", pid).ActionID
			if seen[id] {
				t.Fatalf("round %d: action id %q reused (first actor %s)", round+1, id, first)
			}
			seen[id] = true
			game.State.PendingInterrupt = nil
			game.State.InterruptQueue = nil
			game.NextTurn()
		}
	}
}
