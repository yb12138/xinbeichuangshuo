import { render, screen } from '@testing-library/vue'
import userEvent from '@testing-library/user-event'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import PromptDialog from '../PromptDialog.vue'
import { useInterruptStore } from '../../stores/interrupt.store'
import { useSessionStore } from '../../stores/session.store'
import { useSnapshotStore } from '../../stores/snapshot.store'
import type { Card, GameStateUpdate, PlayerView, Prompt } from '../../types/game'

const submitSelectMock = vi.fn()

vi.mock('../../composables/useSubmitAction', () => ({
  useSubmitAction: () => ({
    submitSelect: submitSelectMock,
    submitCancel: vi.fn(),
    submitConfirm: vi.fn(),
    submitRespondTake: vi.fn(),
    submitRespondCounter: vi.fn(),
    submitRespondDefend: vi.fn(),
    submitAction: vi.fn(),
  }),
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
    name: 'P1',
    camp: 'Red',
    role: 'fighter',
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
    turn_stage: 'ActionExecution',
    current_player: 'p3',
    has_performed_startup: false,
    players: {
      p2: buildPlayer({
        id: 'p2',
        name: 'P2',
        camp: 'Blue',
        heal: 2,
        max_heal: 5,
        hand_count: 3,
        hand: [
          buildCard({ id: 'h0' }),
          buildCard({ id: 'h1', element: 'Water' }),
          buildCard({ id: 'h2', element: 'Thunder' }),
        ],
      }),
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

function healPrompt(): Prompt {
  return {
    type: 'confirm',
    player_id: 'p2',
    choice_type: 'heal',
    message: 'P2 受到伤害，可选择使用治疗抵消：',
    options: [
      { id: '0', label: '不使用治疗' },
      { id: '1', label: '使用 1 点治疗' },
      { id: '2', label: '使用 2 点治疗' },
    ],
    min: 1,
    max: 1,
    presentation: {
      kind: 'numeric',
      numeric_base: 0,
    },
  }
}

describe('PromptDialog', () => {
  beforeEach(() => {
    submitSelectMock.mockReset()
  })

  it('renders heal mitigation as a numeric decision instead of a hand-card picker', async () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    useSessionStore().setRoomInfo('ROOM1', 'p2', 'Blue', 'fighter')
    useSnapshotStore().updateGameState(buildState())
    useInterruptStore().setPrompt(healPrompt())

    render(PromptDialog, {
      global: {
        plugins: [pinia],
      },
    })

    expect(screen.getByTestId('decision-overlay')).toBeInTheDocument()
    expect(screen.getByTestId('numeric-option-0')).toBeInTheDocument()
    expect(screen.getByTestId('numeric-option-2')).toBeInTheDocument()
    expect(screen.queryByText('完成选牌后点击发动')).not.toBeInTheDocument()

    await userEvent.click(screen.getByTestId('numeric-option-2'))

    expect(submitSelectMock).toHaveBeenCalledWith([2])
  })
})
