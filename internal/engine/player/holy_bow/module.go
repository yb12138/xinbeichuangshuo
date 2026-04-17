// gameflow: 圣弓模块入口声明。

package holy_bow

import (
	"starcup-engine/internal/engine/player"
	skills "starcup-engine/internal/engine/skill"
	"starcup-engine/internal/types"
)

// SkillEntries 导出角色技能与策略绑定入口。
func SkillEntries() []player.SkillEntry {
	return []player.SkillEntry{
		{ID: "hb_heavenly_bow", Handler: &skills.HolyBowHeavenlyBowHandler{}},
		{ID: "hb_holy_shard_storm", Handler: &skills.HolyBowShardStormHandler{}},
		{ID: "hb_radiant_descent", Handler: &skills.HolyBowRadiantDescentHandler{}},
		{ID: "hb_light_burst", Handler: &skills.HolyBowLightBurstHandler{}},
		{ID: "hb_meteor_bullet", Handler: &skills.HolyBowMeteorBulletHandler{}},
		{ID: "hb_radiant_cannon", Handler: &skills.HolyBowRadiantCannonHandler{}},
		{ID: "hb_auto_fill", Handler: &skills.HolyBowAutoFillHandler{}},
	}
}

// ChoiceRouteSpecs 导出角色 choice 路由声明。
func ChoiceRouteSpecs() map[string]types.ChoiceRouteSpec {
	return map[string]types.ChoiceRouteSpec{
		"hb_auto_fill_gain":              types.ChoiceRouteRole("holy_bow"),
		"hb_auto_fill_resource":          types.ChoiceRouteRole("holy_bow"),
		"hb_holy_shard_combo":            types.ChoiceRouteRole("holy_bow"),
		"hb_holy_shard_miss_ally_target": types.ChoiceRouteRole("holy_bow"),
		"hb_holy_shard_miss_confirm":     types.ChoiceRouteRole("holy_bow"),
		"hb_holy_shard_miss_x":           types.ChoiceRouteRole("holy_bow"),
		"hb_holy_shard_target":           types.ChoiceRouteRole("holy_bow"),
		"hb_light_burst_mode":            types.ChoiceRouteRole("holy_bow"),
		"hb_light_burst_mode_a_target":   types.ChoiceRouteRole("holy_bow"),
		"hb_light_burst_mode_b_discard":  types.ChoiceRouteRole("holy_bow"),
		"hb_light_burst_mode_b_targets":  types.ChoiceRouteRole("holy_bow"),
		"hb_light_burst_mode_b_x":        types.ChoiceRouteRole("holy_bow"),
		"hb_meteor_bullet_cost":          types.ChoiceRouteRole("holy_bow"),
		"hb_meteor_bullet_target":        types.ChoiceRouteRole("holy_bow"),
		"hb_radiant_cannon_side":         types.ChoiceRouteRole("holy_bow"),
		"hb_radiant_descent_cost":        types.ChoiceRouteRole("holy_bow"),
		"hb_radiant_descent_pick":        types.ChoiceRouteRole("holy_bow"),
	}
}
