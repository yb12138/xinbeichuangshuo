// gameflow: 魔法剑士：暗影形态释放辅助。
package engine

import (
	"fmt"

	"starcup-engine/internal/model"
)

func (e *GameEngine) maybeReleaseMagicSwordsmanShadowAtActionStart(player *model.Player) bool {
	if e == nil || player == nil || !e.isMagicSwordsman(player) {
		return false
	}
	ensurePlayerTokensMap(player)
	if player.TurnState.HasUsedActionSkill {
		return false
	}
	if !hasMagicSwordsmanShadowForm(player) {
		return false
	}
	leaveMagicSwordsmanShadowForm(player)
	e.Log(fmt.Sprintf("%s 脱离暗影形态并转正", player.Name))
	return true
}
