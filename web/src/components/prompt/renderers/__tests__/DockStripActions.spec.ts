import { describe, expect, it } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/vue'
import DockStripActions from '../DockStripActions.vue'

describe('DockStripActions', () => {
  it('shows the live count against the cap and only confirms when allowed', async () => {
    const { emitted, rerender } = render(DockStripActions, {
      props: {
        message: '请选择要弃置的1张手牌：',
        rules: ['同系 · 任意张'],
        selected: 0,
        max: 1,
        canConfirm: false,
        showConfirm: true,
        showCancel: true,
      },
    })
    expect(screen.getByText('请选择要弃置的1张手牌：')).toBeTruthy()
    expect(screen.getByText('同系 · 任意张')).toBeTruthy()
    expect(screen.getByTestId('dock-strip-actions').textContent).toContain('已选 0/1')
    expect((screen.getByTestId('prompt-confirm-btn') as HTMLButtonElement).disabled).toBe(true)

    await rerender({ selected: 1, canConfirm: true })
    expect(screen.getByTestId('dock-strip-actions').textContent).toContain('已选 1/1')
    await fireEvent.click(screen.getByTestId('prompt-confirm-btn'))
    await fireEvent.click(screen.getByTestId('prompt-cancel-btn'))
    expect(emitted().confirm).toHaveLength(1)
    expect(emitted().cancel).toHaveLength(1)
  })

  it('hides confirm for single-click target pickers', () => {
    render(DockStripActions, {
      props: { message: '选择目标', selected: 0, max: 1, canConfirm: false, showConfirm: false, showCancel: false },
    })
    expect(screen.queryByTestId('prompt-confirm-btn')).toBeNull()
    expect(screen.queryByTestId('prompt-cancel-btn')).toBeNull()
  })
})
