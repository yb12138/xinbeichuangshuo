import { describe, expect, it } from 'vitest'
import {
  promptImageButtonKindByOption,
} from '../promptButtonRules'

describe('promptButtonRules', () => {
  it('maps backend button labels to stable image kinds', () => {
    expect(promptImageButtonKindByOption({ buttonLabel: '承受伤害' })).toBe('action')
    expect(promptImageButtonKindByOption({ buttonLabel: '取消' })).toBe('cancel')
    expect(promptImageButtonKindByOption({ buttonLabel: '选择卡牌' })).toBe('card')
    expect(promptImageButtonKindByOption({ buttonLabel: '确认' })).toBe('confirm')
  })
})
