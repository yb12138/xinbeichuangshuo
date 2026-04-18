// gameflow: 魔弓模块入口声明。

package magic_bow

import (
	"starcup-engine/internal/engine/player"
	skills "starcup-engine/internal/engine/skill"
	"starcup-engine/internal/types"
)

// RoleEntry 导出角色统一入口定义。
func RoleEntry() player.RoleEntry {
	return player.RoleEntry{
		ID:               "magic_bow",
		Choices:          NewChoiceHandler(),
		Skills:           SkillEntries(),
		ChoiceRouteSpecs: ChoiceRouteSpecs(),
	}
}

// SkillEntries 导出角色技能与策略绑定入口。
func SkillEntries() []player.SkillEntry {
	return []player.SkillEntry{
		{ID: "mb_magic_pierce", Handler: &skills.MagicBowMagicPierceHandler{}},
		{
			ID:      "mb_thunder_scatter",
			Handler: &skills.MagicBowThunderScatterHandler{},
			Policy: types.SkillPolicy{
				TargetRules: types.TargetRuleSet{
					Count: types.TargetCountRule{Min: 0, Max: 1, Err: "雷光散射至多指定1名敌方角色作为额外目标"},
					Slots: []types.TargetSlotRule{
						{Index: 0, Camp: types.TargetCampEnemy, Err: "雷光散射的额外目标必须是敌方角色"},
					},
				},
			},
		},
		{ID: "mb_multi_shot", Handler: &skills.MagicBowMultiShotHandler{}},
		{ID: "mb_charge", Handler: &skills.MagicBowChargeHandler{}},
		{
			ID:      "mb_demon_eye",
			Handler: &skills.MagicBowDemonEyeHandler{},
			Policy: types.SkillPolicy{
				TargetRules: types.TargetRuleSet{
					Count: types.TargetCountRule{Min: 0, Max: 1, Err: "魔眼需要且仅能指定1名其他角色"},
					Slots: []types.TargetSlotRule{
						{Index: 0, Self: types.TargetSelfOther, Err: "魔眼不能以自己为目标"},
					},
				},
			},
		},
		// 内部回调技能：用于"充能"弃牌后的继续流程
		{ID: "mb_charge_followup_discard", Handler: &skills.MagicBowChargeFollowupDiscardHandler{}},
	}
}

// ChoiceRouteSpecs 导出角色 choice 路由声明。
func ChoiceRouteSpecs() map[string]types.ChoiceRouteSpec {
	return map[string]types.ChoiceRouteSpec{
		"mb_charge_discard_pick":    types.ChoiceRouteRole("magic_bow"),
		"mb_charge_draw_x":          types.ChoiceRouteRole("magic_bow"),
		"mb_charge_place_cards":     types.ChoiceRouteRole("magic_bow"),
		"mb_charge_place_count":     types.ChoiceRouteRole("magic_bow"),
		"mb_demon_eye_charge_card":  types.ChoiceRouteRole("magic_bow"),
		"mb_demon_eye_pick":         types.ChoiceRouteRole("magic_bow"),
		"mb_demon_eye_target":       types.ChoiceRouteTargetPrompt("mb"),
		"mb_multi_shot_target":      types.ChoiceRouteTargetPrompt("mb"),
		"mb_thunder_scatter_extra":  types.ChoiceRouteRole("magic_bow"),
		"mb_thunder_scatter_target": types.ChoiceRouteTargetPrompt("mb"),
	}
}
