package moon

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

const NewMoonCap = 2

func NewMoon(p *model.Player) int {
	return player.TokenValue(p, "mg_new_moon", NewMoonCap)
}

func AddNewMoon(p *model.Player, delta int) int {
	return player.AddToken(p, "mg_new_moon", delta, NewMoonCap)
}

const PetrifyCap = 3

func Petrify(p *model.Player) int {
	return player.TokenValue(p, "mg_petrify", PetrifyCap)
}

func AddPetrify(p *model.Player, delta int) int {
	return player.AddToken(p, "mg_petrify", delta, PetrifyCap)
}
