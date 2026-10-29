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
})
