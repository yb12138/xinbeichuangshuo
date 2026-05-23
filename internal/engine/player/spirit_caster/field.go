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

// SyncPowerToken 清理旧版灵力 Token 镜像；妖力数量由场上盖牌实时派生。
func SyncPowerToken(p *model.Player) {
	if p == nil || p.Tokens == nil {
		return
	}
	delete(p.Tokens, "sc_power_count")
}

// AddPowerCard 将卡牌作为灵力盖牌加入玩家场区。
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
