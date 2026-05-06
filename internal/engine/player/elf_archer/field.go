// gameflow: 精灵射手祝福（FieldCard）管理。

package elf_archer

import (
	"strings"

	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

const elfBlessingPrefix = "elf_blessing:"

// CountBlessings 统计祝福盖牌数量。
func CountBlessings(p *model.Player) int {
	return player.CoverCountByEffect(p, model.EffectElfBlessing)
}

// BlessingCovers 按效果收集祝福盖牌。
func BlessingCovers(p *model.Player) []*model.FieldCard {
	return player.CoverCardsByEffect(p, model.EffectElfBlessing)
}

// BlessingCards 收集祝福盖牌中的原始卡牌。
func BlessingCards(p *model.Player) []model.Card {
	covers := BlessingCovers(p)
	if len(covers) == 0 {
		return nil
	}
	out := make([]model.Card, 0, len(covers))
	for _, fc := range covers {
		if fc == nil {
			continue
		}
		out = append(out, fc.Card)
	}
	return out
}

// IsBlessingCard 检查卡牌是否为祝福卡牌（通过遍历玩家场上 EffectElfBlessing 盖牌判定）。
func IsBlessingCard(p *model.Player, c model.Card) bool {
	if p == nil {
		return false
	}
	for _, fc := range p.Field {
		if fc == nil || fc.Mode != model.FieldCover || fc.Effect != model.EffectElfBlessing {
			continue
		}
		if fc.Card.ID == c.ID {
			return true
		}
	}
	return false
}

// RemoveBlessingByCardID 按 cardID 移除祝福盖牌。
func RemoveBlessingByCardID(p *model.Player, cardID string) bool {
	if p == nil || cardID == "" {
		return false
	}
	for _, fc := range p.Field {
		if fc == nil || fc.Mode != model.FieldCover || fc.Effect != model.EffectElfBlessing {
			continue
		}
		if fc.Card.ID == cardID {
			p.RemoveFieldCard(fc)
			return true
		}
	}
	return false
}

// BlessingHandIndices 返回手牌中祝福卡的索引（ID 以 "elf_blessing:" 为前缀）。
func BlessingHandIndices(p *model.Player) []int {
	if p == nil {
		return nil
	}
	var idxs []int
	for i, c := range p.Hand {
		if strings.HasPrefix(c.ID, elfBlessingPrefix) {
			idxs = append(idxs, i)
		}
	}
	return idxs
}

// SyncBlessings 从场上祝福盖牌同步 p.Blessings 和 p.CharaZone。
func SyncBlessings(p *model.Player) {
	if p == nil {
		return
	}
	blessings := BlessingCards(p)
	p.Blessings = blessings
	blessingIDs := map[string]bool{}
	for _, c := range blessings {
		if c.ID != "" {
			blessingIDs[c.ID] = true
		}
	}
	newZone := make([]string, 0, len(p.CharaZone)+len(blessings))
	zoneHas := map[string]bool{}
	for _, z := range p.CharaZone {
		if !strings.HasPrefix(z, elfBlessingPrefix) {
			newZone = append(newZone, z)
			zoneHas[z] = true
			continue
		}
		cardID := strings.TrimPrefix(z, elfBlessingPrefix)
		if blessingIDs[cardID] {
			newZone = append(newZone, z)
			zoneHas[z] = true
		}
	}
	for _, c := range blessings {
		if c.ID == "" {
			continue
		}
		key := elfBlessingPrefix + c.ID
		if zoneHas[key] {
			continue
		}
		newZone = append(newZone, key)
	}
	p.CharaZone = newZone
}
