// gameflow: 鬼术师模块入口声明。

package onmyoji

import (
	"fmt"
	"strings"

	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
	"starcup-engine/internal/types"
)

// RoleEntry 导出角色统一入口定义。
func RoleEntry() player.RoleEntry {
	return player.RoleEntry{
		ID:               "onmyoji",
		Defaults:         ApplyDefaults,
		Choices:          NewChoiceHandler(),
		Skills:           SkillEntries(),
		ChoiceRouteSpecs: ChoiceRouteSpecs(),
		TimingHookSpecs: []player.TimingHookSpec{
			{Timing: player.TimingOnTurnEnd, Priority: 100, Hook: turnEndDarkRitualHook},
			{Timing: player.TimingOnCombatInteraction, Priority: 100, Hook: combatInteractionTimingHook},
			{Timing: player.TimingOnCounterElementCheck, Priority: 100, Hook: factionElementHook},
			{Timing: player.TimingOnCounterResolve, Priority: 100, Hook: factionResolveHook},
		},
		SkillUsabilityCheckers: map[string]player.SkillUsabilityChecker{
			"onmyoji_shikigami_descend": CheckShikigamiDescendUsability,
		},
		MaybeDarkRitual: MaybeDarkRitual,
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
	p.Tokens["onmyoji_ghost_fire"] = 0
}

// SkillEntries 导出角色技能与策略绑定入口。
func SkillEntries() []player.SkillEntry {
	return []player.SkillEntry{
		{
			ID:      "onmyoji_shikigami_descend",
			Handler: &OnmyojiShikigamiDescendHandler{},
			Policy: types.SkillPolicy{
				ValidateDiscardedCards: func(ctx types.PolicyContext) error {
					if len(ctx.DiscardedCards) != 2 {
						return fmt.Errorf("式神降临需要弃置2张手牌")
					}
					f1 := strings.TrimSpace(ctx.DiscardedCards[0].Faction)
					f2 := strings.TrimSpace(ctx.DiscardedCards[1].Faction)
					if f1 == "" || f2 == "" || f1 != f2 {
						return fmt.Errorf("式神降临需要弃置2张命格相同的手牌")
					}
					return nil
				},
			},
		},
		{ID: "onmyoji_yinyang_shift", Handler: &OnmyojiYinYangShiftHandler{}},
		{ID: "onmyoji_shikigami_shift", Handler: &OnmyojiShikigamiShiftHandler{}},
		{ID: "onmyoji_dark_ritual", Handler: &OnmyojiDarkRitualHandler{}},
		{ID: "onmyoji_binding", Handler: &OnmyojiBindingHandler{}},
		{
			ID:      "onmyoji_life_barrier",
			Handler: &OnmyojiLifeBarrierHandler{},
			Policy: types.SkillPolicy{
				TargetRules: types.TargetRuleSet{
					Count: types.TargetCountRule{Min: 0, Max: 0},
				},
			},
		},
	}
}

// ChoiceRouteSpecs 导出角色 choice 路由声明。
func ChoiceRouteSpecs() map[string]types.ChoiceRouteSpec {
	return map[string]types.ChoiceRouteSpec{
		"onmyoji_binding_card":                types.ChoiceRouteRole("onmyoji"),
		"onmyoji_binding_confirm":             types.ChoiceRouteRole("onmyoji"),
		"onmyoji_binding_counter_target":      types.ChoiceRouteRole("onmyoji"),
		"onmyoji_binding_pick":                types.ChoiceRouteRole("onmyoji"),
		"onmyoji_dark_ritual_pick":            types.ChoiceRouteRole("onmyoji"),
		"onmyoji_dark_ritual_target":          types.ChoiceRouteRole("onmyoji"),
		"onmyoji_life_barrier_mode":           types.ChoiceRouteRole("onmyoji"),
		"onmyoji_life_barrier_release_combo":  types.ChoiceRouteRole("onmyoji"),
		"onmyoji_life_barrier_release_target": types.ChoiceRouteRole("onmyoji"),
		"onmyoji_life_barrier_support_target": types.ChoiceRouteRole("onmyoji"),
		"onmyoji_shikigami_pick":              types.ChoiceRouteRole("onmyoji"),
		"onmyoji_shikigami_shift_pick":        types.ChoiceRouteRole("onmyoji"),
		"onmyoji_yinyang_card":                types.ChoiceRouteRole("onmyoji"),
		"onmyoji_yinyang_confirm":             types.ChoiceRouteRole("onmyoji"),
		"onmyoji_yinyang_counter_target":      types.ChoiceRouteRole("onmyoji"),
		"onmyoji_yinyang_shift_pick":          types.ChoiceRouteRole("onmyoji"),
	}
}
