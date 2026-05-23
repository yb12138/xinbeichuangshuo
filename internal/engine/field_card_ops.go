// gameflow: 场上牌通用操作（移除、查找等）。

package engine

import (
	"fmt"

	"starcup-engine/internal/model"
)

func (e *GameEngine) RemoveFieldCardAt(targetID string, fieldIndex int, sourceID string) (model.Card, error) {
	target := e.State.Players[targetID]
	if target == nil {
		return model.Card{}, fmt.Errorf("目标不存在")
	}
	if fieldIndex < 0 || fieldIndex >= len(target.Field) {
		return model.Card{}, fmt.Errorf("无效的场上牌索引")
	}
	fc := target.Field[fieldIndex]
	if fc == nil {
		return model.Card{}, fmt.Errorf("场上牌不存在")
	}

	target.Field = append(target.Field[:fieldIndex], target.Field[fieldIndex+1:]...)
	e.State.DiscardPile = append(e.State.DiscardPile, fc.Card)
	e.Log(fmt.Sprintf("%s 的场上牌被移除: %s", target.Name, fc.Effect))
	if fc.Mode == model.FieldEffect {
		e.emitBuffRemovedDispatch(sourceID, targetID, fc.Effect)
	}
	return fc.Card, nil
}
