// ============================================================
// Blood Priestess (血之巫女) Protocol Harness Scenarios
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
  availableSkill,
  type ProtocolHarnessScenario,
} from './builders';

// ---- Player IDs ----
export const BP_PLAYER_ID = 'bp_player';
export const ALLY_PLAYER_ID = 'ally_1';
export const ENEMY_PLAYER_ID = 'enemy_1';
export const ENEMY_2_PLAYER_ID = 'enemy_2';

// ---- Skill IDs ----
export const BP_BLOOD_SORROW_SKILL_ID = 'bp_blood_sorrow';
export const BP_BACKFLOW_SKILL_ID = 'bp_backflow';
export const BP_BLOOD_WAIL_SKILL_ID = 'bp_blood_wail';
export const BP_SHARED_LIFE_SKILL_ID = 'bp_shared_life';
export const BP_BLOOD_CURSE_SKILL_ID = 'bp_blood_curse';

// ---- Blood Priestess character definition ----
const bloodPriestessCharacter = characterView({
  id: 'blood_priestess',
  name: '血之巫女',
  title: '鲜血守护者',
  faction: '星杯',
  skills: [
    {
      id: BP_BLOOD_SORROW_SKILL_ID,
      title: '血之哀伤',
      description: '启动技能，启动阶段分支选择',
      type: 2,
      min_targets: 0,
      max_targets: 0,
      target_type: 0,
      cost_gem: 0,
      cost_crystal: 0,
      cost_discards: 0,
    },
    {
      id: BP_BACKFLOW_SKILL_ID,
      title: '逆流',
      description: '法术技能，确认后弃2张牌',
      type: 2,
      min_targets: 0,
      max_targets: 0,
      target_type: 0,
      cost_gem: 0,
      cost_crystal: 0,
      cost_discards: 0,
    },
    {
      id: BP_BLOOD_WAIL_SKILL_ID,
      title: '血之悲鸣',
      description: '独有法术技能，弃独有卡→目标→X值选择',
      type: 2,
      min_targets: 0,
      max_targets: 0,
      target_type: 0,
      cost_gem: 0,
      cost_crystal: 0,
      cost_discards: 0,
    },
    {
      id: BP_SHARED_LIFE_SKILL_ID,
      title: '同生共死',
      description: '法术技能，确认后目标选择',
      type: 2,
      min_targets: 0,
      max_targets: 0,
      target_type: 0,
      cost_gem: 0,
      cost_crystal: 0,
      cost_discards: 0,
    },
    {
      id: BP_BLOOD_CURSE_SKILL_ID,
      title: '血之诅咒',
      description: '法术技能，确认后目标选择→弃3张牌',
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

const enemy2Character = characterView({
  id: 'enemy_2_char',
  name: '恶徒',
  title: '黑暗使者',
  faction: '异端',
  skills: [],
});

// ---- Helper functions ----
function bloodPriestessHand() {
  return [
    card({ id: 'card_1', name: '火焰斩', type: 'Attack', element: 'Fire' }),
    card({ id: 'card_2', name: '水涟斩', type: 'Attack', element: 'Water' }),
    card({ id: 'card_3', name: '风刃', type: 'Attack', element: 'Wind' }),
    card({ id: 'card_4', name: '血之悲鸣', type: 'Magic', element: 'Dark', exclusive_char1: 'blood_priestess' }),
    card({ id: 'card_5', name: '暗月法术', type: 'Magic', element: 'Dark' }),
  ];
}

function bloodPriestessPlayerView(options: {
  is_active?: boolean;
  heal?: number;
  max_heal?: number;
} = {}) {
  return playerView({
    id: BP_PLAYER_ID,
    name: 'E2E Blood Priestess',
    camp: 'Red',
    role: 'blood_priestess',
    hand: bloodPriestessHand(),
    hand_count: bloodPriestessHand().length,
    heal: options.heal ?? 2,
    max_heal: options.max_heal ?? 4,
    is_active: options.is_active ?? true,
  });
}

// ============================================================
// 血之哀伤 (bp_blood_sorrow) - Startup skill with branch choice
// ============================================================

export function bloodSorrowScenario(): ProtocolHarnessScenario {
  const characters = [bloodPriestessCharacter, allyCharacter, enemyCharacter];

  const bloodPriestess = bloodPriestessPlayerView({ is_active: true });

  const players = [
    bloodPriestess,
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
    myPlayerId: BP_PLAYER_ID,
    myPlayerName: 'E2E Blood Priestess',
    characters,
    players: [
      playerInfo({ id: BP_PLAYER_ID, name: 'E2E Blood Priestess', camp: 'Red', char_role: 'blood_priestess', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy Bot', camp: 'Blue', char_role: 'villain' }),
    ],
    initialState: syncState({
      turn_player_id: BP_PLAYER_ID,
      turn_stage: 'StartupPhase',
      available_skills: [availableSkill({ id: BP_BLOOD_SORROW_SKILL_ID, title: '血之哀伤' })],
      characters,
      players,
    }),
  };
}

export function bloodSorrowBranchPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: BP_PLAYER_ID,
    message: '【血之哀伤】启动阶段，请选择：',
    choice_type: 'bp_blood_sorrow_branch',
    skill_id: BP_BLOOD_SORROW_SKILL_ID,
    options: [
      { id: 'transfer', label: '转移' },
      { id: 'remove', label: '移除' },
      { id: 'skip', label: '跳过' },
    ],
    presentation: { kind: 'branch_select', layout: 'overlay' },
    min: 1,
    max: 1,
  } satisfies Prompt);
}

export function bloodSorrowTargetPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: BP_PLAYER_ID,
    message: '【血之哀伤】请选择目标：',
    choice_type: 'bp_blood_sorrow_target',
    skill_id: BP_BLOOD_SORROW_SKILL_ID,
    options: [
      { id: ENEMY_PLAYER_ID, label: '恶徒' },
    ],
    min: 1,
    max: 1,
  } satisfies Prompt);
}

