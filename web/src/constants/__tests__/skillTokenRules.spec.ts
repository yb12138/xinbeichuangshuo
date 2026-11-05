import { describe, expect, it } from 'vitest'
import { SKILL_TOKEN_RULES, skillTokenShortfallReason } from '../skillTokenRules'

describe('skillTokenRules fallback thresholds', () => {
  it('战纹碎击 requires 战纹≥1, not 魔纹', () => {
    expect(skillTokenShortfallReason('hom_rune_smash', { hom_war_rune: 1, hom_magic_rune: 0 })).toBe('')
    expect(skillTokenShortfallReason('hom_rune_smash', { hom_war_rune: 0, hom_magic_rune: 3 })).toContain('战纹不足')
  })

  it('魔纹融合 requires 魔纹≥1', () => {
    expect(skillTokenShortfallReason('hom_glyph_fusion', { hom_magic_rune: 1 })).toBe('')
    expect(skillTokenShortfallReason('hom_glyph_fusion', { hom_magic_rune: 0 })).toContain('魔纹不足')
  })

  it('判决天平 does not require 审判', () => {
    expect(SKILL_TOKEN_RULES.arbiter_balance).toBeUndefined()
    expect(skillTokenShortfallReason('arbiter_balance', { judgment: 0 })).toBe('')
  })

  it('treats missing tokens as zero', () => {
    expect(skillTokenShortfallReason('hero_roar', undefined)).toContain('怒气不足（需要 1，当前 0）')
  })
})
