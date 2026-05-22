// gameflow: 狂战士模块入口声明。

package berserker

import (
	"starcup-engine/internal/engine/player"
)

// RoleEntry 导出角色统一入口定义。
func RoleEntry() player.RoleEntry {
	return player.RoleEntry{
		ID:      "berserker",
		Choices: NewChoiceHandler(),
		Skills:  SkillEntries(),
	}
}

// SkillEntries 导出角色技能与策略绑定入口。
func SkillEntries() []player.SkillEntry {
	return []player.SkillEntry{
		{ID: "berserker_frenzy", Handler: &BerserkerFrenzyHandler{}},
		{ID: "berserker_tear", Handler: &BerserkerTearHandler{}},
		{ID: "blood_roar", Handler: &BloodRoarHandler{}},
		{ID: "blood_blade", Handler: &BloodBladeHandler{}},
	}
}
