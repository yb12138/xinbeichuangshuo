package moon

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

func EnterDarkMoonForm(p *model.Player) bool {
	return player.SetForm(p, model.FormMoonGoddessDarkMoon)
}

func LeaveDarkMoonForm(p *model.Player) bool {
	return player.ClearForm(p, model.FormMoonGoddessDarkMoon)
}
