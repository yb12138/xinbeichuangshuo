import { describe, expect, it } from 'vitest'
import { extractGameplayEventsFromTimeline } from '../gameplayTimeline'
import type { TimelineEvent } from '../protocol'

describe('extractGameplayEventsFromTimeline', () => {
  it('converts supported gameplay timeline events back into gameplay events', () => {
    const events: TimelineEvent[] = [
      {
        event_id: 1,
        turn_id: 1,
        chain_id: 'chain_1',
        type: 'TimelineEffectResolved',
        outcome: 'TimelineOutcomeSuccess',
        visibility: 'TimelineVisibilityPublic',
        message: 'Alice 发动了技能',
        gameplay_type: 'log',
      },
      {
        event_id: 2,
        turn_id: 1,
        chain_id: 'chain_2',
        type: 'TimelineCombatResolved',
        outcome: 'TimelineOutcomeSuccess',
        visibility: 'TimelineVisibilityPublic',
        actor_user_id: 'p1',
        actor_name: 'Alice',
        target_user_ids: ['p2'],
        target_name: 'Bob',
        damage: 2,
        message: '造成 2 点伤害',
        gameplay_type: 'damage_dealt',
      },
    ]

    expect(extractGameplayEventsFromTimeline(events)).toEqual([
      {
        event_type: 'log',
        message: 'Alice 发动了技能',
      },
      {
        event_type: 'damage_dealt',
        source_id: 'p1',
        source_name: 'Alice',
        target_id: 'p2',
        target_name: 'Bob',
        damage: 2,
        damage_type: '',
        message: '造成 2 点伤害',
      },
    ])
  })

  it('keeps error events in timeline replay but still skips prompt-like interrupts', () => {
    const events: TimelineEvent[] = [
      {
        event_id: 3,
        turn_id: 1,
        chain_id: 'chain_3',
        type: 'TimelineInterruptRaised',
        outcome: 'TimelineOutcomeSuccess',
        visibility: 'TimelineVisibilityPublic',
        gameplay_type: 'prompt',
      },
      {
        event_id: 4,
        turn_id: 1,
        chain_id: 'chain_4',
        type: 'TimelineEffectResolved',
        outcome: 'TimelineOutcomeSuccess',
        visibility: 'TimelineVisibilityPublic',
        message: '技能发动失败',
        gameplay_type: 'error',
      },
    ]

    expect(extractGameplayEventsFromTimeline(events)).toEqual([
      {
        event_type: 'error',
        message: '技能发动失败',
      },
    ])
  })

  it('rebuilds card reveal, combat cue, draw and action-step events from typed fields', () => {
    const events: TimelineEvent[] = [
      {
        event_id: 5,
        turn_id: 2,
        chain_id: 'chain_5',
        type: 'TimelineActionDeclared',
        outcome: 'TimelineOutcomeSuccess',
        visibility: 'TimelineVisibilityPublic',
        actor_user_id: 'p1',
        actor_name: 'Alice',
        action_type: 'magic',
        cards: [
          {
            id: 'card-1',
            name: '魔弹',
            type: 'Magic',
            element: 'Dark',
            damage: 2,
            description: 'test',
          },
        ],
        hidden: false,
        gameplay_type: 'card_revealed',
      },
      {
        event_id: 6,
        turn_id: 2,
        chain_id: 'chain_6',
        type: 'TimelineActionDeclared',
        outcome: 'TimelineOutcomeSuccess',
        visibility: 'TimelineVisibilityPublic',
        actor_user_id: 'p1',
        target_user_ids: ['p2'],
        cue_phase: 'counter',
        gameplay_type: 'combat_cue',
      },
      {
        event_id: 7,
        turn_id: 2,
        chain_id: 'chain_7',
        type: 'TimelineEffectResolved',
        outcome: 'TimelineOutcomeSuccess',
        visibility: 'TimelineVisibilityPublic',
        actor_user_id: 'p2',
        actor_name: 'Bob',
        draw_count: 2,
        reason: 'action',
        gameplay_type: 'draw_cards',
      },
      {
        event_id: 8,
        turn_id: 2,
        chain_id: 'chain_8',
        type: 'TimelineEffectResolved',
        outcome: 'TimelineOutcomeSuccess',
        visibility: 'TimelineVisibilityPublic',
        message: 'Alice 使用魔弹',
        detail_kind: 'summary',
        gameplay_type: 'action_step',
      },
    ]

    expect(extractGameplayEventsFromTimeline(events)).toEqual([
      {
        event_type: 'card_revealed',
        player_id: 'p1',
        player_name: 'Alice',
        cards: [
          {
            id: 'card-1',
            name: '魔弹',
            type: 'Magic',
            element: 'Dark',
            damage: 2,
            description: 'test',
          },
        ],
        action_type: 'magic',
        hidden: false,
      },
      {
        event_type: 'combat_cue',
        attacker_id: 'p1',
        target_id: 'p2',
        phase: 'counter',
      },
      {
        event_type: 'draw_cards',
        player_id: 'p2',
        player_name: 'Bob',
        draw_count: 2,
        reason: 'action',
      },
      {
        event_type: 'action_step',
        line: 'Alice 使用魔弹',
        kind: 'summary',
      },
    ])
  })
})
