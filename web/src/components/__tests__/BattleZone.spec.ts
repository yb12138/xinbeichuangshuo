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

interface TestRect {
  x: number
  y: number
  width: number
  height: number
}

const defaultLoopSize = { width: 900, height: 360 }

function cssNumber(element: HTMLElement, name: string) {
  return Number.parseFloat(element.style.getPropertyValue(name))
}

function percentToPx(value: number, axis: 'x' | 'y') {
  return value / 100 * (axis === 'x' ? defaultLoopSize.width : defaultLoopSize.height)
}

function rectFromCenter(x: number, y: number, width: number, height: number): TestRect {
  return {
    x: x - width / 2,
    y: y - height / 2,
    width,
    height,
  }
}

function inflateRect(rect: TestRect, amount: number): TestRect {
  return {
    x: rect.x - amount,
    y: rect.y - amount,
    width: rect.width + amount * 2,
    height: rect.height + amount * 2,
  }
}

function rectIntersects(a: TestRect, b: TestRect) {
  return a.x < b.x + b.width &&
    a.x + a.width > b.x &&
    a.y < b.y + b.height &&
    a.y + a.height > b.y
}

function loopPacketRect(element: HTMLElement) {
  return rectFromCenter(
    percentToPx(cssNumber(element, '--loop-packet-x'), 'x'),
    percentToPx(cssNumber(element, '--loop-packet-y'), 'y'),
    cssNumber(element, '--loop-packet-width'),
    cssNumber(element, '--loop-packet-height'),
  )
}

function loopNodeRect(element: HTMLElement) {
  return rectFromCenter(
    percentToPx(cssNumber(element, '--loop-node-x'), 'x'),
    percentToPx(cssNumber(element, '--loop-node-y'), 'y'),
    cssNumber(element, '--loop-node-width'),
    cssNumber(element, '--loop-node-height'),
  )
}

function loopActorRect(element: HTMLElement) {
  return rectFromCenter(
    percentToPx(cssNumber(element, '--loop-x'), 'x'),
    percentToPx(cssNumber(element, '--loop-y'), 'y'),
    90,
    122,
  )
}

function loopNodeCenter(element: HTMLElement) {
  return {
    x: cssNumber(element, '--loop-node-x'),
    y: cssNumber(element, '--loop-node-y'),
  }
}

function quadraticPoint(start: number, control: number, end: number, t: number) {
  const inverse = 1 - t
  return inverse * inverse * start + 2 * inverse * t * control + t * t * end
}

