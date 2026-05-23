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

function buildSyncPlayer(id: string, camp: 'Red' | 'Blue') {
  return {
    id,
    name: id === 'p1' ? 'Alice' : `Player${id.slice(1)}`,
    camp,
    role: 'hero',
    hand_count: 0,
    max_hand: 6,
    exclusive_card_count: 0,
    hand: [],
    exclusive_cards: [],
    field: [],
    heal: 0,
    max_heal: 5,
    gem: 0,
    crystal: 0,
    is_active: false,
    buffs: [],
    tokens: {},
  }
}

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
      presentation: {
        kind: 'branch_select',
        numeric_base: 0,
      },
    })

    const payload: SyncStatePayload = {
      room_state: 'Playing',
      turn_stage: 'Main',
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
    expect(snapshotStore.turnStage).toBe('Main')
    expect(snapshotStore.players.p1?.name).toBe('Alice')
    expect(snapshotStore.redGems).toBe(2)
    expect(interruptStore.currentPrompt).toBeNull()
  })

  it('pops the current turn player when action stage begins', () => {
    const { handlers, battleFxStore, sessionStore } = buildHandlers()
    sessionStore.setRoomInfo('ROOM1', 'p1', 'Red', 'hero')
    sessionStore.updateRoomPlayers([
      { id: 'p1', name: 'Alice', camp: 'Red', char_role: 'hero', ready: true, is_online: true },
      { id: 'p2', name: 'Bob', camp: 'Red', char_role: 'hero', ready: true, is_online: true },
      { id: 'p3', name: 'Cara', camp: 'Blue', char_role: 'hero', ready: true, is_online: true },
      { id: 'p4', name: 'Dora', camp: 'Blue', char_role: 'hero', ready: true, is_online: true },
      { id: 'p5', name: 'Evan', camp: 'Blue', char_role: 'hero', ready: true, is_online: true },
      { id: 'p6', name: 'Faye', camp: 'Red', char_role: 'hero', ready: true, is_online: true },
    ], 'p1')

    handlers.handleSyncState({
      room_state: 'Playing',
      turn_stage: 'ActionExecution',
      turn_player_id: 'p4',
      has_performed_startup: true,
      morale_red: 15,
      morale_blue: 15,
      cups_red: 0,
      cups_blue: 0,
      stones_red: [0, 0],
      stones_blue: [0, 0],
      deck_count: 30,
      discard_count: 0,
      available_skills: [],
      characters: [],
      players: [
        buildSyncPlayer('p1', 'Red'),
        buildSyncPlayer('p2', 'Red'),
        buildSyncPlayer('p3', 'Blue'),
        buildSyncPlayer('p4', 'Blue'),
        buildSyncPlayer('p5', 'Blue'),
        buildSyncPlayer('p6', 'Red'),
      ],
    })

    expect(battleFxStore.initiatorFocus?.playerId).toBe('p4')
    expect(battleFxStore.initiatorFocus?.side).toBe('right')
  })

  it('routes RequireAction into prompt or waiting state based on target player', () => {
    const { handlers, interruptStore, sessionStore } = buildHandlers()
    sessionStore.setRoomInfo('ROOM1', 'p1', 'Red', 'hero')

    const prompt: Prompt = {
      type: 'confirm',
      player_id: 'p2',
      message: '请选择',
      options: [{ id: 'confirm', label: '确认', button_label: '确认' }],
      min: 1,
      max: 1,
      presentation: {
        kind: 'branch_select',
        numeric_base: 0,
      },
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

  it('focuses the prompt actor when a prompt arrives', () => {
    const { handlers, battleFxStore, sessionStore } = buildHandlers()
    sessionStore.setRoomInfo('ROOM1', 'p1', 'Red', 'hero')

    battleFxStore.startSkillInitiatorFocus('p1', 'skill')
    expect(battleFxStore.initiatorFocus?.playerId).toBe('p1')

    handlers.handleRequireAction({
      interrupt_type: 'WaitChoice',
      target_user_id: 'p1',
      timeout: 0,
      msg: '请选择目标',
      prompt: {
        type: 'confirm',
        player_id: 'p1',
        message: '【挑衅】请选择一名目标对手：',
        choice_type: 'hero_taunt_target',
        options: [{ id: 'p2', label: '目标', button_label: '目标' }],
        min: 1,
        max: 1,
        presentation: {
          kind: 'target_picker',
          numeric_base: 0,
        },
      },
    })

    vi.advanceTimersByTime(1100)

    expect(battleFxStore.initiatorFocus?.playerId).toBe('p1')
    expect(battleFxStore.initiatorFocus?.mode).toBe('skill')
  })

  it('focuses the waiting response player even when the prompt is for another client', () => {
    const { handlers, battleFxStore, interruptStore, sessionStore } = buildHandlers()
    sessionStore.setRoomInfo('ROOM1', 'p1', 'Red', 'hero')

    handlers.handleRequireAction({
      interrupt_type: 'WaitChoice',
      target_user_id: 'p2',
      timeout: 0,
      msg: '请选择应战方式',
      prompt: {
        type: 'confirm',
        player_id: 'p2',
        message: '请选择应战、防御或命中',
        attacker_id: 'p1',
        options: [
          { id: 'counter', label: '应战', button_label: '应战' },
          { id: 'defend', label: '防御', button_label: '防御' },
          { id: 'take', label: '命中', button_label: '命中' },
        ],
        min: 1,
        max: 1,
        presentation: {
          kind: 'branch_select',
          numeric_base: 0,
        },
      },
    })

    expect(interruptStore.currentPrompt).toBeNull()
    expect(interruptStore.waitingFor).toBe('p2')
    expect(battleFxStore.initiatorFocus?.playerId).toBe('p2')
    expect(battleFxStore.initiatorFocus?.mode).toBe('response')
  })

  it('focuses the player who resolves a combat response cue', () => {
    const { handlers, battleFxStore } = buildHandlers()

    handlers.handleGameplayEvent({
      event_type: 'combat_cue',
      attacker_id: 'p1',
      target_id: 'p2',
      phase: 'defend',
    })

    expect(battleFxStore.initiatorFocus?.playerId).toBe('p2')
    expect(battleFxStore.initiatorFocus?.mode).toBe('response')
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
