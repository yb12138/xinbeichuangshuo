// gameflow: 红莲骑士模块入口声明。

package crimson_knight

import (
	"starcup-engine/internal/engine/player"
	skills "starcup-engine/internal/engine/skill"
	"starcup-engine/internal/types"
)

// SkillEntries 导出角色技能与策略绑定入口。
func SkillEntries() []player.SkillEntry {
	return []player.SkillEntry{
		{ID: "crk_crimson_pact", Handler: &skills.CrimsonKnightCrimsonPactHandler{}},
		{ID: "crk_crimson_faith", Handler: &skills.CrimsonKnightCrimsonFaithHandler{}},
		{ID: "crk_bloody_prayer", Handler: &skills.CrimsonKnightBloodyPrayerHandler{}},
		{ID: "crk_killing_feast", Handler: &skills.CrimsonKnightKillingFeastHandler{}},
		{ID: "crk_hot_blood", Handler: &skills.CrimsonKnightHotBloodHandler{}},
		{ID: "crk_calm_mind", Handler: &skills.CrimsonKnightCalmMindHandler{}},
		{ID: "crk_crimson_cross", Handler: &skills.CrimsonKnightCrimsonCrossHandler{}},
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
