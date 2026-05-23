import { render, screen, waitFor } from '@testing-library/vue'
import userEvent from '@testing-library/user-event'
import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import ActionPanel from '../ActionPanel.vue'
import { useInterruptStore } from '../../stores/interrupt.store'
import { useSessionStore } from '../../stores/session.store'
import { useSnapshotStore } from '../../stores/snapshot.store'
import type { AvailableSkill, CharacterView, GameStateUpdate, PlayerView } from '../../types/game'

const mocks = vi.hoisted(() => ({
  submitUseSkill: vi.fn(),
}))

vi.mock('../../composables/useSubmitAction', () => ({
  useSubmitAction: () => ({
    submitAttack: vi.fn(),
    submitMagic: vi.fn(),
    submitUseSkill: mocks.submitUseSkill,
    submitPass: vi.fn(),
    submitCannotAct: vi.fn(),
    submitBuy: vi.fn(),
    submitSynthesize: vi.fn(),
    submitExtract: vi.fn(),
    cheatSkill: vi.fn(),
    cheatToken: vi.fn(),
    cheatSet: vi.fn(),
    cheatEffect: vi.fn(),
    cheatGiveExclusive: vi.fn(),
    cheatGiveByElement: vi.fn(),
    cheatGiveByFaction: vi.fn(),
    cheatGiveMagicByName: vi.fn(),
    cheatDiscard: vi.fn(),
  }),
}))

function buildPlayer(overrides: Partial<PlayerView> = {}): PlayerView {
  return {
    id: 'p1',
    name: '血之巫女',
    camp: 'Red',
    role: 'blood_priestess',
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
    is_active: true,
    buffs: [],
    tokens: {},
    indicators: {},
    ...overrides,
  }
}

function buildCharacter(overrides: Partial<CharacterView> = {}): CharacterView {
  return {
    id: 'blood_priestess',
    name: '血之巫女',
    title: '',
    faction: '血',
    skills: [
      {
        id: 'bp_shared_life',
        title: '同生共死',
        description: '将【同生共死】放置于目标角色面前。',
        type: 2,
        min_targets: 0,
        max_targets: 0,
        target_type: 0,
        cost_gem: 0,
        cost_crystal: 0,
        cost_discards: 0,
        require_exclusive: true,
      },
    ],
    ...overrides,
  }
}

function buildState(overrides: Partial<GameStateUpdate> = {}): GameStateUpdate {
  return {
    turn_stage: 'ActionExecution',
    current_player: 'p1',
    has_performed_startup: false,
    players: {
      p1: buildPlayer(),
    },
    red_morale: 15,
    blue_morale: 15,
    red_cups: 0,
    blue_cups: 0,
    red_gems: 0,
    blue_gems: 0,
    red_crystals: 0,
    blue_crystals: 0,
    deck_count: 30,
    discard_count: 0,
    available_skills: [],
    characters: [buildCharacter()],
    ...overrides,
  }
}

const ELEMENTALIST_EXCLUSIVE_PROMPT_SKILLS = [
  {
    id: 'elementalist_thunder_strike',
    title: '雷击',
    description: '独有技法术：可额外弃1张雷系牌。',
  },
  {
    id: 'elementalist_freeze',
    title: '冰冻',
    description: '独有技法术：可额外弃1张水系牌。',
  },
  {
    id: 'elementalist_wind_blade',
    title: '风刃',
    description: '独有技法术：可额外弃1张风系牌。',
  },
  {
    id: 'elementalist_meteor',
    title: '陨石',
    description: '独有技法术：可额外弃1张地系牌。',
  },
  {
    id: 'elementalist_fireball',
    title: '火球',
    description: '独有技法术：可额外弃1张火系牌。',
  },
] as const

function buildElementalistExclusiveSkill(skill: typeof ELEMENTALIST_EXCLUSIVE_PROMPT_SKILLS[number]): AvailableSkill {
  return {
    id: skill.id,
    title: skill.title,
    description: skill.description,
    min_targets: 2,
    max_targets: 2,
    target_type: 5,
    cost_gem: 0,
    cost_crystal: 0,
    cost_discards: 1,
    require_exclusive: true,
  }
}

