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

function medusaDarkMoonPickPrompt(): Prompt {
  return {
    type: 'confirm',
    player_id: 'p2',
    choice_type: 'mg_medusa_darkmoon_pick',
    message: '【美杜莎之眼】请选择要展示并移除的同系闇月：',
    options: [
      { id: '0', label: '移除闇月[暗月法术/Magic/Dark]', field_index: 0 },
      { id: '1', label: '移除闇月[火焰斩/Attack/Fire]', field_index: 1 },
    ],
    min: 1,
    max: 1,
    presentation: {
      kind: 'card_picker',
      layout: 'field_cover',
    },
  }
}

function moonCycleBranchPrompt(): Prompt {
  return {
    type: 'confirm',
    player_id: 'p2',
    choice_type: 'mg_moon_cycle_mode',
    message: '【月之轮回】请选择发动分支：',
    cancelable: true,
    options: [
      { id: 'decline', label: '不发动' },
      { id: 'branch1', label: '分支①：移除1个闇月，令目标角色+1治疗' },
      { id: 'branch2', label: '分支②：移除1点治疗，你+1新月' },
    ],
    min: 1,
    max: 1,
    presentation: {
      kind: 'branch_select',
      layout: 'overlay',
    },
  }
}

function weakPrompt(): Prompt {
  return {
    type: 'confirm',
    player_id: 'p2',
    choice_type: 'weak',
    message: '【虚弱状态】测试玩家1，你需要做出选择：',
    options: [
      { id: 'draw_continue', label: '摸3张牌继续执行后续行动' },
      { id: 'skip_turn', label: '跳过此回合' },
    ],
    min: 1,
    max: 1,
    presentation: {
      kind: 'branch_select',
      layout: 'overlay',
    },
  }
}

