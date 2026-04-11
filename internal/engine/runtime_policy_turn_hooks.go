// gameflow: 回合级策略钩子注册。

package engine

import (
	"fmt"

	"starcup-engine/internal/engine/skill"
	"starcup-engine/internal/model"
)

func turnBeforeStartButterflyDancerWitherExpiryHook(e *GameEngine, player *model.Player) bool {
	if !e.isButterflyDancer(player) {
		return false
	}
	ensurePlayerTokensMap(player)
	if player.Tokens["bt_wither_active"] <= 0 {
		return false
	}
	player.Tokens["bt_wither_active"] = 0
	e.Log(fmt.Sprintf("%s 的 [凋零] 效果到期：对方士气下限保护已解除", player.Name))
	return false
}

func turnStartBloodPriestessBleedHook(e *GameEngine, player *model.Player) bool {
	if !e.isBloodPriestess(player) {
		return false
	}
	if !hasBloodPriestessBleedingForm(player) || player.TurnState.UsedSkillCounts["bp_bleed_tick"] > 0 {
		return false
	}
	player.TurnState.UsedSkillCounts["bp_bleed_tick"] = 1
	e.Log(fmt.Sprintf("%s 的 [流血] 生效：回合开始对自己造成1点法术伤害", player.Name))
	e.AddPendingDamage(model.PendingDamage{
		SourceID:   player.ID,
		TargetID:   player.ID,
		Damage:     1,
		DamageType: model.MagicDamage,
	})
	e.enterDamageResolution(model.TurnStageTurnStart)
	return true
}

func beforeActionPoisonHook(e *GameEngine, player *model.Player) bool {
	if player == nil {
		return false
	}
	for _, fc := range player.Field {
		if fc == nil || fc.Mode != model.FieldEffect || fc.Hook != model.FieldHookOnBeforeAction || fc.Effect != model.EffectPoison {
			continue
		}
		allowCrimsonFaithHeal := fc.SourceID != "" && fc.SourceID == player.ID
		e.AddPendingDamage(model.PendingDamage{
			SourceID:              fc.SourceID,
			TargetID:              player.ID,
			Damage:                1,
			DamageType:            "poison",
			AllowCrimsonFaithHeal: allowCrimsonFaithHeal,
		})
		player.RemoveFieldCard(fc)
		e.State.DiscardPile = append(e.State.DiscardPile, fc.Card)
		e.Log(fmt.Sprintf("[Effect] %s 受到中毒伤害", player.Name))
		e.Log(fmt.Sprintf("[Field] %s 面前的【%s】触发效果并被弃置", player.Name, fc.Card.Name))
		e.enterDamageResolution(model.TurnStageBeforeAction)
		return true
	}
	return false
}

func beforeActionFiveElementsBindHook(e *GameEngine, player *model.Player) bool {
	if player == nil {
		return false
	}
	for _, fc := range player.Field {
		if fc == nil || fc.Mode != model.FieldEffect || fc.Hook != model.FieldHookOnBeforeAction || fc.Effect != model.EffectFiveElementsBind {
			continue
		}
		sealCount := 0
		for _, fieldPlayer := range e.GetAllPlayers() {
			if fieldPlayer == nil {
				continue
			}
			for _, fieldCard := range fieldPlayer.Field {
				if fieldCard == nil || fieldCard.Mode != model.FieldEffect || !model.IsElementalSealEffect(fieldCard.Effect) {
					continue
				}
				sealCount++
				if sealCount >= 2 {
					break
				}
			}
			if sealCount >= 2 {
				break
			}
		}
		e.PushInterrupt(&model.Interrupt{
			Type:     model.InterruptChoice,
			PlayerID: player.ID,
			Context: map[string]interface{}{
				"choice_type": "five_elements_bind",
				"player_id":   player.ID,
				"draw_count":  2 + sealCount,
			},
		})
		e.Log(fmt.Sprintf("[Buff] %s 触发五系束缚判定，等待玩家选择...", player.Name))
		return true
	}
	return false
}

func beforeActionWeakHook(e *GameEngine, player *model.Player) bool {
	if player == nil || !player.HasFieldEffect(model.EffectWeak) {
		return false
	}
	e.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: player.ID,
		Context: map[string]interface{}{
			"choice_type": "weak",
		},
	})
	e.Log(fmt.Sprintf("[Buff] %s 触发虚弱判定，等待玩家选择...", player.Name))
	return true
}

