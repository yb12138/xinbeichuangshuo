// gameflow: 场牌（FieldCard）管理基础设施。

package player

import "starcup-engine/internal/model"

// CoverCardsByEffect 按效果收集场上盖牌（按场上顺序）。
func CoverCardsByEffect(p *model.Player, effect model.EffectType) []*model.FieldCard {
	if p == nil {
		return nil
	}
	var out []*model.FieldCard
	for _, fc := range p.Field {
		if fc == nil || fc.Mode != model.FieldCover || fc.Effect != effect {
			continue
		}
		out = append(out, fc)
	}
	return out
}

// CoverCountByEffect 统计指定效果的场上盖牌数量。
func CoverCountByEffect(p *model.Player, effect model.EffectType) int {
	return len(CoverCardsByEffect(p, effect))
}

// CoverCountByEffectAndElement 统计指定效果（可按元素过滤）的场上盖牌数量。
func CoverCountByEffectAndElement(p *model.Player, effect model.EffectType, element model.Element) int {
	count := 0
	for _, fc := range CoverCardsByEffect(p, effect) {
		if element != "" && fc.Card.Element != element {
			continue
		}
		count++
	}
	return count
}

// RemoveFirstCoverByEffectAndElement 按场上顺序移除第一张匹配的盖牌。
func RemoveFirstCoverByEffectAndElement(p *model.Player, effect model.EffectType, element model.Element) (model.Card, bool) {
	if p == nil {
		return model.Card{}, false
	}
	for _, fc := range p.Field {
		if fc == nil || fc.Mode != model.FieldCover || fc.Effect != effect {
			continue
		}
		if element != "" && fc.Card.Element != element {
			continue
		}
		card := fc.Card
		p.RemoveFieldCard(fc)
		return card, true
	}
	return model.Card{}, false
}

// RemoveCoverCardByEffectAndID 按效果和卡牌ID移除盖牌。
func RemoveCoverCardByEffectAndID(p *model.Player, effect model.EffectType, cardID string) bool {
	if p == nil || cardID == "" {
		return false
	}
	for _, fc := range p.Field {
		if fc == nil || fc.Mode != model.FieldCover || fc.Effect != effect {
			continue
		}
		if fc.Card.ID == cardID {
			p.RemoveFieldCard(fc)
			return true
		}
	}
	return false
}
