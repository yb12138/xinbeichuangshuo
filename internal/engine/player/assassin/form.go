package assassin

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

func InStealthForm(p *model.Player) bool {
	return player.HasForm(p, model.FormAssassinStealth)
}

func EnterStealthForm(p *model.Player) bool {
	return player.SetForm(p, model.FormAssassinStealth)
}

func LeaveStealthForm(p *model.Player) bool {
	return player.ClearForm(p, model.FormAssassinStealth)
}
