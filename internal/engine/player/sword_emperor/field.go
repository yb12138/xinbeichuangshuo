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

// SyncSwordSoulToken 将剑魂盖牌数量同步到 player.Tokens（已弃用：派生值实时计算）。
func SyncSwordSoulToken(p *model.Player) {
	// 不再同步到 Tokens，服务端 buildStateForPlayer 实时计算 SwordSoulCount
}
