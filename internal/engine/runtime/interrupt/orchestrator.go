// gameflow: 中断统一调度（动作 + Prompt）。

package interrupt

import (
	"fmt"

	"starcup-engine/internal/model"
)

// Orchestrator 绑定宿主与规则表。
type Orchestrator struct {
	engine           EngineInterface
	actionRules      *ActionRules
	promptRules      *PromptRules
	deferredAfterPop []func(EngineInterface)
}

// NewOrchestrator 创建编排器。
func NewOrchestrator(engine EngineInterface, actionRules *ActionRules, promptRules *PromptRules) *Orchestrator {
	return &Orchestrator{
		engine:      engine,
		actionRules: actionRules,
		promptRules: promptRules,
	}
}

// DispatchAction 校验并分派中断输入。
func (o *Orchestrator) DispatchAction(act model.PlayerAction) error {
	if o == nil || o.engine == nil {
		return fmt.Errorf("interrupt orchestrator unavailable")
	}
	st := o.engine.GetState()
	if st == nil || st.PendingInterrupt == nil {
		return fmt.Errorf("没有待处理的中断")
	}
	intr := st.PendingInterrupt
	if act.PlayerID != intr.PlayerID {
		return fmt.Errorf("当前不是等待你的响应")
	}
	rule := o.actionRules.Get(intr.Type)
	if rule == nil {
		return fmt.Errorf("未知的中断类型: %s", intr.Type)
	}
	if len(rule.Allowed) > 0 && !rule.Allowed[act.Type] {
		if rule.InvalidActionMessage != "" {
			return fmt.Errorf(rule.InvalidActionMessage)
		}
		return fmt.Errorf("当前中断类型不支持该指令")
	}
	if rule.HandleResult == nil {
		return fmt.Errorf("中断规则未配置处理器: %s", intr.Type)
	}
	before := st.PendingInterrupt
	result, err := rule.HandleResult(o.engine, act)
	if err != nil {
		return err
	}
	if result.Consumed && before != nil && st.PendingInterrupt == before {
		if result.AfterPop != nil {
			o.deferredAfterPop = append(o.deferredAfterPop, result.AfterPop)
		}
		o.PopInterrupt()
	}
	return nil
}

// BuildInterruptPrompt 构建当前挂起中断的 Prompt。
func (o *Orchestrator) BuildInterruptPrompt() *model.Prompt {
	if o == nil || o.engine == nil {
		return nil
	}
	st := o.engine.GetState()
	if st == nil || st.PendingInterrupt == nil {
		return nil
	}
	b := o.promptRules.Get(st.PendingInterrupt.Type)
	if b == nil {
		return nil
	}
	return b(o.engine)
}

// PushInterrupt 挂起中断：无当前 Pending 则立即生效并同步阶段，否则入队。
func (o *Orchestrator) PushInterrupt(interrupt *model.Interrupt) {
	if o == nil || o.engine == nil || interrupt == nil {
		return
	}
	st := o.engine.GetState()
	if st == nil {
		return
	}
	if st.PendingInterrupt == nil {
		st.SetPendingInterrupt(interrupt)
		o.engine.ApplyInterruptPhase(interrupt)
		choiceType := ""
		if data, ok := interrupt.Context.(map[string]interface{}); ok {
			if ct, ok := data["choice_type"].(string); ok {
				choiceType = ct
			}
		}
		if choiceType != "" {
			o.engine.Log(fmt.Sprintf("[Interrupt] Pending=%s Player=%s Choice=%s", interrupt.Type, interrupt.PlayerID, choiceType))
		} else {
			o.engine.Log(fmt.Sprintf("[Interrupt] Pending=%s Player=%s", interrupt.Type, interrupt.PlayerID))
		}
		o.engine.NotifyInterruptPrompt()
		return
	}
	st.EnqueueInterrupt(interrupt)
	o.engine.Log(fmt.Sprintf("新中断入队等待: %s (Player: %s)", interrupt.Type, interrupt.PlayerID))
}

// RemoveQueuedInterruptByPredicate 从中断队列中移除所有满足 predicate 的中断。
func (o *Orchestrator) RemoveQueuedInterruptByPredicate(predicate func(*model.Interrupt) bool) {
	if o == nil || o.engine == nil || predicate == nil {
		return
	}
	st := o.engine.GetState()
	if st == nil || len(st.InterruptQueue) == 0 {
		return
	}
	filtered := make([]*model.Interrupt, 0, len(st.InterruptQueue))
	for _, intr := range st.InterruptQueue {
		if !predicate(intr) {
			filtered = append(filtered, intr)
		}
	}
	if len(filtered) != len(st.InterruptQueue) {
		st.InterruptQueue = filtered
		st.TouchInterruptRevision()
	}
}

// PopInterrupt 弹出当前中断；若队列非空则激活下一个并同步阶段。
func (o *Orchestrator) PopInterrupt() {
	if o == nil || o.engine == nil {
		return
	}
	st := o.engine.GetState()
	if st == nil {
		return
	}
	popped := st.PendingInterrupt
	st.SetPendingInterrupt(nil)
	if len(st.InterruptQueue) > 0 {
		nextInterrupt := st.InterruptQueue[0]
		st.InterruptQueue = st.InterruptQueue[1:]
		st.TouchInterruptRevision()
		st.SetPendingInterrupt(nextInterrupt)
		o.engine.Log(fmt.Sprintf("[System] 队列弹出中断: %s", nextInterrupt.Type))
		o.engine.ApplyInterruptPhase(nextInterrupt)
		o.engine.NotifyInterruptPrompt()
	} else {
		o.engine.Log("[System] 所有中断处理完毕，恢复主流程")
	}
	o.engine.ReconcileSubflowAfterInterruptPop(popped)
	o.drainDeferredAfterPop()
}

func (o *Orchestrator) drainDeferredAfterPop() {
	if o == nil || o.engine == nil {
		return
	}
	st := o.engine.GetState()
	for st != nil && st.PendingInterrupt == nil && len(o.deferredAfterPop) > 0 {
		after := o.deferredAfterPop[0]
		o.deferredAfterPop = o.deferredAfterPop[1:]
		if after != nil {
			after(o.engine)
		}
	}
}
