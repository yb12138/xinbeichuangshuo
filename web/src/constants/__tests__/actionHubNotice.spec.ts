import { describe, expect, it } from 'vitest'
import { parseActionHubNotice } from '../actionHubNotice'

describe('parseActionHubNotice', () => {
  it('hides the default prompt', () => {
    expect(parseActionHubNotice('请选择行动类型')).toMatchObject({ kind: 'default', text: '' })
    expect(parseActionHubNotice('')).toMatchObject({ kind: 'default' })
    expect(parseActionHubNotice(undefined)).toMatchObject({ kind: 'default' })
  })

  it('classifies extra actions and strips the trailing default sentence', () => {
    expect(parseActionHubNotice('当前为额外攻击行动，仅可执行攻击。请选择行动类型')).toMatchObject({
      kind: 'extra',
      text: '当前为额外攻击行动，仅可执行攻击',
      extraLimit: '仅可攻击',
    })
    expect(parseActionHubNotice('当前为额外法术行动，仅可执行法术。请选择行动类型').extraLimit).toBe('仅可法术')
    expect(parseActionHubNotice('当前为额外行动，可执行攻击或法术。请选择行动类型').extraLimit).toBe('攻击或法术')
  })

  it('flags an extra action with nothing playable', () => {
    expect(parseActionHubNotice('当前为额外行动阶段，但你没有满足约束的可执行动作。可选择跳过本次额外行动。').kind).toBe('extra-none')
  })

  it('treats role/rule constraints as warnings', () => {
    expect(parseActionHubNotice('你的审判已达上限：本行动阶段必须发动【末日审判】。').kind).toBe('rule')
    expect(parseActionHubNotice('当前规则要求：本行动阶段只能主动攻击。').kind).toBe('rule')
    expect(parseActionHubNotice('你处于【百式幻龙拳】状态：本行动阶段只能主动攻击 圣女；若本行动阶段结束仍处于该形态，则自动转正。').kind).toBe('rule')
    expect(parseActionHubNotice('你受到约束效果影响：本次行动阶段必须且只能主动攻击 圣女，或选择跳过行动并移除此牌。').kind).toBe('rule')
  })

  it('extracts the 充盈 bonus from any variant', () => {
    expect(parseActionHubNotice('请选择行动类型（【充盈】下一次主动攻击伤害额外+2）')).toMatchObject({ kind: 'default', fullnessBonus: 2 })
    expect(parseActionHubNotice('当前为额外攻击行动，仅可执行攻击。请选择行动类型（【充盈】下一次主动攻击伤害额外+1）')).toMatchObject({
      kind: 'extra',
      extraLimit: '仅可攻击',
      fullnessBonus: 1,
    })
  })

  it('passes unknown text through', () => {
    expect(parseActionHubNotice('某个新的提示')).toMatchObject({ kind: 'other', text: '某个新的提示' })
  })
})
