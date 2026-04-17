// gameflow: 格斗家模块入口声明。

package fighter

import (
	"starcup-engine/internal/engine/player"
	skills "starcup-engine/internal/engine/skill"
	"starcup-engine/internal/types"
)

// SkillEntries 导出角色技能与策略绑定入口。
func SkillEntries() []player.SkillEntry {
	return []player.SkillEntry{
		{ID: "fighter_psi_field", Handler: &skills.FighterPsiFieldHandler{}},
		{ID: "fighter_charge_strike", Handler: &skills.FighterChargeStrikeHandler{}},
		{ID: "fighter_psi_bullet", Handler: &skills.FighterPsiBulletHandler{}},
		{ID: "fighter_hundred_dragon", Handler: &skills.FighterHundredDragonHandler{}},
		{ID: "fighter_burst_crash", Handler: &skills.FighterBurstCrashHandler{}},
		{ID: "fighter_war_god_drive", Handler: &skills.FighterWarGodDriveHandler{}},
		{ID: "fighter_war_god_drive_followup", Handler: &skills.FighterWarGodDriveFollowupHandler{}},
	}
}

// ChoiceRouteSpecs 导出角色 choice 路由声明。
func ChoiceRouteSpecs() map[string]types.ChoiceRouteSpec {
	return map[string]types.ChoiceRouteSpec{
		"fighter_hundred_dragon_target": types.ChoiceRouteTargetPrompt("fighter"),
		"fighter_psi_bullet_target":     types.ChoiceRouteTargetPrompt("fighter"),
	}
}
