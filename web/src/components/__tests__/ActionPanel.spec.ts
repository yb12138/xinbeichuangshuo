import { render, screen, waitFor } from '@testing-library/vue'
import userEvent from '@testing-library/user-event'
import { createPinia, setActivePinia } from 'pinia'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import ActionPanel from '../ActionPanel.vue'
import { useInterruptStore } from '../../stores/interrupt.store'
import { useSessionStore } from '../../stores/session.store'
import { useSnapshotStore } from '../../stores/snapshot.store'
import type { AvailableSkill, GameStateUpdate, PlayerInfo, PlayerView, Prompt } from '../../types/game'

const submitCannotActMock = vi.fn()
const submitPassMock = vi.fn()
const submitSelectMock = vi.fn()

vi.mock('../../composables/useSubmitAction', () => ({
  useSubmitAction: () => ({
    submitCannotAct: submitCannotActMock,
    submitPass: submitPassMock,
    submitSelect: submitSelectMock,
    submitBuy: vi.fn(),
    submitSynthesize: vi.fn(),
    submitExtract: vi.fn(),
    submitUseSkill: vi.fn(),
    submitAction: vi.fn(),
    submitConfirm: vi.fn(),
    submitCancel: vi.fn(),
    submitSelectedBoardTarget: vi.fn(),
    submitRespondTake: vi.fn(),
    submitRespondCounter: vi.fn(),
    submitRespondDefend: vi.fn(),
    cheatSet: vi.fn(),
    cheatAddCard: vi.fn(),
    cheatDiscard: vi.fn(),
  }),
}))

vi.mock('../PromptDialog.vue', () => ({
  default: {
    name: 'PromptDialogStub',
    template: '<div data-testid="prompt-dialog-stub">Prompt</div>',
  },
}))

function buildPlayer(overrides: Partial<PlayerView> = {}): PlayerView {
  return {
    id: 'p1',
    name: '玩家',
    camp: 'Red',
    role: 'fighter',
    hand_count: 0,
    max_hand: 6,
    exclusive_card_count: 0,
    hand: [],
    exclusive_cards: [],
    field: [],
    heal: 3,
    max_heal: 5,
    gem: 0,
    crystal: 0,
    is_active: true,
    buffs: [],
    tokens: {},
    indicators: {},
    ...overrides,
  }
}

function buildPlayerInfo(player: PlayerView): PlayerInfo {
  return {
    id: player.id,
    name: player.name,
    camp: player.camp,
    char_role: player.role,
    ready: true,
    is_online: true,
  }
}

function buildState(players: Record<string, PlayerView>, availableSkills: AvailableSkill[] = []): GameStateUpdate {
  return {
    turn_stage: 'ActionExecution',
    current_player: 'p1',
    has_performed_startup: false,
    players,
    red_morale: 15,
    blue_morale: 15,
    red_cups: 0,
    blue_cups: 0,
    red_gems: 3,
    blue_gems: 0,
    red_crystals: 0,
    blue_crystals: 0,
    deck_count: 30,
    discard_count: 0,
    available_skills: availableSkills,
  }
}

function actionHubPrompt(options: Prompt['options']): Prompt {
  return {
    type: 'confirm',
    player_id: 'p1',
    choice_type: 'action_hub',
    message: '请选择行动',
    options,
    special_options: [],
    min: 1,
    max: 1,
    presentation: {
      kind: 'action_hub',
      numeric_base: 0,
    },
  }
}

function targetPickerPrompt(): Prompt {
  return {
    type: 'choose_target',
    player_id: 'p1',
    choice_type: 'target',
    message: '请选择目标',
    options: [{ id: '0', label: '目标', button_label: '选择', target_id: 'p2' }],
    min: 1,
    max: 1,
    presentation: {
      kind: 'target_picker',
      numeric_base: 0,
      target_filter: 'custom',
    },
  }
}

