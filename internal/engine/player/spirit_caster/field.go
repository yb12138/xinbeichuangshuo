// gameflow: 灵符师灵力（FieldCard）管理。

package spirit_caster

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

// PowerCovers 按效果收集灵力盖牌。
func PowerCovers(p *model.Player) []*model.FieldCard {
	return player.CoverCardsByEffect(p, model.EffectSpiritCasterPower)
}

// PowerCount 统计灵力盖牌数量。element 为空时统计全部。
func PowerCount(p *model.Player, element model.Element) int {
	if element == "" {
		return player.CoverCountByEffect(p, model.EffectSpiritCasterPower)
	}
	return player.CoverCountByEffectAndElement(p, model.EffectSpiritCasterPower, element)
}

// SyncPowerToken 将灵力数量同步到玩家 Token（已弃用：派生值实时计算）。
func SyncPowerToken(p *model.Player) {
	// 不再同步到 Tokens，服务端 buildStateForPlayer 实时计算 PowerCount
}

// AddPowerCard 将卡牌作为灵力盖牌加入玩家场区，并同步 Token。
func AddPowerCard(p *model.Player, card model.Card) {
	p.AddFieldCard(&model.FieldCard{
		Card:     card,
		Mode:     model.FieldCover,
		Effect:   model.EffectSpiritCasterPower,
		Hook:     model.FieldHookManual,
		OwnerID:  p.ID,
		SourceID: p.ID,
	})
	SyncPowerToken(p)
}