// ============================================================
// 逆流 (bp_backflow) - Magic skill with discard
// ============================================================

export function backflowScenario(): ProtocolHarnessScenario {
  const characters = [bloodPriestessCharacter, allyCharacter, enemyCharacter];

  const bloodPriestess = bloodPriestessPlayerView({ is_active: true });

  const players = [
    bloodPriestess,
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
    myPlayerId: BP_PLAYER_ID,
    myPlayerName: 'E2E Blood Priestess',
    characters,
    players: [
      playerInfo({ id: BP_PLAYER_ID, name: 'E2E Blood Priestess', camp: 'Red', char_role: 'blood_priestess', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy Bot', camp: 'Blue', char_role: 'villain' }),
    ],
    initialState: syncState({
      turn_player_id: BP_PLAYER_ID,
      turn_stage: 'ActionExecution',
      available_skills: [availableSkill({ id: BP_BACKFLOW_SKILL_ID, title: '逆流' })],
      characters,
      players,
    }),
  };
}

export function backflowConfirmPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: BP_PLAYER_ID,
    message: '【逆流】是否发动？',
    choice_type: 'bp_backflow_confirm',
    skill_id: BP_BACKFLOW_SKILL_ID,
    options: [
      { id: '0', label: '发动' },
      { id: '1', label: '不发动' },
    ],
    min: 1,
    max: 1,
  } satisfies Prompt);
}

export function backflowDiscardPrompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_cards',
    player_id: BP_PLAYER_ID,
    message: '【逆流】请选择弃置2张牌：',
    choice_type: 'bp_backflow_discard',
    skill_id: BP_BACKFLOW_SKILL_ID,
    options: [
      { id: '0', label: '1: 火焰斩 (火 Attack)' },
      { id: '1', label: '2: 水涟斩 (水 Attack)' },
      { id: '2', label: '3: 风刃 (风 Attack)' },
    ],
    min: 2,
    max: 2,
  } satisfies Prompt);
}

// ============================================================
// 血之悲鸣 (bp_blood_wail) - Unique magic skill with discard unique card, target, X value
// ============================================================

