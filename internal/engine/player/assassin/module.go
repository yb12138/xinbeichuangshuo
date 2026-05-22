// gameflow: 暗杀者模块入口声明。

package assassin

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
	"starcup-engine/internal/types"
)

// RoleEntry 导出角色统一入口定义。
func RoleEntry() player.RoleEntry {
	return player.RoleEntry{
		ID: "assassin",
		HandLimit: player.HandLimitRuleFuncs{
			Modifier: func(p *model.Player, current int) int {
				if player.HasForm(p, model.FormAssassinStealth) {
					return current - 1
				}
				return current
			},
		},
		TargetFilter: player.TargetFilterRuleFuncs{
			CannotBeTarget: func(p *model.Player) bool {
				return player.HasForm(p, model.FormAssassinStealth)
			},
		},
		FlowContinuationHandlers: map[model.FlowContinuationKind]player.FlowContinuationHandler{
			model.FlowContinuationAfterDraw: handleStealthAfterDraw,
		},
		Choices:          NewChoiceHandler(),
		Skills:           SkillEntries(),
		ChoiceRouteSpecs: ChoiceRouteSpecs(),
		TimingHookSpecs: []player.TimingHookSpec{
			{Timing: player.TimingBeforeAction, Priority: 200, Hook: beforeActionStealthReleaseHook},
			{Timing: player.TimingDamageSourceDeal, Priority: 600, Hook: damageCalculateHook},
			{Timing: player.TimingAttackNoResponse, Priority: 200, Hook: attackGatingHook},
		},
	}
}

// SkillEntries 导出角色技能与策略绑定入口。
func SkillEntries() []player.SkillEntry {
	return []player.SkillEntry{
		{ID: "backlash", Handler: &BacklashHandler{}},
		{ID: "water_shadow", Handler: &WaterShadowHandler{}},
		{ID: "stealth", Handler: &StealthHandler{}},
	}
}

// ChoiceRouteSpecs 导出角色 choice 路由声明。
func ChoiceRouteSpecs() map[string]types.ChoiceRouteSpec {
	return map[string]types.ChoiceRouteSpec{
		"assassin_stealth_draw": types.ChoiceRouteRole("assassin"),
	}
}
