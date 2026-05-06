package elf_archer

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

func InRitualForm(p *model.Player) bool {
	return player.HasForm(p, model.FormElfArcherRitual)
}

func EnterRitualForm(p *model.Player) bool {
	return player.SetForm(p, model.FormElfArcherRitual)
}

func LeaveRitualForm(p *model.Player) bool {
	return player.ClearForm(p, model.FormElfArcherRitual)
}
