// gameflow: 魔法弓手：充能计数与盖牌辅助。

package engine

import (
	"starcup-engine/internal/model"
)

func magicBowChargeCount(player *model.Player, element model.Element) int {
	return coverCountByEffectAndElement(player, model.EffectMagicBowCharge, element)
}

func syncMagicBowChargeToken(player *model.Player) {
	// no-op: mb_charge_count 在服务端 buildStateForPlayer 中按场上盖牌派生写入 PlayerView.tokens
}

func addMagicBowChargeCards(player *model.Player, cards []model.Card) int {
	if player == nil || len(cards) == 0 {
		return 0
	}
	room := magicBowChargeCapEngine - magicBowChargeCount(player, "")
	if room <= 0 {
		return 0
	}
	added := 0
	for _, c := range cards {
		if added >= room {
			break
		}
		player.AddFieldCard(&model.FieldCard{
			Card:     c,
			OwnerID:  player.ID,
			SourceID: player.ID,
			Mode:     model.FieldCover,
			Effect:   model.EffectMagicBowCharge,
		})
		added++
	}
	syncMagicBowChargeToken(player)
	return added
}

func removeMagicBowChargeByElement(player *model.Player, element model.Element) (model.Card, bool) {
	card, ok := removeFirstCoverByEffectAndElement(player, model.EffectMagicBowCharge, element)
	if !ok {
		return model.Card{}, false
	}
	syncMagicBowChargeToken(player)
	return card, true
}
