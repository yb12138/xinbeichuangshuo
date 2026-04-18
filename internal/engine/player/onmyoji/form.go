package onmyoji

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

func InShikigamiForm(p *model.Player) bool {
	return player.HasForm(p, model.FormOnmyojiShikigami)
}

func LeaveShikigamiForm(p *model.Player) bool {
	return player.ClearForm(p, model.FormOnmyojiShikigami)
}
