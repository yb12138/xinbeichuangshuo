import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useBattleReviewStore } from '../battleReview.store'

describe('useBattleReviewStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  afterEach(() => {
    vi.restoreAllMocks()
    vi.useRealTimers()
  })

  it('deduplicates battle feed entries within the short merge window', () => {
    const store = useBattleReviewStore()
    const nowSpy = vi.spyOn(Date, 'now')

    nowSpy.mockReturnValue(1000)
    store.addBattleFeed({
      type: 'turn',
      title: '回合开始：勇者',
      actorId: 'p1',
      actorName: '勇者',
    })

    nowSpy.mockReturnValue(1200)
    store.addBattleFeed({
      type: 'turn',
      title: '回合开始：勇者',
      actorId: 'p1',
      actorName: '勇者',
    })

    nowSpy.mockReturnValue(1400)
    store.addBattleFeed({
      type: 'turn',
      title: '回合开始：勇者',
      actorId: 'p1',
      actorName: '勇者',
    })

    expect(store.battleFeed).toHaveLength(2)
    expect(store.battleFeed[0]?.timestamp).toBe(1000)
    expect(store.battleFeed[1]?.timestamp).toBe(1400)
  })

  it('consumes the freshest matching morale hint and drops stale ones', () => {
    const store = useBattleReviewStore()
    const nowSpy = vi.spyOn(Date, 'now')

    nowSpy.mockReturnValue(1000)
    store.pushMoraleHint({
      source: '旧提示',
      raw: '[Old] 红方士气-1',
      camp: 'Red',
      loss: 1,
    })

    nowSpy.mockReturnValue(5000)
    store.pushMoraleHint({
      source: '蓝方受创',
      raw: '[New] 蓝方士气-2',
      camp: 'Blue',
      loss: 2,
    })

    nowSpy.mockReturnValue(24000)
    const matched = store.consumeMoraleHint('Blue', 2)

    expect(matched?.source).toBe('蓝方受创')
    expect(store.moraleHints).toHaveLength(0)
  })

  it('ranks morale bursts by loss size and then by recency', () => {
    const store = useBattleReviewStore()
    const nowSpy = vi.spyOn(Date, 'now')

    nowSpy.mockReturnValue(1000)
    store.recordMoraleChange('Red', 10, 8, {
      id: 1,
      timestamp: 1000,
      source: '小伤害',
      raw: '[A]',
      camp: 'Red',
      loss: 2,
    })

    nowSpy.mockReturnValue(2000)
    store.recordMoraleChange('Blue', 12, 8, {
      id: 2,
      timestamp: 2000,
      source: '重击',
      raw: '[B]',
      camp: 'Blue',
      loss: 4,
    })

    nowSpy.mockReturnValue(3000)
    store.recordMoraleChange('Red', 8, 5, {
      id: 3,
      timestamp: 3000,
      source: '追击',
      raw: '[C]',
      camp: 'Red',
      loss: 3,
    })

    nowSpy.mockReturnValue(4000)
    store.recordMoraleChange('Blue', 8, 9, {
      id: 4,
      timestamp: 4000,
      source: '恢复',
      raw: '[D]',
      camp: 'Blue',
    })

    expect(store.moraleBurstRanking.map((item) => `${item.camp}:${item.delta}`)).toEqual([
      'Blue:-4',
      'Red:-3',
      'Red:-2',
    ])
  })
})
