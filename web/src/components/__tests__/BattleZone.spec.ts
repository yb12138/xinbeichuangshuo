import { render, screen, waitFor } from '@testing-library/vue'
import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, nextTick } from 'vue'
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
      expect(screen.getAllByText('发起攻击').length).toBeGreaterThan(0)
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

  it('lays out played cards in the center and connects actor, card and target with mist', async () => {
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

    const stackItem = container.querySelector('[data-narrative-stack-id="card-1"]')
    const playedCard = container.querySelector('.narrative-played-card')
    await waitFor(() => {
      expect(container.querySelectorAll('.narrative-mist').length).toBeGreaterThanOrEqual(2)
    })
    const mistSegments = Array.from(container.querySelectorAll('.narrative-mist'))

    expect(stackItem?.classList.contains('narrative-stack-item--card')).toBe(true)
    expect(stackItem?.classList.contains('narrative-stack-item--latest')).toBe(true)
    expect(stackItem?.getAttribute('style')).toContain('--stack-order')
    expect(playedCard?.classList.contains('narrative-played-card--attack')).toBe(true)
    expect(mistSegments.some(segment => segment.classList.contains('narrative-mist--from-actor'))).toBe(true)
    expect(mistSegments.some(segment => segment.classList.contains('narrative-mist--to-actor'))).toBe(true)
    expect(container.querySelector('.narrative-link')).toBeNull()
  })

  it('keeps settled actor cards in their seat column while cards stay in the central layout', async () => {
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
    const stackItem = container.querySelector('[data-narrative-card-id="1"]')

    expect(settledCard?.classList.contains('narrative-settled-card')).toBe(true)
    expect(settledCard?.classList.contains('narrative-settled-card--side-left')).toBe(true)
    expect(settledCard?.classList.contains('narrative-settled-card--row-bottom')).toBe(true)
    expect(stackItem?.classList.contains('narrative-stack-item')).toBe(true)
    expect(stackItem?.classList.contains('narrative-stack-item--from-left')).toBe(true)
    expect(stackItem?.getAttribute('style')).toContain('--stack-order')
  })

  it('plays card steps one by one before entering review', async () => {
    vi.useFakeTimers()
    const snapshotStore = useSnapshotStore()
    const battleFxStore = useBattleFxStore()
    snapshotStore.setCharacters([
      { id: 'assassin', name: '暗杀者', title: '', faction: '', skills: [] },
      { id: 'wind_sword_saint', name: '风之剑圣', title: '', faction: '', skills: [] },
    ])
    snapshotStore.updateGameState({
      turn_stage: 'ActionExecution',
      current_player: 'p1',
      has_performed_startup: true,
      players: {
        p1: buildPlayer({ id: 'p1', name: '暗杀者玩家', role: 'assassin', camp: 'Blue' }),
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
    battleFxStore.addNarrativeCard('p2', buildCard({ id: 'holy-shield', name: '圣光防御' }), 'defend', 'p1')

    const { container } = render(BattleZone, {
      props: {
        actorSeatPositions: {
          p1: 1,
          p2: 4,
        },
      },
    })

    const first = container.querySelector<HTMLElement>('[data-narrative-stack-id="card-1"]')
    const hiddenNext = container.querySelector<HTMLElement>('[data-narrative-stack-id="card-2"]')

    expect(first?.classList.contains('narrative-stack-item--latest')).toBe(false)
    expect(first?.classList.contains('narrative-stack-item--active-step')).toBe(true)
    expect(hiddenNext).toBeNull()

    vi.advanceTimersByTime(950)
    await nextTick()

    const latest = container.querySelector<HTMLElement>('[data-narrative-stack-id="card-2"]')
    expect(first?.classList.contains('narrative-stack-item--step-completed')).toBe(true)
    expect(latest?.classList.contains('narrative-stack-item--latest')).toBe(true)
    expect(latest?.classList.contains('narrative-stack-item--respond')).toBe(true)
    expect(latest?.classList.contains('narrative-stack-item--active-step')).toBe(true)
    expect(Number(latest?.style.zIndex || 0)).toBeGreaterThan(Number(first?.style.zIndex || 0))
    expect(first?.style.getPropertyValue('--stack-order')).toBe('0')
    expect(latest?.style.getPropertyValue('--stack-order')).toBe('1')
    vi.useRealTimers()
  })

  it('keeps card and skill chain items grouped in the final review', async () => {
    vi.useFakeTimers()
    const snapshotStore = useSnapshotStore()
    const battleFxStore = useBattleFxStore()
    snapshotStore.setCharacters([
      { id: 'sealer', name: '封印师', title: '', faction: '', skills: [] },
      { id: 'blade_master', name: '风之剑圣', title: '', faction: '', skills: [] },
    ])
    snapshotStore.updateGameState({
      turn_stage: 'ActionExecution',
      current_player: 'p1',
      has_performed_startup: true,
      players: {
        p1: buildPlayer({ id: 'p1', name: '封印师玩家', role: 'sealer', camp: 'Blue' }),
        p2: buildPlayer({ id: 'p2', name: '风之剑圣玩家', role: 'blade_master', camp: 'Red' }),
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

    battleFxStore.addNarrativeSkill('p1', '水之封印', '目标本回合无法使用水系牌', ['p2'])
    battleFxStore.addNarrativeSkill('p1', '法术激荡', '额外获得1次攻击行动', [])
    battleFxStore.addNarrativeCard('p1', buildCard({ id: 'earth-slash', name: '地裂斩', element: 'Earth' }), 'attack', 'p2')

    const { container } = render(BattleZone)

    expect(screen.getByText('水之封印')).toBeTruthy()
    expect(screen.queryByText('法术激荡')).toBeNull()

    vi.advanceTimersByTime(1120)
    await nextTick()

    expect(screen.getByText('法术激荡')).toBeTruthy()
    expect(screen.queryByText('地裂斩')).toBeNull()

    vi.advanceTimersByTime(1120)
    await nextTick()

    expect(screen.getByText('地裂斩')).toBeTruthy()

    vi.advanceTimersByTime(940)
    await nextTick()

    const items = Array.from(container.querySelectorAll<HTMLElement>('.narrative-stack-item'))
    const groups = Array.from(container.querySelectorAll<HTMLElement>('.narrative-step-group'))
    expect(items).toHaveLength(3)
    expect(groups).toHaveLength(3)
    expect(container.querySelector('.narrative-stack-lane--review')).not.toBeNull()
    expect(items.map(item => item.style.getPropertyValue('--stack-order'))).toEqual(['0', '1', '2'])
    expect(items.every(item => item.style.getPropertyValue('--stack-offset-x') === '')).toBe(true)
    vi.useRealTimers()
  })

  it('renders skill activations as central tokens with skill mist', async () => {
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

    battleFxStore.addNarrativeSkill('p1', '圣光裁决', '对目标造成2点法术伤害', ['p2'])

    const { container } = render(BattleZone, {
      props: {
        actorSeatPositions: {
          p1: 0,
          p2: 3,
        },
      },
    })

    expect(screen.getByText('圣光裁决')).toBeTruthy()
    expect(screen.queryByText('对目标造成2点法术伤害')).toBeNull()
    expect(container.querySelector('[data-narrative-skill-id="skill-1"]')).not.toBeNull()
    expect(container.querySelector('.narrative-skill-token__ring')).toBeNull()
    await waitFor(() => {
      expect(container.querySelectorAll('.narrative-mist--skill').length).toBeGreaterThanOrEqual(2)
    })
  })

  it('shows damage near mist target endpoints', async () => {
    const snapshotStore = useSnapshotStore()
    const battleFxStore = useBattleFxStore()
    snapshotStore.setCharacters([
      { id: 'assassin', name: '暗杀者', title: '', faction: '', skills: [] },
      { id: 'angel', name: '天使', title: '', faction: '', skills: [] },
    ])
    snapshotStore.updateGameState({
      turn_stage: 'ActionExecution',
      current_player: 'p1',
      has_performed_startup: true,
      players: {
        p1: buildPlayer({ id: 'p1', name: '暗杀者玩家', role: 'assassin', camp: 'Blue' }),
        p2: buildPlayer({ id: 'p2', name: '天使玩家', role: 'angel', camp: 'Red' }),
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
    battleFxStore.addNarrativeDamage('p1', 'p2', 2, 'Attack')

    const { container } = render(BattleZone, {
      props: {
        actorSeatPositions: {
          p1: 1,
          p2: 4,
        },
      },
    })

    await waitFor(() => {
      expect(container.querySelector('.narrative-mist__damage-label')?.textContent).toContain('-2')
    })
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
