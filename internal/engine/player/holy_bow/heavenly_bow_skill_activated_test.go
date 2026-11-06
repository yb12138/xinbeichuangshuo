package holy_bow_test

import (
	"testing"

	"starcup-engine/internal/engine"
	"starcup-engine/internal/model"
	"starcup-engine/internal/testutils"
)

func TestHolyBow_HeavenlyBowFaithGainEmitsSkillActivated(t *testing.T) {
	obs := &testutils.CaptureObserver{}
	game := engine.NewGameEngine(obs)
	if err := game.AddPlayer("p1", "HolyBow", "holy_bow", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := game.AddPlayer("p2", "Enemy", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	nonHoly := holyBowTestCard("atk2", "火斩", model.CardTypeAttack, model.ElementFire, 2)
	nonHoly.Faction = "幻"
	game.HandlePostAttackHitEffects(&model.PendingDamage{
		SourceID: "p1", TargetID: "p2", Damage: 2, DamageType: model.AttackDamage, Card: &nonHoly,
	})
	if n := len(obs.SkillActivated("hb_heavenly_bow")); n != 0 {
		t.Fatalf("non-holy hit should not emit, got %d", n)
	}

	holy := holyBowTestCard("atk1", "圣斩", model.CardTypeAttack, model.ElementLight, 2)
	holy.Faction = "圣"
	game.HandlePostAttackHitEffects(&model.PendingDamage{
		SourceID: "p1", TargetID: "p2", Damage: 2, DamageType: model.AttackDamage, Card: &holy,
	})
	testutils.RequireSingleSkillActivation(t, obs, "hb_heavenly_bow", "p1", "p1")
}
