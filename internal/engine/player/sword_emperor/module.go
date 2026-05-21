// gameflow: 剑帝模块入口声明。

package sword_emperor

import (
	"starcup-engine/internal/engine/player"
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
		TimingHookSpecs: []player.TimingHookSpec{
			{Timing: player.TimingOnDamageCalculate, Priority: 800, Hook: damageCalculateHook},
			{Timing: player.TimingOnAttackStateReset, Priority: 100, Hook: attackStateResetHook},
			{Timing: player.TimingOnAttackMiss, Priority: 500, Hook: attackMissHook},
			{Timing: player.TimingOnAttackMiss, Priority: 510, Hook: angelSoulMissHook},
			{Timing: player.TimingOnAttackMiss, Priority: 520, Hook: demonSoulMissHook},
			{Timing: player.TimingOnAttackMiss, Priority: 900, Hook: attackMissCleanupHook},
			{Timing: player.TimingOnDamageAfterTaken, Priority: 100, Hook: angelSoulHitHook},
		},
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
}

// SkillEntries 导出角色技能与策略绑定入口。
func SkillEntries() []player.SkillEntry {
	return []player.SkillEntry{
		{ID: "se_sword_soul_guard", Handler: &SwordEmperorSwordSoulGuardHandler{}},
		{ID: "se_feint", Handler: &SwordEmperorFeintHandler{}},
		{ID: "se_sword_qi_slash", Handler: &SwordEmperorSwordQiSlashHandler{}},
		{ID: "se_angel_soul", Handler: &SwordEmperorAngelSoulHandler{}},
		{ID: "se_demon_soul", Handler: &SwordEmperorDemonSoulHandler{}},
		{ID: "se_indomitable_will", Handler: &SwordEmperorIndomitableWillHandler{}},
	}
}

// ChoiceRouteSpecs 导出角色 choice 路由声明。
func ChoiceRouteSpecs() map[string]types.ChoiceRouteSpec {
	return map[string]types.ChoiceRouteSpec{
		"se_soul_pick":             types.ChoiceRouteRole("sword_emperor"),
		"se_sword_qi_slash_target": types.ChoiceRouteRole("sword_emperor"),
		"se_sword_qi_slash_x":      types.ChoiceRouteRole("sword_emperor"),
		"se_sword_rain_discard":    types.ChoiceRouteRole("sword_emperor"),
		"se_sword_rain_target":     types.ChoiceRouteRole("sword_emperor"),
	}
}
