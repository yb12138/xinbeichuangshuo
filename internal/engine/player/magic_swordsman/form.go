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
	// 魔剑士暗影抗拒：在你的行动阶段你始终不能使用法术牌（与形态无关）。
	// 回合外仍可用魔弹/圣光响应，由 CanUseShadowRejectResponse 与响应流程处理。
	return player.IsCharacter(p, "magic_swordsman")
}

func CanUseShadowRejectResponse(p *model.Player, currentTurnPlayerID string) bool {
	return player.IsCharacter(p, "magic_swordsman") && p.ID != currentTurnPlayerID
}
