// gameflow: 暗杀者延迟后续处理。

package assassin

import (
	"fmt"

	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// FollowupSpecs 导出角色延迟后续声明。
func FollowupSpecs() map[string]engineplayer.FollowupSpec {
	return map[string]engineplayer.FollowupSpec{
		"assassin_stealth_apply": {Label: "Assassin", Resolve: resolveStealthApply},
	}
}

func resolveStealthApply(rt engineplayer.ChoiceRuntime, f model.DeferredFollowup) error {
	user := rt.GetPlayers()[f.UserID]
	if user == nil {
		return fmt.Errorf("暗杀者潜行后续执行者不存在: %s", f.UserID)
	}
	if !engineplayer.IsCharacter(user, "assassin") {
		return fmt.Errorf("仅暗杀者可执行潜行后续")
	}
	applyStealth(rt, user)
	return nil
}

// applyStealth 本地实现进潜行形态：设形态、横置、检查手牌上限。
func applyStealth(rt engineplayer.ChoiceRuntime, player *model.Player) {
	if player == nil {
		return
	}
	defer rt.PoseChangeGuard()
	engineplayer.SetForm(player, model.FormAssassinStealth)
	rt.Log(fmt.Sprintf("%s 进入潜行形态：转为横置，手牌上限-1，无法成为主动攻击目标", player.Name))
	rt.CheckHandLimit(player.ID, true)
}
