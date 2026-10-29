import { render, screen } from '@testing-library/vue'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'
import TargetPickerPromptRenderer from '../TargetPickerPromptRenderer.vue'

function baseProps() {
  return {
    visible: true,
    message: '请选择目标角色',
    showConfirm: true,
    canConfirm: true,
  }
}

describe('TargetPickerPromptRenderer', () => {
  it('renders the target picker hint message', () => {
    render(TargetPickerPromptRenderer, {
      props: baseProps(),
    })

    expect(screen.getByTestId('target-picker-prompt')).toBeInTheDocument()
    expect(screen.getByText('请选择目标角色')).toBeInTheDocument()
  })

  it('does not render a confirm button when manual confirm is not needed', () => {
    render(TargetPickerPromptRenderer, {
      props: {
        ...baseProps(),
        showConfirm: false,
      },
    })

    expect(screen.queryByTestId('prompt-confirm-btn')).not.toBeInTheDocument()
  })

  it('emits confirm when the confirm button is clicked', async () => {
    const user = userEvent.setup()
    const { emitted } = render(TargetPickerPromptRenderer, {
      props: baseProps(),
    })

    await user.click(screen.getByTestId('prompt-confirm-btn'))

    expect(emitted('confirm')).toEqual([[]])
  })

  it('does not emit confirm while disabled', async () => {
    const user = userEvent.setup()
    const { emitted } = render(TargetPickerPromptRenderer, {
      props: {
        ...baseProps(),
        canConfirm: false,
      },
    })

    await user.click(screen.getByTestId('prompt-confirm-btn'))

    expect(emitted('confirm')).toBeUndefined()
  })

  it('renders confirm as a gold pill with a text label', () => {
    render(TargetPickerPromptRenderer, { props: baseProps() })

    const confirm = screen.getByTestId('prompt-confirm-btn')
    expect(confirm).toHaveTextContent('确认')
    expect(confirm).toHaveClass('dc-pill', 'dc-pill--gold')
  })

  it('does not render while hidden', () => {
    render(TargetPickerPromptRenderer, {
      props: {
        ...baseProps(),
        visible: false,
      },
    })

    expect(screen.queryByTestId('target-picker-prompt')).not.toBeInTheDocument()
    expect(screen.queryByText('请选择目标角色')).not.toBeInTheDocument()
  })
})
