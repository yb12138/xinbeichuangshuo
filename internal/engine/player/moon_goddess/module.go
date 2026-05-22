// gameflow: 月神模块入口声明。

package moon

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
	"starcup-engine/internal/types"
)

// RoleEntry 导出角色统一入口定义。
func RoleEntry() player.RoleEntry {
	return player.RoleEntry{
		ID:               "moon_goddess",
		Defaults:         ApplyDefaults,
		Choices:          NewChoiceHandler(),
		Skills:           SkillEntries(),
		ChoiceRouteSpecs: ChoiceRouteSpecs(),
		TimingHookSpecs: []player.TimingHookSpec{
			{Timing: player.TimingAttackNoResponse, Priority: 200, Hook: attackGatingHook},
			{Timing: player.TimingTurnEndPreExtra, Priority: 100, Hook: turnEndMoonCycleHook},
			{Timing: player.TimingPostDamageResolved, Priority: 900, Hook: postDamageResolvedHook},
			{Timing: player.TimingTurnEndFinal, Priority: 100, Hook: turnEndFinalHook},
			{Timing: player.TimingAttackDeclareInterrupt, Priority: 100, Hook: medusaInterruptHook},
		},
	}
}

// ApplyDefaults 初始化月神的基础指示物。
func ApplyDefaults(p *model.Player) {
	if p == nil || p.Tokens == nil {
		return
	}
	p.Tokens["mg_new_moon"] = 0
	p.Tokens["mg_petrify"] = 0
}

// SkillEntries 导出角色技能与策略绑定入口。
func SkillEntries() []player.SkillEntry {
	return []player.SkillEntry{
		{ID: "mg_new_moon_shelter", Handler: &MoonGoddessNewMoonShelterHandler{}},
		{ID: "mg_dark_moon_curse", Handler: &MoonGoddessDarkMoonCurseHandler{}},
		{ID: "mg_medusa_eye", Handler: &MoonGoddessMedusaEyeHandler{}},
		{ID: "mg_moon_cycle", Handler: &MoonGoddessMoonCycleHandler{}},
		{ID: "mg_blasphemy", Handler: &MoonGoddessBlasphemyHandler{}},
		{ID: "mg_darkmoon_slash", Handler: &MoonGoddessDarkMoonSlashHandler{}},
		{ID: "mg_pale_moon", Handler: &MoonGoddessPaleMoonHandler{}},
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
