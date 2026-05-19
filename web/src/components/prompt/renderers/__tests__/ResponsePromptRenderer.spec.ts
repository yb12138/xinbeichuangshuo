import { fireEvent, render, screen } from '@testing-library/vue'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'
import ResponsePromptRenderer from '../ResponsePromptRenderer.vue'

function baseOptions() {
  return [
    {
      id: 'take',
      buttonLabel: '承受命中',
      disabled: false,
      kind: 'take' as const,
      imageSrc: '/assets/ui/prompt_btn_take.png',
      imageReady: true,
      fallbackText: '命',
      enlarged: false,
    },
    {
      id: 'defend',
      buttonLabel: '防御',
      disabled: false,
      kind: 'defend' as const,
      imageSrc: '/assets/ui/prompt_btn_defend.png',
      imageReady: true,
      fallbackText: '防',
      enlarged: true,
    },
    {
      id: 'counter',
      buttonLabel: '应战',
      disabled: true,
      kind: 'counter' as const,
      imageSrc: '/assets/ui/prompt_btn_counter.png',
      imageReady: false,
      fallbackText: '应',
      enlarged: false,
    },
  ]
}

function baseProps() {
  return {
    visible: true,
    hint: '此次攻击系别：火系（应战需同系或暗灭）',
    options: baseOptions(),
  }
}

describe('ResponsePromptRenderer', () => {
  it('renders attack element hint and response buttons', () => {
    render(ResponsePromptRenderer, {
      props: baseProps(),
    })

    expect(screen.getByTestId('response-prompt')).toBeInTheDocument()
    expect(screen.getByText('此次攻击系别：火系（应战需同系或暗灭）')).toBeInTheDocument()
    expect(screen.getByTestId('prompt-option-take')).toHaveClass('prompt-response-btn--take')
    expect(screen.getByTestId('prompt-option-defend')).toHaveClass('prompt-response-btn--defend')
    expect(screen.getByTestId('prompt-option-counter')).toHaveClass('prompt-response-btn--counter')
    expect(screen.getByTestId('prompt-option-defend')).toHaveClass('prompt-response-btn--large')
    expect(screen.getByText('应')).toBeInTheDocument()
  })

  it('emits selected option id', async () => {
    const user = userEvent.setup()
    const { emitted } = render(ResponsePromptRenderer, {
      props: baseProps(),
    })

    await user.click(screen.getByTestId('prompt-option-defend'))

    expect(emitted('select')).toEqual([['defend']])
  })

  it('does not emit select when an option is disabled', async () => {
    const user = userEvent.setup()
    const { emitted } = render(ResponsePromptRenderer, {
      props: baseProps(),
    })

    await user.click(screen.getByTestId('prompt-option-counter'))

    expect(emitted('select')).toBeUndefined()
  })

  it('emits image error with option id', async () => {
    const { container, emitted } = render(ResponsePromptRenderer, {
      props: baseProps(),
    })

    const image = container.querySelector('img')
    expect(image).not.toBeNull()
    await fireEvent.error(image!)

    expect(emitted('imageError')).toEqual([['take']])
  })

  it('does not render while hidden', () => {
    render(ResponsePromptRenderer, {
      props: {
        ...baseProps(),
        visible: false,
      },
    })

    expect(screen.queryByTestId('response-prompt')).not.toBeInTheDocument()
  })
})
