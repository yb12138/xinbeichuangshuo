// ============================================================
// Sword Emperor (剑帝) Protocol Harness Scenarios
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
export const SE_PLAYER_ID = 'se_player';
export const ALLY_PLAYER_ID = 'ally_1';
export const ENEMY_PLAYER_ID = 'enemy_1';
export const ENEMY_2_PLAYER_ID = 'enemy_2';

// ---- Skill IDs ----
export const SE_SWORD_QI_SLASH_SKILL_ID = 'se_sword_qi_slash';
export const SE_ANGEL_SOUL_SKILL_ID = 'se_angel_soul';
export const SE_DEMON_SOUL_SKILL_ID = 'se_demon_soul';

// ---- Sword Emperor character definition ----
const swordEmperorCharacter = characterView({
  id: 'sword_emperor',
  name: '剑帝',
  title: '剑气之主',
  faction: '星杯',
  skills: [
    {
      id: SE_SWORD_QI_SLASH_SKILL_ID,
      title: '剑气斩',
      description: '响应技能，攻击命中后发动，移除剑气点数后选择目标',
      type: 0,
      min_targets: 0,
      max_targets: 0,
      target_type: 0,
      cost_gem: 0,
      cost_crystal: 0,
      cost_discards: 0,
    },
    {
      id: SE_ANGEL_SOUL_SKILL_ID,
      title: '天使之魂',
      description: '响应技能，攻击前发动',
      type: 0,
      min_targets: 0,
      max_targets: 0,
      target_type: 0,
      cost_gem: 0,
      cost_crystal: 0,
      cost_discards: 0,
    },
    {
      id: SE_DEMON_SOUL_SKILL_ID,
      title: '恶魔之魂',
      description: '响应技能，攻击前发动',
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

const enemy2Character = characterView({
  id: 'enemy_2_char',
  name: '恶徒',
  title: '黑暗使者',
  faction: '异端',
  skills: [],
});

// ---- Helper functions ----
function swordEmperorHand() {
  return [
    card({ id: 'card_1', name: '火焰斩', type: 'Attack', element: 'Fire' }),
    card({ id: 'card_2', name: '水涟斩', type: 'Attack', element: 'Water' }),
    card({ id: 'card_3', name: '风刃', type: 'Attack', element: 'Wind' }),
  ];
}

function swordEmperorPlayerView(options: {
  is_active?: boolean;
  heal?: number;
  max_heal?: number;
  sword_qi?: number;
} = {}) {
  return playerView({
    id: SE_PLAYER_ID,
    name: 'E2E Sword Emperor',
    camp: 'Red',
    role: 'sword_emperor',
    hand: swordEmperorHand(),
    hand_count: swordEmperorHand().length,
    heal: options.heal ?? 2,
    max_heal: options.max_heal ?? 4,
    is_active: options.is_active ?? true,
    tokens: {
      se_sword_qi: options.sword_qi ?? 0,
    },
  });
}

// ============================================================
// 剑气斩 (se_sword_qi_slash) - Response skill after attack hit
// ============================================================

export function swordQiSlashScenario(options: { sword_qi?: number; attacker_id?: string } = {}): ProtocolHarnessScenario {
  const sword_qi = options.sword_qi ?? 3;
  const attacker_id = options.attacker_id ?? SE_PLAYER_ID;
  const characters = [swordEmperorCharacter, allyCharacter, enemyCharacter, enemy2Character];

  const swordEmperor = swordEmperorPlayerView({ sword_qi, is_active: attacker_id === SE_PLAYER_ID });

  const players = [
    swordEmperor,
    playerView({
      id: ENEMY_PLAYER_ID,
      name: 'Enemy Bot',
      camp: 'Blue',
      role: 'villain',
      hand: [],
      hand_count: 3,
      heal: 1,
      max_heal: 2,
      is_active: attacker_id !== SE_PLAYER_ID,
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
    myPlayerId: SE_PLAYER_ID,
    myPlayerName: 'E2E Sword Emperor',
    characters,
    players: [
      playerInfo({ id: SE_PLAYER_ID, name: 'E2E Sword Emperor', camp: 'Red', char_role: 'sword_emperor', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy Bot', camp: 'Blue', char_role: 'villain' }),
      playerInfo({ id: ENEMY_2_PLAYER_ID, name: 'Enemy Bot 2', camp: 'Blue', char_role: 'villain' }),
    ],
    initialState: syncState({
      turn_player_id: attacker_id,
      turn_stage: 'ActionExecution',
      available_skills: [],
      characters,
      players,
    }),
  };
}

export function swordQiSlashConfirmPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: SE_PLAYER_ID,
    message: '【剑气斩】攻击命中，是否发动？',
    choice_type: 'se_sword_qi_slash_confirm',
    skill_id: SE_SWORD_QI_SLASH_SKILL_ID,
    options: [
      { id: '0', label: '发动' },
      { id: '1', label: '不发动' },
    ],
    min: 1,
    max: 1,
  } satisfies Prompt);
}

export function swordQiSlashRemovePrompt(xMax: number): WsMessage {
  const options: { id: string; label: string }[] = [];
  for (let i = 1; i <= xMax; i++) {
    options.push({ id: `${i}`, label: `移除${i}点剑气` });
  }
  return requireActionMessage({
    type: 'confirm',
    player_id: SE_PLAYER_ID,
    message: `【剑气斩】请选择移除X点剑气（1≤X≤${xMax}）：`,
    choice_type: 'se_sword_qi_slash_remove',
    skill_id: SE_SWORD_QI_SLASH_SKILL_ID,
    options,
    presentation: { kind: 'branch_select', layout: 'overlay' },
    min: 1,
    max: 1,
  } satisfies Prompt);
}

export function swordQiSlashTargetPrompt(xValue: number): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: SE_PLAYER_ID,
    message: `【剑气斩】请选择除攻击目标外的任意一名其他角色（造成${xValue}点法术伤害）：`,
    choice_type: 'se_sword_qi_slash_target',
    skill_id: SE_SWORD_QI_SLASH_SKILL_ID,
    options: [
      { id: ENEMY_2_PLAYER_ID, label: '恶徒2' },
    ],
    min: 1,
    max: 1,
  } satisfies Prompt);
}

// ============================================================
// 天使之魂 (se_angel_soul) - Response skill before attack
// ============================================================

export function angelSoulScenario(): ProtocolHarnessScenario {
  const characters = [swordEmperorCharacter, allyCharacter, enemyCharacter];

  const swordEmperor = swordEmperorPlayerView({ is_active: true });

  const players = [
    swordEmperor,
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
    myPlayerId: SE_PLAYER_ID,
    myPlayerName: 'E2E Sword Emperor',
    characters,
    players: [
      playerInfo({ id: SE_PLAYER_ID, name: 'E2E Sword Emperor', camp: 'Red', char_role: 'sword_emperor', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy Bot', camp: 'Blue', char_role: 'villain' }),
    ],
    initialState: syncState({
      turn_player_id: SE_PLAYER_ID,
      turn_stage: 'ActionExecution',
      available_skills: [],
      characters,
      players,
    }),
  };
}

export function angelSoulConfirmPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: SE_PLAYER_ID,
    message: '【天使之魂】攻击前，是否发动？',
    choice_type: 'se_angel_soul_confirm',
    skill_id: SE_ANGEL_SOUL_SKILL_ID,
    options: [
      { id: '0', label: '发动' },
      { id: '1', label: '不发动' },
    ],
    min: 1,
    max: 1,
  } satisfies Prompt);
}

// ============================================================
// 恶魔之魂 (se_demon_soul) - Response skill before attack
// ============================================================

export function demonSoulScenario(): ProtocolHarnessScenario {
  const characters = [swordEmperorCharacter, allyCharacter, enemyCharacter];

  const swordEmperor = swordEmperorPlayerView({ is_active: true });

  const players = [
    swordEmperor,
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
    myPlayerId: SE_PLAYER_ID,
    myPlayerName: 'E2E Sword Emperor',
    characters,
    players: [
      playerInfo({ id: SE_PLAYER_ID, name: 'E2E Sword Emperor', camp: 'Red', char_role: 'sword_emperor', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy Bot', camp: 'Blue', char_role: 'villain' }),
    ],
    initialState: syncState({
      turn_player_id: SE_PLAYER_ID,
      turn_stage: 'ActionExecution',
      available_skills: [],
      characters,
      players,
    }),
  };
}

export function demonSoulConfirmPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: SE_PLAYER_ID,
    message: '【恶魔之魂】攻击前，是否发动？',
    choice_type: 'se_demon_soul_confirm',
    skill_id: SE_DEMON_SOUL_SKILL_ID,
    options: [
      { id: '0', label: '发动' },
      { id: '1', label: '不发动' },
    ],
    min: 1,
    max: 1,
  } satisfies Prompt);
}