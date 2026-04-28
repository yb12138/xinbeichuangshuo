// gameflow: Skill Handler 注册表核心。

package skills

import (
	"starcup-engine/internal/engine/player"
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

// BaseHandler 是 player.BaseHandler 的类型别名，保持向后兼容。
type BaseHandler = player.BaseHandler

// InitHandlers 空函数（保留兼容，实际注册在 engine/skill_registry.go）
func InitHandlers() {
}
