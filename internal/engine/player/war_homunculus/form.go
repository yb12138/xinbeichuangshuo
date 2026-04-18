package war_homunculus

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

func InBurstForm(p *model.Player) bool {
	return player.HasForm(p, model.FormWarHomunculusBurst)
}

func EnterBurstForm(p *model.Player) bool {
	return player.SetForm(p, model.FormWarHomunculusBurst)
}

func LeaveBurstForm(p *model.Player) bool {
	return player.ClearForm(p, model.FormWarHomunculusBurst)
}
