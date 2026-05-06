// gameflow: 顺序多选提交（一次 CmdSelect 多张）时循环单步 HandleSelect。

package engine

import (
	"fmt"

	"starcup-engine/internal/engine/core/runtimeutil"
	choicert "starcup-engine/internal/engine/runtime/choice"
	"starcup-engine/internal/model"
)

func (e *GameEngine) runSequentialChoiceSelections(playerID, choiceType string, selections []int) error {
	if len(selections) == 0 {
		return fmt.Errorf("请先选择手牌后再提交")
	}
	if e.State.PendingInterrupt == nil || e.State.PendingInterrupt.Type != model.InterruptChoice {
		return fmt.Errorf("当前不存在可处理的选牌中断")
	}
	ctxData, ok := choiceCtxAsAnyMap(e.State.PendingInterrupt.Context)
	if !ok {
		return fmt.Errorf("选牌上下文错误")
	}
	// Read the actual choice_type from context if not provided correctly
	actualChoiceType := choiceType
	if ctxChoiceType, ok := ctxData["choice_type"].(string); ok && ctxChoiceType != "" {
		actualChoiceType = ctxChoiceType
	}
	need, supported := choicert.SequentialRemainingCount(actualChoiceType, choiceCtxAsInterfaceMap(ctxData))
	if !supported {
		return fmt.Errorf("当前选择类型不支持多选提交流程")
	}
	if e.choiceEngine == nil {
		return fmt.Errorf("选择引擎未初始化")
	}
	// need=0 表示弹性数量（如欺诈允许2-3张），由用户决定提交几张。
	if need == 0 {
		need = len(selections)
		if need == 0 {
			return nil
		}
	}
	if need < 1 {
		need = 1
	}
	if len(selections) != need {
		return fmt.Errorf("需要选择 %d 张牌", need)
	}
	// 在循环开始前，将 selections（option indices）转换为候选索引（牌堆索引）
	remainingIndices := runtimeutil.ParseChoiceIntSlice(ctxData["remaining_indices"])
	totalCount := len(selections)

	processOne := func(loopIdx int, idx int) error {
		if e.State.PendingInterrupt != nil {
			newCtxData, ok := choiceCtxAsAnyMap(e.State.PendingInterrupt.Context)
			if ok {
				ctxData = newCtxData
			}
		}
		// sequential_remaining = 本轮之后还有多少张牌待处理（0 = 最后一张）
		ctxData["sequential_remaining"] = totalCount - loopIdx - 1
		if _, err := e.choiceEngine.HandleSelectResult(playerID, idx, ctxData); err != nil {
			return err
		}
		if e.State.PendingInterrupt != nil {
			newCtxData, ok := choiceCtxAsAnyMap(e.State.PendingInterrupt.Context)
			if ok {
				ctxData = newCtxData
			}
		}
		return nil
	}

	if len(remainingIndices) == 0 {
		// 所有手牌都可选，直接使用 selections 作为牌堆索引
		for i, idx := range selections {
			if err := processOne(i, idx); err != nil {
				return err
			}
			if i < len(selections)-1 {
				newChoiceType, _ := ctxData["choice_type"].(string)
				if newChoiceType != actualChoiceType {
					break
				}
			}
		}
		return nil
	}
	// 将 option indices 转换为 candidate indices
	candidateIndices := make([]int, 0, len(selections))
	for _, optIdx := range selections {
		cardIdx, ok := runtimeutil.ResolveSelectionToCandidate(optIdx, remainingIndices)
		if !ok || cardIdx < 0 {
			return fmt.Errorf("无效的选项索引: %d", optIdx)
		}
		candidateIndices = append(candidateIndices, cardIdx)
	}
	for i, cardIdx := range candidateIndices {
		if err := processOne(i, cardIdx); err != nil {
			return err
		}
		if i < len(candidateIndices)-1 {
			newChoiceType, _ := ctxData["choice_type"].(string)
			if newChoiceType != actualChoiceType {
				break
			}
		}
	}
	return nil
}
