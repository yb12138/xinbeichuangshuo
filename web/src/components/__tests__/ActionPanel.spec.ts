import { render, screen } from '@testing-library/vue'
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
    expect(sharedLifeButton).toHaveTextContent('缺少可用于发动的「同生共死」专属/独有牌')
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
})
