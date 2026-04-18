// gameflow: 英灵人形模块入口声明。

package war_homunculus

import (
	"starcup-engine/internal/engine/player"
	skills "starcup-engine/internal/engine/skill"
	"starcup-engine/internal/model"
	"starcup-engine/internal/types"
)

// RoleEntry 导出角色统一入口定义。
func RoleEntry() player.RoleEntry {
	return player.RoleEntry{
		ID:               "war_homunculus",
		Defaults:         ApplyDefaults,
		Choices:          NewChoiceHandler(),
		Skills:           SkillEntries(),
		ChoiceRouteSpecs: ChoiceRouteSpecs(),
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
		{ID: "hom_battle_pattern", Handler: &skills.HomunculusBattlePatternHandler{}},
		{ID: "hom_rage_suppress", Handler: &skills.HomunculusRageSuppressHandler{}},
		{ID: "hom_rune_smash", Handler: &skills.HomunculusRuneSmashHandler{}},
		{ID: "hom_glyph_fusion", Handler: &skills.HomunculusGlyphFusionHandler{}},
		{ID: "hom_rune_reforge", Handler: &skills.HomunculusRuneReforgeHandler{}},
		{ID: "hom_dual_echo", Handler: &skills.HomunculusDualEchoHandler{}},
	}
}

// ChoiceRouteSpecs 导出角色 choice 路由声明。
func ChoiceRouteSpecs() map[string]types.ChoiceRouteSpec {
	return map[string]types.ChoiceRouteSpec{
		"hom_dual_echo_target":          types.ChoiceRouteRole("war_homunculus"),
		"hom_glyph_fusion_cards":        types.ChoiceRouteRole("war_homunculus"),
		"hom_glyph_fusion_pick":         types.ChoiceRouteRole("war_homunculus"),
		"hom_glyph_fusion_x":            types.ChoiceRouteRole("war_homunculus"),
		"hom_glyph_fusion_y":            types.ChoiceRouteRole("war_homunculus"),
		"hom_rune_reforge_distribution": types.ChoiceRouteRole("war_homunculus"),
		"hom_rune_reforge_pick":         types.ChoiceRouteRole("war_homunculus"),
		"hom_rune_smash_cards":          types.ChoiceRouteRole("war_homunculus"),
		"hom_rune_smash_x":              types.ChoiceRouteRole("war_homunculus"),
		"hom_rune_smash_y":              types.ChoiceRouteRole("war_homunculus"),
	}
}
