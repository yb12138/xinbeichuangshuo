// gameflow: 剑帝模块入口声明。

package sword_emperor

import (
	"starcup-engine/internal/engine/player"
	skills "starcup-engine/internal/engine/skill"
	"starcup-engine/internal/types"
)

// SkillEntries 导出角色技能与策略绑定入口。
func SkillEntries() []player.SkillEntry {
	return []player.SkillEntry{
		{ID: "se_sword_soul_guard", Handler: &skills.SwordEmperorSwordSoulGuardHandler{}},
		{ID: "se_feint", Handler: &skills.SwordEmperorFeintHandler{}},
		{ID: "se_sword_qi_slash", Handler: &skills.SwordEmperorSwordQiSlashHandler{}},
		{ID: "se_angel_soul", Handler: &skills.SwordEmperorAngelSoulHandler{}},
		{ID: "se_demon_soul", Handler: &skills.SwordEmperorDemonSoulHandler{}},
		{ID: "se_angel_soul_hit", Handler: &skills.SwordEmperorAngelSoulHitHandler{}},
		{ID: "se_angel_soul_miss", Handler: &skills.SwordEmperorAngelSoulMissHandler{}},
		{ID: "se_demon_soul_miss", Handler: &skills.SwordEmperorDemonSoulMissHandler{}},
		{ID: "se_indomitable_will", Handler: &skills.SwordEmperorIndomitableWillHandler{}},
	}
}

// ChoiceRouteSpecs 导出角色 choice 路由声明。
func ChoiceRouteSpecs() map[string]types.ChoiceRouteSpec {
	return map[string]types.ChoiceRouteSpec{
		"se_soul_pick":             types.ChoiceRouteRole("sword_emperor"),
		"se_sword_qi_slash_target": types.ChoiceRouteTargetPrompt("se"),
		"se_sword_qi_slash_x":      types.ChoiceRouteRole("sword_emperor"),
	}
}
