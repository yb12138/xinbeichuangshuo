// gameflow: 勇者回合钩子。

package hero

import (
	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// ActiveTauntSource 查找对指定玩家生效的挑衅来源。
func ActiveTauntSource(rt engineplayer.ChoiceRuntime, player *model.Player) *model.Player {
	if player == nil {
		return nil
	}
	tauntCard := engineplayer.GetFieldEffectCard(player, model.EffectHeroTaunt)
	if tauntCard == nil {
		ConsumeTauntRestriction(rt, player)
		return nil
	}
	src := rt.GetPlayers()[tauntCard.SourceID]
	if src == nil || src.Camp == player.Camp {
		ConsumeTauntRestriction(rt, player)
		return nil
	}
	return src
}

// ConsumeTauntRestriction 消耗挑衅约束（移除场牌+重置状态）。
func ConsumeTauntRestriction(rt engineplayer.ChoiceRuntime, player *model.Player) {
	if player == nil {
		return
	}
	player.TurnState.UsedSkillCounts["hero_taunt_active_turn"] = 0
	rt.RemoveFieldCard(player.ID, model.EffectHeroTaunt)
}
