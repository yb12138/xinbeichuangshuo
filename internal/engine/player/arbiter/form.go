package arbiter

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

func InJudgmentForm(p *model.Player) bool {
	return player.HasForm(p, model.FormArbiterJudgment)
}

func EnterJudgmentForm(p *model.Player) bool {
	return player.SetForm(p, model.FormArbiterJudgment)
}
