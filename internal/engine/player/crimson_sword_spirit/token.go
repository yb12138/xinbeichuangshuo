package crimson_sword_spirit

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

const BloodCap = 3

func Blood(p *model.Player) int {
	return player.TokenValue(p, "css_blood", BloodCap)
}

func AddBlood(p *model.Player, delta int) int {
	return player.AddToken(p, "css_blood", delta, BloodCap)
}
