// gameflow: 蝶舞者模块入口声明。

package butterfly_dancer

import (
	"starcup-engine/internal/engine/player"
	skills "starcup-engine/internal/engine/skill"
	"starcup-engine/internal/types"
)

// SkillEntries 导出角色技能与策略绑定入口。
func SkillEntries() []player.SkillEntry {
	return []player.SkillEntry{
		{ID: "bt_life_fire", Handler: &skills.ButterflyLifeFireHandler{}},
		{ID: "bt_dance", Handler: &skills.ButterflyDanceHandler{}},
		{ID: "bt_poison_pow", Handler: &skills.ButterflyPoisonPowderHandler{}},
		{ID: "bt_pilgrimage", Handler: &skills.ButterflyPilgrimageHandler{}},
		{ID: "bt_mirror", Handler: &skills.ButterflyMirrorHandler{}},
		{ID: "bt_wither", Handler: &skills.ButterflyWitherHandler{}},
		{ID: "bt_chrysalis", Handler: &skills.ButterflyChrysalisHandler{}},
		{ID: "bt_reverse_butterfly", Handler: &skills.ButterflyReverseHandler{}},
	}
}

// ChoiceRouteSpecs 导出角色 choice 路由声明。
func ChoiceRouteSpecs() map[string]types.ChoiceRouteSpec {
	return map[string]types.ChoiceRouteSpec{
		"bt_cocoon_pick":          types.ChoiceRouteRole("butterfly"),
		"bt_dance_discard":        types.ChoiceRouteRole("butterfly"),
		"bt_dance_mode":           types.ChoiceRouteRole("butterfly"),
		"bt_mirror_pair":          types.ChoiceRouteRole("butterfly"),
		"bt_pilgrimage_pick":      types.ChoiceRouteRole("butterfly"),
		"bt_poison_pick":          types.ChoiceRouteRole("butterfly"),
		"bt_reverse_branch1_pick": types.ChoiceRouteRole("butterfly"),
		"bt_reverse_branch2_cost": types.ChoiceRouteRole("butterfly"),
		"bt_reverse_mode":         types.ChoiceRouteRole("butterfly"),
		"bt_reverse_target":       types.ChoiceRouteRole("butterfly"),
		"bt_wither_confirm":       types.ChoiceRouteRole("butterfly"),
		"bt_wither_target":        types.ChoiceRouteRole("butterfly"),
	}
}
