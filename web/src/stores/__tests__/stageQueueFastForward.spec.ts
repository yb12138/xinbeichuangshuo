import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useStageQueueStore, type AtomicStep } from '../stageQueue.store'

/** 一个完整动作：开场归档上一手 → 连线 → 叙述行 → 技能开场 → 伤害数值。 */
function action(id: string, from: string, to: string): AtomicStep[] {
  return [
    { type: 'mark-previous' },
    { type: 'link', linkId: id, fromPlayerId: from, toPlayerId: to, color: 'attack', label: `攻击·${id}` },
    { type: 'narration-row', html: `${from} 攻击 ${to}（${id}）`, tag: '攻击', tagColor: 'attack' },
    { type: 'skill-bubble', playerId: from, skillName: `技能${id}`, effectText: '', duration: 1500 },
    { type: 'value-chip', playerId: to, text: '-2', color: 'damage' },
  ]
}

describe('stageQueue fastForward（后台标签页切回来）', () => {
  let store: ReturnType<typeof useStageQueueStore>

  beforeEach(() => {
    // 入队时间取 performance.now()；固定为 0，测试里用 fastForward(at) 控制"过了多久"
    vi.spyOn(performance, 'now').mockReturnValue(0)
    setActivePinia(createPinia())
    store = useStageQueueStore()
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('settles every backlog action except the latest one, without flashing their effects', () => {
    store.enqueue(...action('a1', 'p1', 'p2'), ...action('a2', 'p3', 'p4'), ...action('a3', 'p5', 'p6'))

    store.fastForward(1000)

    // 只剩最近一个动作待播
    expect(store.queue[0]?.type).toBe('mark-previous')
    expect(store.queue).toHaveLength(action('a3', 'p5', 'p6').length)
    // 跳过的动作已结算：a2 是当前舞台（下一步 mark-previous 会把它降为上一手），a1 已进历史
    expect(store.stageRows.map((row) => row.html)).toEqual(['p3 攻击 p4（a2）'])
    expect(store.previousAction?.rows.map((row) => row.html)).toEqual(['p1 攻击 p2（a1）'])
    // 瞬时特效不残留
    expect(store.activeSkillBubble).toBeNull()
    expect(store.valueChips).toHaveLength(0)
    expect([...store.activeLinks.values()].every((link) => !link.animating)).toBe(true)
  })

  it('plays the latest action normally right after fast-forwarding', () => {
    store.enqueue(...action('a1', 'p1', 'p2'), ...action('a2', 'p3', 'p4'))
    store.fastForward(1000)

    // 最新动作按正常节奏逐步播放
    store.tick(1000)
    store.tick(1300)
    store.tick(1700)
    store.tick(2000)
    expect(store.activeSkillBubble?.skillName).toBe('技能a2')
    expect(store.previousAction?.rows.map((row) => row.html)).toEqual(['p1 攻击 p2（a1）'])
    expect(store.stageRows.map((row) => row.html)).toEqual(['p3 攻击 p4（a2）'])
  })

  it('keeps a single freshly-arrived action intact', () => {
    const steps = action('a1', 'p1', 'p2')
    store.enqueue(...steps)

    store.fastForward(1000)

    expect(store.queue).toHaveLength(steps.length)
    expect(store.stageRows).toHaveLength(0)
  })

  it('settles a single long action that has been waiting in the background', () => {
    // 一个很长的动作（同一 flow 持续追加），切走期间全部到达：整段已过时，直接结算
    store.enqueue(...action('a1', 'p1', 'p2'))

    store.fastForward(20000)

    expect(store.queue).toHaveLength(0)
    expect(store.isPlaying).toBe(false)
    expect(store.stageRows.map((row) => row.html)).toEqual(['p1 攻击 p2（a1）'])
    expect(store.activeSkillBubble).toBeNull()
    expect(store.valueChips).toHaveLength(0)
  })

  it('keeps steps that arrived just now even when older backlog is settled', () => {
    store.enqueue(...action('a1', 'p1', 'p2'))
    vi.spyOn(performance, 'now').mockReturnValue(19500)
    store.enqueue({ type: 'value-chip', playerId: 'p2', text: '-1', color: 'damage' })

    store.fastForward(20000)

    expect(store.queue.map((step) => step.type)).toEqual(['value-chip'])
    expect(store.stageRows.map((row) => row.html)).toEqual(['p1 攻击 p2（a1）'])
  })
})
