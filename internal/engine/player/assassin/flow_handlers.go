// gameflow: 暗杀者 FlowContinuation 处理函数。

package assassin

import (
	"fmt"

	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

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

// handleStealthAfterDraw 处理 after_draw 流程边界：摸牌后进入潜行。
func handleStealthAfterDraw(rt engineplayer.ChoiceRuntime, cont model.FlowContinuation) error {
	user := rt.LookupPlayer(cont.PlayerID)
	if user == nil {
		return fmt.Errorf("暗杀者潜行后续执行者不存在: %s", cont.PlayerID)
	}
	if !engineplayer.IsCharacter(user, "assassin") {
		return fmt.Errorf("仅暗杀者可执行潜行后续")
	}
	applyStealth(rt, user)
	return nil
}
