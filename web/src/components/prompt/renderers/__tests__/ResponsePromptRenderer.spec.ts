import { render, screen } from '@testing-library/vue'
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
      consequence: '承受火系攻击伤害',
    },
    {
      id: 'defend',
      buttonLabel: '防御',
      disabled: false,
      kind: 'defend' as const,
      consequence: '抵挡本次攻击',
    },
    {
      id: 'counter',
      buttonLabel: '应战',
      disabled: true,
      kind: 'counter' as const,
      consequence: '打出火系攻击或暗灭',
      disabledReason: '没有可用的同系牌',
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
  it('renders the element hint and one tinted card per response with its consequence', () => {
    render(ResponsePromptRenderer, {
      props: baseProps(),
    })

    expect(screen.getByTestId('response-prompt')).toBeInTheDocument()
    expect(screen.getByText('此次攻击系别：火系（应战需同系或暗灭）')).toBeInTheDocument()
    const take = screen.getByTestId('prompt-option-take')
    expect(take).toHaveClass('dc-tile--take')
    expect(take).toHaveTextContent('命中')
    expect(take).toHaveTextContent('承受火系攻击伤害')
    expect(screen.getByTestId('prompt-option-defend')).toHaveClass('dc-tile--defend')
    expect(screen.getByTestId('prompt-option-defend')).toHaveTextContent('防御')
    const counter = screen.getByTestId('prompt-option-counter')
    expect(counter).toHaveClass('dc-tile--counter')
    expect(counter).toHaveTextContent('应战 / 传递')
    // 禁用时写原因而不是后果
    expect(counter).toHaveTextContent('没有可用的同系牌')
    expect(counter).not.toHaveTextContent('打出火系攻击或暗灭')
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
