// gameflow: ChoiceSpec 单点声明一条 choice_type 的 Prompt 与输入处理。

package choice

import "starcup-engine/internal/model"

// HandleResult 描述一次 choice 输入是否完成当前中断。
type HandleResult struct {
	ConsumedInterrupt bool
	AfterConsume      func(h Host)
}

// ChoiceSpec 描述一种 InterruptChoice 的交互；未注册或未实现字段由 Engine 返回错误。
type ChoiceSpec struct {
	Type string

	AutoConsume         bool
	ConsumesInterrupt   func(ctxData map[string]any) bool
	SequentialRemaining func(ctxData map[string]any) (int, bool)

	BuildPrompt func(h Host, choiceType, playerID string, player *model.Player, data map[string]any) *model.Prompt

	OnSelect func(h Host, playerID string, selectionIndex int, ctxData map[string]any) (bool, error)

	OnMultiSelect func(h Host, playerID string, selections []int, ctxData map[string]any) (bool, error)

	OnCancel func(h Host, playerID string, ctxData map[string]any) (bool, error)

	AfterConsume func(h Host, ctxData map[string]any)
}
