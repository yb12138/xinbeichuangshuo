import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { createGameplayMessageHandlers } from '../gameplayMessageHandlers'
import { useBattleFxStore } from '../../stores/battlefx.store'
import { useBattleReviewStore } from '../../stores/battleReview.store'
import { useInterruptStore } from '../../stores/interrupt.store'
import { useMatchLifecycleStore } from '../../stores/matchLifecycle.store'
import { useSessionStore } from '../../stores/session.store'
import { useSnapshotStore } from '../../stores/snapshot.store'
import { useTimelineStore } from '../../stores/timeline.store'
import { useUiStore } from '../../stores/ui.store'
import type { Prompt } from '../../types/game'
import type { RequireActionPayload, SyncStatePayload, TimelineNotifyPayload } from '../protocol'

function buildSyncPlayer(id: string, camp: 'Red' | 'Blue') {
  return {
    id,
    name: id === 'p1' ? 'Alice' : `Player${id.slice(1)}`,
    camp,
    role: 'hero',
    hand_count: 0,
    max_hand: 6,
    exclusive_card_count: 0,
    hand: [],
    exclusive_cards: [],
    field: [],
    heal: 0,
    max_heal: 5,
    gem: 0,
    crystal: 0,
    is_active: false,
    buffs: [],
    tokens: {},
  }
}

function buildHandlers() {
  const interruptStore = useInterruptStore()
  const sessionStore = useSessionStore()
  const snapshotStore = useSnapshotStore()
  const timelineStore = useTimelineStore()
  const uiStore = useUiStore()
  const battleFxStore = useBattleFxStore()
  const battleReviewStore = useBattleReviewStore()
  const matchLifecycleStore = useMatchLifecycleStore()

  return {
    interruptStore,
    sessionStore,
    snapshotStore,
    timelineStore,
    uiStore,
    battleFxStore,
    battleReviewStore,
    matchLifecycleStore,
    handlers: createGameplayMessageHandlers({
      interruptStore,
      sessionStore,
      snapshotStore,
      timelineStore,
      uiStore,
      battleFxStore,
      battleReviewStore,
      matchLifecycleStore,
    }),
  }
}

