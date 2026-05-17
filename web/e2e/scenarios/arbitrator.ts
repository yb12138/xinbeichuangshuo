// ============================================================
// Arbitrator (仲裁者) Protocol Harness Scenarios
// ============================================================

import type { AvailableSkill, Card, Prompt } from '../../src/types/game';
import type { WsMessage } from '../../src/network/protocol';
import {
  availableSkill,
  card,
  characterView,
  playerInfo,
  playerView,
  requireActionMessage,
  syncState,
  type ProtocolHarnessScenario,
} from './builders';

export const ARBITRATOR_PLAYER_ID = 'arbitrator_player';
export const ENEMY_PLAYER_ID = 'enemy_1';
export const ALLY_PLAYER_ID = 'ally_1';

export const ARBITRATOR_RITUAL_INTERRUPT_ID = 'arbitrator_ritual_interrupt';
export const ARBITRATOR_DOOM_JUDGMENT_ID = 'arbitrator_doom_judgment';
export const ARBITRATOR_ARBITRATION_RITUAL_ID = 'arbitrator_arbitration_ritual';
export const ARBITRATOR_JUDGMENT_BALANCE_ID = 'arbitrator_judgment_balance';

const arbitratorCharacter = characterView({
  id: 'arbitrator',
  name: '仲裁者',
  title: '血',
  faction: '血',
  skills: [
    {
      id: 'arbitrator_rule',
      title: '仲裁法则',
      description: '游戏初始时，你+2［水晶］。',
      type: 0, // 被动
      min_targets: 0, max_targets: 0, target_type: 0,
    },
    {
      id: ARBITRATOR_RITUAL_INTERRUPT_ID,
      title: '仪式中断',
      description: '（仅［审判形态］下发动）［转正］脱离［审判形态］，我方［战绩区］+1［宝石］。',
      type: 1, // 启动
      min_targets: 0, max_targets: 0, target_type: 0,
    },
    {
      id: ARBITRATOR_DOOM_JUDGMENT_ID,
      title: '末日审判',
      description: '（移除所有［审判］）对目标角色造成等量的法术伤害③；在你的行动阶段开始时，若［审判］已达到上限，该行动阶段你必须发动［末日审判］。',
      type: 2, // 法术
      min_targets: 1, max_targets: 1, target_type: 3,
    },
    {
      id: 'arbitrator_wave',
      title: '审判浪潮',
      description: '（你每次承受伤害⑥）你+1［审判］。',
      type: 0, // 被动
      min_targets: 0, max_targets: 0, target_type: 0,
    },
    {
      id: ARBITRATOR_ARBITRATION_RITUAL_ID,
      title: '仲裁仪式',
      description: '［持续］［宝石］［横置］转为［审判形态］，你的手牌上限恒定为5；每次在你的回合开始时，你+1［审判］。',
      type: 1, // 启动(大招)
      min_targets: 0, max_targets: 0, target_type: 0,
    },
    {
      id: ARBITRATOR_JUDGMENT_BALANCE_ID,
      title: '判决天平',
      description: '［水晶］你+1［审判］，再选择以下一项发动：●弃掉你的所有手牌。●将你的手牌补到上限［强制］，我方战绩区+1［宝石］。',
      type: 2, // 法术(大招)
      min_targets: 0, max_targets: 0, target_type: 0,
    },
  ],
});

const enemyCharacter = characterView({
  id: 'enemy_char', name: '守卫', title: '测试目标', faction: '异端', skills: [],
});

const allyCharacter = characterView({
  id: 'ally_char', name: '勇者', title: '测试队友', faction: '血', skills: [],
});

const defaultCharacters = [arbitratorCharacter, enemyCharacter, allyCharacter];

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function arbitratorHand(): Card[] {
  return [
    card({ id: 'arb-fire-atk1', name: '火焰斩', type: 'Attack', element: 'Fire' }),
    card({ id: 'arb-fire-atk2', name: '烈焰斩', type: 'Attack', element: 'Fire' }),
    card({ id: 'arb-water-atk', name: '水刃', type: 'Attack', element: 'Water' }),
    card({ id: 'arb-magic', name: '火球', type: 'Magic', element: 'Fire' }),
    card({ id: 'arb-thunder', name: '雷击', type: 'Magic', element: 'Thunder' }),
  ];
}

function arbitratorAvailableSkill(skill: Partial<AvailableSkill> & { id: string; title: string }): AvailableSkill {
  return availableSkill({
    description: '',
    target_type: 0,
    min_targets: 0, max_targets: 0,
    ...skill,
  });
}

// ---------------------------------------------------------------------------
// Scenario Factory
// ---------------------------------------------------------------------------

