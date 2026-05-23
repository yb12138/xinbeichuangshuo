// gameflow: 英灵人形模块入口声明。

package war_homunculus

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
	"starcup-engine/internal/types"
)

// RoleEntry 导出角色统一入口定义。
func RoleEntry() player.RoleEntry {
	return player.RoleEntry{
		ID:       "war_homunculus",
		Defaults: ApplyDefaults,
		HandLimit: player.HandLimitRuleFuncs{
			Modifier: func(p *model.Player, current int) int {
				if player.HasForm(p, model.FormWarHomunculusBurst) {
					return current + 1
				}
				return current
			},
		},
		Choices:          NewChoiceHandler(),
		ChoiceSpecs:      ChoiceSpecs(),
		Skills:           SkillEntries(),
		ChoiceRouteSpecs: ChoiceRouteSpecs(),
		TimingHookSpecs: []player.TimingHookSpec{
			{Timing: player.TimingTurnEndFinal, Priority: 700, Hook: turnEndHook},
		},
	}
}

// ChoiceSpecs 导出角色 choice 声明。
func ChoiceSpecs() []player.ChoiceSpec {
	return []player.ChoiceSpec{
		{
			ChoiceType:        "hom_glyph_fusion_cards",
			HandleMultiSelect: handleRuneCardsMultiSelect(true),
		},
		{
			ChoiceType:        "hom_rune_smash_cards",
			HandleMultiSelect: handleRuneCardsMultiSelect(false),
		},
		{
			ChoiceType:        "hom_rune_reforge_distribution",
			HandleMultiSelect: handleRuneReforgeAllocate,
		},
		{
			ChoiceType:        "hom_rune_reforge_allocate",
			HandleMultiSelect: handleRuneReforgeAllocate,
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
	p.Tokens["hom_war_rune"] = 3
	p.Tokens["hom_magic_rune"] = 0
}

// SkillEntries 导出角色技能与策略绑定入口。
func SkillEntries() []player.SkillEntry {
	return []player.SkillEntry{
		{ID: "hom_battle_pattern", Handler: &HomunculusBattlePatternHandler{}},
		{
			ID:      "hom_rage_suppress",
			Handler: &HomunculusRageSuppressHandler{},
			Policy: types.SkillPolicy{
				ExclusiveResponseGroup: "war_homunculus_rage_response",
			},
		},
		{ID: "hom_rune_smash", Handler: &HomunculusRuneSmashHandler{}},
		{
			ID:      "hom_glyph_fusion",
			Handler: &HomunculusGlyphFusionHandler{},
			Policy: types.SkillPolicy{
				ExclusiveResponseGroup: "war_homunculus_rage_response",
			},
		},
		{ID: "hom_rune_reforge", Handler: &HomunculusRuneReforgeHandler{}},
		{ID: "hom_dual_echo", Handler: &HomunculusDualEchoHandler{}},
	}
}

// ChoiceRouteSpecs 导出角色 choice 路由声明。
func ChoiceRouteSpecs() map[string]types.ChoiceRouteSpec {
	return map[string]types.ChoiceRouteSpec{
		"hom_dual_echo_target":          types.ChoiceRouteRoleWithPhaseSync("war_homunculus", string(player.InterruptPhaseSyncDamageResolution)),
		"hom_glyph_fusion_cards":        types.ChoiceRouteRole("war_homunculus"),
		"hom_glyph_fusion_pick":         types.ChoiceRouteRole("war_homunculus"),
		"hom_glyph_fusion_x":            types.ChoiceRouteRole("war_homunculus"),
		"hom_glyph_fusion_y":            types.ChoiceRouteRole("war_homunculus"),
		"hom_rune_reforge_allocate":     types.ChoiceRouteRole("war_homunculus"),
		"hom_rune_reforge_distribution": types.ChoiceRouteRole("war_homunculus"),
		"hom_rune_reforge_pick":         types.ChoiceRouteRole("war_homunculus"),
		"hom_rune_smash_cards":          types.ChoiceRouteRole("war_homunculus"),
		"hom_rune_smash_x":              types.ChoiceRouteRole("war_homunculus"),
		"hom_rune_smash_y":              types.ChoiceRouteRole("war_homunculus"),
	}
}
