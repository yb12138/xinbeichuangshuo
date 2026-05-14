// ============================================================
// Fighter (格斗家) Protocol Harness Scenarios
// ============================================================

import type { Prompt } from '../../src/types/game';
import type { WsMessage } from '../../src/network/protocol';
import {
  card,
  characterView,
  playerInfo,
  playerView,
  requireActionMessage,
  syncState,
  type ProtocolHarnessScenario,
} from './builders';

// ---- Player IDs ----
export const FIGHTER_PLAYER_ID = 'fighter_player';
export const ALLY_PLAYER_ID = 'ally_1';
export const ENEMY_PLAYER_ID = 'enemy_1';

// ---- Skill IDs ----
export const FIGHTER_CHARGE_ATTACK_SKILL_ID = 'fighter_charge_attack';
export const FIGHTER_BURST_CRASH_SKILL_ID = 'fighter_burst_crash';
export const FIGHTER_BULLET_SKILL_ID = 'fighter_bullet';
export const FIGHTER_HUNDRED_DRAGON_SKILL_ID = 'fighter_hundred_dragon';
export const FIGHTER_HEAVEN_DRIVE_SKILL_ID = 'fighter_heaven_drive';

// ---- Fighter character definition ----
const fighterCharacter = characterView({
  id: 'fighter',
  name: '格斗家',
  title: '斗气战士',
  faction: '星杯',
  skills: [
    {
      id: FIGHTER_CHARGE_ATTACK_SKILL_ID,
      title: '蓄力一击',
      description: '攻击时发动，若斗气未到上限，伤害+X',
      type: 0,
      min_targets: 0,
      max_targets: 0,
      target_type: 0,
      cost_gem: 0,
      cost_crystal: 0,
      cost_discards: 0,
    },
    {
      id: FIGHTER_BURST_CRASH_SKILL_ID,
      title: '气绝崩击',
      description: '攻击时发动，消耗斗气，效果由后端结算',
      type: 0,
      min_targets: 0,
      max_targets: 0,
      target_type: 0,
      cost_gem: 0,
      cost_crystal: 0,
      cost_discards: 0,
    },
    {
      id: FIGHTER_BULLET_SKILL_ID,
      title: '念弹',
      description: '法术行动后发动',
      type: 0,
      min_targets: 0,
      max_targets: 0,
      target_type: 0,
      cost_gem: 0,
      cost_crystal: 0,
      cost_discards: 0,
    },
    {
      id: FIGHTER_HUNDRED_DRAGON_SKILL_ID,
      title: '百式幻龙拳',
      description: '启动技能，消耗3点斗气',
      type: 2,
      min_targets: 0,
      max_targets: 0,
      target_type: 0,
      cost_gem: 0,
      cost_crystal: 0,
      cost_discards: 0,
    },
    {
      id: FIGHTER_HEAVEN_DRIVE_SKILL_ID,
      title: '斗神天驱',
      description: '启动技能，消耗水晶',
      type: 2,
      min_targets: 0,
      max_targets: 0,
      target_type: 0,
      cost_gem: 0,
      cost_crystal: 0,
      cost_discards: 0,
    },
  ],
});

const allyCharacter = characterView({
  id: 'ally_char',
  name: '圣女',
  title: '光之守护',
  faction: '星杯',
  skills: [],
});

const enemyCharacter = characterView({
  id: 'enemy_char',
  name: '魔神',
  title: '暗影之王',
  faction: '异端',
  skills: [],
});

// ---- Helper functions ----
function fighterHand() {
  return [
    card({ id: 'card_1', name: '火焰斩', type: 'Attack', element: 'Fire' }),
    card({ id: 'card_2', name: '火焰斩', type: 'Attack', element: 'Fire' }),
    card({ id: 'card_3', name: '水涟斩', type: 'Attack', element: 'Water' }),
    card({ id: 'card_4', name: '寒冰箭', type: 'Magic', element: 'Water' }),
    card({ id: 'card_5', name: '圣光', type: 'Magic', element: 'Light' }),
  ];
}

function fighterPlayerView(options: {
  is_active?: boolean;
  qi?: number;
  max_qi?: number;
  gems?: number;
  crystals?: number;
} = {}) {
  return playerView({
    id: FIGHTER_PLAYER_ID,
    name: 'E2E Fighter',
    camp: 'Red',
    role: 'fighter',
    hand: fighterHand(),
    hand_count: fighterHand().length,
    heal: 2,
    max_heal: 4,
    is_active: options.is_active ?? true,
    gem: options.gems ?? 0,
    crystal: options.crystals ?? 0,
    tokens: {
      fighter_qi: options.qi ?? 0,
    },
  });
}

