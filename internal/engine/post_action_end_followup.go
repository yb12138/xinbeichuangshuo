package engine

import (
	"fmt"

	"starcup-engine/internal/model"
)

func buildPostActionEndDeferredFollowupHandlers() map[string]deferredFollowupHandler {
	return map[string]deferredFollowupHandler{
		model.DeferredFollowupPostActionEnd: {
			label:   "PostActionEnd",
			resolve: (*GameEngine).resolvePostActionEndFollowup,
		},
	}
}

func (e *GameEngine) enqueuePostActionEndFollowup(playerID string, actionType model.ActionType) {
	if e == nil || e.State == nil || playerID == "" || actionType == "" {
		return
	}
	for _, followup := range e.State.DeferredFollowups {
		if followup.Type != model.DeferredFollowupPostActionEnd || followup.UserID != playerID {
			continue
		}
		if action, _ := followup.Data["action_type"].(string); action == string(actionType) {
			return
		}
	}
	e.enqueueDeferredFollowup(model.DeferredFollowup{
		Type:   model.DeferredFollowupPostActionEnd,
		UserID: playerID,
		Data: map[string]interface{}{
			"action_type": string(actionType),
		},
	})
}

func (e *GameEngine) resolvePostActionEndFollowup(f model.DeferredFollowup) error {
	player := e.State.Players[f.UserID]
	if player == nil {
		return fmt.Errorf("行动结束后续执行者不存在: %s", f.UserID)
	}
	rawActionType := ""
	if f.Data != nil {
		rawActionType, _ = f.Data["action_type"].(string)
	}
	actionType := model.ActionType(rawActionType)
	if actionType == "" {
		return fmt.Errorf("行动结束后续缺少 action_type")
	}
	e.handlePostActionEndEffects(player, actionType)
	return nil
}
