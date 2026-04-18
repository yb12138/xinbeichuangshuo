package soul_sorcerer

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

const BlueSoulCap = 6

func BlueSoul(p *model.Player) int {
	return player.TokenValue(p, "ss_blue_soul", BlueSoulCap)
}

func AddBlueSoul(p *model.Player, delta int) int {
	return player.AddToken(p, "ss_blue_soul", delta, BlueSoulCap)
}

const YellowSoulCap = 6

func YellowSoul(p *model.Player) int {
	return player.TokenValue(p, "ss_yellow_soul", YellowSoulCap)
}

func AddYellowSoul(p *model.Player, delta int) int {
	return player.AddToken(p, "ss_yellow_soul", delta, YellowSoulCap)
}
