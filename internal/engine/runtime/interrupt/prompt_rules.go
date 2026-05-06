// gameflow: 中断阶段 Prompt 构建表。

package interrupt

import (
	"fmt"

	"starcup-engine/internal/model"
)

// PromptBuilder 由宿主实现具体 UI 提示。
type PromptBuilder func(e EngineInterface) *model.Prompt

// PromptRules 提示构建注册表。
type PromptRules struct {
	builders map[model.InterruptType]PromptBuilder
}

// NewPromptRules 创建空表。
func NewPromptRules() *PromptRules {
	return &PromptRules{builders: make(map[model.InterruptType]PromptBuilder)}
}

// Register 注册构建器；重复注册 panic。
func (r *PromptRules) Register(intrType model.InterruptType, builder PromptBuilder) {
	if r == nil || builder == nil {
		return
	}
	if _, exists := r.builders[intrType]; exists {
		panic(fmt.Sprintf("interrupt: duplicate PromptRules.Register: %s", intrType))
	}
	r.builders[intrType] = builder
}

// Get 获取构建器。
func (r *PromptRules) Get(intrType model.InterruptType) PromptBuilder {
	if r == nil {
		return nil
	}
	return r.builders[intrType]
}
