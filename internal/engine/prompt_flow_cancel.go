package engine

import (
	"fmt"

	"starcup-engine/internal/model"
)

func (e *GameEngine) cancelPromptFlowChoice(playerID string, ctxData map[string]any) (bool, error) {
	ctx := choiceCtxAsInterfaceMap(ctxData)
	flow := model.PromptFlowFromContext(ctx)
	if flow == nil {
		return false, nil
	}
	result, err := flow.HandleCancel()
	if err != nil {
		return false, err
	}
	switch result.Action {
	case model.PromptFlowCancelBack:
		step, ok := flow.StepSpec(result.StepID)
		if !ok || step.ChoiceType == "" {
			return false, fmt.Errorf("prompt flow %q missing choice_type for step %q", flow.FlowID, result.StepID)
		}
		ctxData["choice_type"] = step.ChoiceType
		ctxData[model.PromptFlowContextKey] = flow
		if e.State != nil && e.State.PendingInterrupt != nil {
			e.State.PendingInterrupt.Context = ctxData
		}
		e.NotifyInterruptPrompt()
		return true, nil
	case model.PromptFlowCancelDeclined:
		ctxData[model.PromptFlowContextKey] = flow
		if e.State != nil && e.State.PendingInterrupt != nil {
			e.State.PendingInterrupt.Context = ctxData
		}
		e.PopInterrupt()
		return true, nil
	case model.PromptFlowCancelAborted:
		e.PopInterrupt()
		return true, nil
	default:
		return false, fmt.Errorf("当前选择不可取消")
	}
}
