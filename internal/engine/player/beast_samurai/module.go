// gameflow: 兽灵武士模块入口声明。

package beast_samurai

import (
	"starcup-engine/internal/engine/player"
	skills "starcup-engine/internal/engine/skill"
	"starcup-engine/internal/types"
)

// SkillEntries 导出角色技能与策略绑定入口。
func SkillEntries() []player.SkillEntry {
	return []player.SkillEntry{
		{ID: "bs_warrior_zanshin", Handler: &skills.BeastSamuraiWarriorZanshinHandler{}},
		{ID: "bs_one_strike_no_thought", Handler: &skills.BeastSamuraiOneStrikeNoThoughtHandler{}},
		{ID: "bs_one_strike_intercept", Handler: &skills.BeastSamuraiOneStrikeInterceptHandler{}},
		{ID: "bs_beast_soul_will", Handler: &skills.BeastSamuraiBeastSoulWillHandler{}},
		{ID: "bs_beast_soul_alert", Handler: &skills.BeastSamuraiBeastSoulAlertHandler{}},
		{ID: "bs_beast_return", Handler: &skills.BeastSamuraiBeastReturnHandler{}},
		{ID: "bs_iaijutsu_turn_end_drain", Handler: &skills.BeastSamuraiIaijutsuTurnEndDrainHandler{}},
		{ID: "bs_iaijutsu_exit_on_deal_damage", Handler: &skills.BeastSamuraiIaijutsuExitOnDealDamageHandler{}},
		{ID: "bs_iaijutsu_exit_on_zero", Handler: &skills.BeastSamuraiIaijutsuExitOnZeroHandler{}},
		{ID: "bs_iaijutsu_tapped_target_boost", Handler: &skills.BeastSamuraiIaijutsuTappedBoostHandler{}},
		{ID: "bs_reversal_iaijutsu", Handler: &skills.BeastSamuraiReversalIaijutsuSlashHandler{}},
		{ID: "bs_iaijutsu_style", Handler: &skills.BeastSamuraiIaijutsuStyleHandler{}},
	}
}

// ChoiceRouteSpecs 导出角色 choice 路由声明。
func ChoiceRouteSpecs() map[string]types.ChoiceRouteSpec {
	return map[string]types.ChoiceRouteSpec{
		"bs_alert_source_discard":        types.ChoiceRouteRole("beast"),
		"bs_beast_return_self_discard":   types.ChoiceRouteRole("beast"),
		"bs_beast_return_source_discard": types.ChoiceRouteRole("beast"),
		"bs_beast_return_x":              types.ChoiceRouteRole("beast"),
		"bs_iaijutsu_draw_pick":          types.ChoiceRouteRole("beast"),
		"bs_iaijutsu_mode_pick":          types.ChoiceRouteRole("beast"),
		"bs_iaijutsu_style_discard":      types.ChoiceRouteRole("beast"),
		"bs_iaijutsu_style_mode":         types.ChoiceRouteRole("beast"),
		"bs_reversal_target_discard":     types.ChoiceRouteRole("beast"),
		"bs_reversal_x":                  types.ChoiceRouteRole("beast"),
	}
}
