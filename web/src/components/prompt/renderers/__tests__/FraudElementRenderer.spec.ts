import { render, screen } from '@testing-library/vue'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'
import FraudElementRenderer from '../FraudElementRenderer.vue'

function baseProps() {
  return {
    visible: true,
    title: '请选择本次攻击系别',
    options: [
      { id: 'water', title: '水系', glyph: '水', tone: 'prompt-fraud-card--water' },
      { id: 'fire', title: '火系', glyph: '火', tone: 'prompt-fraud-card--fire' },
    ],
  }
}

describe('FraudElementRenderer', () => {
  it('renders title and fraud element cards', () => {
    render(FraudElementRenderer, {
      props: baseProps(),
    })

    expect(screen.getByTestId('decision-overlay')).toBeInTheDocument()
    expect(screen.getByText('请选择本次攻击系别')).toBeInTheDocument()
    expect(screen.getByTestId('prompt-option-water')).toBeInTheDocument()
    expect(screen.getByTestId('prompt-option-fire')).toBeInTheDocument()
  })

  it('emits select when option button is clicked', async () => {
    const user = userEvent.setup()
    const { emitted } = render(FraudElementRenderer, {
      props: baseProps(),
    })

    await user.click(screen.getByTestId('prompt-option-fire'))

    expect(emitted('select')).toEqual([['fire']])
  })

  it('renders generic tone fallback option', () => {
    render(FraudElementRenderer, {
      props: {
        ...baseProps(),
        options: [{ id: 'light', title: '光系', glyph: '光', tone: 'prompt-fraud-card--generic' }],
      },
    })

    expect(screen.getByTestId('prompt-option-light')).toHaveClass('prompt-fraud-card--generic')
    expect(screen.getByText('光系')).toBeInTheDocument()
  })

  it('does not render when visible is false', () => {
    render(FraudElementRenderer, {
      props: {
        ...baseProps(),
        visible: false,
      },
    })

    expect(screen.queryByTestId('decision-overlay')).not.toBeInTheDocument()
    expect(screen.queryByTestId('prompt-option-water')).not.toBeInTheDocument()
  })
})
