import { render } from '@testing-library/vue'
import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it } from 'vitest'
import StageOverlay from '../StageOverlay.vue'
import { useStageQueueStore } from '../../stores/stageQueue.store'

describe('StageOverlay combat chain', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  // 被动 / 响应技能挂在它反应的那一步下面；只写技能名看不出效果，结果（"信仰 +1"）要一起显示。
  it('shows the outcome of a trigger skill hanging under the hit step (天之弓 → 信仰 +1)', async () => {
    const store = useStageQueueStore()
    store.enqueue(
      { type: 'mark-previous' },
      { type: 'link', linkId: 'e1', fromPlayerId: 'p1', toPlayerId: 'p2', color: 'attack', label: '攻击·地裂斩' },
      {
        type: 'narration-row',
        html: '圣弓 攻击 兽灵武士',
        tag: '攻击',
        tagColor: 'attack',
        chain: { linkId: 'e1', fromPlayerId: 'p1', toPlayerId: 'p2', verb: '攻击', card: { name: '地裂斩', element: 'Earth', type: 'Attack' } },
      },
      {
        type: 'narration-row',
        html: '圣弓 发动 天之弓',
        tag: '技能',
        tagColor: 'skill',
        skill: { skillId: 'hb_heavenly_bow', skillName: '天之弓', actorId: 'p1', targetIds: ['p1'], opening: 'trigger', effects: [], outcome: '信仰 +1' },
      },
    )
    for (let t = 0; t < 20; t += 1) store.tick(t * 1000)

    const { findAllByTestId } = render(StageOverlay, { props: { rootEl: null } })
    const triggers = await findAllByTestId('stage-chain-trigger')
    expect(triggers).toHaveLength(1)
    expect(triggers[0]!.textContent?.replace(/\s+/g, ' ').trim()).toBe('✦ 天之弓信仰 +1')
  })

  it('shows a camp morale drop as its own line under the chain', async () => {
    const store = useStageQueueStore()
    store.enqueue(
      { type: 'mark-previous' },
      { type: 'link', linkId: 'e1', fromPlayerId: 'p1', toPlayerId: 'p2', color: 'attack', label: '攻击·火焰斩' },
      {
        type: 'narration-row',
        html: '圣弓 攻击 兽灵武士',
        tag: '攻击',
        tagColor: 'attack',
        chain: { linkId: 'e1', fromPlayerId: 'p1', toPlayerId: 'p2', verb: '攻击', card: { name: '火焰斩', element: 'Fire', type: 'Attack' } },
      },
      { type: 'narration-row', html: '兽灵武士 承受', tag: '承受', tagColor: 'attack' },
      { type: 'narration-row', html: '红方士气 15 → 14', tag: '士气', tagColor: 'attack' },
    )
    for (let t = 0; t < 20; t += 1) store.tick(t * 1000)

    const { findByTestId } = render(StageOverlay, { props: { rootEl: null } })
    expect((await findByTestId('stage-chain-morale')).textContent?.trim()).toBe('红方士气 15 → 14')
  })

  // 回合开始的场上效果结算（虚弱）也在战斗链里画成卡，不再退回"当前行动"文字框。
  it('draws a field effect resolution (虚弱) as a chain node, followed by its morale line', async () => {
    const store = useStageQueueStore()
    store.enqueue(
      { type: 'mark-previous' },
      {
        type: 'narration-row',
        html: '吟游诗人 虚弱 → 狂战士',
        tag: '效果',
        tagColor: 'skill',
        effect: { actorId: 'p1', targetIds: ['p2'], label: '虚弱', card: { name: '虚弱', element: 'Fire', type: 'Magic' }, kind: 'effect' },
      },
      { type: 'narration-row', html: '红方士气 15 → 14（狂战士爆牌）', tag: '士气', tagColor: 'attack' },
    )
    for (let t = 0; t < 20; t += 1) store.tick(t * 1000)

    const { findAllByTestId, findByTestId, queryByText } = render(StageOverlay, { props: { rootEl: null } })
    const nodes = await findAllByTestId('stage-chain-node')
    expect(nodes).toHaveLength(1)
    expect(nodes[0]!.textContent).toContain('虚弱 生效')
    expect((await findByTestId('stage-chain-morale')).textContent?.trim()).toBe('红方士气 15 → 14（狂战士爆牌）')
    // 兜底文字框已移除
    expect(queryByText('当前行动')).toBeNull()
  })

  it('shows a move with only a morale result as a chain, not a text box', async () => {
    const store = useStageQueueStore()
    store.enqueue(
      { type: 'mark-previous' },
      { type: 'narration-row', html: '蓝方士气 15 → 13', tag: '士气', tagColor: 'attack' },
    )
    for (let t = 0; t < 20; t += 1) store.tick(t * 1000)

    const { findByTestId, queryByText, queryAllByTestId } = render(StageOverlay, { props: { rootEl: null } })
    expect((await findByTestId('stage-chain-morale')).textContent?.trim()).toBe('蓝方士气 15 → 13')
    expect(queryAllByTestId('stage-chain-node')).toHaveLength(0)
    expect(queryByText('当前行动')).toBeNull()
  })

  it('marks the card name and 命格 on chain cards', async () => {
    const store = useStageQueueStore()
    store.enqueue(
      { type: 'mark-previous' },
      { type: 'link', linkId: 'e1', fromPlayerId: 'p1', toPlayerId: 'p2', color: 'attack', label: '攻击·地裂斩' },
      {
        type: 'narration-row',
        html: '圣弓 攻击 兽灵武士',
        tag: '攻击',
        tagColor: 'attack',
        chain: { linkId: 'e1', fromPlayerId: 'p1', toPlayerId: 'p2', verb: '攻击', card: { name: '地裂斩', element: 'Earth', type: 'Attack', faction: '技' } },
      },
    )
    for (let t = 0; t < 20; t += 1) store.tick(t * 1000)

    const { findByTestId } = render(StageOverlay, { props: { rootEl: null } })
    expect((await findByTestId('stage-chain-card-name')).textContent?.trim()).toBe('地裂斩')
    expect((await findByTestId('stage-chain-fate')).textContent?.trim()).toBe('技')
  })

  it('shows the cards a skill reveals to everyone (圣屑飓暴) under the skill node', async () => {
    const store = useStageQueueStore()
    store.enqueue(
      { type: 'mark-previous' },
      {
        type: 'narration-row',
        html: '圣弓 发动【圣屑飓暴】',
        tag: '技能',
        tagColor: 'skill',
        skill: {
          skillId: 'hb_holy_shard_storm',
          skillName: '圣屑飓暴',
          actorId: 'p1',
          targetIds: ['p2'],
          opening: 'active',
          effects: [],
          revealed: [
            { name: '火焰斩', element: 'Fire', type: 'Attack', faction: '血' },
            { name: '火焰斩', element: 'Fire', type: 'Attack', faction: '幻' },
          ],
        },
      },
    )
    for (let t = 0; t < 20; t += 1) store.tick(t * 1000)

    const { findByTestId } = render(StageOverlay, { props: { rootEl: null } })
    const revealed = await findByTestId('stage-chain-revealed')
    const minis = revealed.querySelectorAll('.stage-chain__mini')
    expect(minis).toHaveLength(2)
    expect(revealed.textContent).toContain('血')
    expect(revealed.textContent).toContain('幻')
    expect(revealed.textContent).toContain('火焰斩')
  })
})
