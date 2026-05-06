package blaze_witch

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// AttackElement returns the effective attack element for a blaze witch in flame form.
func AttackElement(p *model.Player, card model.Card) model.Element {
	if p == nil || p.Tokens == nil {
		return card.Element
	}
	if !player.IsCharacter(p, "blaze_witch") || !player.HasForm(p, model.FormBlazeWitchFlame) {
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

// ApplyAttackCardTransform transforms the card element for a blaze witch.
func ApplyAttackCardTransform(p *model.Player, card model.Card) model.Card {
	card.Element = AttackElement(p, card)
	return card
}
