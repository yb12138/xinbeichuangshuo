// gameflow: 剑圣风怒等专用中断 Prompt。

package engine

import "starcup-engine/internal/model"

func (e *GameEngine) buildHolySwordDrawPrompt() *model.Prompt {
	interrupt := e.State.PendingInterrupt
	if interrupt == nil {
		return nil
	}
	playerID := interrupt.PlayerID

	return &model.Prompt{
		Type:     model.PromptConfirm,
		PlayerID: playerID,
		Message:  "【圣剑】第3次攻击结束！选择摸X张牌然后弃X张牌 (X=0-3)：",
		Options: []model.PromptOption{
			{ID: "0", Label: "X=0"},
			{ID: "1", Label: "X=1"},
			{ID: "2", Label: "X=2"},
			{ID: "3", Label: "X=3"},
		},
		Min: 1,
		Max: 1,
	}
}
