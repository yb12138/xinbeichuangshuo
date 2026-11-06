import { describe, expect, it } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/vue'
import AllocationOverlayRenderer from '../AllocationOverlayRenderer.vue'
import DirectionPromptRenderer from '../DirectionPromptRenderer.vue'
import { DECISION_DOCKED } from '../../dockContext'

const dockedGlobal = { provide: { [DECISION_DOCKED as symbol]: true } }

describe('docked allocation', () => {
  const props = {
    visible: true,
    title: '圣疗：分配治疗',
    rows: [
      { id: 'p1', label: '圣女' },
      { id: 'p2', label: '天使' },
    ],
    values: [1, 0],
    remaining: 2,
    total: 3,
    canSubmit: true,
  }

  it('shows the running total and steps values with −/+', async () => {
    const { emitted } = render(AllocationOverlayRenderer, { props, global: dockedGlobal })
    expect(screen.getByTestId('allocation-summary').textContent).toContain('已分配 1/3')
    expect(screen.getByTestId('allocation-value-0').textContent).toBe('1')
    expect((screen.getByTestId('allocation-dec-1') as HTMLButtonElement).disabled).toBe(true)
    await fireEvent.click(screen.getByTestId('allocation-inc-1'))
    await fireEvent.click(screen.getByTestId('allocation-dec-0'))
    expect(emitted().change).toEqual([
      [1, 1],
      [0, 0],
    ])
    await fireEvent.click(screen.getByTestId('allocation-submit'))
    expect(emitted().submit).toHaveLength(1)
  })

  it('blocks + once everything is allocated', () => {
    render(AllocationOverlayRenderer, { props: { ...props, values: [2, 1], remaining: 0 }, global: dockedGlobal })
    expect((screen.getByTestId('allocation-inc-0') as HTMLButtonElement).disabled).toBe(true)
    expect((screen.getByTestId('allocation-inc-1') as HTMLButtonElement).disabled).toBe(true)
  })
})

describe('docked direction', () => {
  it('renders two ring-arrow buttons and selects on click', async () => {
    const { emitted } = render(DirectionPromptRenderer, {
      props: {
        visible: true,
        title: '【魔弹掌控】请选择魔弹传递方向',
        options: [
          { id: 'normal', label: '默认方向', hint: '顺时针传给下一名对手', icon: 'arrow-right' },
          { id: 'reverse', label: '逆向传递', hint: '逆时针传递', icon: 'arrow-left' },
        ],
      },
      global: dockedGlobal,
    })
    const reverse = screen.getByTestId('branch-option-1')
    expect(reverse.className).toContain('dir-docked__btn--ccw')
    expect(screen.getByText('顺时针传给下一名对手')).toBeTruthy()
    await fireEvent.click(reverse)
    expect(emitted().select).toEqual([['reverse']])
  })
})
