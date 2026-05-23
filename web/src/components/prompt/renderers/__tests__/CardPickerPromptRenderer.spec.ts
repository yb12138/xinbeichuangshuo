import { fireEvent, render, screen } from '@testing-library/vue'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'
import CardPickerPromptRenderer from '../CardPickerPromptRenderer.vue'

function baseProps() {
  return {
    visible: true,
    message: '完成选牌后点击发动',
    canConfirm: true,
    showCancel: false,
    confirmTitle: '发动',
    confirmAriaLabel: '发动',
    confirmImageSrc: '/assets/ui/action_confirm.png',
    confirmImageReady: true,
    confirmFallbackText: '确',
    cancelImageSrc: '/assets/ui/action_cancel_btn.png',
    cancelImageReady: true,
    cancelFallbackText: '消',
    cancelTitle: '取消',
    cancelAriaLabel: '取消',
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

  it('emits image error events and shows fallback text when images are unavailable', async () => {
    const { container, emitted, rerender } = render(CardPickerPromptRenderer, {
      props: {
        ...baseProps(),
        showCancel: true,
      },
    })

    const images = container.querySelectorAll('img')
    expect(images.length).toBe(2)
    await fireEvent.error(images[0]!)
    await fireEvent.error(images[1]!)
    expect(emitted('confirmImageError')).toEqual([[]])
    expect(emitted('cancelImageError')).toEqual([[]])

    await rerender({
      ...baseProps(),
      showCancel: true,
      confirmImageReady: false,
      cancelImageReady: false,
    })

    expect(screen.getByText('确')).toBeInTheDocument()
    expect(screen.getByText('消')).toBeInTheDocument()
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
