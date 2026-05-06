// gameflow: 玩家能量上限规则接口定义。

package player

import "starcup-engine/internal/model"

// EnergyCapRule 定义玩家能量上限规则（与 HandLimitRule 同模式）。
type EnergyCapRule interface {
	ModifierCap(player *model.Player, current int) int
}

// EnergyCapRuleFuncs 通过函数组合实现 EnergyCapRule。
type EnergyCapRuleFuncs struct {
	Modifier func(player *model.Player, current int) int
}

func (r EnergyCapRuleFuncs) ModifierCap(player *model.Player, current int) int {
	if r.Modifier == nil {
		return current
	}
	return r.Modifier(player, current)
}

var DefaultEnergyCapRule EnergyCapRule = EnergyCapRuleFuncs{}