function setupPanel(options: { prompt?: Prompt; availableSkills?: AvailableSkill[] } = {}) {
  const pinia = createPinia()
  setActivePinia(pinia)
  const me = buildPlayer()
  useSessionStore().setRoomInfo('ROOM1', 'p1', 'Red', 'fighter')
  useSessionStore().updateRoomPlayers([buildPlayerInfo(me)], 'p1')
  useSnapshotStore().updateGameState(buildState({ p1: me }, options.availableSkills))
  const interruptStore = useInterruptStore()
  if (options.prompt) {
    interruptStore.setPrompt(options.prompt)
  }
  const utils = render(ActionPanel, { global: { plugins: [pinia] } })
  return { ...utils, interruptStore }
}

describe('ActionPanel collapsed action hub', () => {
  beforeEach(() => {
    submitCannotActMock.mockReset()
    submitPassMock.mockReset()
    submitSelectMock.mockReset()
    document.body.className = ''
  })

  afterEach(() => {
    document.body.className = ''
  })

  it('shows only the compact action trigger until opened', async () => {
    setupPanel()

    expect(screen.getByTestId('action-hub-trigger')).toBeInTheDocument()
    expect(screen.queryByTestId('action-hub-menu')).not.toBeInTheDocument()
    expect(screen.queryByTestId('action-magic')).not.toBeInTheDocument()

    await userEvent.click(screen.getByTestId('action-hub-trigger'))

    expect(screen.getByTestId('action-hub-menu')).toBeInTheDocument()
    expect(screen.getByTestId('action-attack')).toBeInTheDocument()
    expect(screen.getByTestId('action-magic')).toBeInTheDocument()
    expect(screen.getByTestId('action-pass')).toBeInTheDocument()
  })

  it('closes the menu on outside click and Escape', async () => {
    setupPanel()

    await userEvent.click(screen.getByTestId('action-hub-trigger'))
    expect(screen.getByTestId('action-hub-menu')).toBeInTheDocument()
    await userEvent.click(document.body)
    await waitFor(() => expect(screen.queryByTestId('action-hub-menu')).not.toBeInTheDocument())

    await userEvent.click(screen.getByTestId('action-hub-trigger'))
    expect(screen.getByTestId('action-hub-menu')).toBeInTheDocument()
    await userEvent.keyboard('{Escape}')
    await waitFor(() => expect(screen.queryByTestId('action-hub-menu')).not.toBeInTheDocument())
  })

  it('closes before entering attack and magic subflows', async () => {
    let { interruptStore, unmount } = setupPanel()

    await userEvent.click(screen.getByTestId('action-hub-trigger'))
    await userEvent.click(screen.getByTestId('action-magic'))
    await waitFor(() => expect(screen.queryByTestId('action-hub-menu')).not.toBeInTheDocument())
    expect(interruptStore.actionMode).toBe('magic')

    unmount()
    const secondPanel = setupPanel()
    interruptStore = secondPanel.interruptStore
    unmount = secondPanel.unmount
    await userEvent.click(screen.getByTestId('action-hub-trigger'))
    await userEvent.click(screen.getByTestId('action-attack'))
    await waitFor(() => expect(screen.queryByTestId('action-hub-menu')).not.toBeInTheDocument())
    expect(interruptStore.actionMode).toBe('attack')
    unmount()
  })

  it('keeps single-option action prompts compact and clickable', async () => {
    setupPanel({
      prompt: actionHubPrompt([
        { id: 'cannot_act', label: '无法行动', button_label: '无法行动' },
      ]),
    })

    await userEvent.click(screen.getByTestId('action-hub-trigger'))

    expect(screen.getByTestId('action-cannot-act')).toBeInTheDocument()
    expect(screen.queryByTestId('action-attack')).not.toBeInTheDocument()
    expect(screen.queryByTestId('action-magic')).not.toBeInTheDocument()

    await userEvent.click(screen.getByTestId('action-cannot-act'))

    expect(submitCannotActMock).toHaveBeenCalled()
  })

  it('does not collapse non-hub inline prompts', () => {
    setupPanel({ prompt: targetPickerPrompt() })

    expect(screen.queryByTestId('action-hub-trigger')).not.toBeInTheDocument()
    expect(screen.getByTestId('prompt-dialog-stub')).toBeInTheDocument()
  })
})
