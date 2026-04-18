package magic_lancer

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

func InPhantomForm(p *model.Player) bool {
	return player.HasForm(p, model.FormMagicLancerPhantom)
}

func LeavePhantomForm(p *model.Player) bool {
	return player.ClearForm(p, model.FormMagicLancerPhantom)
}
