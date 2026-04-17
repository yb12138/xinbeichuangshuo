// gameflow: 英灵人形模块入口声明。

package war_homunculus

import (
	"starcup-engine/internal/engine/player"
	skills "starcup-engine/internal/engine/skill"
	"starcup-engine/internal/types"
)

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
		"hom_glyph_fusion_cards":        types.ChoiceRouteRole("homunculus"),
		"hom_glyph_fusion_pick":         types.ChoiceRouteRole("homunculus"),
		"hom_glyph_fusion_x":            types.ChoiceRouteRole("homunculus"),
		"hom_glyph_fusion_y":            types.ChoiceRouteRole("homunculus"),
		"hom_rune_reforge_distribution": types.ChoiceRouteRole("homunculus"),
		"hom_rune_reforge_pick":         types.ChoiceRouteRole("homunculus"),
		"hom_rune_smash_cards":          types.ChoiceRouteRole("homunculus"),
		"hom_rune_smash_x":              types.ChoiceRouteRole("homunculus"),
		"hom_rune_smash_y":              types.ChoiceRouteRole("homunculus"),
	}
}