export function bloodWailScenario(): ProtocolHarnessScenario {
  const characters = [bloodPriestessCharacter, allyCharacter, enemyCharacter, enemy2Character];

  const bloodPriestess = bloodPriestessPlayerView({ is_active: true });

  const players = [
    bloodPriestess,
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
    myPlayerId: BP_PLAYER_ID,
    myPlayerName: 'E2E Blood Priestess',
    characters,
    players: [
      playerInfo({ id: BP_PLAYER_ID, name: 'E2E Blood Priestess', camp: 'Red', char_role: 'blood_priestess', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy Bot', camp: 'Blue', char_role: 'villain' }),
      playerInfo({ id: ENEMY_2_PLAYER_ID, name: 'Enemy Bot 2', camp: 'Blue', char_role: 'villain' }),
    ],
    initialState: syncState({
      turn_player_id: BP_PLAYER_ID,
      turn_stage: 'ActionExecution',
      available_skills: [availableSkill({ id: BP_BLOOD_WAIL_SKILL_ID, title: '血之悲鸣' })],
      characters,
      players,
    }),
  };
}

export function bloodWailConfirmPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: BP_PLAYER_ID,
    message: '【血之悲鸣】是否发动？（弃置独有法术牌）',
    choice_type: 'bp_blood_wail_confirm',
    skill_id: BP_BLOOD_WAIL_SKILL_ID,
    options: [
      { id: '0', label: '发动' },
      { id: '1', label: '不发动' },
    ],
    min: 1,
    max: 1,
  } satisfies Prompt);
}

export function bloodWailUniqueCardPrompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_card',
    player_id: BP_PLAYER_ID,
    message: '【血之悲鸣】请选择弃置独有法术牌：',
    choice_type: 'bp_blood_wail_unique_card',
    skill_id: BP_BLOOD_WAIL_SKILL_ID,
    options: [
      { id: '3', label: '4: 血之悲鸣 (暗 Magic)' },
    ],
    min: 1,
    max: 1,
  } satisfies Prompt);
}

export function bloodWailTargetPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: BP_PLAYER_ID,
    message: '【血之悲鸣】请选择目标：',
    choice_type: 'bp_blood_wail_target',
    skill_id: BP_BLOOD_WAIL_SKILL_ID,
    options: [
      { id: ENEMY_PLAYER_ID, label: '恶徒' },
      { id: ENEMY_2_PLAYER_ID, label: '恶徒2' },
    ],
    min: 1,
    max: 1,
  } satisfies Prompt);
}

export function bloodWailXPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: BP_PLAYER_ID,
    message: '【血之悲鸣】请选择X值（X<3）：',
    choice_type: 'bp_blood_wail_x',
    skill_id: BP_BLOOD_WAIL_SKILL_ID,
    options: [
      { id: '1', label: 'X=1' },
      { id: '2', label: 'X=2' },
    ],
    presentation: { kind: 'branch_select', layout: 'overlay' },
    min: 1,
    max: 1,
  } satisfies Prompt);
}

// ============================================================
// 同生共死 (bp_shared_life) - Magic skill with target selection
// ============================================================

export function sharedLifeScenario(): ProtocolHarnessScenario {
  const characters = [bloodPriestessCharacter, allyCharacter, enemyCharacter, enemy2Character];

  const bloodPriestess = bloodPriestessPlayerView({ is_active: true });

  const players = [
    bloodPriestess,
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
    myPlayerId: BP_PLAYER_ID,
    myPlayerName: 'E2E Blood Priestess',
    characters,
    players: [
      playerInfo({ id: BP_PLAYER_ID, name: 'E2E Blood Priestess', camp: 'Red', char_role: 'blood_priestess', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy Bot', camp: 'Blue', char_role: 'villain' }),
      playerInfo({ id: ENEMY_2_PLAYER_ID, name: 'Enemy Bot 2', camp: 'Blue', char_role: 'villain' }),
    ],
    initialState: syncState({
      turn_player_id: BP_PLAYER_ID,
      turn_stage: 'ActionExecution',
      available_skills: [availableSkill({ id: BP_SHARED_LIFE_SKILL_ID, title: '同生共死' })],
      characters,
      players,
    }),
  };
}

export function sharedLifeConfirmPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: BP_PLAYER_ID,
    message: '【同生共死】是否发动？',
    choice_type: 'bp_shared_life_confirm',
    skill_id: BP_SHARED_LIFE_SKILL_ID,
    options: [
      { id: '0', label: '发动' },
      { id: '1', label: '不发动' },
    ],
    min: 1,
    max: 1,
  } satisfies Prompt);
}

