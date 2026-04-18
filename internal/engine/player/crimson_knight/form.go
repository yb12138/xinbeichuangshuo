package crimson_knight

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

func InHotBloodedForm(p *model.Player) bool {
	return player.HasForm(p, model.FormCrimsonKnightHotBlooded)
}

func EnterHotBloodedForm(p *model.Player) bool {
	return player.SetForm(p, model.FormCrimsonKnightHotBlooded)
}

func LeaveHotBloodedForm(p *model.Player) bool {
	return player.ClearForm(p, model.FormCrimsonKnightHotBlooded)
}
