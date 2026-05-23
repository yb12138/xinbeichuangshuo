import { fireEvent, render, screen } from '@testing-library/vue'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'
import ExtractPromptRenderer from '../ExtractPromptRenderer.vue'

function baseProps() {
  return {
    visible: true,
    options: [
      { id: 'ruby', label: '红宝石' },
      { id: 'crystal', label: '蓝水晶' },
    ],
    selectedIndexes: [],
    min: 1,
    max: 2,
    confirmImageSrc: '/assets/ui/action_confirm.png',
    confirmImageReady: true,
    confirmFallbackText: '确',
  }
}

describe('ExtractPromptRenderer', () => {
  it('renders ruby and crystal labels', () => {
    render(ExtractPromptRenderer, {
      props: baseProps(),
    })

    expect(screen.getByTestId('extract-prompt')).toBeInTheDocument()
    expect(screen.getByText('♦ 红宝石')).toBeInTheDocument()
    expect(screen.getByText('🔷 蓝水晶')).toBeInTheDocument()
  })

  it('emits toggle(index) when option button is clicked', async () => {
    const user = userEvent.setup()
    const { emitted } = render(ExtractPromptRenderer, {
      props: baseProps(),
    })

    await user.click(screen.getByTestId('extract-option-1'))

    expect(emitted('toggle')).toEqual([[1]])
  })

  it('applies selected class by selectedIndexes', () => {
    render(ExtractPromptRenderer, {
      props: {
        ...baseProps(),
        selectedIndexes: [1],
      },
    })

    expect(screen.getByTestId('extract-option-1')).toHaveClass('extract-option-btn--selected')
    expect(screen.getByTestId('extract-option-0')).not.toHaveClass('extract-option-btn--selected')
  })

  it('does not emit confirm while disabled', async () => {
    const user = userEvent.setup()
    const { emitted } = render(ExtractPromptRenderer, {
      props: {
        ...baseProps(),
        min: 2,
        selectedIndexes: [0],
      },
    })

    const confirmButton = screen.getByTestId('prompt-confirm-btn')
    expect(confirmButton).toBeDisabled()

    await user.click(confirmButton)

    expect(emitted('confirm')).toBeUndefined()
  })

  it('emits confirm when confirmation is enabled', async () => {
    const user = userEvent.setup()
    const { emitted } = render(ExtractPromptRenderer, {
      props: {
        ...baseProps(),
        selectedIndexes: [0, 1],
      },
    })

    const confirmButton = screen.getByTestId('prompt-confirm-btn')
    expect(confirmButton).toHaveAttribute('title', '确认提炼（2/2）')
    expect(confirmButton).toHaveAttribute('aria-label', '确认提炼（2/2）')

    await user.click(confirmButton)

    expect(emitted('confirm')).toEqual([[]])
  })

  it('emits confirmImageError when confirm image fails', async () => {
    const { container, emitted } = render(ExtractPromptRenderer, {
      props: baseProps(),
    })

    const image = container.querySelector('img')
    expect(image).not.toBeNull()
    await fireEvent.error(image!)

    expect(emitted('confirmImageError')).toEqual([[]])
  })

  it('does not render while hidden', () => {
    render(ExtractPromptRenderer, {
      props: {
        ...baseProps(),
        visible: false,
      },
    })

    expect(screen.queryByTestId('extract-prompt')).not.toBeInTheDocument()
    expect(screen.queryByTestId('prompt-confirm-btn')).not.toBeInTheDocument()
  })
})
