// gameflow: 手牌上限、可见性等规则。

package engine

import (
	engineplayer "starcup-engine/internal/engine/player"
	"starcup-engine/internal/engine/rolecatalog"
	"starcup-engine/internal/model"
)

var roleRegistry = rolecatalog.BuildRoleRegistry()

func roleIDForHandLimitRule(player *model.Player) string {
	if player == nil {
		return ""
	}
	if player.Character != nil && player.Character.ID != "" {
		return player.Character.ID
	}
	return player.Role
}

func (e *GameEngine) roleHandLimitRule(player *model.Player) engineplayer.HandLimitRule {
	return roleRegistry.HandLimitRule(roleIDForHandLimitRule(player))
}

func (e *GameEngine) baseMaxHand(player *model.Player) int {
	if player == nil {
		return 0
	}
	if player.MaxHand > 0 {
		return player.MaxHand
	}
	// 角色基础手牌上限默认值。
	return 6
}

func (e *GameEngine) roleFixedMaxHandCapValue(player *model.Player) (int, bool) {
	if player == nil {
		return 0, false
	}
	return e.roleHandLimitRule(player).HardCap(player)
}

func (e *GameEngine) hasMercyFixedMaxHandCap(player *model.Player) bool {
	if player == nil {
		return false
	}
	for _, fc := range player.Field {
		if fc != nil && fc.Mode == model.FieldEffect && fc.Effect == model.EffectMercy {
			return true
		}
	}
	return false
}

// GetMaxHand 计算玩家的动态手牌上限
func (e *GameEngine) GetMaxHand(player *model.Player) int {
	if player == nil {
		return 0
	}
	base := e.baseMaxHand(player)
	if fixed, ok := e.fixedMaxHandCapValue(player); ok {
		if fixed < 0 {
			return 0
		}
		return fixed
	}

	maxHand := e.applyRoleMaxHandModifiers(player, base)
	if maxHand < 0 {
		return 0
	}
	return maxHand
}

func (e *GameEngine) fixedMaxHandCapValue(player *model.Player) (int, bool) {
	if fixed, ok := e.roleFixedMaxHandCapValue(player); ok {
		return fixed, true
	}
	if e.hasMercyFixedMaxHandCap(player) {
		return 7, true
	}
	return 0, false
}

// HasFixedMaxHandCap 判断玩家是否有固定手牌上限（实现 HandLimitModifierEngine 接口）。
func (e *GameEngine) HasFixedMaxHandCap(player *model.Player) bool {
	_, ok := e.fixedMaxHandCapValue(player)
	return ok
}

func (e *GameEngine) applyRoleMaxHandModifiers(player *model.Player, maxHand int) int {
	maxHand = e.roleHandLimitRule(player).ModifierCap(player, maxHand)
	maxHand += e.collectHandLimitModifiers(player)
	return maxHand
}

// collectHandLimitModifiers 收集所有角色的全局手牌上限修改。
func (e *GameEngine) collectHandLimitModifiers(player *model.Player) int {
	delta := 0
	for _, entry := range roleRegistry.Entries() {
		if entry.HandLimitModifier == nil {
			continue
		}
		delta += entry.HandLimitModifier(e, player)
	}
	return delta
}
