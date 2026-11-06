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

  it('does not show the main action hub while reconnecting into a combat response snapshot', () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    useSessionStore().setRoomInfo('ROOM', 'p1', 'Red', 'blood_priestess')
    useSnapshotStore().updateGameState(buildState({
      combat_stage: 'CombatHitCheck',
      subflow: '',
      current_player: 'p1',
    }))

    render(ActionPanel, {
      global: {
        plugins: [pinia],
        stubs: {
          PromptDialog: true,
        },
      },
    })

    expect(screen.queryByTestId('action-attack')).not.toBeInTheDocument()
    expect(screen.queryByTestId('action-magic')).not.toBeInTheDocument()
    expect(screen.queryByTestId('action-special')).not.toBeInTheDocument()
  })

  it('does not show the main action hub while reconnecting into a response subflow', () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    useSessionStore().setRoomInfo('ROOM', 'p1', 'Red', 'blood_priestess')
    useSnapshotStore().updateGameState(buildState({
      combat_stage: '',
      subflow: 'Response',
      current_player: 'p1',
    }))

    render(ActionPanel, {
      global: {
        plugins: [pinia],
        stubs: {
          PromptDialog: true,
        },
      },
    })

    expect(screen.queryByTestId('action-attack')).not.toBeInTheDocument()
    expect(screen.queryByTestId('action-magic')).not.toBeInTheDocument()
    expect(screen.queryByTestId('action-special')).not.toBeInTheDocument()
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
      undefined
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
      undefined
    )
    expect(useInterruptStore().skillMode).not.toBe('choosing_target')
  })

  it('greys out a skill the server marks disabled and shows its reason instead of erroring on click', async () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    const stormSkill: AvailableSkill = {
      id: 'hb_holy_shard_storm',
      title: '圣屑飓暴',
      description: '弃2张同系攻击牌并展示，视为1次圣命格同系主动攻击。',
      min_targets: 0,
      max_targets: 0,
      target_type: 0,
      cost_gem: 0,
      cost_crystal: 0,
      cost_discards: 0,
      disabled: true,
      disabled_reason: '手牌中没有2张同系攻击牌。',
    }
    useSessionStore().setRoomInfo('ROOM', 'p1', 'Blue', 'holy_bow')
    useSnapshotStore().updateGameState(buildState({
      players: {
        p1: buildPlayer({ id: 'p1', name: '圣弓', camp: 'Blue', role: 'holy_bow' }),
        p2: buildPlayer({ id: 'p2', name: '目标', camp: 'Red', role: 'enemy', is_active: false }),
      },
      available_skills: [stormSkill],
      characters: [buildCharacter({ id: 'holy_bow', name: '圣弓', skills: [{ ...stormSkill, type: 2 }] })],
    }))
    useInterruptStore().setSkillMode('choosing_skill')

    render(ActionPanel, { global: { plugins: [pinia], stubs: { PromptDialog: true } } })

    expect(screen.getByText('手牌中没有2张同系攻击牌。')).toBeTruthy()
    await userEvent.click(screen.getByTestId('skill-hb_holy_shard_storm'))
    expect(mocks.submitUseSkill).not.toHaveBeenCalled()
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
        [0]
      )
      unmount()
    }
  })
})

