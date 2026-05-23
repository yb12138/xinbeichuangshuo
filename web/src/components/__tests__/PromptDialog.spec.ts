import { render, screen } from '@testing-library/vue'
import userEvent from '@testing-library/user-event'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { nextTick } from 'vue'
import PromptDialog from '../PromptDialog.vue'
import { useInterruptStore } from '../../stores/interrupt.store'
import { useSessionStore } from '../../stores/session.store'
import { useSnapshotStore } from '../../stores/snapshot.store'
import type { Card, GameStateUpdate, PlayerView, Prompt } from '../../types/game'

const submitSelectMock = vi.fn()
const submitCancelMock = vi.fn()
const submitRespondTakeMock = vi.fn()
const submitRespondCounterMock = vi.fn()
const submitRespondDefendMock = vi.fn()
const submitSelectCardIDsMock = vi.fn()

vi.mock('../../composables/useSubmitAction', () => ({
  useSubmitAction: () => ({
    submitSelect: submitSelectMock,
    submitCancel: submitCancelMock,
    submitConfirm: vi.fn(),
    submitRespondTake: submitRespondTakeMock,
    submitRespondCounter: submitRespondCounterMock,
    submitRespondDefend: submitRespondDefendMock,
    submitSelectCardIDs: submitSelectCardIDsMock,
    submitAction: vi.fn(),
  }),
}))

function buildCard(overrides: Partial<Card> = {}): Card {
  return {
    id: 'card-1',
    name: '烈焰斩',
    type: 'Attack',
    element: 'Fire',
    damage: 2,
    description: 'test card',
    ...overrides,
  }
}

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
    ...overrides,
  }
}

