// gameflow: 月神模块入口声明。

package moon

import (
	"starcup-engine/internal/engine/player"
	skills "starcup-engine/internal/engine/skill"
	"starcup-engine/internal/types"
)

// SkillEntries 导出角色技能与策略绑定入口。
func SkillEntries() []player.SkillEntry {
	return []player.SkillEntry{
		{ID: "mg_new_moon_shelter", Handler: &skills.MoonGoddessNewMoonShelterHandler{}},
		{ID: "mg_dark_moon_curse", Handler: &skills.MoonGoddessDarkMoonCurseHandler{}},
		{ID: "mg_medusa_eye", Handler: &skills.MoonGoddessMedusaEyeHandler{}},
		{ID: "mg_moon_cycle", Handler: &skills.MoonGoddessMoonCycleHandler{}},
		{ID: "mg_blasphemy", Handler: &skills.MoonGoddessBlasphemyHandler{}},
		{ID: "mg_darkmoon_slash", Handler: &skills.MoonGoddessDarkMoonSlashHandler{}},
		{ID: "mg_pale_moon", Handler: &skills.MoonGoddessPaleMoonHandler{}},
	}
}

// ChoiceRouteSpecs 导出角色 choice 路由声明。
func ChoiceRouteSpecs() map[string]types.ChoiceRouteSpec {
	return map[string]types.ChoiceRouteSpec{
		"mg_blasphemy_target":       types.ChoiceRouteRole("moon_goddess"),
		"mg_dark_moon_curse_pick":   types.ChoiceRouteRole("moon_goddess"),
		"mg_darkmoon_slash_x":       types.ChoiceRouteRole("moon_goddess"),
		"mg_medusa_darkmoon_pick":   types.ChoiceRouteRole("moon_goddess"),
		"mg_medusa_magic_discard":   types.ChoiceRouteRole("moon_goddess"),
		"mg_moon_cycle_heal_target": types.ChoiceRouteRole("moon_goddess"),
		"mg_moon_cycle_mode":        types.ChoiceRouteRole("moon_goddess"),
		"mg_pale_moon_discard":      types.ChoiceRouteRole("moon_goddess"),
		"mg_pale_moon_mode":         types.ChoiceRouteRole("moon_goddess"),
		"mg_pale_moon_target":       types.ChoiceRouteRole("moon_goddess"),
		"mg_pale_moon_x":            types.ChoiceRouteRole("moon_goddess"),
	}
}
