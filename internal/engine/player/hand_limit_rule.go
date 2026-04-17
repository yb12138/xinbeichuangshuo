// gameflow: 玩家规则接口定义。

package player

import "starcup-engine/internal/model"

// HandLimitRule 定义玩家手牌上限规则。
type HandLimitRule interface {
	HardCap(player *model.Player) (int, bool)
	ModifierCap(player *model.Player, current int) int
}

// HandLimitRuleFuncs 通过函数组合实现 HandLimitRule。
type HandLimitRuleFuncs struct {
	Hard     func(player *model.Player) (int, bool)
	Modifier func(player *model.Player, current int) int
}

func (r HandLimitRuleFuncs) HardCap(player *model.Player) (int, bool) {
	if r.Hard == nil {
		return 0, false
	}
	return r.Hard(player)
}

func (r HandLimitRuleFuncs) ModifierCap(player *model.Player, current int) int {
	if r.Modifier == nil {
		return current
	}
	return r.Modifier(player, current)
}

var DefaultHandLimitRule HandLimitRule = HandLimitRuleFuncs{}
