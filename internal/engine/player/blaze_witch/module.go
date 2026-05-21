// gameflow: 烈焰魔女模块入口声明。

package blaze_witch

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
	"starcup-engine/internal/types"
)

// RoleEntry 导出角色统一入口定义。
func RoleEntry() player.RoleEntry {
	return player.RoleEntry{
		ID:       "blaze_witch",
		Defaults: ApplyDefaults,
		HandLimit: player.HandLimitRuleFuncs{
			Modifier: func(p *model.Player, current int) int {
				if player.HasForm(p, model.FormBlazeWitchFlame) {
					return current + p.Tokens["bw_rebirth"] - 2
				}
				return current
			},
		},
		Choices:          NewChoiceHandler(),
		ChoiceSpecs:      ChoiceSpecs(),
		Skills:           SkillEntries(),
		ChoiceRouteSpecs: ChoiceRouteSpecs(),
		TimingHookSpecs: []player.TimingHookSpec{
			{Timing: player.TimingBeforeAction, Priority: 100, Hook: beforeActionFlameReleaseHook},
			{Timing: player.TimingPostDamageResolved, Priority: 200, Hook: postDamageResolvedHook},
			{Timing: player.TimingOnDamageAfterApply, Priority: 100, Hook: afterApplyHook},
			{Timing: player.TimingOnMoraleLossApplied, Priority: 100, Hook: moraleLossHook},
			{Timing: player.TimingOnAttackCardTransform, Priority: 100, Hook: attackCardTransformHook},
		},
		AttackCardElementTransform: AttackElement,
		AttackElementResolver:      AttackElement,
	}
}

// ChoiceSpecs 导出角色 choice 声明。
func ChoiceSpecs() []player.ChoiceSpec {
	return []player.ChoiceSpec{
		{ChoiceType: "bw_mana_inversion_cards", HandleMultiSelect: handleBlazeWitchManaInversionCardsMultiSelect},
	}
}

// ApplyDefaults 初始化角色默认属性。
func ApplyDefaults(p *model.Player) {
	if p == nil {
		return
	}
	if p.Tokens == nil {
		p.Tokens = map[string]int{}
	}
	p.Tokens["bw_rebirth"] = 0
}

// SkillEntries 导出角色技能与策略绑定入口。
func SkillEntries() []player.SkillEntry {
	return []player.SkillEntry{
		{ID: "bw_rebirth_clock", Handler: &BlazeWitchRebirthClockHandler{}},
		{
			ID:      "bw_blazing_codex",
			Handler: &BlazeWitchBlazingCodexHandler{},
			Policy: types.SkillPolicy{
				TargetRules: types.TargetRuleSet{
					Count: types.TargetCountRule{Min: 1, Max: 1, Err: "苍炎法典需要且仅能指定1名其他角色"},
					Slots: []types.TargetSlotRule{
						{Index: 0, Self: types.TargetSelfOther, Err: "苍炎法典不能以自己为目标"},
					},
				},
			},
		},
		{
			ID:      "bw_heavenfire_cleave",
			Handler: &BlazeWitchHeavenfireCleaveHandler{},
			Policy: types.SkillPolicy{
				TargetRules: types.TargetRuleSet{
					Count: types.TargetCountRule{Min: 1, Max: 1, Err: "天火断空需要且仅能指定1名其他角色"},
					Slots: []types.TargetSlotRule{
						{Index: 0, Self: types.TargetSelfOther, Err: "天火断空不能以自己为目标"},
					},
				},
			},
		},
		{ID: "bw_witch_wrath", Handler: &BlazeWitchWitchWrathHandler{}},
		{ID: "bw_substitute_doll", Handler: &BlazeWitchSubstituteDollHandler{}},
		{ID: "bw_pain_link", Handler: &BlazeWitchPainLinkHandler{}},
		{ID: "bw_mana_inversion", Handler: &BlazeWitchManaInversionHandler{}},
	}
}

// ChoiceRouteSpecs 导出角色 choice 路由声明。
func ChoiceRouteSpecs() map[string]types.ChoiceRouteSpec {
	return map[string]types.ChoiceRouteSpec{
		"bw_mana_inversion_cards":   types.ChoiceRouteRole("blaze_witch"),
		"bw_mana_inversion_target":  types.ChoiceRouteRole("blaze_witch"),
		"bw_pain_link_pick":         types.ChoiceRouteRole("blaze_witch"),
		"bw_substitute_doll_card":   types.ChoiceRouteRole("blaze_witch"),
		"bw_substitute_doll_target": types.ChoiceRouteRole("blaze_witch"),
		"bw_witch_wrath_draw":       types.ChoiceRouteRole("blaze_witch"),
	}
}
