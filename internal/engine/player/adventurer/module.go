// gameflow: 冒险家模块入口声明。

package adventurer

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/types"
)

// RoleEntry 导出角色统一入口定义。
func RoleEntry() player.RoleEntry {
	return player.RoleEntry{
		ID:               "adventurer",
		Choices:          NewChoiceHandler(),
		Skills:           SkillEntries(),
		ChoiceRouteSpecs: ChoiceRouteSpecs(),
		SkillUsabilityCheckers: map[string]player.SkillUsabilityChecker{
			"adventurer_fraud": CheckFraudUsability,
		},
	}
}

// SkillEntries 导出角色技能与策略绑定入口。
func SkillEntries() []player.SkillEntry {
	return []player.SkillEntry{
		{
			ID:      "adventurer_fraud",
			Handler: &AdventurerFraudHandler{},
			Policy: types.SkillPolicy{
				SkipAutoPhaseEnd: true,
			},
		},
		{ID: "adventurer_lucky_fortune", Handler: &AdventurerLuckyFortuneHandler{}},
		{ID: "adventurer_underground_law", Handler: &AdventurerUndergroundLawHandler{}},
		{ID: "adventurer_paradise", Handler: &AdventurerParadiseHandler{}},
		{ID: "adventurer_steal_sky", Handler: &AdventurerStealSkyHandler{}},
	}
}

// ChoiceRouteSpecs 导出角色 choice 路由声明。
func ChoiceRouteSpecs() map[string]types.ChoiceRouteSpec {
	return map[string]types.ChoiceRouteSpec{
		"adventurer_fraud_attack_element": types.ChoiceRouteRole("adventurer"),
		"adventurer_fraud_pick":           types.ChoiceRouteRole("adventurer"),
		"adventurer_paradise_pick":        types.ChoiceRouteRole("adventurer"),
		"adventurer_paradise_target":      types.ChoiceRouteRole("adventurer"),
		"adventurer_steal_sky_mode":       types.ChoiceRouteRole("adventurer"),
	}
}
