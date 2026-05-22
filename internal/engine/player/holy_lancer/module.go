// gameflow: 圣枪骑士模块入口声明。

package holy_lancer

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
	"starcup-engine/internal/types"
)

// RoleEntry 导出角色统一入口定义。
func RoleEntry() player.RoleEntry {
	return player.RoleEntry{
		ID:               "holy_lancer",
		Defaults:         ApplyDefaults,
		Choices:          NewChoiceHandler(),
		Skills:           SkillEntries(),
		ChoiceRouteSpecs: ChoiceRouteSpecs(),
		TimingHookSpecs: []player.TimingHookSpec{
			{Timing: player.TimingAttackStateReset, Priority: 100, Hook: attackStateResetHook},
			{Timing: player.TimingAttackNoResponse, Priority: 200, Hook: attackGatingHook},
			{Timing: player.TimingTurnEndFinal, Priority: 900, Hook: turnEndHook},
			{Timing: player.TimingResponseSkillSkip, Priority: 100, Hook: responseSkillSkipHook},
			{Timing: player.TimingPlayerSetup, Priority: 100, Hook: playerSetupHook},
			{Timing: player.TimingCampCupChanged, Priority: 100, Hook: campCupChangedHook},
		},
		SkillUsabilityCheckers: map[string]player.SkillUsabilityChecker{
			"holy_lancer_punishment": CheckPunishmentUsability,
		},
	}
}

// ApplyDefaults 初始化角色默认属性。
func ApplyDefaults(p *model.Player) {
	if p == nil {
		return
	}
	p.MaxHeal = 2
}

// SkillEntries 导出角色技能与策略绑定入口。
func SkillEntries() []player.SkillEntry {
	return []player.SkillEntry{
		{ID: "holy_lancer_revelation", Handler: &HolyLancerRevelationHandler{}},
		{ID: "holy_lancer_radiance", Handler: &HolyLancerRadianceHandler{}},
		{
			ID:      "holy_lancer_punishment",
			Handler: &HolyLancerPunishmentHandler{},
			Policy: types.SkillPolicy{
				TargetRules: types.TargetRuleSet{
					Count: types.TargetCountRule{Min: 1, Max: 1, Err: "惩戒需要指定1名其他角色"},
					Slots: []types.TargetSlotRule{
						{Index: 0, Self: types.TargetSelfOther, Err: "惩戒目标必须是其他角色"},
					},
					Checks: []types.TargetCheckRule{
						{Kind: types.TargetCheckTargetMinHeal, Index: 0, Min: 1, Err: "惩戒目标至少需要有1点治疗"},
					},
				},
			},
		},
		{ID: "holy_lancer_holy_strike", Handler: &HolyLancerHolyStrikeHandler{}},
		{ID: "holy_lancer_sky_spear", Handler: &HolyLancerSkySpearHandler{}},
		{ID: "holy_lancer_earth_spear", Handler: &HolyLancerEarthSpearHandler{}},
		{ID: "holy_lancer_prayer", Handler: &HolyLancerPrayerHandler{}},
	}
}

// ChoiceRouteSpecs 导出角色 choice 路由声明。
func ChoiceRouteSpecs() map[string]types.ChoiceRouteSpec {
	return map[string]types.ChoiceRouteSpec{
		"holy_lancer_earth_spear_x":   types.ChoiceRouteRole("holy_lancer"),
		"holy_lancer_revelation_pick": types.ChoiceRouteRole("holy_lancer"),
	}
}
