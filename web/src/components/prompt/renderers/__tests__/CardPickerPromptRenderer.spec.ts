import { render, screen } from '@testing-library/vue'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'
import CardPickerPromptRenderer from '../CardPickerPromptRenderer.vue'

function baseProps() {
  return {
    visible: true,
    message: '完成选牌后点击发动',
    canConfirm: true,
    showCancel: false,
  }
}

describe('CardPickerPromptRenderer', () => {
  it('renders message text', () => {
    render(CardPickerPromptRenderer, {
      props: baseProps(),
    })

    expect(screen.getByTestId('card-picker-prompt')).toBeInTheDocument()
    expect(screen.getByText('完成选牌后点击发动')).toBeInTheDocument()
  })

  it('emits confirm when enabled and does not emit when disabled', async () => {
    const user = userEvent.setup()
    const { emitted, rerender } = render(CardPickerPromptRenderer, {
      props: baseProps(),
    })

    await user.click(screen.getByTestId('prompt-confirm-btn'))
    expect(emitted('confirm')).toEqual([[]])

    await rerender({
      ...baseProps(),
      canConfirm: false,
    })

    await user.click(screen.getByTestId('prompt-confirm-btn'))
    expect(emitted('confirm')).toEqual([[]])
  })

  it('renders cancel button only when showCancel is true and emits cancel', async () => {
    const user = userEvent.setup()
    const { emitted, rerender } = render(CardPickerPromptRenderer, {
      props: {
        ...baseProps(),
        showCancel: true,
      },
    })

    expect(screen.getByTestId('prompt-cancel-btn')).toBeInTheDocument()
    await user.click(screen.getByTestId('prompt-cancel-btn'))
    expect(emitted('cancel')).toEqual([[]])

    await rerender({
      ...baseProps(),
      showCancel: false,
    })

    expect(screen.queryByTestId('prompt-cancel-btn')).not.toBeInTheDocument()
  })

  it('renders confirm / cancel as labelled pills', () => {
    render(CardPickerPromptRenderer, {
      props: { ...baseProps(), showCancel: true, confirmLabel: '发动', cancelLabel: '不发动' },
    })

    expect(screen.getByTestId('prompt-confirm-btn')).toHaveTextContent('发动')
    expect(screen.getByTestId('prompt-confirm-btn')).toHaveClass('dc-pill--gold')
    expect(screen.getByTestId('prompt-cancel-btn')).toHaveTextContent('不发动')
  })

  it('does not render when hidden', () => {
    render(CardPickerPromptRenderer, {
      props: {
        ...baseProps(),
        visible: false,
      },
    })

    expect(screen.queryByTestId('card-picker-prompt')).not.toBeInTheDocument()
    expect(screen.queryByText('完成选牌后点击发动')).not.toBeInTheDocument()
  })
})