export function sharedLifeTargetPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: BP_PLAYER_ID,
    message: '【同生共死】请选择目标：',
    choice_type: 'bp_shared_life_target',
    skill_id: BP_SHARED_LIFE_SKILL_ID,
    options: [
      { id: ENEMY_PLAYER_ID, label: '恶徒' },
      { id: ENEMY_2_PLAYER_ID, label: '恶徒2' },
    ],
    min: 1,
    max: 1,
  } satisfies Prompt);
}

// ============================================================
// 血之诅咒 (bp_blood_curse) - Magic skill with target then discard 3 cards
// ============================================================

export function bloodCurseScenario(): ProtocolHarnessScenario {
  const characters = [bloodPriestessCharacter, allyCharacter, enemyCharacter, enemy2Character];

  const bloodPriestess = bloodPriestessPlayerView({ is_active: true });

  const players = [
    bloodPriestess,
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
    myPlayerId: BP_PLAYER_ID,
    myPlayerName: 'E2E Blood Priestess',
    characters,
    players: [
      playerInfo({ id: BP_PLAYER_ID, name: 'E2E Blood Priestess', camp: 'Red', char_role: 'blood_priestess', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy Bot', camp: 'Blue', char_role: 'villain' }),
      playerInfo({ id: ENEMY_2_PLAYER_ID, name: 'Enemy Bot 2', camp: 'Blue', char_role: 'villain' }),
    ],
    initialState: syncState({
      turn_player_id: BP_PLAYER_ID,
      turn_stage: 'ActionExecution',
      available_skills: [availableSkill({ id: BP_BLOOD_CURSE_SKILL_ID, title: '血之诅咒' })],
      characters,
      players,
    }),
  };
}

export function bloodCurseConfirmPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: BP_PLAYER_ID,
    message: '【血之诅咒】是否发动？',
    choice_type: 'bp_blood_curse_confirm',
    skill_id: BP_BLOOD_CURSE_SKILL_ID,
    options: [
      { id: '0', label: '发动' },
      { id: '1', label: '不发动' },
    ],
    min: 1,
    max: 1,
  } satisfies Prompt);
}

export function bloodCurseTargetPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: BP_PLAYER_ID,
    message: '【血之诅咒】请选择目标：',
    choice_type: 'bp_blood_curse_target',
    skill_id: BP_BLOOD_CURSE_SKILL_ID,
    options: [
      { id: ENEMY_PLAYER_ID, label: '恶徒' },
      { id: ENEMY_2_PLAYER_ID, label: '恶徒2' },
    ],
    min: 1,
    max: 1,
  } satisfies Prompt);
}

export function bloodCurseDiscardPrompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_cards',
    player_id: BP_PLAYER_ID,
    message: '【血之诅咒】请选择弃置3张牌：',
    choice_type: 'bp_blood_curse_discard',
    skill_id: BP_BLOOD_CURSE_SKILL_ID,
    options: [
      { id: '0', label: '1: 火焰斩 (火 Attack)' },
      { id: '1', label: '2: 水涟斩 (水 Attack)' },
      { id: '2', label: '3: 风刃 (风 Attack)' },
      { id: '3', label: '4: 血之悲鸣 (暗 Magic)' },
      { id: '4', label: '5: 暗月法术 (暗 Magic)' },
    ],
    min: 3,
    max: 3,
  } satisfies Prompt);
}