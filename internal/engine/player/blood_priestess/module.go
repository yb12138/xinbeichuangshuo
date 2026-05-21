// gameflow: 血之巫女模块入口声明。

package blood_priestess

import (
	"fmt"
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
	"starcup-engine/internal/types"
)

// RoleEntry 导出角色统一入口定义。
func RoleEntry() player.RoleEntry {
	return player.RoleEntry{
		ID:                "blood_priestess",
		Defaults:          ApplyDefaults,
		StarterCards:      StarterCards,
		HandLimitModifier: SharedLifeHandLimitModifier,
		FlowContinuationHandlers: map[model.FlowContinuationKind]player.FlowContinuationHandler{
			model.FlowContinuationAfterDraw:   handleSharedLifeAfterDraw,
			model.FlowContinuationAfterDamage: handleBloodPriestessAfterDamage,
		},
		Choices:          NewChoiceHandler(),
		ChoiceSpecs:      ChoiceSpecs(),
		Skills:           SkillEntries(),
		ChoiceRouteSpecs: ChoiceRouteSpecs(),
		TimingHookSpecs: []player.TimingHookSpec{
			{Timing: player.TimingPostActionEnd, Priority: 100, Hook: postActionEndBleedExitHook, RoleFilter: &player.HookRoleNone},
			{Timing: player.TimingOnMoraleLossApplied, Priority: 100, Hook: moraleLossHook},
			{Timing: player.TimingOnTurnStart, Priority: 100, Hook: turnStartBleedTickHook},
		},
	}
}

// ChoiceSpecs 导出角色 choice 声明。
func ChoiceSpecs() []player.ChoiceSpec {
	return []player.ChoiceSpec{
		{
			ChoiceType:          "bp_curse_discard",
			SequentialRemaining: player.ChoiceRemainingFromFlowSelectionCount(bloodCurseDiscardNeedStep, bloodCurseDiscardCardsStep),
		},
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
		{ID: "bp_blood_sorrow", Handler: &BloodSorrowHandler{}},
		{ID: "bp_bleeding", Handler: &BleedingHandler{}},
		{ID: "bp_backflow", Handler: &BackflowHandler{}},
		{ID: "bp_blood_wail", Handler: &BloodWailHandler{}},
		{
			ID:      "bp_shared_life",
			Handler: &SharedLifeHandler{},
			Policy: types.SkillPolicy{
				ManualExclusiveCard: true,
			},
		},
		{ID: "bp_blood_curse", Handler: &BloodCurseHandler{}},
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
