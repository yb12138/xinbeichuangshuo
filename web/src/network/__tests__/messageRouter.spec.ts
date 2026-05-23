import { describe, expect, it, vi } from 'vitest'
import { routeWsMessage } from '../messageRouter'
import { normalizeWsMessage } from '../protocol'
import type { RoutedWsMessage, UnknownWsMessage } from '../protocol'

function routed(raw: unknown): RoutedWsMessage {
  const msg = normalizeWsMessage(raw)
  if (!msg) throw new Error('invalid test envelope')
  return msg
}

describe('routeWsMessage', () => {
  it('dispatches known commands to their dedicated handlers', () => {
    const handlers = {
      onRoomEvent: vi.fn(),
      onSyncState: vi.fn(),
      onRequireAction: vi.fn(),
      onNotifyTimeline: vi.fn(),
      onProtocolError: vi.fn(),
      onChatMessage: vi.fn(),
      onUnknown: vi.fn(),
    }

    const cases: Array<[RoutedWsMessage, keyof typeof handlers]> = [
      [routed({ Cmd: 'RoomEvent', Data: { action: 'joined' } }), 'onRoomEvent'],
      [routed({ Cmd: 'SyncState', Data: { room_state: 'Playing' } }), 'onSyncState'],
      [routed({ Cmd: 'RequireAction', Data: { interrupt_type: 'Prompt' } }), 'onRequireAction'],
      [routed({ Cmd: 'NotifyTimeline', Data: { room_id: 'ROOM1' } }), 'onNotifyTimeline'],
      [routed({ Cmd: 'ChatMessage', Data: { message: 'hello' } }), 'onChatMessage'],
      [routed({ Cmd: 'ProtocolError', Data: { code: 'unknown_cmd', message: '未知命令' } }), 'onProtocolError'],
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
      onProtocolError: vi.fn(),
      onUnknown: vi.fn(),
    }

    routeWsMessage({ Known: false, Cmd: 'FutureCommand', Data: { foo: 1 } } satisfies UnknownWsMessage, handlers)

    expect(handlers.onUnknown).toHaveBeenCalledWith('FutureCommand', { foo: 1 })
  })

  it('treats legacy NotifyEvent envelopes as unknown commands', () => {
    const handlers = {
      onRoomEvent: vi.fn(),
      onSyncState: vi.fn(),
      onRequireAction: vi.fn(),
      onNotifyTimeline: vi.fn(),
      onProtocolError: vi.fn(),
      onUnknown: vi.fn(),
    }

    routeWsMessage({ Known: false, Cmd: 'NotifyEvent', Data: { event_type: 'log' } } satisfies UnknownWsMessage, handlers)

    expect(handlers.onUnknown).toHaveBeenCalledWith('NotifyEvent', { event_type: 'log' })
  })

  it('ignores chat messages when the chat handler is omitted', () => {
    const handlers = {
      onRoomEvent: vi.fn(),
      onSyncState: vi.fn(),
      onRequireAction: vi.fn(),
      onNotifyTimeline: vi.fn(),
      onProtocolError: vi.fn(),
      onUnknown: vi.fn(),
    }

    expect(() => {
      routeWsMessage(routed({ Cmd: 'ChatMessage', Data: { message: 'hello' } }), handlers)
    }).not.toThrow()
    expect(handlers.onUnknown).not.toHaveBeenCalled()
  })
})
