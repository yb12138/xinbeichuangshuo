import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { createGameplayMessageHandlers } from '../gameplayMessageHandlers'
import { useBattleFxStore } from '../../stores/battlefx.store'
import { useBattleReviewStore } from '../../stores/battleReview.store'
import { useInterruptStore } from '../../stores/interrupt.store'
import { useMatchLifecycleStore } from '../../stores/matchLifecycle.store'
import { useSessionStore } from '../../stores/session.store'
import { useSnapshotStore } from '../../stores/snapshot.store'
import { useTimelineStore } from '../../stores/timeline.store'
import { useUiStore } from '../../stores/ui.store'
import type { Prompt } from '../../types/game'
import type { RequireActionPayload, SyncStatePayload, TimelineNotifyPayload } from '../protocol'

function buildHandlers() {
  const interruptStore = useInterruptStore()
  const sessionStore = useSessionStore()
  const snapshotStore = useSnapshotStore()
  const timelineStore = useTimelineStore()
  const uiStore = useUiStore()
  const battleFxStore = useBattleFxStore()
  const battleReviewStore = useBattleReviewStore()
  const matchLifecycleStore = useMatchLifecycleStore()

  return {
    interruptStore,
    sessionStore,
    snapshotStore,
    timelineStore,
    uiStore,
    battleFxStore,
    battleReviewStore,
    matchLifecycleStore,
    handlers: createGameplayMessageHandlers({
      interruptStore,
      sessionStore,
      snapshotStore,
      timelineStore,
      uiStore,
      battleFxStore,
      battleReviewStore,
      matchLifecycleStore,
    }),
  }
}

describe('createGameplayMessageHandlers', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.runOnlyPendingTimers()
    vi.useRealTimers()
  })

  it('applies SyncState to snapshot and marks the match as started', () => {
    const { handlers, interruptStore, sessionStore, snapshotStore } = buildHandlers()
    sessionStore.setRoomInfo('ROOM1', 'p1', 'Red', 'hero')
    interruptStore.setPrompt({
      type: 'confirm',
      player_id: 'p1',
      message: '旧提示',
      options: [],
      min: 1,
      max: 1,
    })

    const payload: SyncStatePayload = {
      room_state: 'Playing',
      current_phase: 'Main',
      turn_player_id: 'p1',
      has_performed_startup: true,
      morale_red: 15,
      morale_blue: 13,
      cups_red: 0,
      cups_blue: 1,
      stones_red: [2, 1],
      stones_blue: [1, 3],
      deck_count: 18,
      discard_count: 4,
      available_skills: [],
      characters: [],
      players: [
        {
          id: 'p1',
          name: 'Alice',
          camp: 'Red',
          role: 'hero',
          hand_count: 2,
          max_hand: 6,
          exclusive_card_count: 0,
          hand: [],
          blessings: [],
          exclusive_cards: [],
          field: [],
          heal: 1,
          max_heal: 5,
          gem: 1,
          crystal: 2,
          is_active: true,
          buffs: [],
          tokens: {},
        },
      ],
    }

    handlers.handleSyncState(payload)

    expect(sessionStore.gameStarted).toBe(true)
    expect(snapshotStore.phase).toBe('Main')
    expect(snapshotStore.players.p1?.name).toBe('Alice')
    expect(snapshotStore.redGems).toBe(2)
    expect(interruptStore.currentPrompt).toBeNull()
  })

  it('routes RequireAction into prompt or waiting state based on target player', () => {
    const { handlers, interruptStore, sessionStore } = buildHandlers()
    sessionStore.setRoomInfo('ROOM1', 'p1', 'Red', 'hero')

    const prompt: Prompt = {
      type: 'confirm',
      player_id: 'p2',
      message: '请选择',
      options: [{ id: 'confirm', label: '确认' }],
      min: 1,
      max: 1,
    }

    const waitingPayload: RequireActionPayload = {
      interrupt_type: 'WaitChoice',
      target_user_id: 'p2',
      timeout: 0,
      msg: '等待对手操作',
      prompt,
    }

    handlers.handleRequireAction(waitingPayload)

    expect(interruptStore.currentPrompt).toBeNull()
    expect(interruptStore.waitingFor).toBe('p2')

    handlers.handleRequireAction({
      ...waitingPayload,
      target_user_id: 'p1',
      prompt: { ...prompt, player_id: 'p1' },
    })

    expect(interruptStore.currentPrompt?.player_id).toBe('p1')
    expect(interruptStore.waitingFor).toBe('')
  })

  it('replays NotifyTimeline payloads into timeline entries and battle effects', () => {
    const { handlers, timelineStore, battleFxStore } = buildHandlers()

    const payload: TimelineNotifyPayload = {
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
          gameplay_type: 'damage_dealt',
          actor_user_id: 'p1',
          actor_name: 'Alice',
          target_user_ids: ['p2'],
          target_name: 'Bob',
          damage: 3,
          damage_type: 'Attack',
          message: '造成 3 点伤害',
        },
      ],
    }

    handlers.handleNotifyTimeline(payload)

    expect(timelineStore.entries).toHaveLength(1)
    expect(timelineStore.entries[0]?.gameplayType).toBe('damage_dealt')
    expect(battleFxStore.damageEffects).toHaveLength(1)
    expect(battleFxStore.damageEffects[0]?.targetId).toBe('p2')
    expect(battleFxStore.damageEffects[0]?.damage).toBe(3)
  })
})
