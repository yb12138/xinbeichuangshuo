import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useInterruptStore } from '../../stores/interrupt.store'
import { useSessionStore } from '../../stores/session.store'
import { useSnapshotStore } from '../../stores/snapshot.store'
import { useSubmitAction } from '../useSubmitAction'
import type { Card, GameStateUpdate, PlayerView } from '../../types/game'

const wsMock = {
  disconnect: vi.fn(),
  sendAction: vi.fn(),
  sendRoomAction: vi.fn(),
  sendChat: vi.fn(),
  attack: vi.fn(),
  magic: vi.fn(),
  useSkill: vi.fn(),
  pass: vi.fn(),
  confirm: vi.fn(),
  cancel: vi.fn(),
  select: vi.fn(),
  respond: vi.fn(),
  buy: vi.fn(),
  extract: vi.fn(),
  cheatSkill: vi.fn(),
  cheatToken: vi.fn(),
  cheatSet: vi.fn(),
  cheatEffect: vi.fn(),
  cheatGiveExclusive: vi.fn(),
  cheatGiveByElement: vi.fn(),
  cheatGiveByFaction: vi.fn(),
  cheatGiveMagicByName: vi.fn(),
}

vi.mock('../useWebSocket', () => ({
  useWebSocket: () => wsMock,
}))

function buildCard(overrides: Partial<Card> = {}): Card {
  return {
    id: 'card-1',
    name: '烈焰斩',
    type: 'Attack',
    element: 'Fire',
    damage: 2,
    description: 'test card',
    ...overrides,
  }
}

function buildPlayer(overrides: Partial<PlayerView> = {}): PlayerView {
  return {
    id: 'p1',
    name: '我方',
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

describe('useSubmitAction', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    for (const value of Object.values(wsMock)) {
      value.mockReset()
    }

    const sessionStore = useSessionStore()
    const snapshotStore = useSnapshotStore()

    sessionStore.setRoomInfo('ROOM1', 'p1', 'Red', 'hero')
    snapshotStore.updateGameState(buildState({
      players: {
        p1: buildPlayer({
          id: 'p1',
          name: '我方',
          camp: 'Red',
          hand_count: 2,
          hand: [
            buildCard({ id: 'attack-1', name: '烈焰斩', type: 'Attack' }),
            buildCard({ id: 'magic-1', name: '魔弹', type: 'Magic', element: 'Dark' }),
          ],
        }),
        p2: buildPlayer({
          id: 'p2',
          name: '敌方',
          camp: 'Blue',
          role: 'enemy',
        }),
      },
    }))
  })

  it('shows an error when countering without a selected card', () => {
    const interruptStore = useInterruptStore()
    const actions = useSubmitAction()

    const ok = actions.submitRespondCounter(false)

    expect(ok).toBe(false)
    expect(interruptStore.errorMessage).toBe('请先选择一张应战牌')
    expect(wsMock.respond).not.toHaveBeenCalled()
  })

  it('clears stale board-card selection before sending an action', () => {
    const interruptStore = useInterruptStore()
    const actions = useSubmitAction()

    interruptStore.setActionMode('attack')
    interruptStore.setSelectedCardForAction(99)

    const ok = actions.submitSelectedBoardTarget('p2')

    expect(ok).toBe(false)
    expect(interruptStore.selectedCardForAction).toBeNull()
    expect(interruptStore.errorMessage).toBe('所选卡牌已变化，请重新选择')
    expect(wsMock.attack).not.toHaveBeenCalled()
  })

  it('sends the selected attack card to the clicked target', () => {
    const interruptStore = useInterruptStore()
    const actions = useSubmitAction()

    interruptStore.setActionMode('attack')
    interruptStore.setSelectedCardForAction(0)

    const ok = actions.submitSelectedBoardTarget('p2')

    expect(ok).toBe(true)
    expect(wsMock.attack).toHaveBeenCalledWith('p2', 0)
  })
})
