import { render, screen } from '@testing-library/vue'
import userEvent from '@testing-library/user-event'
import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent } from 'vue'
import GameBoard from '../GameBoard.vue'
import { useInterruptStore } from '../../stores/interrupt.store'
import { useSessionStore } from '../../stores/session.store'
import { useSnapshotStore } from '../../stores/snapshot.store'
import type { GameStateUpdate, PlayerInfo, PlayerView, Prompt } from '../../types/game'

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
    submitUseSkill: vi.fn(),
    submitSelectedBoardTarget: vi.fn(),
  }),
}))

const PlayerAreaStub = defineComponent({
  name: 'PlayerArea',
  props: {
    player: { type: Object, required: true },
    selectable: { type: Boolean, default: false },
  },
  emits: ['select'],
  template: `
    <button
      type="button"
      :data-testid="'player-area-' + player.id"
      :disabled="!selectable"
      @click="$emit('select', player.id)"
    >
      {{ player.name }}
    </button>
  `,
})

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
    heal: 0,
    max_heal: 0,
    gem: 0,
    crystal: 0,
    is_active: false,
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

function buildState(players: Record<string, PlayerView>): GameStateUpdate {
  return {
    turn_stage: 'ActionExecution',
    current_player: 'p1',
    has_performed_startup: false,
    players,
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
  }
}

function targetPickerPrompt(): Prompt {
  return {
    type: 'confirm',
    player_id: 'p1',
    choice_type: 'same_name_target',
    message: '请选择目标',
    options: [
      { id: '0', target_id: 'p2', label: '目标一号', button_label: '选择' },
      { id: '1', target_id: 'p3', label: '目标二号', button_label: '选择' },
    ],
    min: 1,
    max: 1,
    presentation: {
      kind: 'target_picker',
      target_filter: 'custom',
      numeric_base: 0,
    },
  }
}

describe('GameBoard target picker', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    submitSelectMock.mockReset()
  })

  it('selects target picker options by target_id instead of player names or labels', async () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    const me = buildPlayer({ id: 'p1', name: '同名玩家', camp: 'Red', role: 'fighter' })
    const targetA = buildPlayer({ id: 'p2', name: '同名玩家', camp: 'Blue', role: 'fighter' })
    const targetB = buildPlayer({ id: 'p3', name: '同名玩家', camp: 'Blue', role: 'fighter' })
    const players = { p1: me, p2: targetA, p3: targetB }

    useSessionStore().setRoomInfo('ROOM1', 'p1', 'Red', 'fighter')
    useSessionStore().updateRoomPlayers(Object.values(players).map(buildPlayerInfo), 'p1')
    useSnapshotStore().updateGameState(buildState(players))
    useInterruptStore().setPrompt(targetPickerPrompt())

    render(GameBoard, {
      global: {
        plugins: [pinia],
        stubs: {
          PlayerArea: PlayerAreaStub,
          ActionPanel: true,
          BattleZone: true,
          CardComponent: true,
          SkillDetailModal: true,
          VfxLayer: true,
          ActionTimeline: true,
          StatusEffectIcon: true,
        },
      },
    })

    await userEvent.click(screen.getByTestId('player-area-p3'))

    expect(submitSelectMock).toHaveBeenCalledWith([1])
  })
})
