import { render } from '@testing-library/vue'
import { describe, expect, it } from 'vitest'
import CardComponent from '../CardComponent.vue'
import type { Card } from '../../types/game'

function buildCard(overrides: Partial<Card> = {}): Card {
  return { id: 'c1', name: '火焰斩', type: 'Attack', element: 'Fire', damage: 2, description: '', ...overrides }
}

describe('CardComponent 命格', () => {
  it('shows the faction as a seal at the top of the card instead of in the footer ribbon', () => {
    const { getByTestId, container } = render(CardComponent, { props: { card: buildCard({ faction: '圣' }) } })
    const seal = getByTestId('card-fate-seal')
    expect(seal.textContent?.trim()).toBe('圣')
    expect(seal.getAttribute('title')).toBe('圣命格')
    expect(container.querySelector('.card-ribbon-text')?.textContent).not.toContain('圣')
  })

  it('omits the seal for cards without a faction', () => {
    const { queryByTestId } = render(CardComponent, { props: { card: buildCard() } })
    expect(queryByTestId('card-fate-seal')).toBeNull()
  })
})
