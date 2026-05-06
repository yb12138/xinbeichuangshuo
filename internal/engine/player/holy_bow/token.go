package holy_bow

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

const FaithCap = 10

func Faith(p *model.Player) int {
	return player.TokenValue(p, "hb_faith", FaithCap)
}

func AddFaith(p *model.Player, delta int) int {
	return player.AddToken(p, "hb_faith", delta, FaithCap)
}

const CannonCap = 1

func Cannon(p *model.Player) int {
	return player.TokenValue(p, "hb_cannon", CannonCap)
}

func AddCannon(p *model.Player, delta int) int {
	return player.AddToken(p, "hb_cannon", delta, CannonCap)
}