export function arbitratorScenario(options: {
  hand?: Card[];
  availableSkills?: AvailableSkill[];
  crystal?: number;
  gem?: number;
  turnStage?: string;
  tokens?: Record<string, number>;
  inJudgmentForm?: boolean;
} = {}): ProtocolHarnessScenario {
  const hand = options.hand ?? arbitratorHand();
  const buffs = options.inJudgmentForm ? [{ id: 'judgment_form', name: '审判形态', duration: 1, value: 0, source_id: ARBITRATOR_ARBITRATION_RITUAL_ID }] : [];
  const players = [
    playerView({
      id: ARBITRATOR_PLAYER_ID,
      name: 'E2E Arbitrator',
      camp: 'Red',
      role: 'arbitrator',
      hand,
      hand_count: hand.length,
      crystal: options.crystal ?? 2, // 仲裁法则：游戏开始+2水晶
      gem: options.gem ?? 0,
      is_active: true,
      tokens: options.tokens ?? {},
      buffs,
    }),
    playerView({
      id: ENEMY_PLAYER_ID,
      name: 'Enemy E1',
      camp: 'Blue',
      role: 'enemy_char',
      hand: [card({ id: 'en-card-1', name: '测试牌', type: 'Attack', element: 'Fire' })],
      hand_count: 1, max_hand: 6,
      heal: 0, max_heal: 4,
      is_active: false,
    }),
    playerView({
      id: ALLY_PLAYER_ID,
      name: 'Ally A1',
      camp: 'Red',
      role: 'ally_char',
      hand: [card({ id: 'al-card-1', name: '测试牌', type: 'Attack', element: 'Water' })],
      hand_count: 1, max_hand: 6,
      heal: 0, max_heal: 4,
      is_active: false,
    }),
  ];

  return {
    roomCode: 'MOCK',
    myPlayerId: ARBITRATOR_PLAYER_ID,
    myPlayerName: 'E2E Arbitrator',
    characters: defaultCharacters,
    players: [
      playerInfo({ id: ARBITRATOR_PLAYER_ID, name: 'E2E Arbitrator', camp: 'Red', char_role: 'arbitrator', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy E1', camp: 'Blue', char_role: 'enemy_char' }),
      playerInfo({ id: ALLY_PLAYER_ID, name: 'Ally A1', camp: 'Red', char_role: 'ally_char' }),
    ],
    initialState: syncState({
      turn_player_id: ARBITRATOR_PLAYER_ID,
      turn_stage: options.turnStage ?? 'ActionExecution',
      available_skills: options.availableSkills ?? [],
      characters: defaultCharacters,
      players,
    }),
  };
}

// ============================================================
// Ritual Interrupt (仪式中断) - 启动技能
// ============================================================

export function ritualInterruptScenario(): ProtocolHarnessScenario {
  return arbitratorScenario({
    inJudgmentForm: true,
    turnStage: 'StartupPhase',
    availableSkills: [
      arbitratorAvailableSkill({
        id: ARBITRATOR_RITUAL_INTERRUPT_ID, title: '仪式中断',
      }),
    ],
  });
}

export function ritualInterruptPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: ARBITRATOR_PLAYER_ID,
    message: '【仪式中断】是否脱离［审判形态］，我方战绩区+1宝石？',
    choice_type: 'arbiter_ritual_interrupt',
    options: [
      { id: 'confirm', label: '发动' },
      { id: 'skip', label: '跳过' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}

// ============================================================
// Doom Judgment (末日审判) - 法术技能
// ============================================================

export function doomJudgmentScenario(options: {
  judgmentCount?: number;
  force?: boolean;
} = {}): ProtocolHarnessScenario {
  return arbitratorScenario({
    tokens: { judgment: options.judgmentCount ?? 3 },
    turnStage: 'ActionExecution',
    availableSkills: [
      arbitratorAvailableSkill({
        id: ARBITRATOR_DOOM_JUDGMENT_ID, title: '末日审判',
        min_targets: 1, max_targets: 1, target_type: 3,
      }),
    ],
  });
}

export function doomJudgmentTargetPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: ARBITRATOR_PLAYER_ID,
    message: '【末日审判】请选择一名目标角色，移除所有［审判］造成等量法术伤害：',
    choice_type: 'arbiter_doom_judgment_target',
    options: [
      { id: ENEMY_PLAYER_ID, label: 'Enemy E1' },
      { id: ALLY_PLAYER_ID, label: 'Ally A1' },
      { id: ARBITRATOR_PLAYER_ID, label: '自己' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}

export function doomJudgmentForcePrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: ARBITRATOR_PLAYER_ID,
    message: '【末日审判］［审判］已达上限，必须发动该技能。请选择目标：',
    choice_type: 'arbiter_doom_judgment_force',
    options: [
      { id: ENEMY_PLAYER_ID, label: 'Enemy E1' },
      { id: ALLY_PLAYER_ID, label: 'Ally A1' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}

// ============================================================
// Arbitration Ritual (仲裁仪式) - 启动技能(大招)
// ============================================================

export function arbitrationRitualScenario(options: {
  gem?: number;
} = {}): ProtocolHarnessScenario {
  return arbitratorScenario({
    gem: options.gem ?? 1,
    turnStage: 'StartupPhase',
    availableSkills: [
      arbitratorAvailableSkill({
        id: ARBITRATOR_ARBITRATION_RITUAL_ID, title: '仲裁仪式',
      }),
    ],
  });
}

export function arbitrationRitualPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: ARBITRATOR_PLAYER_ID,
    message: '【仲裁仪式】是否消耗1个红宝石发动，［横置］转为［审判形态］？',
    choice_type: 'arbiter_arbitration_ritual',
    options: [
      { id: 'confirm', label: '发动' },
      { id: 'skip', label: '跳过' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}

// ============================================================
// Judgment Balance (判决天平) - 法术技能(大招)
// ============================================================

export function judgmentBalanceScenario(): ProtocolHarnessScenario {
  return arbitratorScenario({
    crystal: 1,
    turnStage: 'ActionExecution',
    availableSkills: [
      arbitratorAvailableSkill({
        id: ARBITRATOR_JUDGMENT_BALANCE_ID, title: '判决天平',
      }),
    ],
  });
}

export function judgmentBalanceBranchPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: ARBITRATOR_PLAYER_ID,
    message: '【判决天平】请选择一项发动：',
    choice_type: 'arbiter_balance_mode',
    options: [
      { id: 'discard_all', label: '弃掉所有手牌' },
      { id: 'fill_hand', label: '补牌到上限，战绩区+1宝石' },
    ],
    min: 1, max: 1,
    presentation: { kind: 'branch_select', layout: 'overlay' },
  } satisfies Prompt);
}