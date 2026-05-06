package blood_priestess

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

func InBleedingForm(p *model.Player) bool {
	return player.HasForm(p, model.FormBloodPriestessBleeding)
}

func EnterBleedingForm(p *model.Player) bool {
	return player.SetForm(p, model.FormBloodPriestessBleeding)
}

func LeaveBleedingForm(p *model.Player) bool {
	return player.ClearForm(p, model.FormBloodPriestessBleeding)
}
