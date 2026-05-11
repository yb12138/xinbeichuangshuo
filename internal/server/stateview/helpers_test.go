package stateview_test

import (
	"testing"

	"starcup-engine/internal/model"
	"starcup-engine/internal/server/stateview"
)

func TestIsSkillBlockedBySkillGate(t *testing.T) {
	// 空玩家
	if stateview.IsSkillBlockedBySkillGate(nil, "ml_fullness") {
		t.Fatal("nil player should return false")
	}
	// 空技能ID
	p := &model.Player{ActiveRuleModifiers: map[string]*model.RuleModifierInstance{}}
	if stateview.IsSkillBlockedBySkillGate(p, "") {
		t.Fatal("empty skillID should return false")
	}
	// 无规则
	if stateview.IsSkillBlockedBySkillGate(p, "ml_fullness") {
		t.Fatal("no modifiers should return false")
	}

	// 有SkillGate规则锁定技能
	p.ActiveRuleModifiers["ml_dark_release_lock_turn"] = &model.RuleModifierInstance{
		Domain: model.RuleModifierDomainSkillGate,
		SkillGatePayload: &model.RuleSkillGatePayload{
			Mode:     model.SkillGateDisallowList,
			SkillIDs: []string{"ml_fullness", "ml_black_spear"},
		},
	}

	// 被锁定的技能
	if !stateview.IsSkillBlockedBySkillGate(p, "ml_fullness") {
		t.Fatal("ml_fullness should be blocked")
	}
	if !stateview.IsSkillBlockedBySkillGate(p, "ml_black_spear") {
		t.Fatal("ml_black_spear should be blocked")
	}

	// 未被锁定的技能
	if stateview.IsSkillBlockedBySkillGate(p, "ml_dark_release") {
		t.Fatal("ml_dark_release should NOT be blocked")
	}
	if stateview.IsSkillBlockedBySkillGate(p, "ml_phantom_stardust") {
		t.Fatal("ml_phantom_stardust should NOT be blocked")
	}
}

func TestGetBlockedSkillIDsBySkillGate(t *testing.T) {
	// 空玩家
	if ids := stateview.GetBlockedSkillIDsBySkillGate(nil); ids != nil {
		t.Fatalf("nil player should return nil, got %v", ids)
	}

	// 无规则
	p := &model.Player{ActiveRuleModifiers: map[string]*model.RuleModifierInstance{}}
	if ids := stateview.GetBlockedSkillIDsBySkillGate(p); ids != nil {
		t.Fatalf("no modifiers should return nil, got %v", ids)
	}

	// 有SkillGate规则
	p.ActiveRuleModifiers["ml_dark_release_lock_turn"] = &model.RuleModifierInstance{
		Domain: model.RuleModifierDomainSkillGate,
		SkillGatePayload: &model.RuleSkillGatePayload{
			Mode:     model.SkillGateDisallowList,
			SkillIDs: []string{"ml_fullness", "ml_black_spear"},
		},
	}

	ids := stateview.GetBlockedSkillIDsBySkillGate(p)
	if len(ids) != 2 {
		t.Fatalf("expected 2 blocked skills, got %d", len(ids))
	}
	foundFullness := false
	foundBlackSpear := false
	for _, id := range ids {
		if id == "ml_fullness" {
			foundFullness = true
		}
		if id == "ml_black_spear" {
			foundBlackSpear = true
		}
	}
	if !foundFullness || !foundBlackSpear {
		t.Fatal("expected to find ml_fullness and ml_black_spear in blocked list")
	}
}
