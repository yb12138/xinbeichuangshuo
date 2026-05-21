import { describe, expect, it } from 'vitest'
import {
  skillCanUseDiscardCard,
  skillCostTextOverride,
  skillDiscardGuideText,
  skillSpecificDisabledReason,
} from '../skillDisplayRules'

describe('skillDisplayRules', () => {
  it('shares the same fusion discard rule for active and chain fusion', () => {
    expect(skillCanUseDiscardCard('magic_bullet_fusion', { element: 'Fire' })).toBe(true)
    expect(skillCanUseDiscardCard('magic_bullet_fusion_chain', { element: 'Earth' })).toBe(true)
    expect(skillCanUseDiscardCard('magic_bullet_fusion', { element: 'Water' })).toBe(false)
    expect(skillDiscardGuideText('magic_bullet_fusion')).toBe('需要火系或地系牌')
    expect(skillDiscardGuideText('magic_bullet_fusion_chain')).toBe('需要火系或地系牌')
    expect(skillSpecificDisabledReason('magic_bullet_fusion')).toContain('魔弹融合')
    expect(skillSpecificDisabledReason('magic_bullet_fusion_chain')).toContain('魔弹融合')
  })

  it('keeps skill-specific button copy stable for special cases', () => {
    expect(skillCostTextOverride('priest_water_power')).toBe('弃1水牌+交1手牌(若有)')
    expect(skillCostTextOverride('ss_soul_recall')).toBe('弃法术牌')
    expect(skillDiscardGuideText('priest_water_power')).toContain('第一张需水系')
  })
})
