import { describe, expect, it } from 'vitest'
import { cannotBeActiveAttackTarget, isSkillTargetSelectionValid } from '../skillTargetRules'
import type { PlayerView } from '../../types/game'

function player(overrides: Partial<PlayerView> = {}): PlayerView {
  return {
    id: 'p1', name: 'P1', camp: 'Red', role: 'assassin', hand_count: 0, max_hand: 6,
    exclusive_card_count: 0, field: [], heal: 0, max_heal: 2, gem: 0, crystal: 0,
    is_active: false, buffs: [], ...overrides,
  }
}

describe('skillTargetRules', () => {
  it('prefers the server-sent not_attack_targetable flag', () => {
    expect(cannotBeActiveAttackTarget(player({ not_attack_targetable: true }))).toBe(true)
    // 服务端明确说可以被攻击时，以服务端为准（即使本地看到潜行形态字段）
    expect(cannotBeActiveAttackTarget(player({ form: 'assassin_stealth_form', not_attack_targetable: false }))).toBe(false)
  })

  it('detects stealth form and stealth field effect', () => {
    expect(cannotBeActiveAttackTarget(player())).toBe(false)
    expect(cannotBeActiveAttackTarget(player({ form: 'assassin_stealth_form' }))).toBe(true)
    expect(cannotBeActiveAttackTarget(player({
      field: [{ mode: 'Effect', effect: 'Stealth' } as PlayerView['field'][number]],
    }))).toBe(true)
  })

  it('validates count, distinctness and candidacy', () => {
    const candidates = new Set(['a', 'b', 'c'])
    const bounds = { min: 1, max: 2 }
    expect(isSkillTargetSelectionValid([], bounds, candidates)).toBe(false)
    expect(isSkillTargetSelectionValid(['a'], bounds, candidates)).toBe(true)
    expect(isSkillTargetSelectionValid(['a', 'b'], bounds, candidates)).toBe(true)
    expect(isSkillTargetSelectionValid(['a', 'a'], bounds, candidates)).toBe(false)
    expect(isSkillTargetSelectionValid(['a', 'b', 'c'], bounds, candidates)).toBe(false)
    expect(isSkillTargetSelectionValid(['z'], bounds, candidates)).toBe(false)
  })
})
