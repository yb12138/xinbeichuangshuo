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

func BlocksMagicCasting(p *model.Player) bool {
	// 魔剑士暗影抗拒：暗影形态下行动阶段不能使用法术牌
	return player.IsCharacter(p, "magic_swordsman") && InShadowForm(p)
}

func CanUseShadowRejectResponse(p *model.Player, currentTurnPlayerID string) bool {
	return player.IsCharacter(p, "magic_swordsman") && p.ID != currentTurnPlayerID
}
