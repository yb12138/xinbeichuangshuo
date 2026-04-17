// gameflow: 血色剑灵模块入口声明。

package crimson_sword_spirit

import (
	"starcup-engine/internal/engine/player"
	skills "starcup-engine/internal/engine/skill"
	"starcup-engine/internal/types"
)

// SkillEntries 导出角色技能与策略绑定入口。
func SkillEntries() []player.SkillEntry {
	return []player.SkillEntry{
		{ID: "css_blood_thorns", Handler: &skills.CrimsonBloodThornsHandler{}},
		{ID: "css_crimson_flash", Handler: &skills.CrimsonFlashHandler{}},
		{
			ID:      "css_blood_rose",
			Handler: &skills.CrimsonBloodRoseHandler{},
			Policy: types.SkillPolicy{
				TargetRules: types.TargetRuleSet{
					Count: types.TargetCountRule{Min: 2, Max: 2, Err: "血染蔷薇需要恰好指定2名目标"},
					Slots: []types.TargetSlotRule{
						{Index: 1, Camp: types.TargetCampAlly, Err: "血染蔷薇的第2个目标必须是我方角色"},
					},
				},
			},
		},
		{ID: "css_blood_barrier", Handler: &skills.CrimsonBloodBarrierHandler{}},
		{ID: "css_rose_courtyard", Handler: &skills.CrimsonRoseCourtyardHandler{}},
		{ID: "css_dance", Handler: &skills.CrimsonDanceHandler{}},
	}
}

// ChoiceRouteSpecs 导出角色 choice 路由声明。
func ChoiceRouteSpecs() map[string]types.ChoiceRouteSpec {
	return map[string]types.ChoiceRouteSpec{
		"css_dance_mode":          types.ChoiceRouteRole("crimson_sword_spirit"),
		"css_rose_courtyard_pick": types.ChoiceRouteRole("crimson_sword_spirit"),
	}
}
