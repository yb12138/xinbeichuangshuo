package beast_samurai

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

const ZanshinCap = 4

func Zanshin(p *model.Player) int {
	return player.TokenValue(p, "bs_zanshin", ZanshinCap)
}

func AddZanshin(p *model.Player, delta int) int {
	return player.AddToken(p, "bs_zanshin", delta, ZanshinCap)
}

const BeastSoulCap = 2

func BeastSoul(p *model.Player) int {
	return player.TokenValue(p, "bs_beast_soul", BeastSoulCap)
}

func AddBeastSoul(p *model.Player, delta int, ignoreCap bool) int {
	return player.AddTokenIgnoreCap(p, "bs_beast_soul", delta, BeastSoulCap, ignoreCap)
}

// ClearAttackTokens zeros the reversal-pending token.
func ClearAttackTokens(p *model.Player) {
	player.EnsurePlayerTokensMap(p)
	p.Tokens["bs_reversal_pending_x"] = 0
}
