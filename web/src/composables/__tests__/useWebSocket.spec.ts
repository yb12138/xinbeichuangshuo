import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useWebSocket } from '../useWebSocket'
import { useBattleReviewStore } from '../../stores/battleReview.store'
import { useSessionStore } from '../../stores/session.store'
import { useSnapshotStore } from '../../stores/snapshot.store'
import { useTimelineStore } from '../../stores/timeline.store'
import { useInterruptStore } from '../../stores/interrupt.store'
import type { WsMessage } from '../../network/protocol'

class FakeStorage {
  private data = new Map<string, string>()

  getItem(key: string) {
    return this.data.get(key) ?? null
  }

  setItem(key: string, value: string) {
    this.data.set(key, value)
  }
}

class FakeWebSocket {
  static OPEN = 1
  static CLOSED = 3
  static instances: FakeWebSocket[] = []

  url: string
  readyState = 0
  sent: string[] = []
  onopen: WebSocket['onopen'] = null
  onmessage: WebSocket['onmessage'] = null
  onerror: WebSocket['onerror'] = null
  onclose: WebSocket['onclose'] = null

  constructor(url: string) {
    this.url = url
    FakeWebSocket.instances.push(this)
  }

  send(data: string) {
    this.sent.push(data)
  }

  close() {
    this.readyState = FakeWebSocket.CLOSED
    this.onclose?.call(this as unknown as WebSocket, new CloseEvent('close'))
  }

  open() {
    this.readyState = FakeWebSocket.OPEN
    this.onopen?.call(this as unknown as WebSocket, new Event('open'))
  }

  receive(message: WsMessage) {
    this.onmessage?.call(
      this as unknown as WebSocket,
      new MessageEvent('message', { data: JSON.stringify(message) }),
    )
  }
}

describe('useWebSocket integration', () => {
  let actions: ReturnType<typeof useWebSocket> | null = null
  let storage: FakeStorage

  beforeEach(() => {
    setActivePinia(createPinia())
    vi.useFakeTimers()
    vi.stubEnv('VITE_WS_URL', 'ws://example.test/ws')
    storage = new FakeStorage()
    FakeWebSocket.instances = []
    vi.stubGlobal('WebSocket', FakeWebSocket)
    vi.stubGlobal('window', {
      location: {
        protocol: 'http:',
        hostname: 'localhost',
        host: 'localhost:5173',
      },
      localStorage: storage,
    })
    actions = useWebSocket()
  })

  afterEach(() => {
    actions?.disconnect()
    actions = null
    vi.runOnlyPendingTimers()
    vi.useRealTimers()
    vi.unstubAllEnvs()
    vi.unstubAllGlobals()
  })

  it('connects with stored reconnect info, routes inbound messages into stores, and sends envelopes', () => {
    const sessionStore = useSessionStore()
    const snapshotStore = useSnapshotStore()
    const timelineStore = useTimelineStore()
    const battleReviewStore = useBattleReviewStore()

    storage.setItem('xbs_reconnect_ROOM1_Alice', JSON.stringify({
      room_code: 'ROOM1',
      player_id: 'p9',
      player_name: 'Alice',
      token: 'saved-token',
    }))

    actions!.connect('ROOM1', 'Alice')
    const socket = FakeWebSocket.instances[0]
    expect(socket?.url).toBe(
      'ws://example.test/ws?room=ROOM1&name=Alice&player_id=p9&reconnect_token=saved-token'
    )

    socket?.open()
    socket?.receive({
      Cmd: 'RoomEvent',
      Data: {
        action: 'assigned',
        room_code: 'ROOM1',
        player_id: 'p1',
        camp: 'Red',
        char_role: 'hero',
        reconnect_token: 'token-1',
      },
    })
    socket?.receive({
      Cmd: 'SyncState',
      Data: {
        room_state: 'Playing',
        turn_stage: 'Main',
        turn_player_id: 'p1',
        has_performed_startup: false,
        morale_red: 15,
        morale_blue: 14,
        cups_red: 0,
        cups_blue: 1,
        stones_red: [2, 1],
        stones_blue: [1, 3],
        deck_count: 20,
        discard_count: 2,
        available_skills: [],
        characters: [],
        players: [
          {
            id: 'p1',
            name: 'Alice',
            camp: 'Red',
            role: 'hero',
            hand_count: 1,
            max_hand: 6,
            exclusive_card_count: 0,
            hand: [],
            exclusive_cards: [],
            field: [],
            heal: 3,
            max_heal: 5,
            gem: 1,
            crystal: 0,
            is_active: true,
            buffs: [],
            tokens: {},
          },
        ],
      },
    })
    socket?.receive({
      Cmd: 'NotifyTimeline',
      Data: {
        room_id: 'ROOM1',
        seq_start: 1,
        seq_end: 1,
        is_replay: false,
        events: [
          {
            event_id: 1,
            turn_id: 1,
            chain_id: 'chain_1',
            type: 'TimelineCombatResolved',
            outcome: 'TimelineOutcomeSuccess',
            visibility: 'TimelineVisibilityPublic',
            actor_user_id: 'p1',
            actor_name: 'Alice',
            target_user_ids: ['p2'],
            target_name: 'Bob',
            damage: 2,
            damage_type: 'Attack',
            message: '造成 2 点伤害',
            gameplay_type: 'damage_dealt',
          },
        ],
      },
    })
    socket?.receive({
      Cmd: 'ProtocolError',
      Data: {
        code: 'unknown_cmd',
        message: '未知命令',
        cmd: 'RoomAction',
      },
    })

    expect(sessionStore.myPlayerId).toBe('p1')
    expect(sessionStore.reconnectToken).toBe('token-1')
    expect(sessionStore.gameStarted).toBe(true)
    expect(snapshotStore.turnStage).toBe('Main')
    expect(snapshotStore.players.p1?.name).toBe('Alice')
    expect(timelineStore.entries).toHaveLength(1)
    expect(storage.getItem('xbs_reconnect_ROOM1_Alice')).toContain('token-1')
    expect(useInterruptStore().errorMessage).toBe('未知命令')
    expect(battleReviewStore.logs).toContain('[WS][ProtocolError] unknown_cmd: 未知命令')

    actions!.sendChat('hello')
    expect(socket?.sent).toContain('{"Cmd":"ChatMessage","Data":{"message":"hello"}}')
    expect(battleReviewStore.logs).toContain('[WS] 连接成功')
  })

  it('disconnect resets state without creating a new reconnect attempt', () => {
    const sessionStore = useSessionStore()

    actions!.connect('ROOM2', 'Bob')
    const socket = FakeWebSocket.instances[0]
    socket?.open()
    sessionStore.setRoomInfo('ROOM2', 'p2', 'Blue', 'mage')

    actions!.disconnect()
    vi.runAllTimers()

    expect(sessionStore.isInRoom).toBe(false)
    expect(sessionStore.isConnected).toBe(false)
    expect(FakeWebSocket.instances).toHaveLength(1)
  })
})