func turnStartValkyrieMilitaryGloryHook(e *GameEngine, player *model.Player) bool {
	if e == nil || player == nil || !isCharacter(player, "valkyrie") {
		return false
	}
	ensurePlayerTokensMap(player)
	if player.Tokens["valkyrie_spirit"] <= 0 || player.TurnState.UsedSkillCounts["valkyrie_military_glory"] > 0 {
		return false
	}
	ctx := e.buildTimedContext(player, nil, model.TimingOnTurnStart, &model.EventContext{
		Type:     model.EventTurnStart,
		SourceID: player.ID,
	})
	handler := skills.GetHandler("valkyrie_military_glory")
	if handler == nil || !handler.CanUse(ctx) {
		return false
	}
	player.TurnState.UsedSkillCounts["valkyrie_military_glory"] = 1
	if err := handler.Execute(ctx); err != nil {
		e.Log(fmt.Sprintf("[Skill Error] 军威神光执行失败: %v", err))
		return false
	}
	e.recordSkillUsage(player.ID, "军威神光", model.SkillTypeStartup)
	return e.State.PendingInterrupt != nil
}

func startupHeroExhaustionReleaseHook(e *GameEngine, player *model.Player) bool {
	if e == nil || player == nil || !e.isHero(player) || player.TurnState.HasUsedActionSkill || !hasHeroExhaustionForm(player) {
		return false
	}
	ensurePlayerTokensMap(player)
	if player.Tokens["hero_exhaustion_release_pending"] <= 0 {
		return false
	}
	beforePoses := e.snapshotPlayerPoses()
	leaveHeroExhaustionForm(player)
	player.Tokens["hero_exhaustion_release_pending"] = 0
	e.Log(fmt.Sprintf("%s 的 [精疲力竭] 结束：转正，手牌上限恢复，并对自己造成3点法术伤害", player.Name))
	e.dispatchOrientationChanges(beforePoses)
	e.AddPendingDamage(model.PendingDamage{
		SourceID:   player.ID,
		TargetID:   player.ID,
		Damage:     3,
		DamageType: model.MagicAttack,
	})
	return true
}

func startupArbiterForcedDoomsdayHook(e *GameEngine, player *model.Player) bool {
	if e == nil || player == nil || !isCharacter(player, "arbiter") {
		return false
	}
	ensurePlayerTokensMap(player)
	if player.Tokens["judgment"] < 4 || player.TurnState.UsedSkillCounts["arbiter_skip_forced_doomsday"] != 0 || player.TurnState.UsedSkillCounts["arbiter_forced_doomsday_done_turn"] != 0 {
		player.TurnState.UsedSkillCounts["arbiter_forced_doomsday_pending"] = 0
		return false
	}
	if len(e.campEnemyIDs(player.Camp)) == 0 {
		player.TurnState.UsedSkillCounts["arbiter_forced_doomsday_pending"] = 0
		return false
	}
	if player.TurnState.UsedSkillCounts["arbiter_forced_doomsday_pending"] == 0 {
		e.Log(fmt.Sprintf("%s 的审判已达上限：本行动阶段必须发动 [末日审判]", player.Name))
	}
	player.TurnState.UsedSkillCounts["arbiter_forced_doomsday_pending"] = 1
	return false
}

func startupHeroTauntHook(e *GameEngine, player *model.Player) bool {
	if e == nil || player == nil {
		return false
	}
	player.TurnState.UsedSkillCounts["hero_taunt_active_turn"] = 0
	if player.TurnState.UsedSkillCounts["arbiter_forced_doomsday_pending"] > 0 {
		return false
	}
	src := activeHeroTauntSource(e, player)
	if src == nil {
		return false
	}
	player.TurnState.UsedSkillCounts["hero_taunt_active_turn"] = 1
	e.Log(fmt.Sprintf("[Taunt] %s 在本行动阶段受到 %s 的【挑衅】影响：必须且只能主动攻击该勇者，或选择跳过行动并移除此牌", player.Name, model.GetPlayerDisplayName(src)))
	return false
}

func activeHeroTauntSource(e *GameEngine, player *model.Player) *model.Player {
	if e == nil || player == nil {
		return nil
	}
	tauntCard := getHeroTauntCard(player)
	if tauntCard == nil {
		clearHeroTauntRestriction(e, player)
		return nil
	}
	src := e.State.Players[tauntCard.SourceID]
	if src == nil || src.Camp == player.Camp {
		clearHeroTauntRestriction(e, player)
		return nil
	}
	return src
}

func clearHeroTauntRestriction(e *GameEngine, player *model.Player) {
	consumeHeroTauntRestriction(e, player)
}

func consumeHeroTauntRestriction(e *GameEngine, player *model.Player) {
	if e == nil || player == nil {
		return
	}
	player.TurnState.UsedSkillCounts["hero_taunt_active_turn"] = 0
	e.RemoveFieldCard(player.ID, model.EffectHeroTaunt)
}

func hasPlayableAttackCard(player *model.Player) bool {
	if player == nil {
		return false
	}
	for idx := 0; idx < playableCardCount(player); idx++ {
		card, _, _, ok := getPlayableCardByIndex(player, idx)
		if ok && card.Type == model.CardTypeAttack {
			return true
		}
	}
	return false
}
