// gameflow: 剑帝场牌（FieldCard）管理。

package sword_emperor

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// SwordSoulCap 剑魂盖牌上限。
const SwordSoulCap = 3

// SwordSoulCards 按场上顺序收集剑魂盖牌。
func SwordSoulCards(p *model.Player) []*model.FieldCard {
	return player.CoverCardsByEffect(p, model.EffectSwordSoul)
}

// SwordSoulCount 统计剑魂盖牌数量。
func SwordSoulCount(p *model.Player) int {
	return player.CoverCountByEffect(p, model.EffectSwordSoul)
}

// SyncSwordSoulToken 将剑魂盖牌数量同步到 player.Tokens。
func SyncSwordSoulToken(p *model.Player) {
	player.EnsurePlayerTokensMap(p)
	p.Tokens["se_sword_soul_count"] = SwordSoulCount(p)
}
