// gameflow: 勇者模块入口声明。

package hero

import (
	"starcup-engine/internal/engine/player"
	skills "starcup-engine/internal/engine/skill"
	"starcup-engine/internal/types"
)

// SkillEntries 导出角色技能与策略绑定入口。
func SkillEntries() []player.SkillEntry {
	return []player.SkillEntry{
		{ID: "hero_heart", Handler: &skills.HeroHeartHandler{}},
		{ID: "hero_roar", Handler: &skills.HeroRoarHandler{}},
		{ID: "hero_forbidden_power", Handler: &skills.HeroForbiddenPowerHandler{}},
		{ID: "hero_exhaustion", Handler: &skills.HeroExhaustionHandler{}},
		{ID: "hero_calm_mind", Handler: &skills.HeroCalmMindHandler{}},
		{ID: "hero_taunt", Handler: &skills.HeroTauntHandler{}},
		{ID: "hero_dead_duel", Handler: &skills.HeroDeadDuelHandler{}},
	}
}

// ChoiceRouteSpecs 导出角色 choice 路由声明。
func ChoiceRouteSpecs() map[string]types.ChoiceRouteSpec {
	return map[string]types.ChoiceRouteSpec{
		"hero_roar_draw": types.ChoiceRouteRole("hero"),
	}
}
