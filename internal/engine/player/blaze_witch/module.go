// gameflow: 烈焰魔女模块入口声明。

package blaze_witch

import (
	"starcup-engine/internal/engine/player"
	skills "starcup-engine/internal/engine/skill"
	"starcup-engine/internal/model"
	"starcup-engine/internal/types"
)

// RoleEntry 导出角色统一入口定义。
func RoleEntry() player.RoleEntry {
	return player.RoleEntry{
		ID:               "blaze_witch",
		Defaults:         ApplyDefaults,
		Choices:          NewChoiceHandler(),
		Skills:           SkillEntries(),
		ChoiceRouteSpecs: ChoiceRouteSpecs(),
		TimingHookSpecs: []player.TimingHookSpec{
			{Timing: player.TimingPostDamageResolved, Priority: 200, Hook: postDamageResolvedHook},
		},
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
	p.Tokens["bw_flame_release_pending"] = 0
}

// SkillEntries 导出角色技能与策略绑定入口。
func SkillEntries() []player.SkillEntry {
	return []player.SkillEntry{
		{ID: "bw_rebirth_clock", Handler: &skills.BlazeWitchRebirthClockHandler{}},
		{
			ID:      "bw_blazing_codex",
			Handler: &skills.BlazeWitchBlazingCodexHandler{},
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
			Handler: &skills.BlazeWitchHeavenfireCleaveHandler{},
			Policy: types.SkillPolicy{
				TargetRules: types.TargetRuleSet{
					Count: types.TargetCountRule{Min: 1, Max: 1, Err: "天火断空需要且仅能指定1名其他角色"},
					Slots: []types.TargetSlotRule{
						{Index: 0, Self: types.TargetSelfOther, Err: "天火断空不能以自己为目标"},
					},
				},
			},
		},
		{ID: "bw_witch_wrath", Handler: &skills.BlazeWitchWitchWrathHandler{}},
		{ID: "bw_substitute_doll", Handler: &skills.BlazeWitchSubstituteDollHandler{}},
		{ID: "bw_pain_link", Handler: &skills.BlazeWitchPainLinkHandler{}},
		{ID: "bw_mana_inversion", Handler: &skills.BlazeWitchManaInversionHandler{}},
	}
}

// ChoiceRouteSpecs 导出角色 choice 路由声明。
func ChoiceRouteSpecs() map[string]types.ChoiceRouteSpec {
	return map[string]types.ChoiceRouteSpec{
		"bw_mana_inversion_cards":   types.ChoiceRouteRole("blaze_witch"),
		"bw_mana_inversion_target":  types.ChoiceRouteRole("blaze_witch"),
		"bw_mana_inversion_x":       types.ChoiceRouteRole("blaze_witch"),
		"bw_pain_link_pick":         types.ChoiceRouteRole("blaze_witch"),
		"bw_substitute_doll_card":   types.ChoiceRouteRole("blaze_witch"),
		"bw_substitute_doll_target": types.ChoiceRouteRole("blaze_witch"),
		"bw_witch_wrath_draw":       types.ChoiceRouteRole("blaze_witch"),
	}
}
