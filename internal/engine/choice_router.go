// gameflow: 技能/系统选项输入按 choice_type 分派。

package engine

import (
	"fmt"

	"starcup-engine/internal/model"
)

// ==================== 类型定义 ====================

type choicePromptBuilder func(e *GameEngine, choiceType, playerID string, player *model.Player, data map[string]any) *model.Prompt

type choiceSingleInputHandler func(e *GameEngine, playerID string, selectionIndex int, ctxData map[string]any) (bool, error)

type choiceMultiInputHandler func(e *GameEngine, playerID string, selections []int) error

type choiceCancelHandler func(e *GameEngine, playerID string) error

// ==================== 注册表 ====================

var (
	choicePromptBuilders = make(map[string]choicePromptBuilder)
	choiceSingleHandlers = make(map[string]choiceSingleInputHandler)
	choiceMultiHandlers  = make(map[string]choiceMultiInputHandler)
	choiceCancelHandlers = make(map[string]choiceCancelHandler)
)

// RegisterChoicePrompt 注册选择提示构建器
func RegisterChoicePrompt(choiceType string, builder choicePromptBuilder) {
	if _, exists := choicePromptBuilders[choiceType]; !exists {
		choicePromptBuilders[choiceType] = builder
	}
}

// RegisterChoiceSingleHandler 注册单选输入处理器
func RegisterChoiceSingleHandler(choiceType string, handler choiceSingleInputHandler) {
	if _, exists := choiceSingleHandlers[choiceType]; !exists {
		choiceSingleHandlers[choiceType] = handler
	}
}

// RegisterChoiceMultiHandler 注册多选输入处理器
func RegisterChoiceMultiHandler(choiceType string, handler choiceMultiInputHandler) {
	if _, exists := choiceMultiHandlers[choiceType]; !exists {
		choiceMultiHandlers[choiceType] = handler
	}
}

// RegisterChoiceCancelHandler 注册取消处理器
func RegisterChoiceCancelHandler(choiceType string, handler choiceCancelHandler) {
	if _, exists := choiceCancelHandlers[choiceType]; !exists {
		choiceCancelHandlers[choiceType] = handler
	}
}

// ==================== 路由函数 ====================

func (e *GameEngine) buildRegisteredChoicePrompt(choiceType, playerID string, player *model.Player, data map[string]any) *model.Prompt {
	if builder, ok := choicePromptBuilders[choiceType]; ok {
		return builder(e, choiceType, playerID, player, data)
	}
	return nil
}

func (e *GameEngine) handleRegisteredChoiceInput(playerID string, selectionIndex int, ctxData map[string]any) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	if handler, ok := choiceSingleHandlers[choiceType]; ok {
		return handler(e, playerID, selectionIndex, ctxData)
	}
	return false, nil
}

func (e *GameEngine) handleRegisteredChoiceMultiSelect(playerID, choiceType string, selections []int) (bool, error) {
	if handler, ok := choiceMultiHandlers[choiceType]; ok {
		return true, handler(e, playerID, selections)
	}
	return false, nil
}

func (e *GameEngine) handleRegisteredChoiceCancel(playerID, choiceType string) (bool, error) {
	if handler, ok := choiceCancelHandlers[choiceType]; ok {
		return true, handler(e, playerID)
	}
	return false, nil
}

// ==================== 核心处理函数 ====================

func (e *GameEngine) handleLegacySequentialCardSelections(playerID string, selections []int) error {
	if len(selections) == 0 {
		return fmt.Errorf("请先选择手牌后再提交")
	}
	if e.State.PendingInterrupt == nil || e.State.PendingInterrupt.Type != model.InterruptChoice {
		return fmt.Errorf("当前不存在可处理的选牌中断")
	}
	ctxData, ok := e.State.PendingInterrupt.Context.(map[string]any)
	if !ok {
		return fmt.Errorf("选牌上下文错误")
	}
	choiceType, _ := ctxData["choice_type"].(string)
	needCount, supported := registeredSequentialCardChoiceRemainingCount(choiceType, ctxData)
	if !supported {
		return fmt.Errorf("当前选择类型不支持多选提交流程")
	}
	if needCount < 1 {
		needCount = 1
	}
	if len(selections) == 1 && needCount != 1 {
		return e.handleWeakChoiceInput(playerID, selections[0])
	}
	if len(selections) != needCount {
		return fmt.Errorf("需要选择 %d 张牌", needCount)
	}
	for _, idx := range selections {
		if err := e.handleWeakChoiceInput(playerID, idx); err != nil {
			return err
		}
	}
	return nil
}

func (e *GameEngine) handleWeakChoiceInput(playerID string, selectionIndex int) error {
	if e.State.PendingInterrupt == nil {
		return fmt.Errorf("没有待处理的中断")
	}

	ctxData, ok := e.State.PendingInterrupt.Context.(map[string]any)
	if !ok {
		return fmt.Errorf("中断上下文格式错误")
	}

	choiceType, _ := ctxData["choice_type"].(string)

	if handled, err := e.handleRegisteredChoiceInput(playerID, selectionIndex, ctxData); handled || err != nil {
		return err
	}

	return fmt.Errorf("未知的选择类型: %s", choiceType)
}

func (e *GameEngine) handleChoiceSelectionInput(playerID string, selectionIndex int) error {
	return e.handleWeakChoiceInput(playerID, selectionIndex)
}

func (e *GameEngine) handleInterruptChoiceAction(act model.PlayerAction) error {
	if act.Type == model.CmdCancel {
		if data, ok := e.State.PendingInterrupt.Context.(map[string]any); ok {
			if ct, _ := data["choice_type"].(string); ct != "" {
				if handled, err := e.handleRegisteredChoiceCancel(act.PlayerID, ct); handled || err != nil {
					return err
				}
			}
		}
	}
	if act.Type == model.CmdSelect {
		if data, ok := e.State.PendingInterrupt.Context.(map[string]any); ok {
			if ct, _ := data["choice_type"].(string); ct != "" {
				if handled, err := e.handleRegisteredChoiceMultiSelect(act.PlayerID, ct, act.Selections); handled || err != nil {
					return err
				}
				if _, isLegacyCardMulti := registeredSequentialCardChoiceRemainingCount(ct, data); isLegacyCardMulti {
					return e.handleLegacySequentialCardSelections(act.PlayerID, act.Selections)
				}
			}
		}
		if len(act.Selections) != 1 {
			return fmt.Errorf("请选择一个选项")
		}
		return e.handleWeakChoiceInput(act.PlayerID, act.Selections[0])
	}
	return fmt.Errorf("当前中断类型不支持该指令")
}

func (e *GameEngine) cancelExtractChoice(playerID string) error {
	e.PopInterrupt()
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

func (e *GameEngine) cancelHomDualEchoChoice(playerID string) error {
	e.PopInterrupt()
	if p := e.State.Players[playerID]; p != nil {
		e.Log("[System] " + p.Name + " 取消了 [双重回响] 的目标选择")
	} else {
		e.Log("[System] " + playerID + " 取消了 [双重回响] 的目标选择")
	}
	if e.State.PendingInterrupt == nil && len(e.State.PendingDamageQueue) > 0 {
		e.enterDamageResolution(nil)
	}
	return nil
}
