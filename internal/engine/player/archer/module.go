// gameflow: 神箭手模块入口声明。

package archer

import (
	"starcup-engine/internal/engine/player"
)

// RoleEntry 导出角色统一入口定义。
func RoleEntry() player.RoleEntry {
	return player.RoleEntry{
		ID:      "archer",
		Choices: NewChoiceHandler(),
		Skills:  SkillEntries(),
	}
}

// SkillEntries 导出角色技能与策略绑定入口。
func SkillEntries() []player.SkillEntry {
	return []player.SkillEntry{
		{ID: "piercing_shot", Handler: &PiercingShotHandler{}},
		{ID: "lightning_arrow", Handler: &LightningArrowHandler{}},
		{ID: "snipe", Handler: &SnipeHandler{}},
		{ID: "precise_shot", Handler: &PreciseShotHandler{}},
		{ID: "flash_trap", Handler: &FlashTrapHandler{}},
	}
}
