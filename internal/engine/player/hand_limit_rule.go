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

// HandLimitModifier 全局手牌上限修改器。
// 当计算任意玩家手牌上限时，遍历所有角色的 Modifier 收集 delta。
// 适用于一个角色的场上效果影响其他玩家手牌上限的场景（如血之巫女同生共死）。
type HandLimitModifier func(engine HandLimitModifierEngine, target *model.Player) int

// HandLimitModifierEngine 抽象手牌上限修改器所需的引擎能力。
type HandLimitModifierEngine interface {
	GetAllPlayers() []*model.Player
	LookupPlayer(playerID string) *model.Player
	HasFixedMaxHandCap(player *model.Player) bool
}
