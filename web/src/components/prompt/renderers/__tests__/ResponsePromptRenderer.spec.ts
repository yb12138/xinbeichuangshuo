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
  it('renders the element hint and one button per response option', () => {
    render(ResponsePromptRenderer, { props: baseProps() })

    expect(screen.getByTestId('response-prompt')).toBeInTheDocument()
    expect(screen.getByText('此次攻击系别：火系（应战需同系或暗灭）')).toBeInTheDocument()
    expect(screen.getByTestId('response-take')).toHaveTextContent('命中')
    expect(screen.getByTestId('response-defend')).toHaveTextContent('防御')
    expect(screen.getByTestId('response-counter')).toHaveTextContent('应战 / 传递')
  })

  it('emits selected option id', async () => {
    const user = userEvent.setup()
    const { emitted } = render(ResponsePromptRenderer, { props: baseProps() })
    await user.click(screen.getByTestId('response-take'))
    expect(emitted().select).toEqual([['take']])
  })

  it('does not emit select when an option is disabled', async () => {
    const user = userEvent.setup()
    const { emitted } = render(ResponsePromptRenderer, { props: baseProps() })
    await user.click(screen.getByTestId('response-counter'))
    expect(emitted().select).toBeUndefined()
  })

  it('renders response prompt through the response renderer and keeps take submit behavior', () => {
    render(ResponsePromptRenderer, { props: baseProps() })
    const actions = screen.getByTestId('response-actions')
    expect(actions.querySelectorAll('button')).toHaveLength(3)
    expect(screen.getByTestId('response-take')).not.toBeDisabled()
    expect(screen.getByTestId('response-counter')).toBeDisabled()
  })

  it('mentions the faction when the responder can counter by 命格 (阴阳师)', () => {
    render(ResponsePromptRenderer, {
      props: {
        ...baseProps(),
        hint: '此次攻击系别：火系·幻命格（应战需同系或暗灭）',
      },
    })
    expect(screen.getByTestId('response-hint')?.textContent).toContain('幻命格')
  })
})