// ============================================================
// 蓄力一击 (fighter_charge_attack) - Attack response when qi not maxed
// ============================================================

export function chargeAttackScenario(options: { qi?: number } = {}): ProtocolHarnessScenario {
  const qi = options.qi ?? 1;
  const characters = [fighterCharacter, allyCharacter, enemyCharacter];

  const fighter = fighterPlayerView({ qi, is_active: true });

  const players = [
    fighter,
    playerView({
      id: ENEMY_PLAYER_ID,
      name: 'Enemy Bot',
      camp: 'Blue',
      role: 'villain',
      hand: [],
      hand_count: 3,
      heal: 1,
      max_heal: 2,
      is_active: false,
    }),
  ];

  return {
    roomCode: 'MOCK',
    myPlayerId: FIGHTER_PLAYER_ID,
    myPlayerName: 'E2E Fighter',
    characters,
    players: [
      playerInfo({ id: FIGHTER_PLAYER_ID, name: 'E2E Fighter', camp: 'Red', char_role: 'fighter', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy Bot', camp: 'Blue', char_role: 'villain' }),
    ],
    initialState: syncState({
      turn_player_id: FIGHTER_PLAYER_ID,
      turn_stage: 'ActionExecution',
      available_skills: [],
      characters,
      players,
    }),
  };
}

export function chargeAttackConfirmPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: FIGHTER_PLAYER_ID,
    message: '【蓄力一击】是否发动？',
    choice_type: 'fighter_charge_attack_confirm',
    skill_id: FIGHTER_CHARGE_ATTACK_SKILL_ID,
    options: [
      { id: '0', label: '发动' },
      { id: '1', label: '不发动' },
    ],
    min: 1,
    max: 1,
  } satisfies Prompt);
}

// ============================================================
// 气绝崩击 (fighter_burst_crash) - Attack response when qi >= 1
// ============================================================

export function burstCrashScenario(options: { qi?: number } = {}): ProtocolHarnessScenario {
  const qi = options.qi ?? 2;
  const characters = [fighterCharacter, allyCharacter, enemyCharacter];

  const fighter = fighterPlayerView({ qi, is_active: true });

  const players = [
    fighter,
    playerView({
      id: ENEMY_PLAYER_ID,
      name: 'Enemy Bot',
      camp: 'Blue',
      role: 'villain',
      hand: [],
      hand_count: 3,
      heal: 1,
      max_heal: 2,
      is_active: false,
    }),
  ];

  return {
    roomCode: 'MOCK',
    myPlayerId: FIGHTER_PLAYER_ID,
    myPlayerName: 'E2E Fighter',
    characters,
    players: [
      playerInfo({ id: FIGHTER_PLAYER_ID, name: 'E2E Fighter', camp: 'Red', char_role: 'fighter', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy Bot', camp: 'Blue', char_role: 'villain' }),
    ],
    initialState: syncState({
      turn_player_id: FIGHTER_PLAYER_ID,
      turn_stage: 'ActionExecution',
      available_skills: [],
      characters,
      players,
    }),
  };
}

export function burstCrashConfirmPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: FIGHTER_PLAYER_ID,
    message: '【气绝崩击】是否发动？消耗1点斗气',
    choice_type: 'fighter_burst_crash_confirm',
    skill_id: FIGHTER_BURST_CRASH_SKILL_ID,
    options: [
      { id: '0', label: '发动' },
      { id: '1', label: '不发动' },
    ],
    min: 1,
    max: 1,
  } satisfies Prompt);
}

// ============================================================
// 蓄力一击 & 气绝崩击 互斥选择 - Attack response when both can trigger
// ============================================================

