// gameflow: 魔剑士模块入口声明。

package magic_swordsman

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
	"starcup-engine/internal/types"
)

// RoleEntry 导出角色统一入口定义。
func RoleEntry() player.RoleEntry {
	return player.RoleEntry{
		ID:               "magic_swordsman",
		Choices:          NewChoiceHandler(),
		Skills:           SkillEntries(),
		ChoiceRouteSpecs: ChoiceRouteSpecs(),
		CannotActChecker: CannotActCheckerFn,
		TimingHookSpecs: []player.TimingHookSpec{
			{Timing: player.TimingOnDamageCalculate, Priority: 200, Hook: damageCalculateHook},
			{Timing: player.TimingOnAttackStateReset, Priority: 100, Hook: attackStateResetHook},
			{Timing: player.TimingOnAttackGating, Priority: 200, Hook: attackGatingHook},
			{Timing: player.TimingPostAttackHit, Priority: 500, Hook: postAttackHitHook},
			{Timing: player.TimingBeforeAction, Priority: 100, Hook: beforeActionShadowReleaseHook},
			{Timing: player.TimingOnCombatInteraction, Priority: 400, Hook: darkElementResponseHook},
			{Timing: player.TimingOnCombatCounterCard, Priority: 100, Hook: shadowRejectMagicBulletHook},
			{Timing: player.TimingAfterCannotAct, Priority: 100, Hook: cannotActRedrawHook},
		},
		SkillUsabilityCheckers: map[string]player.SkillUsabilityChecker{
			"ms_shadow_meteor": CheckShadowMeteorUsability,
		},
		BlocksActionType: func(p *model.Player, at model.ActionType) bool {
			return at == model.ActionMagic && BlocksMagicCasting(p)
		},
	}
}

// SkillEntries 导出角色技能与策略绑定入口。
func SkillEntries() []player.SkillEntry {
	return []player.SkillEntry{
		{ID: "ms_asura_combo", Handler: &MagicSwordsmanAsuraComboHandler{}},
		{ID: "ms_shadow_gather", Handler: &MagicSwordsmanShadowGatherHandler{}},
		{ID: "ms_shadow_power", Handler: &MagicSwordsmanShadowPowerHandler{}},
		{ID: "ms_shadow_reject", Handler: &MagicSwordsmanShadowRejectHandler{}},
		{ID: "ms_shadow_meteor", Handler: &MagicSwordsmanShadowMeteorHandler{}},
		{ID: "ms_yellow_spring", Handler: &MagicSwordsmanYellowSpringHandler{}},
	}
}

// ChoiceRouteSpecs 导出角色 choice 路由声明。
func ChoiceRouteSpecs() map[string]types.ChoiceRouteSpec {
	return map[string]types.ChoiceRouteSpec{
		"ms_asura_combo_pick":              types.ChoiceRouteRole("magic_swordsman"),
		"ms_shadow_gather_pick":            types.ChoiceRouteRole("magic_swordsman"),
		"ms_shadow_meteor_release_confirm": types.ChoiceRouteRole("magic_swordsman"),
	}
}
