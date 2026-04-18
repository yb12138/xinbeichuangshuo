package blaze_witch

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

func InFlameForm(p *model.Player) bool {
	return player.HasForm(p, model.FormBlazeWitchFlame)
}

func LeaveFlameForm(p *model.Player) bool {
	return player.ClearForm(p, model.FormBlazeWitchFlame)
}
