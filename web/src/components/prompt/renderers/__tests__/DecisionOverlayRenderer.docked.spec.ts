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
})
