import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useBattleFxStore } from '../battlefx.store'
import { useBattleReviewStore } from '../battleReview.store'
import { useInterruptStore } from '../interrupt.store'
import { useMatchLifecycleStore } from '../matchLifecycle.store'
import { useSessionStore } from '../session.store'
import { useSnapshotStore } from '../snapshot.store'
import { useUiStore } from '../ui.store'
import type { GameStateUpdate, PlayerView, Prompt } from '../../types/game'

function buildPlayer(overrides: Partial<PlayerView> = {}): PlayerView {
  return {
    id: 'p1',
    name: '勇者',
    camp: 'Red',
    role: 'hero',
    hand_count: 0,
    max_hand: 6,
    exclusive_card_count: 0,
    hand: [],
    blessings: [],
    exclusive_cards: [],
    field: [],
    heal: 0,
    max_heal: 0,
    gem: 0,
    crystal: 0,
    is_active: false,
    buffs: [],
    tokens: {},
    ...overrides,
  }
}

function buildState(overrides: Partial<GameStateUpdate> = {}): GameStateUpdate {
  return {
    phase: 'Main',
    current_player: 'p1',
    has_performed_startup: false,
    players: {
      p1: buildPlayer(),
    },
    red_morale: 15,
    blue_morale: 15,
    red_cups: 0,
    blue_cups: 0,
    red_gems: 0,
    blue_gems: 0,
    red_crystals: 0,
    blue_crystals: 0,
    deck_count: 30,
    discard_count: 0,
    available_skills: [],
    ...overrides,
  }
}

describe('useMatchLifecycleStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.runOnlyPendingTimers()
    vi.useRealTimers()
    vi.restoreAllMocks()
  })

  it('clears morale review and end overlay when a new match starts', () => {
    const lifecycleStore = useMatchLifecycleStore()
    const sessionStore = useSessionStore()
    const battleReviewStore = useBattleReviewStore()
    const uiStore = useUiStore()

    battleReviewStore.pushMoraleHint({
      source: '测试提示',
      raw: '[T] 红方士气-1',
      camp: 'Red',
      loss: 1,
    })
    battleReviewStore.recordMoraleChange('Red', 5, 4, {
      id: 1,
      timestamp: 1,
      source: '测试掉血',
      raw: '[T]',
      camp: 'Red',
      loss: 1,
    })
    uiStore.setGameEnded(true, '旧结束文案', null)

    lifecycleStore.setGameStarted()

    expect(sessionStore.gameStarted).toBe(true)
    expect(uiStore.isGameEnded).toBe(false)
    expect(uiStore.gameEndMessage).toBe('')
    expect(battleReviewStore.moraleHints).toHaveLength(0)
    expect(battleReviewStore.moraleChanges).toHaveLength(0)
  })

  it('builds the end snapshot from latest state and clears transient ui', () => {
    const lifecycleStore = useMatchLifecycleStore()
    const sessionStore = useSessionStore()
    const snapshotStore = useSnapshotStore()
    const interruptStore = useInterruptStore()
    const battleFxStore = useBattleFxStore()
    const battleReviewStore = useBattleReviewStore()
    const uiStore = useUiStore()

    sessionStore.setRoomInfo('ROOM1', 'p1', 'Red', 'hero')
    snapshotStore.updateGameState(buildState({
      red_morale: 0,
      blue_morale: 7,
    }))
    battleReviewStore.recordMoraleChange('Red', 2, 0, {
      id: 1,
      timestamp: 10,
      source: '绝杀',
      raw: '[KO]',
      camp: 'Red',
      loss: 2,
    })

    const prompt: Prompt = {
      type: 'confirm',
      player_id: 'p1',
      message: '请选择',
      options: [{ id: 'ok', label: '确定' }],
      min: 1,
      max: 1,
    }
    interruptStore.setPrompt(prompt)
    battleFxStore.startSkillInitiatorFocus('p1', 'skill')

    lifecycleStore.setGameEnded('蓝方胜利！红方士气归零')

    expect(uiStore.isGameEnded).toBe(true)
    expect(uiStore.gameEndMessage).toBe('蓝方胜利！红方士气归零')
    expect(uiStore.gameEndSnapshot).toMatchObject({
      triggerType: 'morale',
      finalRedMorale: 0,
      finalBlueMorale: 7,
      triggerCamp: 'Red',
      triggerDelta: 2,
      triggerSource: '绝杀',
    })
    expect(interruptStore.currentPrompt).toBeNull()
    expect(interruptStore.waitingFor).toBe('')
    expect(battleFxStore.initiatorFocus).toBeNull()
  })

  it('refreshes the end snapshot after a late final state sync', () => {
    const lifecycleStore = useMatchLifecycleStore()
    const sessionStore = useSessionStore()
    const snapshotStore = useSnapshotStore()
    const battleReviewStore = useBattleReviewStore()
    const uiStore = useUiStore()

    sessionStore.setRoomInfo('ROOM2', 'p1', 'Red', 'hero')
    uiStore.setGameEnded(true, '旧文案', null)
    snapshotStore.updateGameState(buildState({
      red_morale: 0,
      blue_morale: 9,
      red_cups: 1,
      blue_cups: 2,
    }))
    battleReviewStore.recordMoraleChange('Red', 3, 0, {
      id: 1,
      timestamp: 20,
      source: '补发终局状态',
      raw: '[SYNC]',
      camp: 'Red',
      loss: 3,
    })

    const snapshot = lifecycleStore.refreshGameEndSnapshot('蓝方胜利！红方士气归零')

    expect(snapshot).toMatchObject({
      triggerType: 'morale',
      finalRedMorale: 0,
      finalBlueMorale: 9,
      finalRedCups: 1,
      finalBlueCups: 2,
      triggerSource: '补发终局状态',
    })
    expect(uiStore.gameEndMessage).toBe('蓝方胜利！红方士气归零')
    expect(uiStore.gameEndSnapshot).toEqual(snapshot)
  })
})
