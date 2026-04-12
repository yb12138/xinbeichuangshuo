// gameflow: 顺序多选提交（一次 CmdSelect 多张）时循环单步 HandleSelect。

package engine

import (
	"fmt"

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
	need, supported := choicert.SequentialRemainingCount(choiceType, choiceCtxAsInterfaceMap(ctxData))
	if !supported {
		return fmt.Errorf("当前选择类型不支持多选提交流程")
	}
	if need < 1 {
		need = 1
	}
	if e.choiceEngine == nil {
		return fmt.Errorf("选择引擎未初始化")
	}
	if len(selections) == 1 && need != 1 {
		_, err := e.choiceEngine.HandleSelect(playerID, selections[0], ctxData)
		return err
	}
	if len(selections) != need {
		return fmt.Errorf("需要选择 %d 张牌", need)
	}
	for _, idx := range selections {
		if _, err := e.choiceEngine.HandleSelect(playerID, idx, ctxData); err != nil {
			return err
		}
	}
	return nil
}