describe('createGameplayMessageHandlers', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.runOnlyPendingTimers()
    vi.useRealTimers()
  })

  it('applies SyncState to snapshot and marks the match as started', () => {
    const { handlers, interruptStore, sessionStore, snapshotStore } = buildHandlers()
    sessionStore.setRoomInfo('ROOM1', 'p1', 'Red', 'hero')
    interruptStore.setPrompt({
      type: 'confirm',
      player_id: 'p1',
      message: '旧提示',
      options: [],
      min: 1,
      max: 1,
      presentation: {
        kind: 'branch_select',
        numeric_base: 0,
      },
    })

    const payload: SyncStatePayload = {
      room_state: 'Playing',
      turn_stage: 'Main',
      turn_player_id: 'p1',
      has_performed_startup: true,
      morale_red: 15,
      morale_blue: 13,
      cups_red: 0,
      cups_blue: 1,
      stones_red: [2, 1],
      stones_blue: [1, 3],
      deck_count: 18,
      discard_count: 4,
      available_skills: [],
      characters: [],
      players: [
        {
          id: 'p1',
          name: 'Alice',
          camp: 'Red',
          role: 'hero',
          hand_count: 2,
          max_hand: 6,
          exclusive_card_count: 0,
          hand: [],
          exclusive_cards: [],
          field: [],
          heal: 1,
          max_heal: 5,
          gem: 1,
          crystal: 2,
          is_active: true,
          buffs: [],
          tokens: {},
        },
      ],
    }

    handlers.handleSyncState(payload)

    expect(sessionStore.gameStarted).toBe(true)
    expect(snapshotStore.turnStage).toBe('Main')
    expect(snapshotStore.players.p1?.name).toBe('Alice')
    expect(snapshotStore.redGems).toBe(2)
    expect(interruptStore.currentPrompt).toBeNull()
  })

  it('pops the current turn player when action stage begins', () => {
    const { handlers, battleFxStore, sessionStore } = buildHandlers()
    sessionStore.setRoomInfo('ROOM1', 'p1', 'Red', 'hero')
    sessionStore.updateRoomPlayers([
      { id: 'p1', name: 'Alice', camp: 'Red', char_role: 'hero', ready: true, is_online: true },
      { id: 'p2', name: 'Bob', camp: 'Red', char_role: 'hero', ready: true, is_online: true },
      { id: 'p3', name: 'Cara', camp: 'Blue', char_role: 'hero', ready: true, is_online: true },
      { id: 'p4', name: 'Dora', camp: 'Blue', char_role: 'hero', ready: true, is_online: true },
      { id: 'p5', name: 'Evan', camp: 'Blue', char_role: 'hero', ready: true, is_online: true },
      { id: 'p6', name: 'Faye', camp: 'Red', char_role: 'hero', ready: true, is_online: true },
    ], 'p1')

    handlers.handleSyncState({
      room_state: 'Playing',
      turn_stage: 'ActionExecution',
      turn_player_id: 'p4',
      has_performed_startup: true,
      morale_red: 15,
      morale_blue: 15,
      cups_red: 0,
      cups_blue: 0,
      stones_red: [0, 0],
      stones_blue: [0, 0],
      deck_count: 30,
      discard_count: 0,
      available_skills: [],
      characters: [],
      players: [
        buildSyncPlayer('p1', 'Red'),
        buildSyncPlayer('p2', 'Red'),
        buildSyncPlayer('p3', 'Blue'),
        buildSyncPlayer('p4', 'Blue'),
        buildSyncPlayer('p5', 'Blue'),
        buildSyncPlayer('p6', 'Red'),
      ],
    })

    expect(battleFxStore.initiatorFocus?.playerId).toBe('p4')
    expect(battleFxStore.initiatorFocus?.side).toBe('right')
  })

  it('routes RequireAction into prompt or waiting state based on target player', () => {
    const { handlers, interruptStore, sessionStore } = buildHandlers()
    sessionStore.setRoomInfo('ROOM1', 'p1', 'Red', 'hero')

    const prompt: Prompt = {
      type: 'confirm',
      player_id: 'p2',
      message: '请选择',
      options: [{ id: 'confirm', label: '确认', button_label: '确认' }],
      min: 1,
      max: 1,
      presentation: {
        kind: 'branch_select',
        numeric_base: 0,
      },
    }

    const waitingPayload: RequireActionPayload = {
      interrupt_type: 'WaitChoice',
      target_user_id: 'p2',
      timeout: 0,
      msg: '等待对手操作',
      prompt,
    }

    handlers.handleRequireAction(waitingPayload)

    expect(interruptStore.currentPrompt).toBeNull()
    expect(interruptStore.waitingFor).toBe('p2')

    handlers.handleRequireAction({
      ...waitingPayload,
      target_user_id: 'p1',
      prompt: { ...prompt, player_id: 'p1' },
    })

    expect(interruptStore.currentPrompt?.player_id).toBe('p1')
    expect(interruptStore.waitingFor).toBe('')
  })

  it('focuses the prompt actor when a prompt arrives', () => {
    const { handlers, battleFxStore, sessionStore } = buildHandlers()
    sessionStore.setRoomInfo('ROOM1', 'p1', 'Red', 'hero')

    battleFxStore.startSkillInitiatorFocus('p1', 'skill')
    expect(battleFxStore.initiatorFocus?.playerId).toBe('p1')

    handlers.handleRequireAction({
      interrupt_type: 'WaitChoice',
      target_user_id: 'p1',
      timeout: 0,
      msg: '请选择目标',
      prompt: {
        type: 'confirm',
        player_id: 'p1',
        message: '【挑衅】请选择一名目标对手：',
        choice_type: 'hero_taunt_target',
        options: [{ id: 'p2', label: '目标', button_label: '目标' }],
        min: 1,
        max: 1,
        presentation: {
          kind: 'target_picker',
          numeric_base: 0,
        },
      },
    })

    vi.advanceTimersByTime(1100)

    expect(battleFxStore.initiatorFocus?.playerId).toBe('p1')
    expect(battleFxStore.initiatorFocus?.mode).toBe('skill')
  })

  it('focuses the waiting response player even when the prompt is for another client', () => {
    const { handlers, battleFxStore, interruptStore, sessionStore } = buildHandlers()
    sessionStore.setRoomInfo('ROOM1', 'p1', 'Red', 'hero')

    handlers.handleRequireAction({
      interrupt_type: 'WaitChoice',
      target_user_id: 'p2',
      timeout: 0,
      msg: '请选择应战方式',
      prompt: {
        type: 'confirm',
        player_id: 'p2',
        message: '请选择应战、防御或命中',
        attacker_id: 'p1',
        options: [
          { id: 'counter', label: '应战', button_label: '应战' },
          { id: 'defend', label: '防御', button_label: '防御' },
          { id: 'take', label: '命中', button_label: '命中' },
        ],
        min: 1,
        max: 1,
        presentation: {
          kind: 'branch_select',
          numeric_base: 0,
        },
      },
    })

    expect(interruptStore.currentPrompt).toBeNull()
    expect(interruptStore.waitingFor).toBe('p2')
    expect(battleFxStore.initiatorFocus?.playerId).toBe('p2')
    expect(battleFxStore.initiatorFocus?.mode).toBe('response')
  })

  it('focuses the player who resolves a combat response cue', () => {
    const { handlers, battleFxStore } = buildHandlers()

    handlers.handleGameplayEvent({
      event_type: 'combat_cue',
      attacker_id: 'p1',
      target_id: 'p2',
      phase: 'defend',
    })

    expect(battleFxStore.initiatorFocus?.playerId).toBe('p2')
    expect(battleFxStore.initiatorFocus?.mode).toBe('response')
  })

  it('does not add detail action steps to action summaries', () => {
    const { handlers, battleReviewStore } = buildHandlers()

    handlers.handleGameplayEvent({
      event_type: 'action_step',
      line: '中间过程：Bob承受伤害',
      kind: 'detail',
    })

    expect(battleReviewStore.actionSummaryLines).toEqual([])
    expect(battleReviewStore.battleFeed).toEqual([])
  })

  it('adds summary action steps to action summaries and battle feed', () => {
    const { handlers, battleReviewStore } = buildHandlers()

    handlers.handleGameplayEvent({
      event_type: 'action_step',
      line: '回合1：Alice 使用攻击【火焰斩】 -> Bob；Bob 受到2点伤害',
      kind: 'summary',
    })

    expect(battleReviewStore.actionSummaryLines).toEqual([
      '回合1：Alice 使用攻击【火焰斩】 -> Bob；Bob 受到2点伤害',
    ])
    expect(battleReviewStore.battleFeed).toHaveLength(1)
    expect(battleReviewStore.battleFeed[0]?.title).toBe(
      '回合1：Alice 使用攻击【火焰斩】 -> Bob；Bob 受到2点伤害',
    )
  })

  it('does not let skill-looking logs drive battle UI', () => {
    const { handlers, battleFxStore, battleReviewStore, interruptStore, snapshotStore } = buildHandlers()
    snapshotStore.updateGameState({
      turn_stage: 'ActionExecution',
      current_player: 'p1',
      has_performed_startup: true,
      players: {
        p1: buildSyncPlayer('p1', 'Red'),
        p2: buildSyncPlayer('p2', 'Blue'),
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

    handlers.handleGameplayEvent({
      event_type: 'log',
      message: 'Alice 发动 [苍炎法典]，先对 Bob 后对自己各造成2点法术伤害',
    })

    expect(interruptStore.skillEffectToast).toBe('')
    expect(battleFxStore.skillAnnouncements).toHaveLength(0)
    expect(battleReviewStore.battleFeed).toHaveLength(0)
    expect(battleReviewStore.logs).toEqual([
      'Alice 发动 [苍炎法典]，先对 Bob 后对自己各造成2点法术伤害',
    ])
  })

  it('uses skill_activated events for battle announcements', () => {
    const { handlers, battleFxStore, battleReviewStore, snapshotStore } = buildHandlers()
    snapshotStore.updateGameState({
      turn_stage: 'ActionExecution',
      current_player: 'p1',
      has_performed_startup: true,
      players: {
        p1: buildSyncPlayer('p1', 'Red'),
        p2: buildSyncPlayer('p2', 'Blue'),
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

    handlers.handleGameplayEvent({
      event_type: 'skill_activated',
      player_id: 'p1',
      player_name: 'Alice',
      skill_id: 'sage_arcane_codex',
      skill_name: '苍炎法典',
      effect_text: '先对 Bob 后对自己各造成2点法术伤害',
      target_ids: ['p2'],
    })

    expect(battleFxStore.skillAnnouncements).toHaveLength(1)
    expect(battleFxStore.skillAnnouncements[0]).toMatchObject({
      actorId: 'p1',
      actorName: 'Alice',
      skillName: '苍炎法典',
      effectText: '先对 Bob 后对自己各造成2点法术伤害',
      phase: 'featured',
    })
    expect(battleReviewStore.battleFeed[battleReviewStore.battleFeed.length - 1]).toMatchObject({
      type: 'skill',
      title: 'Alice 发动「苍炎法典」',
      detail: '先对 Bob 后对自己各造成2点法术伤害',
      actorId: 'p1',
    })
  })

  it('records state_delta morale and resource changes without log hints', () => {
    const { handlers, battleReviewStore } = buildHandlers()

    handlers.handleGameplayEvent({
      event_type: 'state_delta',
      reason: 'DamageDealt',
      deltas: [
        {
          type: 'morale',
          scope: 'team',
          camp: 'Red',
          field: 'morale',
          before: 15,
          after: 13,
          value: -2,
          reason: 'DamageDealt',
        },
        {
          type: 'team_gem',
          scope: 'team',
          camp: 'Blue',
          field: 'gem',
          before: 0,
          after: 1,
          value: 1,
          reason: 'DamageDealt',
        },
      ],
    })

    expect(battleReviewStore.moraleChanges).toHaveLength(1)
    expect(battleReviewStore.moraleChanges[0]).toMatchObject({
      camp: 'Red',
      before: 15,
      after: 13,
      delta: -2,
      source: 'DamageDealt',
    })
    expect(battleReviewStore.battleFeed.map(entry => entry.title)).toEqual([
      '红方士气 -2',
      '蓝方阵营宝石 +1',
    ])
  })

  it('replays damage timeline payloads into effects without adding battle feed entries', () => {
    const { handlers, timelineStore, battleFxStore } = buildHandlers()

    const payload: TimelineNotifyPayload = {
      room_id: 'ROOM1',
      seq_start: 1,
      seq_end: 1,
      is_replay: false,
      events: [
        {
          event_id: 1,
          turn_id: 1,
          chain_id: 'chain_1',
          type: 'TimelineCombatResolved',
          outcome: 'TimelineOutcomeSuccess',
          visibility: 'TimelineVisibilityPublic',
          gameplay_type: 'damage_dealt',
          actor_user_id: 'p1',
          actor_name: 'Alice',
          target_user_ids: ['p2'],
          target_name: 'Bob',
          damage: 3,
          damage_type: 'Attack',
          message: '造成 3 点伤害',
        },
      ],
    }

    handlers.handleNotifyTimeline(payload)

    expect(timelineStore.entries).toHaveLength(0)
    expect(battleFxStore.damageEffects).toHaveLength(1)
    expect(battleFxStore.damageEffects[0]?.targetId).toBe('p2')
    expect(battleFxStore.damageEffects[0]?.damage).toBe(3)
  })

  it('drives the action narrative directly from real-time gameplay events', () => {
    const { handlers, battleFxStore, sessionStore } = buildHandlers()
    sessionStore.setRoomInfo('ROOM1', 'p1', 'Red', 'hero')

    handlers.handleSyncState({
      room_state: 'Playing',
      turn_stage: 'ActionExecution',
      turn_player_id: 'p1',
      has_performed_startup: true,
      morale_red: 15,
      morale_blue: 15,
      cups_red: 0,
      cups_blue: 0,
      stones_red: [0, 0],
      stones_blue: [0, 0],
      deck_count: 30,
      discard_count: 0,
      available_skills: [],
      characters: [],
      players: [
        buildSyncPlayer('p1', 'Red'),
        buildSyncPlayer('p2', 'Blue'),
        buildSyncPlayer('p3', 'Blue'),
      ],
    })

    expect(battleFxStore.actionNarrative?.currentActionPlayerId).toBe('p1')
    expect(battleFxStore.actionNarrative?.featuredActorId).toBe('p1')

    handlers.handleGameplayEvent({
      event_type: 'combat_cue',
      attacker_id: 'p1',
      target_id: 'p2',
      phase: 'attack',
    })
    handlers.handleGameplayEvent({
      event_type: 'card_revealed',
      player_id: 'p1',
      player_name: 'Alice',
      action_type: 'attack',
      cards: [{
        id: 'fire-slash',
        name: '火焰斩',
        type: 'Attack',
        element: 'Fire',
        damage: 2,
        description: '',
      }],
      hidden: false,
    })
    handlers.handleGameplayEvent({
      event_type: 'damage_dealt',
      source_id: 'p1',
      source_name: 'Alice',
      target_id: 'p2',
      target_name: 'Player2',
      damage: 2,
      damage_type: 'Attack',
    })

    expect(battleFxStore.actionNarrative?.opposedActorIds).toContain('p2')
    expect(battleFxStore.actionNarrative?.playedCards[0]).toMatchObject({
      playerId: 'p1',
      targetId: 'p2',
      actionType: 'attack',
    })
    expect(battleFxStore.actionNarrative?.events.map(event => event.label)).toEqual(
      expect.arrayContaining(['行动回合', '攻击', '造成 2 点伤害']),
    )
    expect(battleFxStore.damageEffects).toHaveLength(1)
  })

  it('keeps the action narrative through non-action phases until the next player starts acting', () => {
    const { handlers, battleFxStore, sessionStore } = buildHandlers()
    sessionStore.setRoomInfo('ROOM1', 'p1', 'Red', 'hero')

    const baseState = {
      room_state: 'Playing' as const,
      has_performed_startup: true,
      morale_red: 15,
      morale_blue: 15,
      cups_red: 0,
      cups_blue: 0,
      stones_red: [0, 0] as [number, number],
      stones_blue: [0, 0] as [number, number],
      deck_count: 30,
      discard_count: 0,
      available_skills: [],
      characters: [],
      players: [
        buildSyncPlayer('p1', 'Red'),
        buildSyncPlayer('p2', 'Blue'),
      ],
    }

    handlers.handleSyncState({
      ...baseState,
      turn_stage: 'ActionExecution',
      turn_player_id: 'p1',
    })
    handlers.handleGameplayEvent({
      event_type: 'skill_activated',
      player_id: 'p1',
      player_name: 'Alice',
      skill_id: 'water_seal',
      skill_name: '水之封印',
      effect_text: '目标本回合无法使用水系牌',
      target_ids: ['p2'],
    })

    expect(battleFxStore.actionNarrative?.events.map(event => event.label)).toContain('水之封印：目标本回合无法使用水系牌')

    handlers.handleSyncState({
      ...baseState,
      turn_stage: 'Settlement',
      turn_player_id: 'p1',
      combat_stage: '',
      subflow: '',
    })

    expect(battleFxStore.actionNarrative?.events.map(event => event.label)).toContain('水之封印：目标本回合无法使用水系牌')

    handlers.handleSyncState({
      ...baseState,
      turn_stage: 'ActionExecution',
      turn_player_id: 'p2',
    })

    expect(battleFxStore.actionNarrative?.currentActionPlayerId).toBe('p2')
    expect(battleFxStore.actionNarrative?.events.map(event => event.label)).toContain('行动回合')
    expect(battleFxStore.actionNarrative?.events.map(event => event.label)).not.toContain('水之封印：目标本回合无法使用水系牌')
  })

  it('projects structured timeline replay into the central narrative without replaying transient effects', () => {
    const { handlers, battleFxStore, battleReviewStore } = buildHandlers()

    const payload: TimelineNotifyPayload = {
      room_id: 'ROOM1',
      seq_start: 1,
      seq_end: 5,
      is_replay: true,
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
          actor_name: 'Alice',
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
          combat_id: 'nw-t1-p1-c1-combat',
          narrative_kind: 'combat_declared',
          visual_kind: 'none',
          actor_user_id: 'p1',
          target_user_ids: ['p2'],
          cue_phase: 'attack',
          gameplay_type: 'combat_cue',
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
          combat_id: 'nw-t1-p1-c1-combat',
          narrative_kind: 'card_played',
          visual_kind: 'card',
          card_role: 'attack',
          actor_user_id: 'p1',
          actor_name: 'Alice',
          target_user_ids: ['p2'],
          action_type: 'attack',
          cards: [{
            id: 'earth-slash',
            name: '地裂斩',
            type: 'Attack',
            element: 'Earth',
            damage: 2,
            description: '',
          }],
          gameplay_type: 'card_revealed',
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
          narrative_kind: 'skill_declared',
          visual_kind: 'skill_token',
          skill_phase: 'declared',
          actor_user_id: 'p3',
          actor_name: '封印师',
          target_user_ids: ['p2'],
          skill_name: '水之封印',
          effect_text: '放置水系封印',
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
          combat_id: 'nw-t1-p1-c1-combat',
          narrative_kind: 'damage_dealt',
          visual_kind: 'damage',
          actor_user_id: 'p1',
          target_user_ids: ['p2'],
          damage: 2,
          damage_type: 'Attack',
          gameplay_type: 'damage_dealt',
        },
        {
          event_id: 6,
          turn_id: 1,
          chain_id: 'nw-t1-p1',
          type: 'TimelineEffectResolved',
          outcome: 'TimelineOutcomeSuccess',
          visibility: 'TimelineVisibilityPublic',
          narrative_window_id: 'nw-t1-p1',
          action_id: 'nw-t1-p1-a2-skill',
          narrative_kind: 'field_effect_applied',
          visual_kind: 'effect_token',
          effect_type: 'SealWater',
          actor_user_id: 'p3',
          target_user_ids: ['p2'],
          field_card: {
            card: {
              id: 'seal-card',
              name: '水之封印',
              type: 'Magic',
              element: 'Water',
              damage: 0,
              description: '',
            },
            owner_id: 'p2',
            source_id: 'p3',
            mode: 'Effect',
            effect: 'SealWater',
            field_hook: 'OnBeforeAction',
            locked: false,
            duration: 0,
          },
        },
      ],
    }

    handlers.handleNotifyTimeline(payload)

    expect(battleFxStore.actionNarrative?.currentActionPlayerId).toBe('p1')
    expect(battleFxStore.actionNarrative?.playedCards).toHaveLength(1)
    expect(battleFxStore.actionNarrative?.playedCards[0]).toMatchObject({
      playerId: 'p1',
      targetId: 'p2',
      actionType: 'attack',
    })
    expect(battleFxStore.actionNarrative?.events.map(event => event.label)).toEqual(
      expect.arrayContaining(['水之封印：放置水系封印', '造成 2 点伤害']),
    )
    expect(battleFxStore.actionNarrative?.events.map(event => event.label)).not.toContain('效果：水之封印')
    expect(battleFxStore.actionNarrative?.links.some(link => link.fromId === 'p3' && link.toPlayerId === 'p2')).toBe(true)
    expect(battleFxStore.damageEffects).toHaveLength(0)
    expect(battleReviewStore.battleFeed).toHaveLength(0)
  })

  it('backfills a structured card target when the combat cue arrives after the card reveal', () => {
    const { handlers, battleFxStore } = buildHandlers()

    handlers.handleNotifyTimeline({
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
          cards: [{
            id: 'earth-slash',
            name: '地裂斩',
            type: 'Attack',
            element: 'Earth',
            damage: 2,
            description: '',
          }],
          gameplay_type: 'card_revealed',
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
          combat_id: 'nw-t1-p1-c1-combat',
          narrative_kind: 'combat_declared',
          visual_kind: 'none',
          actor_user_id: 'p1',
          target_user_ids: ['p2'],
          cue_phase: 'attack',
          gameplay_type: 'combat_cue',
        },
      ],
    })

    expect(battleFxStore.actionNarrative?.playedCards[0]).toMatchObject({
      playerId: 'p1',
      targetId: 'p2',
      actionType: 'attack',
    })
    expect(battleFxStore.actionNarrative?.links.some(link => link.fromType === 'card' && link.toPlayerId === 'p2')).toBe(true)
  })

  it('projects public discard cards into narrative playback without revealing hidden discards', () => {
    const { handlers, battleFxStore } = buildHandlers()

    handlers.handleNotifyTimeline({
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
          narrative_kind: 'action_started',
          visual_kind: 'action_marker',
          actor_user_id: 'p1',
          action_type: 'skill',
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
          card_role: 'discard',
          actor_user_id: 'p1',
          cards: [{
            id: 'water-cost',
            name: '水纹法术',
            type: 'Magic',
            element: 'Water',
            damage: 0,
            description: '',
          }],
          gameplay_type: 'card_revealed',
        },
        {
          event_id: 3,
          turn_id: 1,
          chain_id: 'nw-t1-p1',
          type: 'TimelineActionDeclared',
          outcome: 'TimelineOutcomeSuccess',
          visibility: 'TimelineVisibilityPublic',
          narrative_window_id: 'nw-t1-p1',
          action_id: 'nw-t1-p1-a1-skill',
          narrative_kind: 'card_played',
          visual_kind: 'none',
          card_role: 'discard',
          actor_user_id: 'p1',
          hidden: true,
          cards: [{
            id: 'hidden-cost',
            name: '隐藏法术',
            type: 'Magic',
            element: 'Fire',
            damage: 0,
            description: '',
          }],
          gameplay_type: 'card_revealed',
        },
      ],
    })

    expect(battleFxStore.actionNarrative?.playedCards).toHaveLength(1)
    expect(battleFxStore.actionNarrative?.playedCards[0]).toMatchObject({
      playerId: 'p1',
      actionType: 'discard',
      card: expect.objectContaining({ name: '水纹法术' }),
    })
    expect(battleFxStore.actionNarrative?.playedCards.map(card => card.card.name)).not.toContain('隐藏法术')
    expect(battleFxStore.narrativePlayback?.steps[0]).toMatchObject({
      kind: 'discard',
      label: '弃牌',
      itemIds: ['card-1'],
    })
  })
})
