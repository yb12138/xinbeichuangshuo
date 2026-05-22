// gameflow: 冒险家模块入口声明。

package adventurer

import (
	"fmt"

	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
	"starcup-engine/internal/types"
)

// RoleEntry 导出角色统一入口定义。
func RoleEntry() player.RoleEntry {
	return player.RoleEntry{
		ID:               "adventurer",
		Choices:          NewChoiceHandler(),
		ChoiceSpecs:      ChoiceSpecs(),
		Skills:           SkillEntries(),
		ChoiceRouteSpecs: ChoiceRouteSpecs(),
		TimingHookSpecs: []player.TimingHookSpec{
			{Timing: player.TimingSpecialActionOverride, Priority: 100, Hook: undergroundLawOverrideHook},
			{Timing: player.TimingSpecialActionOverride, Priority: 200, Hook: extractOverrideHook},
		},
		SpecialActionHook: player.SpecialActionHookSpec{
			BuyRewardOverride: func(p *model.Player, campStones int, maxStones int) player.BuyRewardResult {
				if campStones >= maxStones {
					return player.BuyRewardResult{Handled: true}
				}
				return player.BuyRewardResult{
					Handled:    true,
					AddGems:    2,
					LogMessage: fmt.Sprintf("[Skill] %s 使用了技能: 地下法则（购买改写：战绩区+2宝石）", p.Name),
				}
			},
		},
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
		{ID: "adventurer_steal_sky", Handler: &AdventurerStealSkyHandler{}},
	}
}

// ChoiceRouteSpecs 导出角色 choice 路由声明。
func ChoiceRouteSpecs() map[string]types.ChoiceRouteSpec {
	return map[string]types.ChoiceRouteSpec{
		"adventurer_extract_paradise_check": types.ChoiceRouteRole("adventurer"),
		"adventurer_fraud_attack_element":   types.ChoiceRouteRole("adventurer"),
		"adventurer_fraud_pick":             types.ChoiceRouteRole("adventurer"),
		"adventurer_paradise_pick":          types.ChoiceRouteRole("adventurer"),
		"adventurer_steal_sky_mode":         types.ChoiceRouteRole("adventurer"),
	}
}
