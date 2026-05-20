// gameflow: 兽灵武士模块入口声明。

package beast_samurai

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
	"starcup-engine/internal/types"
)

// RoleEntry 导出角色统一入口定义。
func RoleEntry() player.RoleEntry {
	return player.RoleEntry{
		ID:               "beast_samurai",
		Defaults:         ApplyDefaults,
		Choices:          NewChoiceHandler(),
		ChoiceSpecs:      ChoiceSpecs(),
		Skills:           SkillEntries(),
		ChoiceRouteSpecs: ChoiceRouteSpecs(),
		TimingHookSpecs: []player.TimingHookSpec{
			{Timing: player.TimingOnDamageCalculate, Priority: 900, Hook: damageCalculateHook},
			{Timing: player.TimingOnAttackStateReset, Priority: 100, Hook: attackStateResetHook},
			{Timing: player.TimingOnAttackGating, Priority: 300, Hook: attackGatingHook},
			{Timing: player.TimingPostAttackHit, Priority: 200, Hook: postAttackHitHook},
			{Timing: player.TimingOnAttackMiss, Priority: 100, Hook: attackMissHook},
			{Timing: player.TimingPostDamageResolved, Priority: 100, Hook: postDamageResolvedHook},
			{Timing: player.TimingOnTurnEnd, Priority: 100, Hook: turnEndHook},
			{Timing: player.TimingOnTurnEndFinal, Priority: 100, Hook: turnEndFinalHook},
			{Timing: player.TimingOnResponseSkillAug, Priority: 100, Hook: responseSkillAugmentHook},
		},
	}
}

// ChoiceSpecs 导出角色 choice 声明（含多选弃牌处理器）。
func ChoiceSpecs() []player.ChoiceSpec {
	return []player.ChoiceSpec{
		{ChoiceType: "bs_reversal_target_discard", HandleMultiSelect: handleReversalTargetDiscardMultiSelect},
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
	p.Tokens["bs_zanshin"] = 0
	p.Tokens["bs_beast_soul"] = 0
}

// SkillEntries 导出角色技能与策略绑定入口。
func SkillEntries() []player.SkillEntry {
	return []player.SkillEntry{
		{ID: "bs_warrior_zanshin", Handler: &BeastSamuraiWarriorZanshinHandler{}},
		{ID: "bs_one_strike_no_thought", Handler: &BeastSamuraiOneStrikeNoThoughtHandler{}},
		{ID: "bs_one_strike_intercept", Handler: &BeastSamuraiOneStrikeInterceptHandler{}},
		{ID: "bs_beast_soul_will", Handler: &BeastSamuraiBeastSoulWillHandler{}},
		{ID: "bs_beast_soul_alert", Handler: &BeastSamuraiBeastSoulAlertHandler{}},
		{ID: "bs_beast_return", Handler: &BeastSamuraiBeastReturnHandler{}},
		{ID: "bs_iaijutsu_turn_end_drain", Handler: &BeastSamuraiIaijutsuTurnEndDrainHandler{}},
		{ID: "bs_iaijutsu_exit_on_deal_damage", Handler: &BeastSamuraiIaijutsuExitOnDealDamageHandler{}},
		{ID: "bs_iaijutsu_exit_on_zero", Handler: &BeastSamuraiIaijutsuExitOnZeroHandler{}},
		{ID: "bs_iaijutsu_tapped_target_boost", Handler: &BeastSamuraiIaijutsuTappedBoostHandler{}},
		{ID: "bs_reversal_iaijutsu", Handler: &BeastSamuraiReversalIaijutsuSlashHandler{}},
		{ID: "bs_iaijutsu_style", Handler: &BeastSamuraiIaijutsuStyleHandler{}},
	}
}

// ChoiceRouteSpecs 导出角色 choice 路由声明。
func ChoiceRouteSpecs() map[string]types.ChoiceRouteSpec {
	return map[string]types.ChoiceRouteSpec{
		"bs_alert_source_discard":        types.ChoiceRouteRole("beast_samurai"),
		"bs_alert_target":                types.ChoiceRouteRole("beast_samurai"),
		"bs_beast_return_self_discard":   types.ChoiceRouteRole("beast_samurai"),
		"bs_beast_return_source_discard": types.ChoiceRouteRole("beast_samurai"),
		"bs_beast_return_x":              types.ChoiceRouteRole("beast_samurai"),
		"bs_iaijutsu_style_discard":      types.ChoiceRouteRole("beast_samurai"),
		"bs_iaijutsu_style_mode":         types.ChoiceRouteRole("beast_samurai"),
		"bs_reversal_target_discard":     types.ChoiceRouteRole("beast_samurai"),
		"bs_reversal_x":                  types.ChoiceRouteRole("beast_samurai"),
	}
}
