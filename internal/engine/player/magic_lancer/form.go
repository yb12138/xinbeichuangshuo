package magic_lancer

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

func InPhantomForm(p *model.Player) bool {
	return player.HasForm(p, model.FormMagicLancerPhantom)
}

func LeavePhantomForm(p *model.Player) bool {
	return player.ClearForm(p, model.FormMagicLancerPhantom)
}

func BlocksMagicCasting(p *model.Player) bool {
	// 魔枪黑暗束缚：始终不能使用法术牌
	return player.IsCharacter(p, "magic_lancer")
}
