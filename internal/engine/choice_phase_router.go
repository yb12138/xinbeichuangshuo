package engine

import (
	"strings"

	"starcup-engine/internal/model"
)

type choiceResumeResolver func(currentPoint string, ctxData map[string]interface{}) (string, bool)

var choiceInterruptResumeResolvers = buildChoiceInterruptResumeResolvers()

func resolveChoiceInterruptPhase(currentPoint string, ctxData map[string]interface{}) (string, bool) {
	if ctxData == nil {
		return "", false
	}
	choiceType, _ := ctxData["choice_type"].(string)
	if choiceType == "" {
		return "", false
	}
	resolver, ok := choiceInterruptResumeResolvers[choiceType]
	if !ok {
		return "", false
	}
	return resolver(currentPoint, ctxData)
}

func fixedChoiceResumePoint(point interface{}) choiceResumeResolver {
	normalized := model.NormalizeResumePoint(point)
	return func(_ string, _ map[string]interface{}) (string, bool) {
		return normalized, normalized != ""
	}
}

func currentChoiceResumePointResolver() choiceResumeResolver {
	return func(currentPoint string, _ map[string]interface{}) (string, bool) {
		return model.NormalizeResumePoint(currentPoint), currentPoint != ""
	}
}

func waitingChoiceResumePoint(key string) choiceResumeResolver {
	return func(currentPoint string, ctxData map[string]interface{}) (string, bool) {
		if point := normalizeChoiceResumePoint(ctxData[key]); point != "" {
			return point, true
		}
		point := model.NormalizeResumePoint(currentPoint)
		return point, point != ""
	}
}

func waitingChoiceResumePointOr(key string, fallback interface{}) choiceResumeResolver {
	fallbackPoint := model.NormalizeResumePoint(fallback)
	return func(_ string, ctxData map[string]interface{}) (string, bool) {
		if point := normalizeChoiceResumePoint(ctxData[key]); point != "" {
			return point, true
		}
		return fallbackPoint, fallbackPoint != ""
	}
}

func registerFixedChoiceResumePoints(m map[string]choiceResumeResolver, point interface{}, choiceTypes ...string) {
	for _, choiceType := range choiceTypes {
		m[choiceType] = fixedChoiceResumePoint(point)
	}
}

func buildChoiceInterruptResumeResolvers() map[string]choiceResumeResolver {
	m := map[string]choiceResumeResolver{
		"hero_roar_draw":                waitingChoiceResumePoint("waiting_phase"),
		"bw_witch_wrath_draw":           waitingChoiceResumePoint("waiting_phase"),
		"assassin_stealth_draw":         waitingChoiceResumePoint("waiting_phase"),
		"fighter_hundred_dragon_target": waitingChoiceResumePointOr("waiting_phase", model.TurnStageActionStart),
		"priest_divine_contract_target": waitingChoiceResumePointOr("waiting_phase", model.TurnStageActionStart),
		"priest_divine_contract_x":      waitingChoiceResumePointOr("waiting_phase", model.TurnStageActionStart),
		"basic_effect_pick":             waitingChoiceResumePoint("waiting_phase"),
		"angel_bond_heal_target":        currentChoiceResumePointResolver(),
		"sage_arcane_x":                 currentChoiceResumePointResolver(),
		"sage_holy_x":                   currentChoiceResumePointResolver(),
		"hb_holy_shard_combo":           currentChoiceResumePointResolver(),
		"hb_radiant_descent_cost":       currentChoiceResumePointResolver(),
		"hb_light_burst_mode":           currentChoiceResumePointResolver(),
		"hb_radiant_cannon_side":        currentChoiceResumePointResolver(),
		"ml_fullness_cost_card":         currentChoiceResumePointResolver(),
		"sc_spiritual_collapse_confirm": func(currentPoint string, ctxData map[string]interface{}) (string, bool) {
			if mode, _ := ctxData["mode"].(string); strings.HasPrefix(mode, "sc_hundred_night") {
				return model.NormalizeResumePoint(model.CombatStageCalcDamage), true
			}
			point := model.NormalizeResumePoint(currentPoint)
			return point, point != ""
		},
	}

	registerFixedChoiceResumePoints(m, model.TurnStageBeforeAction,
		"weak",
		"five_elements_bind",
	)
	registerFixedChoiceResumePoints(m, model.CombatStageCalcDamage,
		"hb_holy_shard_miss_confirm",
		"hb_holy_shard_miss_x",
		"hb_holy_shard_miss_ally_target",
		"sage_magic_rebound_confirm",
		"se_sword_qi_slash_x",
		"se_sword_qi_slash_target",
		"ml_black_spear_x",
		"ml_dark_barrier_mode",
		"ml_dark_barrier_x",
		"ml_dark_barrier_cards",
		"ml_stardust_target",
		"sc_hundred_night_power",
		"sc_hundred_night_fire_reveal",
		"sc_hundred_night_target",
		"sc_hundred_night_exclude_pick",
		"ss_link_transfer_x",
		"bd_descent_element",
		"bd_descent_cards",
		"bd_descent_target",
		"bp_blood_wail_x",
		"bt_pilgrimage_pick",
		"bt_poison_pick",
		"bt_mirror_pair",
		"bt_wither_confirm",
		"bt_wither_target",
		"mg_darkmoon_slash_x",
		"mg_blasphemy_target",
		"bs_beast_return_x",
		"bs_reversal_x",
	)
	registerFixedChoiceResumePoints(m, model.SubflowResponse,
		"god_protection_x",
		"hb_meteor_bullet_cost",
		"hb_meteor_bullet_target",
		"ss_convert_color",
		"mg_medusa_darkmoon_pick",
		"mg_medusa_magic_discard",
		"mg_medusa_magic_target",
	)
	registerFixedChoiceResumePoints(m, model.TurnStageActionExecution,
		"bp_shared_life_target",
		"extract",
		"mg_pale_moon_mode",
		"mg_pale_moon_x",
		"mg_pale_moon_target",
		"mg_pale_moon_discard",
	)
	registerFixedChoiceResumePoints(m, model.TurnStageExtraAction,
		"fighter_psi_bullet_target",
		"ss_recall_pick",
		"bp_curse_discard",
		"bt_dance_mode",
		"bt_dance_discard",
		"bt_cocoon_overflow_discard",
		"bt_reverse_mode",
		"bt_reverse_target",
		"bt_reverse_branch2_cost",
		"bt_reverse_branch2_pick",
	)
	registerFixedChoiceResumePoints(m, model.TurnStageActionStart,
		"ss_link_target",
		"bp_blood_sorrow_mode",
		"bp_blood_sorrow_target",
		"bd_rousing_mode",
		"bd_rousing_targets",
		"bd_rousing_discard_cards",
		"bd_hope_draw_confirm",
		"bd_hope_mode",
		"bd_hope_place_target",
		"bd_hope_transfer_target",
		"bd_hope_transfer_discard",
		"bs_iaijutsu_style_mode",
	)
	registerFixedChoiceResumePoints(m, model.TurnStageTurnEnd,
		"buy_resource",
		"bd_victory_mode",
		"bd_victory_extract_stone",
		"mg_moon_cycle_mode",
		"mg_moon_cycle_heal_target",
		"hb_auto_fill_resource",
		"hb_auto_fill_gain",
	)

	return m
}
