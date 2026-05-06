import type { RequireActionPayload, RoomEventPayload, SyncStatePayload, TimelineNotifyPayload, WsMessage } from './protocol'

export interface MessageRouterHandlers {
  onRoomEvent: (payload: RoomEventPayload) => void
  onSyncState: (payload: SyncStatePayload) => void
  onRequireAction: (payload: RequireActionPayload) => void
  onNotifyTimeline: (payload: TimelineNotifyPayload) => void
  onChatMessage?: (payload: Record<string, unknown>) => void
  onUnknown?: (cmd: string, payload: unknown) => void
}

export function routeWsMessage(msg: WsMessage, handlers: MessageRouterHandlers) {
  switch (msg.Cmd) {
    case 'RoomEvent':
      handlers.onRoomEvent(msg.Data as RoomEventPayload)
      break
    case 'SyncState':
      handlers.onSyncState(msg.Data as SyncStatePayload)
      break
    case 'RequireAction':
      handlers.onRequireAction(msg.Data as RequireActionPayload)
      break
    case 'NotifyTimeline':
      handlers.onNotifyTimeline(msg.Data as TimelineNotifyPayload)
      break
    case 'ChatMessage':
      handlers.onChatMessage?.(msg.Data as Record<string, unknown>)
      break
    default:
      handlers.onUnknown?.(msg.Cmd, msg.Data)
      break
  }
}
