// gameflow: 新角色共用小型运行时函数。
// 已迁移到 player 包，此文件保留别名供 engine 内部过渡使用。

package engine

import (
	"starcup-engine/internal/engine/player"
	"starcup-engine/internal/model"
)

func tokenValueBounded(p *model.Player, key string, cap int) int {
	return player.TokenValue(p, key, cap)
}

func addTokenValueBounded(p *model.Player, key string, delta int, cap int) int {
	return player.AddToken(p, key, delta, cap)
}

func addTokenValueBoundedWithIgnoreCap(p *model.Player, key string, delta int, cap int, ignoreCap bool) int {
	return player.AddTokenIgnoreCap(p, key, delta, cap, ignoreCap)
}

func coverCardsByEffect(p *model.Player, effect model.EffectType) []*model.FieldCard {
	return player.CoverCardsByEffect(p, effect)
}

func coverCountByEffectAndElement(p *model.Player, effect model.EffectType, element model.Element) int {
	return player.CoverCountByEffectAndElement(p, effect, element)
}

func removeFirstCoverByEffectAndElement(p *model.Player, effect model.EffectType, element model.Element) (model.Card, bool) {
	return player.RemoveFirstCoverByEffectAndElement(p, effect, element)
}
