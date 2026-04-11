// gameflow: HandleAction 入口：将 PlayerAction 分派到技能/攻击/法术/选择等。

package engine

import (
	"fmt"

	"starcup-engine/internal/model"
)

// HandleAction 核心路由器：处理所有 Action
func (e *GameEngine) HandleAction(act model.PlayerAction) error {
	e.Log(fmt.Sprintf("[Debug] HandleAction 收到指令: %s", act.Type))
	// 规则顺序（统一入口）：
	// 1) 系统指令；2) 中断输入；3) 终局拦截；4) 回合权校验；5) 主流程路由；6) Drive 推进。
	if handled, err := e.handleImmediateAction(act); handled {
		return err
	}
	if handled, err := e.handlePendingInterruptInput(act); handled {
		return err
	}
	if e.State.GameOver {
		return fmt.Errorf("游戏已结束")
	}
	if err := e.validateActionTurnOwnership(act); err != nil {
		return err
	}
	if err := e.dispatchActionByWindow(act); err != nil {
		return err
	}
	e.driveAfterSuccessfulAction()
	return nil
}

func (e *GameEngine) handleImmediateAction(act model.PlayerAction) (bool, error) {
	switch act.Type {
	case model.CmdQuit:
		// 系统级退出：允许在任意状态触发。
		return true, fmt.Errorf("EXIT_GAME")
	case model.CmdHelp:
		// 帮助信息由上层展示，引擎本身无需推进状态机。
		return true, nil
	case model.CmdCheat:
		// Debug 指令：成功后立刻尝试推进，保持状态提示同步。
		if err := e.handleCheat(act); err != nil {
			return true, err
		}
		e.driveAfterSuccessfulAction()
		return true, nil
	default:
		return false, nil
	}
}

func (e *GameEngine) handlePendingInterruptInput(act model.PlayerAction) (bool, error) {
	if e.State.PendingInterrupt == nil {
		return false, nil
	}
	if err := e.handleInterruptAction(act); err != nil {
		return true, err
	}
	// 中断响应成功后立即推进：若中断已弹出则继续主流程，若仍有中断则 Drive 会自然停在等待点。
	e.Drive()
	return true, nil
}

func (e *GameEngine) validateActionTurnOwnership(act model.PlayerAction) error {
	// Start 属于流程入口，不参与“当前行动者”校验。
	if act.Type == model.CmdStart {
		return nil
	}
	// 战斗交互窗口由响应函数校验“是否轮到该目标响应”。
	if e.isCombatInteractionWindow() {
		return nil
	}
	if len(e.State.PlayerOrder) == 0 {
		return fmt.Errorf("当前没有可行动玩家")
	}
	currentPlayer := e.State.PlayerOrder[e.State.CurrentTurn]
	if act.PlayerID != currentPlayer {
		return fmt.Errorf("不是你的回合")
	}
	return nil
}

func (e *GameEngine) dispatchActionByWindow(act model.PlayerAction) error {
	switch {
	case e.isActionSelectionWindow():
		// 行动选择窗口：攻击、法术、技能等主动指令。
		return e.handleActionSelection(act)
	case e.isCombatInteractionWindow():
		// 战斗交互窗口：仅接收响应指令。
		if act.Type != model.CmdRespond {
			return fmt.Errorf("当前必须响应战斗 (使用 take/defend/counter)")
		}
		return e.handleCombatResponse(act)
	default:
		// 其余状态仅保留 Start 作为流程入口。
		if act.Type == model.CmdStart {
			return e.StartGame()
		}
		return fmt.Errorf("当前状态 (%s) 不支持该指令", e.runtimeStateLabel())
	}
}

func (e *GameEngine) driveAfterSuccessfulAction() {
	e.Log(fmt.Sprintf("[Debug] 指令执行成功，准备 Drive. %s, Interrupt: %v", e.runtimeStateLabel(), e.State.PendingInterrupt))
	if e.State.PendingInterrupt != nil {
		e.Log("[Debug] 存在挂起中断，暂不 Drive")
		return
	}
	e.Drive()
}

// handleInterruptAction 专门处理中断状态下的输入
func (e *GameEngine) handleInterruptAction(act model.PlayerAction) error {
	interrupt := e.State.PendingInterrupt
	if interrupt == nil {
		return fmt.Errorf("没有待处理的中断")
	}
	if act.PlayerID != interrupt.PlayerID {
		return fmt.Errorf("当前不是等待你的响应")
	}
	return e.dispatchInterruptAction(interrupt, act)
}

func (e *GameEngine) dispatchInterruptAction(interrupt *model.Interrupt, act model.PlayerAction) error {
	rule, ok := interruptActionRules[interrupt.Type]
	if !ok {
		return fmt.Errorf("未知的中断类型: %s", interrupt.Type)
	}
	if len(rule.allowed) > 0 && !rule.allowed[act.Type] {
		if rule.invalidActionMessage != "" {
			return fmt.Errorf(rule.invalidActionMessage)
		}
		return fmt.Errorf("当前中断类型不支持该指令")
	}
	if rule.handler == nil {
		return fmt.Errorf("中断规则未配置处理器: %s", interrupt.Type)
	}
	return rule.handler(e, act)
}
