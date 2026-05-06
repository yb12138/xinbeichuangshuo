// gameflow: 封印师模块入口声明。

package sealer

import (
	"fmt"
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
	"starcup-engine/internal/types"
)

// RoleEntry 导出角色统一入口定义。
func RoleEntry() player.RoleEntry {
	return player.RoleEntry{
		ID:               "sealer",
		Defaults:         ApplyDefaults,
		StarterCards:     StarterCards,
		Choices:          NewChoiceHandler(),
		Skills:           SkillEntries(),
		ChoiceRouteSpecs: ChoiceRouteSpecs(),
	}
}

// ApplyDefaults 初始化角色默认属性。
func ApplyDefaults(p *model.Player) {
	// sealer 无特殊默认配置
}

// StarterCards 返回开局专属牌列表。
func StarterCards(p *model.Player) []model.Card {
	if p == nil || p.Character == nil {
		return nil
	}
	return []model.Card{
		{
			ID:              fmt.Sprintf("starter-%s-five_elements_bind", p.ID),
			Name:            "五系束缚",
			Type:            model.CardTypeMagic,
			Element:         model.ElementLight,
			Faction:         p.Character.Faction,
			Description:     "封印师开局自带专属技能卡",
			ExclusiveChar1:  p.Character.ID,
			ExclusiveSkill1: "五系束缚",
		},
	}
}

// SkillEntries 导出角色技能与策略绑定入口。
func SkillEntries() []player.SkillEntry {
	return []player.SkillEntry{
		{ID: "magic_surge", Handler: &MagicSurgeHandler{}},
		{
			ID:      "seal_break",
			Handler: &SealBreakHandler{},
			Policy: types.SkillPolicy{
				TargetRules: types.TargetRuleSet{
					Count: types.TargetCountRule{Min: 0, Max: 1},
					Checks: []types.TargetCheckRule{
						{Kind: types.TargetCheckAnyBasicFieldWhenNone, Err: "场上没有可收回的基础效果"},
						{Kind: types.TargetCheckHasBasicFieldOnTarget, Index: 0, Err: "%s 面前没有可收回的基础效果", WithTargetName: true},
					},
				},
			},
		},
		{ID: "five_elements_bind", Handler: &FiveElementsBindHandler{}},
		{ID: "water_seal", Handler: NewWaterSealHandler()},
		{ID: "fire_seal", Handler: NewFireSealHandler()},
		{ID: "earth_seal", Handler: NewEarthSealHandler()},
		{ID: "wind_seal", Handler: NewWindSealHandler()},
		{ID: "thunder_seal", Handler: NewThunderSealHandler()},
	}
}

// ChoiceRouteSpecs 导出角色 choice 路由声明。
func ChoiceRouteSpecs() map[string]types.ChoiceRouteSpec {
	return map[string]types.ChoiceRouteSpec{
		"five_elements_bind":             types.ChoiceRouteRole("sealer"),
		"sealer_five_elements_bind_pick": types.ChoiceRouteRole("sealer"),
	}
}
