import { render, screen } from '@testing-library/vue'
import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it } from 'vitest'
import PlayerArea from '../PlayerArea.vue'
import type { PlayerView } from '../../types/game'

function buildPlayer(overrides: Partial<PlayerView> = {}): PlayerView {
  return {
    id: 'p1',
    name: 'P1',
    camp: 'Red',
    role: 'fighter',
    hand_count: 0,
    max_hand: 6,
    exclusive_card_count: 0,
    hand: [],
    exclusive_cards: [],
    field: [],
    heal: 0,
    max_heal: 0,
    gem: 0,
    crystal: 0,
    is_active: false,
    buffs: [],
    tokens: {},
    indicators: {},
    ...overrides,
  }
}

describe('PlayerArea indicators', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('renders tokens together with derived indicators', () => {
    render(PlayerArea, {
      props: {
        player: buildPlayer({
          tokens: { hero_anger: 2 },
          indicators: { ml_dark_release_next_attack_bonus: 1 },
        }),
      },
      global: { plugins: [createPinia()] },
    })

    expect(screen.getByText('怒气')).toBeTruthy()
    expect(screen.getByText('下次主动攻+伤')).toBeTruthy()
  })

  it('lets indicators override stale legacy token values', () => {
    render(PlayerArea, {
      props: {
        player: buildPlayer({
          tokens: { mb_charge_count: 99 },
          indicators: { mb_charge_count: 1 },
        }),
      },
      global: { plugins: [createPinia()] },
    })

    expect(screen.getByTitle('充能: 1')).toBeTruthy()
    expect(screen.queryByTitle('充能: 99')).toBeNull()
  })
})
