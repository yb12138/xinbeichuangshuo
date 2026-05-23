import { render, screen } from '@testing-library/vue'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'
import AllocationOverlayRenderer from '../AllocationOverlayRenderer.vue'

function baseProps() {
  return {
    visible: true,
    title: '请分配治疗',
    rows: [
      { id: 'p1', label: '目标一' },
      { id: 'p2', label: '目标二' },
    ],
    values: [1, 0],
    remaining: 2,
    total: 3,
    canSubmit: true,
    submitLabel: '确认分配',
  }
}

describe('AllocationOverlayRenderer', () => {
  it('renders title, remaining summary, rows, and numeric values', () => {
    render(AllocationOverlayRenderer, {
      props: baseProps(),
    })

    expect(screen.getByTestId('allocation-overlay')).toBeInTheDocument()
    expect(screen.getByText('请分配治疗')).toBeInTheDocument()
    expect(screen.getByTestId('allocation-summary')).toHaveTextContent('剩余可分配：2 / 3')
    expect(screen.getByText('目标一')).toBeInTheDocument()
    expect(screen.getByText('目标二')).toBeInTheDocument()
    expect(screen.getByTestId('allocation-option-0-0')).toBeInTheDocument()
    expect(screen.getByTestId('allocation-option-0-3')).toBeInTheDocument()
    expect(screen.getByTestId('allocation-option-0-1')).toHaveClass('overlay-saint-heal-tile--active')
  })

  it('emits change with row index and selected value', async () => {
    const user = userEvent.setup()
    const { emitted } = render(AllocationOverlayRenderer, {
      props: baseProps(),
    })

    await user.click(screen.getByTestId('allocation-option-1-2'))

    expect(emitted('change')).toEqual([[1, 2]])
    expect(emitted('submit')).toBeUndefined()
  })

  it('disables values beyond current value plus remaining', async () => {
    const user = userEvent.setup()
    const { emitted } = render(AllocationOverlayRenderer, {
      props: {
        ...baseProps(),
        values: [2, 0],
        remaining: 0,
      },
    })

    expect(screen.getByTestId('allocation-option-0-2')).not.toBeDisabled()
    expect(screen.getByTestId('allocation-option-0-3')).toBeDisabled()
    expect(screen.getByTestId('allocation-option-1-1')).toBeDisabled()

    await user.click(screen.getByTestId('allocation-option-1-1'))

    expect(emitted('change')).toBeUndefined()
  })

  it('emits submit only when the confirm button is enabled', async () => {
    const user = userEvent.setup()
    const { emitted, rerender } = render(AllocationOverlayRenderer, {
      props: {
        ...baseProps(),
        canSubmit: false,
      },
    })

    await user.click(screen.getByTestId('allocation-submit'))
    expect(emitted('submit')).toBeUndefined()

    await rerender({
      ...baseProps(),
      canSubmit: true,
    })
    await user.click(screen.getByTestId('allocation-submit'))

    expect(emitted('submit')).toEqual([[]])
  })

  it('does not render when visible is false', () => {
    render(AllocationOverlayRenderer, {
      props: {
        ...baseProps(),
        visible: false,
      },
    })

    expect(screen.queryByTestId('allocation-overlay')).not.toBeInTheDocument()
    expect(screen.queryByText('请分配治疗')).not.toBeInTheDocument()
  })
})
