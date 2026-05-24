import { createPinia, setActivePinia } from 'pinia'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { useBattleFxStore } from '../battlefx.store'
import { useSessionStore } from '../session.store'
import { useSnapshotStore } from '../snapshot.store'
import type { Card, GameStateUpdate, PlayerInfo, PlayerView } from '../../types/game'

function buildCard(id: string): Card {
  return {
    id,
    name: id,
    type: 'Attack',
    element: 'Fire',
    damage: 2,
    description: '',
  }
}

function buildPlayer(overrides: Partial<PlayerView> = {}): PlayerView {
  return {
    id: 'p1',
    name: '玩家1',
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

describe('battlefx store focus side', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.runOnlyPendingTimers()
    vi.useRealTimers()
  })

  it('uses the displayed roster order to resolve focus side', () => {
    const sessionStore = useSessionStore()
    const snapshotStore = useSnapshotStore()
    const battleFxStore = useBattleFxStore()

    const roster = [
      buildPlayer({ id: 'p1', name: '玩家1', camp: 'Red' }),
      buildPlayer({ id: 'p2', name: '玩家2', camp: 'Red' }),
      buildPlayer({ id: 'p3', name: '玩家3', camp: 'Blue' }),
      buildPlayer({ id: 'p4', name: '玩家4', camp: 'Blue' }),
      buildPlayer({ id: 'p5', name: '玩家5', camp: 'Blue' }),
      buildPlayer({ id: 'p6', name: '玩家6', camp: 'Red' }),
    ]
    const players = Object.fromEntries(roster.map((player) => [player.id, player]))

    sessionStore.setRoomInfo('ROOM1', 'p1', 'Red', 'fighter')
    sessionStore.updateRoomPlayers(roster.map(buildPlayerInfo), 'p1')
    snapshotStore.updateGameState(buildState(players))

    battleFxStore.startSkillInitiatorFocus('p1', 'skill')
    expect(battleFxStore.initiatorFocus?.side).toBe('left')

    battleFxStore.startSkillInitiatorFocus('p4', 'skill')
    expect(battleFxStore.initiatorFocus?.side).toBe('right')
  })

  it('shows skill announcements as featured, settles them, then removes them', () => {
    const battleFxStore = useBattleFxStore()

    battleFxStore.addSkillAnnouncement('p1', 'Alice', '苍炎法典', '对 Bob 造成2点法术伤害')

    expect(battleFxStore.skillAnnouncements).toHaveLength(1)
    expect(battleFxStore.skillAnnouncements[0]?.phase).toBe('featured')
    expect(battleFxStore.skillAnnouncements[0]?.skillName).toBe('苍炎法典')

    vi.advanceTimersByTime(1650)

    expect(battleFxStore.skillAnnouncements[0]?.phase).toBe('settled')

    vi.advanceTimersByTime(7000)

    expect(battleFxStore.skillAnnouncements).toHaveLength(0)
  })

  it('merges repeated skill announcement logs into one richer announcement', () => {
    const battleFxStore = useBattleFxStore()

    battleFxStore.addSkillAnnouncement('p1', 'Alice', '苍炎法典', '技能发动')
    battleFxStore.addSkillAnnouncement('p1', 'Alice', '苍炎法典', '先对 Bob 后对自己各造成2点法术伤害')

    expect(battleFxStore.skillAnnouncements).toHaveLength(1)
    expect(battleFxStore.skillAnnouncements[0]?.effectText).toBe('先对 Bob 后对自己各造成2点法术伤害')
  })

  it('keeps revealed attack and magic cards queued for battlefield reading', () => {
    const battleFxStore = useBattleFxStore()

    battleFxStore.addFlyingCards([buildCard('火焰斩')], 'p1', 'Alice', 'attack', false)
    battleFxStore.addFlyingCards([buildCard('魔弹')], 'p2', 'Bob', 'magic', false)

    expect(battleFxStore.flyingCards).toHaveLength(1)
    expect(battleFxStore.flyingCards[0]?.cards[0]?.name).toBe('火焰斩')

    battleFxStore.settleFlyingCardToBattlefield(battleFxStore.flyingCards[0]?.id ?? -1)

    expect(battleFxStore.flyingCards).toHaveLength(1)
    expect(battleFxStore.flyingCards[0]?.cards[0]?.name).toBe('魔弹')
  })

  it('does not clear revealed cards when damage or response cues arrive', () => {
    const battleFxStore = useBattleFxStore()

    battleFxStore.addFlyingCards([buildCard('火焰斩')], 'p1', 'Alice', 'attack', false)
    const activeId = battleFxStore.flyingCards[0]?.id

    battleFxStore.addDamageEffect('p2', 'Bob', 2, 'Attack')
    battleFxStore.addCombatCue('p1', 'p2', 'take')

    expect(battleFxStore.flyingCards[0]?.id).toBe(activeId)
  })
})
