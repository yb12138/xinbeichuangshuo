// gameflow: 灵符师模块入口声明。

package spirit_caster

import (
	"starcup-engine/internal/engine/player"
	skills "starcup-engine/internal/engine/skill"
	"starcup-engine/internal/model"
)

// RoleEntry 导出角色统一入口定义。
func RoleEntry() player.RoleEntry {
	return player.RoleEntry{
		ID:     "spirit_caster",
		Defaults: ApplyDefaults,
		Choices: NewChoiceHandler(),
		Skills:   SkillEntries(),
	}
}

// ApplyDefaults 初始化角色默认属性。
func ApplyDefaults(p *model.Player) {
	// spirit_caster 无特殊默认配置
}

// SkillEntries 导出角色技能与策略绑定入口。
func SkillEntries() []player.SkillEntry {
	return []player.SkillEntry{
		{ID: "sc_talisman_thunder", Handler: &skills.SpiritCasterTalismanThunderHandler{}},
		{ID: "sc_talisman_wind", Handler: &skills.SpiritCasterTalismanWindHandler{}},
		{ID: "sc_incantation", Handler: &skills.SpiritCasterIncantationHandler{}},
		{ID: "sc_hundred_night", Handler: &skills.SpiritCasterHundredNightHandler{}},
		{ID: "sc_spiritual_collapse", Handler: &skills.SpiritCasterSpiritualCollapseHandler{}},
	}
}
