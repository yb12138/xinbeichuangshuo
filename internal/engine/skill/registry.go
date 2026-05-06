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
