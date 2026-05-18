// gameflow: 选择顺序剩余数的通用构造器。

package player

import (
	"starcup-engine/internal/engine/core/runtimeutil"
	"starcup-engine/internal/model"
)

// ChoiceRemainingFromSelectionKey 根据 ctxData 中的数值字段计算剩余数量。
func ChoiceRemainingFromSelectionKey(key string) ChoiceSequentialRemaining {
	return func(ctxData map[string]interface{}) (int, bool) {
		selectedCount := len(runtimeutil.ParseChoiceIntSlice(ctxData["selected_indices"]))
		return runtimeutil.ToIntContextValue(ctxData[key]) - selectedCount, true
	}
}

// ChoiceRemainingFromSelectionKeyFloor 根据 ctxData 中的数值字段计算剩余数量，并给需求数量设置下限。
func ChoiceRemainingFromSelectionKeyFloor(key string, floor int) ChoiceSequentialRemaining {
	return func(ctxData map[string]interface{}) (int, bool) {
		need := runtimeutil.ToIntContextValue(ctxData[key])
		if need < floor {
			need = floor
		}
		selectedCount := len(runtimeutil.ParseChoiceIntSlice(ctxData["selected_indices"]))
		return need - selectedCount, true
	}
}

// ChoiceRemainingFromFixedTotal 根据固定总数计算剩余数量。
func ChoiceRemainingFromFixedTotal(total int) ChoiceSequentialRemaining {
	return func(ctxData map[string]interface{}) (int, bool) {
		selectedCount := len(runtimeutil.ParseChoiceIntSlice(ctxData["selected_indices"]))
		return total - selectedCount, true
	}
}

// ChoiceRemainingFromNeedAndSelected 根据 need/selected 字段计算剩余数量。
func ChoiceRemainingFromNeedAndSelected(needKey, selectedKey string) ChoiceSequentialRemaining {
	return func(ctxData map[string]interface{}) (int, bool) {
		return runtimeutil.ToIntContextValue(ctxData[needKey]) - runtimeutil.ToIntContextValue(ctxData[selectedKey]), true
	}
}

// ChoiceRemainingFromFlowSelectionCount derives sequential remaining count from
// PromptFlowState accumulated selections.
func ChoiceRemainingFromFlowSelectionCount(countStepID, selectedStepID string) ChoiceSequentialRemaining {
	return func(ctxData map[string]interface{}) (int, bool) {
		flow := model.PromptFlowFromContext(ctxData)
		if flow == nil {
			return 0, false
		}
		need := flow.Selection(countStepID).Count
		selectedCount := len(flow.Selection(selectedStepID).OptionIndexes)
		return need - selectedCount, true
	}
}

// ChoiceRemainingFromFlexibleRange 表示一次提交允许在 min/max 范围内弹性结束。
func ChoiceRemainingFromFlexibleRange(minCount, maxCount int) ChoiceSequentialRemaining {
	return func(ctxData map[string]interface{}) (int, bool) {
		selectedCount := len(runtimeutil.ParseChoiceIntSlice(ctxData["selected_indices"]))
		switch {
		case selectedCount == 0:
			return 0, true
		case selectedCount < minCount:
			return minCount - selectedCount, true
		case selectedCount < maxCount:
			return maxCount - selectedCount, true
		default:
			return 0, true
		}
	}
}
