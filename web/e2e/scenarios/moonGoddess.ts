// ============================================================
// Moon Goddess (月之女神) Protocol Harness Scenarios
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
export const MG_PLAYER_ID = 'mg_player';
export const ENEMY_PLAYER_ID = 'enemy_1';
export const ENEMY_2_PLAYER_ID = 'enemy_2';

// ---- Skill IDs ----
export const MG_NEW_MOON_SHelter_SKILL_ID = 'mg_new_moon_shelter';
export const MG_MEDUSA_EYE_SKILL_ID = 'mg_medusa_eye';
export const MG_MOON_CYCLE_SKILL_ID = 'mg_moon_cycle';
export const MG_MOON_READ_SKILL_ID = 'mg_blasphemy';
export const MG_DARKMOON_SLASH_SKILL_ID = 'mg_darkmoon_slash';
export const MG_PALE_MOON_SKILL_ID = 'mg_pale_moon';

// ---- Moon Goddess character definition ----
const moonGoddessCharacter = characterView({
  id: 'moon_goddess',
  name: '月之女神',
  title: '月光守护者',
  faction: '星杯',
  skills: [
    {
      id: MG_NEW_MOON_SHelter_SKILL_ID,
      title: '新月庇护',
      description: '响应技能，触发条件时弹出确认',
      type: 0,
      min_targets: 0,
      max_targets: 0,
      target_type: 0,
      cost_gem: 0,
      cost_crystal: 0,
      cost_discards: 0,
    },
    {
      id: MG_MEDUSA_EYE_SKILL_ID,
      title: '美杜莎之眼',
      description: '响应技能，闇月选择→弃牌→目标选择',
      type: 0,
      min_targets: 0,
      max_targets: 0,
      target_type: 0,
      cost_gem: 0,
      cost_crystal: 0,
      cost_discards: 0,
    },
    {
      id: MG_MOON_CYCLE_SKILL_ID,
      title: '月之轮回',
      description: '响应技能，回合结束时触发分支选择',
      type: 0,
      min_targets: 0,
      max_targets: 0,
      target_type: 0,
      cost_gem: 0,
      cost_crystal: 0,
      cost_discards: 0,
    },
    {
      id: MG_MOON_READ_SKILL_ID,
      title: '月渎',
      description: '响应技能，法术伤害后摸牌触发',
      type: 0,
      min_targets: 0,
      max_targets: 0,
      target_type: 0,
      cost_gem: 0,
      cost_crystal: 0,
      cost_discards: 0,
    },
    {
      id: MG_DARKMOON_SLASH_SKILL_ID,
      title: '闇月斩',
      description: '响应技能，确认→选择X值',
      type: 0,
      min_targets: 0,
      max_targets: 0,
      target_type: 0,
      cost_gem: 0,
      cost_crystal: 0,
      cost_discards: 0,
    },
    {
      id: MG_PALE_MOON_SKILL_ID,
      title: '苍白之月',
      description: '响应技能，确认→分支选择',
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

const enemyCharacter = characterView({
  id: 'enemy_char',
  name: '魔神',
  title: '暗影之王',
  faction: '异端',
  skills: [],
});

const enemy2Character = characterView({
  id: 'enemy_2_char',
  name: '恶徒',
  title: '黑暗使者',
  faction: '异端',
  skills: [],
});

// ---- Helper functions ----
function moonGoddessHand() {
  return [
    card({ id: 'card_1', name: '火焰斩', type: 'Attack', element: 'Fire' }),
    card({ id: 'card_2', name: '水涟斩', type: 'Attack', element: 'Water' }),
    card({ id: 'card_3', name: '暗月法术', type: 'Magic', element: 'Dark' }),
    card({ id: 'card_4', name: '圣光', type: 'Magic', element: 'Light' }),
  ];
}

function moonGoddessPlayerView(options: {
  is_active?: boolean;
  heal?: number;
  max_heal?: number;
  dark_moon_cards?: number;
  new_moon_tokens?: number;
  petrify_tokens?: number;
} = {}) {
  return playerView({
    id: MG_PLAYER_ID,
    name: 'E2E Moon Goddess',
    camp: 'Red',
    role: 'moon_goddess',
    hand: moonGoddessHand(),
    hand_count: moonGoddessHand().length,
    heal: options.heal ?? 2,
    max_heal: options.max_heal ?? 4,
    is_active: options.is_active ?? true,
    tokens: {
      mg_dark_moon: options.dark_moon_cards ?? 0,
      mg_new_moon: options.new_moon_tokens ?? 0,
      mg_petrify: options.petrify_tokens ?? 0,
    },
  });
}

// ============================================================
// 新月庇护 (mg_new_moon_shelter) - 后端通过 response_skills 自动触发
// ============================================================

export function newMoonShelterScenario(): ProtocolHarnessScenario {
  const characters = [moonGoddessCharacter, enemyCharacter];

  const moonGoddess = moonGoddessPlayerView({ is_active: false });

  const players = [
    moonGoddess,
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
    myPlayerId: MG_PLAYER_ID,
    myPlayerName: 'E2E Moon Goddess',
    characters,
    players: [
      playerInfo({ id: MG_PLAYER_ID, name: 'E2E Moon Goddess', camp: 'Red', char_role: 'moon_goddess', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy Bot', camp: 'Blue', char_role: 'villain' }),
    ],
    initialState: syncState({
      turn_player_id: ENEMY_PLAYER_ID,
      turn_stage: 'ActionExecution',
      available_skills: [],
      characters,
      players,
      // 后端会设置 response_skills 触发确认弹框
    }),
  };
}

// ============================================================
// 美杜莎之眼 (mg_medusa_eye) - 后端通过 response_skills 触发
// 闇月选择使用 choice_type: mg_medusa_darkmoon_pick
// 弃牌通过 system_discard_cards，目标通过 min_targets 处理
// ============================================================

export function medusaEyeScenario(options: { dark_moon_cards?: number } = {}): ProtocolHarnessScenario {
  const dark_moon_cards = options.dark_moon_cards ?? 2;
  const characters = [moonGoddessCharacter, enemyCharacter, enemy2Character];

  const moonGoddess = moonGoddessPlayerView({ dark_moon_cards, is_active: true });

  const players = [
    moonGoddess,
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
    myPlayerId: MG_PLAYER_ID,
    myPlayerName: 'E2E Moon Goddess',
    characters,
    players: [
      playerInfo({ id: MG_PLAYER_ID, name: 'E2E Moon Goddess', camp: 'Red', char_role: 'moon_goddess', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy Bot', camp: 'Blue', char_role: 'villain' }),
      playerInfo({ id: ENEMY_2_PLAYER_ID, name: 'Enemy Bot 2', camp: 'Blue', char_role: 'villain' }),
    ],
    initialState: syncState({
      turn_player_id: MG_PLAYER_ID,
      turn_stage: 'ActionExecution',
      available_skills: [],
      characters,
      players,
    }),
  };
}

export function medusaEyeDarkMoonPrompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_card',
    player_id: MG_PLAYER_ID,
    message: '【美杜莎之眼】请选择一张闇月牌（盖牌）：',
    choice_type: 'mg_medusa_darkmoon_pick',
    skill_id: MG_MEDUSA_EYE_SKILL_ID,
    options: [
      { id: '0', label: '1: 暗月法术 (暗 Magic)' },
      { id: '1', label: '2: 火焰斩 (火 Attack)' },
    ],
    min: 1,
    max: 1,
  } satisfies Prompt);
}

// ============================================================
// 月之轮回 (mg_moon_cycle) - Response skill at turn end with branch choice
// ============================================================

export function moonCycleScenario(options: { dark_moon_cards?: number; heal?: number } = {}): ProtocolHarnessScenario {
  const dark_moon_cards = options.dark_moon_cards ?? 1;
  const heal = options.heal ?? 2;
  const characters = [moonGoddessCharacter, enemyCharacter];

  const moonGoddess = moonGoddessPlayerView({ dark_moon_cards, heal, is_active: false });

  const players = [
    moonGoddess,
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
    myPlayerId: MG_PLAYER_ID,
    myPlayerName: 'E2E Moon Goddess',
    characters,
    players: [
      playerInfo({ id: MG_PLAYER_ID, name: 'E2E Moon Goddess', camp: 'Red', char_role: 'moon_goddess', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy Bot', camp: 'Blue', char_role: 'villain' }),
    ],
    initialState: syncState({
      turn_player_id: ENEMY_PLAYER_ID,
      turn_stage: 'EndPhase',
      available_skills: [],
      characters,
      players,
    }),
  };
}

export function moonCycleBranchPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: MG_PLAYER_ID,
    message: '【月之轮回】回合结束，请选择分支：',
    choice_type: 'mg_moon_cycle_mode',
    skill_id: MG_MOON_CYCLE_SKILL_ID,
    options: [
      { id: 'branch1', label: '分支一：移除1个闇月，目标+1治疗' },
      { id: 'branch2', label: '分支二：移除1个治疗，+1新月' },
    ],
    presentation: { kind: 'branch_select', layout: 'overlay' },
    min: 1,
    max: 1,
  } satisfies Prompt);
}

export function moonCycleTargetPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: MG_PLAYER_ID,
    message: '【月之轮回】分支一：请选择目标令其+1治疗：',
    choice_type: 'mg_moon_cycle_heal_target',
    skill_id: MG_MOON_CYCLE_SKILL_ID,
    options: [
      { id: ENEMY_PLAYER_ID, label: '恶徒' },
    ],
    min: 1,
    max: 1,
  } satisfies Prompt);
}

// ============================================================
// 月渎 (mg_blasphemy) - 后端通过 response_skills 自动触发
// 目标通过 min_targets 处理
// ============================================================

export function moonReadScenario(options: { heal?: number } = {}): ProtocolHarnessScenario {
  const heal = options.heal ?? 2;
  const characters = [moonGoddessCharacter, enemyCharacter, enemy2Character];

  const moonGoddess = moonGoddessPlayerView({ heal, is_active: true });

  const players = [
    moonGoddess,
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
    myPlayerId: MG_PLAYER_ID,
    myPlayerName: 'E2E Moon Goddess',
    characters,
    players: [
      playerInfo({ id: MG_PLAYER_ID, name: 'E2E Moon Goddess', camp: 'Red', char_role: 'moon_goddess', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy Bot', camp: 'Blue', char_role: 'villain' }),
      playerInfo({ id: ENEMY_2_PLAYER_ID, name: 'Enemy Bot 2', camp: 'Blue', char_role: 'villain' }),
    ],
    initialState: syncState({
      turn_player_id: MG_PLAYER_ID,
      turn_stage: 'ActionExecution',
      available_skills: [],
      characters,
      players,
      // 后端会设置 response_skills 触发确认弹框
    }),
  };
}

// ============================================================
// 闇月斩 (mg_darkmoon_slash) - 后端通过 response_skills 自动触发
// ============================================================

export function darkmoonSlashScenario(): ProtocolHarnessScenario {
  const characters = [moonGoddessCharacter, enemyCharacter];

  const moonGoddess = moonGoddessPlayerView({ is_active: true });

  const players = [
    moonGoddess,
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
    myPlayerId: MG_PLAYER_ID,
    myPlayerName: 'E2E Moon Goddess',
    characters,
    players: [
      playerInfo({ id: MG_PLAYER_ID, name: 'E2E Moon Goddess', camp: 'Red', char_role: 'moon_goddess', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy Bot', camp: 'Blue', char_role: 'villain' }),
    ],
    initialState: syncState({
      turn_player_id: MG_PLAYER_ID,
      turn_stage: 'ActionExecution',
      available_skills: [],
      characters,
      players,
      // 后端会设置 response_skills 触发确认弹框
    }),
  };
}

// ============================================================
// 苍白之月 (mg_pale_moon) - 后端通过 response_skills 自动触发
// 分支选择使用 choice_type: mg_pale_moon_mode
// X值选择使用 choice_type: mg_pale_moon_x
// 弃牌通过 system_discard_cards，目标通过 min_targets 处理
// ============================================================

export function paleMoonScenario(options: { petrify_tokens?: number; new_moon_tokens?: number } = {}): ProtocolHarnessScenario {
  const petrify_tokens = options.petrify_tokens ?? 3;
  const new_moon_tokens = options.new_moon_tokens ?? 2;
  const characters = [moonGoddessCharacter, enemyCharacter, enemy2Character];

  const moonGoddess = moonGoddessPlayerView({ petrify_tokens, new_moon_tokens, is_active: true });

  const players = [
    moonGoddess,
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
    myPlayerId: MG_PLAYER_ID,
    myPlayerName: 'E2E Moon Goddess',
    characters,
    players: [
      playerInfo({ id: MG_PLAYER_ID, name: 'E2E Moon Goddess', camp: 'Red', char_role: 'moon_goddess', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy Bot', camp: 'Blue', char_role: 'villain' }),
      playerInfo({ id: ENEMY_2_PLAYER_ID, name: 'Enemy Bot 2', camp: 'Blue', char_role: 'villain' }),
    ],
    initialState: syncState({
      turn_player_id: MG_PLAYER_ID,
      turn_stage: 'ActionExecution',
      available_skills: [],
      characters,
      players,
      // 后端会设置 response_skills 触发确认弹框
    }),
  };
}

export function paleMoonBranchPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: MG_PLAYER_ID,
    message: '【苍白之月】请选择分支：',
    choice_type: 'mg_pale_moon_mode',
    skill_id: MG_PALE_MOON_SKILL_ID,
    options: [
      { id: 'branch1', label: '分支一：移除3点石化，下次攻击无法应战，+1攻击行动，额外回合' },
      { id: 'branch2', label: '分支二：移除X点新月，+1石化，弃牌，对目标造成X+1法术伤害' },
    ],
    presentation: { kind: 'branch_select', layout: 'overlay' },
    min: 1,
    max: 1,
  } satisfies Prompt);
}

export function paleMoonXPrompt(xMax: number): WsMessage {
  const options: { id: string; label: string }[] = [];
  for (let i = 1; i <= xMax; i++) {
    options.push({ id: `${i}`, label: `移除${i}点新月` });
  }
  return requireActionMessage({
    type: 'confirm',
    player_id: MG_PLAYER_ID,
    message: `【苍白之月】分支二：请选择移除X点新月（1≤X≤${xMax}）：`,
    choice_type: 'mg_pale_moon_x',
    skill_id: MG_PALE_MOON_SKILL_ID,
    options,
    presentation: { kind: 'branch_select', layout: 'overlay' },
    min: 1,
    max: 1,
  } satisfies Prompt);
}