// gameflow: 天使模块入口声明。

package angel

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/types"
)

// RoleEntry 导出角色统一入口定义。
func RoleEntry() player.RoleEntry {
	return player.RoleEntry{
		ID:               "angel",
		Choices:          NewChoiceHandler(),
		Skills:           SkillEntries(),
		ChoiceRouteSpecs: ChoiceRouteSpecs(),
	}
}

// SkillEntries 导出角色技能与策略绑定入口。
func SkillEntries() []player.SkillEntry {
	return []player.SkillEntry{
		{ID: "angel_bond", Handler: &AngelBondHandler{}},
		{
			ID:      "angel_blessing",
			Handler: &AngelBlessingHandler{},
			Policy: types.SkillPolicy{
				TargetRules: types.TargetRuleSet{
					Count:       types.TargetCountRule{Min: 1, Max: 2, Err: "天使祝福只能指定 1 名或 2 名目标"},
					Distinct:    true,
					DistinctErr: "天使祝福指定 2 名目标时不能重复选择同一角色",
				},
			},
		},
		{
			ID:      "angel_cleanse",
			Handler: &AngelCleanseHandler{},
			Policy: types.SkillPolicy{
				TargetRules: types.TargetRuleSet{
					Count: types.TargetCountRule{Min: 0, Max: 1, Err: "风之洁净最多指定1名目标"},
				},
			},
		},
		{ID: "angel_song", Handler: &AngelSongHandler{}},
		{ID: "god_protection", Handler: &GodProtectionHandler{}},
		{ID: "angel_wall", Handler: &AngelWallHandler{}},
	}
}

// ChoiceRouteSpecs 导出角色 choice 路由声明。
func ChoiceRouteSpecs() map[string]types.ChoiceRouteSpec {
	return map[string]types.ChoiceRouteSpec{
		"angel_bond_heal_target": types.ChoiceRouteRole("angel"),
		"god_protection_x":       types.ChoiceRouteRole("angel"),
	}
}