function distanceToAttackCurve(point: { x: number; y: number }) {
  let best = Number.POSITIVE_INFINITY
  for (let index = 0; index <= 100; index += 1) {
    const t = index / 100
    const x = quadraticPoint(50, 87.82, 50, t)
    const y = quadraticPoint(23.2, 52, 80.8, t)
    best = Math.min(best, Math.hypot(point.x - x, point.y - y))
  }
  return best
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

    const { container } = render(BattleZone)

    expect(screen.getAllByText('天使').length).toBeGreaterThan(0)
    expect(screen.getAllByText('狂战士').length).toBeGreaterThan(0)
    await waitFor(() => {
      expect(container.querySelector('[data-testid="narrative-combat-loop"]')).not.toBeNull()
    })
    expect(screen.getAllByText('攻击').length).toBeGreaterThan(0)
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

  it('lays out played combat cards on their action flow edge', async () => {
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

    const loop = container.querySelector('[data-testid="narrative-combat-loop"]')
    const stackItem = container.querySelector('[data-narrative-stack-id="card-1"]')

    expect(loop).not.toBeNull()
    expect(container.querySelectorAll('.narrative-loop-actor')).toHaveLength(2)
    expect(container.querySelectorAll('.narrative-loop-link')).toHaveLength(1)
    expect(stackItem?.classList.contains('narrative-loop-card')).toBe(true)
    expect(stackItem?.classList.contains('narrative-loop-action--red')).toBe(true)
    expect(screen.getByText('地裂斩')).toBeTruthy()
    expect(container.querySelector('.narrative-stack-lane')).toBeNull()
    expect(container.querySelector('.narrative-mist-layer')).toBeNull()
  })

  it('uses the combat loop for settled actors when their card is part of the flow', async () => {
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

    expect(container.querySelector('[data-testid="narrative-combat-loop"]')).not.toBeNull()
    expect(settledCard?.classList.contains('narrative-loop-actor')).toBe(true)
    expect(stackItem?.classList.contains('narrative-loop-card')).toBe(true)
    expect(stackItem?.classList.contains('narrative-loop-action--gold')).toBe(true)
    expect(container.querySelector('.narrative-settled-card')).toBeNull()
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

    expect(container.querySelector('[data-testid="narrative-combat-loop"]')).not.toBeNull()
    expect(first?.classList.contains('narrative-loop-card')).toBe(true)
    expect(first?.classList.contains('narrative-loop-action--step-active')).toBe(true)
    expect(hiddenNext?.classList.contains('narrative-loop-card')).toBe(true)
    expect(hiddenNext?.classList.contains('narrative-loop-action--step-pending')).toBe(true)

    vi.advanceTimersByTime(950)
    await nextTick()

    const latest = container.querySelector<HTMLElement>('[data-narrative-stack-id="card-2"]')
    expect(first?.classList.contains('narrative-loop-action--step-completed')).toBe(true)
    expect(latest?.classList.contains('narrative-loop-action--respond')).toBe(true)
    expect(latest?.classList.contains('narrative-loop-action--step-active')).toBe(true)
    expect(Number(latest?.style.zIndex || 0)).toBeGreaterThan(Number(first?.style.zIndex || 0))
    expect(first?.style.getPropertyValue('--loop-card-x')).not.toBe('')
    expect(latest?.style.getPropertyValue('--loop-card-x')).not.toBe('')
    vi.useRealTimers()
  })

  it('renders attack, counter and miss outcome in the unified combat loop', async () => {
    vi.useFakeTimers()
    const snapshotStore = useSnapshotStore()
    const battleFxStore = useBattleFxStore()
    snapshotStore.setCharacters([
      { id: 'wind_sword_saint', name: '风之剑圣', title: '', faction: '', skills: [] },
      { id: 'berserker', name: '狂战士', title: '', faction: '', skills: [] },
      { id: 'assassin', name: '暗杀者', title: '', faction: '', skills: [] },
    ])
    snapshotStore.updateGameState({
      turn_stage: 'ActionExecution',
      current_player: 'p1',
      has_performed_startup: true,
      players: {
        p1: buildPlayer({ id: 'p1', name: '风之剑圣玩家', role: 'wind_sword_saint', camp: 'Blue' }),
        p2: buildPlayer({ id: 'p2', name: '狂战士玩家', role: 'berserker', camp: 'Red' }),
        p3: buildPlayer({ id: 'p3', name: '暗杀者玩家', role: 'assassin', camp: 'Blue' }),
      },
      red_morale: 15,
      blue_morale: 15,
      red_cups: 0,
      blue_cups: 0,
      red_gems: 0,
      blue_gems: 0,
      blue_crystals: 0,
      red_crystals: 0,
      deck_count: 30,
      discard_count: 0,
      available_skills: [],
    })

    battleFxStore.applyStructuredTimelineNarrative({
      room_id: 'ROOM1',
      seq_start: 1,
      seq_end: 4,
      is_replay: false,
      events: [
        {
          event_id: 1,
          turn_id: 1,
          chain_id: 'nw-t1-p1',
          type: 'TimelineEffectResolved',
          outcome: 'TimelineOutcomeSuccess',
          visibility: 'TimelineVisibilityPublic',
          narrative_window_id: 'nw-t1-p1',
          action_id: 'nw-t1-p1-a1-attack',
          narrative_kind: 'action_started',
          visual_kind: 'action_marker',
          actor_user_id: 'p1',
          target_user_ids: ['p2'],
          action_type: 'attack',
        },
        {
          event_id: 2,
          turn_id: 1,
          chain_id: 'nw-t1-p1',
          type: 'TimelineActionDeclared',
          outcome: 'TimelineOutcomeSuccess',
          visibility: 'TimelineVisibilityPublic',
          narrative_window_id: 'nw-t1-p1',
          action_id: 'nw-t1-p1-a1-attack',
          narrative_kind: 'card_played',
          visual_kind: 'card',
          card_role: 'attack',
          actor_user_id: 'p1',
          target_user_ids: ['p2'],
          cards: [buildCard({ id: 'water-slash', name: '水涟斩', element: 'Water' })],
        },
        {
          event_id: 3,
          turn_id: 1,
          chain_id: 'nw-t1-p1',
          type: 'TimelineActionDeclared',
          outcome: 'TimelineOutcomeSuccess',
          visibility: 'TimelineVisibilityPublic',
          narrative_window_id: 'nw-t1-p1',
          action_id: 'nw-t1-p1-a1-attack',
          combat_id: 'nw-t1-p1-c1-counter',
          narrative_kind: 'card_played',
          visual_kind: 'card',
          card_role: 'counter',
          actor_user_id: 'p2',
          target_user_ids: ['p3'],
          cards: [buildCard({ id: 'counter-slash', name: '水涟斩', element: 'Water' })],
        },
        {
          event_id: 4,
          turn_id: 1,
          chain_id: 'nw-t1-p1',
          type: 'TimelineEffectResolved',
          outcome: 'TimelineOutcomeSuccess',
          visibility: 'TimelineVisibilityPublic',
          narrative_window_id: 'nw-t1-p1',
          action_id: 'nw-t1-p1-a1-attack',
          combat_id: 'nw-t1-p1-c1-counter',
          narrative_kind: 'field_effect_applied',
          visual_kind: 'effect_token',
          effect_type: 'attack_miss',
          actor_user_id: 'p1',
          target_user_ids: ['p2'],
          summary: '未命中',
        },
      ],
    })

    const { container } = render(BattleZone, {
      props: {
        actorSeatPositions: {
          p1: 4,
          p2: 1,
          p3: 0,
        },
      },
    })

    vi.advanceTimersByTime(940 + 940 + 1080 + 20)
    await nextTick()

    const loop = container.querySelector('[data-testid="narrative-combat-loop"]')

    expect(loop).not.toBeNull()
    expect(loop?.querySelector('[data-narrative-actor-id="p1"]')).not.toBeNull()
    expect(loop?.querySelector('[data-narrative-actor-id="p2"]')).not.toBeNull()
    expect(loop?.querySelector('[data-narrative-actor-id="p3"]')).not.toBeNull()
    expect(loop?.querySelector('[data-narrative-stack-id="card-1"]')).not.toBeNull()
    expect(loop?.querySelector('[data-narrative-stack-id="card-2"]')).not.toBeNull()
    expect(container.querySelectorAll('.narrative-loop-link')).toHaveLength(2)
    expect(screen.getByText('未命中')).toBeTruthy()
    expect(container.querySelector('.narrative-stack-lane')).toBeNull()
    expect(container.querySelector('.narrative-mist-layer')).toBeNull()
    vi.useRealTimers()
  })

  it('renders a multi-actor combat loop with cards on each flow edge', async () => {
    const snapshotStore = useSnapshotStore()
    const battleFxStore = useBattleFxStore()
    snapshotStore.setCharacters([
      { id: 'wind_sword_saint', name: '风之剑圣', title: '', faction: '', skills: [] },
      { id: 'sealer', name: '封印师', title: '', faction: '', skills: [] },
      { id: 'angel', name: '天使', title: '', faction: '', skills: [] },
      { id: 'berserker', name: '狂战士', title: '', faction: '', skills: [] },
    ])
    snapshotStore.updateGameState({
      turn_stage: 'ActionExecution',
      current_player: 'p1',
      has_performed_startup: true,
      players: {
        p1: buildPlayer({ id: 'p1', name: '风之剑圣玩家', role: 'wind_sword_saint', camp: 'Blue' }),
        p2: buildPlayer({ id: 'p2', name: '封印师玩家', role: 'sealer', camp: 'Red' }),
        p3: buildPlayer({ id: 'p3', name: '天使玩家', role: 'angel', camp: 'Blue' }),
        p4: buildPlayer({ id: 'p4', name: '狂战士玩家', role: 'berserker', camp: 'Red' }),
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

    battleFxStore.applyStructuredTimelineNarrative({
      room_id: 'ROOM1',
      seq_start: 1,
      seq_end: 5,
      is_replay: false,
      events: [
        {
          event_id: 1,
          turn_id: 1,
          chain_id: 'nw-loop',
          type: 'TimelineEffectResolved',
          outcome: 'TimelineOutcomeSuccess',
          visibility: 'TimelineVisibilityPublic',
          narrative_window_id: 'nw-loop',
          action_id: 'nw-loop-attack',
          narrative_kind: 'action_started',
          visual_kind: 'action_marker',
          actor_user_id: 'p1',
          target_user_ids: ['p2'],
          action_type: 'attack',
        },
        {
          event_id: 2,
          turn_id: 1,
          chain_id: 'nw-loop',
          type: 'TimelineActionDeclared',
          outcome: 'TimelineOutcomeSuccess',
          visibility: 'TimelineVisibilityPublic',
          narrative_window_id: 'nw-loop',
          action_id: 'nw-loop-attack',
          narrative_kind: 'card_played',
          visual_kind: 'card',
          card_role: 'attack',
          actor_user_id: 'p1',
          target_user_ids: ['p2'],
          cards: [buildCard({ id: 'fire-slash', name: '火焰斩', element: 'Fire', damage: 2 })],
        },
        {
          event_id: 3,
          turn_id: 1,
          chain_id: 'nw-loop',
          type: 'TimelineActionDeclared',
          outcome: 'TimelineOutcomeSuccess',
          visibility: 'TimelineVisibilityPublic',
          narrative_window_id: 'nw-loop',
          action_id: 'nw-loop-counter-1',
          combat_id: 'nw-loop-c1',
          narrative_kind: 'card_played',
          visual_kind: 'card',
          card_role: 'counter',
          actor_user_id: 'p2',
          target_user_ids: ['p3'],
          cards: [buildCard({ id: 'water-slash', name: '水涟斩', element: 'Water', damage: 3 })],
        },
        {
          event_id: 4,
          turn_id: 1,
          chain_id: 'nw-loop',
          type: 'TimelineActionDeclared',
          outcome: 'TimelineOutcomeSuccess',
          visibility: 'TimelineVisibilityPublic',
          narrative_window_id: 'nw-loop',
          action_id: 'nw-loop-counter-2',
          combat_id: 'nw-loop-c2',
          narrative_kind: 'card_played',
          visual_kind: 'card',
          card_role: 'counter',
          actor_user_id: 'p3',
          target_user_ids: ['p4'],
          cards: [buildCard({ id: 'light-slash', name: '光辉斩', element: 'Light', damage: 2 })],
        },
        {
          event_id: 5,
          turn_id: 1,
          chain_id: 'nw-loop',
          type: 'TimelineActionDeclared',
          outcome: 'TimelineOutcomeSuccess',
          visibility: 'TimelineVisibilityPublic',
          narrative_window_id: 'nw-loop',
          action_id: 'nw-loop-counter-3',
          combat_id: 'nw-loop-c3',
          narrative_kind: 'card_played',
          visual_kind: 'card',
          card_role: 'counter',
          actor_user_id: 'p4',
          target_user_ids: ['p1'],
          cards: [buildCard({ id: 'earth-slash', name: '地裂斩', element: 'Earth', damage: 4 })],
        },
      ],
    })

    const { container } = render(BattleZone)

    const loop = container.querySelector('[data-testid="narrative-combat-loop"]')
    expect(loop).not.toBeNull()
    expect(container.querySelectorAll('.narrative-loop-actor')).toHaveLength(4)
    expect(container.querySelectorAll('.narrative-loop-link')).toHaveLength(4)
    expect(container.querySelectorAll('.narrative-loop-action--red.narrative-loop-link')).toHaveLength(1)
    expect(container.querySelectorAll('.narrative-loop-action--gold.narrative-loop-link')).toHaveLength(3)
    expect(container.querySelectorAll('.narrative-loop-card')).toHaveLength(4)
    expect(screen.getAllByText('风之剑圣').length).toBeGreaterThan(0)
    expect(screen.getAllByText('封印师').length).toBeGreaterThan(0)
    expect(screen.getAllByText('天使').length).toBeGreaterThan(0)
    expect(screen.getAllByText('狂战士').length).toBeGreaterThan(0)
    expect(container.querySelector('.narrative-stack-lane')).toBeNull()
    expect(container.querySelector('.narrative-mist-layer')).toBeNull()
    expect(container.querySelector('.narrative-actors')).toBeNull()
  })

  it('renders a backend action flow as the central loop', async () => {
    const snapshotStore = useSnapshotStore()
    const battleFxStore = useBattleFxStore()
    snapshotStore.setCharacters([
      { id: 'wind_sword_saint', name: '风之剑圣', title: '', faction: '', skills: [] },
      { id: 'sealer', name: '封印师', title: '', faction: '', skills: [] },
      { id: 'angel', name: '天使', title: '', faction: '', skills: [] },
      { id: 'berserker', name: '狂战士', title: '', faction: '', skills: [] },
    ])
    snapshotStore.updateGameState({
      turn_stage: 'ActionExecution',
      current_player: 'p1',
      has_performed_startup: true,
      players: {
        p1: buildPlayer({ id: 'p1', name: '风之剑圣玩家', role: 'wind_sword_saint', camp: 'Blue' }),
        p2: buildPlayer({ id: 'p2', name: '封印师玩家', role: 'sealer', camp: 'Red' }),
        p3: buildPlayer({ id: 'p3', name: '天使玩家', role: 'angel', camp: 'Blue' }),
        p4: buildPlayer({ id: 'p4', name: '狂战士玩家', role: 'berserker', camp: 'Red' }),
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

    battleFxStore.applyStructuredTimelineNarrative({
      room_id: 'ROOM1',
      seq_start: 1,
      seq_end: 1,
      is_replay: false,
      events: [{
        event_id: 1,
        turn_id: 1,
        chain_id: 'nw-flow',
        type: 'TimelineEffectResolved',
        outcome: 'TimelineOutcomeSuccess',
        visibility: 'TimelineVisibilityPublic',
        narrative_window_id: 'nw-flow',
        action_id: 'nw-flow-a1-attack',
        narrative_kind: 'action_started',
        visual_kind: 'action_marker',
        actor_user_id: 'p1',
        action_type: 'attack',
      }],
      action_flows: [{
        flow_id: 'nw-flow-a1-attack',
        action_id: 'nw-flow-a1-attack',
        narrative_window_id: 'nw-flow',
        actor_user_id: 'p1',
        action_type: 'attack',
        actors: [
          { player_id: 'p1', order: 1 },
          { player_id: 'p2', order: 2 },
          { player_id: 'p3', order: 3 },
          { player_id: 'p4', order: 4 },
        ],
        edges: [
          {
            id: 'edge-1',
            order: 1,
            from_user_id: 'p1',
            to_user_id: 'p2',
            phase: 'attack',
            cards: [buildCard({ id: 'fire-slash', name: '火焰斩', element: 'Fire', damage: 2 })],
            outcome: 'hit',
            damage: 2,
          },
          {
            id: 'edge-2',
            order: 2,
            from_user_id: 'p2',
            to_user_id: 'p3',
            phase: 'counter',
            cards: [buildCard({ id: 'water-slash', name: '水涟斩', element: 'Water', damage: 3 })],
            outcome: 'pending',
          },
          {
            id: 'edge-3',
            order: 3,
            from_user_id: 'p3',
            to_user_id: 'p4',
            phase: 'counter',
            cards: [buildCard({ id: 'light-slash', name: '光辉斩', element: 'Light', damage: 2 })],
            outcome: 'pending',
          },
          {
            id: 'edge-4',
            order: 4,
            from_user_id: 'p4',
            to_user_id: 'p1',
            phase: 'counter',
            cards: [buildCard({ id: 'earth-slash', name: '地裂斩', element: 'Earth', damage: 4 })],
            outcome: 'pending',
          },
        ],
        logs: [
          { order: 1, text: '攻击 风之剑圣 -> 封印师 | 伤害: 2' },
          { order: 2, text: '应战 封印师 -> 天使' },
          { order: 3, text: '应战 天使 -> 狂战士' },
          { order: 4, text: '应战 狂战士 -> 风之剑圣' },
        ],
      }],
    })

    const { container } = render(BattleZone)

    expect(battleFxStore.actionNarrative).toBeNull()
    expect(container.querySelector('[data-testid="narrative-combat-loop"]')).not.toBeNull()
    expect(container.querySelectorAll('.narrative-loop-actor')).toHaveLength(4)
    expect(container.querySelectorAll('.narrative-loop-link')).toHaveLength(4)
    expect(container.querySelectorAll('.narrative-loop-card')).toHaveLength(4)
    expect(screen.getByText('伤害 2')).toBeTruthy()
    expect(screen.queryByText('战斗日志 (循环)')).toBeNull()
    expect(container.querySelector('.narrative-stack-lane')).toBeNull()
  })

  it('renders backend magic edges with inserted skill nodes', async () => {
    const snapshotStore = useSnapshotStore()
    const battleFxStore = useBattleFxStore()
    snapshotStore.setCharacters([
      { id: 'sealer', name: '封印师', title: '', faction: '', skills: [] },
      { id: 'berserker', name: '狂战士', title: '', faction: '', skills: [] },
      { id: 'angel', name: '天使', title: '', faction: '', skills: [] },
    ])
    snapshotStore.updateGameState({
      turn_stage: 'ActionExecution',
      current_player: 'p1',
      has_performed_startup: true,
      players: {
        p1: buildPlayer({ id: 'p1', name: '封印师玩家', role: 'sealer', camp: 'Blue' }),
        p2: buildPlayer({ id: 'p2', name: '狂战士玩家', role: 'berserker', camp: 'Red' }),
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

    battleFxStore.applyStructuredTimelineNarrative({
      room_id: 'ROOM1',
      seq_start: 1,
      seq_end: 1,
      is_replay: false,
      events: [{
        event_id: 1,
        turn_id: 1,
        chain_id: 'nw-magic',
        type: 'TimelineEffectResolved',
        outcome: 'TimelineOutcomeSuccess',
        visibility: 'TimelineVisibilityPublic',
        narrative_window_id: 'nw-magic',
        action_id: 'nw-magic-a1-magic',
        narrative_kind: 'action_started',
        visual_kind: 'action_marker',
        actor_user_id: 'p1',
        action_type: 'magic',
      }],
      action_flows: [{
        flow_id: 'nw-magic-a1-magic',
        action_id: 'nw-magic-a1-magic',
        narrative_window_id: 'nw-magic',
        actor_user_id: 'p1',
        action_type: 'magic',
        actors: [
          { player_id: 'p1', order: 1 },
          { player_id: 'p2', order: 2 },
          { player_id: 'p3', order: 3 },
        ],
        edges: [{
          id: 'magic-edge',
          order: 1,
          from_user_id: 'p1',
          to_user_id: 'p2',
          phase: 'magic',
          cards: [buildCard({ id: 'water-seal', name: '水之封印', type: 'Magic', element: 'Water', damage: 0 })],
          node_ids: ['skill-node', 'skill-node-duplicate'],
          outcome: 'resolved',
        }],
        nodes: [
          {
            id: 'skill-node',
            order: 1,
            kind: 'skill',
            actor_user_id: 'p3',
            target_user_ids: ['p1'],
            anchor_edge_id: 'magic-edge',
            skill_name: '神圣庇护',
            effect_text: '响应法术结算',
            outcome: 'resolved',
          },
          {
            id: 'skill-node-duplicate',
            order: 2,
            kind: 'skill',
            actor_user_id: 'p3',
            target_user_ids: ['p1'],
            anchor_edge_id: 'magic-edge',
            skill_name: '神圣庇护',
            effect_text: '响应法术结算完成',
            outcome: 'resolved',
          },
        ],
        logs: [
          { order: 1, text: '法术 封印师 -> 狂战士 | 水之封印 | 已结算' },
          { order: 2, text: '技能 天使：神圣庇护' },
        ],
      }],
    })

    const { container } = render(BattleZone)

    expect(container.querySelector('[data-testid="narrative-combat-loop"]')).not.toBeNull()
    expect(container.querySelector('.narrative-loop-action--phase-magic')).not.toBeNull()
    expect(container.querySelector('[data-narrative-flow-node-id="skill-node"]')).not.toBeNull()
    expect(container.querySelector('[data-narrative-flow-node-id="skill-node-duplicate"]')).toBeNull()
    expect(container.querySelectorAll('.narrative-loop-node--skill')).toHaveLength(1)
    expect(screen.queryByText('行动日志')).toBeNull()
    expect(screen.getByText('水之封印')).toBeTruthy()
    expect(screen.getByText('神圣庇护')).toBeTruthy()
    expect(screen.queryByText(/响应法术结算/)).toBeNull()
    const skillNode = container.querySelector<HTMLElement>('[data-narrative-flow-node-id="skill-node"]')
    const magicPacket = container.querySelector<HTMLElement>('[data-narrative-packet-id="packet-magic-edge"]')
    expect(skillNode).not.toBeNull()
    expect(magicPacket).not.toBeNull()
    expect(rectIntersects(inflateRect(loopNodeRect(skillNode!), 4), loopPacketRect(magicPacket!))).toBe(false)
    for (const actor of Array.from(container.querySelectorAll<HTMLElement>('.narrative-loop-actor'))) {
      expect(rectIntersects(inflateRect(loopNodeRect(skillNode!), 4), loopActorRect(actor))).toBe(false)
    }
  })

  it('places attack skill nodes on their anchored attack line', async () => {
    const snapshotStore = useSnapshotStore()
    const battleFxStore = useBattleFxStore()
    snapshotStore.setCharacters([
      { id: 'blade_master', name: '剑术大师', title: '', faction: '', skills: [] },
      { id: 'berserker', name: '狂战士', title: '', faction: '', skills: [] },
    ])
    snapshotStore.updateGameState({
      turn_stage: 'ActionExecution',
      current_player: 'p1',
      has_performed_startup: true,
      players: {
        p1: buildPlayer({ id: 'p1', name: '剑术大师玩家', role: 'blade_master', camp: 'Blue' }),
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

    battleFxStore.applyStructuredTimelineNarrative({
      room_id: 'ROOM1',
      seq_start: 1,
      seq_end: 1,
      is_replay: false,
      events: [],
      action_flows: [{
        flow_id: 'nw-gale-a1-attack',
        action_id: 'nw-gale-a1-attack',
        narrative_window_id: 'nw-gale',
        actor_user_id: 'p1',
        action_type: 'attack',
        actors: [
          { player_id: 'p1', order: 1 },
          { player_id: 'p2', order: 2 },
        ],
        edges: [{
          id: 'gale-attack-edge',
          order: 1,
          from_user_id: 'p1',
          to_user_id: 'p2',
          phase: 'attack',
          cards: [buildCard({ id: 'water-slash', name: '水涟斩', element: 'Water', damage: 2 })],
          node_ids: ['gale-node', 'trap-node'],
          outcome: 'resolved',
        }],
        nodes: [
          {
            id: 'gale-node',
            order: 1,
            kind: 'skill',
            actor_user_id: 'p1',
            target_user_ids: ['p2'],
            anchor_edge_id: 'gale-attack-edge',
            skill_name: '疾风技',
            effect_text: '额外+1攻击行动',
            outcome: 'resolved',
          },
          {
            id: 'trap-node',
            order: 2,
            kind: 'skill',
            actor_user_id: 'p2',
            target_user_ids: ['p1'],
            anchor_edge_id: 'gale-attack-edge',
            skill_name: '闪光陷阱',
            effect_text: '',
            outcome: 'resolved',
          },
        ],
        logs: [],
      }],
    })

    const { container } = render(BattleZone)

    const skillNode = container.querySelector<HTMLElement>('[data-narrative-flow-node-id="gale-node"]')
    const trapNode = container.querySelector<HTMLElement>('[data-narrative-flow-node-id="trap-node"]')
    const packet = container.querySelector<HTMLElement>('[data-narrative-packet-id="packet-gale-attack-edge"]')
    expect(skillNode).not.toBeNull()
    expect(trapNode).not.toBeNull()
    expect(packet).not.toBeNull()
    expect(skillNode?.classList.contains('narrative-loop-node--anchored')).toBe(true)
    expect(trapNode?.classList.contains('narrative-loop-node--anchored')).toBe(true)
    expect(screen.getByText('疾风技')).toBeTruthy()
    expect(screen.getByText('闪光陷阱')).toBeTruthy()
    expect(screen.getByText('额外+1攻击行动')).toBeTruthy()

    const skillNodes = [skillNode!, trapNode!]
    for (const node of skillNodes) {
      const center = loopNodeCenter(node)
      expect(distanceToAttackCurve(center)).toBeLessThan(18)
      expect(rectIntersects(inflateRect(loopNodeRect(node), 4), loopPacketRect(packet!))).toBe(false)
    }
    expect(rectIntersects(inflateRect(loopNodeRect(skillNode!), 4), loopNodeRect(trapNode!))).toBe(false)
  })

  it('places multiple anchored skill nodes in separate non-overlapping edge slots', async () => {
    const snapshotStore = useSnapshotStore()
    const battleFxStore = useBattleFxStore()
    snapshotStore.setCharacters([
      { id: 'sealer', name: '封印师', title: '', faction: '', skills: [] },
      { id: 'berserker', name: '狂战士', title: '', faction: '', skills: [] },
      { id: 'angel', name: '天使', title: '', faction: '', skills: [] },
      { id: 'wind_sword_saint', name: '风之剑圣', title: '', faction: '', skills: [] },
    ])
    snapshotStore.updateGameState({
      turn_stage: 'ActionExecution',
      current_player: 'p1',
      has_performed_startup: true,
      players: {
        p1: buildPlayer({ id: 'p1', name: '封印师玩家', role: 'sealer', camp: 'Blue' }),
        p2: buildPlayer({ id: 'p2', name: '狂战士玩家', role: 'berserker', camp: 'Red' }),
        p3: buildPlayer({ id: 'p3', name: '天使玩家', role: 'angel', camp: 'Blue' }),
        p4: buildPlayer({ id: 'p4', name: '风之剑圣玩家', role: 'wind_sword_saint', camp: 'Red' }),
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

    battleFxStore.applyStructuredTimelineNarrative({
      room_id: 'ROOM1',
      seq_start: 1,
      seq_end: 1,
      is_replay: false,
      events: [],
      action_flows: [{
        flow_id: 'nw-skill-collision-a1-magic',
        action_id: 'nw-skill-collision-a1-magic',
        narrative_window_id: 'nw-skill-collision',
        actor_user_id: 'p1',
        action_type: 'magic',
        actors: [
          { player_id: 'p1', order: 1 },
          { player_id: 'p2', order: 2 },
          { player_id: 'p3', order: 3 },
          { player_id: 'p4', order: 4 },
        ],
        edges: [{
          id: 'crowded-magic-edge',
          order: 1,
          from_user_id: 'p1',
          to_user_id: 'p2',
          phase: 'magic',
          cards: [buildCard({ id: 'water-seal', name: '水之封印', type: 'Magic', element: 'Water', damage: 0 })],
          node_ids: ['skill-a', 'skill-b'],
          outcome: 'resolved',
        }],
        nodes: [
          {
            id: 'skill-a',
            order: 1,
            kind: 'skill',
            actor_user_id: 'p3',
            target_user_ids: ['p1'],
            anchor_edge_id: 'crowded-magic-edge',
            skill_name: '神圣庇护',
            effect_text: '响应法术结算',
            outcome: 'resolved',
          },
          {
            id: 'skill-b',
            order: 2,
            kind: 'skill',
            actor_user_id: 'p4',
            target_user_ids: ['p1'],
            anchor_edge_id: 'crowded-magic-edge',
            skill_name: '剑心',
            effect_text: '响应法术结算',
            outcome: 'resolved',
          },
        ],
        logs: [],
      }],
    })

    const { container } = render(BattleZone)

    const packet = container.querySelector<HTMLElement>('[data-narrative-packet-id="packet-crowded-magic-edge"]')
    const skillNodes = Array.from(container.querySelectorAll<HTMLElement>('.narrative-loop-node--skill'))
    expect(packet).not.toBeNull()
    expect(skillNodes).toHaveLength(2)
    const nodeRects = skillNodes.map(node => loopNodeRect(node))
    expect(rectIntersects(inflateRect(nodeRects[0]!, 4), nodeRects[1]!)).toBe(false)
    for (const nodeRect of nodeRects) {
      expect(rectIntersects(inflateRect(nodeRect, 4), loopPacketRect(packet!))).toBe(false)
      for (const actor of Array.from(container.querySelectorAll<HTMLElement>('.narrative-loop-actor'))) {
        expect(rectIntersects(inflateRect(nodeRect, 4), loopActorRect(actor))).toBe(false)
      }
    }
  })

  it('keeps passive skill nodes anchored to empty skill edges away from related attack cards', async () => {
    const snapshotStore = useSnapshotStore()
    const battleFxStore = useBattleFxStore()
    snapshotStore.setCharacters([
      { id: 'berserker', name: '狂战士', title: '', faction: '', skills: [] },
      { id: 'wind_sword_saint', name: '风之剑圣', title: '', faction: '', skills: [] },
    ])
    snapshotStore.updateGameState({
      turn_stage: 'ActionExecution',
      current_player: 'p1',
      has_performed_startup: true,
      players: {
        p1: buildPlayer({ id: 'p1', name: '狂战士玩家', role: 'berserker', camp: 'Red' }),
        p2: buildPlayer({ id: 'p2', name: '风之剑圣玩家', role: 'wind_sword_saint', camp: 'Blue' }),
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

    battleFxStore.applyStructuredTimelineNarrative({
      room_id: 'ROOM1',
      seq_start: 1,
      seq_end: 1,
      is_replay: false,
      events: [],
      action_flows: [{
        flow_id: 'nw-berserk-a1-attack',
        action_id: 'nw-berserk-a1-attack',
        narrative_window_id: 'nw-berserk',
        actor_user_id: 'p1',
        action_type: 'attack',
        actors: [
          { player_id: 'p1', order: 1 },
          { player_id: 'p2', order: 2 },
        ],
        edges: [
          {
            id: 'attack-edge',
            order: 1,
            from_user_id: 'p1',
            to_user_id: 'p2',
            phase: 'attack',
            cards: [buildCard({ id: 'thunder-slash', name: '雷光斩', element: 'Thunder', damage: 2 })],
            outcome: 'hit',
            damage: 4,
          },
          {
            id: 'berserk-skill-edge',
            order: 2,
            from_user_id: 'p1',
            to_user_id: 'p2',
            phase: 'skill',
            node_ids: ['berserk-node'],
            outcome: 'resolved',
          },
        ],
        nodes: [{
          id: 'berserk-node',
          order: 2,
          kind: 'skill',
          actor_user_id: 'p1',
          target_user_ids: ['p2'],
          anchor_edge_id: 'berserk-skill-edge',
          skill_name: '狂化',
          effect_text: '攻击伤害增加',
          outcome: 'resolved',
        }],
        logs: [],
      }],
    })

    const { container } = render(BattleZone)

    const attackPacket = container.querySelector<HTMLElement>('[data-narrative-packet-id="packet-attack-edge"]')
    const skillNode = container.querySelector<HTMLElement>('[data-narrative-flow-node-id="berserk-node"]')
    expect(attackPacket).not.toBeNull()
    expect(skillNode).not.toBeNull()
    expect(screen.getByText('狂化')).toBeTruthy()
    expect(rectIntersects(inflateRect(loopNodeRect(skillNode!), 4), loopPacketRect(attackPacket!))).toBe(false)
  })

  it('filters redundant backend resolution and anchored effect nodes', async () => {
    const snapshotStore = useSnapshotStore()
    const battleFxStore = useBattleFxStore()
    snapshotStore.setCharacters([
      { id: 'wind_sword_saint', name: '风之剑圣', title: '', faction: '', skills: [] },
      { id: 'berserker', name: '狂战士', title: '', faction: '', skills: [] },
      { id: 'angel', name: '天使', title: '', faction: '', skills: [] },
    ])
    snapshotStore.updateGameState({
      turn_stage: 'ActionExecution',
      current_player: 'p1',
      has_performed_startup: true,
      players: {
        p1: buildPlayer({ id: 'p1', name: '风之剑圣玩家', role: 'wind_sword_saint', camp: 'Blue' }),
        p2: buildPlayer({ id: 'p2', name: '狂战士玩家', role: 'berserker', camp: 'Red' }),
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

    battleFxStore.applyStructuredTimelineNarrative({
      room_id: 'ROOM1',
      seq_start: 1,
      seq_end: 1,
      is_replay: false,
      events: [{
        event_id: 1,
        turn_id: 1,
        chain_id: 'nw-miss',
        type: 'TimelineEffectResolved',
        outcome: 'TimelineOutcomeSuccess',
        visibility: 'TimelineVisibilityPublic',
        narrative_window_id: 'nw-miss',
        action_id: 'nw-miss-a1-attack',
        narrative_kind: 'action_started',
        visual_kind: 'action_marker',
        actor_user_id: 'p1',
        action_type: 'attack',
      }],
      action_flows: [{
        flow_id: 'nw-miss-a1-attack',
        action_id: 'nw-miss-a1-attack',
        narrative_window_id: 'nw-miss',
        actor_user_id: 'p1',
        action_type: 'attack',
        actors: [
          { player_id: 'p1', order: 1 },
          { player_id: 'p2', order: 2 },
          { player_id: 'p3', order: 3 },
        ],
        edges: [{
          id: 'miss-edge',
          order: 1,
          from_user_id: 'p1',
          to_user_id: 'p2',
          phase: 'attack',
          cards: [buildCard({ id: 'water-slash', name: '水涟斩', element: 'Water', damage: 2 })],
          node_ids: ['legacy-resolution', 'legacy-effect', 'skill-node'],
          outcome: 'miss',
          label: '未命中',
        }],
        nodes: [
          {
            id: 'legacy-resolution',
            order: 1,
            kind: 'resolution',
            actor_user_id: 'p1',
            target_user_ids: ['p2'],
            anchor_edge_id: 'miss-edge',
            outcome: 'miss',
            label: '未命中',
          },
          {
            id: 'legacy-effect',
            order: 2,
            kind: 'effect',
            actor_user_id: 'p1',
            target_user_ids: ['p2'],
            anchor_edge_id: 'miss-edge',
            outcome: 'resolved',
            label: '已结算',
          },
          {
            id: 'skill-node',
            order: 3,
            kind: 'skill',
            actor_user_id: 'p3',
            target_user_ids: ['p1'],
            anchor_edge_id: 'miss-edge',
            skill_name: '神圣庇护',
            effect_text: '响应攻击结算',
            outcome: 'resolved',
          },
        ],
        logs: [{ order: 1, text: '攻击 风之剑圣 -> 狂战士 | 未命中' }],
      }],
    })

    const { container } = render(BattleZone)

    expect(container.querySelector('[data-testid="narrative-combat-loop"]')).not.toBeNull()
    expect(screen.getByText('未命中')).toBeTruthy()
    expect(container.querySelector('.narrative-loop-node--resolution')).toBeNull()
    expect(container.querySelector('[data-narrative-flow-node-id="legacy-resolution"]')).toBeNull()
    expect(container.querySelector('[data-narrative-flow-node-id="legacy-effect"]')).toBeNull()
    expect(container.querySelector('[data-narrative-flow-node-id="skill-node"]')).not.toBeNull()
    expect(screen.getByText('神圣庇护')).toBeTruthy()
  })

  it('merges backend card nodes into edge game cards', async () => {
    const snapshotStore = useSnapshotStore()
    const battleFxStore = useBattleFxStore()
    snapshotStore.setCharacters([
      { id: 'wind_sword_saint', name: '风之剑圣', title: '', faction: '', skills: [] },
      { id: 'berserker', name: '狂战士', title: '', faction: '', skills: [] },
      { id: 'assassin', name: '暗杀者', title: '', faction: '', skills: [] },
    ])
    snapshotStore.updateGameState({
      turn_stage: 'ActionExecution',
      current_player: 'p1',
      has_performed_startup: true,
      players: {
        p1: buildPlayer({ id: 'p1', name: '风之剑圣玩家', role: 'wind_sword_saint', camp: 'Blue' }),
        p2: buildPlayer({ id: 'p2', name: '狂战士玩家', role: 'berserker', camp: 'Red' }),
        p3: buildPlayer({ id: 'p3', name: '暗杀者玩家', role: 'assassin', camp: 'Blue' }),
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

    battleFxStore.applyStructuredTimelineNarrative({
      room_id: 'ROOM1',
      seq_start: 1,
      seq_end: 1,
      is_replay: false,
      events: [],
      action_flows: [{
        flow_id: 'nw-card-node-a1-attack',
        action_id: 'nw-card-node-a1-attack',
        narrative_window_id: 'nw-card-node',
        actor_user_id: 'p1',
        action_type: 'attack',
        actors: [
          { player_id: 'p1', order: 1 },
          { player_id: 'p2', order: 2 },
          { player_id: 'p3', order: 3 },
        ],
        edges: [
          {
            id: 'attack-edge',
            order: 1,
            from_user_id: 'p1',
            to_user_id: 'p2',
            phase: 'attack',
            cards: [buildCard({ id: 'water-slash', name: '水涟斩', element: 'Water' })],
          },
          {
            id: 'counter-edge',
            order: 2,
            from_user_id: 'p2',
            to_user_id: 'p3',
            phase: 'counter',
            node_ids: ['counter-card-node'],
          },
          {
            id: 'counter-edge-2',
            order: 3,
            from_user_id: 'p2',
            to_user_id: 'p3',
            phase: 'counter',
            cards: [buildCard({ id: 'light-slash', name: '光辉斩', element: 'Light', damage: 2 })],
          },
        ],
        nodes: [{
          id: 'counter-card-node',
          order: 1,
          kind: 'card',
          actor_user_id: 'p2',
          target_user_ids: ['p3'],
          cards: [buildCard({ id: 'earth-slash', name: '地裂斩', element: 'Earth', damage: 3 })],
          label: '地裂斩',
        }],
        logs: [],
      }],
    })

    const { container } = render(BattleZone)

    expect(container.querySelector('[data-testid="narrative-combat-loop"]')).not.toBeNull()
    expect(container.querySelector('.narrative-loop-node--card')).toBeNull()
    expect(container.querySelector('[data-narrative-flow-node-id="counter-card-node"]')).toBeNull()
    expect(container.querySelectorAll('.narrative-loop-card')).toHaveLength(3)
    expect(container.querySelectorAll('.narrative-loop-card .game-card')).toHaveLength(3)
    expect(container.querySelectorAll('.narrative-loop-packet')).toHaveLength(3)
    const cards = Array.from(container.querySelectorAll<HTMLElement>('.narrative-loop-card'))
    const notes = Array.from(container.querySelectorAll<HTMLElement>('.narrative-loop-note'))
    const packets = Array.from(container.querySelectorAll<HTMLElement>('.narrative-loop-packet'))
    const counterCard = cards[1]
    const counterNote = notes[1]
    const repeatedCounterCard = cards[2]
    const counterPacket = packets[1]
    const repeatedCounterPacket = packets[2]
    expect(counterCard).toBeDefined()
    expect(counterNote).toBeDefined()
    expect(repeatedCounterCard).toBeDefined()
    expect(counterPacket).toBeDefined()
    expect(repeatedCounterPacket).toBeDefined()
    expect(counterPacket?.contains(counterNote || null)).toBe(true)
    expect(counterNote?.style.getPropertyValue('--loop-label-x')).toBe(counterCard?.style.getPropertyValue('--loop-card-x'))
    expect(counterNote?.style.getPropertyValue('--loop-label-y')).toBe(counterCard?.style.getPropertyValue('--loop-card-y'))
    const counterCardPoint = `${counterCard?.style.getPropertyValue('--loop-card-x')}:${counterCard?.style.getPropertyValue('--loop-card-y')}`
    const repeatedCounterCardPoint = `${repeatedCounterCard?.style.getPropertyValue('--loop-card-x')}:${repeatedCounterCard?.style.getPropertyValue('--loop-card-y')}`
    const counterPacketPoint = `${counterPacket?.style.getPropertyValue('--loop-packet-x')}:${counterPacket?.style.getPropertyValue('--loop-packet-y')}`
    const repeatedCounterPacketPoint = `${repeatedCounterPacket?.style.getPropertyValue('--loop-packet-x')}:${repeatedCounterPacket?.style.getPropertyValue('--loop-packet-y')}`
    expect(repeatedCounterCardPoint).not.toBe(counterCardPoint)
    expect(repeatedCounterPacketPoint).not.toBe(counterPacketPoint)
    expect(screen.getByText('地裂斩')).toBeTruthy()
  })

  it('does not render duplicate empty attack edge for a counter card edge', async () => {
    const snapshotStore = useSnapshotStore()
    const battleFxStore = useBattleFxStore()
    snapshotStore.setCharacters([
      { id: 'berserker', name: '狂战士', title: '', faction: '', skills: [] },
      { id: 'wind_sword_saint', name: '风之剑圣', title: '', faction: '', skills: [] },
      { id: 'sealer', name: '封印师', title: '', faction: '', skills: [] },
    ])
    snapshotStore.updateGameState({
      turn_stage: 'ActionExecution',
      current_player: 'p1',
      has_performed_startup: true,
      players: {
        p1: buildPlayer({ id: 'p1', name: '狂战士玩家', role: 'berserker', camp: 'Blue' }),
        p2: buildPlayer({ id: 'p2', name: '风之剑圣玩家', role: 'wind_sword_saint', camp: 'Red' }),
        p3: buildPlayer({ id: 'p3', name: '封印师玩家', role: 'sealer', camp: 'Blue' }),
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

    battleFxStore.applyStructuredTimelineNarrative({
      room_id: 'ROOM1',
      seq_start: 1,
      seq_end: 1,
      is_replay: false,
      events: [],
      action_flows: [{
        flow_id: 'nw-duplicate-counter-a1-attack',
        action_id: 'nw-duplicate-counter-a1-attack',
        narrative_window_id: 'nw-duplicate-counter',
        actor_user_id: 'p1',
        action_type: 'attack',
        actors: [
          { player_id: 'p1', order: 1 },
          { player_id: 'p2', order: 2 },
          { player_id: 'p3', order: 3 },
        ],
        edges: [
          {
            id: 'attack-edge',
            order: 1,
            from_user_id: 'p1',
            to_user_id: 'p2',
            phase: 'attack',
            cards: [buildCard({ id: 'fire-slash', name: '火焰斩', element: 'Fire' })],
            outcome: 'miss',
            label: '未命中',
          },
          {
            id: 'counter-edge',
            order: 2,
            from_user_id: 'p2',
            to_user_id: 'p3',
            phase: 'counter',
            node_ids: ['counter-card-node'],
            outcome: 'resolved',
          },
          {
            id: 'duplicate-empty-attack-edge',
            order: 3,
            from_user_id: 'p2',
            to_user_id: 'p3',
            phase: 'attack',
            outcome: 'resolved',
          },
        ],
        nodes: [{
          id: 'counter-card-node',
          order: 1,
          kind: 'card',
          actor_user_id: 'p2',
          target_user_ids: ['p3'],
          anchor_edge_id: 'counter-edge',
          cards: [buildCard({ id: 'counter-fire-slash', name: '火焰斩', element: 'Fire' })],
          label: '火焰斩',
        }],
        logs: [],
      }],
    })

    const { container } = render(BattleZone)

    const links = Array.from(container.querySelectorAll<SVGGElement>('.narrative-loop-link'))
    expect(links).toHaveLength(2)
    expect(links.filter(link => link.classList.contains('narrative-loop-action--phase-attack'))).toHaveLength(1)
    expect(links.filter(link => link.classList.contains('narrative-loop-action--phase-counter'))).toHaveLength(1)
    expect(container.querySelectorAll('.narrative-loop-card')).toHaveLength(2)
    expect(container.querySelector('[data-narrative-packet-id="packet-duplicate-empty-attack-edge"]')).toBeNull()
  })

  it('merges empty attack miss resolution into the matching counter edge', async () => {
    const snapshotStore = useSnapshotStore()
    const battleFxStore = useBattleFxStore()
    snapshotStore.setCharacters([
      { id: 'berserker', name: '狂战士', title: '', faction: '', skills: [] },
      { id: 'wind_sword_saint', name: '风之剑圣', title: '', faction: '', skills: [] },
      { id: 'sealer', name: '封印师', title: '', faction: '', skills: [] },
    ])
    snapshotStore.updateGameState({
      turn_stage: 'ActionExecution',
      current_player: 'p1',
      has_performed_startup: true,
      players: {
        p1: buildPlayer({ id: 'p1', name: '狂战士玩家', role: 'berserker', camp: 'Blue' }),
        p2: buildPlayer({ id: 'p2', name: '风之剑圣玩家', role: 'wind_sword_saint', camp: 'Red' }),
        p3: buildPlayer({ id: 'p3', name: '封印师玩家', role: 'sealer', camp: 'Blue' }),
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

    battleFxStore.applyStructuredTimelineNarrative({
      room_id: 'ROOM1',
      seq_start: 1,
      seq_end: 1,
      is_replay: false,
      events: [],
      action_flows: [{
        flow_id: 'nw-counter-miss-a1-attack',
        action_id: 'nw-counter-miss-a1-attack',
        narrative_window_id: 'nw-counter-miss',
        actor_user_id: 'p1',
        action_type: 'attack',
        actors: [
          { player_id: 'p1', order: 1 },
          { player_id: 'p2', order: 2 },
          { player_id: 'p3', order: 3 },
        ],
        edges: [
          {
            id: 'attack-edge',
            order: 1,
            from_user_id: 'p1',
            to_user_id: 'p2',
            phase: 'attack',
            cards: [buildCard({ id: 'fire-slash', name: '火焰斩', element: 'Fire' })],
            outcome: 'miss',
            label: '未命中',
          },
          {
            id: 'counter-edge',
            order: 2,
            from_user_id: 'p2',
            to_user_id: 'p3',
            phase: 'counter',
            cards: [buildCard({ id: 'counter-water-slash', name: '水涟斩', element: 'Water' })],
            outcome: 'pending',
          },
          {
            id: 'duplicate-empty-counter-miss-edge',
            order: 3,
            from_user_id: 'p2',
            to_user_id: 'p3',
            phase: 'attack',
            outcome: 'miss',
            label: '未命中',
          },
        ],
        logs: [],
      }],
    })

    const { container } = render(BattleZone)

    const links = Array.from(container.querySelectorAll<SVGGElement>('.narrative-loop-link'))
    expect(links).toHaveLength(2)
    expect(links.filter(link => link.classList.contains('narrative-loop-action--phase-attack'))).toHaveLength(1)
    const counterLinks = links.filter(link => link.classList.contains('narrative-loop-action--phase-counter'))
    expect(counterLinks).toHaveLength(1)
    expect(counterLinks[0]?.classList.contains('narrative-loop-action--outcome-miss')).toBe(true)
    expect(container.querySelectorAll('.narrative-loop-card')).toHaveLength(2)
    expect(container.querySelector('[data-narrative-packet-id="packet-duplicate-empty-counter-miss-edge"]')).toBeNull()
    expect(screen.getAllByText('未命中')).toHaveLength(2)
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
    expect(container.querySelector('.narrative-skill-token__shine')).not.toBeNull()
    await waitFor(() => {
      expect(container.querySelectorAll('.narrative-mist--skill').length).toBeGreaterThanOrEqual(2)
    })
  })

  it('shows elemental seal cards directly in the center without step groups', async () => {
    const snapshotStore = useSnapshotStore()
    const battleFxStore = useBattleFxStore()
    snapshotStore.setCharacters([
      { id: 'sealer', name: '封印师', title: '', faction: '', skills: [] },
      { id: 'berserker', name: '狂战士', title: '', faction: '', skills: [] },
    ])
    snapshotStore.updateGameState({
      turn_stage: 'ActionExecution',
      current_player: 'p1',
      has_performed_startup: true,
      players: {
        p1: buildPlayer({ id: 'p1', name: '封印师玩家', role: 'sealer', camp: 'Blue' }),
        p2: buildPlayer({
          id: 'p2',
          name: '狂战士玩家',
          role: 'berserker',
          camp: 'Red',
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

    battleFxStore.applyStructuredTimelineNarrative({
      room_id: 'ROOM1',
      seq_start: 1,
      seq_end: 3,
      is_replay: false,
      events: [
        {
          event_id: 1,
          turn_id: 1,
          chain_id: 'nw-t1-p1',
          type: 'TimelineEffectResolved',
          outcome: 'TimelineOutcomeSuccess',
          visibility: 'TimelineVisibilityPublic',
          narrative_window_id: 'nw-t1-p1',
          action_id: 'nw-t1-p1-a1-skill',
          narrative_kind: 'skill_declared',
          visual_kind: 'none',
          actor_user_id: 'p1',
          target_user_ids: ['p2'],
          skill_name: '水之封印',
        },
        {
          event_id: 2,
          turn_id: 1,
          chain_id: 'nw-t1-p1',
          type: 'TimelineActionDeclared',
          outcome: 'TimelineOutcomeSuccess',
          visibility: 'TimelineVisibilityPublic',
          narrative_window_id: 'nw-t1-p1',
          action_id: 'nw-t1-p1-a1-skill',
          narrative_kind: 'card_played',
          visual_kind: 'card',
          card_role: 'field_effect',
          actor_user_id: 'p1',
          cards: [buildCard({ id: 'water-seal-card', name: '水涟斩', element: 'Water' })],
        },
        {
          event_id: 3,
          turn_id: 1,
          chain_id: 'nw-t1-p1',
          type: 'TimelineEffectResolved',
          outcome: 'TimelineOutcomeSuccess',
          visibility: 'TimelineVisibilityPublic',
          narrative_window_id: 'nw-t1-p1',
          action_id: 'nw-t1-p1-a1-skill',
          narrative_kind: 'field_effect_applied',
          visual_kind: 'effect_token',
          effect_type: 'SealWater',
          actor_user_id: 'p1',
          target_user_ids: ['p2'],
        },
      ],
    })

    const { container } = render(BattleZone, {
      props: {
        actorSeatPositions: {
          p1: 0,
          p2: 3,
        },
      },
    })

    expect(screen.getByText('水涟斩')).toBeTruthy()
    expect(container.querySelector('.narrative-seal-card-stage')).not.toBeNull()
    expect(container.querySelector('.narrative-played-card--field_effect')).not.toBeNull()
    expect(container.querySelector('.narrative-step-group')).toBeNull()
    expect(container.querySelector('.narrative-step-group__items')).toBeNull()
    expect(container.querySelector('[data-narrative-skill-id]')).toBeNull()
    await waitFor(() => {
      expect(container.querySelectorAll('.narrative-mist--skill').length).toBeGreaterThanOrEqual(2)
    })
  })

  it('shows damage on the loop target and flow note', async () => {
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

    expect(container.querySelector('[data-testid="narrative-combat-loop"]')).not.toBeNull()
    expect(container.querySelector('.narrative-loop-actor__damage')?.textContent).toContain('-2')
    expect(screen.getByText('伤害 2')).toBeTruthy()
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