describe('ActionPanel action hub notice and end-turn', () => {
  function hubPrompt(message: string, optionIds: string[] = ['attack', 'magic', 'cannot_act']) {
    return {
      type: 'confirm',
      player_id: 'p1',
      choice_type: 'action_hub',
      message,
      options: optionIds.map((id) => ({ id, label: id, button_label: id })),
      min: 1,
      max: 1,
      presentation: { kind: 'action_hub', numeric_base: 0 },
    } as unknown as Parameters<ReturnType<typeof useInterruptStore>['setPrompt']>[0]
  }

  function renderHub(message: string, optionIds?: string[]) {
    const pinia = createPinia()
    setActivePinia(pinia)
    useSessionStore().setRoomInfo('ROOM', 'p1', 'Red', 'blood_priestess')
    useSnapshotStore().updateGameState(buildState())
    useInterruptStore().setPrompt(hubPrompt(message, optionIds))
    return render(ActionPanel, { global: { plugins: [pinia], stubs: { PromptDialog: true } } })
  }

  it('hides the default "请选择行动类型" notice', () => {
    renderHub('请选择行动类型')
    expect(screen.getByTestId('action-attack')).toBeInTheDocument()
    expect(screen.queryByTestId('action-hub-notice')).toBeNull()
  })

  it('shows rule constraints as a warning without the trailing default sentence', () => {
    renderHub('你的审判已达上限：本行动阶段必须发动【末日审判】。')
    const notice = screen.getByTestId('action-hub-notice')
    expect(notice).toHaveTextContent('必须发动【末日审判】')
    expect(notice.className).toContain('dock-notice--rule')
  })

  it('moves the 充盈 bonus into a footer chip', () => {
    renderHub('请选择行动类型（【充盈】下一次主动攻击伤害额外+2）')
    expect(screen.queryByTestId('action-hub-notice')).toBeNull()
    expect(screen.getByTestId('action-hub-fullness')).toHaveTextContent('+2')
  })

  it('never renders an end-turn button; cannot-act sits in the footer', () => {
    renderHub('请选择行动类型')
    expect(screen.queryByTestId('action-pass')).toBeNull()
    expect(screen.queryByText('结束回合')).toBeNull()
    const cannotAct = screen.getByTestId('action-cannot-act')
    expect(cannotAct.closest('footer')).not.toBeNull()
  })

  it('labels cannot-act as giving up the extra action during an extra action', () => {
    renderHub('当前为额外攻击行动，仅可执行攻击。请选择行动类型', ['attack', 'cannot_act'])
    expect(screen.getByTestId('action-cannot-act')).toHaveTextContent('放弃额外行动')
    expect(screen.queryByTestId('action-hub-notice')).toBeNull()
    expect(screen.getByText('仅可攻击')).toBeInTheDocument()
  })
})

