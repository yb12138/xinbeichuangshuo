// gameflow: ChoiceEngine 仅走 ChoiceSpec，未注册即报错。

package choice

import (
	"fmt"

	"starcup-engine/internal/model"
)

// Engine 统一选择交互引擎。
type Engine struct {
	reg  *SpecRegistry
	host Host
}

// NewEngine 创建引擎；随后必须 SetHost。
func NewEngine(reg *SpecRegistry) *Engine {
	if reg == nil {
		reg = NewSpecRegistry()
	}
	return &Engine{reg: reg}
}

// SetHost 注入宿主（通常在 NewGameEngine 末尾调用）。
func (e *Engine) SetHost(h Host) {
	e.host = h
}

// Registry 返回内部注册表（供 bootstrap 注册 ChoiceSpec）。
func (e *Engine) Registry() *SpecRegistry {
	if e == nil {
		return nil
	}
	return e.reg
}

func (e *Engine) SequentialRemainingCount(choiceType string, ctxData map[string]any) (int, bool) {
	if e == nil || e.reg == nil {
		return 0, false
	}
	return e.reg.SequentialRemainingCount(choiceType, ctxData)
}

// BuildPrompt 构建 Prompt；必须存在带 BuildPrompt 的 ChoiceSpec。
func (e *Engine) BuildPrompt(choiceType, playerID string, player *model.Player, data map[string]any) (*model.Prompt, error) {
	if e == nil || e.host == nil {
		return nil, fmt.Errorf("choice: engine or host is nil")
	}
	if choiceType == "" {
		return nil, fmt.Errorf("choice: empty choice_type")
	}
	spec := e.reg.Get(choiceType)
	if spec == nil {
		return nil, fmt.Errorf("choice: unregistered choice_type %q", choiceType)
	}
	if spec.BuildPrompt == nil {
		return nil, fmt.Errorf("choice: no BuildPrompt for %q", choiceType)
	}
	p := spec.BuildPrompt(e.host, choiceType, playerID, player, data)
	if p == nil {
		return nil, fmt.Errorf("choice: BuildPrompt returned nil for %q", choiceType)
	}
	return p, nil
}

// HandleSelectResult 单选；必须存在 OnSelect，并显式返回消费结果。
func (e *Engine) HandleSelectResult(playerID string, selectionIndex int, ctxData map[string]any) (HandleResult, error) {
	if e == nil || e.host == nil {
		return HandleResult{}, fmt.Errorf("choice: engine or host is nil")
	}
	choiceType, _ := ctxData["choice_type"].(string)
	if choiceType == "" {
		return HandleResult{}, fmt.Errorf("choice: context missing choice_type")
	}
	spec := e.reg.Get(choiceType)
	if spec == nil || spec.OnSelect == nil {
		return HandleResult{}, fmt.Errorf("choice: unregistered or no OnSelect for %q", choiceType)
	}
	handled, err := spec.OnSelect(e.host, playerID, selectionIndex, ctxData)
	if err != nil {
		return HandleResult{}, err
	}
	if !handled {
		return HandleResult{}, fmt.Errorf("choice: OnSelect returned unhandled for %q", choiceType)
	}
	return e.choiceResult(spec, ctxData), nil
}

// HandleMultiSelectResult 多选；必须存在 OnMultiSelect，并显式返回消费结果。
func (e *Engine) HandleMultiSelectResult(playerID, choiceType string, selections []int, ctxData map[string]any) (HandleResult, error) {
	if e == nil || e.host == nil {
		return HandleResult{}, fmt.Errorf("choice: engine or host is nil")
	}
	if choiceType == "" {
		return HandleResult{}, fmt.Errorf("choice: empty choice_type for multi-select")
	}
	spec := e.reg.Get(choiceType)
	if spec == nil || spec.OnMultiSelect == nil {
		return HandleResult{}, fmt.Errorf("choice: unregistered or no OnMultiSelect for %q", choiceType)
	}
	handled, err := spec.OnMultiSelect(e.host, playerID, selections, ctxData)
	if err != nil {
		return HandleResult{}, err
	}
	if !handled {
		return HandleResult{}, fmt.Errorf("choice: OnMultiSelect returned unhandled for %q", choiceType)
	}
	return e.choiceResult(spec, ctxData), nil
}

// HandleCancelResult 取消；必须存在 OnCancel，并显式返回消费结果。
func (e *Engine) HandleCancelResult(playerID string, ctxData map[string]any) (HandleResult, error) {
	if e == nil || e.host == nil {
		return HandleResult{}, fmt.Errorf("choice: engine or host is nil")
	}
	choiceType, _ := ctxData["choice_type"].(string)
	if choiceType == "" {
		return HandleResult{}, fmt.Errorf("choice: context missing choice_type for cancel")
	}
	spec := e.reg.Get(choiceType)
	if spec == nil {
		return HandleResult{}, fmt.Errorf("choice: unregistered choice_type %q", choiceType)
	}
	if spec.OnCancel == nil {
		return HandleResult{}, fmt.Errorf("choice: unregistered or no OnCancel for %q", choiceType)
	}
	handled, err := spec.OnCancel(e.host, playerID, ctxData)
	if err != nil {
		return HandleResult{}, err
	}
	if !handled {
		return HandleResult{}, fmt.Errorf("choice: OnCancel returned unhandled for %q", choiceType)
	}
	return e.choiceResult(spec, ctxData), nil
}

func (e *Engine) choiceResult(spec *ChoiceSpec, ctxData map[string]any) HandleResult {
	return HandleResult{
		ConsumedInterrupt: choiceAutoConsume(spec, ctxData),
		AfterConsume:      choiceAfterConsume(spec, ctxData),
	}
}

func choiceAutoConsume(spec *ChoiceSpec, ctxData map[string]any) bool {
	if spec == nil {
		return false
	}
	if spec.ConsumesInterrupt != nil {
		return spec.ConsumesInterrupt(ctxData)
	}
	return spec.AutoConsume
}

func choiceAfterConsume(spec *ChoiceSpec, ctxData map[string]any) func(h Host) {
	if spec == nil || spec.AfterConsume == nil {
		return nil
	}
	return func(h Host) { spec.AfterConsume(h, ctxData) }
}
