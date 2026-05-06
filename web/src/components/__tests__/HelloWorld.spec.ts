import { render, screen } from '@testing-library/vue'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'
import HelloWorld from '../HelloWorld.vue'

describe('HelloWorld component interaction', () => {
  it('increments count text after clicking the button', async () => {
    render(HelloWorld, {
      props: {
        msg: 'Front-end interaction test',
      },
    })

    const button = screen.getByRole('button', { name: /count is 0/i })
    await userEvent.click(button)

    expect(screen.getByRole('button', { name: /count is 1/i })).toBeInTheDocument()
  })
})
