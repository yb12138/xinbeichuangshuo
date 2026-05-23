package server

import (
	"encoding/json"
	"testing"
	"time"

	"starcup-engine/internal/engine"
	"starcup-engine/internal/model"
)

func readWSMessage(t *testing.T, ch <-chan []byte) WSMessage {
	t.Helper()
	select {
	case raw := <-ch:
		var msg WSMessage
		if err := json.Unmarshal(raw, &msg); err != nil {
			t.Fatalf("unmarshal ws message: %v", err)
		}
		return msg
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout waiting websocket message")
		return WSMessage{}
	}
}

func readProtocolErrorPayload(t *testing.T, ch <-chan []byte) ProtocolErrorPayload {
	t.Helper()
	msg := readWSMessage(t, ch)
	if msg.Cmd != CmdProtocolError {
		t.Fatalf("expected %s, got %s", CmdProtocolError, msg.Cmd)
	}
	var payload ProtocolErrorPayload
	if err := json.Unmarshal(msg.Data, &payload); err != nil {
		t.Fatalf("unmarshal protocol error payload: %v", err)
	}
	return payload
}

func TestHandleMessage_UnknownCmdReturnsProtocolError(t *testing.T) {
	room := NewRoom("STRICT_CMD")
	client := &Client{
		Room:     room,
		Send:     make(chan []byte, 1),
		PlayerID: "p1",
		Name:     "Alice",
	}

	room.HandleMessage(client, &WSMessage{Cmd: "UnknownCmd", Data: mustMarshal(map[string]interface{}{})})

	payload := readProtocolErrorPayload(t, client.Send)
	if payload.Code != protocolErrorCodeUnknownCmd {
		t.Fatalf("expected code %s, got %s", protocolErrorCodeUnknownCmd, payload.Code)
	}
	if payload.Cmd != "UnknownCmd" {
		t.Fatalf("expected cmd UnknownCmd, got %s", payload.Cmd)
	}
}

func TestHandleAction_InvalidJSONReturnsProtocolError(t *testing.T) {
	room := NewRoom("STRICT_ACTION_JSON")
	room.Engine = engine.NewGameEngine(room)
	room.Started = true
	client := &Client{
		Room:     room,
		Send:     make(chan []byte, 1),
		PlayerID: "p1",
		Name:     "Alice",
	}

	room.handleAction(client, json.RawMessage(`{"action_type":`))

	payload := readProtocolErrorPayload(t, client.Send)
	if payload.Code != protocolErrorCodeInvalidJSON {
		t.Fatalf("expected code %s, got %s", protocolErrorCodeInvalidJSON, payload.Code)
	}
	if payload.Cmd != CmdSubmitAction {
		t.Fatalf("expected cmd %s, got %s", CmdSubmitAction, payload.Cmd)
	}
}

func TestHandleAction_UnknownActionTypeReturnsProtocolError(t *testing.T) {
	room := NewRoom("STRICT_ACTION_TYPE")
	room.Engine = engine.NewGameEngine(room)
	room.Started = true
	if err := room.Engine.AddPlayer("p1", "Alice", "berserker", model.RedCamp); err != nil {
		t.Fatal(err)
	}
	client := &Client{
		Room:     room,
		Send:     make(chan []byte, 1),
		PlayerID: "p1",
		Name:     "Alice",
	}

	room.handleAction(client, mustMarshal(ClientActionRequest{ActionType: "UnknownAction"}))

	payload := readProtocolErrorPayload(t, client.Send)
	if payload.Code != protocolErrorCodeUnknownActionType {
		t.Fatalf("expected code %s, got %s", protocolErrorCodeUnknownActionType, payload.Code)
	}
	if payload.Cmd != CmdSubmitAction {
		t.Fatalf("expected cmd %s, got %s", CmdSubmitAction, payload.Cmd)
	}
}

func TestHandleRoomAction_InvalidJSONReturnsProtocolError(t *testing.T) {
	room := NewRoom("STRICT_ROOM_JSON")
	client := &Client{
		Room:     room,
		Send:     make(chan []byte, 1),
		PlayerID: "p1",
		Name:     "Alice",
	}

	room.handleRoomAction(client, json.RawMessage(`{"action":`))

	payload := readProtocolErrorPayload(t, client.Send)
	if payload.Code != protocolErrorCodeInvalidJSON {
		t.Fatalf("expected code %s, got %s", protocolErrorCodeInvalidJSON, payload.Code)
	}
	if payload.Cmd != CmdRoomAction {
		t.Fatalf("expected cmd %s, got %s", CmdRoomAction, payload.Cmd)
	}
}

func TestHandleRoomAction_UnknownActionReturnsProtocolError(t *testing.T) {
	room := NewRoom("STRICT_ROOM_ACTION")
	client := &Client{
		Room:     room,
		Send:     make(chan []byte, 1),
		PlayerID: "p1",
		Name:     "Alice",
	}

	room.handleRoomAction(client, mustMarshal(RoomActionRequest{Action: "unknown_action"}))

	payload := readProtocolErrorPayload(t, client.Send)
	if payload.Code != protocolErrorCodeUnknownRoomAction {
		t.Fatalf("expected code %s, got %s", protocolErrorCodeUnknownRoomAction, payload.Code)
	}
	if payload.Cmd != CmdRoomAction {
		t.Fatalf("expected cmd %s, got %s", CmdRoomAction, payload.Cmd)
	}
	if payload.Context["room_code"] != "STRICT_ROOM_ACTION" {
		t.Fatalf("expected room_code context STRICT_ROOM_ACTION, got %+v", payload.Context)
	}
	if payload.Context["player_id"] != "p1" {
		t.Fatalf("expected player_id context p1, got %+v", payload.Context)
	}
}
