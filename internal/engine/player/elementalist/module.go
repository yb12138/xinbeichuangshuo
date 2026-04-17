// gameflow: 元素师模块入口声明。

package elementalist

import (
	"starcup-engine/internal/engine/player"
	skills "starcup-engine/internal/engine/skill"
	"starcup-engine/internal/types"
)

// SkillEntries 导出角色技能与策略绑定入口。
func SkillEntries() []player.SkillEntry {
	return []player.SkillEntry{
		{ID: "elementalist_absorb", Handler: &skills.ElementalistAbsorbHandler{}},
		{ID: "elementalist_ignite", Handler: &skills.ElementalistIgniteHandler{}},
		{ID: "elementalist_thunder_strike", Handler: &skills.ElementalistThunderStrikeHandler{}},
		{
			ID:      "elementalist_freeze",
			Handler: &skills.ElementalistFreezeHandler{},
			Policy: types.SkillPolicy{
				TargetRules: types.TargetRuleSet{
					Count: types.TargetCountRule{Min: 2, Max: 2, Err: "冰冻需要指定2名目标"},
				},
			},
		},
		{ID: "elementalist_wind_blade", Handler: &skills.ElementalistWindBladeHandler{}},
		{ID: "elementalist_meteor", Handler: &skills.ElementalistMeteorHandler{}},
		{ID: "elementalist_fireball", Handler: &skills.ElementalistFireballHandler{}},
		{ID: "elementalist_moonlight", Handler: &skills.ElementalistMoonlightHandler{}},
	}
}

// ChoiceRouteSpecs 导出角色 choice 路由声明。
func ChoiceRouteSpecs() map[string]types.ChoiceRouteSpec {
	return map[string]types.ChoiceRouteSpec{
		"elementalist_bonus_card":      types.ChoiceRouteRole("elementalist"),
		"elementalist_primordial_pick": types.ChoiceRouteRole("elementalist"),
	}
}
