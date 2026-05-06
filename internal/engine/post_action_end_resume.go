// gameflow: 行动结束后的核心恢复状态。

package engine

import (
	"fmt"

	"starcup-engine/internal/model"
)

func (e *GameEngine) queuePostActionEndResume(playerID string, actionType model.ActionType) {
	if e == nil || e.State == nil || playerID == "" || actionType == "" {
		return
	}
	if e.postActionEndResume != nil && e.postActionEndResume.playerID == playerID && e.postActionEndResume.actionType == actionType {
		return
	}
	e.postActionEndResume = &postActionEndResumeState{playerID: playerID, actionType: actionType}
}

func (e *GameEngine) processPostActionEndResume() bool {
	if e == nil || e.postActionEndResume == nil || e.State.PendingInterrupt != nil {
		return false
	}
	state := e.postActionEndResume
	e.postActionEndResume = nil
	player := e.State.Players[state.playerID]
	if player == nil {
		e.Log(fmt.Sprintf("[Warn] 行动结束恢复失败：执行者不存在 %s", state.playerID))
		return true
	}
	e.handlePostActionEndEffects(player, state.actionType)
	return true
}
