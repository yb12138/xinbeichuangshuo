package engine

import (
	"starcup-engine/internal/model"
)

func getHeroTauntCard(player *model.Player) *model.FieldCard {
	return getFieldEffectCard(player, model.EffectHeroTaunt)
}
