// gameflow: 血之巫女模块入口声明。

package blood_priestess

import (
	"starcup-engine/internal/engine/player"
	skills "starcup-engine/internal/engine/skill"
	"starcup-engine/internal/types"
)

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
		"bp_shared_life_pick":    types.ChoiceRouteRole("blood_priestess"),
		"bp_shared_life_target":  types.ChoiceRouteRole("blood_priestess"),
	}
}
