// gameflow: 圣弓策略 Hook 声明式注册。

package holy_bow

import (
	"fmt"

	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// holyGloryExitHook 圣光荣耀退出策略。
func holyGloryExitHook(rt engineplayer.HookRuntime, ctx engineplayer.TimingHookContext) engineplayer.TimingHookResult {
	player := ctx.Player
	if player == nil {
		return engineplayer.TimingHookResult{}
	}

	// 检查是否为圣弓角色且处于圣煌形态
	if !rt.IsCharacter(player, "holy_bow") || !rt.HasForm(player, model.FormHolyBowHolyGlory) {
		return engineplayer.TimingHookResult{}
	}

	beforePoses := rt.SnapshotPlayerPoses()
	rt.ClearForm(player, model.FormHolyBowHolyGlory)
	rt.Heal(player.ID, 1)
	if player.TurnState.UsedSkillCounts == nil {
		player.TurnState.UsedSkillCounts = map[string]int{}
	}
	player.TurnState.UsedSkillCounts["hb_special"] = 1
	rt.Log(fmt.Sprintf("%s 在圣煌形态下执行特殊行动，脱离圣煌形态并获得1点治疗", player.Name))
	rt.DispatchOrientationChanges(beforePoses)

	return engineplayer.TimingHookResult{Handled: true}
}
