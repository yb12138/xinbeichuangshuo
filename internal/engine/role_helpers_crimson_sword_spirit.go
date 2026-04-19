// gameflow: 绯剑灵专属 helpers。

package engine

import (
	"starcup-engine/internal/model"
)

func (e *GameEngine) isRoseCourtyardActive() bool {
	for _, p := range e.State.Players {
		if p == nil || !e.isCrimsonSwordSpirit(p) {
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
