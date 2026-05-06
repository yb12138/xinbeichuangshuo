// gameflow: 魔弓充能（FieldCard）管理。

package magic_bow

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

const ChargeCap = 8

// ChargeCount 统计魔弓充能盖牌数量。element 为空时统计全部。
func ChargeCount(p *model.Player, element model.Element) int {
	if element == "" {
		return player.CoverCountByEffect(p, model.EffectMagicBowCharge)
	}
	return player.CoverCountByEffectAndElement(p, model.EffectMagicBowCharge, element)
}

// SyncChargeToken 将充能数量同步到玩家 Token（已弃用：派生值实时计算）。
func SyncChargeToken(p *model.Player) {
	// 不再同步到 Tokens，服务端 buildStateForPlayer 实时计算 ChargeCount
}

// AddChargeCards 将卡牌作为充能盖牌加入玩家场区，并同步 Token。
func AddChargeCards(p *model.Player, cards []model.Card) {
	for _, c := range cards {
		p.AddFieldCard(&model.FieldCard{
			Card:     c,
			Mode:     model.FieldCover,
			Effect:   model.EffectMagicBowCharge,
			Hook:     model.FieldHookManual,
			OwnerID:  p.ID,
			SourceID: p.ID,
		})
	}
	SyncChargeToken(p)
}

// RemoveChargeByElement 按元素移除第一张匹配的充能盖牌，并同步 Token。
func RemoveChargeByElement(p *model.Player, element model.Element) (model.Card, bool) {
	card, ok := player.RemoveFirstCoverByEffectAndElement(p, model.EffectMagicBowCharge, element)
	if ok {
		SyncChargeToken(p)
	}
	return card, ok
}
