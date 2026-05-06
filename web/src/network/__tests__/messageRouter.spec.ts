import { describe, expect, it, vi } from 'vitest'
import { routeWsMessage } from '../messageRouter'
import type { WsMessage } from '../protocol'

describe('routeWsMessage', () => {
  it('dispatches known commands to their dedicated handlers', () => {
    const handlers = {
      onRoomEvent: vi.fn(),
      onSyncState: vi.fn(),
      onRequireAction: vi.fn(),
      onNotifyTimeline: vi.fn(),
      onChatMessage: vi.fn(),
      onUnknown: vi.fn(),
    }

    const cases: Array<[WsMessage, keyof typeof handlers]> = [
      [{ Cmd: 'RoomEvent', Data: { action: 'joined' } }, 'onRoomEvent'],
      [{ Cmd: 'SyncState', Data: { room_state: 'Playing' } }, 'onSyncState'],
      [{ Cmd: 'RequireAction', Data: { interrupt_type: 'Prompt' } }, 'onRequireAction'],
      [{ Cmd: 'NotifyTimeline', Data: { room_id: 'ROOM1' } }, 'onNotifyTimeline'],
      [{ Cmd: 'ChatMessage', Data: { message: 'hello' } }, 'onChatMessage'],
    ]

    for (const [message, expectedHandler] of cases) {
      routeWsMessage(message, handlers)
      expect(handlers[expectedHandler]).toHaveBeenCalledWith(message.Data)
    }

    expect(handlers.onUnknown).not.toHaveBeenCalled()
  })

  it('passes unknown commands to the fallback handler', () => {
    const handlers = {
      onRoomEvent: vi.fn(),
      onSyncState: vi.fn(),
      onRequireAction: vi.fn(),
      onNotifyTimeline: vi.fn(),
      onUnknown: vi.fn(),
    }

    routeWsMessage({ Cmd: 'FutureCommand', Data: { foo: 1 } }, handlers)

    expect(handlers.onUnknown).toHaveBeenCalledWith('FutureCommand', { foo: 1 })
  })

  it('treats legacy NotifyEvent envelopes as unknown commands', () => {
    const handlers = {
      onRoomEvent: vi.fn(),
      onSyncState: vi.fn(),
      onRequireAction: vi.fn(),
      onNotifyTimeline: vi.fn(),
      onUnknown: vi.fn(),
    }

    routeWsMessage({ Cmd: 'NotifyEvent', Data: { event_type: 'log' } }, handlers)

    expect(handlers.onUnknown).toHaveBeenCalledWith('NotifyEvent', { event_type: 'log' })
  })

  it('ignores chat messages when the chat handler is omitted', () => {
    const handlers = {
      onRoomEvent: vi.fn(),
      onSyncState: vi.fn(),
      onRequireAction: vi.fn(),
      onNotifyTimeline: vi.fn(),
      onUnknown: vi.fn(),
    }

    expect(() => {
      routeWsMessage({ Cmd: 'ChatMessage', Data: { message: 'hello' } }, handlers)
    }).not.toThrow()
    expect(handlers.onUnknown).not.toHaveBeenCalled()
  })
})
