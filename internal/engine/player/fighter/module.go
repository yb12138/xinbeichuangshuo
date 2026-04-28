// gameflow: 格斗家模块入口声明。

package fighter

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
	"starcup-engine/internal/types"
)

// RoleEntry 导出角色统一入口定义。
func RoleEntry() player.RoleEntry {
	return player.RoleEntry{
		ID:               "fighter",
		Defaults:         ApplyDefaults,
		Choices:          NewChoiceHandler(),
		Skills:           SkillEntries(),
		ChoiceRouteSpecs: ChoiceRouteSpecs(),
		TimingHookSpecs: []player.TimingHookSpec{
			{Timing: player.TimingOnDamageCalculate, Priority: 400, Hook: damageCalculateHook},
			{Timing: player.TimingOnAttackDeclared, Priority: 200, Hook: pendingDamageInitHook},
			{Timing: player.TimingOnAttackStateReset, Priority: 100, Hook: attackStateResetHook},
			{Timing: player.TimingOnAttackGating, Priority: 200, Hook: attackGatingHook},
			{Timing: player.TimingOnAttackMiss, Priority: 200, Hook: attackMissHook},
			{Timing: player.TimingOnTurnEnd, Priority: 200, Hook: turnEndHook},
			{Timing: player.TimingBeforeActionOption, Priority: 300, Hook: beforeActionOptionHook},
			{Timing: player.TimingBeforeActionValidation, Priority: 300, Hook: beforeActionValidationHook},
			{Timing: player.TimingOnResponseSkillNormalize, Priority: 100, Hook: responseSkillNormalizeHook},
			{Timing: player.TimingOnResponseSkillAdvance, Priority: 200, Hook: responseSkillAdvanceHook},
		},
		BlocksActionType: func(p *model.Player, at model.ActionType) bool {
			return at == model.ActionMagic && BlocksMagicCasting(p)
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
	p.Tokens["fighter_qi"] = 0
}

// SkillEntries 导出角色技能与策略绑定入口。
func SkillEntries() []player.SkillEntry {
	return []player.SkillEntry{
		{ID: "fighter_psi_field", Handler: &FighterPsiFieldHandler{}},
		{ID: "fighter_charge_strike", Handler: &FighterChargeStrikeHandler{}},
		{ID: "fighter_psi_bullet", Handler: &FighterPsiBulletHandler{}},
		{ID: "fighter_hundred_dragon", Handler: &FighterHundredDragonHandler{}},
		{ID: "fighter_burst_crash", Handler: &FighterBurstCrashHandler{}},
		{ID: "fighter_war_god_drive", Handler: &FighterWarGodDriveHandler{}},
		{ID: "fighter_war_god_drive_followup", Handler: &FighterWarGodDriveFollowupHandler{}},
	}
}

// ChoiceRouteSpecs 导出角色 choice 路由声明。
func ChoiceRouteSpecs() map[string]types.ChoiceRouteSpec {
	return map[string]types.ChoiceRouteSpec{
		"fighter_hundred_dragon_target": types.ChoiceRouteTargetPrompt("fighter"),
		"fighter_psi_bullet_target":     types.ChoiceRouteTargetPrompt("fighter"),
	}
}
