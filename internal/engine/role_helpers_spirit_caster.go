// gameflow: 灵力施法者：灵力盖牌与计数辅助。
package engine

import (
	"fmt"

	"starcup-engine/internal/engine/core/runtimeutil"
	"starcup-engine/internal/model"
)

func spiritCasterPowerCovers(player *model.Player) []*model.FieldCard {
	return coverCardsByEffect(player, model.EffectSpiritCasterPower)
}

func spiritCasterPowerCount(player *model.Player, element model.Element) int {
	return coverCountByEffectAndElement(player, model.EffectSpiritCasterPower, element)
}

func syncSpiritCasterPowerToken(player *model.Player) {
	// no-op: sc_power_count 在服务端 buildStateForPlayer 中按场上盖牌派生写入 PlayerView.tokens
}

func addSpiritCasterPowerCard(player *model.Player, card model.Card) bool {
	if player == nil {
		return false
	}
	player.AddFieldCard(&model.FieldCard{
		Card:     card,
		OwnerID:  player.ID,
		SourceID: player.ID,
		Mode:     model.FieldCover,
		Effect:   model.EffectSpiritCasterPower,
	})
	syncSpiritCasterPowerToken(player)
	return true
}

// gameflow: 通灵师：念咒、百鬼夜行等。

func (e *GameEngine) continueSpiritCasterTalisman(user *model.Player, skillID string, targetIDs []string) error {
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	switch skillID {
	case "sc_talisman_thunder":
		if e.CanPayCrystalCost(user.ID, 1) {
			e.PushInterrupt(&model.Interrupt{
				Type:     model.InterruptChoice,
				PlayerID: user.ID,
				Context: map[string]interface{}{
					"choice_type": "sc_spiritual_collapse_confirm",
					"user_id":     user.ID,
					"mode":        "sc_talisman_thunder",
					"target_ids":  append([]string{}, targetIDs...),
				},
			})
			return nil
		}
		e.resolveSpiritCasterThunderDamage(user, targetIDs, 0)
	case "sc_talisman_wind":
		return e.startSpiritCasterWindDiscardFlow(user, targetIDs)
	default:
		return fmt.Errorf("未知灵符技能: %s", skillID)
	}
	return nil
}

func (e *GameEngine) resolveSpiritCasterThunderDamage(user *model.Player, targetIDs []string, bonus int) {
	if user == nil {
		return
	}
	damage := 1 + bonus
	if damage < 0 {
		damage = 0
	}
	targetSet := runtimeutil.IDsToSet(runtimeutil.DedupeIDs(targetIDs))
	ordered := e.reverseOrderTargetIDsFrom(user.ID, true)
	hitCount := 0
	for _, targetID := range ordered {
		if !targetSet[targetID] {
			continue
		}
		e.AddPendingDamage(model.PendingDamage{
			SourceID:   user.ID,
			TargetID:   targetID,
			Damage:     damage,
			DamageType: model.MagicAttack,
		})
		hitCount++
	}
	e.Log(fmt.Sprintf("%s 发动 [灵符-雷鸣]：对%d名角色各造成%d点法术伤害", user.Name, hitCount, damage))
	e.routePendingDamageWithDefaultReturn(model.TurnStageExtraAction)
}

func (e *GameEngine) startSpiritCasterWindDiscardFlow(user *model.Player, targetIDs []string) error {
	if user == nil {
		return fmt.Errorf("玩家不存在")
	}
	targetSet := runtimeutil.IDsToSet(runtimeutil.DedupeIDs(targetIDs))
	orderedAll := e.reverseOrderTargetIDsFrom(user.ID, true)
	ordered := make([]string, 0, len(targetIDs))
	for _, playerID := range orderedAll {
		if !targetSet[playerID] {
			continue
		}
		ordered = append(ordered, playerID)
	}
	if len(ordered) == 0 {
		e.Log(fmt.Sprintf("%s 的 [灵符-风行]：无有效目标", user.Name))
		return nil
	}

	cursor := 0
	for cursor < len(ordered) {
		target := e.State.Players[ordered[cursor]]
		if target == nil || len(target.Hand) == 0 {
			if target != nil {
				e.Log(fmt.Sprintf("%s 的 [灵符-风行]：%s 无手牌可弃置", user.Name, target.Name))
			}
			cursor++
			continue
		}
		break
	}
	if cursor >= len(ordered) {
		e.Log(fmt.Sprintf("%s 的 [灵符-风行]：所有目标均无手牌可弃置", user.Name))
		return nil
	}

	currentTargetID := ordered[cursor]
	e.PushInterrupt(&model.Interrupt{
		Type:     model.InterruptChoice,
		PlayerID: currentTargetID,
		Context: map[string]interface{}{
			"choice_type":        "sc_talisman_wind_discard",
			"user_id":            user.ID,
			"ordered_target_ids": ordered,
			"cursor":             cursor,
			"current_target_id":  currentTargetID,
		},
	})
	return nil
}
