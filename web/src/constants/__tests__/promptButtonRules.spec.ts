import { describe, expect, it } from 'vitest'
import {
  isDeclineLabel,
  promptImageButtonKindByOption,
  responseOptionKind,
} from '../promptButtonRules'

describe('promptButtonRules', () => {
  it('detects response option kinds and ignores missile direction choices', () => {
    expect(responseOptionKind({ id: 'take', label: '承受伤害' })).toBe('take')
    expect(responseOptionKind({ id: 'counter', label: '打出【魔弹】传递' })).toBe('counter')
    expect(responseOptionKind({ id: 'defend', label: '使用【圣光】抵挡' })).toBe('defend')
    expect(responseOptionKind({ id: 'normal', label: '顺时针' })).toBeNull()
    expect(responseOptionKind({ id: 'reverse', label: '逆时针' })).toBeNull()
  })

  it('does not treat weakness skip-action label as a decline/cancel label', () => {
    expect(isDeclineLabel('跳过行动阶段 (移除虚弱)')).toBe(false)
    expect(isDeclineLabel('跳过')).toBe(true)
    expect(isDeclineLabel('跳过本次')).toBe(true)
  })

  it('maps common prompt buttons to stable image kinds', () => {
    expect(promptImageButtonKindByOption({ id: 'take', label: '承受伤害' })).toBe('take')
    expect(promptImageButtonKindByOption({ id: 'cancel', label: '取消' })).toBe('cancel')
    expect(promptImageButtonKindByOption({ id: '0', label: '0: 火球术' })).toBe('card')
    expect(promptImageButtonKindByOption({ id: 'confirm', label: '确认发动' })).toBe('confirm')
  })
})

