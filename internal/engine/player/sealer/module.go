// gameflow: 封印师模块入口声明。

package sealer

import (
	"starcup-engine/internal/engine/player"
	skills "starcup-engine/internal/engine/skill"
	"starcup-engine/internal/types"
)

// SkillEntries 导出角色技能与策略绑定入口。
func SkillEntries() []player.SkillEntry {
	return []player.SkillEntry{
		{ID: "magic_surge", Handler: &skills.MagicSurgeHandler{}},
		{
			ID:      "seal_break",
			Handler: &skills.SealBreakHandler{},
			Policy: types.SkillPolicy{
				TargetRules: types.TargetRuleSet{
					Count: types.TargetCountRule{Min: 0, Max: 1},
					Checks: []types.TargetCheckRule{
						{Kind: types.TargetCheckAnyBasicFieldWhenNone, Err: "场上没有可收回的基础效果"},
						{Kind: types.TargetCheckHasBasicFieldOnTarget, Index: 0, Err: "%s 面前没有可收回的基础效果", WithTargetName: true},
					},
				},
			},
		},
		{ID: "five_elements_bind", Handler: &skills.FiveElementsBindHandler{}},
		{ID: "water_seal", Handler: skills.NewWaterSealHandler()},
		{ID: "fire_seal", Handler: skills.NewFireSealHandler()},
		{ID: "earth_seal", Handler: skills.NewEarthSealHandler()},
		{ID: "wind_seal", Handler: skills.NewWindSealHandler()},
		{ID: "thunder_seal", Handler: skills.NewThunderSealHandler()},
	}
}

// ChoiceRouteSpecs 导出角色 choice 路由声明。
func ChoiceRouteSpecs() map[string]types.ChoiceRouteSpec {
	return map[string]types.ChoiceRouteSpec{
		"sealer_five_elements_bind_pick": types.ChoiceRouteRole("sealer"),
	}
}
