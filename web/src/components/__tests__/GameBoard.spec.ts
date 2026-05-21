import { render, screen } from '@testing-library/vue'
import userEvent from '@testing-library/user-event'
import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent } from 'vue'
import GameBoard from '../GameBoard.vue'
import { useInterruptStore } from '../../stores/interrupt.store'
import { useSessionStore } from '../../stores/session.store'
import { useSnapshotStore } from '../../stores/snapshot.store'
import type { Card, CharacterView, GameStateUpdate, PlayerInfo, PlayerView, Prompt } from '../../types/game'

const submitSelectMock = vi.fn()
const submitSelectCardIDsMock = vi.fn()

vi.mock('../../composables/useSubmitAction', () => ({
  useSubmitAction: () => ({
    submitSelect: submitSelectMock,
    submitSelectCardIDs: submitSelectCardIDsMock,
    submitCancel: vi.fn(),
    submitConfirm: vi.fn(),
    submitRespondTake: vi.fn(),
    submitRespondCounter: vi.fn(),
    submitRespondDefend: vi.fn(),
    submitAction: vi.fn(),
    submitUseSkill: vi.fn(),
    submitSelectedBoardTarget: vi.fn(),
    takeoverPlayer: vi.fn(),
    dissolveRoom: vi.fn(),
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

function buildCard(overrides: Partial<Card> = {}): Card {
  return {
    id: 'card-1',
    name: '火焰斩',
    type: 'Attack',
    element: 'Fire',
    damage: 2,
    description: 'test',
    ...overrides,
  }
}

function buildState(
  players: Record<string, PlayerView>,
  overrides: Partial<GameStateUpdate> = {},
): GameStateUpdate {
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
    ...overrides,
  }
}

function buildCharacter(overrides: Partial<CharacterView> = {}): CharacterView {
  return {
    id: 'fighter',
    name: '战士',
    title: '',
    faction: '',
    skills: [],
    ...overrides,
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

function actionHubPrompt(): Prompt {
  return {
    type: 'confirm',
    player_id: 'p1',
    choice_type: 'action_hub',
    message: '请选择行动',
    options: [
      { id: 'attack', label: '攻击', button_label: '攻击' },
      { id: 'magic', label: '法术', button_label: '法术' },
      { id: 'cannot_act', label: '无法行动', button_label: '无法行动' },
    ],
    min: 1,
    max: 1,
    presentation: {
      kind: 'action_hub',
      numeric_base: 0,
    },
  }
}

function handCardPickerPrompt(cardSource: 'hand' | 'proxy' = 'hand'): Prompt {
  return {
    type: 'choose_cards',
    player_id: 'p1',
    choice_type: 'test_discard_card',
    message: '请选择弃置1张手牌',
    options: [
      { id: '0', label: '1: 火焰斩', button_label: '选择', card_id: 'card-1' },
    ],
    min: 1,
    max: 1,
    presentation: {
      kind: 'card_picker',
      card_source: cardSource,
      numeric_base: 0,
    },
  }
}

describe('GameBoard target picker', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    submitSelectMock.mockReset()
    submitSelectCardIDsMock.mockReset()
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

  it('keeps enemy players selectable after choosing an attack card from the action hub', () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    const me = buildPlayer({
      id: 'p1',
      name: '攻击者',
      camp: 'Red',
      role: 'fighter',
      hand: [{
        id: 'atk-1',
        name: '火焰斩',
        type: 'Attack',
        element: 'Fire',
        damage: 2,
        description: 'test',
      }],
      hand_count: 1,
    })
    const target = buildPlayer({ id: 'p2', name: '可攻击目标', camp: 'Blue', role: 'fighter' })
    const stealthed = buildPlayer({
      id: 'p3',
      name: '潜行目标',
      camp: 'Blue',
      role: 'fighter',
      field: [{
        card: buildCard({ id: 'stealth-effect' }),
        mode: 'Effect',
        effect: 'Stealth',
        source_id: 'p3',
        owner_id: 'p3',
        field_hook: 'Manual',
        locked: false,
        duration: 0,
      }],
    })
    const players = { p1: me, p2: target, p3: stealthed }

    useSessionStore().setRoomInfo('ROOM1', 'p1', 'Red', 'fighter')
    useSessionStore().updateRoomPlayers(Object.values(players).map(buildPlayerInfo), 'p1')
    useSnapshotStore().updateGameState(buildState(players))
    const interruptStore = useInterruptStore()
    interruptStore.setPrompt(actionHubPrompt())
    interruptStore.setActionMode('attack')
    interruptStore.setSelectedHandIndexForAction(0)

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

    expect(screen.getByTestId('player-area-p2')).not.toBeDisabled()
    expect(screen.getByTestId('player-area-p3')).toBeDisabled()
  })

  it('selects a single hand card picker card without auto-submitting', async () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    const me = buildPlayer({
      id: 'p1',
      name: '弃牌者',
      camp: 'Red',
      role: 'fighter',
      hand: [buildCard({ id: 'card-1', name: '火焰斩' })],
      hand_count: 1,
    })
    const players = { p1: me }

    useSessionStore().setRoomInfo('ROOM1', 'p1', 'Red', 'fighter')
    useSessionStore().updateRoomPlayers(Object.values(players).map(buildPlayerInfo), 'p1')
    useSnapshotStore().updateGameState(buildState(players))
    const interruptStore = useInterruptStore()
    interruptStore.setPrompt(handCardPickerPrompt())

    render(GameBoard, {
      global: {
        plugins: [pinia],
        stubs: {
          PlayerArea: PlayerAreaStub,
          ActionPanel: true,
          BattleZone: true,
          SkillDetailModal: true,
          VfxLayer: true,
          ActionTimeline: true,
          StatusEffectIcon: true,
        },
      },
    })

    await userEvent.click(screen.getByTestId('hand-card-0'))

    expect(interruptStore.selectedHandIndexes).toEqual([0])
    expect(submitSelectCardIDsMock).not.toHaveBeenCalled()
    expect(submitSelectMock).not.toHaveBeenCalled()
  })

  it('allows proxy card picker cards that map to my hand', async () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    const me = buildPlayer({
      id: 'p1',
      name: '弃牌者',
      camp: 'Red',
      role: 'magic_lancer',
      hand: [buildCard({ id: 'card-1', name: '幻影星尘' })],
      hand_count: 1,
    })
    const players = { p1: me }

    useSessionStore().setRoomInfo('ROOM1', 'p1', 'Red', 'magic_lancer')
    useSessionStore().updateRoomPlayers(Object.values(players).map(buildPlayerInfo), 'p1')
    useSnapshotStore().updateGameState(buildState(players))
    const interruptStore = useInterruptStore()
    interruptStore.setPrompt(handCardPickerPrompt('proxy'))

    render(GameBoard, {
      global: {
        plugins: [pinia],
        stubs: {
          PlayerArea: PlayerAreaStub,
          ActionPanel: true,
          BattleZone: true,
          SkillDetailModal: true,
          VfxLayer: true,
          ActionTimeline: true,
          StatusEffectIcon: true,
        },
      },
    })

    await userEvent.click(screen.getByTestId('hand-card-0'))

    expect(interruptStore.selectedHandIndexes).toEqual([0])
    expect(submitSelectCardIDsMock).not.toHaveBeenCalled()
    expect(submitSelectMock).not.toHaveBeenCalled()
  })

  it('renders the current player in the same 3+3 side layout as other players', () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    const roster = Array.from({ length: 6 }, (_, index) => {
      const n = index + 1
      return buildPlayer({
        id: `p${n}`,
        name: `玩家${n}`,
        camp: n <= 3 ? 'Red' : 'Blue',
        role: 'fighter',
        heal: n,
        max_heal: 5,
      })
    })
    const players = Object.fromEntries(roster.map((player) => [player.id, player]))

    useSessionStore().setRoomInfo('ROOM1', 'p1', 'Red', 'fighter')
    useSessionStore().updateRoomPlayers(roster.map(buildPlayerInfo), 'p1')
    useSnapshotStore().updateGameState(buildState(players, {
      characters: [buildCharacter({ id: 'fighter', name: '剑斗士' })],
    }))

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

    expect(screen.getAllByTestId(/^player-area-p/)).toHaveLength(6)
    expect(screen.getByTestId('player-area-p1')).toBeInTheDocument()
    expect(screen.getByText('治疗 1/5')).toBeInTheDocument()
    expect(document.querySelector('.my-status-name')).toHaveTextContent('剑斗士')
  })

})
