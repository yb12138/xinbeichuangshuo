import { describe, expect, it } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/vue'
import DecisionOverlayRenderer from '../DecisionOverlayRenderer.vue'
import { DECISION_DOCKED } from '../../dockContext'

const baseProps = {
  visible: true,
  title: '【舞动】请选择先执行的动作：',
  mode: 'text' as const,
  options: [
    { id: '0', label: '摸1张牌', buttonLabel: '摸1张牌' },
    { id: '1', label: '弃1张牌', buttonLabel: '弃1张牌' },
  ],
  canCancel: false,
  cancelLabel: '取消',
  cancelOptionId: '',
  handCount: 4,
}

describe('DecisionOverlayRenderer (docked)', () => {
  it('renders option cards with a hand-count preview and only submits on 确认', async () => {
    const { emitted } = render(DecisionOverlayRenderer, {
      props: baseProps,
      global: { provide: { [DECISION_DOCKED as symbol]: true } },
    })
    expect(screen.getByText('手牌 4 → 5')).toBeTruthy()
    expect(screen.getByText('手牌 4 → 3')).toBeTruthy()
    const confirm = screen.getByTestId('decision-confirm-btn') as HTMLButtonElement
    expect(confirm.disabled).toBe(true)

    await fireEvent.click(screen.getByTestId('branch-option-1'))
    expect(emitted().select).toBeUndefined()
    await fireEvent.click(confirm)
    expect(emitted().select).toEqual([['1']])
  })

  it('keeps the immediate-click list outside the dock', async () => {
    const { emitted } = render(DecisionOverlayRenderer, { props: baseProps })
    expect(screen.queryByTestId('decision-confirm-btn')).toBeNull()
    await fireEvent.click(screen.getByTestId('branch-option-0'))
    expect(emitted().select).toEqual([['0']])
  })

  it('yes-no: answers straight from the footer, 发动 is the gold button', async () => {
    const { emitted } = render(DecisionOverlayRenderer, {
      props: {
        ...baseProps,
        mode: 'yes-no' as const,
        options: [
          { id: '0', label: '发动', buttonLabel: '发动' },
          { id: '1', label: '取消', buttonLabel: '取消' },
        ],
      },
      global: { provide: { [DECISION_DOCKED as symbol]: true } },
    })
    const fire = screen.getByTestId('prompt-option-0')
    expect(fire.className).toContain('decision-actions__gold')
    expect(screen.getByTestId('prompt-option-1').className).toContain('decision-actions__ghost')
    await fireEvent.click(fire)
    expect(emitted().select).toEqual([['0']])
  })

  it('activation-cost: shows the effect and fires on 发动', async () => {
    const { emitted } = render(DecisionOverlayRenderer, {
      props: {
        ...baseProps,
        mode: 'activation-cost' as const,
        options: [],
        activationHint: '消耗1点水晶，对目标造成1点法术伤害',
        activationOptionId: '0',
        canCancel: true,
        cancelOptionId: 'cancel',
      },
      global: { provide: { [DECISION_DOCKED as symbol]: true } },
    })
    expect(screen.getByText('消耗1点水晶，对目标造成1点法术伤害')).toBeTruthy()
    await fireEvent.click(screen.getByTestId('prompt-option-confirm'))
    expect(emitted().select).toEqual([['0']])
    await fireEvent.click(screen.getByTestId('prompt-cancel-btn'))
    expect(emitted().cancel).toEqual([['cancel']])
  })

  it('numeric: picking a tile shows its consequence, 确认 submits it', async () => {
    const { emitted } = render(DecisionOverlayRenderer, {
      props: {
        ...baseProps,
        mode: 'numeric' as const,
        options: [
          { id: '0', label: '消耗0点水晶', buttonLabel: '0' },
          { id: '2', label: '消耗2点水晶，抵御2点士气下降', buttonLabel: '2' },
        ],
      },
      global: { provide: { [DECISION_DOCKED as symbol]: true } },
    })
    await fireEvent.click(screen.getByTestId('numeric-option-2'))
    expect(screen.getByText('消耗2点水晶，抵御2点士气下降')).toBeTruthy()
    expect(emitted().select).toBeUndefined()
    await fireEvent.click(screen.getByTestId('decision-confirm-btn'))
    expect(emitted().select).toEqual([['2']])
  })

  it('shows a disabled branch greyed out with its reason and does not let it be picked', async () => {
    render(DecisionOverlayRenderer, {
      props: {
        ...baseProps,
        options: [
          { id: '0', label: '分支①：对目标造成1点法术伤害', buttonLabel: '' },
          { id: '1', label: '分支②：移除2个茧，然后移除1个蛹', buttonLabel: '', disabled: true, disabledReason: '需要至少1个蛹' },
        ],
      },
      global: { provide: { [DECISION_DOCKED as symbol]: true } },
    })
    const branch2 = screen.getByTestId('branch-option-1') as HTMLButtonElement
    expect(branch2.disabled).toBe(true)
    expect(screen.getByTestId('decision-card-reason-1').textContent).toBe('需要至少1个蛹')
    expect(screen.queryByTestId('decision-card-reason-0')).toBeNull()
    await fireEvent.click(branch2)
    expect((screen.getByTestId('decision-confirm-btn') as HTMLButtonElement).disabled).toBe(true)
  })
})


describe('DecisionOverlayRenderer (docked) layout', () => {
  // .overlay-panel-body 是纵向 flex；选项卡容器若与它共用同一元素，两张卡会被上下堆叠而不是并排。
  it('keeps the option-card row in its own container, not on the column-flex panel body', () => {
    const { container } = render(DecisionOverlayRenderer, {
      props: baseProps,
      global: { provide: { [DECISION_DOCKED as symbol]: true } },
    })
    const tiles = container.querySelector('.dc-tiles')
    expect(tiles).not.toBeNull()
    expect(tiles!.classList.contains('overlay-panel-body')).toBe(false)
    expect(tiles!.parentElement?.classList.contains('overlay-panel-body')).toBe(true)
  })
})
