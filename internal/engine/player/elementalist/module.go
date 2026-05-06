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
					Count: types.TargetCountRule{Min: 2, Max: 2, Err: "冰冻需要指定2名目标"},
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
		"elementalist_bonus_card":      types.ChoiceRouteRole("elementalist"),
		"elementalist_primordial_pick": types.ChoiceRouteRole("elementalist"),
	}
}
