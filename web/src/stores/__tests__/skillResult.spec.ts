import { describe, expect, it } from 'vitest'
import { skillResult, skillResultSummary } from '../skillResult'
import type { ActionFlowNodeDTO, TimelineDelta } from '../../types/generated'

const cocoon = (owner: string): TimelineDelta => ({
  type: 'field_card_added',
  target_user_id: owner,
  field_card: {
    card: { id: 'x', name: '茧' } as never,
    owner_id: owner,
    source_id: owner,
    mode: 'Cover',
    effect: 'ButterflyCocoon',
    field_hook: '',
    locked: false,
    duration: 0,
  },
})

function node(partial: Partial<ActionFlowNodeDTO>): ActionFlowNodeDTO {
  return { id: 'n', order: 1, kind: 'skill', actor_user_id: 'p1', ...partial }
}

describe('skillResult', () => {
  it('splits the owner gem cost from effects and counts covers (蛹化)', () => {
    const result = skillResult(
      node({
        deltas: [
          { type: 'player_gem', target_user_id: 'p1', field: 'gem', value: -1 },
          { type: 'token', target_user_id: 'p1', field: 'token:bt_pupa', value: 1 },
          cocoon('p1'), cocoon('p1'), cocoon('p1'), cocoon('p1'),
          { type: 'deck_count', field: 'deck_count', value: -4 },
        ],
      }),
    )
    expect(result.costs.map((c) => c.text)).toEqual(['宝石 1'])
    expect(skillResultSummary(result)).toBe('蛹 +1 · 茧 ×4')
    expect(result.effects.every((e) => e.targetId === undefined)).toBe(true)
  })

  it('nets repeated changes and drops zero results', () => {
    const result = skillResult(
      node({
        deltas: [
          { type: 'token', target_user_id: 'p1', field: 'token:bt_pupa', value: 1 },
          { type: 'token', target_user_id: 'p1', field: 'token:bt_pupa', value: -1 },
          { type: 'heal', target_user_id: 'p2', field: 'heal', value: 2 },
        ],
      }),
    )
    expect(result.effects).toEqual([{ text: '治疗 +2', targetId: 'p2', tone: 'gain' }])
  })

  it('renders combat effects from the cause (毒粉 / 式神咒束)', () => {
    const names: Record<string, string> = { p4: '精灵射手' }
    const result = skillResult(
      node({
        effects: [
          { kind: 'damage_modified', summary: '伤害 +1', target_ids: ['p2'] },
          { kind: 'attack_taken_over', summary: 'A → B', target_ids: ['p2', 'p4'] },
        ],
      }),
      (id) => names[id] || id,
    )
    expect(skillResultSummary(result)).toBe('伤害 +1 · 代为应战 → 精灵射手')
  })
})

describe('skillResult persistent effects', () => {
  it('describes a persistent effect starting under its skill (凋零)', () => {
    const result = skillResult(
      node({ deltas: [{ type: 'token', target_user_id: 'p1', field: 'indicator:bt_wither_active', value: 1 }] }),
    )
    expect(skillResultSummary(result)).toBe('凋零生效')
  })

  describe('手牌流动', () => {
    const hand = (id: string, before: number, after: number): TimelineDelta => ({
      type: 'hand_count', target_user_id: id, field: 'hand_count', before, after, value: after - before,
    })

    it('天使祝福：两名目标各交 1 张给天使，代价水系牌不算交出（实测数据）', () => {
      const n = node({
        actor_user_id: 'p3',
        target_user_ids: ['p1', 'p5'],
        cards: [{ id: 'w', name: '水涟斩' } as never],
        deltas: [
          { type: 'discard_count', field: 'discard_count', value: 1 },
          hand('p3', 5, 4),
          hand('p3', 4, 5),
          hand('p5', 4, 3),
          hand('p1', 4, 3),
          hand('p3', 5, 6),
        ],
      })
      const result = skillResult(n, (id) => ({ p1: '暗杀者', p5: '封印师' })[id] || id)
      expect(result.transfers).toEqual([
        { from: 'p5', to: 'p3', count: 1 },
        { from: 'p1', to: 'p3', count: 1 },
      ])
      expect(skillResultSummary(result)).toBe('封印师 交来 1 张 · 暗杀者 交来 1 张')
      expect(skillResultSummary(result, { skipTransfers: true })).toBe('')
    })

    it('单目标交 2 张', () => {
      const result = skillResult(node({ actor_user_id: 'p3', deltas: [hand('p1', 4, 2), hand('p3', 4, 6)] }))
      expect(result.transfers).toEqual([{ from: 'p1', to: 'p3', count: 2 }])
    })

    it('持有者给队友牌（水之神力）', () => {
      const result = skillResult(node({ actor_user_id: 'p1', deltas: [hand('p1', 4, 3), hand('p2', 3, 4)] }), (id) => (id === 'p2' ? '圣女' : id))
      expect(result.transfers).toEqual([{ from: 'p1', to: 'p2', count: 1 }])
      expect(skillResultSummary(result)).toBe('交给 圣女 1 张')
    })

    it('让别人弃牌：无对应增加 = 弃入弃牌堆', () => {
      const result = skillResult(node({ actor_user_id: 'p1', deltas: [hand('p4', 4, 3), hand('p5', 2, 1)] }), (id) => id.toUpperCase())
      expect(result.transfers).toEqual([
        { from: 'p4', to: '', count: 1 },
        { from: 'p5', to: '', count: 1 },
      ])
      expect(skillResultSummary(result)).toBe('P4 弃 1 张 · P5 弃 1 张')
    })

    it('从牌库摸来的不配对成别人交的牌', () => {
      const result = skillResult(node({
        actor_user_id: 'p1',
        deltas: [hand('p2', 4, 3), hand('p1', 4, 6), { type: 'deck_count', field: 'deck_count', value: -1 }],
      }))
      // 持有者自己从牌库摸的那张是过程，不报
      expect(result.transfers).toEqual([{ from: 'p2', to: 'p1', count: 1 }])
    })

    it('魔爆冲击：两名目标各弃 1 张（实测：展示牌含目标的弃牌，不能当作施法者代价）', () => {
      const result = skillResult(node({
        actor_user_id: 'p1',
        cards: [{ id: 'c', name: '圣盾' } as never, { id: 'm', name: '魔弹' } as never, { id: 'h', name: '圣光' } as never],
        deltas: [
          hand('p1', 5, 4), { type: 'discard_count', field: 'discard_count', value: 1 },
          hand('p2', 6, 5), { type: 'discard_count', field: 'discard_count', value: 1 },
          hand('p5', 6, 5), { type: 'discard_count', field: 'discard_count', value: 1 },
        ],
      }))
      expect(result.transfers).toEqual([
        { from: 'p2', to: '', count: 1 },
        { from: 'p5', to: '', count: 1 },
      ])
    })

    it('只有持有者自己摸 / 弃牌不报', () => {
      expect(skillResult(node({ deltas: [hand('p1', 4, 6), { type: 'deck_count', field: 'deck_count', value: -2 }] })).transfers).toEqual([])
      expect(skillResult(node({ deltas: [hand('p1', 4, 3)] })).transfers).toEqual([])
    })
  })
})
