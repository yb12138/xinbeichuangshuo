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

  it('builds a real-time action narrative from combat cues, cards and damage', () => {
    const battleFxStore = useBattleFxStore()

    battleFxStore.beginActionNarrative('p1')
    expect(battleFxStore.actionNarrative?.featuredActorId).toBe('p1')

    battleFxStore.addCombatCue('p1', 'p2', 'attack')
    battleFxStore.addNarrativeCard('p1', buildCard('火焰斩'), 'attack')

    expect(battleFxStore.actionNarrative?.opposedActorIds).toContain('p2')
    expect(battleFxStore.actionNarrative?.playedCards[0]).toMatchObject({
      playerId: 'p1',
      targetId: 'p2',
      actionType: 'attack',
    })

    battleFxStore.addCombatCue('p2', 'p3', 'counter')
    expect(battleFxStore.actionNarrative?.featuredActorId).toBe('p2')
    expect(battleFxStore.actionNarrative?.settledActorIds).toContain('p1')
    expect(battleFxStore.actionNarrative?.opposedActorIds).toContain('p3')

    battleFxStore.addNarrativeDamage('p2', 'p3', 2, 'Attack')
    expect(battleFxStore.actionNarrative?.events.map(event => event.label)).toContain('造成 2 点伤害')

    battleFxStore.clearActionNarrative()
    expect(battleFxStore.actionNarrative).toBeNull()
  })

  it('builds readable playback steps from structured narrative events', () => {
    const battleFxStore = useBattleFxStore()

    battleFxStore.applyStructuredTimelineNarrative({
      room_id: 'ROOM1',
      seq_start: 1,
      seq_end: 5,
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
          cards: [buildCard('破空箭')],
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
          cards: [buildCard('地裂斩')],
        },
        {
          event_id: 4,
          turn_id: 1,
          chain_id: 'nw-t1-p1',
          type: 'TimelineEffectResolved',
          outcome: 'TimelineOutcomeSuccess',
          visibility: 'TimelineVisibilityPublic',
          narrative_window_id: 'nw-t1-p1',
          action_id: 'nw-t1-p1-a2-skill',
          narrative_kind: 'skill_triggered',
          visual_kind: 'skill_token',
          skill_phase: 'triggered',
          actor_user_id: 'p3',
          skill_name: '水影',
          effect_text: '弃置1张牌后承受伤害',
        },
        {
          event_id: 5,
          turn_id: 1,
          chain_id: 'nw-t1-p1',
          type: 'TimelineCombatResolved',
          outcome: 'TimelineOutcomeSuccess',
          visibility: 'TimelineVisibilityPublic',
          narrative_window_id: 'nw-t1-p1',
          action_id: 'nw-t1-p1-a1-attack',
          combat_id: 'nw-t1-p1-c1-counter',
          narrative_kind: 'damage_dealt',
          visual_kind: 'damage',
          actor_user_id: 'p2',
          target_user_ids: ['p3'],
          damage: 3,
          damage_type: 'Attack',
        },
        {
          event_id: 6,
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

    expect(battleFxStore.narrativePlayback?.steps.map(step => step.kind)).toEqual([
      'combat',
      'response',
      'skill',
      'damage',
      'effect',
    ])
    expect(battleFxStore.narrativePlayback?.activeStepId).toBe('card-1')
    expect(battleFxStore.narrativePlayback?.steps[0]).toMatchObject({
      label: '发起攻击',
      status: 'active',
      itemIds: ['card-1'],
    })

    vi.advanceTimersByTime(950)

    expect(battleFxStore.narrativePlayback?.steps[0]?.status).toBe('completed')
    expect(battleFxStore.narrativePlayback?.steps[1]).toMatchObject({
      label: '应战',
      status: 'active',
      itemIds: ['card-2'],
    })
    expect(battleFxStore.narrativePlayback?.steps[4]).toMatchObject({
      kind: 'effect',
      label: '效果：未命中',
    })
  })

  it('projects placed field effect cards without duplicate effect tokens', () => {
    const battleFxStore = useBattleFxStore()

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
          cards: [buildCard('水涟斩')],
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
          field_card: {
            card: buildCard('水涟斩'),
            owner_id: 'p2',
            source_id: 'p1',
            mode: 'Effect',
            effect: 'SealWater',
            field_hook: '',
            locked: false,
            duration: 0,
          },
        },
      ],
    })

    expect(battleFxStore.actionNarrative?.playedCards).toHaveLength(1)
    expect(battleFxStore.actionNarrative?.playedCards[0]).toMatchObject({
      actionType: 'field_effect',
      targetId: 'p2',
    })
    expect(battleFxStore.actionNarrative?.events).toHaveLength(0)
    expect(battleFxStore.narrativePlayback?.steps).toHaveLength(1)
    expect(battleFxStore.narrativePlayback?.steps[0]).toMatchObject({
      kind: 'skill',
      label: '施加封印',
      targetIds: ['p2'],
    })
  })
})
