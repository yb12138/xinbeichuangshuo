// gameflow: 回合级策略钩子注册（角色逻辑委托到 player/<role>/ 包）。

package engine

import (
	"fmt"

	arbiterpkg "starcup-engine/internal/engine/player/arbiter"
	bloodpriestess "starcup-engine/internal/engine/player/blood_priestess"
	butterflydancer "starcup-engine/internal/engine/player/butterfly_dancer"
	heropkg "starcup-engine/internal/engine/player/hero"
	valkyriepkg "starcup-engine/internal/engine/player/valkyrie"
	"starcup-engine/internal/model"
)

func turnStartBloodPriestessBleedHook(e *GameEngine, player *model.Player) bool {
	return bloodpriestess.BleedTick(newRoleChoiceRuntime(e), player)
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
	return valkyriepkg.MilitaryGlory(newRoleChoiceRuntime(e), player)
}

func startupHeroExhaustionReleaseHook(e *GameEngine, player *model.Player) bool {
	return heropkg.ExhaustionRelease(newRoleChoiceRuntime(e), player)
}

func startupArbiterForcedDoomsdayHook(e *GameEngine, player *model.Player) bool {
	return arbiterpkg.ForcedDoomsdayStartup(newRoleChoiceRuntime(e), player)
}

func startupHeroTauntHook(e *GameEngine, player *model.Player) bool {
	return heropkg.TauntStartup(newRoleChoiceRuntime(e), player)
}

func activeHeroTauntSource(e *GameEngine, player *model.Player) *model.Player {
	return heropkg.ActiveTauntSource(newRoleChoiceRuntime(e), player)
}

func clearHeroTauntRestriction(e *GameEngine, player *model.Player) {
	consumeHeroTauntRestriction(e, player)
}

func consumeHeroTauntRestriction(e *GameEngine, player *model.Player) {
	heropkg.ConsumeTauntRestriction(newRoleChoiceRuntime(e), player)
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

func (e *GameEngine) ResolveButterflyChrysalis(userID string) error {
	return butterflydancer.ResolveChrysalis(newRoleChoiceRuntime(e), userID)
}

func (e *GameEngine) StartButterflyReverse(userID string) error {
	return butterflydancer.StartReverse(newRoleChoiceRuntime(e), userID)
}
