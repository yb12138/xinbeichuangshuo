package model

import "fmt"

const PromptFlowContextKey = "flow"

const (
	CancelPolicyDeny    = "deny"
	CancelPolicyAbort   = "abort"
	CancelPolicyDecline = "decline"
	CancelPolicyBack    = "back"
)

type PromptFlowCancelAction string

const (
	PromptFlowCancelDenied   PromptFlowCancelAction = "denied"
	PromptFlowCancelAborted  PromptFlowCancelAction = "aborted"
	PromptFlowCancelDeclined PromptFlowCancelAction = "declined"
	PromptFlowCancelBack     PromptFlowCancelAction = "back"
)

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
	StepSpecs             map[string]PromptFlowStepSpec  `json:"-"`
}

type PromptFlowStepSpec struct {
	ID           string
	ChoiceType   string
	CancelPolicy string
}

type PromptFlowCancelResult struct {
	Action PromptFlowCancelAction
	StepID string
}

type PromptFlowRuntime struct {
	FlowID string
	steps  map[string]PromptFlowStepSpec
	order  []string
}

func NewPromptFlowState(flowID, stepID string) *PromptFlowState {
	return &PromptFlowState{
		FlowID:                flowID,
		StepID:                stepID,
		AccumulatedSelections: map[string]PromptFlowSelection{},
	}
}

func NewPromptFlowRuntime(flowID string, steps []PromptFlowStepSpec) (*PromptFlowRuntime, error) {
	if flowID == "" {
		return nil, fmt.Errorf("prompt flow runtime requires flow_id")
	}
	if len(steps) == 0 {
		return nil, fmt.Errorf("prompt flow %q requires at least one step", flowID)
	}
	rt := &PromptFlowRuntime{
		FlowID: flowID,
		steps:  map[string]PromptFlowStepSpec{},
		order:  make([]string, 0, len(steps)),
	}
	for _, step := range steps {
		if step.ID == "" {
			return nil, fmt.Errorf("prompt flow %q has empty step_id", flowID)
		}
		if _, exists := rt.steps[step.ID]; exists {
			return nil, fmt.Errorf("prompt flow %q has duplicate step_id %q", flowID, step.ID)
		}
		rt.steps[step.ID] = step
		rt.order = append(rt.order, step.ID)
	}
	return rt, nil
}

func MustNewPromptFlowRuntime(flowID string, steps []PromptFlowStepSpec) *PromptFlowRuntime {
	rt, err := NewPromptFlowRuntime(flowID, steps)
	if err != nil {
		panic(err)
	}
	return rt
}

func (rt *PromptFlowRuntime) Begin() *PromptFlowState {
	if rt == nil || len(rt.order) == 0 {
		return nil
	}
	return rt.newState(rt.order[0])
}

func (rt *PromptFlowRuntime) BeginAt(stepID string) (*PromptFlowState, error) {
	if rt == nil {
		return nil, fmt.Errorf("prompt flow runtime is nil")
	}
	if stepID == "" {
		return nil, fmt.Errorf("prompt flow %q requires step_id", rt.FlowID)
	}
	if _, ok := rt.Step(stepID); !ok {
		return nil, fmt.Errorf("prompt flow %q unknown step %q", rt.FlowID, stepID)
	}
	return rt.newState(stepID), nil
}

func (rt *PromptFlowRuntime) MustBeginAt(stepID string) *PromptFlowState {
	flow, err := rt.BeginAt(stepID)
	if err != nil {
		panic(err)
	}
	return flow
}

func (rt *PromptFlowRuntime) Step(stepID string) (PromptFlowStepSpec, bool) {
	if rt == nil || stepID == "" {
		return PromptFlowStepSpec{}, false
	}
	step, ok := rt.steps[stepID]
	return step, ok
}

func (rt *PromptFlowRuntime) newState(stepID string) *PromptFlowState {
	flow := NewPromptFlowState(rt.FlowID, stepID)
	flow.StepSpecs = make(map[string]PromptFlowStepSpec, len(rt.steps))
	for id, spec := range rt.steps {
		flow.StepSpecs[id] = spec
	}
	return flow
}

