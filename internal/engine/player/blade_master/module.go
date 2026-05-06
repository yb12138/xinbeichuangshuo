// gameflow: 风之剑圣模块入口声明。

package blade_master

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// RoleEntry 导出角色统一入口定义。
func RoleEntry() player.RoleEntry {
	return player.RoleEntry{
		ID:      "blade_master",
		Choices: NewChoiceHandler(),
		Skills:  SkillEntries(),
		InterruptSpecs: []player.InterruptSpec{
			{
				Type:                 model.InterruptHolySwordDraw,
				PhaseSync:            player.InterruptPhaseSyncCombatDraw,
				BuildPrompt:          buildHolySwordDrawPrompt,
				HandleActionResult:   handleHolySwordDrawAction,
				AllowedActionTypes:   []model.PlayerActionType{model.CmdSelect},
				InvalidActionMessage: "当前为【圣剑】后续选择阶段，请提交选择",
			},
		},
		TimingHookSpecs: []player.TimingHookSpec{
			{Timing: player.TimingPostActionEnd, Priority: 100, Hook: MaybeHolySwordDrawInterrupt},
		},
	}
}

// SkillEntries 导出角色技能与策略绑定入口。
func SkillEntries() []player.SkillEntry {
	return []player.SkillEntry{
		{ID: "wind_fury", Handler: &WindFuryHandler{}},
		{ID: "holy_sword", Handler: &HolySwordHandler{}},
		{ID: "sword_shadow", Handler: &SwordShadowHandler{}},
		{ID: "gale_skill", Handler: &GaleSkillHandler{}},
		{ID: "gale_slash", Handler: &GaleSlashHandler{}},
	}
}
