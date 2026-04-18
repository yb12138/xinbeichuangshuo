// gameflow: 血色剑灵模块入口声明。

package crimson_sword_spirit

import (
	"fmt"
	"starcup-engine/internal/engine/player"
	skills "starcup-engine/internal/engine/skill"
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
		{ID: "css_blood_thorns", Handler: &skills.CrimsonBloodThornsHandler{}},
		{ID: "css_crimson_flash", Handler: &skills.CrimsonFlashHandler{}},
		{
			ID:      "css_blood_rose",
			Handler: &skills.CrimsonBloodRoseHandler{},
			Policy: types.SkillPolicy{
				TargetRules: types.TargetRuleSet{
					Count: types.TargetCountRule{Min: 2, Max: 2, Err: "血染蔷薇需要恰好指定2名目标"},
					Slots: []types.TargetSlotRule{
						{Index: 1, Camp: types.TargetCampAlly, Err: "血染蔷薇的第2个目标必须是我方角色"},
					},
				},
			},
		},
		{ID: "css_blood_barrier", Handler: &skills.CrimsonBloodBarrierHandler{}},
		{ID: "css_rose_courtyard", Handler: &skills.CrimsonRoseCourtyardHandler{}},
		{ID: "css_dance", Handler: &skills.CrimsonDanceHandler{}},
	}
}

// ChoiceRouteSpecs 导出角色 choice 路由声明。
func ChoiceRouteSpecs() map[string]types.ChoiceRouteSpec {
	return map[string]types.ChoiceRouteSpec{
		"css_dance_mode":          types.ChoiceRouteRole("crimson_sword_spirit"),
		"css_rose_courtyard_pick": types.ChoiceRouteRole("crimson_sword_spirit"),
	}
}
