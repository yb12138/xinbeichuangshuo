package holy_bow

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

func InHolyGloryForm(p *model.Player) bool {
	return player.HasForm(p, model.FormHolyBowHolyGlory)
}

func EnterHolyGloryForm(p *model.Player) bool {
	return player.SetForm(p, model.FormHolyBowHolyGlory)
}

func LeaveHolyGloryForm(p *model.Player) bool {
	return player.ClearForm(p, model.FormHolyBowHolyGlory)
}
