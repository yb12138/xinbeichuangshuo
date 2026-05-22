// gameflow: 元素师模块入口声明。

package elementalist

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/types"
)

// RoleEntry 导出角色统一入口定义。
func RoleEntry() player.RoleEntry {
	return player.RoleEntry{
		ID:               "elementalist",
		Choices:          NewChoiceHandler(),
		Skills:           SkillEntries(),
		ChoiceRouteSpecs: ChoiceRouteSpecs(),
		SkillUsabilityCheckers: map[string]player.SkillUsabilityChecker{
			"elementalist_ignite": CheckIgniteUsability,
		},
	}
}

// SkillEntries 导出角色技能与策略绑定入口。
func SkillEntries() []player.SkillEntry {
	return []player.SkillEntry{
		{ID: "elementalist_absorb", Handler: &ElementalistAbsorbHandler{}},
		{ID: "elementalist_ignite", Handler: &ElementalistIgniteHandler{}},
		{ID: "elementalist_thunder_strike", Handler: &ElementalistThunderStrikeHandler{}},
		{
			ID:      "elementalist_freeze",
			Handler: &ElementalistFreezeHandler{},
			Policy: types.SkillPolicy{
				TargetRules: types.TargetRuleSet{
					// 分步选择模式：由后端流程控制，前端不需要一次性选目标
					// 使用 Err 字段触发 HasCountOverride，使 min/max=0生效
					Count: types.TargetCountRule{Min: 0, Max: 0, Err: "分步选择"},
				},
			},
		},
		{ID: "elementalist_wind_blade", Handler: &ElementalistWindBladeHandler{}},
		{ID: "elementalist_meteor", Handler: &ElementalistMeteorHandler{}},
		{ID: "elementalist_fireball", Handler: &ElementalistFireballHandler{}},
		{ID: "elementalist_moonlight", Handler: &ElementalistMoonlightHandler{}},
	}
}

// ChoiceRouteSpecs 导出角色 choice 路由声明。
func ChoiceRouteSpecs() map[string]types.ChoiceRouteSpec {
	return map[string]types.ChoiceRouteSpec{
		"elementalist_bonus_card":           types.ChoiceRouteRole("elementalist"),
		"elementalist_primordial_pick":      types.ChoiceRouteRole("elementalist"),
		"elementalist_freeze_damage_target": types.ChoiceRouteRole("elementalist"),
		"elementalist_freeze_heal_target":   types.ChoiceRouteRole("elementalist"),
	}
}
