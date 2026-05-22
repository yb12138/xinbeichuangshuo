// gameflow: 圣弓模块入口声明。

package holy_bow

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
	"starcup-engine/internal/types"
)

// RoleEntry 导出角色统一入口定义。
func RoleEntry() player.RoleEntry {
	return player.RoleEntry{
		ID:               "holy_bow",
		Defaults:         ApplyDefaults,
		Choices:          NewChoiceHandler(),
		ChoiceSpecs:      ChoiceSpecs(),
		Skills:           SkillEntries(),
		ChoiceRouteSpecs: ChoiceRouteSpecs(),
		TimingHookSpecs: []player.TimingHookSpec{
			{Timing: player.TimingTurnStart, Priority: 100, Hook: turnStartResetHook},
			{Timing: player.TimingDamageSourceDeal, Priority: 700, Hook: damageCalculateHook},
			{Timing: player.TimingPostAttackHit, Priority: 100, Hook: postAttackHitHook},
			{Timing: player.TimingTurnEndFinal, Priority: 800, Hook: turnEndAutoFillHook},
			{Timing: player.TimingAttackMiss, Priority: 400, Hook: attackMissHook},
			{Timing: player.TimingOnSpecialActionPost, Priority: 100, Hook: holyGloryExitHook},
		},
	}
}

// ChoiceSpecs 导出角色 choice 声明。
func ChoiceSpecs() []player.ChoiceSpec {
	return []player.ChoiceSpec{
		{ChoiceType: "hb_holy_shard_combo", HandleMultiSelect: handleHolyShardComboMultiSelect},
		{ChoiceType: "hb_light_burst_mode_b_discard", SequentialRemaining: player.ChoiceRemainingFromFlowSelectionCount(lightBurstStepModeBX, lightBurstStepModeBDiscard)},
	}
}

// ApplyDefaults 初始化角色默认属性。
func ApplyDefaults(p *model.Player) {
	if p == nil {
		return
	}
	p.Crystal += 2
	p.MaxHeal += 1
	if p.Tokens == nil {
		p.Tokens = map[string]int{}
	}
	p.Tokens["hb_cannon"] = 1
	p.Tokens["hb_faith"] = 0
}

// SkillEntries 导出角色技能与策略绑定入口。
func SkillEntries() []player.SkillEntry {
	return []player.SkillEntry{
		{ID: "hb_heavenly_bow", Handler: &HolyBowHeavenlyBowHandler{}},
		{ID: "hb_holy_shard_storm", Handler: &HolyBowShardStormHandler{}},
		{ID: "hb_radiant_descent", Handler: &HolyBowRadiantDescentHandler{}},
		{ID: "hb_light_burst", Handler: &HolyBowLightBurstHandler{}},
		{ID: "hb_meteor_bullet", Handler: &HolyBowMeteorBulletHandler{}},
		{ID: "hb_radiant_cannon", Handler: &HolyBowRadiantCannonHandler{}},
		{ID: "hb_auto_fill", Handler: &HolyBowAutoFillHandler{}},
	}
}

// ChoiceRouteSpecs 导出角色 choice 路由声明。
func ChoiceRouteSpecs() map[string]types.ChoiceRouteSpec {
	return map[string]types.ChoiceRouteSpec{
		"hb_auto_fill_gain":              types.ChoiceRouteRole("holy_bow"),
		"hb_auto_fill_resource":          types.ChoiceRouteRole("holy_bow"),
		"hb_holy_shard_combo":            types.ChoiceRouteRole("holy_bow"),
		"hb_holy_shard_miss_ally_target": types.ChoiceRouteRole("holy_bow"),
		"hb_holy_shard_miss_confirm":     types.ChoiceRouteRole("holy_bow"),
		"hb_holy_shard_miss_x":           types.ChoiceRouteRole("holy_bow"),
		"hb_holy_shard_target":           types.ChoiceRouteRole("holy_bow"),
		"hb_light_burst_mode":            types.ChoiceRouteRole("holy_bow"),
		"hb_light_burst_mode_a_target":   types.ChoiceRouteRole("holy_bow"),
		"hb_light_burst_mode_b_discard":  types.ChoiceRouteRole("holy_bow"),
		"hb_light_burst_mode_b_targets":  types.ChoiceRouteRole("holy_bow"),
		"hb_light_burst_mode_b_x":        types.ChoiceRouteRole("holy_bow"),
		"hb_meteor_bullet_cost":          types.ChoiceRouteRole("holy_bow"),
		"hb_meteor_bullet_target":        types.ChoiceRouteRole("holy_bow"),
		"hb_radiant_cannon_side":         types.ChoiceRouteRole("holy_bow"),
		"hb_radiant_descent_cost":        types.ChoiceRouteRole("holy_bow"),
		"hb_radiant_descent_pick":        types.ChoiceRouteRole("holy_bow"),
	}
}
