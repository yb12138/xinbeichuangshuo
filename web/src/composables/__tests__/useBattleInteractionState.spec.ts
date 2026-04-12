import { beforeEach, describe, expect, it } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useBattleInteractionState } from '../useBattleInteractionState'
import { useSessionStore } from '../../stores/session.store'
import { useSnapshotStore } from '../../stores/snapshot.store'
import type { AvailableSkill, Card, CharacterView, GameStateUpdate, PlayerView, SkillView } from '../../types/game'

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

