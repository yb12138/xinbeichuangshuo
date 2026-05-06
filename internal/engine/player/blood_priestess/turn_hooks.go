// gameflow: 血祭司回合钩子。

package blood_priestess

import (
	"fmt"

	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// BleedTick 回合开始时流血效果触发。
func BleedTick(rt engineplayer.ChoiceRuntime, player *model.Player) bool {
	if player == nil || !engineplayer.IsCharacter(player, "blood_priestess") {
		return false
	}
	if !InBleedingForm(player) || player.TurnState.UsedSkillCounts["bp_bleed_tick"] > 0 {
		return false
	}
	player.TurnState.UsedSkillCounts["bp_bleed_tick"] = 1
	rt.Log(fmt.Sprintf("%s 的 [流血] 生效：回合开始对自己造成1点法术伤害", player.Name))
	rt.AddPendingDamage(model.PendingDamage{
		SourceID:   player.ID,
		TargetID:   player.ID,
		Damage:     1,
		DamageType: model.MagicDamage,
	})
	rt.EnterDamageResolution(model.TurnStageTurnStart)
	return true
}

// postActionEndBleedExitHook 行动结束后检查是否需要脱离流血形态。
func postActionEndBleedExitHook(rt engineplayer.HookRuntime, ctx engineplayer.TimingHookContext) engineplayer.TimingHookResult {
	for _, pid := range rt.GetPlayerOrder() {
		player := rt.GetPlayers()[pid]
		if player == nil || !engineplayer.IsCharacter(player, "blood_priestess") {
			continue
		}
		if !InBleedingForm(player) {
			continue
		}
		if len(player.Hand) >= 3 {
			continue
		}
		defer rt.PoseChangeGuard()
		LeaveBleedingForm(player)
		rt.Log(fmt.Sprintf("%s 的 [流血·手牌不足脱离] 生效：行动结束时手牌<3，脱离流血形态", player.Name))
	}
	return engineplayer.TimingHookResult{}
}
