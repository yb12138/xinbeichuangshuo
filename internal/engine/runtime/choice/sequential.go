// gameflow: 顺序多选牌（一次 CmdSelect 提交多张）的剩余张数解析。

package choice

import "starcup-engine/internal/engine/core/runtimeutil"

// SequentialRemainingCount 返回当前上下文还需选择的张数；不支持则 ok=false。
func SequentialRemainingCount(choiceType string, ctxData map[string]interface{}) (need int, ok bool) {
	resolver, found := sequentialCardResolvers[choiceType]
	if !found {
		return 0, false
	}
	return resolver(ctxData)
}

type sequentialCountResolver func(ctxData map[string]interface{}) (int, bool)

// HasSequentialMulti 表示该 choice_type 支持「一次提交多张手牌」的顺序多选语义。
func HasSequentialMulti(choiceType string) bool {
	_, ok := sequentialCardResolvers[choiceType]
	return ok
}

var sequentialCardResolvers = map[string]sequentialCountResolver{
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
	"bp_curse_discard":              remainingCountFromSelectionKey("discard_count"),
	"mb_demon_eye_charge_card": func(ctxData map[string]interface{}) (int, bool) {
		need := runtimeutil.ToIntContextValue(ctxData["need_count"])
		if need <= 0 {
			need = 1
		}
		return need - len(runtimeutil.ParseChoiceIntSlice(ctxData["selected_indices"])), true
	},
	"adventurer_fraud_pick": func(ctxData map[string]interface{}) (int, bool) {
		selectedCount := len(runtimeutil.ParseChoiceIntSlice(ctxData["selected_indices"]))
		// 欺诈允许选择2-3张牌；首次选择返回0表示弹性数量，由用户决定提交几张。
		if selectedCount == 0 {
			return 0, true
		}
		if selectedCount < 2 {
			return 2 - selectedCount, true
		}
		if selectedCount == 2 {
			return 1, true // 可选第三张
		}
		return 0, true
	},
}

func remainingCountFromSelectionKey(key string) sequentialCountResolver {
	return func(ctxData map[string]interface{}) (int, bool) {
		selectedCount := len(runtimeutil.ParseChoiceIntSlice(ctxData["selected_indices"]))
		return runtimeutil.ToIntContextValue(ctxData[key]) - selectedCount, true
	}
}

func remainingCountFromFixedTotal(total int) sequentialCountResolver {
	return func(ctxData map[string]interface{}) (int, bool) {
		selectedCount := len(runtimeutil.ParseChoiceIntSlice(ctxData["selected_indices"]))
		return total - selectedCount, true
	}
}

func remainingCountFromNeedAndSelected(needKey, selectedKey string) sequentialCountResolver {
	return func(ctxData map[string]interface{}) (int, bool) {
		return runtimeutil.ToIntContextValue(ctxData[needKey]) - runtimeutil.ToIntContextValue(ctxData[selectedKey]), true
	}
}
