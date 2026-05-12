// gameflow: 弃牌核心操作（从手牌移除 → 通知观察者）。

package engine

import "starcup-engine/internal/model"

// PerformDiscardFromHand 从手牌中移除指定牌并通知观察者。
// 注意：弃牌堆的添加由调用方在士气损失结算后统一处理（因为存在"吸收"机制）。
func (e *GameEngine) PerformDiscardFromHand(player *model.Player, indices []int) ([]model.Card, error) {
	cards, err := e.discardCardsFromHand(player, indices)
	if err != nil {
		return nil, err
	}
	e.notifyHiddenDiscard(player.ID, cards)
	return cards, nil
}
