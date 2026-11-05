package moon

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// NewMoonCap 新月上限：与选择流实际生效的上限（choices.go moonGoddessNewMoonCapEngine）一致。
const NewMoonCap = moonGoddessNewMoonCapEngine

func NewMoon(p *model.Player) int {
	return player.TokenValue(p, "mg_new_moon", NewMoonCap)
}

func AddNewMoon(p *model.Player, delta int) int {
	return player.AddToken(p, "mg_new_moon", delta, NewMoonCap)
}

// PetrifyCap 石化上限：与选择流实际生效的上限一致。
const PetrifyCap = moonGoddessPetrifyCapEngine

func Petrify(p *model.Player) int {
	return player.TokenValue(p, "mg_petrify", PetrifyCap)
}

func AddPetrify(p *model.Player, delta int) int {
	return player.AddToken(p, "mg_petrify", delta, PetrifyCap)
}
