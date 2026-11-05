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

	if !rt.IsCharacter(player, "holy_bow") {
		return engineplayer.TimingHookResult{}
	}
	// 自动填充：任意形态下执行过特殊行动（购买/合成/提炼）都记为"已执行特殊行动"
	if player.TurnState.UsedSkillCounts == nil {
		player.TurnState.UsedSkillCounts = map[string]int{}
	}
	player.TurnState.UsedSkillCounts["hb_special"] = 1
	// 仅圣煌形态下脱离形态并+1治疗
	if !rt.HasForm(player, model.FormHolyBowHolyGlory) {
		return engineplayer.TimingHookResult{}
	}

	beforePoses := rt.SnapshotPlayerPoses()
	rt.ClearForm(player, model.FormHolyBowHolyGlory)
	rt.Heal(player.ID, 1)
	rt.Log(fmt.Sprintf("%s 在圣煌形态下执行特殊行动，脱离圣煌形态并获得1点治疗", player.Name))
	rt.DispatchOrientationChanges(beforePoses)

	return engineplayer.TimingHookResult{Handled: true}
}
