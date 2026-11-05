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

  it('still shows catalog action skills when backend available_skills is empty in own ActionExecution turn', () => {
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
    expect(interaction.effectiveAvailableSkills.value.map((s) => s.id)).toEqual(['fire_seal'])
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

  it('builds playable blessing cards from field covers', () => {
    const sessionStore = useSessionStore()
    const snapshotStore = useSnapshotStore()
    sessionStore.setRoomInfo('ROOM', 'p1', 'Red', 'elf_archer')

    snapshotStore.updateGameState(buildState({
      players: {
        p1: buildPlayer({
          id: 'p1',
          role: 'elf_archer',
          hand: [buildCard({ id: 'hand-attack-1', type: 'Attack', element: 'Fire' })],
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

describe('useBattleInteractionState skill target legality', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  function setup(skill: AvailableSkill, playerOverrides: Record<string, Partial<PlayerView>> = {}) {
    useSessionStore().setRoomInfo('ROOM', 'p1', 'Red', 'sealer')
    useSnapshotStore().updateGameState(buildState({
      players: {
        p1: buildPlayer({ id: 'p1', camp: 'Red', ...playerOverrides.p1 }),
        p2: buildPlayer({ id: 'p2', camp: 'Red', role: 'saintess', ...playerOverrides.p2 }),
        p3: buildPlayer({ id: 'p3', camp: 'Blue', role: 'berserker', ...playerOverrides.p3 }),
        p4: buildPlayer({ id: 'p4', camp: 'Blue', role: 'assassin', ...playerOverrides.p4 }),
      },
      available_skills: [skill],
    }))
    const interruptStore = useInterruptStore()
    interruptStore.setSelectedSkill(skill)
    const interaction = useBattleInteractionState()
    return {
      interruptStore,
      interaction,
      ids: () => interaction.targetablePlayersForSkill.value.map((p) => p.id),
    }
  }

  function plainSkill(overrides: Partial<AvailableSkill>): AvailableSkill {
    return buildAvailableSkill({
      id: 'test_skill',
      cost_discards: 0,
      discard_element: undefined,
      require_exclusive: false,
      place_card: false,
      place_effect: undefined,
      ...overrides,
    })
  }

  it.each([
    [1, ['p1']],
    [2, ['p3', 'p4']],
    [3, ['p2']],
    [4, ['p1', 'p2']],
    [5, ['p1', 'p2', 'p3', 'p4']],
    [6, ['p1', 'p2', 'p3', 'p4']],
  ])('filters candidates by target_type %i', (targetType, expected) => {
    const { ids } = setup(plainSkill({ target_type: targetType }))
    expect(ids()).toEqual(expected)
  })

  it('excludes self and targets without heal for 惩戒', () => {
    const { ids } = setup(plainSkill({ id: 'holy_lancer_punishment', target_type: 5 }), {
      p1: { heal: 3 },
      p2: { heal: 1 },
      p3: { heal: 0 },
      p4: { heal: 2 },
    })
    expect(ids()).toEqual(['p2', 'p4'])
  })

  it('excludes self for 苍炎法典 / 天火断空', () => {
    for (const id of ['bw_blazing_codex', 'bw_heavenfire_cleave']) {
      setActivePinia(createPinia())
      const { ids } = setup(plainSkill({ id, target_type: 5 }))
      expect(ids()).toEqual(['p2', 'p3', 'p4'])
    }
  })

  it('only offers players with a basic effect for 封印破碎', () => {
    const poison: FieldCard = {
      card: buildCard({ id: 'poison', type: 'Magic', name: '中毒' }),
      owner_id: 'p3',
      source_id: 'p1',
      mode: 'Effect',
      effect: 'Poison',
      field_hook: 'Manual',
      locked: false,
      duration: 0,
    }
    const { ids } = setup(plainSkill({ id: 'seal_break', target_type: 5, min_targets: 0 }), { p3: { field: [poison] } })
    expect(ids()).toEqual(['p3'])
  })

  it('never offers non-targetable skills', () => {
    const { ids } = setup(plainSkill({ target_type: 0 }))
    expect(ids()).toEqual([])
  })

  it('rejects duplicate targets and enforces min/max on confirm', () => {
    const { interruptStore, interaction } = setup(plainSkill({ id: 'angel_blessing', target_type: 6, min_targets: 1, max_targets: 2 }))
    expect(interaction.canConfirmSkill.value).toBe(false)
    interruptStore.setSkillTargetIds(['p2'])
    expect(interaction.canConfirmSkill.value).toBe(true)
    interruptStore.setSkillTargetIds(['p2', 'p3'])
    expect(interaction.canConfirmSkill.value).toBe(true)
    interruptStore.setSkillTargetIds(['p2', 'p2'])
    expect(interaction.canConfirmSkill.value).toBe(false)
    interruptStore.setSkillTargetIds(['p2', 'p3', 'p4'])
    expect(interaction.canConfirmSkill.value).toBe(false)
  })

  it('blocks confirm when a selected target is no longer a legal candidate', () => {
    const { interruptStore, interaction } = setup(plainSkill({ target_type: 2 }))
    interruptStore.setSkillTargetIds(['p2'])
    expect(interaction.canConfirmSkill.value).toBe(false)
    interruptStore.setSkillTargetIds(['p3'])
    expect(interaction.canConfirmSkill.value).toBe(true)
  })

  it('excludes stealth-form players from attack targets', () => {
    useSessionStore().setRoomInfo('ROOM', 'p1', 'Red', 'sealer')
    useSnapshotStore().updateGameState(buildState({
      players: {
        p1: buildPlayer({ id: 'p1', camp: 'Red', hand: [buildCard()] }),
        p3: buildPlayer({ id: 'p3', camp: 'Blue', role: 'berserker' }),
        p4: buildPlayer({ id: 'p4', camp: 'Blue', role: 'assassin', form: 'assassin_stealth_form' }),
      },
    }))
    const interruptStore = useInterruptStore()
    interruptStore.setActionMode('attack')
    const interaction = useBattleInteractionState()
    expect(interaction.targetablePlayers.value.map((p) => p.id)).toEqual(['p3'])
  })
})
