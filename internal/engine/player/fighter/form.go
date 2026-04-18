package fighter

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

func InHundredDragonForm(p *model.Player) bool {
	return player.HasForm(p, model.FormFighterHundredDragon)
}

func LeaveHundredDragonForm(p *model.Player) bool {
	return player.ClearForm(p, model.FormFighterHundredDragon)
}
