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

// postActionEndBleedExitHook 手牌<3 时立即脱离流血形态。
// 挂在状态检查（每次推进流程前）和行动结束两个时点，保证手牌一少于 3 张就脱离。
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
		func() {
			defer rt.PoseChangeGuard()()
			LeaveBleedingForm(player)
		}()
		engineplayer.NotifySkillActivated(rt, player, player.ID, "bp_bleeding", player.ID)
		rt.Log(fmt.Sprintf("%s 的 [流血·手牌不足脱离] 生效：手牌<3，立即脱离流血形态", player.Name))
		// 脱离流血后同生共死的上限修正由 +1 变为 -2，需要重新检查相关玩家的手牌上限（不打断当前回合）。
		crt := rt.AsChoiceRuntime()
		for _, otherID := range rt.GetPlayerOrder() {
			other := rt.GetPlayers()[otherID]
			if other != nil && len(other.Hand) > rt.GetMaxHand(other) && crt != nil {
				crt.CheckHandLimit(other.ID, true)
			}
		}
	}
	return engineplayer.TimingHookResult{}
}
