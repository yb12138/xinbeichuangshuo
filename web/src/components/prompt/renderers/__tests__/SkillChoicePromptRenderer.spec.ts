import { fireEvent, render, screen } from '@testing-library/vue'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'
import SkillChoicePromptRenderer from '../SkillChoicePromptRenderer.vue'

function baseButtons() {
  return [
    {
      id: 'confirm',
      label: '发动',
      disabled: false,
      cancel: false,
      imageSrc: '/assets/ui/action_confirm.png',
      imageReady: true,
      fallbackText: '确',
    },
    {
      id: 'cancel',
      label: '取消',
      disabled: false,
      cancel: true,
      imageSrc: '/assets/ui/action_cancel_btn.png',
      imageReady: true,
      fallbackText: '消',
    },
  ]
}

function baseBranches() {
  return [
    {
      id: 'wind_fury',
      title: '风怒',
      description: '额外攻击一次',
      cost: '[消耗1风]',
      disabled: false,
    },
    {
      id: 'holy_light',
      title: '圣光',
      description: '恢复治疗',
      disabled: false,
    },
  ]
}

describe('SkillChoicePromptRenderer', () => {
  it('renders inline skill buttons and emits selected option id', async () => {
    const user = userEvent.setup()
    const { emitted } = render(SkillChoicePromptRenderer, {
      props: {
        inlineVisible: true,
        overlayVisible: false,
        title: '是否发动【风怒】',
        buttons: baseButtons(),
        branches: [],
      },
    })

    expect(screen.getByTestId('skill-choice-prompt')).toBeInTheDocument()
    expect(screen.getByText('是否发动【风怒】')).toBeInTheDocument()
    expect(screen.getByTestId('prompt-option-confirm')).toHaveAttribute('aria-label', '发动')

    await user.click(screen.getByTestId('prompt-option-confirm'))

    expect(emitted('select')).toEqual([['confirm']])
  })

  it('does not emit select for a disabled inline button', async () => {
    const user = userEvent.setup()
    const { emitted } = render(SkillChoicePromptRenderer, {
      props: {
        inlineVisible: true,
        overlayVisible: false,
        title: '是否发动【风怒】',
        buttons: [
          {
            ...baseButtons()[0]!,
            disabled: true,
          },
        ],
        branches: [],
      },
    })

    const button = screen.getByTestId('prompt-option-confirm')
    expect(button).toBeDisabled()

    await user.click(button)

    expect(emitted('select')).toBeUndefined()
  })

  it('emits imageError with option id when inline image fails', async () => {
    const { container, emitted } = render(SkillChoicePromptRenderer, {
      props: {
        inlineVisible: true,
        overlayVisible: false,
        title: '是否发动【风怒】',
        buttons: baseButtons(),
        branches: [],
      },
    })

    const image = container.querySelector('img')
    expect(image).not.toBeNull()
    await fireEvent.error(image!)

    expect(emitted('imageError')).toEqual([['confirm']])
  })

  it('renders inline fallback text when image is not ready', () => {
    render(SkillChoicePromptRenderer, {
      props: {
        inlineVisible: true,
        overlayVisible: false,
        title: '是否发动【风怒】',
        buttons: [
          {
            ...baseButtons()[0]!,
            imageReady: false,
          },
        ],
        branches: [],
      },
    })

    expect(screen.getByText('确')).toBeInTheDocument()
  })

  it('renders overlay branches and emits branch / skip selections', async () => {
    const user = userEvent.setup()
    const { emitted } = render(SkillChoicePromptRenderer, {
      props: {
        inlineVisible: false,
        overlayVisible: true,
        title: '请选择要发动的技能',
        buttons: [],
        branches: baseBranches(),
      },
    })

    expect(screen.getByTestId('skill-branch-overlay')).toBeInTheDocument()
    expect(screen.getByTestId('decision-overlay')).toBeInTheDocument()
    expect(screen.getByText('请选择要发动的技能')).toBeInTheDocument()
    expect(screen.getByText('风怒')).toBeInTheDocument()
    expect(screen.getByText('额外攻击一次')).toBeInTheDocument()
    expect(screen.getByText('[消耗1风]')).toBeInTheDocument()

    await user.click(screen.getByTestId('branch-option-1'))
    await user.click(screen.getByTestId('prompt-option-skip'))

    expect(emitted('select')).toEqual([['holy_light'], ['skip']])
  })

  it('does not render hidden inline or overlay surfaces', () => {
    render(SkillChoicePromptRenderer, {
      props: {
        inlineVisible: false,
        overlayVisible: false,
        title: '请选择',
        buttons: baseButtons(),
        branches: baseBranches(),
      },
    })

    expect(screen.queryByTestId('skill-choice-prompt')).not.toBeInTheDocument()
    expect(screen.queryByTestId('skill-branch-overlay')).not.toBeInTheDocument()
  })
})