func (rt *PromptFlowRuntime) CurrentStep(flow *PromptFlowState) (PromptFlowStepSpec, bool) {
	if rt == nil || flow == nil || flow.FlowID != rt.FlowID {
		return PromptFlowStepSpec{}, false
	}
	return rt.Step(flow.StepID)
}

func (f *PromptFlowState) CurrentStepSpec() (PromptFlowStepSpec, bool) {
	if f == nil || f.StepID == "" || f.StepSpecs == nil {
		return PromptFlowStepSpec{}, false
	}
	step, ok := f.StepSpecs[f.StepID]
	return step, ok
}

func (f *PromptFlowState) StepSpec(stepID string) (PromptFlowStepSpec, bool) {
	if f == nil || stepID == "" || f.StepSpecs == nil {
		return PromptFlowStepSpec{}, false
	}
	step, ok := f.StepSpecs[stepID]
	return step, ok
}

func (f *PromptFlowState) HandleCancel() (PromptFlowCancelResult, error) {
	step, ok := f.CurrentStepSpec()
	if !ok {
		return PromptFlowCancelResult{Action: PromptFlowCancelDenied}, fmt.Errorf("prompt flow current step is invalid")
	}
	switch step.CancelPolicy {
	case CancelPolicyBack:
		previous, moved := f.Back()
		if !moved {
			return PromptFlowCancelResult{Action: PromptFlowCancelDenied, StepID: f.StepID}, fmt.Errorf("当前步骤无法返回")
		}
		return PromptFlowCancelResult{Action: PromptFlowCancelBack, StepID: previous}, nil
	case CancelPolicyAbort:
		f.Clear()
		return PromptFlowCancelResult{Action: PromptFlowCancelAborted}, nil
	case CancelPolicyDecline:
		f.PutSelection(f.StepID, PromptFlowSelection{})
		return PromptFlowCancelResult{Action: PromptFlowCancelDeclined, StepID: f.StepID}, nil
	default:
		return PromptFlowCancelResult{Action: PromptFlowCancelDenied, StepID: f.StepID}, fmt.Errorf("当前步骤不可取消")
	}
}

func (rt *PromptFlowRuntime) MoveTo(flow *PromptFlowState, stepID string) error {
	if rt == nil {
		return fmt.Errorf("prompt flow runtime is nil")
	}
	if flow == nil || flow.FlowID != rt.FlowID {
		return fmt.Errorf("prompt flow %q state missing or mismatched", rt.FlowID)
	}
	if _, ok := rt.Step(stepID); !ok {
		return fmt.Errorf("prompt flow %q unknown step %q", rt.FlowID, stepID)
	}
	flow.Advance(stepID)
	return nil
}

func (rt *PromptFlowRuntime) RecordAndMove(flow *PromptFlowState, selection PromptFlowSelection, nextStepID string) error {
	if rt == nil {
		return fmt.Errorf("prompt flow runtime is nil")
	}
	if flow == nil || flow.FlowID != rt.FlowID {
		return fmt.Errorf("prompt flow %q state missing or mismatched", rt.FlowID)
	}
	current, ok := rt.CurrentStep(flow)
	if !ok {
		return fmt.Errorf("prompt flow %q unknown current step %q", rt.FlowID, flow.StepID)
	}
	flow.PutSelection(current.ID, selection)
	return rt.MoveTo(flow, nextStepID)
}

func (rt *PromptFlowRuntime) HandleCancel(flow *PromptFlowState) (PromptFlowCancelResult, error) {
	if flow != nil && flow.StepSpecs == nil && rt != nil {
		flow.StepSpecs = make(map[string]PromptFlowStepSpec, len(rt.steps))
		for id, spec := range rt.steps {
			flow.StepSpecs[id] = spec
		}
	}
	return flow.HandleCancel()
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

func (f *PromptFlowState) Back() (string, bool) {
	if f == nil || len(f.History) == 0 {
		return "", false
	}
	previous := f.History[len(f.History)-1]
	f.History = f.History[:len(f.History)-1]
	f.StepID = previous
	return previous, true
}

func (f *PromptFlowState) Clear() {
	if f == nil {
		return
	}
	f.StepID = ""
	f.History = nil
	f.AccumulatedSelections = map[string]PromptFlowSelection{}
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
