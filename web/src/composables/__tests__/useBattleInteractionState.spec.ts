import { beforeEach, describe, expect, it } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useBattleInteractionState } from '../useBattleInteractionState'
import { useInterruptStore } from '../../stores/interrupt.store'
import { useSessionStore } from '../../stores/session.store'
import { useSnapshotStore } from '../../stores/snapshot.store'
import type { AvailableSkill, Card, CharacterView, FieldCard, GameStateUpdate, PlayerView, SkillView } from '../../types/game'

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

function buildPlayer(overrides: Partial<PlayerView> = {}): PlayerView {
  return {
    id: 'p1',
    name: 'P1',
    camp: 'Red',
    role: 'sealer',
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
    is_active: true,
    buffs: [],
    tokens: {},
    ...overrides,
  }
}

function buildSkill(overrides: Partial<SkillView> = {}): SkillView {
  return {
    id: 'fire_seal',
    title: '火之封印',
    description: 'test',
    type: 2,
    min_targets: 1,
    max_targets: 1,
    target_type: 2,
    cost_gem: 0,
    cost_crystal: 0,
    cost_discards: 1,
    discard_element: 'Fire',
    require_exclusive: true,
    ...overrides,
  }
}

function buildCoverFieldCard(overrides: Partial<FieldCard> = {}): FieldCard {
  return {
    card: buildCard({ id: 'cover-card-1', type: 'Attack', element: 'Fire' }),
    owner_id: 'p1',
    source_id: 'p1',
    mode: 'Cover',
    effect: 'ElfBlessing',
    field_hook: 'Manual',
    locked: false,
    duration: 0,
    ...overrides,
  }
}

function buildCharacter(overrides: Partial<CharacterView> = {}): CharacterView {
  return {
    id: 'sealer',
    name: '封印师',
    title: '幻',
    faction: '幻',
    skills: [buildSkill()],
    ...overrides,
  }
}

function buildAvailableSkill(overrides: Partial<AvailableSkill> = {}): AvailableSkill {
  return {
    id: 'fire_seal',
    title: '火之封印',
    description: 'test',
    min_targets: 1,
    max_targets: 1,
    target_type: 2,
    cost_gem: 0,
    cost_crystal: 0,
    cost_discards: 1,
    discard_element: 'Fire',
    require_exclusive: true,
    place_card: true,
    place_effect: 'SealFire',
    ...overrides,
  }
}

function buildState(overrides: Partial<GameStateUpdate> = {}): GameStateUpdate {
  return {
    turn_stage: 'ActionExecution',
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
    characters: [buildCharacter()],
    ...overrides,
  }
}

describe('useBattleInteractionState available skill source of truth', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('uses backend available_skills during own ActionExecution turn', () => {
    const sessionStore = useSessionStore()
    const snapshotStore = useSnapshotStore()
    sessionStore.setRoomInfo('ROOM', 'p1', 'Red', 'sealer')

    snapshotStore.updateGameState(buildState({
      available_skills: [buildAvailableSkill()],
    }))

    const interaction = useBattleInteractionState()
    expect(interaction.effectiveAvailableSkills.value.map((s) => s.id)).toEqual(['fire_seal'])
  })

  it('does not fall back to static character skills when backend available_skills is empty in own ActionExecution turn', () => {
    const sessionStore = useSessionStore()
    const snapshotStore = useSnapshotStore()
    sessionStore.setRoomInfo('ROOM', 'p1', 'Red', 'sealer')

    snapshotStore.updateGameState(buildState({
      players: {
        p1: buildPlayer({
          hand_count: 1,
          hand: [
            buildCard({
              exclusive_char1: 'sealer',
              exclusive_skill1: '火之封印',
            }),
          ],
        }),
      },
      available_skills: [],
      characters: [buildCharacter()],
    }))

    const interaction = useBattleInteractionState()
    expect(interaction.effectiveAvailableSkills.value).toEqual([])
  })
})

describe('useBattleInteractionState skill target candidates', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('allows angel_wall to target both ally and enemy players when target_type is any', () => {
    const sessionStore = useSessionStore()
    const snapshotStore = useSnapshotStore()
    const interruptStore = useInterruptStore()
    sessionStore.setRoomInfo('ROOM', 'p1', 'Red', 'angel')

    const angelWallSkill = buildAvailableSkill({
      id: 'angel_wall',
      title: '天使之墙',
      target_type: 5,
      min_targets: 1,
      max_targets: 1,
      discard_element: undefined,
      place_effect: 'Shield',
    })

    snapshotStore.updateGameState(buildState({
      players: {
        p1: buildPlayer({ id: 'p1', camp: 'Red', role: 'angel' }),
        p2: buildPlayer({ id: 'p2', camp: 'Red', role: 'saintess' }),
        p3: buildPlayer({ id: 'p3', camp: 'Blue', role: 'berserker' }),
      },
      available_skills: [angelWallSkill],
    }))
    interruptStore.setSelectedSkill(angelWallSkill)

    const interaction = useBattleInteractionState()
    expect(interaction.targetablePlayersForSkill.value.map((p) => p.id)).toEqual(['p1', 'p2', 'p3'])
  })

  it('builds playable blessing cards from field covers when blessings mirror is empty', () => {
    const sessionStore = useSessionStore()
    const snapshotStore = useSnapshotStore()
    sessionStore.setRoomInfo('ROOM', 'p1', 'Red', 'elf_archer')

    snapshotStore.updateGameState(buildState({
      players: {
        p1: buildPlayer({
          id: 'p1',
          role: 'elf_archer',
          hand: [buildCard({ id: 'hand-attack-1', type: 'Attack', element: 'Fire' })],
          blessings: [],
          field: [
            buildCoverFieldCard({
              card: buildCard({ id: 'bless-attack-1', type: 'Attack', element: 'Wind' }),
              effect: 'ElfBlessing',
            }),
          ],
        }),
      },
    }))

    const interaction = useBattleInteractionState()
    expect(interaction.myPlayableCards.value.map((item) => ({ id: item.card.id, source: item.source, index: item.index }))).toEqual([
      { id: 'hand-attack-1', source: 'hand', index: 0 },
      { id: 'bless-attack-1', source: 'blessing', index: 1 },
    ])
  })
})
