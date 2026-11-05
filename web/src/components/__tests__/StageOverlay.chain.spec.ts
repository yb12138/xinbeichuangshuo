import { fireEvent, render } from '@testing-library/vue'
import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it } from 'vitest'
import StageOverlay from '../StageOverlay.vue'
import { useStageQueueStore } from '../../stores/stageQueue.store'
import { useSnapshotStore } from '../../stores/snapshot.store'

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

  // 战斗链回放：▲ 往前翻已结束的链，最多 5 条；▼ 回到较新的链。
  it('replays up to five earlier combat chains', async () => {
    const store = useStageQueueStore()
    const chain = (n: number) => [
      { type: 'mark-previous' as const },
      { type: 'narration-row' as const, html: `第${n}手`, tag: '记录', tagColor: 'grey' as const },
    ]
    // 7 手：第7手在舞台，第6手是上一手（被压下），回放可翻第6手 ~ 第2手
    for (let n = 1; n <= 7; n += 1) store.enqueue(...chain(n))
    for (let t = 0; t < 40; t += 1) store.tick(t * 1000)
    expect(store.chainArchive).toHaveLength(5)

    const { findByTestId, queryByTestId } = render(StageOverlay, { props: { rootEl: null } })
    const tag = await findByTestId('stage-chain-tag')
    const older = await findByTestId('stage-chain-replay-older')
    const text = () => document.querySelector('.stage-chain')?.textContent || ''
    expect(text()).toContain('第7手')
    expect(queryByTestId('stage-chain-replay-newer')).toBeNull()

    await fireEvent.click(older)
    expect(tag.textContent?.trim()).toBe('回放 · 前 1 条')
    expect(text()).toContain('第6手')

    for (let i = 0; i < 6; i += 1) await fireEvent.click(older)
    expect(tag.textContent?.trim()).toBe('回放 · 前 5 条')
    expect(text()).toContain('第2手')
    expect((older as HTMLButtonElement).disabled).toBe(true)

    await fireEvent.click(await findByTestId('stage-chain-replay-newer'))
    expect(text()).toContain('第3手')
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

  // 截图里的一回合：剑帝雷光斩 → 封印师应战转给精灵射手（承受），剑气斩追加地裂斩 → 贤者。
  // 两次攻击各占一行；被动挂在剑帝名下；承受与伤害写在各自那一行的行尾。
  it('splits a turn into one row per attack with the verdict at the row end', async () => {
    const snapshot = useSnapshotStore()
    snapshot.players = {
      p1: { id: 'p1', name: 'A3', camp: 'Blue', role: 'sword_emperor' },
      p2: { id: 'p2', name: 'B1', camp: 'Red', role: 'sealer' },
      p3: { id: 'p3', name: 'A2', camp: 'Blue', role: 'elf_archer' },
      p4: { id: 'p4', name: 'B3', camp: 'Red', role: 'sage' },
    } as never
    snapshot.characters = {
      sword_emperor: { name: '剑帝' },
      sealer: { name: '封印师' },
      elf_archer: { name: '精灵射手' },
      sage: { name: '贤者' },
    } as never
    const thunder = { name: '雷光斩', element: 'Thunder', type: 'Attack', damage: 2 } as const
    const store = useStageQueueStore()
    store.enqueue(
      { type: 'mark-previous' },
      { type: 'link', linkId: 'e1', fromPlayerId: 'p1', toPlayerId: 'p2', color: 'attack', label: '攻击' },
      { type: 'narration-row', html: '', tag: '攻击', tagColor: 'attack', chain: { linkId: 'e1', fromPlayerId: 'p1', toPlayerId: 'p2', verb: '攻击', card: thunder } },
      { type: 'link', linkId: 'e2', fromPlayerId: 'p2', toPlayerId: 'p3', color: 'attack', label: '应战' },
      { type: 'narration-row', html: '', tag: '应战', tagColor: 'attack', chain: { linkId: 'e2', fromPlayerId: 'p2', toPlayerId: 'p3', verb: '应战', card: thunder } },
      {
        type: 'narration-row',
        html: '',
        tag: '技能',
        tagColor: 'skill',
        skill: { skillId: 'se_guard', skillName: '剑魂守护', actorId: 'p1', targetIds: ['p1'], opening: 'trigger', effects: [], anchorLinkId: 'e1' },
      },
      { type: 'narration-row', html: '', tag: '承受', tagColor: 'grey', take: { playerId: 'p3', sourceId: 'p2', linkId: 'e3', damage: 0, heal: false } },
      { type: 'take-result', linkId: 'e3', damage: 2, heal: false },
      { type: 'narration-row', html: '红方战绩区 蓝水晶 +1', tag: '资源', tagColor: 'heal' },
      { type: 'link', linkId: 'e4', fromPlayerId: 'p1', toPlayerId: 'p4', color: 'attack', label: '攻击' },
      {
        type: 'narration-row',
        html: '',
        tag: '攻击',
        tagColor: 'attack',
        chain: { linkId: 'e4', fromPlayerId: 'p1', toPlayerId: 'p4', verb: '攻击', card: { name: '地裂斩', element: 'Earth', type: 'Attack', damage: 2 } as const },
      },
      {
        type: 'narration-row',
        html: '',
        tag: '技能',
        tagColor: 'skill',
        skill: { skillId: 'se_demon', skillName: '恶魔之魂', actorId: 'p1', targetIds: ['p1'], opening: 'trigger', effects: [], outcome: '伤害 +1', anchorLinkId: 'e4' },
      },
      { type: 'narration-row', html: '', tag: '承受', tagColor: 'grey', take: { playerId: 'p4', sourceId: 'p1', linkId: 'e5', damage: 3, heal: false } },
    )
    for (let t = 0; t < 30; t += 1) store.tick(t * 1000)

    const { findAllByTestId, queryAllByText } = render(StageOverlay, { props: { rootEl: null } })
    const rows = await findAllByTestId('stage-chain-row')
    expect(rows).toHaveLength(2)
    const text = (el: Element) => el.textContent?.replace(/\s+/g, ' ').trim() || ''

    const titles = (await findAllByTestId('stage-chain-row-title')).map(text)
    expect(titles).toEqual(['主攻击', '追加攻击'])
    // 第一行：雷光斩 + 应战雷光斩两张牌，同一步不重复
    expect(rows[0]!.querySelectorAll('[data-testid="stage-chain-node"]')).toHaveLength(2)
    expect(text(rows[0]!.querySelector('[data-testid="stage-chain-verdict"]')!)).toBe('命中2 伤害')
    expect(text(rows[0]!)).toContain('剑帝：')
    expect(text(rows[0]!)).toContain('剑魂守护')
    expect(text(rows[0]!)).toContain('红方战绩区 蓝水晶 +1')
    // 第二行：伤害算式写出恶魔之魂的 +1
    expect(text(rows[1]!.querySelector('[data-testid="stage-chain-verdict"]')!)).toBe('命中3 伤害')
    expect(text(rows[1]!.querySelector('[data-testid="stage-chain-formula"]')!)).toBe('基础 2 · 恶魔之魂 +1 → 3')
    // 承受不再单列成文字
    expect(queryAllByText(/承受/)).toHaveLength(0)
  })
})
