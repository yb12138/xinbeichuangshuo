// gameflow: 瘟疫法师模块入口声明。

package plague_mage

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
	"starcup-engine/internal/types"
)

// RoleEntry 导出角色统一入口定义。
func RoleEntry() player.RoleEntry {
	return player.RoleEntry{
		ID:               "plague_mage",
		Defaults:         ApplyDefaults,
		Choices:          NewChoiceHandler(),
		Skills:           SkillEntries(),
		ChoiceRouteSpecs: ChoiceRouteSpecs(),
		TimingHookSpecs: []player.TimingHookSpec{
			{Timing: player.TimingOnHealResist, Priority: 300, Hook: healResistHook},
			{Timing: player.TimingOnTurnEnd, Priority: 400, Hook: turnEndHook},
		},
	}
}

// ApplyDefaults 初始化角色默认属性。
func ApplyDefaults(player *model.Player) {
	if player == nil {
		return
	}
	player.MaxHeal = 5
}

// SkillEntries 导出角色技能与策略绑定入口。
func SkillEntries() []player.SkillEntry {
	return []player.SkillEntry{
		{ID: "plague_immortal", Handler: &PlagueImmortalHandler{}},
		{ID: "plague_blasphemy", Handler: &PlagueBlasphemyHandler{}},
		{ID: "plague_outbreak", Handler: &PlagueOutbreakHandler{}},
		{
			ID:      "plague_death_touch",
			Handler: &PlagueDeathTouchHandler{},
			Policy: types.SkillPolicy{
				SkipAutoPhaseEnd: true,
			},
		},
		{ID: "plague_toxic_nova", Handler: &PlagueToxicNovaHandler{}},
	}
}

// ChoiceRouteSpecs 导出角色 choice 路由声明。
func ChoiceRouteSpecs() map[string]types.ChoiceRouteSpec {
	return map[string]types.ChoiceRouteSpec{
		"plague_death_touch_cards":    types.ChoiceRouteRole("plague_mage"),
		"plague_death_touch_element":  types.ChoiceRouteRole("plague_mage"),
		"plague_death_touch_target":   types.ChoiceRouteRole("plague_mage"),
		"plague_death_touch_x":        types.ChoiceRouteRole("plague_mage"),
		"plague_death_touch_y":        types.ChoiceRouteRole("plague_mage"),
		"plague_mage_toxic_nova_pick": types.ChoiceRouteRole("plague_mage"),
	}
}
