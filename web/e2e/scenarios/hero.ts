// ============================================================
// Hero (勇者) Protocol Harness Scenarios
// ============================================================

import type { Prompt, AvailableSkill } from '../../src/types/game';
import type { WsMessage } from '../../src/network/protocol';
import {
  card,
  characterView,
  playerInfo,
  playerView,
  requireActionMessage,
  syncState,
  availableSkill,
  type ProtocolHarnessScenario,
} from './builders';

// ---- Player IDs ----
export const HERO_PLAYER_ID = 'hero_player';
export const ALLY_PLAYER_ID = 'ally_1';
export const ENEMY_PLAYER_ID = 'enemy_1';
export const ENEMY_2_PLAYER_ID = 'enemy_2';

// ---- Skill IDs ----
export const HERO_ROAR_SKILL_ID = 'hero_roar';
export const HERO_CALM_MIND_SKILL_ID = 'hero_calm_mind';
export const HERO_FORBIDDEN_POWER_SKILL_ID = 'hero_forbidden_power';
export const HERO_TAUNT_SKILL_ID = 'hero_taunt';
export const HERO_DEAD_DUEL_SKILL_ID = 'hero_dead_duel';

// ---- Hero character definition ----
const heroCharacter = characterView({
  id: 'hero',
  name: '勇者',
  title: '星杯勇者',
  faction: '星杯',
  skills: [
    {
      id: HERO_ROAR_SKILL_ID,
      title: '怒吼',
      description: '主动攻击时发动，移除1点怒气，摸0-1张牌，攻击伤害+2；未命中+1知性',
      type: 0,
      min_targets: 0,
      max_targets: 0,
      target_type: 0,
      cost_gem: 0,
      cost_crystal: 0,
      cost_discards: 0,
    },
    {
      id: HERO_CALM_MIND_SKILL_ID,
      title: '明镜止水',
      description: '主动攻击时发动，消耗4点知性，使攻击无法被应战',
      type: 0,
      min_targets: 0,
      max_targets: 0,
      target_type: 0,
      cost_gem: 0,
      cost_crystal: 0,
      cost_discards: 0,
    },
    {
      id: HERO_FORBIDDEN_POWER_SKILL_ID,
      title: '禁断之力',
      description: '攻击命中或未命中后发动，消耗1个水晶，效果由后端结算',
      type: 0,
      min_targets: 0,
      max_targets: 0,
      target_type: 0,
      cost_gem: 0,
      cost_crystal: 0,
      cost_discards: 0,
    },
    {
      id: HERO_TAUNT_SKILL_ID,
      title: '挑衅',
      description: '消耗1点怒气，选择一名目标对手，使其必须攻击你',
      type: 2,
      min_targets: 1,
      max_targets: 1,
      target_type: 1,
      cost_gem: 0,
      cost_crystal: 0,
      cost_discards: 0,
    },
    {
      id: HERO_DEAD_DUEL_SKILL_ID,
      title: '死斗',
      description: '承受法术伤害时发动，消耗1个红宝石',
      type: 0,
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
function heroHand() {
  return [
    card({ id: 'card_1', name: '火焰斩', type: 'Attack', element: 'Fire' }),
    card({ id: 'card_2', name: '火焰斩', type: 'Attack', element: 'Fire' }),
    card({ id: 'card_3', name: '水涟斩', type: 'Attack', element: 'Water' }),
    card({ id: 'card_4', name: '寒冰箭', type: 'Magic', element: 'Water' }),
    card({ id: 'card_5', name: '圣光', type: 'Magic', element: 'Light' }),
  ];
}

function heroPlayerView(options: {
  is_active?: boolean;
  anger?: number;
  wisdom?: number;
  gems?: number;
  crystals?: number;
} = {}) {
  return playerView({
    id: HERO_PLAYER_ID,
    name: 'E2E Hero',
    camp: 'Red',
    role: 'hero',
    hand: heroHand(),
    hand_count: heroHand().length,
    heal: 2,
    max_heal: 4,
    is_active: options.is_active ?? true,
    gem: options.gems ?? 0,
    crystal: options.crystals ?? 0,
    tokens: {
      hero_anger: options.anger ?? 0,
      hero_wisdom: options.wisdom ?? 0,
    },
  });
}

// ============================================================
// 怒吼 (hero_roar) - Attack response with anger token
// ============================================================

export function roarScenario(options: { anger?: number } = {}): ProtocolHarnessScenario {
  const anger = options.anger ?? 1;
  const characters = [heroCharacter, allyCharacter, enemyCharacter];

  const hero = heroPlayerView({ anger, is_active: true });

  const players = [
    hero,
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
    myPlayerId: HERO_PLAYER_ID,
    myPlayerName: 'E2E Hero',
    characters,
    players: [
      playerInfo({ id: HERO_PLAYER_ID, name: 'E2E Hero', camp: 'Red', char_role: 'hero', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy Bot', camp: 'Blue', char_role: 'villain' }),
    ],
    initialState: syncState({
      turn_player_id: HERO_PLAYER_ID,
      turn_stage: 'ActionExecution',
      available_skills: [],
      characters,
      players,
    }),
  };
}

export function roarConfirmPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: HERO_PLAYER_ID,
    message: '【怒吼】是否发动？（主动攻击时发动，移除1点怒气）',
    choice_type: 'hero_roar_confirm',
    skill_id: HERO_ROAR_SKILL_ID,
    options: [
      { id: '0', label: '发动', button_label: '发动' },
      { id: '1', label: '不发动', button_label: '不发动' },
    ],
    presentation: { kind: 'branch_select', layout: 'overlay', numeric_base: 0 },
    min: 1,
    max: 1,
  } satisfies Prompt);
}

export function roarDrawPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: HERO_PLAYER_ID,
    message: '【怒吼】是否摸1张牌？',
    choice_type: 'hero_roar_draw',
    skill_id: HERO_ROAR_SKILL_ID,
    options: [
      { id: '0', label: '摸1张牌', button_label: '摸1张牌' },
      { id: '1', label: '不摸牌', button_label: '不摸牌' },
    ],
    presentation: { kind: 'branch_select', layout: 'overlay', numeric_base: 0 },
    min: 1,
    max: 1,
  } satisfies Prompt);
}

