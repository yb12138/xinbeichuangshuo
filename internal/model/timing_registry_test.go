package model

import "testing"

func TestTimingDescriptorOfCoversCurrentFlowTimings(t *testing.T) {
	timings := []Timing{
		TimingUnknown,
		TimingOnGameStart,
		TimingOnCampChanged,
		TimingOnSkillExecuted,
		TimingOnMagicDeclared,
		TimingOnDamageApplied,
		TimingMoraleLossCheck,
		TimingOnHealOverflow,
		TimingOnFieldMarkChanged,
		TimingOnOrientationChanged,
		TimingOnTurnEnd,
	}

	assertTimingDescriptors(t, timings)
}

func TestTimingDescriptorOfCoversCurrentRoleHookTimings(t *testing.T) {
	timings := []Timing{
		"post_action_end",
		"post_attack_hit",
		"post_damage_resolved",
		"on_turn_before_start",
		"on_turn_start",
		"on_turn_end",
		"on_turn_end_final",
		"before_action",
		"on_action_end",
		"on_attack_gating",
		"on_attack_card_hook",
		"on_attack_state_reset",
		"on_attack_target_ctx",
		"on_attack_miss",
		"on_counter_policy",
		"on_defend_validation",
		"on_response_skill_aug",
		"on_response_skill_normalize",
		"on_response_skill_advance",
		"on_response_skill_skip",
		"on_combat_interaction",
		"on_counter_card_policy",
		"on_counter_element_check",
		"on_counter_resolve",
		"on_magic_missile_defend",
		"on_magic_missile_counter",
		"on_magic_missile_response_skill_aug",
		"on_damage_before_taken",
		"on_damage_after_taken",
		"on_damage_applied",
		"on_damage_after_apply",
		"on_heal_resist",
		"on_heal_cap_calculate",
		"on_game_start",
		"on_player_added",
		"on_camp_changed",
		"on_player_setup",
		"on_camp_cup_changed",
		"on_morale_loss_applied",
		"before_action_option",
		"before_action_validation",
		"on_attack_declared_interrupt",
		"on_combat_counter_card",
		"on_after_cannot_act",
		"on_special_action_override",
		"on_special_action_post",
		"on_skill_post",
		"on_attack_card_transform",
	}

	assertTimingDescriptors(t, timings)
}

func TestTimingRegistryRejectsUnknownTiming(t *testing.T) {
	if _, ok := TimingDescriptorOf("missing.timing"); ok {
		t.Fatal("unknown timing should not have a descriptor")
	}
	if got := TimingCategoryOf("missing.timing"); got != TimingCategoryUnknown {
		t.Fatalf("unknown timing category = %q, want %q", got, TimingCategoryUnknown)
	}
}

func TestTimingRegistryCategoriesCurrentTimelines(t *testing.T) {
	tests := map[Timing]TimingCategory{
		TimingTurnStart:             TimingCategoryTurn,
		TimingAttackDeclare:         TimingCategoryAttack,
		TimingOnMagicDeclared:       TimingCategoryMagic,
		TimingMagicMissileResponse:  TimingCategoryMagic,
		TimingDamageTaken:           TimingCategoryDamage,
		TimingMoraleLossCheck:       TimingCategorySettle,
		Timing("on_player_added"):   TimingCategorySystem,
		Timing("on_counter_policy"): TimingCategoryAttack,
	}
	for timing, want := range tests {
		if got := TimingCategoryOf(timing); got != want {
			t.Fatalf("TimingCategoryOf(%q) = %q, want %q", timing, got, want)
		}
	}
}

func TestTimingDescriptorOfCoversRulebookTimings(t *testing.T) {
	timings := []Timing{
		TimingGameInitial,
		TimingTurnBeforeStart,
		TimingTurnStart,
		TimingActionBefore,
		TimingActionStart,
		TimingActionDuring,
		TimingActionEnd,
		TimingActionPost,
		TimingTurnEnd,
		TimingAttackDeclare,
		TimingAttackSelectTarget,
		TimingAttackPlayCard,
		TimingAttackModifyCard,
		TimingAttackCommitted,
		TimingAttackForceHitCheck,
		TimingAttackNoResponseCheck,
		TimingAttackResponse,
		TimingAttackHit,
		TimingAttackMiss,
		TimingMagicDeclare,
		TimingMagicSelectTarget,
		TimingMagicValidate,
		TimingMagicResolve,
		TimingMagicHealOverflow,
		TimingMagicMissileResponse,
		TimingMagicMissileDefend,
		TimingMagicMissileCounter,
		TimingMagicMissileResponseSkill,
		TimingDamageSourceDeal,
		TimingDamageTargetBefore,
		TimingHealBefore,
		TimingHealUse,
		TimingHealCap,
		TimingDamageApplied,
		TimingDamageTaken,
		TimingSettleDraw,
		TimingSettleDiscard,
		TimingCardPlayedRevealed,
		TimingSettleHandLimit,
		TimingMoraleLossCheck,
		TimingMoraleLossApplied,
		TimingGameEndCheck,
		TimingDamageResolved,
	}

	assertTimingDescriptors(t, timings)
}

func assertTimingDescriptors(t *testing.T, timings []Timing) {
	t.Helper()
	for _, timing := range timings {
		desc, ok := TimingDescriptorOf(timing)
		if !ok {
			t.Fatalf("missing descriptor for timing %q", timing)
		}
		if desc.ID != timing {
			t.Fatalf("descriptor ID for %q = %q", timing, desc.ID)
		}
		if desc.RuleName == "" {
			t.Fatalf("descriptor for %q has empty rule name", timing)
		}
	}
}
