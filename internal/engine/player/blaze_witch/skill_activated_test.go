package blaze_witch_test

import (
	"testing"

	"starcup-engine/internal/engine"
	"starcup-engine/internal/model"
	"starcup-engine/internal/testutils"
)

func TestBlazeWitchRebirthClock_EmitsSkillActivatedOnlyWhenRebirthGained(t *testing.T) {
	obs := &testutils.CaptureObserver{}
	game := engine.NewGameEngine(obs)
	if err := game.AddPlayer("p1", "Blaze", "blaze_witch", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	p1 := game.State.Players["p1"]
	ctx := &model.Context{Flags: map[string]bool{"FromDamageDraw": true, "IsMagicDamage": true}}

	p1.Hand = makeBlazeWitchTestCards(8)
	game.CheckHandLimitCtx(p1, ctx)
	if err := game.ConfirmDiscard("p1", []int{0, 1}); err != nil {
		t.Fatalf("confirm discard failed: %v", err)
	}
	testutils.RequireSingleSkillActivation(t, obs, "bw_rebirth_clock", "p1", "p1")

	// 已达上限：重生不变，不应再次发出事件。
	p1.Tokens["bw_rebirth"] = 4
	p1.Hand = makeBlazeWitchTestCards(8)
	game.CheckHandLimitCtx(p1, ctx)
	if err := game.ConfirmDiscard("p1", []int{0, 1}); err != nil {
		t.Fatalf("confirm discard at cap failed: %v", err)
	}
	testutils.RequireSingleSkillActivation(t, obs, "bw_rebirth_clock", "p1", "p1")
}