function buildState(overrides: Partial<GameStateUpdate> = {}): GameStateUpdate {
  return {
    turn_stage: 'ActionExecution',
    current_player: 'p3',
    has_performed_startup: false,
    players: {
      p2: buildPlayer({
        id: 'p2',
        name: 'P2',
        camp: 'Blue',
        heal: 2,
        max_heal: 5,
        hand_count: 3,
        hand: [
          buildCard({ id: 'h0' }),
          buildCard({ id: 'h1', element: 'Water' }),
          buildCard({ id: 'h2', element: 'Thunder' }),
        ],
      }),
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
    ...overrides,
  }
}

function medusaDarkMoonPickPrompt(): Prompt {
  return {
    type: 'confirm',
    player_id: 'p2',
    message: '【美杜莎之眼】请选择要展示并移除的同系闇月：',
    options: [
      { id: '0', label: '移除闇月[暗月法术/Magic/Dark]', button_label: '移除闇月[0]', field_index: 0 },
      { id: '1', label: '移除闇月[火焰斩/Attack/Fire]', button_label: '移除闇月[1]', field_index: 1 },
    ],
    min: 1,
    max: 1,
    presentation: {
      kind: 'card_picker',
      layout: 'field_cover',
      numeric_base: 0,
      card_source: 'field',
      card_filter: 'effect:MoonDarkMoon',
    },
  }
}

function moonCycleBranchPrompt(): Prompt {
  return {
    type: 'confirm',
    player_id: 'p2',
    choice_type: 'mg_moon_cycle_mode',
    message: '【月之轮回】请选择发动分支：',
    options: [
      { id: 'decline', label: '不发动', button_label: '不发动' },
      { id: 'branch1', label: '分支①：移除1个闇月，令目标角色+1治疗', button_label: '分支①：移除1个闇月，令目标角色+1治疗' },
      { id: 'branch2', label: '分支②：移除1点治疗，你+1新月', button_label: '分支②：移除1点治疗，你+1新月' },
    ],
    min: 1,
    max: 1,
    presentation: {
      kind: 'branch_select',
      layout: 'overlay',
      numeric_base: 0,
      cancel_policy: 'decline',
      has_decline: true,
      decline_index: 0,
    },
  }
}

function weakPrompt(): Prompt {
  return {
    type: 'confirm',
    player_id: 'p2',
    choice_type: 'weak',
    message: '【虚弱状态】测试玩家1，你需要做出选择：',
    options: [
      { id: 'draw_continue', label: '摸3张牌继续执行后续行动', button_label: '摸3张牌继续执行后续行动' },
      { id: 'skip_turn', label: '跳过此回合', button_label: '跳过此回合' },
    ],
    min: 1,
    max: 1,
    presentation: {
      kind: 'branch_select',
      layout: 'overlay',
      numeric_base: 0,
    },
  }
}

function magicBulletDirectionPrompt(): Prompt {
  return {
    type: 'confirm',
    player_id: 'p2',
    message: '【魔弹掌控】请选择方向：',
    options: [
      { id: 'normal', label: '正向', button_label: '正向' },
      { id: 'reverse', label: '逆向', button_label: '逆向' },
    ],
    min: 1,
    max: 1,
    presentation: {
      kind: 'branch_select',
      layout: 'overlay',
      numeric_base: 0,
    },
  }
}

function radiantCannonDirectionPrompt(): Prompt {
  return {
    type: 'confirm',
    player_id: 'p2',
    choice_type: 'hb_radiant_cannon_side',
    message: '【圣煌辉光炮】请选择士气对齐方向：',
    options: [
      { id: 'red', label: '红方士气', button_label: '红方', hint: '对齐到红方' },
      { id: 'blue', label: '蓝方士气', button_label: '蓝方', hint: '对齐到蓝方' },
    ],
    min: 1,
    max: 1,
    presentation: {
      kind: 'branch_select',
      layout: 'overlay',
      numeric_base: 0,
    },
  }
}

function moonCycleTargetPrompt(): Prompt {
  return {
    type: 'confirm',
    player_id: 'p2',
    choice_type: 'mg_moon_cycle_heal_target',
    message: '【月之轮回】请选择获得1点治疗的角色：',
    options: [
      { id: '0', target_id: 'p3', label: '目标玩家', button_label: '目标玩家' },
    ],
    min: 1,
    max: 1,
    presentation: {
      kind: 'target_picker',
      target_filter: 'custom',
      numeric_base: 0,
    },
  }
}

function healPrompt(): Prompt {
  return {
    type: 'confirm',
    player_id: 'p2',
    choice_type: 'heal',
    message: 'P2 受到伤害，可选择使用治疗抵消：',
    options: [
      { id: '0', label: '不使用治疗', button_label: '0' },
      { id: '1', label: '使用 1 点治疗', button_label: '1' },
      { id: '2', label: '使用 2 点治疗', button_label: '2' },
    ],
    min: 1,
    max: 1,
    presentation: {
      kind: 'numeric',
      numeric_base: 0,
    },
  }
}

function saintHealAllocatePrompt(): Prompt {
  return {
    type: 'confirm',
    player_id: 'p2',
    message: '【圣疗】请分配 3 点治疗：',
    options: [
      { id: '0', label: '治疗目标一', button_label: '治疗目标一', target_id: 'p2' },
      { id: '1', label: '治疗目标二', button_label: '治疗目标二', target_id: 'p3' },
    ],
    min: 1,
    max: 1,
    presentation: {
      kind: 'branch_select',
      layout: 'heal_allocate',
      numeric_base: 0,
    },
  }
}

function runeReforgeAllocatePrompt(): Prompt {
  return {
    type: 'confirm',
    player_id: 'p2',
    message: '【符文改造】请分配战纹/魔纹：',
    options: [
      { id: 'war', label: '战纹', button_label: '战纹' },
      { id: 'magic', label: '魔纹', button_label: '魔纹' },
    ],
    min: 1,
    max: 1,
    presentation: {
      kind: 'branch_select',
      layout: 'rune_allocate',
      numeric_base: 0,
    },
  }
}

function bloodyPrayerAllocatePrompt(): Prompt {
  return {
    type: 'confirm',
    player_id: 'p2',
    choice_type: 'crk_bloody_prayer_split',
    message: '【血腥祷言】请选择治疗分配：所有人加起来的治疗点数必须等于 3',
    options: [
      { id: 'p2', label: '治疗目标一（治疗:2）', button_label: '治疗目标一', target_id: 'p2' },
      { id: 'p3', label: '治疗目标二（治疗:1）', button_label: '治疗目标二', target_id: 'p3' },
    ],
    min: 2,
    max: 2,
    presentation: {
      kind: 'numeric',
      layout: 'blood_prayer_allocate',
      numeric_base: 0,
    },
  }
}

function fraudAttackElementPrompt(): Prompt {
  return {
    type: 'confirm',
    player_id: 'p2',
    message: '【欺诈】请选择本次攻击系别：',
    options: [
      { id: 'water', label: '水', button_label: '水' },
      { id: 'fire', label: '火', button_label: '火' },
      { id: 'earth', label: '地', button_label: '地' },
      { id: 'wind', label: '风', button_label: '风' },
      { id: 'thunder', label: '雷', button_label: '雷' },
    ],
    min: 1,
    max: 1,
    presentation: {
      kind: 'branch_select',
      layout: 'fraud_attack_element',
      numeric_base: 0,
    },
  }
}

function responsePrompt(): Prompt {
  return {
    type: 'confirm',
    player_id: 'p2',
    message: '请选择响应方式',
    options: [
      { id: 'counter', label: '应战', button_label: '应战' },
      { id: 'take', label: '承受命中', button_label: '承受命中' },
      { id: 'defend', label: '防御', button_label: '防御' },
    ],
    min: 1,
    max: 1,
    attack_element: 'Fire',
    presentation: {
      kind: 'response',
      numeric_base: 0,
    },
  }
}

function handCardPickerPrompt(overrides: Partial<Prompt> = {}): Prompt {
  return {
    type: 'confirm',
    player_id: 'p2',
    message: '请选择要发动的手牌',
    options: [
      { id: 'pick-h0', label: '选择火焰斩', button_label: '选择火焰斩', card_id: 'h0' },
      { id: 'pick-h1', label: '选择水涟斩', button_label: '选择水涟斩', card_id: 'h1' },
    ],
    min: 1,
    max: 1,
    presentation: {
      kind: 'card_picker',
      card_source: 'hand',
      numeric_base: 0,
    },
    ...overrides,
  }
}

function handCardPickerWithDeclinePrompt(): Prompt {
  return {
    type: 'confirm',
    player_id: 'p2',
    message: '请选择是否发动该卡牌效果',
    options: [
      { id: 'decline', label: '取消', button_label: '取消' },
      { id: 'pick-h0', label: '选择火焰斩', button_label: '选择火焰斩', card_id: 'h0' },
    ],
    min: 0,
    max: 1,
    presentation: {
      kind: 'card_picker',
      card_source: 'hand',
      numeric_base: 0,
      cancel_policy: 'decline',
      has_decline: true,
      decline_index: 0,
    },
  }
}

function extractPrompt(): Prompt {
  return {
    type: 'confirm',
    player_id: 'p2',
    message: '请选择提炼目标',
    options: [
      { id: 'ruby', label: '红宝石', button_label: '红宝石' },
      { id: 'crystal', label: '蓝水晶', button_label: '蓝水晶' },
      { id: 'extra', label: '红宝石', button_label: '红宝石' },
    ],
    min: 1,
    max: 2,
    presentation: {
      kind: 'branch_select',
      layout: 'extract',
      numeric_base: 0,
    },
  }
}

function singleSkillChoicePrompt(): Prompt {
  return {
    type: 'choose_skill',
    player_id: 'p2',
    message: '请选择是否发动技能',
    options: [
      { id: 'fire_blast', label: '炎爆术', button_label: '发动' },
    ],
    min: 1,
    max: 1,
    presentation: {
      kind: 'skill_choice',
      numeric_base: 0,
    },
  }
}

function multiSkillChoicePrompt(): Prompt {
  return {
    type: 'choose_skill',
    player_id: 'p2',
    message: '请选择要发动的技能',
    options: [
      { id: 'skill_a', label: '烈焰突刺[消耗1]', button_label: '发动', hint: '造成2点火焰伤害' },
      { id: 'skill_b', label: '冰封结界[消耗2]', button_label: '发动', hint: '获得1层护盾' },
      { id: 'skill_c', label: '雷鸣冲击[消耗3]', button_label: '发动', hint: '随机打击2次' },
    ],
    min: 1,
    max: 1,
    presentation: {
      kind: 'skill_choice',
      numeric_base: 0,
      cancel_policy: 'decline',
    },
  }
}

function colonLabelSkillChoicePrompt(): Prompt {
  return {
    type: 'choose_skill',
    player_id: 'p2',
    message: '请选择要发动的技能',
    options: [
      { id: 'skill_a', label: '烈焰突刺：造成2点火焰伤害', button_label: '发动' },
      { id: 'skill_b', label: '冰封结界', button_label: '发动' },
    ],
    min: 1,
    max: 1,
    presentation: {
      kind: 'skill_choice',
      numeric_base: 0,
      cancel_policy: 'decline',
    },
  }
}

describe('PromptDialog', () => {
  beforeEach(() => {
    submitSelectMock.mockReset()
    submitCancelMock.mockReset()
    submitRespondTakeMock.mockReset()
    submitRespondCounterMock.mockReset()
    submitRespondDefendMock.mockReset()
    submitSelectCardIDsMock.mockReset()
  })

  it('shows full weakness labels from presentation', async () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    useSessionStore().setRoomInfo('ROOM1', 'p2', 'Blue', 'fighter')
    useSnapshotStore().updateGameState(buildState({
      players: {
        p2: buildPlayer({
          id: 'p2',
          name: '测试玩家1',
          camp: 'Blue',
          role: 'fighter',
          hand: [buildCard({ id: 'c1', name: '火球', type: 'Magic', element: 'Fire' })],
          field: [],
          is_active: true,
          buffs: [],
          tokens: {},
        }),
      },
    }))
    useInterruptStore().setPrompt({
      type: 'confirm',
      player_id: 'p2',
      message: '【虚弱状态】测试玩家1，你需要做出选择：',
      options: [
        { id: '0', label: '摸3张牌继续执行后续行动', button_label: '摸3张牌继续执行后续行动' },
        { id: '1', label: '跳过此回合', button_label: '跳过此回合' },
      ],
      min: 1,
      max: 1,
      presentation: {
        kind: 'branch_select',
        layout: 'overlay',
        numeric_base: 0,
      },
    })

    render(PromptDialog, {
      global: {
        plugins: [pinia],
      },
    })

    expect(screen.getByText('摸3张牌继续执行后续行动')).toBeInTheDocument()
    expect(screen.getByText('跳过此回合')).toBeInTheDocument()
    expect(screen.queryByText(/^1$/)).not.toBeInTheDocument()
    expect(screen.queryByText(/^2$/)).not.toBeInTheDocument()
  })

  it('does not render medusa dark moon pick as decision overlay', async () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    useSessionStore().setRoomInfo('ROOM1', 'p2', 'Blue', 'moon_goddess')
    useSnapshotStore().updateGameState(buildState())
    useInterruptStore().setPrompt(medusaDarkMoonPickPrompt())

    render(PromptDialog, {
      global: {
        plugins: [pinia],
      },
    })

    expect(screen.queryByTestId('decision-overlay')).not.toBeInTheDocument()
    expect(screen.queryByText('移除闇月[暗月法术/Magic/Dark]')).not.toBeInTheDocument()
  })

  it('renders moon cycle branch prompt with decline and both branches', async () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    useSessionStore().setRoomInfo('ROOM1', 'p2', 'Blue', 'moon_goddess')
    useSnapshotStore().updateGameState(buildState({
      players: {
        p2: buildPlayer({
          id: 'p2',
          name: '测试玩家2',
          camp: 'Blue',
          role: 'moon_goddess',
          hand: [],
          field: [],
          is_active: true,
          buffs: [],
          tokens: {},
        }),
        p3: buildPlayer({
          id: 'p3',
          name: '干扰玩家',
          camp: 'Red',
          role: 'fighter',
          hand: [],
          field: [],
          is_active: false,
          buffs: [],
          tokens: {},
        }),
      },
    }))
    useInterruptStore().setPrompt(moonCycleBranchPrompt())

    render(PromptDialog, {
      global: {
        plugins: [pinia],
      },
    })

    expect(screen.getByTestId('decision-overlay')).toBeInTheDocument()
    expect(screen.getByTestId('prompt-option-decline')).toBeInTheDocument()
    expect(screen.queryByTestId('prompt-cancel-btn')).not.toBeInTheDocument()
    expect(screen.getByText('不发动')).toBeInTheDocument()
    expect(screen.getByText('分支①：移除1个闇月，令目标角色+1治疗')).toBeInTheDocument()
    expect(screen.getByText('分支②：移除1点治疗，你+1新月')).toBeInTheDocument()

    await userEvent.click(screen.getByTestId('prompt-option-decline'))
    expect(submitSelectMock).toHaveBeenCalledWith([0])
    expect(submitCancelMock).not.toHaveBeenCalled()
  })

  it('renders moon cycle healing target as a target-selection hint instead of option buttons', async () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    useSessionStore().setRoomInfo('ROOM1', 'p2', 'Blue', 'moon_goddess')
    useSnapshotStore().updateGameState(buildState({
      players: {
        p2: buildPlayer({
          id: 'p2',
          name: '测试玩家2',
          camp: 'Blue',
          role: 'moon_goddess',
          is_active: true,
        }),
        p3: buildPlayer({
          id: 'p3',
          name: '目标玩家',
          camp: 'Red',
          role: 'fighter',
        }),
      },
    }))
    useInterruptStore().setPrompt(moonCycleTargetPrompt())

    render(PromptDialog, {
      global: {
        plugins: [pinia],
      },
    })

    expect(screen.getByText('【月之轮回】请选择获得1点治疗的角色：')).toBeInTheDocument()
    expect(screen.queryByTestId('decision-overlay')).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: '目标玩家' })).not.toBeInTheDocument()
  })

  it('renders weakness choice as a decision overlay instead of a hand-card picker', async () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    useSessionStore().setRoomInfo('ROOM1', 'p2', 'Blue', 'fighter')
    useSnapshotStore().updateGameState(buildState())
    useInterruptStore().setPrompt(weakPrompt())

    render(PromptDialog, {
      global: {
        plugins: [pinia],
      },
    })

    expect(screen.getByTestId('decision-overlay')).toBeInTheDocument()
    expect(screen.getByText('摸3张牌继续执行后续行动')).toBeInTheDocument()
    expect(screen.getByText('跳过此回合')).toBeInTheDocument()
    expect(screen.queryByText('完成选牌后点击发动')).not.toBeInTheDocument()

    await userEvent.click(screen.getByText('摸3张牌继续执行后续行动'))

    expect(submitSelectMock).toHaveBeenCalledWith([0])
  })

  it('renders magic bullet direction through the direction renderer and submits reverse as option index 1', async () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    useSessionStore().setRoomInfo('ROOM1', 'p2', 'Blue', 'fighter')
    useSnapshotStore().updateGameState(buildState())
    useInterruptStore().setPrompt(magicBulletDirectionPrompt())

    render(PromptDialog, {
      global: {
        plugins: [pinia],
      },
    })

    expect(screen.getByTestId('direction-prompt')).toBeInTheDocument()
    expect(screen.queryByTestId('prompt-dialog')).not.toBeInTheDocument()
    expect(screen.getByTestId('branch-option-0')).toBeInTheDocument()

    await userEvent.click(screen.getByTestId('direction-option-reverse'))

    expect(submitSelectMock).toHaveBeenCalledWith([1])
  })

  it('renders radiant cannon direction through the direction renderer and submits the second option index', async () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    useSessionStore().setRoomInfo('ROOM1', 'p2', 'Blue', 'fighter')
    useSnapshotStore().updateGameState(buildState())
    useInterruptStore().setPrompt(radiantCannonDirectionPrompt())

    render(PromptDialog, {
      global: {
        plugins: [pinia],
      },
    })

    expect(screen.getByTestId('direction-prompt')).toBeInTheDocument()
    expect(screen.queryByTestId('prompt-dialog')).not.toBeInTheDocument()
    expect(screen.getByTestId('branch-option-0')).toBeInTheDocument()

    await userEvent.click(screen.getByTestId('direction-option-blue'))

    expect(submitSelectMock).toHaveBeenCalledWith([1])
  })

  it('renders heal mitigation as a numeric decision instead of a hand-card picker', async () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    useSessionStore().setRoomInfo('ROOM1', 'p2', 'Blue', 'fighter')
    useSnapshotStore().updateGameState(buildState())
    useInterruptStore().setPrompt(healPrompt())

    render(PromptDialog, {
      global: {
        plugins: [pinia],
      },
    })

    expect(screen.getByTestId('decision-overlay')).toBeInTheDocument()
    expect(screen.getByTestId('numeric-option-0')).toBeInTheDocument()
    expect(screen.getByTestId('numeric-option-2')).toBeInTheDocument()
    expect(screen.queryByText('完成选牌后点击发动')).not.toBeInTheDocument()

    await userEvent.click(screen.getByTestId('numeric-option-2'))

    expect(submitSelectMock).toHaveBeenCalledWith([2])
  })

  it('renders fraud attack element prompt through the fraud renderer and submits option index', async () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    useSessionStore().setRoomInfo('ROOM1', 'p2', 'Blue', 'trickster')
    useSnapshotStore().updateGameState(buildState())
    useInterruptStore().setPrompt(fraudAttackElementPrompt())

    render(PromptDialog, {
      global: {
        plugins: [pinia],
      },
    })

    expect(screen.getByTestId('decision-overlay')).toBeInTheDocument()
    expect(screen.getByText('水涟斩')).toBeInTheDocument()
    expect(screen.getByText('火焰斩')).toBeInTheDocument()
    expect(screen.queryByTestId('numeric-option-1')).not.toBeInTheDocument()

    await userEvent.click(screen.getByTestId('prompt-option-fire'))

    expect(submitSelectMock).toHaveBeenCalledWith([1])
  })

  it('renders saint heal allocation through the allocation renderer and submits values', async () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    useSessionStore().setRoomInfo('ROOM1', 'p2', 'Blue', 'saintess')
    useSnapshotStore().updateGameState(buildState())
    useInterruptStore().setPrompt(saintHealAllocatePrompt())

    render(PromptDialog, {
      global: {
        plugins: [pinia],
      },
    })

    expect(screen.getByTestId('allocation-overlay')).toBeInTheDocument()
    expect(screen.getByText('治疗目标一')).toBeInTheDocument()
    expect(screen.getByText('治疗目标二')).toBeInTheDocument()
    expect(screen.queryByTestId('numeric-option-1')).not.toBeInTheDocument()

    await userEvent.click(screen.getByTestId('allocation-option-0-1'))
    await userEvent.click(screen.getByTestId('allocation-option-1-2'))
    await userEvent.click(screen.getByTestId('allocation-submit'))

    expect(submitSelectMock).toHaveBeenCalledWith([1, 2])
  })

  it('keeps rune allocation submit disabled until all runes are allocated', async () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    useSessionStore().setRoomInfo('ROOM1', 'p2', 'Blue', 'war_homunculus')
    useSnapshotStore().updateGameState(buildState())
    useInterruptStore().setPrompt(runeReforgeAllocatePrompt())

    render(PromptDialog, {
      global: {
        plugins: [pinia],
      },
    })

    expect(screen.getByTestId('allocation-overlay')).toBeInTheDocument()
    expect(screen.getByTestId('allocation-submit')).toBeDisabled()

    await userEvent.click(screen.getByTestId('allocation-option-0-2'))
    expect(screen.getByTestId('allocation-submit')).toBeDisabled()

    await userEvent.click(screen.getByTestId('allocation-option-1-1'))
    await userEvent.click(screen.getByTestId('allocation-submit'))

    expect(submitSelectMock).toHaveBeenCalledWith([2, 1])
  })

  it('renders bloody prayer allocation through the allocation renderer and submits values', async () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    useSessionStore().setRoomInfo('ROOM1', 'p2', 'Blue', 'crimson_knight')
    useSnapshotStore().updateGameState(buildState())
    useInterruptStore().setPrompt(bloodyPrayerAllocatePrompt())

    render(PromptDialog, {
      global: {
        plugins: [pinia],
      },
    })

    expect(screen.getByTestId('allocation-overlay')).toBeInTheDocument()
    expect(screen.getByText('治疗目标一（治疗:2）')).toBeInTheDocument()
    expect(screen.getByText('治疗目标二（治疗:1）')).toBeInTheDocument()

    await userEvent.click(screen.getByTestId('allocation-option-0-1'))
    await userEvent.click(screen.getByTestId('allocation-option-1-2'))
    await userEvent.click(screen.getByTestId('allocation-submit'))

    expect(submitSelectMock).toHaveBeenCalledWith([1, 2])
  })

  it('renders response prompt through the response renderer and keeps take submit behavior', async () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    useSessionStore().setRoomInfo('ROOM1', 'p2', 'Blue', 'fighter')
    useSnapshotStore().updateGameState(buildState())
    useInterruptStore().setPrompt(responsePrompt())

    render(PromptDialog, {
      global: {
        plugins: [pinia],
      },
    })

    expect(screen.getByTestId('response-prompt')).toBeInTheDocument()
    expect(screen.getByText('此次攻击系别：火系（应战需同系或暗灭）')).toBeInTheDocument()
    expect(screen.getByTestId('prompt-option-take')).toBeInTheDocument()
    expect(screen.getByTestId('prompt-option-defend')).toBeInTheDocument()
    expect(screen.getByTestId('prompt-option-counter')).toBeInTheDocument()
    expect(screen.queryByTestId('decision-overlay')).not.toBeInTheDocument()

    await userEvent.click(screen.getByTestId('prompt-option-take'))

    expect(submitRespondTakeMock).toHaveBeenCalledOnce()
    expect(submitSelectMock).not.toHaveBeenCalled()
  })

  it('renders card picker confirm-only via card picker renderer and keeps confirm disabled/enabled behavior', async () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    useSessionStore().setRoomInfo('ROOM1', 'p2', 'Blue', 'fighter')
    useSnapshotStore().updateGameState(buildState())
    const interruptStore = useInterruptStore()
    interruptStore.setPrompt(handCardPickerPrompt())

    render(PromptDialog, {
      global: {
        plugins: [pinia],
      },
    })

    expect(screen.getByTestId('card-picker-prompt')).toBeInTheDocument()
    expect(screen.queryByTestId('prompt-cancel-btn')).not.toBeInTheDocument()

    const confirmBtn = screen.getByTestId('prompt-confirm-btn')
    expect(confirmBtn).toBeDisabled()
    await userEvent.click(confirmBtn)
    expect(submitSelectCardIDsMock).not.toHaveBeenCalled()

    interruptStore.setSelectedHandIndexes([0])
    await nextTick()

    expect(screen.getByTestId('prompt-confirm-btn')).not.toBeDisabled()
    await userEvent.click(screen.getByTestId('prompt-confirm-btn'))
    expect(submitSelectCardIDsMock).toHaveBeenCalledWith(['h0'])
  })

  it('submits proxy card picker selections by matching card_id to hand cards', async () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    useSessionStore().setRoomInfo('ROOM1', 'p2', 'Blue', 'magic_lancer')
    useSnapshotStore().updateGameState(buildState())
    const interruptStore = useInterruptStore()
    interruptStore.setPrompt(handCardPickerPrompt({
      presentation: {
        kind: 'card_picker',
        card_source: 'proxy',
        numeric_base: 0,
      },
    }))

    render(PromptDialog, {
      global: {
        plugins: [pinia],
      },
    })

    interruptStore.setSelectedHandIndexes([1])
    await nextTick()

    await userEvent.click(screen.getByTestId('prompt-confirm-btn'))
    expect(submitSelectCardIDsMock).toHaveBeenCalledWith(['h1'])
  })

  it('submits elf blessing cover picker selections by matching card_id to playable blessing cards', async () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    useSessionStore().setRoomInfo('ROOM1', 'p2', 'Blue', 'elf_archer')
    useSnapshotStore().updateGameState(buildState({
      players: {
        p2: buildPlayer({
          id: 'p2',
          name: 'Elf',
          camp: 'Blue',
          role: 'elf_archer',
          hand: [
            buildCard({ id: 'h0', name: '火焰斩', type: 'Attack', element: 'Fire' }),
          ],
          field: [
            {
              mode: 'Cover',
              effect: 'ElfBlessing',
              card: buildCard({ id: 'blessing-1', name: '祝福圣盾', type: 'Magic', element: 'Light' }),
              owner_id: 'p2',
              source_id: 'p2',
              field_hook: 'Manual',
              locked: false,
              duration: 0,
            },
          ],
        }),
      },
    }))
    const interruptStore = useInterruptStore()
    interruptStore.setPrompt(handCardPickerPrompt({
      options: [
        { id: '1', label: '2: 祝福圣盾', button_label: '选择', card_id: 'blessing-1' },
      ],
      presentation: {
        kind: 'card_picker',
        card_source: 'hand',
        card_filter: 'magic_or_elf_blessing',
        numeric_base: 0,
      },
    }))

    render(PromptDialog, {
      global: {
        plugins: [pinia],
      },
    })

    interruptStore.setSelectedHandIndexes([1])
    await nextTick()

    await userEvent.click(screen.getByTestId('prompt-confirm-btn'))
    expect(submitSelectCardIDsMock).toHaveBeenCalledWith(['blessing-1'])
  })

  it('renders card picker decline row via renderer and keeps cancel/confirm submission behavior', async () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    useSessionStore().setRoomInfo('ROOM1', 'p2', 'Blue', 'fighter')
    useSnapshotStore().updateGameState(buildState())
    useInterruptStore().setPrompt(handCardPickerWithDeclinePrompt())

    render(PromptDialog, {
      global: {
        plugins: [pinia],
      },
    })

    expect(screen.getByTestId('card-picker-prompt')).toBeInTheDocument()
    expect(screen.getByTestId('prompt-cancel-btn')).toBeInTheDocument()
    expect(screen.getByText('请选择是否发动该卡牌效果')).toBeInTheDocument()

    await userEvent.click(screen.getByTestId('prompt-cancel-btn'))
    expect(submitCancelMock).toHaveBeenCalledOnce()

    await userEvent.click(screen.getByTestId('prompt-confirm-btn'))
    expect(submitSelectCardIDsMock).toHaveBeenCalledWith([])
  })

  it('renders extract prompt through extract renderer and confirms selected option indexes', async () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    useSessionStore().setRoomInfo('ROOM1', 'p2', 'Blue', 'fighter')
    useSnapshotStore().updateGameState(buildState())
    useInterruptStore().setPrompt(extractPrompt())

    render(PromptDialog, {
      global: {
        plugins: [pinia],
      },
    })

    expect(screen.getByTestId('extract-prompt')).toBeInTheDocument()
    expect(screen.getByTestId('extract-option-0')).toBeInTheDocument()
    expect(screen.getByTestId('extract-option-1')).toBeInTheDocument()
    expect(screen.getByTestId('extract-option-2')).toBeInTheDocument()

    await userEvent.click(screen.getByTestId('extract-option-0'))
    await userEvent.click(screen.getByTestId('extract-option-1'))
    await userEvent.click(screen.getByTestId('prompt-confirm-btn'))

    expect(submitSelectMock).toHaveBeenCalledWith([0, 1])
  })

  it('keeps single skill choice confirm behavior and submits the selected index', async () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    useSessionStore().setRoomInfo('ROOM1', 'p2', 'Blue', 'fighter')
    useSnapshotStore().updateGameState(buildState())
    useInterruptStore().setPrompt(singleSkillChoicePrompt())

    render(PromptDialog, {
      global: {
        plugins: [pinia],
      },
    })

    expect(screen.getByTestId('prompt-option-fire_blast')).toBeInTheDocument()
    expect(screen.queryByTestId('skill-branch-overlay')).not.toBeInTheDocument()

    await userEvent.click(screen.getByTestId('prompt-option-fire_blast'))

    expect(submitSelectMock).toHaveBeenCalledWith([0])
  })

  it('renders multi skill choice overlay and keeps select/skip submission behavior', async () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    useSessionStore().setRoomInfo('ROOM1', 'p2', 'Blue', 'fighter')
    useSnapshotStore().updateGameState(buildState())
    useInterruptStore().setPrompt(multiSkillChoicePrompt())

    render(PromptDialog, {
      global: {
        plugins: [pinia],
      },
    })

    expect(screen.getByTestId('skill-branch-overlay')).toBeInTheDocument()
    expect(screen.getByTestId('decision-overlay')).toBeInTheDocument()
    expect(screen.getByTestId('branch-option-1')).toBeInTheDocument()

    await userEvent.click(screen.getByTestId('branch-option-1'))
    expect(submitSelectMock).toHaveBeenCalledWith([1])

    submitSelectMock.mockReset()

    await userEvent.click(screen.getByTestId('prompt-option-skip'))
    expect(submitCancelMock).toHaveBeenCalledOnce()
    expect(submitSelectMock).not.toHaveBeenCalled()
  })

  it('does not parse colon labels as old service skill title shorthand', async () => {
    const pinia = createPinia()
    setActivePinia(pinia)

    useSessionStore().setRoomInfo('ROOM1', 'p2', 'Blue', 'fighter')
    useSnapshotStore().updateGameState(buildState())
    useInterruptStore().setPrompt(colonLabelSkillChoicePrompt())

    render(PromptDialog, {
      global: {
        plugins: [pinia],
      },
    })

    expect(screen.getAllByText('烈焰突刺：造成2点火焰伤害').length).toBeGreaterThan(0)
    expect(screen.queryByText(/^烈焰突刺$/)).not.toBeInTheDocument()
  })
})