describe('ActionPanel skill step confirmation', () => {
  const discardSkill: AvailableSkill = {
    id: 'test_discard_target',
    title: '弃牌打击',
    description: '弃1张牌，对1名目标造成伤害。',
    min_targets: 1,
    max_targets: 1,
    target_type: 2,
    cost_gem: 0,
    cost_crystal: 0,
    cost_discards: 1,
  }

  function setup(skill: AvailableSkill) {
    const pinia = createPinia()
    setActivePinia(pinia)
    mocks.submitUseSkill.mockReset()
    useSessionStore().setRoomInfo('ROOM', 'p1', 'Red', 'hero')
    useSnapshotStore().updateGameState(buildState({
      players: {
        p1: buildPlayer({
          id: 'p1',
          role: 'hero',
          hand_count: 2,
          hand: [
            { id: 'c1', name: '火焰斩', type: 'Attack', element: 'Fire', damage: 2, description: '' },
            { id: 'c2', name: '水涟斩', type: 'Attack', element: 'Water', damage: 2, description: '' },
          ],
        }),
        p2: buildPlayer({ id: 'p2', name: '目标', camp: 'Blue', role: 'enemy', is_active: false }),
      },
      available_skills: [skill],
      characters: [buildCharacter({ id: 'hero', name: '英雄', skills: [{ ...skill, type: 2 }] })],
    }))
    const store = useInterruptStore()
    store.setSelectedSkill(skill)
    store.setSkillMode('choosing_discard')
    render(ActionPanel, { global: { plugins: [pinia], stubs: { PromptDialog: true } } })
    return store
  }

  it('waits for "下一步" after the discard is full, and "上一步" keeps the discard', async () => {
    const store = setup(discardSkill)
    const next = screen.getByTestId('skill-discard-confirm-btn')
    expect(next).toBeDisabled()
    expect(next).toHaveTextContent('下一步')

    store.setSkillDiscardHandIndexes([1])
    await waitFor(() => expect(screen.getByTestId('skill-discard-confirm-btn')).not.toBeDisabled())
    expect(store.skillMode).toBe('choosing_discard')

    await userEvent.click(screen.getByTestId('skill-discard-confirm-btn'))
    expect(store.skillMode).toBe('choosing_target')

    await userEvent.click(screen.getByTestId('skill-target-back-btn'))
    expect(store.skillMode).toBe('choosing_discard')
    expect(store.skillDiscardHandIndexes).toEqual([1])
  })

  it('submits a targeted skill only through "确认发动"', async () => {
    const store = setup(discardSkill)
    store.setSkillDiscardHandIndexes([0])
    store.setSkillMode('choosing_target')
    await waitFor(() => expect(screen.getByTestId('skill-target-confirm-btn')).toBeDisabled())

    store.setSkillTargetIds(['p2'])
    await waitFor(() => expect(screen.getByTestId('skill-target-confirm-btn')).not.toBeDisabled())
    expect(mocks.submitUseSkill).not.toHaveBeenCalled()

    await userEvent.click(screen.getByTestId('skill-target-confirm-btn'))
    expect(mocks.submitUseSkill).toHaveBeenCalledWith('test_discard_target', ['p2'], [0])
  })

  it('labels the discard step "确认发动" and hides the target step for self-target skills', async () => {
    const store = setup({ ...discardSkill, id: 'test_discard_self', target_type: 1, min_targets: 0, max_targets: 0 })
    expect(screen.queryByText('② 选目标')).toBeNull()
    expect(screen.getByTestId('skill-discard-confirm-btn')).toHaveTextContent('确认发动')

    store.setSkillDiscardHandIndexes([0])
    await waitFor(() => expect(screen.getByTestId('skill-discard-confirm-btn')).not.toBeDisabled())
    await userEvent.click(screen.getByTestId('skill-discard-confirm-btn'))
    expect(mocks.submitUseSkill).toHaveBeenCalledWith('test_discard_self', ['p1'], [0])
  })

  it('hints which target will be replaced once a multi-target skill is full', async () => {
    const store = setup({ ...discardSkill, id: 'test_multi', cost_discards: 0, min_targets: 1, max_targets: 2 })
    store.setSkillMode('choosing_target')
    store.setSkillTargetIds(['p2', 'p1'])
    await waitFor(() => expect(screen.getByTestId('skill-target-replace-hint')).toHaveTextContent('目标'))
  })
})

describe('ActionPanel pending states', () => {
  function renderIdleTurn() {
    const pinia = createPinia()
    setActivePinia(pinia)
    useSessionStore().setRoomInfo('ROOM', 'p1', 'Red', 'hero')
    useSnapshotStore().updateGameState(buildState())
    render(ActionPanel, { global: { plugins: [pinia], stubs: { PromptDialog: true } } })
    return useInterruptStore()
  }

  it('hides the action hub while a submission awaits the server', async () => {
    const store = renderIdleTurn()
    expect(screen.getByTestId('action-attack')).toBeInTheDocument()
    store.beginSubmit()
    await waitFor(() => expect(screen.queryByTestId('action-attack')).toBeNull())
    expect(screen.getByTestId('dock-pending-panel')).toHaveTextContent('等待服务器结算')
    store.endSubmit()
    await waitFor(() => expect(screen.getByTestId('action-attack')).toBeInTheDocument())
  })

  it('shows a light "即将轮到你" status while my prompt is gated', async () => {
    const store = renderIdleTurn()
    store.setPromptGated(true)
    await waitFor(() => expect(screen.getByTestId('dock-pending-panel')).toHaveTextContent('战斗结算中，即将轮到你操作'))
    expect(screen.queryByTestId('action-attack')).toBeNull()
  })
})