export function attackSkillChoiceScenario(options: { qi?: number } = {}): ProtocolHarnessScenario {
  const qi = options.qi ?? 2;
  const characters = [fighterCharacter, allyCharacter, enemyCharacter];

  const fighter = fighterPlayerView({ qi, is_active: true });

  const players = [
    fighter,
    playerView({
      id: ENEMY_PLAYER_ID,
      name: 'Enemy Bot',
      camp: 'Blue',
      role: 'villain',
      hand: [],
      hand_count: 3,
      heal: 1,
      max_heal: 2,
      is_active: false,
    }),
  ];

  return {
    roomCode: 'MOCK',
    myPlayerId: FIGHTER_PLAYER_ID,
    myPlayerName: 'E2E Fighter',
    characters,
    players: [
      playerInfo({ id: FIGHTER_PLAYER_ID, name: 'E2E Fighter', camp: 'Red', char_role: 'fighter', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy Bot', camp: 'Blue', char_role: 'villain' }),
    ],
    initialState: syncState({
      turn_player_id: FIGHTER_PLAYER_ID,
      turn_stage: 'ActionExecution',
      available_skills: [],
      characters,
      players,
    }),
  };
}

export function attackSkillChoicePrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: FIGHTER_PLAYER_ID,
    message: '【格斗家】攻击时可发动的技能（互斥）：',
    choice_type: 'fighter_attack_skill_choice',
    options: [
      { id: FIGHTER_CHARGE_ATTACK_SKILL_ID, label: '发动【蓄力一击】' },
      { id: FIGHTER_BURST_CRASH_SKILL_ID, label: '发动【气绝崩击】' },
      { id: 'skip', label: '跳过' },
    ],
    presentation: { kind: 'branch_select', layout: 'overlay' },
    min: 1,
    max: 1,
  } satisfies Prompt);
}

// ============================================================
// 念弹 (fighter_bullet) - Post-magic action response
// ============================================================

export function bulletScenario(): ProtocolHarnessScenario {
  const characters = [fighterCharacter, allyCharacter, enemyCharacter];

  const fighter = fighterPlayerView({ qi: 1, is_active: true });

  const players = [
    fighter,
    playerView({
      id: ENEMY_PLAYER_ID,
      name: 'Enemy Bot',
      camp: 'Blue',
      role: 'villain',
      hand: [],
      hand_count: 3,
      heal: 1,
      max_heal: 2,
      is_active: false,
    }),
  ];

  return {
    roomCode: 'MOCK',
    myPlayerId: FIGHTER_PLAYER_ID,
    myPlayerName: 'E2E Fighter',
    characters,
    players: [
      playerInfo({ id: FIGHTER_PLAYER_ID, name: 'E2E Fighter', camp: 'Red', char_role: 'fighter', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy Bot', camp: 'Blue', char_role: 'villain' }),
    ],
    initialState: syncState({
      turn_player_id: FIGHTER_PLAYER_ID,
      turn_stage: 'ActionExecution',
      available_skills: [],
      characters,
      players,
    }),
  };
}

export function bulletConfirmPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: FIGHTER_PLAYER_ID,
    message: '【念弹】法术行动后是否发动？',
    choice_type: 'fighter_bullet_confirm',
    skill_id: FIGHTER_BULLET_SKILL_ID,
    options: [
      { id: '0', label: '发动' },
      { id: '1', label: '不发动' },
    ],
    min: 1,
    max: 1,
  } satisfies Prompt);
}

// ============================================================
// 百式幻龙拳 (fighter_hundred_dragon) - Start skill with 3+ qi
// ============================================================

export function hundredDragonScenario(options: { qi?: number } = {}): ProtocolHarnessScenario {
  const qi = options.qi ?? 3;
  const characters = [fighterCharacter, allyCharacter, enemyCharacter];

  const fighter = fighterPlayerView({ qi, is_active: true });

  const players = [
    fighter,
    playerView({
      id: ENEMY_PLAYER_ID,
      name: 'Enemy Bot',
      camp: 'Blue',
      role: 'villain',
      hand: [],
      hand_count: 3,
      heal: 1,
      max_heal: 2,
      is_active: false,
    }),
  ];

  return {
    roomCode: 'MOCK',
    myPlayerId: FIGHTER_PLAYER_ID,
    myPlayerName: 'E2E Fighter',
    characters,
    players: [
      playerInfo({ id: FIGHTER_PLAYER_ID, name: 'E2E Fighter', camp: 'Red', char_role: 'fighter', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy Bot', camp: 'Blue', char_role: 'villain' }),
    ],
    initialState: syncState({
      turn_player_id: FIGHTER_PLAYER_ID,
      turn_stage: 'StartPhase',
      available_skills: [],
      characters,
      players,
    }),
  };
}

export function hundredDragonConfirmPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: FIGHTER_PLAYER_ID,
    message: '【百式幻龙拳】回合开始是否发动？消耗3点斗气',
    choice_type: 'fighter_hundred_dragon_confirm',
    skill_id: FIGHTER_HUNDRED_DRAGON_SKILL_ID,
    options: [
      { id: '0', label: '发动' },
      { id: '1', label: '不发动' },
    ],
    min: 1,
    max: 1,
  } satisfies Prompt);
}

