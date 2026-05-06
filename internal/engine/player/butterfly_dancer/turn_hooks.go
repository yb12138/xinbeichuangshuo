// gameflow: 蝶舞者回合钩子。

package butterfly_dancer

import (
	"fmt"

	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// WitherExpiry 回合开始前检查凋零效果到期。
func WitherExpiry(rt engineplayer.ChoiceRuntime, player *model.Player) bool {
	if player == nil || !engineplayer.IsCharacter(player, "butterfly_dancer") {
		return false
	}
	engineplayer.EnsurePlayerSkillFlowState(player)
	if player.TurnState.SkillFlowState["bt_wither_active"] <= 0 {
		return false
	}
	player.TurnState.SkillFlowState["bt_wither_active"] = 0
	rt.Log(fmt.Sprintf("%s 的 [凋零] 效果到期：对方士气下限保护已解除", player.Name))
	return false
}
