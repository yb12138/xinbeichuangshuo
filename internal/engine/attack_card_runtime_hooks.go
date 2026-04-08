package engine

import "starcup-engine/internal/model"

type attackCardRuntimeTransformHook func(e *GameEngine, player *model.Player, card model.Card) model.Card

func applyBlazeWitchAttackCardRuntimeHook(e *GameEngine, player *model.Player, card model.Card) model.Card {
	if e == nil {
		return card
	}
	return e.applyBlazeWitchAttackCardTransform(player, card)
}

// applyTimingOnAttackDeclaredCardTransforms 在攻击宣言时按固定顺序应用卡面变换规则。
func (e *GameEngine) applyTimingOnAttackDeclaredCardTransforms(player *model.Player, card model.Card) model.Card {
	for _, hook := range e.attackDeclaredCardTransformHooks {
		card = hook(e, player, card)
	}
	return card
}
