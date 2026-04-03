package engine

import "starcup-engine/internal/model"

type attackCardRuntimeTransformHook func(e *GameEngine, player *model.Player, card model.Card) model.Card

var attackCardRuntimeTransformHooks = []attackCardRuntimeTransformHook{
	applyBlazeWitchAttackCardRuntimeHook,
}

func applyBlazeWitchAttackCardRuntimeHook(e *GameEngine, player *model.Player, card model.Card) model.Card {
	if e == nil {
		return card
	}
	return e.applyBlazeWitchAttackCardTransform(player, card)
}

func (e *GameEngine) applyAttackCardRuntimeTransforms(player *model.Player, card model.Card) model.Card {
	for _, hook := range attackCardRuntimeTransformHooks {
		if hook != nil {
			card = hook(e, player, card)
		}
	}
	return card
}
