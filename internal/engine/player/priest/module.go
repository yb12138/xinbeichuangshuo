// gameflow: 神官模块入口声明。

package priest

import (
	"fmt"
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
	"starcup-engine/internal/types"
)

// RoleEntry 导出角色统一入口定义。
func RoleEntry() player.RoleEntry {
	return player.RoleEntry{
		ID:               "priest",
		Defaults:         ApplyDefaults,
		Choices:          NewChoiceHandler(),
		Skills:           SkillEntries(),
		ChoiceRouteSpecs: ChoiceRouteSpecs(),
		TimingHookSpecs: []player.TimingHookSpec{
			{Timing: player.TimingHealCap, Priority: 100, Hook: healCapHook},
		},
		SkillUsabilityCheckers: map[string]player.SkillUsabilityChecker{
			"priest_water_power": CheckWaterPowerDiscardUsability,
		},
	}
}

// ApplyDefaults 初始化角色默认属性。
func ApplyDefaults(p *model.Player) {
	if p == nil {
		return
	}
	p.MaxHeal = 6
}

// SkillEntries 导出角色技能与策略绑定入口。
func SkillEntries() []player.SkillEntry {
	return []player.SkillEntry{
		{ID: "priest_divine_revelation", Handler: &PriestDivineRevelationHandler{}},
		{ID: "priest_divine_bless", Handler: &PriestDivineBlessHandler{}},
		{
			ID:      "priest_water_power",
			Handler: &PriestWaterPowerHandler{},
			Policy: types.SkillPolicy{
				ResolveDiscardCount: func(ctx types.PolicyContext) int {
					return ctx.SkillDef.CostDiscards
				},
				ValidateDiscardedCards: func(ctx types.PolicyContext) error {
					if len(ctx.DiscardedCards) != 2 {
						return fmt.Errorf("水之神力需要选择1张水系牌并额外选择1张手牌交给队友")
					}
					if ctx.DiscardedCards[0].Element != model.ElementWater {
						return fmt.Errorf("水之神力第一张必须弃置水系牌")
					}
					return nil
				},
				ResolveDiscardPile: func(ctx types.PolicyContext) []model.Card {
					if len(ctx.DiscardedCards) >= 2 {
						// 第二张弃牌由技能效果交给队友，不进入弃牌堆。
						return []model.Card{ctx.DiscardedCards[0]}
					}
					return append([]model.Card{}, ctx.DiscardedCards...)
				},
			},
		},
		{ID: "priest_guardian", Handler: &PriestGuardianHandler{}},
		{ID: "priest_divine_contract", Handler: &PriestDivineContractHandler{}},
		{
			ID:      "priest_divine_domain",
			Handler: &PriestDivineDomainHandler{},
			Policy: types.SkillPolicy{
				ResolveDiscardCount: func(ctx types.PolicyContext) int {
					return ctx.SkillDef.CostDiscards
				},
			},
		},
	}
}

// ChoiceRouteSpecs 导出角色 choice 路由声明。
func ChoiceRouteSpecs() map[string]types.ChoiceRouteSpec {
	return map[string]types.ChoiceRouteSpec{
		"priest_divine_contract_target":      types.ChoiceRouteRole("priest"),
		"priest_divine_contract_x":           types.ChoiceRouteRole("priest"),
		"priest_divine_domain_damage_target": types.ChoiceRouteRole("priest"),
		"priest_divine_domain_heal_target":   types.ChoiceRouteRole("priest"),
		"priest_divine_domain_mode":          types.ChoiceRouteRole("priest"),
		"priest_divine_domain_pick":          types.ChoiceRouteRole("priest"),
		"priest_divine_revelation_pick":      types.ChoiceRouteRole("priest"),
	}
}
