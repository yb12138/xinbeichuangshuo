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

// PowerCount 统计灵力盖牌数量。
func PowerCount(p *model.Player) int {
	return player.CoverCountByEffect(p, model.EffectSpiritCasterPower)
}

// SyncPowerToken 将灵力数量同步到玩家 Token。
func SyncPowerToken(p *model.Player) {
	player.EnsurePlayerTokensMap(p)
	p.Tokens["sc_power_count"] = PowerCount(p)
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
