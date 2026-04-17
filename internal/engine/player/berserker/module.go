// gameflow: 狂战士模块入口声明。

package berserker

import (
	"starcup-engine/internal/engine/player"
	skills "starcup-engine/internal/engine/skill"
)

// SkillEntries 导出角色技能与策略绑定入口。
func SkillEntries() []player.SkillEntry {
	return []player.SkillEntry{
		{ID: "berserker_frenzy", Handler: &skills.BerserkerFrenzyHandler{}},
		{ID: "berserker_tear", Handler: &skills.BerserkerTearHandler{}},
		{ID: "blood_roar", Handler: &skills.BloodRoarHandler{}},
		{ID: "blood_blade", Handler: &skills.BloodBladeHandler{}},
	}
}
