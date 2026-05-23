import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it } from 'vitest'
import { useBattleFxStore } from '../battlefx.store'
import { useSessionStore } from '../session.store'
import { useSnapshotStore } from '../snapshot.store'
import type { GameStateUpdate, PlayerInfo, PlayerView } from '../../types/game'

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
})
