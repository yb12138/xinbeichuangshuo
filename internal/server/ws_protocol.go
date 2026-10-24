package server

import (
	"encoding/json"

	"starcup-engine/internal/server/protocol"
)

const (
	CmdSyncState      = protocol.CmdSyncState
	CmdRequireAction  = protocol.CmdRequireAction
	CmdNotifyTimeline = protocol.CmdNotifyTimeline
	CmdSubmitAction   = protocol.CmdSubmitAction
	CmdRoomAction     = protocol.CmdRoomAction
	CmdRoomEvent      = protocol.CmdRoomEvent
	CmdChatMessage    = protocol.CmdChatMessage
	CmdProtocolError  = protocol.CmdProtocolError

	RoomActionDissolveRoom   = protocol.RoomActionDissolveRoom
	RoomActionAddBot         = protocol.RoomActionAddBot
	RoomActionRemoveBot      = protocol.RoomActionRemoveBot
	RoomActionTakeoverPlayer = protocol.RoomActionTakeoverPlayer
	RoomActionChangeCamp     = protocol.RoomActionChangeCamp
	RoomActionChangeRole     = protocol.RoomActionChangeRole
	RoomActionStart          = protocol.RoomActionStart
)

type WSCommand = protocol.WSCommand

type WSMessage = protocol.WSMessage

type ProtocolErrorPayload = protocol.ProtocolErrorPayload

type TargetNode = protocol.TargetNode

type ClientActionRequest = protocol.ClientActionRequest

type RoomActionType = protocol.RoomActionType

type RoomActionRequest = protocol.RoomActionRequest

type SyncStatePayload = protocol.SyncStatePayload

type RequireActionPayload = protocol.RequireActionPayload

type TimelineDelta = protocol.TimelineDelta

type TimelineEvent = protocol.TimelineEvent

type ActionFlowActorDTO = protocol.ActionFlowActorDTO

type ActionFlowNodeDTO = protocol.ActionFlowNodeDTO

type ActionFlowEdgeDTO = protocol.ActionFlowEdgeDTO

type ActionFlowLogDTO = protocol.ActionFlowLogDTO

type ActionFlowDTO = protocol.ActionFlowDTO

type TimelineNotifyPayload = protocol.TimelineNotifyPayload

func newWSMessage(cmd protocol.WSCommand, data interface{}) WSMessage {
	return protocol.WSMessage{
		Cmd:  cmd,
		Data: mustMarshal(data),
	}
}

func mustMarshal(v interface{}) json.RawMessage {
	data, _ := json.Marshal(v)
	return data
}
