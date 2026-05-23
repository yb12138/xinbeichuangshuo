// gameflow: 选择顺序剩余数的通用构造器。

package player

import "starcup-engine/internal/model"

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

// ChoiceRemainingFromFlowFlexibleRange 表示一次提交允许在 min/max 范围内弹性结束，
// 已选数量从 PromptFlowState 读取。
func ChoiceRemainingFromFlowFlexibleRange(selectedStepID string, minCount, maxCount int) ChoiceSequentialRemaining {
	return func(ctxData map[string]interface{}) (int, bool) {
		flow := model.PromptFlowFromContext(ctxData)
		if flow == nil {
			return 0, false
		}
		selectedCount := len(flow.Selection(selectedStepID).OptionIndexes)
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
