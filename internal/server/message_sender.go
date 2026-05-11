package server

import (
	"encoding/json"

	"starcup-engine/internal/model"
	"starcup-engine/internal/server/prompting"
)

func (r *Room) sendToClient(client *Client, cmd string, data interface{}) {
	if client == nil || client.IsBot || client.Disconnected {
		return
	}
	client.SendMessage(newWSMessage(cmd, data))
}

func (r *Room) sendRoomEventToClient(client *Client, event RoomEvent) {
	r.sendToClient(client, CmdRoomEvent, event)
}

func (r *Room) sendNotifyTimelineToClient(client *Client, eventType string, data map[string]interface{}, message string) {
	if client == nil {
		return
	}
	r.sendToClient(client, CmdNotifyTimeline, r.buildTimelineNotify(eventType, data, message))
}

func (r *Room) sendSyncStateToClient(client *Client) {
	if client == nil {
		return
	}
	r.sendToClient(client, CmdSyncState, r.buildSyncStatePayload(client.PlayerID))
}

func (r *Room) sendRequireActionToClient(client *Client, prompt *model.Prompt) {
	if prompt == nil {
		return
	}
	r.sendToClient(client, CmdRequireAction, prompting.BuildRequireActionPayload(prompt))
}

func (r *Room) broadcastHumans(cmd string, data interface{}) {
	msg := newWSMessage(cmd, data)
	raw, _ := json.Marshal(msg)
	for _, c := range r.Clients {
		if c == nil || c.IsBot || c.Disconnected {
			continue
		}
		select {
		case c.Send <- raw:
		default:
		}
	}
}

func (r *Room) broadcastRoomEvent(event RoomEvent) {
	r.broadcastHumans(CmdRoomEvent, event)
}

func (r *Room) broadcastToAll(message []byte) {
	for _, client := range r.Clients {
		if client.IsBot || client.Disconnected {
			continue
		}
		select {
		case client.Send <- message:
		default:
		}
	}
}
