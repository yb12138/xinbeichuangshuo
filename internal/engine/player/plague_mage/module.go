// gameflow: 瘟疫法师模块入口声明。

package plague_mage

import (
	"starcup-engine/internal/engine/player"
	skills "starcup-engine/internal/engine/skill"
	"starcup-engine/internal/types"
)

// SkillEntries 导出角色技能与策略绑定入口。
func SkillEntries() []player.SkillEntry {
	return []player.SkillEntry{
		{ID: "plague_immortal", Handler: &skills.PlagueImmortalHandler{}},
		{ID: "plague_blasphemy", Handler: &skills.PlagueBlasphemyHandler{}},
		{ID: "plague_outbreak", Handler: &skills.PlagueOutbreakHandler{}},
		{
			ID:      "plague_death_touch",
			Handler: &skills.PlagueDeathTouchHandler{},
			Policy: types.SkillPolicy{
				SkipAutoPhaseEnd: true,
			},
		},
		{ID: "plague_toxic_nova", Handler: &skills.PlagueToxicNovaHandler{}},
	}
}

// ChoiceRouteSpecs 导出角色 choice 路由声明。
func ChoiceRouteSpecs() map[string]types.ChoiceRouteSpec {
	return map[string]types.ChoiceRouteSpec{
		"plague_death_touch_cards":    types.ChoiceRouteRole("plague_mage"),
		"plague_death_touch_element":  types.ChoiceRouteRole("plague_mage"),
		"plague_death_touch_target":   types.ChoiceRouteRole("plague_mage"),
		"plague_death_touch_x":        types.ChoiceRouteRole("plague_mage"),
		"plague_death_touch_y":        types.ChoiceRouteRole("plague_mage"),
		"plague_mage_toxic_nova_pick": types.ChoiceRouteRole("plague_mage"),
	}
}
