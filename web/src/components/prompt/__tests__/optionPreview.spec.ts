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

  it('marks a "不…" option as skip', () => {
    expect(optionPreview('不摸牌', undefined, 4)).toMatchObject({ title: '不摸牌', glyph: 'skip' })
  })

  it('prefers the backend hint over inferred text', () => {
    expect(optionPreview('摸1张牌', '先摸再放茧', 4).desc).toBe('先摸再放茧')
  })
})

describe('sharedConsequence', () => {
  it('extracts the "然后…" clause from a skill description', () => {
    expect(sharedConsequence('选择：摸1张牌（强制）或弃1张牌（强制）；然后将牌库顶1张牌面朝下放置为茧。')).toBe('然后将牌库顶1张牌面朝下放置为茧')
  })

  it('does not present the next-step branch choice as a fixed consequence (希望赋格曲)', () => {
    const fugue = '［水晶］你可以先摸1张牌；然后选择：将永恒乐章放置于目标队友面前；或将永恒乐章转移给我方另一名目标角色，你弃1张牌并获得+1治疗或+1灵感。'
    expect(sharedConsequence(fugue)).toBe('')
  })

  it('keeps every "；或…" branch of a shared clause', () => {
    expect(sharedConsequence('你摸1张牌；然后弃1张牌；或+1治疗。')).toBe('然后弃1张牌；或+1治疗')
  })

  it('returns empty when there is no shared clause', () => {
    expect(sharedConsequence('指定最多3名角色各+1［治疗］。')).toBe('')
    expect(sharedConsequence(undefined)).toBe('')
  })
})
