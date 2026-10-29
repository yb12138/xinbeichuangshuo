package holy_lancer_test

import (
	"testing"

	"starcup-engine/internal/engine"
	"starcup-engine/internal/model"
	"starcup-engine/internal/testutils"
)

func TestHolyLancerRevelation_EmitsSkillActivatedWhenMaxHealRestored(t *testing.T) {
	obs := &testutils.CaptureObserver{}
	game := engine.NewGameEngine(obs)
	if err := game.AddPlayer("p1", "HolyLancer", "holy_lancer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Dummy", "berserker", model.BlueCamp); err != nil {
		t.Fatal(err)
	}
	if n := len(obs.SkillActivated("holy_lancer_revelation")); n != 0 {
		t.Fatalf("setup sync should not emit, got %d", n)
	}

	// 对方先拿星杯：治疗上限 3→2，神圣启示失效，不发事件
	if !game.AddCampCup(model.BlueCamp) {
		t.Fatalf("blue cup gain failed")
	}
	if n := len(obs.SkillActivated("holy_lancer_revelation")); n != 0 {
		t.Fatalf("losing revelation should not emit, got %d", n)
	}

	// 我方追平：治疗上限 2→3，神圣启示生效
	if !game.AddCampCup(model.RedCamp) {
		t.Fatalf("red cup gain failed")
	}
	if got := game.State.Players["p1"].MaxHeal; got != 3 {
		t.Fatalf("expected max heal 3, got %d", got)
	}
	testutils.RequireSingleSkillActivation(t, obs, "holy_lancer_revelation", "p1", "p1")

	// 再拿一杯：上限保持 3，不重复发事件
	if !game.AddCampCup(model.RedCamp) {
		t.Fatalf("red cup gain failed")
	}
	testutils.RequireSingleSkillActivation(t, obs, "holy_lancer_revelation", "p1", "p1")
}
