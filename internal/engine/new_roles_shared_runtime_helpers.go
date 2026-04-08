package engine

import "starcup-engine/internal/model"

// tokenValueBounded 读取并规范化玩家 token 值：
// 小于 0 归零；cap >= 0 时按上限裁剪；并回写到玩家状态。
func tokenValueBounded(player *model.Player, key string, cap int) int {
	if player == nil {
		return 0
	}
	ensurePlayerTokensMap(player)
	v := player.Tokens[key]
	if v < 0 {
		v = 0
	}
	if cap >= 0 && v > cap {
		v = cap
	}
	player.Tokens[key] = v
	return v
}

// addTokenValueBounded 在规范化基础上增减 token，并应用统一上限规则。
func addTokenValueBounded(player *model.Player, key string, delta int, cap int) int {
	return addTokenValueBoundedWithIgnoreCap(player, key, delta, cap, false)
}

// addTokenValueBoundedWithIgnoreCap 允许按场景跳过上限裁剪（仅保留非负约束）。
func addTokenValueBoundedWithIgnoreCap(player *model.Player, key string, delta int, cap int, ignoreCap bool) int {
	if player == nil {
		return 0
	}
	ensurePlayerTokensMap(player)
	baseCap := cap
	if ignoreCap {
		baseCap = -1
	}
	v := tokenValueBounded(player, key, baseCap) + delta
	if v < 0 {
		v = 0
	}
	if !ignoreCap && cap >= 0 && v > cap {
		v = cap
	}
	player.Tokens[key] = v
	return v
}

// coverCardsByEffect 收集指定场上盖牌效果的所有卡牌引用（按场上顺序）。
func coverCardsByEffect(player *model.Player, effect model.EffectType) []*model.FieldCard {
	if player == nil {
		return nil
	}
	var out []*model.FieldCard
	for _, fc := range player.Field {
		if fc == nil || fc.Mode != model.FieldCover || fc.Effect != effect {
			continue
		}
		out = append(out, fc)
	}
	return out
}

// coverCountByEffectAndElement 统计指定效果（可按元素过滤）的场上盖牌数量。
func coverCountByEffectAndElement(player *model.Player, effect model.EffectType, element model.Element) int {
	count := 0
	for _, fc := range coverCardsByEffect(player, effect) {
		if element != "" && fc.Card.Element != element {
			continue
		}
		count++
	}
	return count
}

// removeFirstCoverByEffectAndElement 按场上顺序移除第一张匹配的盖牌。
func removeFirstCoverByEffectAndElement(player *model.Player, effect model.EffectType, element model.Element) (model.Card, bool) {
	if player == nil {
		return model.Card{}, false
	}
	for _, fc := range player.Field {
		if fc == nil || fc.Mode != model.FieldCover || fc.Effect != effect {
			continue
		}
		if element != "" && fc.Card.Element != element {
			continue
		}
		card := fc.Card
		player.RemoveFieldCard(fc)
		return card, true
	}
	return model.Card{}, false
}
