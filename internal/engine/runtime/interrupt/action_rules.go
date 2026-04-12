// gameflow: 中断阶段 PlayerAction 路由表。

package interrupt

import (
	"fmt"

	"starcup-engine/internal/model"
)

// ActionHandler 处理某类中断下的玩家指令。
type ActionHandler func(e EngineInterface, act model.PlayerAction) error

// ActionRule 单条中断的动作规则。
type ActionRule struct {
	Allowed              map[model.PlayerActionType]bool
	InvalidActionMessage string
	Handler              ActionHandler
}

// ActionRules 动作规则注册表。
type ActionRules struct {
	rules map[model.InterruptType]*ActionRule
}

// NewActionRules 创建空表。
func NewActionRules() *ActionRules {
	return &ActionRules{rules: make(map[model.InterruptType]*ActionRule)}
}

// Register 注册规则；重复注册 panic。
func (r *ActionRules) Register(intrType model.InterruptType, rule *ActionRule) {
	if r == nil || rule == nil {
		return
	}
	if _, exists := r.rules[intrType]; exists {
		panic(fmt.Sprintf("interrupt: duplicate ActionRules.Register: %s", intrType))
	}
	r.rules[intrType] = rule
}

// Get 获取规则。
func (r *ActionRules) Get(intrType model.InterruptType) *ActionRule {
	if r == nil {
		return nil
	}
	return r.rules[intrType]
}
