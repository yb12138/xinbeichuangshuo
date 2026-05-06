// gameflow: 贤者模块入口声明。

package sage

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
	"starcup-engine/internal/types"
)

// RoleEntry 导出角色统一入口定义。
func RoleEntry() player.RoleEntry {
	return player.RoleEntry{
		ID: "sage",
		EnergyCapRule: player.EnergyCapRuleFuncs{
			Modifier: func(_ *model.Player, current int) int { return current + 1 },
		},
		Choices:          NewChoiceHandler(),
		ChoiceSpecs:      ChoiceSpecs(),
		Skills:           SkillEntries(),
		ChoiceRouteSpecs: ChoiceRouteSpecs(),
		TimingHookSpecs: []player.TimingHookSpec{
			{Timing: player.TimingPostDamageResolved, Priority: 400, Hook: postDamageResolvedHook},
		},
	}
}

// ChoiceSpecs 导出角色 choice 声明。
func ChoiceSpecs() []player.ChoiceSpec {
	return []player.ChoiceSpec{
		{ChoiceType: "sage_arcane_cards", SequentialRemaining: player.ChoiceRemainingFromSelectionKey("x_value")},
		{ChoiceType: "sage_holy_cards", SequentialRemaining: player.ChoiceRemainingFromSelectionKey("x_value")},
		{ChoiceType: "sage_magic_rebound_cards", SequentialRemaining: player.ChoiceRemainingFromSelectionKey("x_value")},
	}
}

// SkillEntries 导出角色技能与策略绑定入口。
func SkillEntries() []player.SkillEntry {
	return []player.SkillEntry{
		{ID: "sage_wisdom_codex", Handler: &SageWisdomCodexHandler{}},
		{ID: "sage_magic_rebound", Handler: &SageMagicReboundHandler{}},
		{
			ID:      "sage_arcane_codex",
			Handler: &SageArcaneCodexHandler{},
			Policy: types.SkillPolicy{
				TargetRules: types.TargetRuleSet{
					Count: types.TargetCountRule{Min: 0, Max: 1, Err: "魔道法典需要且仅能指定1名其他角色"},
					Slots: []types.TargetSlotRule{
						{Index: 0, Self: types.TargetSelfOther, Err: "魔道法典不能以自己为目标"},
					},
				},
			},
		},
		{
			ID:      "sage_holy_codex",
			Handler: &SageHolyCodexHandler{},
			Policy: types.SkillPolicy{
				TargetRules: types.TargetRuleSet{
					Count: types.TargetCountRule{Min: 0, Max: 6},
				},
			},
		},
	}
}

// ChoiceRouteSpecs 导出角色 choice 路由声明。
func ChoiceRouteSpecs() map[string]types.ChoiceRouteSpec {
	return map[string]types.ChoiceRouteSpec{
		"sage_magic_rebound_confirm": types.ChoiceRouteRole("sage"),
		"sage_magic_rebound_x":       types.ChoiceRouteRole("sage"),
		"sage_magic_rebound_element": types.ChoiceRouteRole("sage"),
		"sage_magic_rebound_cards":   types.ChoiceRouteRole("sage"),
		"sage_magic_rebound_target":  types.ChoiceRouteRole("sage"),
		"sage_arcane_cards":          types.ChoiceRouteRole("sage"),
		"sage_arcane_x":              types.ChoiceRouteRole("sage"),
		"sage_arcane_target":         types.ChoiceRouteRole("sage"),
		"sage_holy_cards":            types.ChoiceRouteRole("sage"),
		"sage_holy_x":                types.ChoiceRouteRole("sage"),
		"sage_holy_target_count":     types.ChoiceRouteRole("sage"),
		"sage_holy_targets":          types.ChoiceRouteRole("sage"),
	}
}
