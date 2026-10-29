import { beforeEach, describe, expect, it } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/vue'
import { createPinia, setActivePinia } from 'pinia'
import DecisionDock from '../DecisionDock.vue'
import { useInterruptStore } from '../../stores/interrupt.store'
import { useSessionStore } from '../../stores/session.store'
import { useSnapshotStore } from '../../stores/snapshot.store'
import { useUiStore } from '../../stores/ui.store'

type PromptInput = Parameters<ReturnType<typeof useInterruptStore>['setPrompt']>[0]

function prompt(overrides: Record<string, unknown> = {}): PromptInput {
  return {
    type: 'choice',
    player_id: 'p1',
    message: '【舞动】请选择先执行的动作：',
    options: [
      { id: '0', label: '摸1张牌' },
      { id: '1', label: '弃1张牌' },
    ],
    presentation: { kind: 'branch_select', layout: 'overlay', numeric_base: 0 },
    min: 1,
    max: 1,
    ...overrides,
  } as unknown as PromptInput
}

describe('DecisionDock', () => {
  let pinia: ReturnType<typeof createPinia>

  beforeEach(() => {
    pinia = createPinia()
    setActivePinia(pinia)
    useSessionStore().myPlayerId = 'p1'
    const snapshot = useSnapshotStore()
    snapshot.currentPlayer = 'p1'
  })

  const mount = () => render(DecisionDock, { global: { plugins: [pinia], stubs: { PromptDialog: true } } })

  it('shows my non-hub prompt in a full panel titled by its skill', () => {
    useInterruptStore().setPrompt(prompt())
    mount()
    expect(screen.getByTestId('decision-dock')).toBeTruthy()
    expect(screen.getByTestId('decision-dock-title').textContent).toBe('舞动')
    expect(screen.getByTestId('decision-dock').className).not.toContain('decision-dock--strip')
    expect(screen.queryByTestId('decision-dock-band')).toBeNull()
  })

  it('stays out of the way for the action hub and for other players’ prompts', () => {
    const interrupt = useInterruptStore()
    interrupt.setPrompt(prompt({ presentation: { kind: 'action_hub', numeric_base: 0 } }))
    const { unmount } = mount()
    expect(screen.queryByTestId('decision-dock')).toBeNull()
    unmount()
    interrupt.setPrompt(prompt({ player_id: 'p2' }))
    mount()
    expect(screen.queryByTestId('decision-dock')).toBeNull()
  })

  it('narrows to a strip for prompts answered on the board', () => {
    useInterruptStore().setPrompt(prompt({ presentation: { kind: 'target_picker', numeric_base: 0 } }))
    mount()
    expect(screen.getByTestId('decision-dock').className).toContain('decision-dock--strip')
  })

  it('collapses to a pill and expands back', async () => {
    useInterruptStore().setPrompt(prompt())
    mount()
    await fireEvent.click(screen.getByTestId('decision-dock-collapse'))
    expect(useUiStore().decisionDockCollapsed).toBe(true)
    const pill = await screen.findByTestId('decision-dock-pill')
    expect(pill.textContent).toContain('待决策：舞动')
    await fireEvent.click(pill)
    expect(useUiStore().decisionDockCollapsed).toBe(false)
  })

  it('forces the panel open with an interrupt band on someone else’s turn', () => {
    useSnapshotStore().currentPlayer = 'p2'
    useUiStore().setDecisionDockCollapsed(true)
    useInterruptStore().setPrompt(prompt())
    mount()
    expect(screen.getByTestId('decision-dock-band').textContent).toContain('打断')
    expect(screen.queryByTestId('decision-dock-collapse')).toBeNull()
    expect(screen.queryByTestId('decision-dock-pill')).toBeNull()
  })

  it('shows step progress dots for multi-step flows', () => {
    useInterruptStore().setPrompt(
      prompt({ presentation: { kind: 'branch_select', layout: 'overlay', numeric_base: 0, step_index: 2, total_steps: 3 } }),
    )
    mount()
    const steps = screen.getByTestId('decision-dock-steps')
    expect(steps.textContent).toContain('第 2/3 步')
    expect(steps.querySelectorAll('i')).toHaveLength(3)
    expect(steps.querySelectorAll('i.is-current')).toHaveLength(1)
  })

  it('hides progress for single-step prompts', () => {
    useInterruptStore().setPrompt(prompt())
    mount()
    expect(screen.queryByTestId('decision-dock-steps')).toBeNull()
  })

  it('shows the skill description so the player knows what the branches mean', async () => {
    const snapshot = useSnapshotStore()
    snapshot.characters = {
      holy_bow: {
        id: 'holy_bow',
        name: '圣弓',
        skills: [{ id: 'hb_auto_fill', title: '自动填充', description: '回合结束时……①消耗1水晶，+1信仰或+1治疗' }],
      },
    } as never
    useInterruptStore().setPrompt(prompt({ message: '【自动填充】请选择要发动的分支：' }))
    mount()
    expect(screen.getByTestId('decision-dock-desc').textContent).toContain('①消耗1水晶')
    await fireEvent.click(screen.getByTestId('decision-dock-desc-toggle'))
    expect(screen.queryByTestId('decision-dock-desc')).toBeNull()
  })

  it('collapses the description by default on board-answered strips', () => {
    const snapshot = useSnapshotStore()
    snapshot.characters = {
      x: { id: 'x', name: 'X', skills: [{ id: 'sk', title: '测试', description: '说明文本' }] },
    } as never
    useInterruptStore().setPrompt(prompt({ skill_id: 'sk', presentation: { kind: 'target_picker', numeric_base: 0 } }))
    mount()
    expect(screen.getByTestId('decision-dock-desc-toggle')).toBeTruthy()
    expect(screen.queryByTestId('decision-dock-desc')).toBeNull()
  })
})