describe('ActionPanel skill availability', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    mocks.submitUseSkill.mockReset()
  })

  it('disables shared life when its exclusive card is already away from the player', () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    useSessionStore().setRoomInfo('ROOM', 'p1', 'Red', 'blood_priestess')
    useSnapshotStore().updateGameState(buildState())
    useInterruptStore().setSkillMode('choosing_skill')

    render(ActionPanel, {
      global: {
        plugins: [pinia],
        stubs: {
          PromptDialog: true,
        },
      },
    })

    const sharedLifeButton = screen.getByTestId('skill-bp_shared_life')
    expect(sharedLifeButton).toBeDisabled()
    expect(sharedLifeButton).toHaveTextContent('缺少可用于发动的「同生共死」独有技手牌')
  })

  it('submits server-published targeted skills before backend prompt-driven target selection', async () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    const arcaneCodexSkill: AvailableSkill = {
      id: 'sage_arcane_codex',
      title: '魔道法典',
      description: '［宝石］弃X张异系牌（X>1），对目标角色与自己各造成(X-1)点法术伤害。',
      min_targets: 1,
      max_targets: 1,
      target_type: 5,
      cost_gem: 1,
      cost_crystal: 0,
      cost_discards: 0,
    }
    const sageCharacter = buildCharacter({
      id: 'sage',
      name: '贤者',
      skills: [
        {
          ...arcaneCodexSkill,
          type: 2,
        },
      ],
    })

    useSessionStore().setRoomInfo('ROOM', 'p1', 'Red', 'sage')
    useSnapshotStore().updateGameState(buildState({
      players: {
        p1: buildPlayer({ id: 'p1', role: 'sage', gem: 1 }),
        p2: buildPlayer({ id: 'p2', name: '目标', camp: 'Blue', role: 'enemy', is_active: false }),
      },
      available_skills: [arcaneCodexSkill],
      characters: [sageCharacter],
    }))
    useInterruptStore().setSkillMode('choosing_skill')

    render(ActionPanel, {
      global: {
        plugins: [pinia],
        stubs: {
          PromptDialog: true,
        },
      },
    })

    await userEvent.click(screen.getByTestId('skill-sage_arcane_codex'))

    expect(mocks.submitUseSkill).toHaveBeenCalledWith(
      'sage_arcane_codex',
      [],
      undefined,
      { clearSkillMode: true },
    )
    expect(useInterruptStore().skillMode).not.toBe('choosing_target')
  })

  it('starts adventurer fraud without choosing a target first', async () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    const fraudSkill: AvailableSkill = {
      id: 'adventurer_fraud',
      title: '欺诈',
      description: '主动技能：选择1名敌方角色，弃同系牌将本次视为一次主动攻击。',
      min_targets: 0,
      max_targets: 0,
      target_type: 0,
      cost_gem: 0,
      cost_crystal: 0,
      cost_discards: 0,
    }
    const adventurerCharacter = buildCharacter({
      id: 'adventurer',
      name: '冒险家',
      skills: [
        {
          ...fraudSkill,
          type: 2,
        },
      ],
    })

    useSessionStore().setRoomInfo('ROOM', 'p1', 'Red', 'adventurer')
    useSnapshotStore().updateGameState(buildState({
      players: {
        p1: buildPlayer({ id: 'p1', name: '冒险家', role: 'adventurer' }),
        p2: buildPlayer({ id: 'p2', name: '目标', camp: 'Blue', role: 'enemy', is_active: false }),
      },
      available_skills: [fraudSkill],
      characters: [adventurerCharacter],
    }))
    useInterruptStore().setSkillMode('choosing_skill')

    render(ActionPanel, {
      global: {
        plugins: [pinia],
        stubs: {
          PromptDialog: true,
        },
      },
    })

    await userEvent.click(screen.getByTestId('skill-adventurer_fraud'))

    expect(mocks.submitUseSkill).toHaveBeenCalledWith(
      'adventurer_fraud',
      [],
      undefined,
      { clearSkillMode: true },
    )
    expect(useInterruptStore().skillMode).not.toBe('choosing_target')
  })

  it('requires exclusive-card confirmation before target selection for prayer blessing skills', async () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    const prayerPowerBlessingSkill: AvailableSkill = {
      id: 'prayer_power_blessing',
      title: '威力赐福',
      description: '将独有技手牌当法术牌打出并放置于1名队友面前。',
      min_targets: 1,
      max_targets: 1,
      target_type: 3,
      cost_gem: 0,
      cost_crystal: 0,
      cost_discards: 0,
      require_exclusive: true,
      place_card: true,
      place_effect: 'PowerBlessing',
    }
    const prayerCharacter = buildCharacter({
      id: 'prayer_master',
      name: '祈祷师',
      skills: [
        {
          ...prayerPowerBlessingSkill,
          type: 2,
        },
      ],
    })

    useSessionStore().setRoomInfo('ROOM', 'p1', 'Red', 'prayer_master')
    useSnapshotStore().updateGameState(buildState({
      players: {
        p1: buildPlayer({
          id: 'p1',
          name: '祈祷师',
          role: 'prayer_master',
          hand_count: 1,
          hand: [
            {
              id: 'hand-p1-prayer_power_blessing',
              name: '威力赐福',
              type: 'Magic',
              element: 'Light',
              damage: 0,
              description: '独有技卡',
              exclusive_char1: 'prayer_master',
              exclusive_skill1: '威力赐福',
            },
          ],
        }),
        p2: buildPlayer({ id: 'p2', name: '队友', camp: 'Red', role: 'angel', is_active: false }),
      },
      available_skills: [prayerPowerBlessingSkill],
      characters: [prayerCharacter],
    }))
    useInterruptStore().setSkillMode('choosing_skill')

    render(ActionPanel, {
      global: {
        plugins: [pinia],
        stubs: {
          PromptDialog: true,
        },
      },
    })

    await userEvent.click(screen.getByTestId('skill-prayer_power_blessing'))

    expect(mocks.submitUseSkill).not.toHaveBeenCalled()
    expect(useInterruptStore().skillMode).toBe('choosing_exclusive')
    expect(screen.getByTestId('skill-exclusive-select-panel')).toBeInTheDocument()
    expect(screen.getByTestId('skill-exclusive-confirm-btn')).toBeDisabled()

    useInterruptStore().setSkillDiscardHandIndexes([0])
    await waitFor(() => {
      expect(screen.getByTestId('skill-exclusive-confirm-btn')).not.toBeDisabled()
    })

    await userEvent.click(screen.getByTestId('skill-exclusive-confirm-btn'))

    expect(useInterruptStore().skillMode).toBe('choosing_target')
    expect(useInterruptStore().selectedSkill?.id).toBe('prayer_power_blessing')
  })

  it('keeps all elementalist exclusive prompt-flow skills on the exclusive-card confirmation panel before handing off to the backend prompt flow', async () => {
    for (const skillDef of ELEMENTALIST_EXCLUSIVE_PROMPT_SKILLS) {
      const pinia = createPinia()
      setActivePinia(pinia)
      mocks.submitUseSkill.mockReset()

      const skill = buildElementalistExclusiveSkill(skillDef)
      const elementalistCharacter = buildCharacter({
        id: 'elementalist',
        name: '元素师',
        skills: [
          {
            ...skill,
            type: 2,
          },
        ],
      })

      useSessionStore().setRoomInfo('ROOM', 'p1', 'Red', 'elementalist')
      useSnapshotStore().updateGameState(buildState({
        players: {
          p1: buildPlayer({
            id: 'p1',
            name: '元素师',
            role: 'elementalist',
            hand_count: 1,
            hand: [
              {
                id: `hand-p1-${skill.id}`,
                name: skill.title,
                type: 'Magic',
                element: 'Water',
                damage: 0,
                description: '独有技卡',
                exclusive_char1: 'elementalist',
                exclusive_skill1: skill.title,
              },
            ],
          }),
        },
        available_skills: [skill],
        characters: [elementalistCharacter],
      }))
      useInterruptStore().setSkillMode('choosing_skill')

      const { unmount } = render(ActionPanel, {
        global: {
          plugins: [pinia],
          stubs: {
            PromptDialog: true,
          },
        },
      })

      await userEvent.click(screen.getByTestId(`skill-${skill.id}`))

      expect(mocks.submitUseSkill).not.toHaveBeenCalled()
      expect(useInterruptStore().skillMode).toBe('choosing_exclusive')
      expect(screen.getByTestId('skill-exclusive-select-panel')).toBeInTheDocument()
      expect(screen.getByTestId('skill-exclusive-confirm-btn')).toBeDisabled()

      useInterruptStore().setSkillDiscardHandIndexes([0])
      await waitFor(() => {
        expect(screen.getByTestId('skill-exclusive-confirm-btn')).not.toBeDisabled()
      })

      await userEvent.click(screen.getByTestId('skill-exclusive-confirm-btn'))

      expect(mocks.submitUseSkill).toHaveBeenCalledWith(
        skill.id,
        [],
        [0],
        { clearSkillMode: true },
      )
      unmount()
    }
  })
})
