import { render, screen, waitFor } from '@testing-library/vue'
import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent } from 'vue'
import BattleZone from '../BattleZone.vue'
import { useBattleFxStore } from '../../stores/battlefx.store'
import { useSnapshotStore } from '../../stores/snapshot.store'
import type { Card, PlayerView } from '../../types/game'

vi.mock('../StatusIcons/RoseCourtyardIcon.vue', () => ({
  default: defineComponent({
    name: 'RoseCourtyardIcon',
    template: '<div data-testid="rose-icon" />',
  }),
}))

function buildPlayer(overrides: Partial<PlayerView> = {}): PlayerView {
  return {
    id: 'p1',
    name: '玩家',
    camp: 'Red',
    role: 'angel',
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

function buildCard(overrides: Partial<Card> = {}): Card {
  return {
    id: 'fire-slash',
    name: '火焰斩',
    type: 'Attack',
    element: 'Fire',
    damage: 2,
    description: '',
    ...overrides,
  }
}

describe('BattleZone action narrative', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('renders real-time narrative actors, cards and events', async () => {
    const snapshotStore = useSnapshotStore()
    const battleFxStore = useBattleFxStore()
    snapshotStore.setCharacters([
      { id: 'angel', name: '天使', title: '', faction: '', skills: [] },
      { id: 'berserker', name: '狂战士', title: '', faction: '', skills: [] },
    ])
    snapshotStore.updateGameState({
      turn_stage: 'ActionExecution',
      current_player: 'p1',
      has_performed_startup: true,
      players: {
        p1: buildPlayer({ id: 'p1', name: '天使玩家', role: 'angel', camp: 'Blue' }),
        p2: buildPlayer({ id: 'p2', name: '狂战士玩家', role: 'berserker', camp: 'Red' }),
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
    })

    battleFxStore.beginActionNarrative('p1')
    battleFxStore.addCombatCue('p1', 'p2', 'attack')
    battleFxStore.addNarrativeCard('p1', buildCard(), 'attack')

    render(BattleZone)

    expect(screen.getAllByText('天使').length).toBeGreaterThan(0)
    expect(screen.getAllByText('狂战士').length).toBeGreaterThan(0)
    await waitFor(() => {
      expect(screen.getAllByText('攻击').length).toBeGreaterThan(0)
    })
    expect(screen.getByText('火焰斩')).toBeTruthy()
  })

  it('places narrative actor cards by battle seat direction', () => {
    const snapshotStore = useSnapshotStore()
    const battleFxStore = useBattleFxStore()
    snapshotStore.setCharacters([
      { id: 'angel', name: '天使', title: '', faction: '', skills: [] },
      { id: 'berserker', name: '狂战士', title: '', faction: '', skills: [] },
    ])
    snapshotStore.updateGameState({
      turn_stage: 'ActionExecution',
      current_player: 'p1',
      has_performed_startup: true,
      players: {
        p1: buildPlayer({ id: 'p1', name: '天使玩家', role: 'angel', camp: 'Blue' }),
        p2: buildPlayer({ id: 'p2', name: '狂战士玩家', role: 'berserker', camp: 'Red' }),
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
    })

    battleFxStore.beginActionNarrative('p1')
    battleFxStore.addCombatCue('p1', 'p2', 'attack')

    const { container } = render(BattleZone, {
      props: {
        actorSeatPositions: {
          p1: 3,
          p2: 0,
        },
      },
    })

    const featured = container.querySelector('.narrative-actor-card--featured')
    const opposed = container.querySelector('.narrative-actor-card--opposed')

    expect(featured?.classList.contains('narrative-actor-card--side-right')).toBe(true)
    expect(featured?.classList.contains('narrative-actor-card--row-top')).toBe(true)
    expect(opposed?.classList.contains('narrative-actor-card--side-left')).toBe(true)
    expect(opposed?.classList.contains('narrative-actor-card--row-top')).toBe(true)
  })

  it('draws played cards beside their actor and links the card to the target', async () => {
    const snapshotStore = useSnapshotStore()
    const battleFxStore = useBattleFxStore()
    snapshotStore.setCharacters([
      { id: 'angel', name: '天使', title: '', faction: '', skills: [] },
      { id: 'wind_sword_saint', name: '风之剑圣', title: '', faction: '', skills: [] },
    ])
    snapshotStore.updateGameState({
      turn_stage: 'ActionExecution',
      current_player: 'p1',
      has_performed_startup: true,
      players: {
        p1: buildPlayer({ id: 'p1', name: '天使玩家', role: 'angel', camp: 'Blue' }),
        p2: buildPlayer({ id: 'p2', name: '风之剑圣玩家', role: 'wind_sword_saint', camp: 'Red' }),
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
    })

    battleFxStore.beginActionNarrative('p1')
    battleFxStore.addCombatCue('p1', 'p2', 'attack')
    battleFxStore.addNarrativeCard('p1', buildCard({ name: '地裂斩' }), 'attack')

    const { container } = render(BattleZone, {
      props: {
        actorSeatPositions: {
          p1: 0,
          p2: 3,
        },
      },
    })

    const playedCard = container.querySelector('.narrative-played-card')
    await waitFor(() => {
      expect(container.querySelector('.narrative-link')).not.toBeNull()
    })
    const link = container.querySelector('.narrative-link')

    expect(playedCard?.classList.contains('narrative-played-card--side-left')).toBe(true)
    expect(playedCard?.classList.contains('narrative-played-card--row-top')).toBe(true)
    expect(link?.classList.contains('narrative-link--from-card')).toBe(true)
    expect(link?.classList.contains('narrative-link--from-left')).toBe(true)
    expect(link?.classList.contains('narrative-link--to-right')).toBe(true)
  })

  it('keeps settled actor cards in the same seat layer as cards and links', async () => {
    const snapshotStore = useSnapshotStore()
    const battleFxStore = useBattleFxStore()
    snapshotStore.setCharacters([
      { id: 'sealer', name: '封印师', title: '', faction: '', skills: [] },
      { id: 'angel', name: '天使', title: '', faction: '', skills: [] },
      { id: 'wind_sword_saint', name: '风之剑圣', title: '', faction: '', skills: [] },
    ])
    snapshotStore.updateGameState({
      turn_stage: 'ActionExecution',
      current_player: 'p1',
      has_performed_startup: true,
      players: {
        p1: buildPlayer({ id: 'p1', name: '封印师玩家', role: 'sealer', camp: 'Blue' }),
        p2: buildPlayer({ id: 'p2', name: '风之剑圣玩家', role: 'wind_sword_saint', camp: 'Red' }),
        p3: buildPlayer({ id: 'p3', name: '天使玩家', role: 'angel', camp: 'Blue' }),
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
    })

    battleFxStore.beginActionNarrative('p1')
    battleFxStore.addCombatCue('p1', 'p2', 'attack')
    battleFxStore.addNarrativeCard('p3', buildCard({ name: '地裂斩' }), 'counter', 'p2')
    battleFxStore.settleNarrativeActor('p3')

    const { container } = render(BattleZone, {
      props: {
        actorSeatPositions: {
          p1: 1,
          p2: 4,
          p3: 2,
        },
      },
    })

    await waitFor(() => {
      expect(container.querySelector('[data-narrative-actor-id="p3"]')).not.toBeNull()
    })

    const settledCard = container.querySelector('[data-narrative-actor-id="p3"]')
    const playedCard = container.querySelector('[data-narrative-card-id="1"]')

    expect(settledCard?.classList.contains('narrative-settled-card')).toBe(true)
    expect(settledCard?.classList.contains('narrative-settled-card--side-left')).toBe(true)
    expect(settledCard?.classList.contains('narrative-settled-card--row-bottom')).toBe(true)
    expect(playedCard?.classList.contains('narrative-played-card--side-left')).toBe(true)
    expect(playedCard?.classList.contains('narrative-played-card--row-bottom')).toBe(true)
  })

  it('shows Rose Courtyard inside the narrative layer without an idle label', () => {
    const snapshotStore = useSnapshotStore()
    snapshotStore.updateGameState({
      turn_stage: 'ActionExecution',
      current_player: 'p1',
      has_performed_startup: true,
      players: {
        p1: buildPlayer({
          field: [{
            card: buildCard({ name: '血蔷薇庭院' }),
            owner_id: 'p1',
            source_id: 'p1',
            mode: 'Effect',
            effect: 'RoseCourtyard',
            field_hook: '',
            locked: false,
            duration: 0,
          }],
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
    })

    render(BattleZone)

    expect(screen.getByText('血蔷薇庭院')).toBeTruthy()
    expect(screen.getByText('玩家无法使用治疗抵消伤害')).toBeTruthy()
    expect(screen.queryByText('战区')).toBeNull()
  })
})
