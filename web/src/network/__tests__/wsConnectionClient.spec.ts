import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { createWsConnectionClient } from '../wsConnectionClient'
import { useBattleReviewStore } from '../../stores/battleReview.store'
import { useInterruptStore } from '../../stores/interrupt.store'
import { useMatchLifecycleStore } from '../../stores/matchLifecycle.store'
import { useSessionStore } from '../../stores/session.store'
import type { RoutedWsMessage } from '../protocol'

class FakeSocket {
  url: string
  readyState = 0
  sent: string[] = []
  onopen: WebSocket['onopen'] = null
  onmessage: WebSocket['onmessage'] = null
  onerror: WebSocket['onerror'] = null
  onclose: WebSocket['onclose'] = null

  constructor(url: string) {
    this.url = url
  }

  send(data: string) {
    this.sent.push(data)
  }

  close() {
    this.readyState = 3
    this.onclose?.call(this as unknown as WebSocket, new CloseEvent('close'))
  }

  open() {
    this.readyState = 1
    this.onopen?.call(this as unknown as WebSocket, new Event('open'))
  }

  receive(message: unknown) {
    this.onmessage?.call(
      this as unknown as WebSocket,
      new MessageEvent('message', { data: JSON.stringify(message) }),
    )
  }

  fail() {
    this.onerror?.call(this as unknown as WebSocket, new Event('error'))
  }

  remoteClose() {
    this.readyState = 3
    this.onclose?.call(this as unknown as WebSocket, new CloseEvent('close'))
  }
}

let activeClient: ReturnType<typeof createWsConnectionClient> | null = null

function buildClient(options?: {
  loadReconnectInfo?: (roomCode: string, playerName: string) => { player_id: string; token: string; room_code: string; player_name: string } | null
}) {
  const interruptStore = useInterruptStore()
  const sessionStore = useSessionStore()
  const battleReviewStore = useBattleReviewStore()
  const matchLifecycleStore = useMatchLifecycleStore()
  const handledMessages: RoutedWsMessage[] = []
  const sockets: FakeSocket[] = []
  const createSocket = vi.fn((url: string) => {
    const socket = new FakeSocket(url)
    sockets.push(socket)
    return socket as unknown as WebSocket
  })

  const client = createWsConnectionClient({
    sessionStore,
    interruptStore,
    battleReviewStore,
    matchLifecycleStore,
    getWsUrl: () => 'ws://example.test/ws',
    loadReconnectInfo: options?.loadReconnectInfo,
    createSocket,
    safeStringify: (data) => JSON.stringify(data),
    onMessage: (msg) => {
      handledMessages.push(msg)
    },
  })

  activeClient = client

  return {
    client,
    interruptStore,
    sessionStore,
    battleReviewStore,
    createSocket,
    sockets,
    handledMessages,
  }
}

describe('createWsConnectionClient', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.useFakeTimers()
  })

  afterEach(() => {
    activeClient?.disconnect()
    activeClient = null
    vi.runOnlyPendingTimers()
    vi.useRealTimers()
  })

  it('prefers session reconnect credentials over persisted storage when reconnecting', () => {
    const { client, createSocket, sessionStore, sockets } = buildClient({
      loadReconnectInfo: () => ({
        room_code: 'ROOM1',
        player_id: 'stored-player',
        player_name: 'Alice',
        token: 'stored-token',
      }),
    })
    sessionStore.setRoomInfo('ROOM1', 'p1', 'Red', 'hero')
    sessionStore.setReconnectToken('session-token')

    client.connect('ROOM1', 'Alice')
    sockets[0]?.open()

    expect(createSocket).toHaveBeenCalledWith(
      'ws://example.test/ws?room=ROOM1&name=Alice&player_id=p1&reconnect_token=session-token'
    )
    expect(sessionStore.isConnected).toBe(true)
  })

  it('routes inbound messages, serializes outbound envelopes and surfaces socket errors', () => {
    const { client, interruptStore, battleReviewStore, sockets, handledMessages } = buildClient()

    client.connect('ROOM1', 'Alice')
    sockets[0]?.open()
    client.sendEnvelope({ Cmd: 'ChatMessage', Data: { message: 'hello' } })
    sockets[0]?.receive({ Cmd: 'NotifyTimeline', Data: { room_id: 'ROOM1' } })
    sockets[0]?.receive({ Data: { room_id: 'ROOM1' } })
    sockets[0]?.fail()

    expect(sockets[0]?.sent).toEqual([
      '{"Cmd":"ChatMessage","Data":{"message":"hello"}}',
    ])
    expect(handledMessages).toEqual([
      { Cmd: 'NotifyTimeline', Data: { room_id: 'ROOM1' } },
    ])
    expect(interruptStore.errorMessage).toBe('连接错误')
    expect(battleReviewStore.logs).toContain('[WS][RX] invalid envelope')
    expect(battleReviewStore.logs).toContain('[WS] 连接错误')
  })

  it('retries connection after unexpected close while the session is still in room', () => {
    const { client, createSocket, sessionStore, sockets } = buildClient()
    sessionStore.setRoomInfo('ROOM1', 'p1', 'Red', 'hero')
    sessionStore.setReconnectToken('session-token')

    client.connect('ROOM1', 'Alice')
    sockets[0]?.open()
    sockets[0]?.remoteClose()

    expect(sessionStore.isConnected).toBe(false)
    expect(createSocket).toHaveBeenCalledTimes(1)

    vi.advanceTimersByTime(2000)

    expect(createSocket).toHaveBeenCalledTimes(2)
    expect(createSocket).toHaveBeenLastCalledWith(
      'ws://example.test/ws?room=ROOM1&name=Alice&player_id=p1&reconnect_token=session-token'
    )
  })

  it('manual disconnect closes the socket without scheduling reconnects', () => {
    const { client, createSocket, sessionStore, sockets } = buildClient()
    sessionStore.setRoomInfo('ROOM1', 'p1', 'Red', 'hero')

    client.connect('ROOM1', 'Alice')
    sockets[0]?.open()
    client.disconnect()
    vi.runAllTimers()

    expect(createSocket).toHaveBeenCalledTimes(1)
    expect(sessionStore.isInRoom).toBe(false)
    expect(sessionStore.isConnected).toBe(false)
  })
})
