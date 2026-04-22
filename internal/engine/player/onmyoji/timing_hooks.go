// gameflow: 阴阳师 Timing Hook 实现。

package onmyoji

import (
	"fmt"

	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// turnEndDarkRitualHook 回合结束时黑暗祭礼触发检查。
func turnEndDarkRitualHook(rt engineplayer.HookRuntime, ctx engineplayer.TimingHookContext) engineplayer.TimingHookResult {
	player := rt.GetPlayer(ctx.SourceID)
	if player == nil || !engineplayer.IsCharacter(player, "onmyoji") || player.Tokens == nil || player.Tokens["onmyoji_ghost_fire"] < 3 {
		return engineplayer.TimingHookResult{}
	}
	targetIDs := rt.CampEnemyIDs(player.Camp)
	if len(targetIDs) == 0 {
		return engineplayer.TimingHookResult{}
	}
	rt.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: player.ID,
		Context: map[string]interface{}{
			"choice_type": "onmyoji_dark_ritual_target",
			"user_id":     player.ID,
			"target_ids":  targetIDs,
			"ghost_fire":  player.Tokens["onmyoji_ghost_fire"],
		},
	})
	rt.Log(fmt.Sprintf("%s 的 [黑暗祭礼] 触发，等待选择2点法术伤害目标", player.Name))
	return engineplayer.TimingHookResult{Interrupted: true}
}
