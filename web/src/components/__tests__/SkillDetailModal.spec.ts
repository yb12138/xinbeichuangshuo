import { render, screen } from '@testing-library/vue'
import { describe, expect, it } from 'vitest'
import SkillDetailModal from '../SkillDetailModal.vue'

describe('SkillDetailModal', () => {
  it('shows the exclusive tag for exclusive skills', () => {
    render(SkillDetailModal, {
      props: {
        visible: true,
        character: {
          id: 'fighter',
          name: '战士',
          title: '测试角色',
          faction: 'Red',
          skills: [
            {
              id: 'exclusive-skill',
              title: '独有技',
              description: 'exclusive skill',
              type: 2,
              min_targets: 1,
              max_targets: 1,
              target_type: 2,
              cost_gem: 0,
              cost_crystal: 0,
              cost_discards: 0,
              require_exclusive: true,
            },
          ],
        },
      },
    })

    expect(screen.getByText('独有')).toBeInTheDocument()
  })
})
