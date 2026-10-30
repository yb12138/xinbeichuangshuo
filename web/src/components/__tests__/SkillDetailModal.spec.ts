import { render, screen } from '@testing-library/vue'
import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it } from 'vitest'
import SkillDetailModal from '../SkillDetailModal.vue'
import { useSessionStore } from '../../stores/session.store'
import { useSnapshotStore } from '../../stores/snapshot.store'
import type { CharacterView, SkillView } from '../../types/game'

function skill(overrides: Partial<SkillView>): SkillView {
  return {
    id: 'sk',
    title: '技能',
    description: '描述',
    type: 2,
    min_targets: 0,
    max_targets: 0,
    target_type: 0,
    cost_gem: 0,
    cost_crystal: 0,
    cost_discards: 0,
    ...overrides,
  }
}

const holyBow: CharacterView = {
  id: 'holy_bow',
  name: '圣弓',
  title: '圣',
  faction: '圣',
  skills: [
    skill({ id: 'hb_heavenly_bow', title: '天之弓', type: 0, description: '治疗上限+1。' }),
    skill({ id: 'hb_holy_shard_storm', title: '圣屑飓暴', description: '弃2张同系攻击牌并展示。' }),
    skill({ id: 'hb_light_burst', title: '圣光爆裂', description: '仅圣煌形态可发动：①摸1；②移除X治疗。' }),
    skill({ id: 'hb_auto_fill', title: '自动填充', type: 3, tags: ['Ultimate'], description: '回合结束时：①消耗1水晶；②消耗1宝石。' }),
    skill({ id: 'ex', title: '独有技', require_exclusive: true, cost_gem: 1 }),
  ],
}

describe('SkillDetailModal', () => {
  let pinia: ReturnType<typeof createPinia>

  beforeEach(() => {
    pinia = createPinia()
    setActivePinia(pinia)
  })

  const mount = (character: CharacterView = holyBow) =>
    render(SkillDetailModal, { props: { visible: true, character }, global: { plugins: [pinia] } })

  it('shows name, 命格, type and tag pills', () => {
    mount()
    expect(screen.getByText('圣弓')).toBeInTheDocument()
    expect(screen.getByText('命格：圣')).toBeInTheDocument()
    expect(screen.getByText('被动')).toBeInTheDocument()
    expect(screen.getByText('响应')).toBeInTheDocument()
    expect(screen.getByText('大招')).toBeInTheDocument()
    expect(screen.getByText('仅圣煌形态')).toBeInTheDocument()
    expect(screen.getByText('独有')).toBeInTheDocument()
  })

  it('splits ①② branches into separate lines', () => {
    mount()
    const row = screen.getByTestId('skill-detail-hb_auto_fill')
    expect(row.querySelectorAll('.skill-row__line--branch')).toHaveLength(2)
  })

  it('hides resources outside a game', () => {
    mount()
    expect(screen.queryByTestId('skill-detail-resources')).toBeNull()
  })

  it('shows my resources and the server disabled reason for my own character', () => {
    useSessionStore().myPlayerId = 'p1'
    const snapshot = useSnapshotStore()
    snapshot.players = {
      p1: { id: 'p1', role: 'holy_bow', camp: 'Blue', heal: 1, max_heal: 3, gem: 3, crystal: 2, tokens: {} },
    } as never
    snapshot.availableSkills = [
      { id: 'hb_holy_shard_storm', disabled: true, disabled_reason: '手牌中没有2张同系攻击牌。' },
    ] as never
    mount()
    expect(screen.getByTestId('skill-detail-resources').textContent).toContain('1/3')
    expect(screen.getByTestId('skill-detail-reason-hb_holy_shard_storm').textContent).toContain('手牌中没有2张同系攻击牌')
  })

  it("doesn't judge other players' skills", () => {
    useSessionStore().myPlayerId = 'p2'
    const snapshot = useSnapshotStore()
    snapshot.players = { p1: { id: 'p1', role: 'holy_bow', camp: 'Blue', heal: 0, max_heal: 3, gem: 0, crystal: 0 } } as never
    snapshot.availableSkills = [{ id: 'hb_holy_shard_storm', disabled: true, disabled_reason: 'x' }] as never
    mount()
    expect(screen.queryByTestId('skill-detail-reason-hb_holy_shard_storm')).toBeNull()
  })
})
