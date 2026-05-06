package beast_samurai

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

func InIaijutsuForm(p *model.Player) bool {
	return player.HasForm(p, model.FormBeastSamuraiIaijutsu)
}

func EnterIaijutsuForm(p *model.Player) bool {
	return player.SetForm(p, model.FormBeastSamuraiIaijutsu)
}

func LeaveIaijutsuForm(p *model.Player) bool {
	return player.ClearForm(p, model.FormBeastSamuraiIaijutsu)
}
