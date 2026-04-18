// gameflow: 血之巫女模块入口声明。

package blood_priestess

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
		ID:               "blood_priestess",
		Defaults:         ApplyDefaults,
		StarterCards:     StarterCards,
		Choices:          NewChoiceHandler(),
		Skills:           SkillEntries(),
		ChoiceRouteSpecs: ChoiceRouteSpecs(),
	}
}

// ApplyDefaults 初始化角色默认属性。
func ApplyDefaults(p *model.Player) {
	// blood_priestess 无特殊默认配置
}

// StarterCards 返回开局专属牌列表。
func StarterCards(p *model.Player) []model.Card {
	if p == nil || p.Character == nil {
		return nil
	}
	return []model.Card{
		{
			ID:              fmt.Sprintf("starter-%s-bp_shared_life", p.ID),
			Name:            "同生共死",
			Type:            model.CardTypeMagic,
			Element:         model.ElementDark,
			Faction:         p.Character.Faction,
			Description:     "血之巫女开局自带专属技能卡",
			ExclusiveChar1:  p.Character.ID,
			ExclusiveSkill1: "同生共死",
		},
	}
}

// SkillEntries 导出角色技能与策略绑定入口。
func SkillEntries() []player.SkillEntry {
	return []player.SkillEntry{
		{ID: "bp_blood_sorrow", Handler: &skills.BloodPriestessBloodSorrowHandler{}},
		{ID: "bp_bleeding", Handler: &skills.BloodPriestessBleedingHandler{}},
		{ID: "bp_backflow", Handler: &skills.BloodPriestessBackflowHandler{}},
		{ID: "bp_blood_wail", Handler: &skills.BloodPriestessBloodWailHandler{}},
		{ID: "bp_shared_life", Handler: &skills.BloodPriestessSharedLifeHandler{}},
		{ID: "bp_blood_curse", Handler: &skills.BloodPriestessBloodCurseHandler{}},
	}
}

// ChoiceRouteSpecs 导出角色 choice 路由声明。
func ChoiceRouteSpecs() map[string]types.ChoiceRouteSpec {
	return map[string]types.ChoiceRouteSpec{
		"bp_blood_curse_pick":    types.ChoiceRouteRole("blood_priestess"),
		"bp_blood_sorrow_mode":   types.ChoiceRouteRole("blood_priestess"),
		"bp_blood_sorrow_target": types.ChoiceRouteRole("blood_priestess"),
		"bp_blood_wail_x":        types.ChoiceRouteRole("blood_priestess"),
		"bp_curse_discard":       types.ChoiceRouteRole("blood_priestess"),
		"bp_shared_life_pick":    types.ChoiceRouteRole("blood_priestess"),
		"bp_shared_life_target":  types.ChoiceRouteRole("blood_priestess"),
	}
}
