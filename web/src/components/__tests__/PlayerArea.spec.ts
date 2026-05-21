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

  it('renders sword emperor sword qi with Chinese label', () => {
    render(PlayerArea, {
      props: {
        player: buildPlayer({
          tokens: { se_sword_qi: 2 },
        }),
      },
      global: { plugins: [createPinia()] },
    })

    expect(screen.getByText('剑气')).toBeTruthy()
    expect(screen.getByTitle('剑气: 2')).toBeTruthy()
    expect(screen.queryByText('se_sword_qi')).toBeNull()
  })

  it('keeps magic lancer turn-state indicators out of the token chip row', () => {
    render(PlayerArea, {
      props: {
        player: buildPlayer({
          tokens: { hero_anger: 2 },
          indicators: {
            ml_dark_release_next_attack_bonus: 1,
            ml_dark_release_lock_turn: 1,
          },
        }),
      },
      global: { plugins: [createPinia()] },
    })

    expect(screen.getByText('怒气')).toBeTruthy()
    expect(screen.queryByText('下次主动攻+伤')).toBeNull()
    expect(screen.queryByText('本回合锁技能')).toBeNull()
  })

  it('keeps spirit caster power count out of the token chip row', () => {
    render(PlayerArea, {
      props: {
        player: buildPlayer({
          indicators: { sc_power_count: 2 },
        }),
      },
      global: { plugins: [createPinia()] },
    })

    expect(screen.queryByText('sc_power_count')).toBeNull()
    expect(screen.queryByTitle('sc_power_count: 2')).toBeNull()
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

  it('renders shared life as a portrait badge instead of a token chip', () => {
    render(PlayerArea, {
      props: {
        player: buildPlayer({
          tokens: {
            bp_shared_life_active: 1,
            bp_shared_life_bound: 1,
          },
          indicators: {
            bp_shared_life_active: 1,
            bp_shared_life_bound: 1,
          },
        }),
        turnOrder: 3,
        bloodSharedLifeText: '同生共死',
        bloodSharedLifeTitle: '同生共死：血之巫女 与 圣女 的手牌上限保持联动',
        bloodSharedLifeRole: 'source',
      },
      global: { plugins: [createPinia()] },
    })

    expect(screen.getByText('同生共死')).toBeTruthy()
    expect(screen.queryByTitle('同生共死在场: 1')).toBeNull()
    expect(screen.queryByTitle('同生共死绑定: 1')).toBeNull()
  })
})
