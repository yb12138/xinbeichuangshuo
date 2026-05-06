package server

import "starcup-engine/internal/server/protocol"

const (
	CmdSyncState      = protocol.CmdSyncState
	CmdRequireAction  = protocol.CmdRequireAction
	CmdNotifyTimeline = protocol.CmdNotifyTimeline
	CmdSubmitAction   = protocol.CmdSubmitAction
	CmdRoomAction     = protocol.CmdRoomAction
	CmdRoomEvent      = protocol.CmdRoomEvent
	CmdChatMessage    = protocol.CmdChatMessage
	CmdProtocolError  = protocol.CmdProtocolError
)

type WSMessage = protocol.WSMessage

type ProtocolErrorPayload = protocol.ProtocolErrorPayload

type TargetNode = protocol.TargetNode

type ClientActionRequest = protocol.ClientActionRequest

type RoomActionRequest = protocol.RoomActionRequest

type SyncStatePayload = protocol.SyncStatePayload

type RequireActionPayload = protocol.RequireActionPayload

type TimelineDelta = protocol.TimelineDelta

type TimelineEvent = protocol.TimelineEvent

type TimelineNotifyPayload = protocol.TimelineNotifyPayload

func newWSMessage(cmd string, data interface{}) WSMessage {
	return protocol.WSMessage{
		Cmd:  cmd,
		Data: mustMarshal(data),
	}
}
