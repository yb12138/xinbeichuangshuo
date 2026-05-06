// gameflow: 格斗家 FlowContinuation 处理函数。

package fighter

import (
	"fmt"

	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

func handleFighterAfterDiscard(rt engineplayer.ChoiceRuntime, cont model.FlowContinuation) error {
	if cont.SkillID != "fighter_war_god_drive" {
		return nil
	}
	user := rt.GetPlayers()[cont.PlayerID]
	if user == nil {
		return fmt.Errorf("斗神天驱后续执行者不存在: %s", cont.PlayerID)
	}
	if cont.Data != nil {
		if discarded, ok := cont.Data["discarded_cards"].([]model.Card); ok && len(discarded) > 0 {
			rt.Log(fmt.Sprintf("%s 的 [斗神天驱] 弃置%d张牌", user.Name, len(discarded)))
		}
	}
	rt.Heal(user.ID, 2)
	rt.Log(fmt.Sprintf("%s 的 [斗神天驱] 弃牌完成，获得2点治疗", user.Name))
	return nil
}
