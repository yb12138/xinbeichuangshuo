// gameflow: 仲裁者模块入口声明。

package arbiter

import (
	"starcup-engine/internal/engine/player"
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
		TimingHookSpecs: []player.TimingHookSpec{
			{Timing: player.TimingTurnStart, Priority: 100, Hook: turnStartResetHook},
			{Timing: player.TimingTurnStart, Priority: 200, Hook: turnStartJudgmentUpkeepHook},
			{Timing: player.TimingTurnStart, Priority: 300, Hook: turnStartForcedDoomsdayHook},
			{Timing: player.TimingActionStartOption, Priority: 100, Hook: beforeActionOptionHook, RoleFilter: &player.HookRoleNone},
			{Timing: player.TimingActionStartValidation, Priority: 100, Hook: beforeActionValidationHook, RoleFilter: &player.HookRoleNone},
			{Timing: player.TimingSkillPost, Priority: 100, Hook: skillPostCleanupHook, RoleFilter: &player.HookRoleNone},
		},
	}
}

// ApplyDefaults 初始化角色默认属性。
func ApplyDefaults(p *model.Player) {
	if p == nil {
		return
	}
	p.Crystal += 2
}

// SkillEntries 导出角色技能与策略绑定入口。
func SkillEntries() []player.SkillEntry {
	return []player.SkillEntry{
		{ID: "arbiter_law", Handler: &ArbiterLawHandler{}},
		{ID: "arbiter_judgment_tide", Handler: &ArbiterJudgmentTideHandler{}},
		{ID: "arbiter_ritual", Handler: &ArbiterRitualHandler{}},
		{ID: "arbiter_ritual_break", Handler: &ArbiterRitualBreakHandler{}},
		{ID: "arbiter_doomsday", Handler: &ArbiterDoomsdayHandler{}},
		{ID: "arbiter_balance", Handler: &ArbiterBalanceHandler{}},
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
