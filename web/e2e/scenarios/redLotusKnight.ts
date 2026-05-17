// ============================================================
// RedLotusKnight (红莲骑士) Protocol Harness Scenarios
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

export const RED_LOTUS_KNIGHT_PLAYER_ID = 'red_lotus_knight_player';
export const ENEMY_PLAYER_ID = 'enemy_1';
export const ALLY_PLAYER_ID = 'ally_1';

export const RED_LOTUS_KNIGHT_SCARLET_COVENANT_ID = 'red_lotus_knight_scarlet_covenant';
export const RED_LOTUS_KNIGHT_BLOODY_PRAYER_ID = 'red_lotus_knight_bloody_prayer';
export const RED_LOTUS_KNIGHT_SLAUGHTER_FEAST_ID = 'red_lotus_knight_slaughter_feast';
export const RED_LOTUS_KNIGHT_HOT_BLOOD_BOILING_ID = 'red_lotus_knight_hot_blood_boiling';
export const RED_LOTUS_KNIGHT_MODESTY_ID = 'red_lotus_knight_modesty';
export const RED_LOTUS_KNIGHT_SCARLET_CROSS_ID = 'red_lotus_knight_scarlet_cross';

const redLotusKnightCharacter = characterView({
  id: 'red_lotus_knight',
  name: '红莲骑士',
  title: '咏',
  faction: '咏',
  skills: [
    {
      id: RED_LOTUS_KNIGHT_SCARLET_COVENANT_ID,
      title: '腥红圣约',
      description: '（攻击时发动）本次攻击伤害+1，攻击结束后你+1［治疗］。',
      type: 3, // 响应
      min_targets: 0, max_targets: 0, target_type: 0,
    },
    {
      id: RED_LOTUS_KNIGHT_BLOODY_PRAYER_ID,
      title: '血腥祷言',
      description: '（移除自己X点［治疗］，X最大为2）将等量的［血印］放置于一名对手面前。',
      type: 1, // 启动
      min_targets: 0, max_targets: 0, target_type: 0,
    },
    {
      id: RED_LOTUS_KNIGHT_SLAUGHTER_FEAST_ID,
      title: '杀戮盛宴',
      description: '（命中后发动）移除目标1［血印］，对自己造成1点法术伤害③，本次攻击伤害+2。',
      type: 3, // 响应
      min_targets: 0, max_targets: 0, target_type: 0,
    },
    {
      id: RED_LOTUS_KNIGHT_HOT_BLOOD_BOILING_ID,
      title: '热血沸腾',
      description: '（士气下降时自动进入热血形态）回合结束时脱离热血形态。',
      type: 3, // 被动响应
      min_targets: 0, max_targets: 0, target_type: 0,
    },
    {
      id: RED_LOTUS_KNIGHT_MODESTY_ID,
      title: '戒骄戒躁',
      description: '［水晶］（热血形态中）脱离热血形态，选择：Ⅰ、摸2张牌；Ⅱ、+2［治疗］。',
      type: 3, // 响应(大招)
      min_targets: 0, max_targets: 0, target_type: 0,
    },
    {
      id: RED_LOTUS_KNIGHT_SCARLET_CROSS_ID,
      title: '腥红十字',
      description: '［水晶］移除目标1［血印］，弃1张法术牌［展示］，对目标造成2点法术伤害③。',
      type: 2, // 法术(大招)
      min_targets: 1, max_targets: 1, target_type: 2,
    },
  ],
});

const enemyCharacter = characterView({
  id: 'enemy_char', name: '守卫', title: '测试目标', faction: '异端', skills: [],
});

const allyCharacter = characterView({
  id: 'ally_char', name: '勇者', title: '测试队友', faction: '咏', skills: [],
});

const defaultCharacters = [redLotusKnightCharacter, enemyCharacter, allyCharacter];

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function redLotusKnightHand(): Card[] {
  return [
    card({ id: 'rlk-attack-1', name: '血刃', type: 'Attack', element: 'Fire' }),
    card({ id: 'rlk-attack-2', name: '红莲斩', type: 'Attack', element: 'Fire' }),
    card({ id: 'rlk-magic-1', name: '血术', type: 'Magic', element: 'Fire' }),
    card({ id: 'rlk-magic-2', name: '红莲', type: 'Magic', element: 'Fire' }),
    card({ id: 'rlk-water-magic', name: '治愈', type: 'Magic', element: 'Water' }),
  ];
}