// ============================================================
// 斗神天驱 (fighter_heaven_drive) - Start skill with crystal
// ============================================================

export function heavenDriveScenario(options: { crystals?: number } = {}): ProtocolHarnessScenario {
  const crystals = options.crystals ?? 1;
  const characters = [fighterCharacter, allyCharacter, enemyCharacter];

  const fighter = fighterPlayerView({ qi: 1, crystals, is_active: true });

  const players = [
    fighter,
    playerView({
      id: ENEMY_PLAYER_ID,
      name: 'Enemy Bot',
      camp: 'Blue',
      role: 'villain',
      hand: [],
      hand_count: 3,
      heal: 1,
      max_heal: 2,
      is_active: false,
    }),
  ];

  return {
    roomCode: 'MOCK',
    myPlayerId: FIGHTER_PLAYER_ID,
    myPlayerName: 'E2E Fighter',
    characters,
    players: [
      playerInfo({ id: FIGHTER_PLAYER_ID, name: 'E2E Fighter', camp: 'Red', char_role: 'fighter', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy Bot', camp: 'Blue', char_role: 'villain' }),
    ],
    initialState: syncState({
      turn_player_id: FIGHTER_PLAYER_ID,
      turn_stage: 'StartPhase',
      available_skills: [],
      characters,
      players,
    }),
  };
}

export function heavenDriveConfirmPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: FIGHTER_PLAYER_ID,
    message: '【斗神天驱】回合开始是否发动？消耗1个水晶',
    choice_type: 'fighter_heaven_drive_confirm',
    skill_id: FIGHTER_HEAVEN_DRIVE_SKILL_ID,
    options: [
      { id: '0', label: '发动' },
      { id: '1', label: '不发动' },
    ],
    min: 1,
    max: 1,
  } satisfies Prompt);
}

// ============================================================
// 百式幻龙拳 & 斗神天驱 互斥选择 - Start skill choice when both can trigger
// ============================================================

export function startSkillChoiceScenario(options: { qi?: number; crystals?: number } = {}): ProtocolHarnessScenario {
  const qi = options.qi ?? 3;
  const crystals = options.crystals ?? 1;
  const characters = [fighterCharacter, allyCharacter, enemyCharacter];

  const fighter = fighterPlayerView({ qi, crystals, is_active: true });

  const players = [
    fighter,
    playerView({
      id: ENEMY_PLAYER_ID,
      name: 'Enemy Bot',
      camp: 'Blue',
      role: 'villain',
      hand: [],
      hand_count: 3,
      heal: 1,
      max_heal: 2,
      is_active: false,
    }),
  ];

  return {
    roomCode: 'MOCK',
    myPlayerId: FIGHTER_PLAYER_ID,
    myPlayerName: 'E2E Fighter',
    characters,
    players: [
      playerInfo({ id: FIGHTER_PLAYER_ID, name: 'E2E Fighter', camp: 'Red', char_role: 'fighter', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy Bot', camp: 'Blue', char_role: 'villain' }),
    ],
    initialState: syncState({
      turn_player_id: FIGHTER_PLAYER_ID,
      turn_stage: 'StartPhase',
      available_skills: [],
      characters,
      players,
    }),
  };
}

export function startSkillChoicePrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: FIGHTER_PLAYER_ID,
    message: '【格斗家】回合开始可发动的启动技能（互斥）：',
    choice_type: 'fighter_start_skill_choice',
    options: [
      { id: FIGHTER_HUNDRED_DRAGON_SKILL_ID, label: '发动【百式幻龙拳】' },
      { id: FIGHTER_HEAVEN_DRIVE_SKILL_ID, label: '发动【斗神天驱】' },
      { id: 'skip', label: '跳过' },
    ],
    presentation: { kind: 'branch_select', layout: 'overlay' },
    min: 1,
    max: 1,
  } satisfies Prompt);
}