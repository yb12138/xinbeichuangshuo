package bard

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

func InEternalPrisonerForm(p *model.Player) bool {
	return player.HasForm(p, model.FormBardEternalPrisoner)
}

func EnterEternalPrisonerForm(p *model.Player) bool {
	return player.SetForm(p, model.FormBardEternalPrisoner)
}

func LeaveEternalPrisonerForm(p *model.Player) bool {
	return player.ClearForm(p, model.FormBardEternalPrisoner)
}
