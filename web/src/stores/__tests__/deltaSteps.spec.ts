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
