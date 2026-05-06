package hero

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

func InExhaustionForm(p *model.Player) bool {
	return player.HasForm(p, model.FormHeroExhaustion)
}

func LeaveExhaustionForm(p *model.Player) bool {
	return player.ClearForm(p, model.FormHeroExhaustion)
}
