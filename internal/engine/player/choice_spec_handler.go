// gameflow: 声明式 ChoiceSpec 通用处理器。

package player

import (
	"strconv"

	"starcup-engine/internal/model"
)

// SpecChoiceHandler 声明式选择处理器（使用 ChoiceSpec 注册表）。
type SpecChoiceHandler struct {
	specs map[string]ChoiceSpec
}

// NewSpecChoiceHandler 从 ChoiceSpec 列表创建处理器。
func NewSpecChoiceHandler(specs []ChoiceSpec) ChoiceHandler {
	m := make(map[string]ChoiceSpec, len(specs))
	for _, spec := range specs {
		m[spec.ChoiceType] = spec
	}
	return &SpecChoiceHandler{specs: m}
}

func (h *SpecChoiceHandler) BuildPrompt(rt ChoiceRuntime, choiceType, playerID string, player *model.Player, data map[string]any) *model.Prompt {
	spec, ok := h.specs[choiceType]
	if !ok || spec.BuildPrompt == nil {
		return nil
	}
	return spec.BuildPrompt(rt, playerID, player, data)
}

func (h *SpecChoiceHandler) HandleChoice(rt ChoiceRuntime, playerID string, selectionIndex int, ctxData map[string]any) (bool, error) {
	choiceType, _ := ctxData["choice_type"].(string)
	spec, ok := h.specs[choiceType]
	if !ok || spec.HandleChoice == nil {
		return false, nil
	}
	return spec.HandleChoice(rt, playerID, selectionIndex, ctxData)
}

// ---------------------------------------------------------------------------
// Prompt 构建辅助函数（简化重复代码）
// ---------------------------------------------------------------------------

// PromptBuilder 简化 Prompt 构建的辅助结构。
type PromptBuilder struct {
	p *model.Prompt
}

// NewPrompt 创建 PromptBuilder。
func NewPrompt(playerID string, message string) *PromptBuilder {
	return &PromptBuilder{
		p: &model.Prompt{
			Type:     model.PromptConfirm,
			PlayerID: playerID,
			Message:  message,
			Min:      1,
			Max:      1,
		},
	}
}

// Options 设置选项列表。
func (b *PromptBuilder) Options(opts ...PromptOptionSpec) *PromptBuilder {
	options := make([]model.PromptOption, len(opts))
	for i, opt := range opts {
		options[i] = model.PromptOption{ID: opt.ID, Label: opt.Label}
	}
	b.p.Options = options
	return b
}

// OptionsFromLabels 从标签列表创建选项（ID 自动为索引）。
func (b *PromptBuilder) OptionsFromLabels(labels ...string) *PromptBuilder {
	options := make([]model.PromptOption, len(labels))
	for i, label := range labels {
		options[i] = model.PromptOption{ID: intToString(i), Label: label}
	}
	b.p.Options = options
	return b
}

// Build 返回最终 Prompt。
func (b *PromptBuilder) Build() *model.Prompt {
	return b.p
}

// PromptOptionSpec 选项规格。
type PromptOptionSpec struct {
	ID    string
	Label string
}

// Option 创建选项规格。
func Option(id, label string) PromptOptionSpec {
	return PromptOptionSpec{ID: id, Label: label}
}

// intToString 快速整数转字符串（用于选项 ID）。
func intToString(i int) string {
	if i >= 0 && i < 10 {
		return digits[i : i+1]
	}
	return strconv.Itoa(i)
}

var digits = "0123456789"
