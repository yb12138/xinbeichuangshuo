// gameflow: ChoiceEngine 与 InterruptChoice 动作入口（无弱选择 / 无顺序补丁）。

package engine

import (
	"fmt"

	choicert "starcup-engine/internal/engine/runtime/choice"
	"starcup-engine/internal/model"
)

// choiceHostBridge 将 *GameEngine 注入 runtime/choice（无 Legacy 回退）。
type choiceHostBridge struct {
	e *GameEngine
}

var _ choicert.Host = (*choiceHostBridge)(nil)

func (*choiceHostBridge) ChoiceEngineHost() {}

// choiceCtxAsAnyMap 将中断 Context 转为 map[string]any。
func choiceCtxAsAnyMap(raw interface{}) (map[string]any, bool) {
	m, ok := raw.(map[string]any)
	if !ok {
		return nil, false
	}
	return m, true
}

func choiceCtxAsInterfaceMap(m map[string]any) map[string]interface{} {
	if m == nil {
		return nil
	}
	out := make(map[string]interface{}, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func (e *GameEngine) applyInterruptChoiceSelect(playerID string, selectionIndex int, ctxData map[string]any) error {
	if e.choiceEngine == nil {
		return fmt.Errorf("选择引擎未初始化")
	}
	_, err := e.choiceEngine.HandleSelect(playerID, selectionIndex, ctxData)
	return err
}

func (e *GameEngine) handleInterruptChoiceAction(act model.PlayerAction) error {
	if e.State.PendingInterrupt == nil {
		return fmt.Errorf("没有待处理的中断")
	}
	if e.choiceEngine == nil {
		return fmt.Errorf("选择引擎未初始化")
	}
	data, ok := choiceCtxAsAnyMap(e.State.PendingInterrupt.Context)
	if !ok {
		return fmt.Errorf("中断上下文格式错误")
	}
	if act.Type == model.CmdCancel {
		_, err := e.choiceEngine.HandleCancel(act.PlayerID, data)
		return err
	}
	if act.Type == model.CmdSelect {
		ct, _ := data["choice_type"].(string)
		if ct == "" {
			return fmt.Errorf("中断上下文缺少 choice_type")
		}
		if len(act.Selections) > 1 {
			_, err := e.choiceEngine.HandleMultiSelect(act.PlayerID, ct, act.Selections, data)
			return err
		}
		if len(act.Selections) != 1 {
			return fmt.Errorf("请选择一个选项")
		}
		return e.applyInterruptChoiceSelect(act.PlayerID, act.Selections[0], data)
	}
	return fmt.Errorf("当前中断类型不支持该指令")
}

func (e *GameEngine) cancelExtractChoice(playerID string) error {
	e.PopInterrupt()
	if p := e.State.Players[playerID]; p != nil {
		// 提炼取消属于“行动未提交”，需要回滚预写入的行动收尾标记。
		p.TurnState.LastActionType = ""
		p.TurnState.LastActionCard = nil
		p.TurnState.HasActed = false
	}
	if e.State.PendingInterrupt == nil {
		e.enterActionExecutionStage()
	}
	if p := e.State.Players[playerID]; p != nil {
		e.Log("[System] " + p.Name + " 取消了提炼操作")
	} else {
		e.Log("[System] " + playerID + " 取消了提炼操作")
	}
	return nil
}
