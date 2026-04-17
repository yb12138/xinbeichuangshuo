// gameflow: 暗杀者模块入口声明。

package assassin

import (
	"starcup-engine/internal/engine/player"
	skills "starcup-engine/internal/engine/skill"
	"starcup-engine/internal/types"
)

// SkillEntries 导出角色技能与策略绑定入口。
func SkillEntries() []player.SkillEntry {
	return []player.SkillEntry{
		{ID: "backlash", Handler: &skills.BacklashHandler{}},
		{ID: "water_shadow", Handler: &skills.WaterShadowHandler{}},
		{ID: "stealth", Handler: &skills.StealthHandler{}},
	}
}

// ChoiceRouteSpecs 导出角色 choice 路由声明。
func ChoiceRouteSpecs() map[string]types.ChoiceRouteSpec {
	return map[string]types.ChoiceRouteSpec{
		"assassin_stealth_draw": types.ChoiceRouteRole("assassin"),
	}
}
