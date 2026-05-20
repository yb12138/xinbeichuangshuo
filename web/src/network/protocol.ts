import type {
  ClientActionRequest as GeneratedClientActionRequest,
  ProtocolErrorPayload as GeneratedProtocolErrorPayload,
  RequireActionPayload as GeneratedRequireActionPayload,
  RoomActionRequest as GeneratedRoomActionRequest,
  RoomActionType as GeneratedRoomActionType,
  SyncStatePayload as GeneratedSyncStatePayload,
  TimelineNotifyPayload as GeneratedTimelineNotifyPayload,
  TargetNode as GeneratedTargetNode,
  TimelineDelta as GeneratedTimelineDelta,
  TimelineEvent as GeneratedTimelineEvent,
  WSCommand as GeneratedWSCommand,
  WSMessage,
} from '../types/generated'
import type { Prompt, RoomEvent } from '../types/game'

export type WsMessage<T = unknown> = Omit<WSMessage, 'Data'> & {
  Cmd: GeneratedWSCommand
  Data: T
}

export type WSCommand = GeneratedWSCommand
export type RoomActionType = GeneratedRoomActionType
export type ClientActionRequest = GeneratedClientActionRequest
export type RoomActionRequest = GeneratedRoomActionRequest
export type ProtocolErrorPayload = GeneratedProtocolErrorPayload
export type SyncStatePayload = GeneratedSyncStatePayload
export type TimelineNotifyPayload = GeneratedTimelineNotifyPayload
export type TimelineDelta = GeneratedTimelineDelta
export type TimelineEvent = GeneratedTimelineEvent
export type TargetNode = GeneratedTargetNode

export type RequireActionPayload = Omit<GeneratedRequireActionPayload, 'prompt'> & {
  prompt?: Prompt
}
export type RoomEventPayload = RoomEvent

export type WsInboundMessage =
  | { Cmd: 'RoomEvent'; Data: RoomEventPayload }
  | { Cmd: 'SyncState'; Data: SyncStatePayload }
  | { Cmd: 'RequireAction'; Data: RequireActionPayload }
  | { Cmd: 'NotifyTimeline'; Data: TimelineNotifyPayload }
  | { Cmd: 'ChatMessage'; Data: Record<string, unknown> }
  | { Cmd: 'ProtocolError'; Data: ProtocolErrorPayload }

export type WsOutboundMessage =
  | { Cmd: 'SubmitAction'; Data: ClientActionRequest }
  | { Cmd: 'RoomAction'; Data: RoomActionRequest }
  | { Cmd: 'ChatMessage'; Data: Record<string, string> }

export type UnknownWsMessage = {
  Known: false
  Cmd: string
  Data: unknown
}

export type RoutedWsMessage = WsInboundMessage | UnknownWsMessage

const inboundCommands = new Set<WsInboundMessage['Cmd']>([
  'RoomEvent',
  'SyncState',
  'RequireAction',
  'NotifyTimeline',
  'ChatMessage',
  'ProtocolError',
])

export function normalizeWsMessage(raw: unknown): RoutedWsMessage | null {
  if (!raw || typeof raw !== 'object') return null
  const candidate = raw as { Cmd?: unknown; Data?: unknown }
  if (typeof candidate.Cmd !== 'string') return null
  const data = candidate.Data

  if (!inboundCommands.has(candidate.Cmd as WsInboundMessage['Cmd'])) {
    return {
      Known: false,
      Cmd: candidate.Cmd,
      Data: data,
    }
  }

  return {
    Cmd: candidate.Cmd,
    Data: data,
  } as WsInboundMessage
}
