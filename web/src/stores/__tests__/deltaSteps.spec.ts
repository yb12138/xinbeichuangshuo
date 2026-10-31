import { describe, expect, it } from 'vitest'
import { deltaSteps } from '../stageQueue.store'
import type { TimelineDelta } from '../../types/generated'

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

describe('deltaSteps animation templates', () => {
  it('turns covers placed from the deck into one deck-to-seat flight before the cover marks (蛹化)', () => {
    const steps = deltaSteps(
      [{ type: 'deck_count', field: 'deck_count', value: -4 }, cocoon('p1'), cocoon('p1'), cocoon('p1'), cocoon('p1')],
      'p1',
    )
    const flights = steps.filter((s) => s.type === 'deck-to-seat')
    expect(flights).toEqual([{ type: 'deck-to-seat', playerId: 'p1', count: 4, coverKey: 'ButterflyCocoon' }])
    expect(steps.findIndex((s) => s.type === 'deck-to-seat')).toBeLessThan(steps.findIndex((s) => s.type === 'status-change'))
  })

  it('does not fly cards from the deck when covers come from somewhere else (hand)', () => {
    const steps = deltaSteps([cocoon('p1')], 'p1')
    expect(steps.some((s) => s.type === 'deck-to-seat')).toBe(false)
  })

  it('emits form-change for entering and leaving a form, coloured by the skill element', () => {
    const enter = deltaSteps([{ type: 'form', target_user_id: 'p1', field: 'form', before_text: '', after_text: 'blaze_witch_flame_form' }], 'p1', 'Fire')
    expect(enter).toEqual([{ type: 'form-change', playerId: 'p1', form: 'blaze_witch_flame_form', label: '烈焰形态', element: 'Fire' }])
    const leave = deltaSteps([{ type: 'form', target_user_id: 'p1', field: 'form', before_text: 'blaze_witch_flame_form', after_text: '' }], 'p1')
    expect(leave).toEqual([{ type: 'form-change', playerId: 'p1', form: '', label: '烈焰形态' }])
  })
})

describe('deltaSteps stone flights', () => {
  it('flies gems from the acting seat to the camp bar when the camp gains them', () => {
    const steps = deltaSteps([{ type: 'team_gem', camp: 'Red', field: 'gem', value: 2 }], 'p3')
    expect(steps.filter((s) => s.type === 'stone-flight')).toEqual([{ type: 'stone-flight', playerId: 'p3', camp: 'Red', stone: 'gem', count: 2, toCamp: true }])
    // 同时写进战斗链
    expect(steps.filter((s) => s.type === 'narration-row')).toHaveLength(1)
  })

  it('flies crystals from the camp bar to the player who extracts them', () => {
    const steps = deltaSteps(
      [
        { type: 'team_crystal', camp: 'Blue', field: 'crystal', value: -1 },
        { type: 'player_crystal', camp: 'Blue', target_user_id: 'p1', field: 'crystal', value: 1 },
      ],
      'p1',
    )
    expect(steps.filter((s) => s.type === 'stone-flight')).toEqual([{ type: 'stone-flight', playerId: 'p1', camp: 'Blue', stone: 'crystal', count: 1, toCamp: false }])
  })

  it('does not fly when the camp only spends stones', () => {
    expect(deltaSteps([{ type: 'team_gem', camp: 'Red', field: 'gem', value: -1 }], 'p3').filter((s) => s.type === 'stone-flight')).toEqual([])
  })

  // 服务端阵营差值的 scope 是 "team"（见 timeline_state_delta.go addCampDelta），这里按真实形状构造。
  it('narrates a camp morale drop so the combat chain can show it', () => {
    const steps = deltaSteps([{ type: 'morale', scope: 'team', camp: 'Red', field: 'morale', before: 15, after: 14, value: -1 }])
    expect(steps).toEqual([{ type: 'narration-row', html: '红方士气 15 → 14', tag: '士气', tagColor: 'attack' }])
  })

  it('ignores morale deltas without a camp or without a change', () => {
    expect(deltaSteps([{ type: 'morale', scope: 'team', field: 'morale', before: 15, after: 14 }])).toEqual([])
    expect(deltaSteps([{ type: 'morale', scope: 'team', camp: 'Blue', field: 'morale', before: 9, after: 9 }])).toEqual([])
  })

  it('summarises camp battle-record changes in one row per camp (合成：星石 -3、星杯 +1)', () => {
    // 实际抓包里星杯差值排在最前；输出仍按 红宝石 → 蓝水晶 → 星杯
    const rows = deltaSteps([
      { type: 'team_cup', scope: 'team', camp: 'Red', field: 'cup', before: 0, after: 1, value: 1 },
      { type: 'team_gem', scope: 'team', camp: 'Red', field: 'gem', before: 2, after: 0, value: -2 },
      { type: 'team_crystal', scope: 'team', camp: 'Red', field: 'crystal', before: 1, after: 0, value: -1 },
    ]).filter((s) => s.type === 'narration-row')
    expect(rows).toEqual([{ type: 'narration-row', html: '红方战绩区 红宝石 -2 · 蓝水晶 -1 · 星杯 +1', tag: '资源', tagColor: 'heal' }])
  })
})
