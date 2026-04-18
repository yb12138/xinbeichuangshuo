// gameflow: 剑帝模块入口声明。

package sword_emperor

import (
	"starcup-engine/internal/engine/player"
	skills "starcup-engine/internal/engine/skill"
	"starcup-engine/internal/model"
	"starcup-engine/internal/types"
)

// RoleEntry 导出角色统一入口定义。
func RoleEntry() player.RoleEntry {
	return player.RoleEntry{
		ID:               "sword_emperor",
		Defaults:         ApplyDefaults,
		Choices:          NewChoiceHandler(),
		Skills:           SkillEntries(),
		ChoiceRouteSpecs: ChoiceRouteSpecs(),
	}
}

// ApplyDefaults 初始化角色默认属性。
func ApplyDefaults(p *model.Player) {
	if p == nil {
		return
	}
	if p.Tokens == nil {
		p.Tokens = map[string]int{}
	}
	p.Tokens["se_sword_qi"] = 0
	p.Tokens["se_sword_soul_count"] = 0
}

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
