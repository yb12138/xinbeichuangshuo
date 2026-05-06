import { beforeEach, describe, expect, it } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useGameStore } from '../gameStore'
import { useInterruptStore } from '../interrupt.store'
import { useSessionStore } from '../session.store'
import { useTimelineStore } from '../timeline.store'
import type { GameStateUpdate, PlayerView, Prompt } from '../../types/game'

function buildPlayer(overrides: Partial<PlayerView> = {}): PlayerView {
  return {
    id: 'p1',
    name: 'Alice',
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
    turn_stage: 'Main',
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

describe('useGameStore legacy facade', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('reuses timeline.store payload history instead of maintaining a duplicate buffer', () => {
    const gameStore = useGameStore()
    const timelineStore = useTimelineStore()

    expect(gameStore.timelinePayloads).toBe(timelineStore.payloads)

    gameStore.pushTimelinePayload({
      room_id: 'ROOM1',
      seq_start: 1,
      seq_end: 1,
      is_replay: false,
      events: [],
    })

    expect(timelineStore.payloads).toHaveLength(1)
    timelineStore.clear()
    expect(gameStore.timelinePayloads).toHaveLength(0)
  })

  it('keeps compatibility side effects when syncing state and prompt selection', () => {
    const gameStore = useGameStore()
    const sessionStore = useSessionStore()
    const interruptStore = useInterruptStore()

    sessionStore.setRoomInfo('ROOM1', 'p1', '', '')
    interruptStore.setPrompt({
      type: 'confirm',
      player_id: 'p1',
      message: '旧提示',
      options: [],
      min: 1,
      max: 1,
    })

    gameStore.updateGameState(buildState({
      players: {
        p1: buildPlayer({
          id: 'p1',
          camp: 'Blue',
          role: 'mage',
        }),
      },
    }))

    expect(sessionStore.myCamp).toBe('Blue')
    expect(sessionStore.myCharRole).toBe('mage')
    expect(interruptStore.currentPrompt).toBeNull()

    const prompt: Prompt = {
      type: 'confirm',
      player_id: 'p1',
      message: '请选择',
      options: [{ id: 'ok', label: '确定' }],
      min: 1,
      max: 1,
    }

    gameStore.setPrompt(prompt)
    expect(interruptStore.currentPrompt).toEqual(prompt)
  })

  it('preserves the old action-mode helper behavior', () => {
    const gameStore = useGameStore()
    const interruptStore = useInterruptStore()

    interruptStore.setSelectedCardForAction(3)
    gameStore.setActionModeForAttack('none')

    expect(interruptStore.actionMode).toBe('none')
    expect(interruptStore.selectedCardForAction).toBeNull()
  })
})
