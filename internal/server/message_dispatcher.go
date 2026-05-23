package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"starcup-engine/internal/model"
	"starcup-engine/internal/server/timeline"
)

const (
	protocolErrorCodeUnknownCmd        = "unknown_cmd"
	protocolErrorCodeInvalidJSON       = "invalid_json"
	protocolErrorCodeUnknownActionType = "unknown_action_type"
	protocolErrorCodeUnknownRoomAction = "unknown_room_action"
)

type protocolInputError struct {
	code    string
	message string
	context map[string]interface{}
}

func (e *protocolInputError) Error() string {
	return e.message
}

func newProtocolInputError(code, message string, context map[string]interface{}) *protocolInputError {
	return &protocolInputError{
		code:    code,
		message: message,
		context: context,
	}
}

func (r *Room) sendProtocolErrorToClient(client *Client, code, message string, cmd WSCommand, context map[string]interface{}) {
	payload := ProtocolErrorPayload{
		Code:    code,
		Message: message,
		Cmd:     cmd,
		Context: context,
	}
	r.sendToClient(client, CmdProtocolError, payload)
}

// HandleMessage processes incoming WebSocket messages.
func (r *Room) HandleMessage(client *Client, msg *WSMessage) {
	if msg == nil {
		r.sendProtocolErrorToClient(client, protocolErrorCodeInvalidJSON, "消息体为空", WSCommand(""), nil)
		return
	}
	switch msg.Cmd {
	case CmdSubmitAction:
		r.handleAction(client, msg.Data)
	case CmdChatMessage:
		r.handleChat(client, msg.Data)
	case CmdRoomAction:
		r.handleRoomAction(client, msg.Data)
	default:
		r.sendProtocolErrorToClient(client, protocolErrorCodeUnknownCmd, "未知命令", msg.Cmd, map[string]interface{}{
			"cmd": msg.Cmd,
		})
	}
}

func (r *Room) handleAction(client *Client, payload json.RawMessage) {
	if !r.Started || r.Engine == nil {
		r.sendNotifyTimelineToClient(client, timeline.Payload{Type: "error", Message: "游戏尚未开始"})
		return
	}

	var req ClientActionRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		log.Printf("Error parsing action: %v", err)
		r.sendProtocolErrorToClient(client, protocolErrorCodeInvalidJSON, "SubmitAction 负载不是合法 JSON", CmdSubmitAction, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	action, err := r.translateClientAction(client.PlayerID, req)
	if err != nil {
		var protocolErr *protocolInputError
		if errors.As(err, &protocolErr) {
			r.sendProtocolErrorToClient(client, protocolErr.code, protocolErr.message, CmdSubmitAction, protocolErr.context)
			return
		}
		r.sendNotifyTimelineToClient(client, timeline.Payload{Type: "error", Message: err.Error()})
		return
	}

	if err := r.submitAction(action); err != nil {
		r.sendNotifyTimelineToClient(client, timeline.Payload{Type: "error", Message: err.Error()})
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
		r.sendProtocolErrorToClient(client, protocolErrorCodeInvalidJSON, "ChatMessage 负载不是合法 JSON", CmdChatMessage, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	r.broadcastTimeline(timeline.Payload{
		Type:       "chat",
		Message:    chatMsg.Message,
		PlayerID:   client.PlayerID,
		PlayerName: client.Name,
	})
}