// ============================================================
// 明镜止水 (hero_calm_mind) - Attack response with 4 wisdom
// ============================================================

export function calmMindScenario(options: { wisdom?: number } = {}): ProtocolHarnessScenario {
  const wisdom = options.wisdom ?? 4;
  const characters = [heroCharacter, allyCharacter, enemyCharacter];

  const hero = heroPlayerView({ wisdom, is_active: true });

  const players = [
    hero,
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
    myPlayerId: HERO_PLAYER_ID,
    myPlayerName: 'E2E Hero',
    characters,
    players: [
      playerInfo({ id: HERO_PLAYER_ID, name: 'E2E Hero', camp: 'Red', char_role: 'hero', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy Bot', camp: 'Blue', char_role: 'villain' }),
    ],
    initialState: syncState({
      turn_player_id: HERO_PLAYER_ID,
      turn_stage: 'ActionExecution',
      available_skills: [],
      characters,
      players,
    }),
  };
}

export function calmMindConfirmPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: HERO_PLAYER_ID,
    message: '【明镜止水】是否发动？消耗4点知性，本次攻击无法被应战',
    choice_type: 'hero_calm_mind_confirm',
    skill_id: HERO_CALM_MIND_SKILL_ID,
    options: [
      { id: '0', label: '发动', button_label: '发动' },
      { id: '1', label: '不发动', button_label: '不发动' },
    ],
    presentation: { kind: 'branch_select', layout: 'overlay', numeric_base: 0 },
    min: 1,
    max: 1,
  } satisfies Prompt);
}

// ============================================================
// 禁断之力 (hero_forbidden_power) - Post-attack response with crystal
// ============================================================

export function forbiddenPowerScenario(options: { crystals?: number } = {}): ProtocolHarnessScenario {
  const crystals = options.crystals ?? 1;
  const characters = [heroCharacter, allyCharacter, enemyCharacter];

  const hero = heroPlayerView({ crystals, is_active: false });

  const players = [
    hero,
    playerView({
      id: ENEMY_PLAYER_ID,
      name: 'Enemy Bot',
      camp: 'Blue',
      role: 'villain',
      hand: [],
      hand_count: 3,
      heal: 1,
      max_heal: 2,
      is_active: true,
    }),
  ];

  return {
    roomCode: 'MOCK',
    myPlayerId: HERO_PLAYER_ID,
    myPlayerName: 'E2E Hero',
    characters,
    players: [
      playerInfo({ id: HERO_PLAYER_ID, name: 'E2E Hero', camp: 'Red', char_role: 'hero', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy Bot', camp: 'Blue', char_role: 'villain' }),
    ],
    initialState: syncState({
      turn_player_id: ENEMY_PLAYER_ID,
      turn_stage: 'AttackResolution',
      available_skills: [],
      characters,
      players,
    }),
  };
}

export function forbiddenPowerConfirmPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: HERO_PLAYER_ID,
    message: '【禁断之力】是否发动？消耗1个水晶',
    choice_type: 'hero_forbidden_power_confirm',
    skill_id: HERO_FORBIDDEN_POWER_SKILL_ID,
    options: [
      { id: '0', label: '发动', button_label: '发动' },
      { id: '1', label: '不发动', button_label: '不发动' },
    ],
    presentation: { kind: 'branch_select', layout: 'overlay', numeric_base: 0 },
    min: 1,
    max: 1,
  } satisfies Prompt);
}

