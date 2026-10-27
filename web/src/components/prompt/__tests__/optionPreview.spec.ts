import { describe, expect, it } from 'vitest'
import { optionPreview, sharedConsequence } from '../optionPreview'

describe('optionPreview', () => {
  it('previews hand changes for draw / discard options', () => {
    expect(optionPreview('摸1张牌', undefined, 4)).toMatchObject({ title: '摸1张牌', desc: '手牌 4 → 5', glyph: 'cards', tone: 'gain' })
    expect(optionPreview('弃1张牌', undefined, 4)).toMatchObject({ desc: '手牌 4 → 3', glyph: 'trash', tone: 'loss' })
  })

  it('splits "标题：说明" and "标题（说明）" labels', () => {
    expect(optionPreview('分支①：对目标造成1点法术伤害', undefined, null)).toMatchObject({ title: '分支①', desc: '对目标造成1点法术伤害' })
    expect(optionPreview('默认方向（右手边，前一位对手）', undefined, null)).toMatchObject({ title: '默认方向', desc: '右手边，前一位对手' })
  })

  it('prefers the backend hint over inferred text', () => {
    expect(optionPreview('摸1张牌', '先摸再放茧', 4).desc).toBe('先摸再放茧')
  })
})

describe('sharedConsequence', () => {
  it('extracts the "然后…" clause from a skill description', () => {
    expect(sharedConsequence('选择：摸1张牌（强制）或弃1张牌（强制）；然后将牌库顶1张牌面朝下放置为茧。')).toBe('然后将牌库顶1张牌面朝下放置为茧')
  })

  it('returns empty when there is no shared clause', () => {
    expect(sharedConsequence('指定最多3名角色各+1［治疗］。')).toBe('')
    expect(sharedConsequence(undefined)).toBe('')
  })
})
