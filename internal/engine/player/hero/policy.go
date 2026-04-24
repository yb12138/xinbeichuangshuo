// gameflow: 勇者行动选择策略。

package hero

import (
	"fmt"

	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// TauntOptionPolicy 挑衅行动选项策略。
// hasPlayableAttack reports whether the player has a playable attack card (engine callback).
func TauntOptionPolicy(rt engineplayer.ChoiceRuntime, player *model.Player, mod engineplayer.ActionSelectionModifier, hasPlayableAttack func(*model.Player) bool) {
	if player == nil {
		return
	}
	if player.TurnState.UsedSkillCounts["hero_taunt_active_turn"] <= 0 || player.TurnState.UsedSkillCounts["arbiter_forced_doomsday_pending"] > 0 {
		return
	}
	src := ActiveTauntSource(rt, player)
	if src == nil {
		return
	}
	mod.SetActionRule("force_attack_or_skip", "hero_taunt", 10)
	mod.SetConstrainedTarget(src.ID, model.GetPlayerDisplayName(src))
	mod.SetRuleRequiresSkipOnly(!hasPlayableAttack(player))
}

// TauntValidationPolicy 挑衅行动校验策略。
func TauntValidationPolicy(rt engineplayer.ChoiceRuntime, player *model.Player, mod engineplayer.ActionSelectionValidationModifier, hasPlayableAttack func(*model.Player) bool) {
	if player == nil {
		return
	}
	if player.TurnState.UsedSkillCounts["hero_taunt_active_turn"] <= 0 || player.TurnState.UsedSkillCounts["arbiter_forced_doomsday_pending"] > 0 {
		return
	}
	src := ActiveTauntSource(rt, player)
	if src == nil {
		return
	}
	sourceName := model.GetPlayerDisplayName(src)
	mod.SetActionRule("force_attack_or_skip", "hero_taunt", 10)
	mod.SetConstrainedTarget(src.ID, sourceName)
	mod.SetRuleRequiresSkipOnly(!hasPlayableAttack(player))
	mod.SetForceAttackOnlyMessage(fmt.Sprintf("你受到【挑衅】影响：本次行动阶段必须且只能主动攻击 %s，或选择跳过行动", sourceName))
	mod.SetOnSkipChosen(func(rt engineplayer.ChoiceRuntime, player *model.Player) (bool, error) {
		rt.Log(fmt.Sprintf("[Taunt] %s 选择跳过本次行动阶段，并移除来自 %s 的【挑衅】", player.Name, sourceName))
		ConsumeTauntRestriction(rt, player)
		rt.EnterTurnEndStage()
		return true, nil
	})
	mod.SetOnAttackAccepted(func(rt engineplayer.ChoiceRuntime, player *model.Player, act model.PlayerAction) error {
		// 标记攻击后消耗挑衅效果
		mod.MarkConsumeHeroTauntOnAttack()
		return nil
	})
}

// HasPlayableAttackCard 检查玩家是否有可打出的攻击牌。
func HasPlayableAttackCard(player *model.Player) bool {
	if player == nil {
		return false
	}
	for _, c := range player.Hand {
		if c.Type == model.CardTypeAttack {
			return true
		}
	}
	return false
}
