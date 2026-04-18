package magic_swordsman

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

func InShadowForm(p *model.Player) bool {
	return player.HasForm(p, model.FormMagicSwordsmanShadow)
}

func EnterShadowForm(p *model.Player) bool {
	return player.SetForm(p, model.FormMagicSwordsmanShadow)
}

func LeaveShadowForm(p *model.Player) bool {
	return player.ClearForm(p, model.FormMagicSwordsmanShadow)
}
