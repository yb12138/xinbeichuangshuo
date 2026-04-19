// gameflow: 烈焰魔女：攻击元素转换等辅助函数。

package engine

import (
	"starcup-engine/internal/model"
)

func (e *GameEngine) blazeWitchAttackElement(player *model.Player, card model.Card) model.Element {
	if player == nil || player.Tokens == nil {
		return card.Element
	}
	if !e.isBlazeWitch(player) || !hasBlazeWitchFlameForm(player) {
		return card.Element
	}
	if card.Type != model.CardTypeAttack {
		return card.Element
	}
	if card.Element == model.ElementWater || card.Element == model.ElementDark {
		return card.Element
	}
	return model.ElementFire
}

func (e *GameEngine) applyBlazeWitchAttackCardTransform(player *model.Player, card model.Card) model.Card {
	card.Element = e.blazeWitchAttackElement(player, card)
	return card
}
