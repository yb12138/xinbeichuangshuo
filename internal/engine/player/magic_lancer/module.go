// gameflow: 魔枪模块入口声明。

package magic_lancer

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
	"starcup-engine/internal/types"
)

// RoleEntry 导出角色统一入口定义。
func RoleEntry() player.RoleEntry {
	return player.RoleEntry{
		ID:       "magic_lancer",
		Defaults: ApplyDefaults,
		HandLimit: player.HandLimitRuleFuncs{
			Hard: func(p *model.Player) (int, bool) {
				if player.HasForm(p, model.FormMagicLancerPhantom) {
					return 5, true
				}
				return 0, false
			},
		},
		Choices:          NewChoiceHandler(),
		ChoiceSpecs:      ChoiceSpecs(),
		Skills:           SkillEntries(),
		ChoiceRouteSpecs: ChoiceRouteSpecs(),
		BlocksActionType: func(p *model.Player, at model.ActionType) bool {
			return at == model.ActionMagic && BlocksMagicCasting(p)
		},
		FlowContinuationHandlers: map[model.FlowContinuationKind]player.FlowContinuationHandler{
			model.FlowContinuationAfterDiscard: handleMagicLancerAfterDiscard,
		},
		TimingHookSpecs: []player.TimingHookSpec{
			{Timing: player.TimingDamageSourceDeal, Priority: 300, Hook: damageCalculateHook},
			{Timing: player.TimingPostDamageResolved, Priority: 300, Hook: postDamageResolvedHook},
			{Timing: player.TimingDefendValidation, Priority: 100, Hook: defendValidationHook},
			{Timing: player.TimingMagicMissileDefend, Priority: 100, Hook: magicMissileDefendHook},
			{Timing: player.TimingMagicMissileCounter, Priority: 100, Hook: magicMissileCounterHook},
		},
	}
}

// ChoiceSpecs 导出角色 choice 声明。
func ChoiceSpecs() []player.ChoiceSpec {
	return []player.ChoiceSpec{
		{ChoiceType: "ml_dark_barrier_cards", HandleMultiSelect: handleDarkBarrierCardsMultiSelect},
	}
}

// ApplyDefaults 初始化角色默认属性。
func ApplyDefaults(p *model.Player) {
	// magic_lancer 无特殊默认配置
}

// SkillEntries 导出角色技能与策略绑定入口。
func SkillEntries() []player.SkillEntry {
	return []player.SkillEntry{
		{ID: "ml_dark_release", Handler: &MagicLancerDarkReleaseHandler{}},
		{
			ID:      "ml_phantom_stardust",
			Handler: &MagicLancerPhantomStardustHandler{},
			Policy: types.SkillPolicy{
				TargetRules: types.TargetRuleSet{
					Count: types.TargetCountRule{Min: 0, Max: 1, Err: "幻影星尘需要且仅能指定1名敌方角色"},
					Slots: []types.TargetSlotRule{
						{Index: 0, Camp: types.TargetCampEnemy, Err: "幻影星尘目标必须是敌方角色"},
					},
				},
			},
		},
		{ID: "ml_dark_bind", Handler: &MagicLancerDarkBindHandler{}},
		{ID: "ml_dark_barrier", Handler: &MagicLancerDarkBarrierHandler{}},
		{
			ID:      "ml_fullness",
			Handler: &MagicLancerFullnessHandler{},
			Policy: types.SkillPolicy{
				TargetRules: types.TargetRuleSet{
					Count: types.TargetCountRule{Min: 0, Max: 1, Err: "充盈至多指定1名其他队友"},
					Slots: []types.TargetSlotRule{
						{Index: 0, Camp: types.TargetCampAlly, Self: types.TargetSelfOther, Err: "充盈的可选目标必须是其他队友"},
					},
				},
			},
		},
		{ID: "ml_black_spear", Handler: &MagicLancerBlackSpearHandler{}},
	}
}

// ChoiceRouteSpecs 导出角色 choice 路由声明。
func ChoiceRouteSpecs() map[string]types.ChoiceRouteSpec {
	return map[string]types.ChoiceRouteSpec{
		"ml_black_spear_x":         types.ChoiceRouteRole("magic_lancer"),
		"ml_dark_barrier_cards":    types.ChoiceRouteRole("magic_lancer"),
		"ml_dark_barrier_mode":     types.ChoiceRouteRole("magic_lancer"),
		"ml_dark_release_pick":     types.ChoiceRouteRole("magic_lancer"),
		"ml_fullness_cost_card":    types.ChoiceRouteRole("magic_lancer"),
		"ml_fullness_discard_step": types.ChoiceRouteRole("magic_lancer"),
		"ml_phantom_stardust_pick": types.ChoiceRouteRole("magic_lancer"),
		"ml_stardust_target":       types.ChoiceRouteRole("magic_lancer"),
	}
}