function redLotusKnightAvailableSkill(skill: Partial<AvailableSkill> & { id: string; title: string }): AvailableSkill {
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

export function redLotusKnightScenario(options: {
  hand?: Card[];
  availableSkills?: AvailableSkill[];
  crystal?: number;
  gem?: number;
  turnStage?: string;
  heal?: number;
  buffs?: { id: string; name: string; duration: number; value: number; source_id: string }[];
} = {}): ProtocolHarnessScenario {
  const hand = options.hand ?? redLotusKnightHand();
  const players = [
    playerView({
      id: RED_LOTUS_KNIGHT_PLAYER_ID,
      name: 'E2E RedLotusKnight',
      camp: 'Red',
      role: 'red_lotus_knight',
      hand,
      hand_count: hand.length,
      crystal: options.crystal ?? 0,
      gem: options.gem ?? 0,
      heal: options.heal ?? 2,
      max_heal: 4,
      is_active: true,
      buffs: options.buffs ?? [],
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
      tokens: { blood_mark: 1 },
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
    myPlayerId: RED_LOTUS_KNIGHT_PLAYER_ID,
    myPlayerName: 'E2E RedLotusKnight',
    characters: defaultCharacters,
    players: [
      playerInfo({ id: RED_LOTUS_KNIGHT_PLAYER_ID, name: 'E2E RedLotusKnight', camp: 'Red', char_role: 'red_lotus_knight', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy E1', camp: 'Blue', char_role: 'enemy_char' }),
      playerInfo({ id: ALLY_PLAYER_ID, name: 'Ally A1', camp: 'Red', char_role: 'ally_char' }),
    ],
    initialState: syncState({
      turn_player_id: RED_LOTUS_KNIGHT_PLAYER_ID,
      turn_stage: options.turnStage ?? 'ActionExecution',
      available_skills: options.availableSkills ?? [],
      characters: defaultCharacters,
      players,
    }),
  };
}

// ============================================================
// Scarlet Covenant (腥红圣约) - 响应技能
// ============================================================

export function scarletCovenantScenario(): ProtocolHarnessScenario {
  return redLotusKnightScenario();
}

export function scarletCovenantPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: RED_LOTUS_KNIGHT_PLAYER_ID,
    message: '【腥红圣约】攻击伤害+1，攻击结束后+1治疗？',
    choice_type: 'crk_scarlet_covenant',
    options: [
      { id: 'confirm', label: '发动' },
      { id: 'skip', label: '跳过' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}

// ============================================================
// Bloody Prayer (血腥祷言) - 启动技能
// ============================================================

export function bloodyPrayerScenario(options: {
  heal?: number;
} = {}): ProtocolHarnessScenario {
  return redLotusKnightScenario({
    heal: options.heal ?? 2,
    availableSkills: [
      redLotusKnightAvailableSkill({
        id: RED_LOTUS_KNIGHT_BLOODY_PRAYER_ID, title: '血腥祷言',
      }),
    ],
  });
}

export function bloodyPrayerXPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: RED_LOTUS_KNIGHT_PLAYER_ID,
    message: '【血腥祷言】移除多少治疗放置血印？（X最大为2）',
    choice_type: 'crk_bloody_prayer_x',
    options: [
      { id: '1', label: '移除1治疗' },
      { id: '2', label: '移除2治疗' },
    ],
    min: 1, max: 1,
    presentation: { kind: 'numeric', numeric_base: 0 },
  } satisfies Prompt);
}

export function bloodyPrayerTargetPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: RED_LOTUS_KNIGHT_PLAYER_ID,
    message: '【血腥祷言】请选择一名对手放置血印：',
    choice_type: 'crk_bloody_prayer_target',
    options: [
      { id: ENEMY_PLAYER_ID, label: 'Enemy E1' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}

// ============================================================
// Slaughter Feast (杀戮盛宴) - 响应技能
// ============================================================

export function slaughterFeastScenario(): ProtocolHarnessScenario {
  return redLotusKnightScenario();
}

export function slaughterFeastPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: RED_LOTUS_KNIGHT_PLAYER_ID,
    message: '【杀戮盛宴】移除目标1血印，对自己造成1点法术伤害，攻击伤害+2？',
    choice_type: 'crk_slaughter_feast',
    options: [
      { id: 'confirm', label: '发动' },
      { id: 'skip', label: '跳过' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}

// ============================================================
// Hot Blood Boiling (热血沸腾) - 被动响应技能
// ============================================================

export function hotBloodBoilingScenario(): ProtocolHarnessScenario {
  return redLotusKnightScenario({
    buffs: [{ id: 'hot_blood_form', name: '热血形态', duration: 0, value: 0, source_id: RED_LOTUS_KNIGHT_HOT_BLOOD_BOILING_ID }],
  });
}

// 热血沸腾自动触发，无交互prompt

// ============================================================
// Modesty (戒骄戒躁) - 响应技能(大招)
// ============================================================

export function modestyScenario(): ProtocolHarnessScenario {
  return redLotusKnightScenario({
    crystal: 1,
    buffs: [{ id: 'hot_blood_form', name: '热血形态', duration: 0, value: 0, source_id: RED_LOTUS_KNIGHT_HOT_BLOOD_BOILING_ID }],
  });
}

export function modestyBranchPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: RED_LOTUS_KNIGHT_PLAYER_ID,
    message: '【戒骄戒躁】脱离热血形态，选择：',
    choice_type: 'crk_modesty_branch',
    options: [
      { id: 'draw', label: '摸2张牌' },
      { id: 'heal', label: '+2治疗' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}

// ============================================================
// Scarlet Cross (腥红十字) - 法术技能(大招)
// ============================================================

export function scarletCrossScenario(): ProtocolHarnessScenario {
  return redLotusKnightScenario({
    crystal: 1,
    availableSkills: [
      redLotusKnightAvailableSkill({
        id: RED_LOTUS_KNIGHT_SCARLET_CROSS_ID, title: '腥红十字',
        min_targets: 1, max_targets: 1, target_type: 2,
      }),
    ],
  });
}

export function scarletCrossDiscardPrompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_cards',
    player_id: RED_LOTUS_KNIGHT_PLAYER_ID,
    message: '【腥红十字】请选择弃1张法术牌［展示］：',
    choice_type: 'crk_scarlet_cross_discard',
    options: [
      { id: 'rlk-magic-1', label: '血术（法术）' },
      { id: 'rlk-magic-2', label: '红莲（法术）' },
      { id: 'rlk-water-magic', label: '治愈（法术）' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}

export function scarletCrossTargetPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: RED_LOTUS_KNIGHT_PLAYER_ID,
    message: '【腥红十字】请选择一名对手（需有血印）造成2点法术伤害：',
    choice_type: 'crk_scarlet_cross_target',
    options: [
      { id: ENEMY_PLAYER_ID, label: 'Enemy E1' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}