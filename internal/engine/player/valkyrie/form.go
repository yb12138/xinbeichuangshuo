package valkyrie

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

func InHeroicForm(p *model.Player) bool {
	return player.HasForm(p, model.FormValkyrieHeroic)
}

func EnterHeroicForm(p *model.Player) bool {
	return player.SetForm(p, model.FormValkyrieHeroic)
}

func LeaveHeroicForm(p *model.Player) bool {
	return player.ClearForm(p, model.FormValkyrieHeroic)
}
