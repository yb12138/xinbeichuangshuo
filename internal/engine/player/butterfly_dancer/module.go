// gameflow: 蝶舞者模块入口声明。

package butterfly_dancer

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
	"starcup-engine/internal/types"
)

// RoleEntry 导出角色统一入口定义。
func RoleEntry() player.RoleEntry {
	return player.RoleEntry{
		ID:       "butterfly_dancer",
		Defaults: ApplyDefaults,
		HandLimit: player.HandLimitRuleFuncs{
			Modifier: func(p *model.Player, current int) int {
				current -= Pupa(p)
				if current < 3 {
					return 3
				}
				return current
			},
		},
		Choices:          NewChoiceHandler(),
		ChoiceSpecs:      ChoiceSpecs(),
		Skills:           SkillEntries(),
		ChoiceRouteSpecs: ChoiceRouteSpecs(),
		TimingHookSpecs: []player.TimingHookSpec{
			{Timing: player.TimingOnTurnBeforeStart, Priority: 100, Hook: witherExpiryHook},
			{Timing: player.TimingDamageApplied, Priority: 100, Hook: damageResponseHook, RoleFilter: &player.HookRoleNone},
			{Timing: player.TimingDamageTaken, Priority: 100, Hook: pilgrimageBeforeApplyHook, RoleFilter: &player.HookRoleNone},
		},
		MoraleLossModifier: func(engine player.MoraleLossModifierEngine, camp model.Camp, current int, proposedLoss int, extra player.MoraleLossModifierExtra) int {
			for _, p := range engine.GetAllPlayers() {
				if p.Camp != camp && WitherActive(p) {
					maxLoss := current - 1
					if maxLoss < 0 {
						maxLoss = 0
					}
					if proposedLoss > maxLoss {
						return maxLoss
					}
					return proposedLoss
				}
			}
			return proposedLoss
		},
	}
}

// ChoiceSpecs 导出角色 choice 声明（含多选处理器）。
func ChoiceSpecs() []player.ChoiceSpec {
	return []player.ChoiceSpec{
		{ChoiceType: "bt_cocoon_overflow_discard", HandleMultiSelect: handleCocoonOverflowDiscardMultiSelect},
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
	p.Tokens["bt_pupa"] = 0
}

// SkillEntries 导出角色技能与策略绑定入口。
func SkillEntries() []player.SkillEntry {
	return []player.SkillEntry{
		{ID: "bt_life_fire", Handler: &ButterflyLifeFireHandler{}},
		{ID: "bt_dance", Handler: &ButterflyDanceHandler{}},
		{ID: "bt_poison_powder", Handler: &ButterflyPoisonPowderHandler{}},
		{ID: "bt_pilgrimage", Handler: &ButterflyPilgrimageHandler{}},
		{ID: "bt_mirror", Handler: &ButterflyMirrorHandler{}},
		{ID: "bt_wither", Handler: &ButterflyWitherHandler{}},
		{ID: "bt_chrysalis", Handler: &ButterflyChrysalisHandler{}},
		{ID: "bt_reverse_butterfly", Handler: &ButterflyReverseHandler{}},
	}
}

// ChoiceRouteSpecs 导出角色 choice 路由声明。
func ChoiceRouteSpecs() map[string]types.ChoiceRouteSpec {
	return map[string]types.ChoiceRouteSpec{
		"bt_cocoon_overflow_discard": types.ChoiceRouteRole("butterfly_dancer"),
		"bt_cocoon_pick":             types.ChoiceRouteRole("butterfly_dancer"),
		"bt_dance_discard":           types.ChoiceRouteRole("butterfly_dancer"),
		"bt_dance_mode":              types.ChoiceRouteRole("butterfly_dancer"),
		"bt_mirror_pair":             types.ChoiceRouteRole("butterfly_dancer"),
		"bt_pilgrimage_confirm":      types.ChoiceRouteRole("butterfly_dancer"),
		"bt_pilgrimage_pick":         types.ChoiceRouteRole("butterfly_dancer"),
		"bt_poison_pick":             types.ChoiceRouteRole("butterfly_dancer"),
		"bt_reverse_branch1_pick":    types.ChoiceRouteRole("butterfly_dancer"),
		"bt_reverse_branch2_cost":    types.ChoiceRouteRole("butterfly_dancer"),
		"bt_reverse_branch2_pick":    types.ChoiceRouteRole("butterfly_dancer"),
		"bt_reverse_mode":            types.ChoiceRouteRole("butterfly_dancer"),
		"bt_reverse_target":          types.ChoiceRouteRole("butterfly_dancer"),
		"bt_wither_confirm":          types.ChoiceRouteRole("butterfly_dancer"),
		"bt_wither_target":           types.ChoiceRouteRole("butterfly_dancer"),
	}
}
