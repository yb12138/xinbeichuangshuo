// gameflow: 冒险家模块入口声明。

package adventurer

import (
	"starcup-engine/internal/engine/player"
	skills "starcup-engine/internal/engine/skill"
	"starcup-engine/internal/model"
	"starcup-engine/internal/types"
)

// RoleEntry 导出角色统一入口定义。
func RoleEntry() player.RoleEntry {
	return player.RoleEntry{
		ID:               "adventurer",
		Defaults:         ApplyDefaults,
		Choices:          NewChoiceHandler(),
		Skills:           SkillEntries(),
		ChoiceRouteSpecs: ChoiceRouteSpecs(),
	}
}

// ApplyDefaults 初始化角色默认属性。
func ApplyDefaults(p *model.Player) {
	// adventurer 无特殊默认配置
}

// SkillEntries 导出角色技能与策略绑定入口。
func SkillEntries() []player.SkillEntry {
	return []player.SkillEntry{
		{
			ID:      "adventurer_fraud",
			Handler: &skills.AdventurerFraudHandler{},
			Policy: types.SkillPolicy{
				SkipAutoPhaseEnd: true,
			},
		},
		{ID: "adventurer_lucky_fortune", Handler: &skills.AdventurerLuckyFortuneHandler{}},
		{ID: "adventurer_underground_law", Handler: &skills.AdventurerUndergroundLawHandler{}},
		{ID: "adventurer_paradise", Handler: &skills.AdventurerParadiseHandler{}},
		{ID: "adventurer_steal_sky", Handler: &skills.AdventurerStealSkyHandler{}},
	}
}

// ChoiceRouteSpecs 导出角色 choice 路由声明。
func ChoiceRouteSpecs() map[string]types.ChoiceRouteSpec {
	return map[string]types.ChoiceRouteSpec{
		"adventurer_fraud_attack_element":   types.ChoiceRouteRole("adventurer"),
		"adventurer_fraud_pick":             types.ChoiceRouteRole("adventurer"),
		"adventurer_paradise_pick":          types.ChoiceRouteRole("adventurer"),
		"adventurer_paradise_target":        types.ChoiceRouteRole("adventurer"),
		"adventurer_steal_sky_extra_action": types.ChoiceRouteRole("adventurer"),
		"adventurer_steal_sky_mode":         types.ChoiceRouteRole("adventurer"),
	}
}
