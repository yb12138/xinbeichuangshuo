import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { createRoomMessageHandlers } from '../roomMessageHandlers'
import { useBattleReviewStore } from '../../stores/battleReview.store'
import { useInterruptStore } from '../../stores/interrupt.store'
import { useMatchLifecycleStore } from '../../stores/matchLifecycle.store'
import { useSessionStore } from '../../stores/session.store'
import { useSnapshotStore } from '../../stores/snapshot.store'

function buildRoomHandlers() {
  const sessionStore = useSessionStore()
  const snapshotStore = useSnapshotStore()
  const interruptStore = useInterruptStore()
  const battleReviewStore = useBattleReviewStore()
  const matchLifecycleStore = useMatchLifecycleStore()
  const closeTransport = vi.fn()
  const persistReconnectInfo = vi.fn()

  return {
    sessionStore,
    snapshotStore,
    interruptStore,
    battleReviewStore,
    matchLifecycleStore,
    closeTransport,
    persistReconnectInfo,
    handlers: createRoomMessageHandlers({
      sessionStore,
      snapshotStore,
      interruptStore,
      battleReviewStore,
      matchLifecycleStore,
      closeTransport,
      persistReconnectInfo,
    }),
  }
}

describe('createRoomMessageHandlers', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.runOnlyPendingTimers()
    vi.useRealTimers()
  })

  it('handles assigned events by updating seat info, characters and reconnect token', () => {
    const { handlers, sessionStore, snapshotStore, battleReviewStore, persistReconnectInfo } = buildRoomHandlers()
    sessionStore.setMyName('Alice')

    handlers.handleRoomEvent({
      action: 'assigned',
      room_code: 'ROOM1',
      player_id: 'p1',
      camp: 'Red',
      char_role: 'hero',
      reconnect_token: 'token-1',
      characters: [
        {
          id: 'hero',
          name: '英雄',
          title: '测试角色',
          faction: 'fire',
          skills: [],
        },
      ],
    })

    expect(sessionStore.roomCode).toBe('ROOM1')
    expect(sessionStore.myPlayerId).toBe('p1')
    expect(sessionStore.reconnectToken).toBe('token-1')
    expect(snapshotStore.characters.hero?.name).toBe('英雄')
    expect(battleReviewStore.logs[battleReviewStore.logs.length - 1]).toBe('已加入房间 ROOM1，你是 p1')
    expect(persistReconnectInfo).toHaveBeenCalledWith('ROOM1', 'p1', 'token-1')
  })

  it('handles player_list events by refreshing room players and characters', () => {
    const { handlers, sessionStore, snapshotStore } = buildRoomHandlers()
    sessionStore.setRoomInfo('ROOM1', 'p1', 'Red', 'hero')

    handlers.handleRoomEvent({
      action: 'player_list',
      room_code: 'ROOM1',
      players: [
        { id: 'p1', name: 'Alice', camp: 'Red', char_role: 'hero', ready: true },
        { id: 'p2', name: 'Bob', camp: 'Blue', char_role: 'angel', ready: false },
      ],
      characters: [
        {
          id: 'angel',
          name: '天使',
          title: '测试角色',
          faction: 'light',
          skills: [],
        },
      ],
    })

    expect(sessionStore.roomPlayers).toHaveLength(2)
    expect(sessionStore.myCamp).toBe('Red')
    expect(snapshotStore.characters.angel?.name).toBe('天使')
  })

  it('handles started events by marking the match as started and logging it', () => {
    const { handlers, sessionStore, battleReviewStore } = buildRoomHandlers()

    handlers.handleRoomEvent({
      action: 'started',
      room_code: 'ROOM1',
    })

    expect(sessionStore.gameStarted).toBe(true)
    expect(battleReviewStore.logs[battleReviewStore.logs.length - 1]).toBe('游戏开始！')
  })

  it('handles dissolved events by closing transport, resetting state and surfacing the error', () => {
    const { handlers, sessionStore, interruptStore, battleReviewStore, closeTransport } = buildRoomHandlers()
    sessionStore.setRoomInfo('ROOM1', 'p1', 'Red', 'hero')
    battleReviewStore.addLog('some log')

    handlers.handleRoomEvent({
      action: 'dissolved',
      room_code: 'ROOM1',
      message: '房间已解散',
    })

    expect(closeTransport).toHaveBeenCalledTimes(1)
    expect(sessionStore.isInRoom).toBe(false)
    expect(battleReviewStore.logs).toEqual([])
    expect(interruptStore.errorMessage).toBe('房间已解散')
  })
})
