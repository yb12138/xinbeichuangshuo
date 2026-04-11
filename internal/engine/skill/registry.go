// gameflow: Skill Handler 注册表核心。

package skills

import (
	"starcup-engine/internal/model"
)

var registry = make(map[string]model.SkillHandler)

// Register 注册技能逻辑
func Register(id string, handler model.SkillHandler) {
	if _, exists := registry[id]; !exists {
		registry[id] = handler
	}
}

// GetHandler 获取技能逻辑
func GetHandler(id string) model.SkillHandler {
	return registry[id]
}

// BaseHandler 基础处理器，简化实现
type BaseHandler struct{}

func (h *BaseHandler) CanUse(ctx *model.Context) bool {
	return true
}

func (h *BaseHandler) Execute(ctx *model.Context) error {
	return nil
}

// InitHandlers 空函数（保留兼容，实际注册在 engine/skill_registry.go）
func InitHandlers() {
}
