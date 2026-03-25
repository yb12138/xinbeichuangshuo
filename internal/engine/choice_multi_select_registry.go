package engine

type choiceSequentialCountResolver func(ctxData map[string]interface{}) (int, bool)

var registeredChoiceSequentialCountResolvers = buildChoiceSequentialCountResolvers()

func buildChoiceSequentialCountResolvers() map[string]choiceSequentialCountResolver {
	m := map[string]choiceSequentialCountResolver{
		"hom_rune_smash_cards":          remainingCountFromSelectionKey("x_value"),
		"hom_glyph_fusion_cards":        remainingCountFromSelectionKey("x_value"),
		"plague_death_touch_cards":      remainingCountFromSelectionKey("y_value"),
		"bw_mana_inversion_cards":       remainingCountFromSelectionKey("x_value"),
		"sage_magic_rebound_cards":      remainingCountFromSelectionKey("x_value"),
		"sage_arcane_cards":             remainingCountFromSelectionKey("x_value"),
		"sage_holy_cards":               remainingCountFromSelectionKey("x_value"),
		"mb_charge_place_cards":         remainingCountFromSelectionKey("need_count"),
		"hb_light_burst_mode_b_discard": remainingCountFromSelectionKey("x_value"),
		"ml_dark_barrier_cards":         remainingCountFromSelectionKey("x_value"),
		"bd_descent_cards":              remainingCountFromFixedTotal(2),
		"bd_rousing_discard_cards":      remainingCountFromFixedTotal(2),
		"bd_dissonance_discard_step":    remainingCountFromNeedAndSelected("need_count", "selected_count"),
		"mb_demon_eye_charge_card": func(ctxData map[string]interface{}) (int, bool) {
			need := toIntContextValue(ctxData["need_count"])
			if need <= 0 {
				need = 1
			}
			return need - len(parseChoiceIntSlice(ctxData["selected_indices"])), true
		},
	}
	return m
}

func remainingCountFromSelectionKey(key string) choiceSequentialCountResolver {
	return func(ctxData map[string]interface{}) (int, bool) {
		selectedCount := len(parseChoiceIntSlice(ctxData["selected_indices"]))
		return toIntContextValue(ctxData[key]) - selectedCount, true
	}
}

func remainingCountFromFixedTotal(total int) choiceSequentialCountResolver {
	return func(ctxData map[string]interface{}) (int, bool) {
		selectedCount := len(parseChoiceIntSlice(ctxData["selected_indices"]))
		return total - selectedCount, true
	}
}

func remainingCountFromNeedAndSelected(needKey, selectedKey string) choiceSequentialCountResolver {
	return func(ctxData map[string]interface{}) (int, bool) {
		return toIntContextValue(ctxData[needKey]) - toIntContextValue(ctxData[selectedKey]), true
	}
}

func registeredSequentialCardChoiceRemainingCount(choiceType string, ctxData map[string]interface{}) (int, bool) {
	resolver, ok := registeredChoiceSequentialCountResolvers[choiceType]
	if !ok {
		return 0, false
	}
	return resolver(ctxData)
}
