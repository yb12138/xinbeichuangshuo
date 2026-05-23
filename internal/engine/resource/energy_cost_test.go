package resource

import (
	"testing"

	"starcup-engine/internal/model"
)

func TestSpendSkillEnergyCost_GemSubstitutesCrystalAfterGemCost(t *testing.T) {
	p := &model.Player{Gem: 2, Crystal: 0}

	if !SpendSkillEnergyCost(p, 1, 1) {
		t.Fatal("expected one remaining gem to pay one crystal cost")
	}
	if p.Gem != 0 || p.Crystal != 0 {
		t.Fatalf("expected both gems spent and no crystals, got gem=%d crystal=%d", p.Gem, p.Crystal)
	}
}

func TestCanPaySkillEnergyCost_CrystalCannotSubstituteGem(t *testing.T) {
	p := &model.Player{Gem: 0, Crystal: 2}

	if CanPaySkillEnergyCost(p, 1, 0) {
		t.Fatal("expected blue crystals to not pay a red-gem cost")
	}
}

func TestSpendCrystalCost_PrefersCrystalBeforeGem(t *testing.T) {
	p := &model.Player{Gem: 1, Crystal: 1}

	if !SpendCrystalCost(p, 2) {
		t.Fatal("expected gem to substitute the second crystal")
	}
	if p.Gem != 0 || p.Crystal != 0 {
		t.Fatalf("expected crystal then gem spent, got gem=%d crystal=%d", p.Gem, p.Crystal)
	}
}
