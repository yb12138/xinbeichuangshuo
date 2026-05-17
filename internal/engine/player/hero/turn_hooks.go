// gameflow: 勇者回合钩子。

package hero

import (
	"fmt"

	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// ExhaustionRelease 回合开始前检查精疲力竭结束。
func ExhaustionRelease(rt engineplayer.ChoiceRuntime, player *model.Player) bool {
	if player == nil || !engineplayer.IsCharacter(player, "hero") || player.TurnState.HasUsedActionSkill || !InExhaustionForm(player) {
		return false
	}
	if !hasExhaustionReleasePending(player) {
		return false
	}
	defer rt.PoseChangeGuard()
	LeaveExhaustionForm(player)
	clearExhaustionReleasePending(player)
	rt.Log(fmt.Sprintf("%s 的 [精疲力竭] 结束：转正，手牌上限恢复，并对自己造成3点法术伤害", player.Name))
	rt.AddPendingDamage(model.PendingDamage{
		SourceID:   player.ID,
		TargetID:   player.ID,
		Damage:     3,
		DamageType: model.MagicAttack,
	})
	return true
}

// TauntStartup 回合开始时挑衅约束初始化。
func TauntStartup(rt engineplayer.ChoiceRuntime, player *model.Player) bool {
	if player == nil {
		return false
	}
	player.TurnState.UsedSkillCounts["hero_taunt_active_turn"] = 0
	if player.TurnState.UsedSkillCounts["arbiter_forced_doomsday_pending"] > 0 {
		return false
	}
	src := ActiveTauntSource(rt, player)
	if src == nil {
		return false
	}
	player.TurnState.UsedSkillCounts["hero_taunt_active_turn"] = 1
	rt.Log(fmt.Sprintf("[Taunt] %s 在本行动阶段受到 %s 的【挑衅】影响：必须且只能主动攻击该勇者，或选择跳过行动并移除此牌", player.Name, model.GetPlayerDisplayName(src)))
	return false
}

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
