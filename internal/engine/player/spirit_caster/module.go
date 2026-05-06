// gameflow: 灵符师模块入口声明。

package spirit_caster

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
	"starcup-engine/internal/types"
)

// RoleEntry 导出角色统一入口定义。
func RoleEntry() player.RoleEntry {
	return player.RoleEntry{
		ID:               "spirit_caster",
		Defaults:         ApplyDefaults,
		Choices:          NewChoiceHandler(),
		Skills:           SkillEntries(),
		ChoiceRouteSpecs: ChoiceRouteSpecs(),
	}
}

// ApplyDefaults 初始化角色默认属性。
func ApplyDefaults(p *model.Player) {
	// spirit_caster 无特殊默认配置
}

// ChoiceRouteSpecs 导出角色 choice 路由声明。
func ChoiceRouteSpecs() map[string]types.ChoiceRouteSpec {
	return map[string]types.ChoiceRouteSpec{
		"sc_hundred_night_exclude_pick": types.ChoiceRouteRole("spirit_caster"),
		"sc_hundred_night_fire_reveal":  types.ChoiceRouteRole("spirit_caster"),
		"sc_hundred_night_power":        types.ChoiceRouteRole("spirit_caster"),
		"sc_hundred_night_target":       types.ChoiceRouteRole("spirit_caster"),
		"sc_incant_card":                types.ChoiceRouteRole("spirit_caster"),
		"sc_incant_confirm":             types.ChoiceRouteRole("spirit_caster"),
		"sc_spiritual_collapse_confirm": types.ChoiceRouteRole("spirit_caster"),
		"sc_talisman_wind_discard":      types.ChoiceRouteRole("spirit_caster"),
		"sc_talisman_pick":              types.ChoiceRouteRole("spirit_caster"),
	}
}

// SkillEntries 导出角色技能与策略绑定入口。
func SkillEntries() []player.SkillEntry {
	return []player.SkillEntry{
		{ID: "sc_talisman_thunder", Handler: &SpiritCasterTalismanThunderHandler{}},
		{ID: "sc_talisman_wind", Handler: &SpiritCasterTalismanWindHandler{}},
		{ID: "sc_incantation", Handler: &SpiritCasterIncantationHandler{}},
		{ID: "sc_hundred_night", Handler: &SpiritCasterHundredNightHandler{}},
		{ID: "sc_spiritual_collapse", Handler: &SpiritCasterSpiritualCollapseHandler{}},
	}
}
