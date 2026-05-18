package model

import "fmt"

const PromptFlowContextKey = "flow"

type PromptFlowSelection struct {
	OptionIndexes []int    `json:"option_indexes,omitempty"`
	CardIDs       []string `json:"card_ids,omitempty"`
	TargetIDs     []string `json:"target_ids,omitempty"`
	Element       string   `json:"element,omitempty"`
	Count         int      `json:"count,omitempty"`
}

type PromptFlowState struct {
	FlowID                string                         `json:"flow_id"`
	StepID                string                         `json:"step_id"`
	History               []string                       `json:"history,omitempty"`
	AccumulatedSelections map[string]PromptFlowSelection `json:"accumulated_selections,omitempty"`
}

func NewPromptFlowState(flowID, stepID string) *PromptFlowState {
	return &PromptFlowState{
		FlowID:                flowID,
		StepID:                stepID,
		AccumulatedSelections: map[string]PromptFlowSelection{},
	}
}

func PromptFlowFromContext(ctx map[string]interface{}) *PromptFlowState {
	if ctx == nil {
		return nil
	}
	flow, _ := ctx[PromptFlowContextKey].(*PromptFlowState)
	return flow
}

func RequirePromptFlow(ctx map[string]interface{}, flowID, label string) (*PromptFlowState, error) {
	flow := PromptFlowFromContext(ctx)
	if flow == nil || flow.FlowID != flowID {
		if label == "" {
			label = "选择流程"
		}
		return nil, fmt.Errorf("%s缺少多步流状态", label)
	}
	return flow, nil
}

func SetPromptFlowContext(ctx map[string]interface{}, flow *PromptFlowState) {
	if ctx == nil || flow == nil {
		return
	}
	ctx[PromptFlowContextKey] = flow
}

func (f *PromptFlowState) Advance(stepID string) {
	if f == nil || stepID == "" || f.StepID == stepID {
		return
	}
	if f.StepID != "" {
		f.History = append(f.History, f.StepID)
	}
	f.StepID = stepID
}

func (f *PromptFlowState) PutSelection(stepID string, selection PromptFlowSelection) {
	if f == nil || stepID == "" {
		return
	}
	if f.AccumulatedSelections == nil {
		f.AccumulatedSelections = map[string]PromptFlowSelection{}
	}
	f.AccumulatedSelections[stepID] = selection
}

func (f *PromptFlowState) Selection(stepID string) PromptFlowSelection {
	if f == nil || f.AccumulatedSelections == nil || stepID == "" {
		return PromptFlowSelection{}
	}
	return f.AccumulatedSelections[stepID]
}
