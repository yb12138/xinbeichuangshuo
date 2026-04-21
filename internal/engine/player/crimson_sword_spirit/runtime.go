package crimson_sword_spirit

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// IsRoseCourtyardActive checks whether any crimson sword spirit player has an
// active Rose Courtyard field effect.
func IsRoseCourtyardActive(allPlayers map[string]*model.Player) bool {
	for _, p := range allPlayers {
		if p == nil || !player.IsCharacter(p, "crimson_sword_spirit") {
			continue
		}
		for _, fc := range p.Field {
			if fc != nil && fc.Mode == model.FieldEffect && fc.Effect == model.EffectRoseCourtyard {
				return true
			}
		}
	}
	return false
}
