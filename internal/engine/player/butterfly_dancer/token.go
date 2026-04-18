package butterfly_dancer

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

const PupaCap = -1

func Pupa(p *model.Player) int {
	return player.TokenValue(p, "bt_pupa", PupaCap)
}

func AddPupa(p *model.Player, delta int) int {
	return player.AddToken(p, "bt_pupa", delta, PupaCap)
}
