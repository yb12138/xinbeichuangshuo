import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useInterruptStore, SUBMIT_RESPONSE_TIMEOUT_MS } from '../../stores/interrupt.store'
import { recoverFromRejectedSubmission } from '../submissionRecovery'
import type { AvailableSkill, Prompt } from '../../types/game'

const skill: AvailableSkill = {
  id: 'destruction_storm',
  title: '毁灭风暴',
  description: '',
  min_targets: 2,
  max_targets: 2,
  target_type: 2,
  cost_gem: 1,
  cost_crystal: 0,
  cost_discards: 1,
}

describe('submission lifecycle and recovery', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.useFakeTimers()
  })

  it('restores skill selections when the server rejects the submission', () => {
    const store = useInterruptStore()
    store.setSelectedSkill(skill)
    store.setSkillMode('choosing_target')
    store.setSkillDiscardHandIndexes([2])
    store.setSkillTargetIds(['p2', 'p3'])
    store.beginSubmit()
    expect(store.isSubmitting).toBe(true)

    // 期间有别的代码把状态清了，也能恢复
    store.clearSkillMode()
    recoverFromRejectedSubmission(store, '技能发动失败: 目标不合法')

    expect(store.isSubmitting).toBe(false)
    expect(store.skillMode).toBe('choosing_target')
    expect(store.selectedSkill?.id).toBe('destruction_storm')
    expect(store.skillTargetIds).toEqual(['p2', 'p3'])
    expect(store.skillDiscardHandIndexes).toEqual([2])
  })

  it('clears the counter target but keeps the counter card after a rejected counter', () => {
    const store = useInterruptStore()
    store.setCounterFlowStage('target')
    store.setSelectedHandIndexes([1])
    store.setPromptCounterTarget('p3')
    store.beginSubmit()

    recoverFromRejectedSubmission(store, '不能应战给该目标')

    expect(store.promptCounterTarget).toBe('')
    expect(store.counterFlowStage).toBe('target')
    expect(store.selectedHandIndexes).toEqual([1])
  })

  it('falls back to keyword cleanup when nothing was pending', () => {
    const store = useInterruptStore()
    store.setSelectedSkill(skill)
    store.setSkillMode('choosing_target')
    recoverFromRejectedSubmission(store, '技能发动失败: x')
    expect(store.skillMode).toBe('none')
  })

  it('ends submitting on a new prompt or state update and drops stale selections', () => {
    const store = useInterruptStore()
    store.setSelectedSkill(skill)
    store.setSkillMode('choosing_target')
    store.beginSubmit()
    store.setPrompt(null)
    expect(store.isSubmitting).toBe(false)
    expect(store.skillMode).toBe('none')
    expect(store.hasPendingSubmission()).toBe(false)

    store.beginSubmit()
    store.setPrompt({ type: 'confirm', player_id: 'p1', message: '', options: [], min: 1, max: 1 } as unknown as Prompt)
    expect(store.isSubmitting).toBe(false)

    store.beginSubmit()
    store.syncAfterStateUpdate()
    expect(store.isSubmitting).toBe(false)
  })

  it('keeps local selections when setPrompt(null) arrives without a pending submission', () => {
    const store = useInterruptStore()
    store.setSelectedSkill(skill)
    store.setSkillMode('choosing_target')
    store.setPrompt(null)
    expect(store.skillMode).toBe('choosing_target')
  })

  it('releases the submitting lock after the response timeout', () => {
    const store = useInterruptStore()
    store.beginSubmit()
    vi.advanceTimersByTime(SUBMIT_RESPONSE_TIMEOUT_MS)
    expect(store.isSubmitting).toBe(false)
  })
})
