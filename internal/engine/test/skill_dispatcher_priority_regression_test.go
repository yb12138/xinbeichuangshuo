package engine_test

import (
	"starcup-engine/internal/engine"
	"starcup-engine/internal/testutils"
	"testing"

	"starcup-engine/internal/model"
)

func setCharacterSkillPriority(player *model.Player, skillID string, priority int) bool {
	if player == nil || player.Character == nil {
		return false
	}
	for i := range player.Character.Skills {
		if player.Character.Skills[i].ID == skillID {
			player.Character.Skills[i].Priority = priority
			return true
		}
	}
	return false
}

func TestBeforeMoraleLoss_UsesSkillPriorityOrdering(t *testing.T) {
	game := engine.NewGameEngine(testutils.NoopObserver{})
	if err := game.AddPlayer("p1", "AngelA", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "AngelB", "angel", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p3", "Enemy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	angelA := game.State.Players["p1"]
	angelB := game.State.Players["p2"]
	angelA.Crystal = 2
	angelB.Crystal = 2

	if !setCharacterSkillPriority(angelA, "god_protection", 10) {
		t.Fatalf("missing god_protection on p1")
	}
	if !setCharacterSkillPriority(angelB, "god_protection", 90) {
		t.Fatalf("missing god_protection on p2")
	}

	loss := 2
	lossCtx := game.BuildContext(angelA, game.State.Players["p3"], model.TimingBeforeMoraleLoss, &model.EventContext{
		Type:      model.EventDamage,
		SourceID:  "p3",
		TargetID:  "p1",
		DamageVal: &loss,
	})
	lossCtx.Flags["IsMagicDamage"] = true

	game.Dispatcher().OnTiming(lossCtx.Timing, lossCtx)

	if game.State.PendingInterrupt == nil || game.State.PendingInterrupt.Type != model.InterruptResponseSkill {
		t.Fatalf("expected response interrupt, got %+v", game.State.PendingInterrupt)
	}
	if game.State.PendingInterrupt.PlayerID != "p2" {
		t.Fatalf("expected higher-priority angel p2 first, got %s", game.State.PendingInterrupt.PlayerID)
	}
	if len(game.State.InterruptQueue) == 0 {
		t.Fatalf("expected lower-priority response queued")
	}
	if game.State.InterruptQueue[0].PlayerID != "p1" {
		t.Fatalf("expected p1 queued after p2, got %s", game.State.InterruptQueue[0].PlayerID)
	}
}
