// gameflow: 仲裁者回合钩子。

package arbiter

import (
	"fmt"

	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// ForcedDoomsdayStartup 回合开始时末日审判强制发动检查。
func ForcedDoomsdayStartup(rt engineplayer.ChoiceRuntime, player *model.Player) bool {
	if player == nil || !engineplayer.IsCharacter(player, "arbiter") {
		return false
	}
	engineplayer.EnsurePlayerTokensMap(player)
	if player.Tokens["judgment"] < 4 || player.TurnState.UsedSkillCounts["arbiter_skip_forced_doomsday"] != 0 || player.TurnState.UsedSkillCounts["arbiter_forced_doomsday_done_turn"] != 0 {
		player.TurnState.UsedSkillCounts["arbiter_forced_doomsday_pending"] = 0
		return false
	}
	if len(rt.CampEnemyIDs(player.Camp)) == 0 {
		player.TurnState.UsedSkillCounts["arbiter_forced_doomsday_pending"] = 0
		return false
	}
	if player.TurnState.UsedSkillCounts["arbiter_forced_doomsday_pending"] == 0 {
		rt.Log(fmt.Sprintf("%s 的审判已达上限：本行动阶段必须发动 [末日审判]", player.Name))
	}
	player.TurnState.UsedSkillCounts["arbiter_forced_doomsday_pending"] = 1
	return false
}
