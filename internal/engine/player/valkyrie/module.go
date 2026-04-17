// gameflow: 女武神模块入口声明。

package valkyrie

import (
	"starcup-engine/internal/engine/player"
	skills "starcup-engine/internal/engine/skill"
	"starcup-engine/internal/types"
)

// SkillEntries 导出角色技能与策略绑定入口。
func SkillEntries() []player.SkillEntry {
	return []player.SkillEntry{
		{ID: "valkyrie_divine_pursuit", Handler: &skills.ValkyrieDivinePursuitHandler{}},
		{ID: "valkyrie_order_seal", Handler: &skills.ValkyrieOrderSealHandler{}},
		{ID: "valkyrie_peace_walker", Handler: &skills.ValkyriePeaceWalkerHandler{}},
		{ID: "valkyrie_military_glory", Handler: &skills.ValkyrieMilitaryGloryHandler{}},
		{ID: "valkyrie_heroic_summon", Handler: &skills.ValkyrieHeroicSummonHandler{}},
	}
}

// ChoiceRouteSpecs 导出角色 choice 路由声明。
func ChoiceRouteSpecs() map[string]types.ChoiceRouteSpec {
	return map[string]types.ChoiceRouteSpec{
		"valkyrie_heroic_discard_card":   types.ChoiceRouteRole("valkyrie"),
		"valkyrie_heroic_summon_pick":    types.ChoiceRouteRole("valkyrie"),
		"valkyrie_military_glory_mode":   types.ChoiceRouteRole("valkyrie"),
		"valkyrie_military_glory_target": types.ChoiceRouteRole("valkyrie"),
		"valkyrie_military_glory_x":      types.ChoiceRouteRole("valkyrie"),
	}
}