function moonCycleTargetPrompt(): Prompt {
  return {
    type: 'confirm',
    player_id: 'p2',
    choice_type: 'mg_moon_cycle_heal_target',
    message: '【月之轮回】请选择获得1点治疗的角色：',
    options: [
      { id: 'p3', label: '目标玩家' },
    ],
    min: 1,
    max: 1,
    presentation: {
      kind: 'target_picker',
    },
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

  it('shows full weakness labels when only message hints weakness branch choice', async () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    useSessionStore().setRoomInfo('ROOM1', 'p2', 'Blue', 'fighter')
    useSnapshotStore().updateGameState(buildState({
      players: {
        p2: buildPlayer({
          id: 'p2',
          name: '测试玩家1',
          camp: 'Blue',
          role: 'fighter',
          hand: [buildCard({ id: 'c1', name: '火球', type: 'Magic', element: 'Fire' })],
          field: [],
          is_active: true,
          buffs: [],
          tokens: {},
        }),
      },
    }))
    useInterruptStore().setPrompt({
      type: 'confirm',
      player_id: 'p2',
      message: '【虚弱状态】测试玩家1，你需要做出选择：',
      options: [
        { id: '0', label: '摸3张牌继续执行后续行动' },
        { id: '1', label: '跳过此回合' },
      ],
      min: 1,
      max: 1,
    })

    render(PromptDialog, {
      global: {
        plugins: [pinia],
      },
    })

    expect(screen.getByText('摸3张牌继续执行后续行动')).toBeInTheDocument()
    expect(screen.getByText('跳过此回合')).toBeInTheDocument()
    expect(screen.queryByText(/^1$/)).not.toBeInTheDocument()
    expect(screen.queryByText(/^2$/)).not.toBeInTheDocument()
  })

  it('does not render medusa dark moon pick as decision overlay', async () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    useSessionStore().setRoomInfo('ROOM1', 'p2', 'Blue', 'moon_goddess')
    useSnapshotStore().updateGameState(buildState())
    useInterruptStore().setPrompt(medusaDarkMoonPickPrompt())

    render(PromptDialog, {
      global: {
        plugins: [pinia],
      },
    })

    expect(screen.queryByTestId('decision-overlay')).not.toBeInTheDocument()
    expect(screen.queryByText('移除闇月[暗月法术/Magic/Dark]')).not.toBeInTheDocument()
  })

  it('renders moon cycle branch prompt with decline and both branches', async () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    useSessionStore().setRoomInfo('ROOM1', 'p2', 'Blue', 'moon_goddess')
    useSnapshotStore().updateGameState(buildState({
      players: {
        p2: buildPlayer({
          id: 'p2',
          name: '测试玩家2',
          camp: 'Blue',
          role: 'moon_goddess',
          hand: [],
          field: [],
          is_active: true,
          buffs: [],
          tokens: {},
        }),
        p3: buildPlayer({
          id: 'p3',
          name: '干扰玩家',
          camp: 'Red',
          role: 'fighter',
          hand: [],
          field: [],
          is_active: false,
          buffs: [],
          tokens: {},
        }),
      },
    }))
    useInterruptStore().setPrompt(moonCycleBranchPrompt())

    render(PromptDialog, {
      global: {
        plugins: [pinia],
      },
    })

    expect(screen.getByTestId('decision-overlay')).toBeInTheDocument()
    expect(screen.getByText('不发动')).toBeInTheDocument()
    expect(screen.getByText('分支①：移除1个闇月，令目标角色+1治疗')).toBeInTheDocument()
    expect(screen.getByText('分支②：移除1点治疗，你+1新月')).toBeInTheDocument()
  })

  it('renders moon cycle healing target as a target-selection hint instead of option buttons', async () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    useSessionStore().setRoomInfo('ROOM1', 'p2', 'Blue', 'moon_goddess')
    useSnapshotStore().updateGameState(buildState({
      players: {
        p2: buildPlayer({
          id: 'p2',
          name: '测试玩家2',
          camp: 'Blue',
          role: 'moon_goddess',
          is_active: true,
        }),
        p3: buildPlayer({
          id: 'p3',
          name: '目标玩家',
          camp: 'Red',
          role: 'fighter',
        }),
      },
    }))
    useInterruptStore().setPrompt(moonCycleTargetPrompt())

    render(PromptDialog, {
      global: {
        plugins: [pinia],
      },
    })

    expect(screen.getByText('【月之轮回】请选择获得1点治疗的角色：')).toBeInTheDocument()
    expect(screen.queryByTestId('decision-overlay')).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: '目标玩家' })).not.toBeInTheDocument()
  })

  it('renders weakness choice without presentation as full branch labels', async () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    useSessionStore().setRoomInfo('ROOM1', 'p2', 'Blue', 'fighter')
    useSnapshotStore().updateGameState(buildState())
    const { presentation: _presentation, ...legacyWeak } = weakPrompt()
    useInterruptStore().setPrompt(legacyWeak)

    render(PromptDialog, {
      global: {
        plugins: [pinia],
      },
    })

    expect(screen.getByText('摸3张牌继续执行后续行动')).toBeInTheDocument()
    expect(screen.getByText('跳过此回合')).toBeInTheDocument()
    expect(screen.queryByText('取消')).not.toBeInTheDocument()
    expect(screen.queryByTestId('numeric-option-2')).not.toBeInTheDocument()
  })

  it('renders weakness choice as a decision overlay instead of a hand-card picker', async () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    useSessionStore().setRoomInfo('ROOM1', 'p2', 'Blue', 'fighter')
    useSnapshotStore().updateGameState(buildState())
    useInterruptStore().setPrompt(weakPrompt())

    render(PromptDialog, {
      global: {
        plugins: [pinia],
      },
    })

    expect(screen.getByTestId('decision-overlay')).toBeInTheDocument()
    expect(screen.getByText('摸3张牌继续执行后续行动')).toBeInTheDocument()
    expect(screen.getByText('跳过此回合')).toBeInTheDocument()
    expect(screen.queryByText('完成选牌后点击发动')).not.toBeInTheDocument()

    await userEvent.click(screen.getByText('摸3张牌继续执行后续行动'))

    expect(submitSelectMock).toHaveBeenCalledWith([0])
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
