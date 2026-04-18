// gameflow: 仲裁者模块入口声明。

package arbiter

import (
	"starcup-engine/internal/engine/player"
	skills "starcup-engine/internal/engine/skill"
	"starcup-engine/internal/model"
	"starcup-engine/internal/types"
)

// RoleEntry 导出角色统一入口定义。
func RoleEntry() player.RoleEntry {
	return player.RoleEntry{
		ID:               "arbiter",
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
	p.Crystal += 2
	if p.Tokens == nil {
		p.Tokens = map[string]int{}
	}
	p.Tokens["arbiter_law_inited"] = 1
}

// SkillEntries 导出角色技能与策略绑定入口。
func SkillEntries() []player.SkillEntry {
	return []player.SkillEntry{
		{ID: "arbiter_law", Handler: &skills.ArbiterLawHandler{}},
		{ID: "arbiter_judgment_tide", Handler: &skills.ArbiterJudgmentTideHandler{}},
		{ID: "arbiter_ritual", Handler: &skills.ArbiterRitualHandler{}},
		{ID: "arbiter_ritual_break", Handler: &skills.ArbiterRitualBreakHandler{}},
		{ID: "arbiter_doomsday", Handler: &skills.ArbiterDoomsdayHandler{}},
		{ID: "arbiter_balance", Handler: &skills.ArbiterBalanceHandler{}},
	}
}

// ChoiceRouteSpecs 导出角色 choice 路由声明。
func ChoiceRouteSpecs() map[string]types.ChoiceRouteSpec {
	return map[string]types.ChoiceRouteSpec{
		"arbiter_balance_mode": types.ChoiceRouteRole("arbiter"),
		"arbiter_law_pick":     types.ChoiceRouteRole("arbiter"),
		"arbiter_ritual_pick":  types.ChoiceRouteRole("arbiter"),
	}
}
