// gameflow: 血色剑灵模块入口声明。

package crimson_sword_spirit

import (
	"fmt"
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
	"starcup-engine/internal/types"
)

// RoleEntry 导出角色统一入口定义。
func RoleEntry() player.RoleEntry {
	return player.RoleEntry{
		ID:               "crimson_sword_spirit",
		Defaults:         ApplyDefaults,
		StarterCards:     StarterCards,
		Choices:          NewChoiceHandler(),
		Skills:           SkillEntries(),
		ChoiceRouteSpecs: ChoiceRouteSpecs(),
		TimingHookSpecs: []player.TimingHookSpec{
			{Timing: player.TimingOnDamageAfterApply, Priority: 100, Hook: afterApplyHook},
			{Timing: player.TimingOnHealResist, Priority: 100, Hook: healResistHook},
			{Timing: player.TimingOnTurnEnd, Priority: 500, Hook: turnEndHook},
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
	p.Tokens["css_blood_cap"] = 3
	p.Tokens["css_blood"] = 0
}

// StarterCards 返回开局专属牌列表。
func StarterCards(p *model.Player) []model.Card {
	if p == nil || p.Character == nil {
		return nil
	}
	return []model.Card{
		{
			ID:              fmt.Sprintf("starter-%s-css_rose_courtyard", p.ID),
			Name:            "血蔷薇庭院",
			Type:            model.CardTypeMagic,
			Element:         model.ElementDark,
			Faction:         p.Character.Faction,
			Description:     "血色剑灵开局自带专属技能卡",
			ExclusiveChar1:  p.Character.ID,
			ExclusiveSkill1: "血蔷薇庭院",
		},
	}
}

// SkillEntries 导出角色技能与策略绑定入口。
func SkillEntries() []player.SkillEntry {
	return []player.SkillEntry{
		{ID: "css_blood_thorns", Handler: &CrimsonBloodThornsHandler{}},
		{ID: "css_crimson_flash", Handler: &CrimsonFlashHandler{}},
		{
			ID:      "css_blood_rose",
			Handler: &CrimsonBloodRoseHandler{},
			Policy: types.SkillPolicy{
				TargetRules: types.TargetRuleSet{
					// 分步选择模式：由后端流程控制，前端不需要一次性选目标
					// 使用 Err 字段触发 HasCountOverride，使 min/max=0生效
					Count: types.TargetCountRule{Min: 0, Max: 0, Err: "分步选择"},
				},
			},
		},
		{ID: "css_blood_barrier", Handler: &CrimsonBloodBarrierHandler{}},
		{ID: "css_rose_courtyard", Handler: &CrimsonRoseCourtyardHandler{}},
		{ID: "css_dance", Handler: &CrimsonDanceHandler{}},
	}
}

// ChoiceRouteSpecs 导出角色 choice 路由声明。
func ChoiceRouteSpecs() map[string]types.ChoiceRouteSpec {
	return map[string]types.ChoiceRouteSpec{
		"css_dance_mode":                    types.ChoiceRouteRole("crimson_sword_spirit"),
		"css_rose_courtyard_pick":           types.ChoiceRouteRole("crimson_sword_spirit"),
		"css_blood_rose_remove_heal_target": types.ChoiceRouteRole("crimson_sword_spirit"),
		"css_blood_rose_gain_heal_target":   types.ChoiceRouteRole("crimson_sword_spirit"),
	}
}
