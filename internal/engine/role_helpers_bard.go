// gameflow: 吟游诗人：永恒乐章相关辅助。
package engine

import (
	"fmt"

	"starcup-engine/internal/model"
)

func bardEternalMovementCard(bard *model.Player) model.Card {
	id := "bd_eternal_movement"
	if bard != nil && bard.ID != "" {
		id = "bd_eternal_movement_" + bard.ID
	}
	return model.Card{
		ID:          id,
		Name:        "永恒乐章",
		Type:        model.CardTypeMagic,
		Element:     model.ElementDark,
		Description: "吟游诗人的永恒乐章指示牌",
	}
}

func (e *GameEngine) findBardEternalMovement(bard *model.Player) (*model.Player, *model.FieldCard) {
	return e.findSourceEffectCard(bard, model.EffectBardEternalMovement)
}

func (e *GameEngine) bardEternalHolderID(bard *model.Player) string {
	holder, _ := e.findBardEternalMovement(bard)
	if holder == nil {
		return ""
	}
	return holder.ID
}

func (e *GameEngine) removeBardEternalMovement(bard *model.Player) bool {
	holder, card, ok := e.detachSourceEffectCard(bard, model.EffectBardEternalMovement)
	if !ok {
		return false
	}
	e.State.DiscardPile = append(e.State.DiscardPile, card)
	e.emitBuffRemovedDispatch(bard.ID, holder.ID, model.EffectBardEternalMovement)
	return true
}

func (e *GameEngine) placeBardEternalMovement(bard *model.Player, target *model.Player) error {
	return e.placeBardEternalMovementWithCard(bard, target, bardEternalMovementCard(bard))
}

func (e *GameEngine) placeBardEternalMovementWithCard(bard *model.Player, target *model.Player, card model.Card) error {
	if bard == nil || target == nil {
		return fmt.Errorf("放置永恒乐章时角色不存在")
	}
	if target.Camp != bard.Camp {
		return fmt.Errorf("永恒乐章只能放置在我方角色面前")
	}
	e.removeBardEternalMovement(bard)
	return e.attachSourceEffectCard(bard, target, model.EffectBardEternalMovement, card)
}

func (e *GameEngine) transferBardEternalMovement(bard *model.Player, target *model.Player) error {
	if bard == nil || target == nil {
		return fmt.Errorf("转移永恒乐章时角色不存在")
	}
	if target.Camp != bard.Camp {
		return fmt.Errorf("永恒乐章只能转移给我方角色")
	}
	holder, _ := e.findBardEternalMovement(bard)
	if holder == nil {
		return fmt.Errorf("当前没有永恒乐章可转移")
	}
	if holder.ID == target.ID {
		return fmt.Errorf("永恒乐章已在该角色面前")
	}
	holder, card, ok := e.detachSourceEffectCard(bard, model.EffectBardEternalMovement)
	if !ok || holder == nil {
		return fmt.Errorf("当前没有永恒乐章可转移")
	}
	return e.attachSourceEffectCard(bard, target, model.EffectBardEternalMovement, card)
}

func bardInspiration(player *model.Player) int {
	return tokenValueBounded(player, "bd_inspiration", bardInspirationCapEngine)
}

func addBardInspiration(player *model.Player, delta int) int {
	return addTokenValueBounded(player, "bd_inspiration", delta, bardInspirationCapEngine)
}

// gameflow: 吟游诗人：战歌、魅惑、沉沦协奏曲等技能流与回合钩子。

func parseIntSliceContextValue(raw interface{}) []int {
	result := make([]int, 0)
	switch value := raw.(type) {
	case []int:
		result = append(result, value...)
	case []interface{}:
		for _, item := range value {
			switch v := item.(type) {
			case int:
				result = append(result, v)
			case float64:
				result = append(result, int(v))
			}
		}
	}
	return result
}
