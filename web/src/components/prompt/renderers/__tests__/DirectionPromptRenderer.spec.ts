import { render, screen } from '@testing-library/vue'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'
import DirectionPromptRenderer from '../DirectionPromptRenderer.vue'

function baseProps() {
  return {
    visible: true,
    title: '请选择方向',
    options: [
      {
        id: 'normal',
        label: '正向',
        hint: '沿当前方向推进',
        description: '沿当前方向推进',
        disabled: false,
        tone: 'prompt-direction--normal',
        icon: 'arrow-right',
      },
      {
        id: 'reverse',
        label: '逆向',
        hint: '反向切换',
        description: '反向切换',
        disabled: true,
        tone: 'prompt-direction--reverse',
        icon: 'arrow-left',
      },
    ],
  }
}

describe('DirectionPromptRenderer', () => {
  it('renders title, options and descriptions', () => {
    render(DirectionPromptRenderer, {
      props: baseProps(),
    })

    expect(screen.getByTestId('direction-prompt')).toBeInTheDocument()
    expect(screen.getByTestId('decision-overlay')).toBeInTheDocument()
    expect(screen.getByText('请选择方向')).toBeInTheDocument()
    expect(screen.getByTestId('direction-option-normal')).toBeInTheDocument()
    expect(screen.getByTestId('direction-option-reverse')).toBeInTheDocument()
    expect(screen.getByText('沿当前方向推进')).toBeInTheDocument()
    expect(screen.getByText('反向切换')).toBeInTheDocument()
  })

  it('emits select when option is clicked', async () => {
    const user = userEvent.setup()
    const { emitted } = render(DirectionPromptRenderer, {
      props: baseProps(),
    })

    await user.click(screen.getByTestId('direction-option-normal'))

    expect(emitted('select')).toEqual([['normal']])
  })

  it('does not emit when option is disabled', async () => {
    const user = userEvent.setup()
    const { emitted } = render(DirectionPromptRenderer, {
      props: baseProps(),
    })

    await user.click(screen.getByTestId('direction-option-reverse'))

    expect(emitted('select')).toBeUndefined()
  })

  it('does not render when hidden', () => {
    render(DirectionPromptRenderer, {
      props: {
        ...baseProps(),
        visible: false,
      },
    })

    expect(screen.queryByTestId('direction-prompt')).not.toBeInTheDocument()
    expect(screen.queryByTestId('decision-overlay')).not.toBeInTheDocument()
  })
})