// ============================================================
// 挑衅 (hero_taunt) - Active skill with anger token, target selection
// ============================================================

export function tauntScenario(options: { anger?: number } = {}): ProtocolHarnessScenario {
  const anger = options.anger ?? 1;
  const characters = [heroCharacter, allyCharacter, enemyCharacter];

  const hero = heroPlayerView({ anger, is_active: true });

  const players = [
    hero,
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
    playerView({
      id: ENEMY_2_PLAYER_ID,
      name: 'Enemy Bot 2',
      camp: 'Blue',
      role: 'villain',
      hand: [],
      hand_count: 2,
      heal: 1,
      max_heal: 2,
      is_active: false,
    }),
  ];

  return {
    roomCode: 'MOCK',
    myPlayerId: HERO_PLAYER_ID,
    myPlayerName: 'E2E Hero',
    characters,
    players: [
      playerInfo({ id: HERO_PLAYER_ID, name: 'E2E Hero', camp: 'Red', char_role: 'hero', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy Bot', camp: 'Blue', char_role: 'villain' }),
      playerInfo({ id: ENEMY_2_PLAYER_ID, name: 'Enemy Bot 2', camp: 'Blue', char_role: 'villain' }),
    ],
    initialState: syncState({
      turn_player_id: HERO_PLAYER_ID,
      turn_stage: 'ActionExecution',
      available_skills: [availableSkill({ id: HERO_TAUNT_SKILL_ID, title: '挑衅' })],
      characters,
      players,
    }),
  };
}

export function tauntTargetPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: HERO_PLAYER_ID,
    message: '【挑衅】请选择一名目标对手：',
    choice_type: 'hero_taunt_target',
    skill_id: HERO_TAUNT_SKILL_ID,
    presentation: { kind: 'target_picker', target_filter: 'custom', numeric_base: 0 },
    options: [
      { id: ENEMY_PLAYER_ID, target_id: ENEMY_PLAYER_ID, label: '恶徒', button_label: '选择' },
      { id: ENEMY_2_PLAYER_ID, target_id: ENEMY_2_PLAYER_ID, label: '恶徒2', button_label: '选择' },
    ],
    min: 1,
    max: 1,
  } satisfies Prompt);
}

// ============================================================
// 死斗 (hero_dead_duel) - Magic damage response with gem
// ============================================================

export function deadDuelScenario(options: { gems?: number } = {}): ProtocolHarnessScenario {
  const gems = options.gems ?? 1;
  const characters = [heroCharacter, allyCharacter, enemyCharacter];

  const hero = heroPlayerView({ gems, is_active: false });

  const players = [
    hero,
    playerView({
      id: ENEMY_PLAYER_ID,
      name: 'Enemy Bot',
      camp: 'Blue',
      role: 'villain',
      hand: [],
      hand_count: 3,
      heal: 1,
      max_heal: 2,
      is_active: true,
    }),
  ];

  return {
    roomCode: 'MOCK',
    myPlayerId: HERO_PLAYER_ID,
    myPlayerName: 'E2E Hero',
    characters,
    players: [
      playerInfo({ id: HERO_PLAYER_ID, name: 'E2E Hero', camp: 'Red', char_role: 'hero', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy Bot', camp: 'Blue', char_role: 'villain' }),
    ],
    initialState: syncState({
      turn_player_id: ENEMY_PLAYER_ID,
      turn_stage: 'MagicResolution',
      available_skills: [],
      characters,
      players,
    }),
  };
}

export function deadDuelConfirmPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: HERO_PLAYER_ID,
    message: '【死斗】是否发动？消耗1个红宝石',
    choice_type: 'hero_dead_duel_confirm',
    skill_id: HERO_DEAD_DUEL_SKILL_ID,
    options: [
      { id: '0', label: '发动', button_label: '发动' },
      { id: '1', label: '不发动', button_label: '不发动' },
    ],
    presentation: { kind: 'branch_select', layout: 'overlay', numeric_base: 0 },
    min: 1,
    max: 1,
  } satisfies Prompt);
}
