import type {
  ProtocolErrorPayload,
  RequireActionPayload,
  RoomEventPayload,
  RoutedWsMessage,
  SyncStatePayload,
  TimelineNotifyPayload,
} from './protocol'

export interface MessageRouterHandlers {
  onRoomEvent: (payload: RoomEventPayload) => void
  onSyncState: (payload: SyncStatePayload) => void
  onRequireAction: (payload: RequireActionPayload) => void
  onNotifyTimeline: (payload: TimelineNotifyPayload) => void
  onProtocolError: (payload: ProtocolErrorPayload) => void
  onChatMessage?: (payload: Record<string, unknown>) => void
  onUnknown?: (cmd: string, payload: unknown) => void
}

export function routeWsMessage(msg: RoutedWsMessage, handlers: MessageRouterHandlers) {
  if ('Known' in msg) {
    handlers.onUnknown?.(msg.Cmd, msg.Data)
    return
  }

  switch (msg.Cmd) {
    case 'RoomEvent':
      handlers.onRoomEvent(msg.Data)
      break
    case 'SyncState':
      handlers.onSyncState(msg.Data)
      break
    case 'RequireAction':
      handlers.onRequireAction(msg.Data)
      break
    case 'NotifyTimeline':
      handlers.onNotifyTimeline(msg.Data)
      break
    case 'ChatMessage':
      handlers.onChatMessage?.(msg.Data)
      break
    case 'ProtocolError':
      handlers.onProtocolError(msg.Data)
      break
  }
}
