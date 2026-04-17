// gameflow: 风之剑圣模块入口声明。

package blade_master

import (
	"starcup-engine/internal/engine/player"
	skills "starcup-engine/internal/engine/skill"
)

// SkillEntries 导出角色技能与策略绑定入口。
func SkillEntries() []player.SkillEntry {
	return []player.SkillEntry{
		{ID: "wind_fury", Handler: &skills.WindFuryHandler{}},
		{ID: "holy_sword", Handler: &skills.HolySwordHandler{}},
		{ID: "sword_shadow", Handler: &skills.SwordShadowHandler{}},
		{ID: "gale_skill", Handler: &skills.GaleSkillHandler{}},
		{ID: "gale_slash", Handler: &skills.GaleSlashHandler{}},
	}
}
