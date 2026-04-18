package bard

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

const InspirationCap = 3

func Inspiration(p *model.Player) int {
	return player.TokenValue(p, "bd_inspiration", InspirationCap)
}

func AddInspiration(p *model.Player, delta int) int {
	return player.AddToken(p, "bd_inspiration", delta, InspirationCap)
}
