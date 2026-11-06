import { render, screen } from '@testing-library/vue'
import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it } from 'vitest'
import PlayerArea from '../PlayerArea.vue'
import type { FieldCard, PlayerView } from '../../types/game'

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

function buildEffectField(effect: string): FieldCard {
  return {
    card: {
      id: `field-${effect}`,
      name: effect,
      type: 'Magic',
      element: 'Fire',
      damage: 0,
      description: '',
    },
    owner_id: 'p1',
    source_id: 'p1',
    mode: 'Effect',
    effect,
    field_hook: '',
    locked: false,
    duration: 0,
  }
}

describe('PlayerArea indicators', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('renders sword emperor sword qi as an icon with its Chinese name', () => {
    render(PlayerArea, {
      props: {
        player: buildPlayer({
          tokens: { se_sword_qi: 2 },
        }),
      },
      global: { plugins: [createPinia()] },
    })

    expect(screen.getByRole('img', { name: '剑气' })).toBeTruthy()
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

    expect(screen.getByTitle('怒气: 2')).toBeTruthy()
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

  it('shows cover cards as one card-back mark with a count, and hides the duplicated count token', () => {
    render(PlayerArea, {
      props: {
        player: buildPlayer({
          tokens: { se_sword_soul_count: 2, se_sword_qi: 1 },
          field: [
            { ...buildEffectField('SwordSoul'), mode: 'Cover' },
            { ...buildEffectField('SwordSoul'), mode: 'Cover' },
          ],
        }),
      },
      global: { plugins: [createPinia()] },
    })

    expect(screen.getByTitle('剑魂: 2')).toBeTruthy()
    expect(screen.getByRole('img', { name: '剑魂' })).toBeTruthy()
    expect(screen.queryByTitle('剑魂: 2')?.textContent).toContain('2')
    expect(screen.getAllByTitle(/^剑魂/)).toHaveLength(1)
  })

  it('draws small-cap tokens as a dot grid', () => {
    const { container } = render(PlayerArea, {
      props: { player: buildPlayer({ tokens: { onmyoji_ghost_fire: 2 } }) },
      global: { plugins: [createPinia()] },
    })

    expect(screen.getByTitle('鬼火: 2')).toBeTruthy()
    expect(container.querySelectorAll('.seat-mark-pip')).toHaveLength(3)
    expect(container.querySelectorAll('.seat-mark-pip--on')).toHaveLength(2)
  })

  it('keeps blood cap out of the token chip row', () => {
    render(PlayerArea, {
      props: {
        player: buildPlayer({
          tokens: { css_blood: 2, css_blood_cap: 4 },
        }),
      },
      global: { plugins: [createPinia()] },
    })

    expect(screen.getByTitle('鲜血: 2')).toBeTruthy()
    expect(screen.queryByText('鲜血上限')).toBeNull()
    expect(screen.queryByTitle('css_blood_cap: 4')).toBeNull()
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

  it('renders hundred dragon lock as a portrait badge instead of a link marker', () => {
    render(PlayerArea, {
      props: {
        player: buildPlayer(),
        turnOrder: 2,
        fighterHundredDragonText: '幻龙锁定',
        fighterHundredDragonTitle: '百式幻龙拳：格斗家 锁定 圣女，本行动阶段只能主动攻击该角色',
        fighterHundredDragonRole: 'bound',
      },
      global: { plugins: [createPinia()] },
    })

    expect(screen.getByText('幻龙锁定')).toBeTruthy()
    expect(screen.getByTitle('百式幻龙拳：格斗家 锁定 圣女，本行动阶段只能主动攻击该角色')).toBeTruthy()
  })

  it('renders stable identity, active, and targetable badges', () => {
    render(PlayerArea, {
      props: {
        player: buildPlayer({ is_active: true }),
        isMe: true,
        selectable: true,
      },
      global: { plugins: [createPinia()] },
    })

    const area = screen.getByTestId('player-area-p1')
    expect(area.classList.contains('player-area--me')).toBe(true)
    expect(area.classList.contains('player-area--active')).toBe(true)
    expect(area.classList.contains('player-area--targetable')).toBe(true)
    expect(screen.getByText('你')).toBeTruthy()
    expect(screen.getByText('当前行动')).toBeTruthy()
    expect(screen.getByText('可选')).toBeTruthy()
  })

  it('limits always-visible status effect icons to the first three', () => {
    const { container } = render(PlayerArea, {
      props: {
        player: buildPlayer({
          field: [
            buildEffectField('Shield'),
            buildEffectField('Poison'),
            buildEffectField('Weak'),
            buildEffectField('Stealth'),
          ],
        }),
      },
      global: { plugins: [createPinia()] },
    })

    expect(container.querySelectorAll('.effect-icon-item')).toHaveLength(3)
  })

  it('renders a persistent effect as a glowing seat chip with its rule as tooltip (额外回合)', () => {
    render(PlayerArea, {
      props: { player: buildPlayer({ role: 'moon_goddess', indicators: { mg_extra_turn_pending: 1 } }) },
    })
    const chip = screen.getByTestId('seat-persistent-effect')
    expect(chip.textContent).toContain('额外回合')
    expect(chip.getAttribute('title')).toContain('苍白之月')
  })

  it('keeps a form aura and badge on the seat for as long as the player stays in the form', async () => {
    const { rerender, queryByTestId, getByTestId } = render(PlayerArea, {
      props: { player: buildPlayer({ role: 'holy_bow', form: 'holy_bow_holy_glory_form' }) },
    })
    expect(getByTestId('seat-form-badge').textContent).toContain('圣煌形态')
    expect(getByTestId('seat-form-aura')).toBeTruthy()

    await rerender({ player: buildPlayer({ role: 'holy_bow', form: '' }) })
    expect(queryByTestId('seat-form-aura')).toBeNull()
    expect(queryByTestId('seat-form-badge')).toBeNull()
  })
})
