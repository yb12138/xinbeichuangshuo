import { beforeEach, describe, expect, it } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useTimelineStore } from '../timeline.store'
import type { TimelineNotifyPayload } from '../../network/protocol'

describe('useTimelineStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('keeps raw payloads but only displays player-meaningful replay events', () => {
    const store = useTimelineStore()
    const payload: TimelineNotifyPayload = {
      room_id: 'ROOM1',
      seq_start: 1,
      seq_end: 10,
      is_replay: false,
      events: [
        {
          event_id: 1,
          turn_id: 1,
          chain_id: 'chain_1',
          type: 'TimelineInterruptRaised',
          outcome: 'TimelineOutcomeSuccess',
          visibility: 'TimelineVisibilityPublic',
          gameplay_type: 'prompt',
          message: '请选择响应',
        },
        {
          event_id: 2,
          turn_id: 1,
          chain_id: 'chain_2',
          type: 'TimelineEffectResolved',
          outcome: 'TimelineOutcomeSuccess',
          visibility: 'TimelineVisibilityPublic',
          gameplay_type: 'action_step',
          detail_kind: 'detail',
          message: '中间过程：承受伤害',
        },
        {
          event_id: 3,
          turn_id: 1,
          chain_id: 'chain_3',
          type: 'TimelineEffectResolved',
          outcome: 'TimelineOutcomeSuccess',
          visibility: 'TimelineVisibilityPublic',
          gameplay_type: 'action_step',
          detail_kind: 'summary',
          message: '回合1：Alice 使用攻击【火焰斩】 -> Bob',
        },
        {
          event_id: 4,
          turn_id: 1,
          chain_id: 'chain_4',
          type: 'TimelineCombatResolved',
          outcome: 'TimelineOutcomeSuccess',
          visibility: 'TimelineVisibilityPublic',
          gameplay_type: 'damage_dealt',
          actor_user_id: 'p1',
          target_user_ids: ['p2'],
          damage: 2,
          message: 'Bob 受到2点伤害',
        },
        {
          event_id: 5,
          turn_id: 1,
          chain_id: 'chain_5',
          type: 'TimelineEffectResolved',
          outcome: 'TimelineOutcomeSuccess',
          visibility: 'TimelineVisibilityPublic',
          gameplay_type: 'draw_cards',
          actor_user_id: 'p2',
          actor_name: 'Bob',
          draw_count: 2,
          reason: 'damage_draw',
          message: 'damage_draw',
        },
        {
          event_id: 6,
          turn_id: 1,
          chain_id: 'chain_6',
          type: 'TimelineEffectResolved',
          outcome: 'TimelineOutcomeSuccess',
          visibility: 'TimelineVisibilityPublic',
          gameplay_type: 'draw_cards',
          actor_user_id: 'p1',
          actor_name: 'Alice',
          draw_count: 1,
          reason: 'skill_reward',
          message: 'skill_reward',
        },
        {
          event_id: 7,
          turn_id: 1,
          chain_id: 'chain_7',
          type: 'TimelineEffectResolved',
          outcome: 'TimelineOutcomeSuccess',
          visibility: 'TimelineVisibilityPublic',
          gameplay_type: 'log',
          message: '[Debug] Drive Loop: 0',
        },
        {
          event_id: 8,
          turn_id: 1,
          chain_id: 'chain_8',
          type: 'TimelineEffectResolved',
          outcome: 'TimelineOutcomeSuccess',
          visibility: 'TimelineVisibilityPublic',
          gameplay_type: 'log',
          message: 'Alice 发动 [苍炎法典]，先对 Bob 后对自己各造成2点法术伤害',
        },
        {
          event_id: 9,
          turn_id: 1,
          chain_id: 'chain_9',
          type: 'TimelineEffectResolved',
          outcome: 'TimelineOutcomeSuccess',
          visibility: 'TimelineVisibilityPublic',
          gameplay_type: 'log',
          message: 'Alice 的 [魔能反转] 生效：弃2张法术牌，对 Bob 造成1点法术伤害',
        },
        {
          event_id: 10,
          turn_id: 1,
          chain_id: 'chain_10',
          type: 'TimelineEffectResolved',
          outcome: 'TimelineOutcomeSuccess',
          visibility: 'TimelineVisibilityPublic',
          gameplay_type: 'log',
          message: '[Skill] Alice 使用了技能: 苍炎法典',
        },
        {
          event_id: 11,
          turn_id: 1,
          chain_id: 'chain_11',
          type: 'TimelineActionDeclared',
          outcome: 'TimelineOutcomeSuccess',
          visibility: 'TimelineVisibilityPublic',
          gameplay_type: 'skill_activated',
          actor_user_id: 'p1',
          actor_name: 'Alice',
          skill_id: 'sage_arcane_codex',
          skill_name: '苍炎法典',
          effect_text: '造成法术伤害',
        },
        {
          event_id: 12,
          turn_id: 1,
          chain_id: 'chain_12',
          type: 'TimelineActionDeclared',
          outcome: 'TimelineOutcomeSuccess',
          visibility: 'TimelineVisibilityPublic',
          gameplay_type: 'special_action',
          actor_user_id: 'p1',
          actor_name: 'Alice',
          action_type: 'Buy',
          summary: 'Alice 执行特殊行动【购买】',
        },
        {
          event_id: 13,
          turn_id: 1,
          chain_id: 'chain_13',
          type: 'TimelineEffectResolved',
          outcome: 'TimelineOutcomeSuccess',
          visibility: 'TimelineVisibilityPublic',
          gameplay_type: 'state_delta',
          deltas: [
            {
              type: 'morale',
              scope: 'team',
              camp: 'Red',
              field: 'morale',
              before: 15,
              after: 14,
              value: -1,
            },
          ],
        },
      ],
    }

    store.push(payload)

    expect(store.payloads).toHaveLength(1)
    expect(store.payloads[0]?.events).toHaveLength(13)
    expect(store.entries.map((entry) => entry.title)).toEqual([
      '回合1：Alice 使用攻击【火焰斩】 -> Bob',
      'skill_reward',
      'Alice 发动「苍炎法典」',
      'Alice 执行特殊行动【购买】',
      '红方士气 -1',
    ])
    expect(store.entries.map((entry) => entry.type)).toEqual(['system', 'resource', 'skill', 'resource', 'damage'])
  })
})
