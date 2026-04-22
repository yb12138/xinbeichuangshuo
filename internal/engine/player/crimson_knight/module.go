// gameflow: 红莲骑士模块入口声明。

package crimson_knight

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
	"starcup-engine/internal/types"
)

// RoleEntry 导出角色统一入口定义。
func RoleEntry() player.RoleEntry {
	return player.RoleEntry{
		ID:               "crimson_knight",
		Defaults:         ApplyDefaults,
		Choices:          NewChoiceHandler(),
		Skills:           SkillEntries(),
		ChoiceRouteSpecs: ChoiceRouteSpecs(),
		TimingHookSpecs: []player.TimingHookSpec{
			{Timing: player.TimingOnHealResist, Priority: 200, Hook: healResistHook},
			{Timing: player.TimingOnTurnEnd, Priority: 600, Hook: turnEndHook},
			{Timing: player.TimingOnMoraleLossApplied, Priority: 100, Hook: moraleLossHook},
		},
	}
}

// ApplyDefaults 初始化角色默认属性。
func ApplyDefaults(p *model.Player) {
	if p == nil {
		return
	}
	p.MaxHeal = 4
	if p.Tokens == nil {
		p.Tokens = map[string]int{}
	}
	p.Tokens["crk_blood_mark"] = 0
}

// SkillEntries 导出角色技能与策略绑定入口。
func SkillEntries() []player.SkillEntry {
	return []player.SkillEntry{
		{ID: "crk_crimson_pact", Handler: &CrimsonKnightCrimsonPactHandler{}},
		{ID: "crk_crimson_faith", Handler: &CrimsonKnightCrimsonFaithHandler{}},
		{ID: "crk_bloody_prayer", Handler: &CrimsonKnightBloodyPrayerHandler{}},
		{ID: "crk_killing_feast", Handler: &CrimsonKnightKillingFeastHandler{}},
		{ID: "crk_hot_blood", Handler: &CrimsonKnightHotBloodHandler{}},
		{ID: "crk_calm_mind", Handler: &CrimsonKnightCalmMindHandler{}},
		{ID: "crk_crimson_cross", Handler: &CrimsonKnightCrimsonCrossHandler{}},
	}
}

// ChoiceRouteSpecs 导出角色 choice 路由声明。
func ChoiceRouteSpecs() map[string]types.ChoiceRouteSpec {
	return map[string]types.ChoiceRouteSpec{
		"crk_bloody_prayer_ally_count": types.ChoiceRouteRole("crimson_knight"),
		"crk_bloody_prayer_pick":       types.ChoiceRouteRole("crimson_knight"),
		"crk_bloody_prayer_split":      types.ChoiceRouteRole("crimson_knight"),
		"crk_bloody_prayer_target":     types.ChoiceRouteRole("crimson_knight"),
		"crk_bloody_prayer_x":          types.ChoiceRouteRole("crimson_knight"),
	}
}
