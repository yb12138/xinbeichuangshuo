// gameflow: 女武神模块入口声明。

package valkyrie

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/types"
)

// RoleEntry 导出角色统一入口定义。
func RoleEntry() player.RoleEntry {
	return player.RoleEntry{
		ID:               "valkyrie",
		Choices:          NewChoiceHandler(),
		Skills:           SkillEntries(),
		ChoiceRouteSpecs: ChoiceRouteSpecs(),
		TimingHookSpecs: []player.TimingHookSpec{
			{Timing: player.TimingOnTurnStart, Priority: 100, Hook: turnStartMilitaryGloryHook},
		},
	}
}

// SkillEntries 导出角色技能与策略绑定入口。
func SkillEntries() []player.SkillEntry {
	return []player.SkillEntry{
		{ID: "valkyrie_divine_pursuit", Handler: &ValkyrieDivinePursuitHandler{}},
		{ID: "valkyrie_order_seal", Handler: &ValkyrieOrderSealHandler{}},
		{ID: "valkyrie_peace_walker", Handler: &ValkyriePeaceWalkerHandler{}},
		{ID: "valkyrie_military_glory", Handler: &ValkyrieMilitaryGloryHandler{}},
		{ID: "valkyrie_heroic_summon", Handler: &ValkyrieHeroicSummonHandler{}},
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
