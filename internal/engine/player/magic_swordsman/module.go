// gameflow: 魔剑士模块入口声明。

package magic_swordsman

import (
	"starcup-engine/internal/engine/player"
	skills "starcup-engine/internal/engine/skill"
	"starcup-engine/internal/types"
)

// SkillEntries 导出角色技能与策略绑定入口。
func SkillEntries() []player.SkillEntry {
	return []player.SkillEntry{
		{ID: "ms_asura_combo", Handler: &skills.MagicSwordsmanAsuraComboHandler{}},
		{ID: "ms_shadow_gather", Handler: &skills.MagicSwordsmanShadowGatherHandler{}},
		{ID: "ms_shadow_power", Handler: &skills.MagicSwordsmanShadowPowerHandler{}},
		{ID: "ms_shadow_reject", Handler: &skills.MagicSwordsmanShadowRejectHandler{}},
		{ID: "ms_shadow_meteor", Handler: &skills.MagicSwordsmanShadowMeteorHandler{}},
		{ID: "ms_yellow_spring", Handler: &skills.MagicSwordsmanYellowSpringHandler{}},
	}
}

// ChoiceRouteSpecs 导出角色 choice 路由声明。
func ChoiceRouteSpecs() map[string]types.ChoiceRouteSpec {
	return map[string]types.ChoiceRouteSpec{
		"ms_asura_combo_pick":              types.ChoiceRouteRole("magic_swordsman"),
		"ms_shadow_gather_pick":            types.ChoiceRouteRole("magic_swordsman"),
		"ms_shadow_meteor_release_confirm": types.ChoiceRouteRole("magic_swordsman"),
	}
}
