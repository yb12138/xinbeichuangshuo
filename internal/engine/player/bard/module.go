// gameflow: 吟游诗人模块入口声明。

package bard

import (
	"starcup-engine/internal/engine/player"
	skills "starcup-engine/internal/engine/skill"
	"starcup-engine/internal/types"
)

// SkillEntries 导出角色技能与策略绑定入口。
func SkillEntries() []player.SkillEntry {
	return []player.SkillEntry{
		{ID: "bd_descent_concerto", Handler: &skills.BardDescentConcertoHandler{}},
		{
			ID:      "bd_dissonance_chord",
			Handler: &skills.BardDissonanceChordHandler{},
			Policy: types.SkillPolicy{
				TargetRules: types.TargetRuleSet{
					Count: types.TargetCountRule{Min: 0, Max: 1},
				},
			},
		},
		{ID: "bd_forbidden_verse", Handler: &skills.BardForbiddenVerseHandler{}},
		{ID: "bd_rousing_rhapsody", Handler: &skills.BardRousingRhapsodyHandler{}},
		{ID: "bd_victory_symphony", Handler: &skills.BardVictorySymphonyHandler{}},
		{
			ID:      "bd_hope_fugue",
			Handler: &skills.BardHopeFugueHandler{},
			Policy: types.SkillPolicy{
				ManualExclusiveCard: true,
				TargetRules: types.TargetRuleSet{
					Count: types.TargetCountRule{Min: 0, Max: 1, Err: "希望赋格曲至多指定1名其他队友"},
					Slots: []types.TargetSlotRule{
						{Index: 0, Camp: types.TargetCampAlly, Self: types.TargetSelfOther, Err: "希望赋格曲的目标必须是其他队友"},
					},
				},
			},
		},
	}
}

// ChoiceRouteSpecs 导出角色 choice 路由声明。
func ChoiceRouteSpecs() map[string]types.ChoiceRouteSpec {
	return map[string]types.ChoiceRouteSpec{
		"bd_descent_cards":           types.ChoiceRouteRole("bard"),
		"bd_descent_element":         types.ChoiceRouteRole("bard"),
		"bd_descent_target":          types.ChoiceRouteTargetPrompt("bard"),
		"bd_dissonance_discard_step": types.ChoiceRouteRole("bard"),
		"bd_dissonance_mode":         types.ChoiceRouteRole("bard"),
		"bd_dissonance_pick":         types.ChoiceRouteRole("bard"),
		"bd_dissonance_target":       types.ChoiceRouteTargetPrompt("bard"),
		"bd_dissonance_x":            types.ChoiceRouteRole("bard"),
		"bd_forbidden_verse_pick":    types.ChoiceRouteRole("bard"),
		"bd_hope_draw_confirm":       types.ChoiceRouteRole("bard"),
		"bd_hope_mode":               types.ChoiceRouteRole("bard"),
		"bd_hope_place_target":       types.ChoiceRouteTargetPrompt("bard"),
		"bd_hope_transfer_discard":   types.ChoiceRouteRole("bard"),
		"bd_hope_transfer_target":    types.ChoiceRouteTargetPrompt("bard"),
		"bd_rousing_discard_cards":   types.ChoiceRouteRole("bard"),
		"bd_rousing_mode":            types.ChoiceRouteRole("bard"),
		"bd_rousing_targets":         types.ChoiceRouteRole("bard"),
		"bd_victory_extract_stone":   types.ChoiceRouteRole("bard"),
		"bd_victory_mode":            types.ChoiceRouteRole("bard"),
	}
}
