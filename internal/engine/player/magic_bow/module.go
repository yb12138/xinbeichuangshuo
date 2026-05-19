// gameflow: 魔弓模块入口声明。

package magic_bow

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
	"starcup-engine/internal/types"
)

// RoleEntry 导出角色统一入口定义。
func RoleEntry() player.RoleEntry {
	return player.RoleEntry{
		ID:               "magic_bow",
		Choices:          NewChoiceHandler(),
		ChoiceSpecs:      ChoiceSpecs(),
		Skills:           SkillEntries(),
		ChoiceRouteSpecs: ChoiceRouteSpecs(),
		FlowContinuationHandlers: map[model.FlowContinuationKind]player.FlowContinuationHandler{
			model.FlowContinuationAfterDiscard: handleMagicBowAfterDiscard,
		},
		TimingHookSpecs: []player.TimingHookSpec{
			{Timing: player.TimingOnAttackMiss, Priority: 300, Hook: attackMissHook},
			{Timing: player.TimingOnAttackTargetCtx, Priority: 100, Hook: attackTargetCtxHook},
			{Timing: player.TimingPostAttackHit, Priority: 600, Hook: postAttackHitHook},
		},
		SkillUsabilityCheckers: map[string]player.SkillUsabilityChecker{
			"mb_thunder_scatter": CheckThunderScatterUsability,
		},
	}
}

// ChoiceSpecs 导出角色 choice 声明。
func ChoiceSpecs() []player.ChoiceSpec {
	return []player.ChoiceSpec{
		{
			ChoiceType:        "mb_charge_place_cards",
			HandleMultiSelect: handleChargePlaceCardsMultiSelect,
		},
		{
			ChoiceType:          "mb_demon_eye_charge_card",
			SequentialRemaining: player.ChoiceRemainingFromFlowSelectionCount("need", "cards"),
		},
	}
}

// SkillEntries 导出角色技能与策略绑定入口。
func SkillEntries() []player.SkillEntry {
	return []player.SkillEntry{
		{ID: "mb_magic_pierce", Handler: &MagicBowMagicPierceHandler{}},
		{
			ID:      "mb_thunder_scatter",
			Handler: &MagicBowThunderScatterHandler{},
			Policy: types.SkillPolicy{
				TargetRules: types.TargetRuleSet{
					Count: types.TargetCountRule{Min: 0, Max: 1, Err: "雷光散射至多指定1名敌方角色作为额外目标"},
					Slots: []types.TargetSlotRule{
						{Index: 0, Camp: types.TargetCampEnemy, Err: "雷光散射的额外目标必须是敌方角色"},
					},
				},
			},
		},
		{ID: "mb_multi_shot", Handler: &MagicBowMultiShotHandler{}},
		{ID: "mb_charge", Handler: &MagicBowChargeHandler{}},
		{
			ID:      "mb_demon_eye",
			Handler: &MagicBowDemonEyeHandler{},
			Policy: types.SkillPolicy{
				TargetRules: types.TargetRuleSet{
					Count: types.TargetCountRule{Min: 0, Max: 0},
				},
			},
		},
	}
}

// ChoiceRouteSpecs 导出角色 choice 路由声明。
func ChoiceRouteSpecs() map[string]types.ChoiceRouteSpec {
	return map[string]types.ChoiceRouteSpec{
		"mb_charge_discard_pick":         types.ChoiceRouteRole("magic_bow"),
		"mb_charge_draw_x":               types.ChoiceRouteRole("magic_bow"),
		"mb_charge_place_cards":          types.ChoiceRouteRole("magic_bow"),
		"mb_charge_place_count":          types.ChoiceRouteRole("magic_bow"),
		"mb_demon_eye_charge_card":       types.ChoiceRouteRole("magic_bow"),
		"mb_demon_eye_mode":              types.ChoiceRouteRole("magic_bow"),
		"mb_demon_eye_pick":              types.ChoiceRouteRole("magic_bow"),
		"mb_demon_eye_target":            types.ChoiceRouteRole("magic_bow"),
		"mb_magic_pierce_charge":         types.ChoiceRouteRole("magic_bow"),
		"mb_magic_pierce_hit_bonus":      types.ChoiceRouteRole("magic_bow"),
		"mb_magic_pierce_hit_charge":     types.ChoiceRouteRole("magic_bow"),
		"mb_multi_shot_charge":           types.ChoiceRouteRole("magic_bow"),
		"mb_multi_shot_target":           types.ChoiceRouteRole("magic_bow"),
		"mb_thunder_scatter_base_charge": types.ChoiceRouteRole("magic_bow"),
		"mb_thunder_scatter_extra":       types.ChoiceRouteRole("magic_bow"),
		"mb_thunder_scatter_target":      types.ChoiceRouteRole("magic_bow"),
	}
}
