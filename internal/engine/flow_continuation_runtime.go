// gameflow: FlowContinuation 处理机制（流程边界恢复点触发）。

package engine

import (
	"fmt"

	"starcup-engine/internal/model"
)

// processFlowContinuations 处理指定类型的流程边界恢复点。
func (e *GameEngine) processFlowContinuations(kind model.FlowContinuationKind) {
	if len(e.State.FlowContinuations) == 0 {
		return
	}

	pending := e.State.FlowContinuations
	e.State.FlowContinuations = nil // 清空队列

	for _, cont := range pending {
		if cont.Kind != kind {
			// 不是当前边界，保留到下次
			e.State.FlowContinuations = append(e.State.FlowContinuations, cont)
			continue
		}

		// 查找角色 handler
		entry := roleRegistry.Entry(cont.RoleID)
		if entry.ID == "" {
			e.Log(fmt.Sprintf("[Error] FlowContinuation 角色未注册: %s — 流程配置错误", cont.RoleID))
			e.State.FlowContinuations = append(e.State.FlowContinuations, cont)
			continue
		}

		handler := entry.FlowContinuationHandlers[kind]
		if handler == nil {
			e.Log(fmt.Sprintf("[Error] FlowContinuation handler 未注册: %s/%s — 流程配置错误", cont.RoleID, kind))
			e.State.FlowContinuations = append(e.State.FlowContinuations, cont)
			continue
		}

		// 执行
		rt := NewRoleChoiceRuntime(e)
		if err := handler(rt, cont); err != nil {
			e.Log(fmt.Sprintf("[Error] FlowContinuation 处理失败: %s/%s - %v", cont.RoleID, kind, err))
		}
	}
}
