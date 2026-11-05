import { describe, expect, it } from 'vitest'
import {
  ELEMENTALIST_EXCLUSIVE_PROMPT_FLOW_SKILL_IDS,
  nextSkillTargets,
  requiredSkillDiscardCount,
  requiresExclusiveCardSelectionBeforeTarget,
  skillTargetBounds,
  targetToBeReplaced,
} from '../skillFlowRules'
import type { AvailableSkill } from '../../types/game'

function skill(overrides: Partial<AvailableSkill> = {}): AvailableSkill {
  return {
    id: 'x',
    title: 'X',
    description: '',
    min_targets: 1,
    max_targets: 1,
    target_type: 2,
    cost_gem: 0,
    cost_crystal: 0,
    cost_discards: 0,
    ...overrides,
  }
}

describe('skillFlowRules', () => {
  it('routes every elementalist exclusive skill (cost_discards=1) through exclusive card selection', () => {
    for (const id of ELEMENTALIST_EXCLUSIVE_PROMPT_FLOW_SKILL_IDS) {
      expect(requiresExclusiveCardSelectionBeforeTarget(skill({ id, require_exclusive: true, cost_discards: 1, target_type: 5 }))).toBe(true)
    }
  })

  it('keeps other exclusive skills with discard cost on the discard step', () => {
    expect(requiresExclusiveCardSelectionBeforeTarget(skill({ require_exclusive: true, cost_discards: 1 }))).toBe(false)
    expect(requiresExclusiveCardSelectionBeforeTarget(skill({ require_exclusive: true, cost_discards: 0 }))).toBe(true)
    expect(requiresExclusiveCardSelectionBeforeTarget(skill({ require_exclusive: true, target_type: 0 }))).toBe(false)
    expect(requiresExclusiveCardSelectionBeforeTarget(null)).toBe(false)
  })

  it('caps priest discards at hand size', () => {
    expect(requiredSkillDiscardCount(skill({ id: 'priest_divine_domain', cost_discards: 2 }), 1)).toBe(1)
    expect(requiredSkillDiscardCount(skill({ id: 'other', cost_discards: 2 }), 1)).toBe(2)
    expect(requiredSkillDiscardCount(skill({ cost_discards: 0 }), 5)).toBe(0)
  })

  it('derives target bounds with fallbacks', () => {
    expect(skillTargetBounds(skill({ min_targets: 0, max_targets: 0, target_type: 2 }))).toEqual({ min: 1, max: 1 })
    expect(skillTargetBounds(skill({ min_targets: 1, max_targets: 3 }))).toEqual({ min: 1, max: 3 })
  })

  it('toggles already-selected targets off and replaces the oldest when full', () => {
    expect(nextSkillTargets(['a', 'b'], 'a', 3)).toEqual(['b'])
    expect(nextSkillTargets(['a'], 'b', 3)).toEqual(['a', 'b'])
    expect(nextSkillTargets(['a', 'b'], 'c', 2)).toEqual(['b', 'c'])
    expect(nextSkillTargets(['a'], 'b', 1)).toEqual(['b'])
    expect(nextSkillTargets(['a'], 'a', 1)).toEqual([])
  })

  it('reports which target will be replaced next', () => {
    expect(targetToBeReplaced(['a', 'b'], 2)).toBe('a')
    expect(targetToBeReplaced(['a'], 2)).toBe('')
  })
})
