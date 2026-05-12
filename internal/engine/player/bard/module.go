// gameflow: 吟游诗人模块入口声明。

package bard

import (
	"fmt"

	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
	"starcup-engine/internal/types"
)

// RoleEntry 导出角色统一入口定义。
func RoleEntry() player.RoleEntry {
	return player.RoleEntry{
		ID:               "bard",
		Defaults:         ApplyDefaults,
		StarterCards:     StarterCards,
		Choices:          NewChoiceHandler(),
		ChoiceSpecs:      ChoiceSpecs(),
		Skills:           SkillEntries(),
		ChoiceRouteSpecs: ChoiceRouteSpecs(),
		TimingHookSpecs: []player.TimingHookSpec{
			{Timing: player.TimingOnTurnStart, Priority: 200, Hook: turnStartRousingHook, RoleFilter: &player.HookRoleNone},
			{Timing: player.TimingOnTurnEnd, Priority: 200, Hook: turnEndVictoryHook, RoleFilter: &player.HookRoleNone},
			{Timing: player.TimingPostDamageResolved, Priority: 500, Hook: postDamageResolvedHook},
		},
		SkillUsabilityCheckers: map[string]player.SkillUsabilityChecker{
			"bd_dissonance_chord": CheckDissonanceChordUsability,
		},
	}
}

// ChoiceSpecs 导出角色 choice 声明。
func ChoiceSpecs() []player.ChoiceSpec {
	return []player.ChoiceSpec{
		{ChoiceType: "bd_descent_cards", SequentialRemaining: player.ChoiceRemainingFromFixedTotal(2)},
		{ChoiceType: "bd_dissonance_discard_step", SequentialRemaining: player.ChoiceRemainingFromNeedAndSelected("need_count", "selected_count")},
		{ChoiceType: "bd_rousing_discard_cards", SequentialRemaining: player.ChoiceRemainingFromFixedTotal(2)},
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
	p.Tokens["bd_inspiration"] = 0
}

// StarterCards 返回开局专属牌列表。
func StarterCards(p *model.Player) []model.Card {
	if p == nil || p.Character == nil {
		return nil
	}
	return []model.Card{
		{
			ID:              fmt.Sprintf("starter-%s-bd_eternal_movement", p.ID),
			Name:            "永恒乐章",
			Type:            model.CardTypeMagic,
			Element:         model.ElementDark,
			Faction:         p.Character.Faction,
			Description:     "吟游诗人开局自带专属牌",
			ExclusiveChar1:  p.Character.ID,
		},
	}
}

// SkillEntries 导出角色技能与策略绑定入口。
func SkillEntries() []player.SkillEntry {
	return []player.SkillEntry{
		{ID: "bd_descent_concerto", Handler: &BardDescentConcertoHandler{}},
		{
			ID:      "bd_dissonance_chord",
			Handler: &BardDissonanceChordHandler{},
			Policy: types.SkillPolicy{
				TargetRules: types.TargetRuleSet{
					Count: types.TargetCountRule{Min: 0, Max: 1},
				},
			},
		},
		{ID: "bd_forbidden_verse", Handler: &BardForbiddenVerseHandler{}},
		{ID: "bd_rousing_rhapsody", Handler: &BardRousingRhapsodyHandler{}},
		{ID: "bd_victory_symphony", Handler: &BardVictorySymphonyHandler{}},
		{
			ID:      "bd_hope_fugue",
			Handler: &BardHopeFugueHandler{},
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
		"bd_descent_target":          types.ChoiceRouteRole("bard"),
		"bd_dissonance_discard_step": types.ChoiceRouteRole("bard"),
		"bd_dissonance_mode":         types.ChoiceRouteRole("bard"),
		"bd_dissonance_target":       types.ChoiceRouteRole("bard"),
		"bd_dissonance_x":            types.ChoiceRouteRole("bard"),
		"bd_forbidden_verse_pick":    types.ChoiceRouteRole("bard"),
		"bd_hope_draw_confirm":       types.ChoiceRouteRole("bard"),
		"bd_hope_mode":               types.ChoiceRouteRole("bard"),
		"bd_hope_place_target":       types.ChoiceRouteRole("bard"),
		"bd_hope_transfer_discard":   types.ChoiceRouteRole("bard"),
		"bd_hope_transfer_target":    types.ChoiceRouteRole("bard"),
		"bd_rousing_discard_cards":   types.ChoiceRouteRole("bard"),
		"bd_rousing_mode":            types.ChoiceRouteRole("bard"),
		"bd_rousing_targets":         types.ChoiceRouteRole("bard"),
		"bd_victory_extract_stone":   types.ChoiceRouteRole("bard"),
		"bd_victory_mode":            types.ChoiceRouteRole("bard"),
	}
}
