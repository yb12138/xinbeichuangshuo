package server

import (
	"encoding/json"
	"testing"
	"time"

	"starcup-engine/internal/model"
)

func waitRoomEvent(t *testing.T, ch <-chan []byte, action string) RoomEvent {
	t.Helper()
	timeout := time.After(800 * time.Millisecond)
	for {
		select {
		case raw, ok := <-ch:
			if !ok {
				t.Fatal("channel closed before expected room event")
			}
			var msg WSMessage
			if err := json.Unmarshal(raw, &msg); err != nil {
				continue
			}
			if msg.Type != "room" {
				continue
			}
			var ev RoomEvent
			if err := json.Unmarshal(msg.Payload, &ev); err != nil {
				continue
			}
			if ev.Action == action {
				return ev
			}
		case <-timeout:
			t.Fatalf("timeout waiting room event action=%s", action)
		}
	}
}

func findPlayerInfo(players []PlayerInfo, id string) *PlayerInfo {
	for i := range players {
		if players[i].ID == id {
			return &players[i]
		}
	}
	return nil
}

func TestRoomUnregister_StartedHumanKeepsReconnectableSeat(t *testing.T) {
	room := NewRoom("R001")
	host := &Client{
		Room:     room,
		Send:     make(chan []byte, 64),
		PlayerID: "p2",
		Name:     "Host",
		Camp:     model.BlueCamp,
		CharRole: "angel",
	}
	leaver := &Client{
		Room:           room,
		Send:           make(chan []byte, 64),
		PlayerID:       "p1",
		Name:           "Alice",
		Camp:           model.RedCamp,
		CharRole:       "berserker",
		ReconnectToken: "tok-alice",
	}
	room.Clients["p1"] = leaver
	room.Clients["p2"] = host
	room.Started = true
	room.HostID = "p2"

	room.handleUnregister(leaver)

	kept := room.Clients["p1"]
	if kept == nil {
		t.Fatal("expected disconnected player seat kept for reconnect")
	}
	if kept.IsBot {
		t.Fatal("expected disconnected player NOT auto-switched to bot")
	}
	if !kept.Disconnected {
		t.Fatal("expected kept seat marked disconnected")
	}

	ev := waitRoomEvent(t, host.Send, "player_list")
	p1 := findPlayerInfo(ev.Players, "p1")
	if p1 == nil {
		t.Fatal("expected p1 in player_list")
	}
	if p1.IsOnline {
		t.Fatal("expected p1 is_online=false after disconnect")
	}
	if p1.IsBot {
		t.Fatal("expected p1 still human seat after disconnect")
	}
}

func TestRoomHostCanTakeoverDisconnectedPlayerWithBot(t *testing.T) {
	room := NewRoom("R002")
	host := &Client{
		Room:     room,
		Send:     make(chan []byte, 64),
		PlayerID: "p2",
		Name:     "Host",
		Camp:     model.BlueCamp,
		CharRole: "angel",
	}
	offline := &Client{
		Room:           room,
		Send:           make(chan []byte, 64),
		PlayerID:       "p1",
		Name:           "Alice",
		Camp:           model.RedCamp,
		CharRole:       "berserker",
		ReconnectToken: "tok-alice",
		Disconnected:   true,
	}
	room.Clients["p1"] = offline
	room.Clients["p2"] = host
	room.Started = true
	room.HostID = "p2"

	room.handleRoomAction(host, mustMarshal(map[string]interface{}{
		"action":    "takeover_player",
		"target_id": "p1",
	}))

	takeover := room.Clients["p1"]
	if takeover == nil {
		t.Fatal("expected takeover bot seat exists")
	}
	if !takeover.IsBot {
		t.Fatal("expected p1 switched to bot after host takeover")
	}
	if takeover.BotMode != "takeover" {
		t.Fatalf("expected bot_mode=takeover, got %q", takeover.BotMode)
	}
	if takeover.ReconnectToken != "tok-alice" {
		t.Fatalf("expected reconnect token kept, got %q", takeover.ReconnectToken)
	}

	ev := waitRoomEvent(t, host.Send, "player_list")
	p1 := findPlayerInfo(ev.Players, "p1")
	if p1 == nil {
		t.Fatal("expected p1 in player_list")
	}
	if !p1.IsBot {
		t.Fatal("expected p1 marked bot in player_list")
	}
}

func TestRoomReconnectByRoomAndNameWithoutToken(t *testing.T) {
	room := NewRoom("R003")
	host := &Client{
		Room:     room,
		Send:     make(chan []byte, 64),
		PlayerID: "p2",
		Name:     "Host",
		Camp:     model.BlueCamp,
		CharRole: "angel",
	}
	oldSeat := &Client{
		Room:           room,
		Send:           make(chan []byte, 64),
		PlayerID:       "p1",
		Name:           "Alice",
		Camp:           model.RedCamp,
		CharRole:       "berserker",
		ReconnectToken: "tok-alice-old",
		Disconnected:   true,
	}
	room.Clients["p1"] = oldSeat
	room.Clients["p2"] = host
	room.Started = true
	room.HostID = "p2"

	incoming := &Client{
		Name: "Alice",
		Send: make(chan []byte, 64),
	}
	room.handleRegister(incoming)

	if room.Clients["p1"] != incoming {
		t.Fatal("expected reconnect by name to reclaim p1 seat")
	}
	if incoming.PlayerID != "p1" {
		t.Fatalf("expected incoming player_id=p1, got %q", incoming.PlayerID)
	}
	if incoming.Disconnected {
		t.Fatal("expected incoming client connected after reconnect")
	}
	if incoming.ReconnectToken == "" || incoming.ReconnectToken == "tok-alice-old" {
		t.Fatalf("expected rotated reconnect token, got %q", incoming.ReconnectToken)
	}

	assigned := waitRoomEvent(t, incoming.Send, "assigned")
	if assigned.PlayerID != "p1" {
		t.Fatalf("expected assigned player p1, got %q", assigned.PlayerID)
	}
	if assigned.Message == "" {
		t.Fatal("expected assigned message for reconnect")
	}
}
