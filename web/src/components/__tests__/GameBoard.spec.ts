import { render, screen, waitFor } from '@testing-library/vue'
import userEvent from '@testing-library/user-event'
import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent } from 'vue'
import GameBoard from '../GameBoard.vue'
import { useInterruptStore } from '../../stores/interrupt.store'
import { useSessionStore } from '../../stores/session.store'
import { useSnapshotStore } from '../../stores/snapshot.store'
import type { AvailableSkill, Card, CharacterView, GameStateUpdate, PlayerInfo, PlayerView, Prompt } from '../../types/game'

const submitSelectMock = vi.fn()
const submitSelectCardIDsMock = vi.fn()
const submitUseSkillMock = vi.fn()

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
    submitUseSkill: submitUseSkillMock,
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
    fighterHundredDragonText: { type: String, default: '' },
    fighterHundredDragonTitle: { type: String, default: '' },
  },
  emits: ['select'],
  template: `
    <div>
      <button
        type="button"
        :data-testid="'player-area-' + player.id"
        :disabled="!selectable"
        @click="$emit('select', player.id)"
      >
        {{ player.name }}
      </button>
      <span
        v-if="fighterHundredDragonText"
        :data-testid="'fighter-lock-' + player.id"
        :title="fighterHundredDragonTitle"
      >
        {{ fighterHundredDragonText }}
      </span>
    </div>
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

function discardGuidePrompt(discardReason?: string): Prompt {
  return {
    type: 'choose_cards',
    player_id: 'p1',
    choice_type: 'system_discard_cards',
    message: '请选择弃牌',
    options: [
      { id: '0', label: '1: 火焰斩', button_label: '选择', card_id: 'card-1' },
    ],
    min: 1,
    max: 1,
    presentation: {
      kind: 'card_picker',
      card_source: 'hand',
      card_filter: 'overflow_discard',
      discard_reason: discardReason,
      numeric_base: 0,
    },
  }
}

describe('GameBoard target picker', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    submitSelectMock.mockReset()
    submitSelectCardIDsMock.mockReset()
    submitUseSkillMock.mockReset()
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

  it('shows seal prompt and icon on the target player anchor', () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    const me = buildPlayer({ id: 'p1', name: '封印师玩家', camp: 'Red', role: 'sealer' })
    const target = buildPlayer({
      id: 'p2',
      name: '狂战士玩家',
      camp: 'Blue',
      role: 'fighter',
      field: [{
        card: buildCard({ id: 'water-seal-card', name: '水涟斩', element: 'Water' }),
        owner_id: 'p2',
        source_id: 'p1',
        mode: 'Effect',
        effect: 'SealWater',
        field_hook: '',
        locked: false,
        duration: 0,
      }],
    })
    const players = { p1: me, p2: target }

    useSessionStore().setRoomInfo('ROOM1', 'p1', 'Red', 'sealer')
    useSessionStore().updateRoomPlayers(Object.values(players).map(buildPlayerInfo), 'p1')
    useSnapshotStore().setCharacters([
      buildCharacter({ id: 'sealer', name: '封印师' }),
      buildCharacter({ id: 'fighter', name: '战士' }),
    ])
    useSnapshotStore().updateGameState(buildState(players))

    const { container } = render(GameBoard, {
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

    const targetAnchor = container.querySelector('[data-player-anchor="p2"]')
    expect(targetAnchor?.querySelector('.player-anchor-seal-effects')).not.toBeNull()
    expect(targetAnchor?.querySelector('.player-anchor-seal-pop')?.textContent).toContain('受到水之封印')
    expect(targetAnchor?.querySelector('.player-anchor-seal-icon')).not.toBeNull()
  })

  it('keeps rose courtyard ambient vfx while the field effect is active', async () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    const me = buildPlayer({
      id: 'p1',
      name: '血色剑灵',
      camp: 'Red',
      role: 'crimson_sword_spirit',
      field: [{
        card: buildCard({ id: 'rose-courtyard', name: '血蔷薇庭院', type: 'Magic' }),
        mode: 'Effect',
        effect: 'RoseCourtyard',
        source_id: 'p1',
        owner_id: 'p1',
        field_hook: 'Manual',
        locked: false,
        duration: 0,
      }],
    })
    const target = buildPlayer({ id: 'p2', name: '对手', camp: 'Blue', role: 'fighter' })
    const players = { p1: me, p2: target }

    useSessionStore().setRoomInfo('ROOM1', 'p1', 'Red', 'crimson_sword_spirit')
    useSessionStore().updateRoomPlayers(Object.values(players).map(buildPlayerInfo), 'p1')
    useSnapshotStore().updateGameState(buildState(players))

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

    expect(screen.getByTestId('rose-courtyard-vfx')).toBeTruthy()

    useSnapshotStore().updateGameState(buildState({
      p1: buildPlayer({ ...me, field: [] }),
      p2: target,
    }))

    await waitFor(() => {
      expect(screen.queryByTestId('rose-courtyard-vfx')).toBeNull()
    })
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

  it('pulses selectable enemies while choosing destruction storm targets', () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    const me = buildPlayer({ id: 'p1', name: '魔法少女', camp: 'Red', role: 'magical_girl' })
    const enemyA = buildPlayer({ id: 'p2', name: '目标A', camp: 'Blue', role: 'fighter' })
    const enemyB = buildPlayer({ id: 'p3', name: '目标B', camp: 'Blue', role: 'fighter' })
    const ally = buildPlayer({ id: 'p4', name: '队友', camp: 'Red', role: 'fighter' })
    const players = { p1: me, p2: enemyA, p3: enemyB, p4: ally }

    const destructionStorm: AvailableSkill = {
      id: 'destruction_storm',
      title: '毁灭风暴',
      description: '［宝石］对任2名目标对手各造成2点法术伤害③。',
      min_targets: 2,
      max_targets: 2,
      target_type: 2,
      cost_gem: 1,
      cost_crystal: 0,
      cost_discards: 0,
    }

    useSessionStore().setRoomInfo('ROOM1', 'p1', 'Red', 'magical_girl')
    useSessionStore().updateRoomPlayers(Object.values(players).map(buildPlayerInfo), 'p1')
    useSnapshotStore().updateGameState(buildState(players))
    const interruptStore = useInterruptStore()
    interruptStore.setSelectedSkill(destructionStorm)
    interruptStore.setSkillMode('choosing_target')

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
    expect(screen.getByTestId('player-area-p3')).not.toBeDisabled()
    expect(screen.getByTestId('player-area-p4')).toBeDisabled()
    expect(screen.getByTestId('player-area-p2').closest('.target-guide-pulse')).not.toBeNull()
    expect(screen.getByTestId('player-area-p3').closest('.target-guide-pulse')).not.toBeNull()
    expect(screen.getByTestId('player-area-p4').closest('.target-guide-pulse')).toBeNull()
    expect(document.querySelector('.table-target-guide-hint')).toBeNull()
  })

  it('only allows matching hand cards for exclusive skill selection', async () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    const prayerPowerBlessingSkill: AvailableSkill = {
      id: 'prayer_power_blessing',
      title: '威力赐福',
      description: '将独有技手牌当法术牌打出并放置于1名队友面前。',
      min_targets: 1,
      max_targets: 1,
      target_type: 3,
      cost_gem: 0,
      cost_crystal: 0,
      cost_discards: 0,
      require_exclusive: true,
      place_card: true,
      place_effect: 'PowerBlessing',
    }

    const me = buildPlayer({
      id: 'p1',
      name: '祈祷师',
      camp: 'Red',
      role: 'prayer_master',
      hand: [
        buildCard({
          id: 'card-match',
          name: '威力赐福',
          type: 'Magic',
          element: 'Light',
          description: '独有技手牌',
          exclusive_char1: 'prayer_master',
          exclusive_skill1: '威力赐福',
        }),
        buildCard({
          id: 'card-mismatch',
          name: '火焰斩',
          type: 'Attack',
          element: 'Fire',
          description: '普通手牌',
        }),
      ],
      hand_count: 2,
    })
    const ally = buildPlayer({ id: 'p2', name: '队友', camp: 'Red', role: 'angel' })
    const players = { p1: me, p2: ally }

    useSessionStore().setRoomInfo('ROOM1', 'p1', 'Red', 'prayer_master')
    useSessionStore().updateRoomPlayers(Object.values(players).map(buildPlayerInfo), 'p1')
    useSnapshotStore().updateGameState(buildState(players, {
      available_skills: [prayerPowerBlessingSkill],
      characters: [
        buildCharacter({
          id: 'prayer_master',
          name: '祈祷师',
          skills: [
            {
              ...prayerPowerBlessingSkill,
              type: 2,
            },
          ],
        }),
      ],
    }))

    const interruptStore = useInterruptStore()
    interruptStore.setSkillMode('choosing_exclusive')
    interruptStore.setSelectedSkill(prayerPowerBlessingSkill)

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

    await userEvent.click(screen.getByTestId('hand-card-1'))
    expect(interruptStore.skillDiscardHandIndexes).toEqual([])

    await userEvent.click(screen.getByTestId('hand-card-0'))
    expect(interruptStore.skillDiscardHandIndexes).toEqual([0])

    interruptStore.setSkillMode('choosing_target')
    await userEvent.click(screen.getByTestId('player-area-p2'))
    expect(submitUseSkillMock).toHaveBeenCalledWith(
      'prayer_power_blessing',
      ['p2'],
      [0],
      { clearSkillMode: true },
    )
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

  it('shows overflow discard guide for hand limit discard prompts', () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    const me = buildPlayer({
      id: 'p1',
      hand: [buildCard({ id: 'card-1' })],
      hand_count: 1,
    })
    const players = { p1: me }

    useSessionStore().setRoomInfo('ROOM1', 'p1', 'Red', 'fighter')
    useSessionStore().updateRoomPlayers(Object.values(players).map(buildPlayerInfo), 'p1')
    useSnapshotStore().updateGameState(buildState(players))
    useInterruptStore().setPrompt(discardGuidePrompt('hand_overflow'))

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

    expect(screen.getByText('爆牌弃牌阶段')).toBeInTheDocument()
  })

  it('does not show overflow discard guide for semantic skill discard prompts', () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    const me = buildPlayer({
      id: 'p1',
      hand: [buildCard({ id: 'card-1' })],
      hand_count: 1,
    })
    const players = { p1: me }

    useSessionStore().setRoomInfo('ROOM1', 'p1', 'Red', 'fighter')
    useSessionStore().updateRoomPlayers(Object.values(players).map(buildPlayerInfo), 'p1')
    useSnapshotStore().updateGameState(buildState(players))
    useInterruptStore().setPrompt(discardGuidePrompt('skill_effect'))

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

    expect(screen.queryByText('爆牌弃牌阶段')).not.toBeInTheDocument()
  })

  it('keeps overflow discard guide fallback for older prompt payloads', () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    const me = buildPlayer({
      id: 'p1',
      hand: [buildCard({ id: 'card-1' })],
      hand_count: 1,
    })
    const players = { p1: me }

    useSessionStore().setRoomInfo('ROOM1', 'p1', 'Red', 'fighter')
    useSessionStore().updateRoomPlayers(Object.values(players).map(buildPlayerInfo), 'p1')
    useSnapshotStore().updateGameState(buildState(players))
    useInterruptStore().setPrompt(discardGuidePrompt())

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

    expect(screen.getByText('爆牌弃牌阶段')).toBeInTheDocument()
  })

  it('renders all players around the battle table while keeping my HUD separate from the hand rail', () => {
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

    expect(screen.getByTestId('battle-table-scene')).toBeInTheDocument()
    expect(document.querySelector('.team-track')).toBeNull()
    expect(screen.getAllByTestId(/^battle-table-seat-/)).toHaveLength(6)
    expect(screen.getAllByTestId(/^player-area-p/)).toHaveLength(6)
    expect(screen.getByTestId('player-area-p1')).toBeInTheDocument()
    expect(screen.getByTestId('battle-table-seat-0')).toContainElement(screen.getByTestId('player-area-p1'))
    const statusPortrait = document.querySelector('.my-status-portrait') as HTMLImageElement | null
    const statusStrip = document.querySelector('.my-status-strip') as HTMLElement | null
    const statusOverlay = statusStrip?.querySelector('.my-status-overlay.player-overlay') as HTMLElement | null
    const handRail = document.querySelector('.hand-rail') as HTMLElement | null
    const actionDock = document.querySelector('.right-action-dock') as HTMLElement | null
    expect(statusStrip).not.toBeNull()
    expect(statusOverlay).not.toBeNull()
    expect(statusStrip?.querySelector('.my-status-content')).toBeNull()
    expect(statusOverlay?.textContent).toContain('剑斗士')
    expect(statusOverlay?.textContent).toContain('玩家：玩家1')
    expect(statusOverlay?.textContent).toContain('1/5')
    expect(statusOverlay?.textContent).toContain('0/6')
    expect(handRail).not.toBeNull()
    expect(actionDock).not.toBeNull()
    expect(statusStrip?.compareDocumentPosition(handRail as Node)).toBe(Node.DOCUMENT_POSITION_FOLLOWING)
    expect(handRail?.compareDocumentPosition(actionDock as Node)).toBe(Node.DOCUMENT_POSITION_FOLLOWING)
    expect(statusPortrait).not.toBeNull()
    expect(statusPortrait?.alt).toBe('剑斗士')
    expect(statusPortrait?.getAttribute('src')).toBe('/characters/fighter.png')
    expect(handRail?.querySelector('.my-status-portrait')).toBeNull()
    expect(statusStrip?.classList.contains('my-status-strip--active')).toBe(true)
    expect(actionDock?.classList.contains('right-action-dock--active')).toBe(true)
  })

  it('passes hundred dragon lock badges to source and target players without rendering a link line', () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    const me = buildPlayer({ id: 'p1', name: '格斗家', camp: 'Red', role: 'fighter' })
    const target = buildPlayer({
      id: 'p2',
      name: '圣女',
      camp: 'Blue',
      role: 'saintess',
      field: [{
        card: buildCard({ id: 'hundred-dragon-lock' }),
        mode: 'Effect',
        effect: 'FighterHundredDragonLock',
        source_id: 'p1',
        owner_id: 'p2',
        field_hook: 'Manual',
        locked: false,
        duration: 0,
      }],
    })
    const players = { p1: me, p2: target }

    useSessionStore().setRoomInfo('ROOM1', 'p1', 'Red', 'fighter')
    useSessionStore().updateRoomPlayers(Object.values(players).map(buildPlayerInfo), 'p1')
    useSnapshotStore().updateGameState(buildState(players))

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

    expect(screen.getByTestId('fighter-lock-p1')).toHaveTextContent('幻龙锁定')
    expect(screen.getByTestId('fighter-lock-p2')).toHaveTextContent('幻龙锁定')
    expect(document.querySelector('.link-lines-layer')).toBeNull()
  })

})
