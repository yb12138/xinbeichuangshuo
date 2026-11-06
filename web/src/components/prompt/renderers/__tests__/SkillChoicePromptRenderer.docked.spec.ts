import { describe, expect, it } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/vue'
import SkillChoicePromptRenderer from '../SkillChoicePromptRenderer.vue'
import { DECISION_DOCKED } from '../../dockContext'

const dockedGlobal = { provide: { [DECISION_DOCKED as symbol]: true } }

describe('SkillChoicePromptRenderer (docked)', () => {
  it('single response skill: the card is the description, 发动 fires it, the ghost button declines', async () => {
    const { emitted } = render(SkillChoicePromptRenderer, {
      props: {
        inlineVisible: false,
        overlayVisible: true,
        title: '你触发了响应技能【剑气斩】',
        buttons: [],
        branches: [
          { id: 'se_sword_qi_slash', title: '剑气斩', description: '移除X点剑气，对任意1名角色造成X点法术伤害。', disabled: false },
          { id: 'skip', title: '', description: '不发动响应技能', disabled: false },
        ],
      },
      global: dockedGlobal,
    })
    expect(screen.queryByTestId('skill-branch-overlay')).toBeNull()
    expect(screen.getByText('剑气斩')).toBeTruthy()
    const fire = screen.getByTestId('decision-confirm-btn') as HTMLButtonElement
    expect(fire.disabled).toBe(false)
    await fireEvent.click(fire)
    await fireEvent.click(screen.getByTestId('prompt-option-skip'))
    expect(emitted().select).toEqual([['se_sword_qi_slash'], ['skip']])
    expect(screen.getByTestId('prompt-option-skip').textContent).toContain('不发动响应技能')
  })

  it('several skills: must pick one before 发动', async () => {
    const { emitted } = render(SkillChoicePromptRenderer, {
      props: {
        inlineVisible: false,
        overlayVisible: true,
        title: '选择要发动的技能',
        buttons: [],
        branches: [
          { id: 'a', title: '技能A', description: '', disabled: false },
          { id: 'b', title: '技能B', description: '', disabled: false },
        ],
      },
      global: dockedGlobal,
    })
    const fire = screen.getByTestId('decision-confirm-btn') as HTMLButtonElement
    expect(fire.disabled).toBe(true)
    await fireEvent.click(screen.getByTestId('branch-option-1'))
    await fireEvent.click(fire)
    expect(emitted().select).toEqual([['b']])
  })
})
