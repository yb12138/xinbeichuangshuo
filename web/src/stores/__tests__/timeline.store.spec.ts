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
      seq_end: 5,
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
      ],
    }

    store.push(payload)

    expect(store.payloads).toHaveLength(1)
    expect(store.payloads[0]?.events).toHaveLength(4)
    expect(store.entries.map((entry) => entry.title)).toEqual([
      '回合1：Alice 使用攻击【火焰斩】 -> Bob',
      'Bob 受到2点伤害',
    ])
  })
})
