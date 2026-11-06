import { describe, expect, it } from 'vitest'
import { nodeSteps } from '../stageQueue.store'

describe('signature skills (template 6)', () => {
  it('marks a signature skill opening with the long full-width banner', () => {
    const steps = nodeSteps(
      { id: 'n', order: 1, kind: 'skill', actor_user_id: 'p1', skill_id: 'destruction_storm', skill_name: '毁灭风暴' },
      (id) => id,
      () => '',
    )
    const bubble = steps.find((s) => s.type === 'skill-bubble')
    expect(bubble && bubble.type === 'skill-bubble' && bubble.signature).toBe(true)
  })

  it('keeps ordinary active skills on the seat plaque only', () => {
    const steps = nodeSteps(
      { id: 'n', order: 1, kind: 'skill', actor_user_id: 'p1', skill_id: 'bt_chrysalis', skill_name: '蛹化' },
      (id) => id,
      () => '',
    )
    const bubble = steps.find((s) => s.type === 'skill-bubble')
    expect(bubble && bubble.type === 'skill-bubble' && bubble.signature).toBeFalsy()
  })
})
