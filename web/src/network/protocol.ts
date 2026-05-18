import type {
  ClientActionRequest as GeneratedClientActionRequest,
  RequireActionPayload as GeneratedRequireActionPayload,
  RoomActionRequest as GeneratedRoomActionRequest,
  SyncStatePayload as GeneratedSyncStatePayload,
  TimelineNotifyPayload as GeneratedTimelineNotifyPayload,
  TargetNode as GeneratedTargetNode,
  TimelineDelta as GeneratedTimelineDelta,
  TimelineEvent as GeneratedTimelineEvent,
  WSMessage,
} from '../types/generated'
import type { Prompt, RoomEvent } from '../types/game'

export type WsMessage<T = unknown> = Omit<WSMessage, 'Data'> & {
  Cmd: string
  Data: T
}

export type ClientActionRequest = GeneratedClientActionRequest
export type RoomActionRequest = GeneratedRoomActionRequest
export type SyncStatePayload = GeneratedSyncStatePayload
export type TimelineNotifyPayload = GeneratedTimelineNotifyPayload
export type TimelineDelta = GeneratedTimelineDelta
export type TimelineEvent = GeneratedTimelineEvent
export type TargetNode = GeneratedTargetNode

export type RequireActionPayload = Omit<GeneratedRequireActionPayload, 'prompt'> & {
  prompt?: Prompt
}
export type RoomEventPayload = RoomEvent
