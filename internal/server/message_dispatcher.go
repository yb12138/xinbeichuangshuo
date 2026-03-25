package server

import (
	"encoding/json"
	"fmt"
	"log"

	"starcup-engine/internal/model"
)

// HandleMessage processes incoming WebSocket messages.
func (r *Room) HandleMessage(client *Client, msg *WSMessage) {
	switch msg.Cmd {
	case CmdSubmitAction:
		r.handleAction(client, msg.Data)
	case CmdChatMessage:
		r.handleChat(client, msg.Data)
	case CmdRoomAction:
		r.handleRoomAction(client, msg.Data)
	}
}

func (r *Room) handleAction(client *Client, payload json.RawMessage) {
	if !r.Started || r.Engine == nil {
		r.sendNotifyTimelineToClient(client, "error", map[string]interface{}{"message": "游戏尚未开始"}, "游戏尚未开始")
		return
	}

	var req ClientActionRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		log.Printf("Error parsing action: %v", err)
		return
	}

	action, err := r.translateClientAction(client.PlayerID, req)
	if err != nil {
		r.sendNotifyTimelineToClient(client, "error", map[string]interface{}{"message": err.Error()}, err.Error())
		return
	}

	if err := r.submitAction(action); err != nil {
		r.sendNotifyTimelineToClient(client, "error", map[string]interface{}{"message": err.Error()}, err.Error())
		return
	}
}

func (r *Room) submitAction(action model.PlayerAction) error {
	return r.executeViaActor(func() error {
		return r.submitActionDirect(action)
	})
}

func (r *Room) submitActionDirect(action model.PlayerAction) error {
	r.engineMu.Lock()
	defer r.engineMu.Unlock()

	if !r.Started || r.Engine == nil {
		return fmt.Errorf("游戏尚未开始")
	}

	r.mu.RLock()
	promptEpochBefore := r.botPromptEpoch
	r.mu.RUnlock()

	if err := r.Engine.HandleAction(action); err != nil {
		return err
	}
	r.Engine.Drive()

	r.mu.RLock()
	promptEpochAfter := r.botPromptEpoch
	r.mu.RUnlock()
	if promptEpochAfter == promptEpochBefore && r.Engine.State.PendingInterrupt == nil {
		r.OnGameEvent(model.GameEvent{Type: model.EventStateUpdate})
	}
	return nil
}

func (r *Room) handleChat(client *Client, payload json.RawMessage) {
	var chatMsg struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal(payload, &chatMsg); err != nil {
		return
	}

	r.broadcastTimeline("chat", map[string]interface{}{
		"player_id":   client.PlayerID,
		"player_name": client.Name,
		"message":     chatMsg.Message,
	}, chatMsg.Message)
}
