package server

import (
	"encoding/json"
	"strings"
	"testing"

	"starcup-engine/internal/engine"
	"starcup-engine/internal/model"
)

func TestBuildStateForPlayer_UsesIndicatorsForDerivedDisplayState(t *testing.T) {
	room := NewRoom("INDICATORS")
	room.Engine = engine.NewGameEngine(room)
	room.Started = true

	if err := room.Engine.AddPlayer("p1", "Alice", "magic_lancer", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	if err := room.Engine.AddPlayer("p2", "Bob", "angel", model.BlueCamp); err != nil {
		t.Fatal(err)
	}

	p1 := room.Engine.State.Players["p1"]
	p2 := room.Engine.State.Players["p2"]
	p1.Tokens["mb_charge_count"] = 99
	p1.Tokens["hero_anger"] = 2
	p1.Field = append(p1.Field,
		&model.FieldCard{
			Card:    model.Card{ID: "charge-1", Name: "充能A", Type: model.CardTypeAttack, Element: model.ElementFire},
			OwnerID: p1.ID,
			Mode:    model.FieldCover,
			Effect:  model.EffectMagicBowCharge,
		},
		&model.FieldCard{
			Card:    model.Card{ID: "soul-1", Name: "剑魂A", Type: model.CardTypeAttack, Element: model.ElementWater},
			OwnerID: p1.ID,
			Mode:    model.FieldCover,
			Effect:  model.EffectSwordSoul,
		},
	)
	p2.Field = append(p2.Field, &model.FieldCard{
		Card:     model.Card{ID: "shared-life-1", Name: "同生共死", Type: model.CardTypeMagic, Element: model.ElementDark},
		OwnerID:  p2.ID,
		SourceID: p1.ID,
		Mode:     model.FieldEffect,
		Effect:   model.EffectBloodSharedLife,
	})
	room.Engine.ApplyNextAttackDamageRule(p1.ID, "ml_dark_release_next_attack_bonus", "ml_dark_release", 1, model.RuleLifeUntilTurnEnd)
	room.Engine.ApplySkillGateRule(p1.ID, "ml_dark_release_lock_turn", "ml_dark_release", []string{"ml_fullness"}, model.RuleLifeUntilTurnEnd)

	view := room.buildStateForPlayer("p1").Players["p1"]
	if got := view.Tokens["hero_anger"]; got != 2 {
		t.Fatalf("expected real token to remain in tokens, got %d", got)
	}
	if _, ok := view.Tokens["mb_charge_count"]; ok {
		t.Fatalf("expected derived mirror to be removed from tokens, got %+v", view.Tokens)
	}

	expectedIndicators := map[string]int{
		"mb_charge_count":                   1,
		"se_sword_soul_count":               1,
		"bp_shared_life_active":             1,
		"ml_dark_release_next_attack_bonus": 1,
		"ml_dark_release_lock_turn":         1,
	}
	for key, want := range expectedIndicators {
		if got := view.Indicators[key]; got != want {
			t.Fatalf("indicator %s: want %d, got %d in %+v", key, want, got, view.Indicators)
		}
	}

	raw, err := json.Marshal(view)
	if err != nil {
		t.Fatalf("marshal view: %v", err)
	}
	text := string(raw)
	for _, legacy := range []string{
		`"mb_charge_count":99`,
		`"ml_dark_release_next_attack_bonus":1`,
		`"ml_dark_release_lock_turn":1`,
		`"se_sword_soul_count":1`,
		`"bp_shared_life_active":1`,
	} {
		if strings.Contains(text, legacy) && !strings.Contains(text, `"indicators"`) {
			t.Fatalf("legacy derived field leaked outside indicators: %s", text)
		}
	}
	var decoded map[string]interface{}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("decode view json: %v", err)
	}
	for _, legacyKey := range []string{
		"elf_blessing_count",
		"mb_charge_count",
		"sc_power_count",
		"mg_dark_moon_count",
		"bt_cocoon_count",
		"bp_shared_life_active",
		"bp_shared_life_bound",
		"ml_dark_release_next_attack_bonus",
		"ml_dark_release_lock_turn",
		"se_sword_soul_count",
	} {
		if _, ok := decoded[legacyKey]; ok {
			t.Fatalf("legacy top-level key %s leaked in JSON: %s", legacyKey, text)
		}
	}
	if _, ok := decoded["indicators"]; !ok {
		t.Fatalf("expected indicators in JSON, got %s", text)
	}
}
