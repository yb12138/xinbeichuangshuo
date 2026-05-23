import { render, screen } from '@testing-library/vue'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'
import DecisionOverlayRenderer from '../DecisionOverlayRenderer.vue'

function baseProps() {
  return {
    visible: true,
    title: '请选择',
    mode: 'text' as const,
    options: [
      {
        id: 'branch1',
        label: '分支①：移除1个闇月，令目标角色+1治疗',
        buttonLabel: '分支①：移除1个闇月，令目标角色+1治疗',
      },
      {
        id: 'branch2',
        label: '分支②：移除1点治疗，你+1新月',
        buttonLabel: '分支②：移除1点治疗，你+1新月',
      },
    ],
    activationHint: '',
    activationOptionId: '',
    activationDisabled: false,
    canCancel: true,
    cancelLabel: '取消',
    cancelOptionId: 'cancel',
  }
}

describe('DecisionOverlayRenderer', () => {
  it('renders text mode and emits select', async () => {
    const user = userEvent.setup()
    const { emitted } = render(DecisionOverlayRenderer, {
      props: baseProps(),
    })

    expect(screen.getByTestId('decision-overlay')).toBeInTheDocument()
    await user.click(screen.getByTestId('prompt-option-branch1'))

    expect(emitted('select')).toEqual([['branch1']])
  })

  it('renders numeric mode and emits select', async () => {
    const user = userEvent.setup()
    const { emitted } = render(DecisionOverlayRenderer, {
      props: {
        ...baseProps(),
        mode: 'numeric',
        options: [
          { id: '0', label: '不使用治疗', buttonLabel: '0' },
          { id: '1', label: '使用 1 点治疗', buttonLabel: '1' },
          { id: '2', label: '使用 2 点治疗', buttonLabel: '2' },
        ],
      },
    })

    await user.click(screen.getByTestId('numeric-option-2'))

    expect(emitted('select')).toEqual([['2']])
  })

  it('renders activation-cost mode and emits select', async () => {
    const user = userEvent.setup()
    const { emitted } = render(DecisionOverlayRenderer, {
      props: {
        ...baseProps(),
        mode: 'activation-cost',
        activationHint: '你需要支付 1 点代价以发动技能',
        activationOptionId: 'confirm',
      },
    })

    expect(screen.getByText('你需要支付 1 点代价以发动技能')).toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: '确认发动' }))

    expect(emitted('select')).toEqual([['confirm']])
  })

  it('renders yes-no mode and emits selected option id', async () => {
    const user = userEvent.setup()
    const { emitted } = render(DecisionOverlayRenderer, {
      props: {
        ...baseProps(),
        mode: 'yes-no',
        options: [
          { id: 'yes', label: '是', buttonLabel: '是' },
          { id: 'no', label: '否', buttonLabel: '否' },
        ],
      },
    })

    await user.click(screen.getByTestId('prompt-option-no'))

    expect(emitted('select')).toEqual([['no']])
    expect(screen.queryByTestId('prompt-cancel-btn')).not.toBeInTheDocument()
  })

  it('does not emit select for disabled numeric options', async () => {
    const user = userEvent.setup()
    const { emitted } = render(DecisionOverlayRenderer, {
      props: {
        ...baseProps(),
        mode: 'numeric',
        options: [
          { id: '0', label: '不使用治疗', buttonLabel: '0' },
          { id: '1', label: '使用 1 点治疗', buttonLabel: '1', disabled: true },
        ],
      },
    })

    await user.click(screen.getByTestId('numeric-option-1'))

    expect(emitted('select')).toBeUndefined()
  })

  it('emits cancel when cancel button is clicked', async () => {
    const user = userEvent.setup()
    const { emitted } = render(DecisionOverlayRenderer, {
      props: baseProps(),
    })

    await user.click(screen.getByTestId('prompt-cancel-btn'))

    expect(emitted('cancel')).toEqual([['cancel']])
  })
})
