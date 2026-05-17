// ============================================================
// WindSwordSaint (风之剑圣) Protocol Harness Scenarios
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

export const WSS_PLAYER_ID = 'wss_player';
export const ENEMY_PLAYER_ID = 'enemy_1';
export const ALLY_PLAYER_ID = 'ally_1';

export const WSS_WIND_FURY_ID = 'wss_wind_fury';
export const WSS_HOLY_SWORD_ID = 'wss_holy_sword';
export const WSS_SWORD_SHADOW_ID = 'wss_sword_shadow';
export const WSS_GALE_SKILL_ID = 'wss_gale_skill'; // 独有牌 疾风技
export const WSS_WIND_BLADE_ID = 'wss_wind_blade'; // 独有牌 列风技

const windSwordSaintCharacter = characterView({
  id: 'wind_sword_saint',
  name: '风之剑圣',
  title: '技',
  faction: '技',
  skills: [
    {
      id: WSS_WIND_FURY_ID,
      title: '风怒追击',
      description: '［回合限定］（［攻击行动］结束时发动）额外+1风系［攻击行动］。',
      type: 3, // 响应
      min_targets: 0, max_targets: 0, target_type: 0,
    },
    {
      id: WSS_HOLY_SWORD_ID,
      title: '圣剑',
      description: '若你的主动攻击为你本次行动阶段的第三次［攻击行动］，则此攻击强制命中。本次［攻击行动］结束后，你摸X张牌，弃X张牌（X<4）。',
      type: 0, // 被动
      min_targets: 0, max_targets: 0, target_type: 0,
    },
    {
      id: WSS_SWORD_SHADOW_ID,
      title: '剑影',
      description: '［回合限定］［水晶］（［攻击行动］结束时发动）额外+1［攻击行动］。',
      type: 3, // 响应(大招)
      min_targets: 0, max_targets: 0, target_type: 0,
    },
  ],
});

const enemyCharacter = characterView({
  id: 'enemy_char', name: '守卫', title: '测试目标', faction: '异端', skills: [],
});

const allyCharacter = characterView({
  id: 'ally_char', name: '勇者', title: '测试队友', faction: '技', skills: [],
});

const defaultCharacters = [windSwordSaintCharacter, enemyCharacter, allyCharacter];

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function wssHand(): Card[] {
  return [
    card({ id: 'wss-wind-atk1', name: '风刃', type: 'Attack', element: 'Wind' }),
    card({ id: 'wss-wind-atk2', name: '疾风斩', type: 'Attack', element: 'Wind' }),
    card({ id: 'wss-wind-atk3', name: '风暴斩', type: 'Attack', element: 'Wind' }),
    card({ id: WSS_GALE_SKILL_ID, name: '疾风技', type: 'Attack', element: 'Wind' }),
    card({ id: WSS_WIND_BLADE_ID, name: '列风技', type: 'Attack', element: 'Wind' }),
    card({ id: 'wss-fire-atk', name: '火焰斩', type: 'Attack', element: 'Fire' }),
  ];
}

function wssAvailableSkill(skill: Partial<AvailableSkill> & { id: string; title: string }): AvailableSkill {
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

export function wssScenario(options: {
  hand?: Card[];
  availableSkills?: AvailableSkill[];
  crystal?: number;
  gem?: number;
  turnStage?: string;
  attackCount?: number; // 本回合攻击次数
} = {}): ProtocolHarnessScenario {
  const hand = options.hand ?? wssHand();
  const players = [
    playerView({
      id: WSS_PLAYER_ID,
      name: 'E2E WindSwordSaint',
      camp: 'Red',
      role: 'wind_sword_saint',
      hand,
      hand_count: hand.length,
      crystal: options.crystal ?? 0,
      gem: options.gem ?? 0,
      is_active: true,
      tokens: { attack_count: options.attackCount ?? 0 },
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
      buffs: [{ id: 'shield', name: '圣盾', duration: 1, value: 1, source_id: 'shield' }],
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
    myPlayerId: WSS_PLAYER_ID,
    myPlayerName: 'E2E WindSwordSaint',
    characters: defaultCharacters,
    players: [
      playerInfo({ id: WSS_PLAYER_ID, name: 'E2E WindSwordSaint', camp: 'Red', char_role: 'wind_sword_saint', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy E1', camp: 'Blue', char_role: 'enemy_char' }),
      playerInfo({ id: ALLY_PLAYER_ID, name: 'Ally A1', camp: 'Red', char_role: 'ally_char' }),
    ],
    initialState: syncState({
      turn_player_id: WSS_PLAYER_ID,
      turn_stage: options.turnStage ?? 'ActionExecution',
      available_skills: options.availableSkills ?? [],
      characters: defaultCharacters,
      players,
    }),
  };
}

// ============================================================
// Wind Fury (风怒追击) - 响应技能
// ============================================================

export function windFuryScenario(): ProtocolHarnessScenario {
  return wssScenario({
    attackCount: 1, // 第一次攻击结束
  });
}

export function windFuryPrompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_skill',
    player_id: WSS_PLAYER_ID,
    message: '【风怒追击］攻击行动结束，是否额外+1风系［攻击行动］？',
    choice_type: 'response_skill_choice',
    options: [
      { id: 'wind_fury', label: '发动风怒追击' },
      { id: 'skip', label: '跳过' },
    ],
    min: 1, max: 1,
    presentation: { kind: 'skill_choice', layout: 'overlay' },
  } satisfies Prompt);
}

// ============================================================
// Holy Sword (圣剑) - 被动技能
// ============================================================

export function holySwordScenario(options: {
  attackCount?: number;
} = {}): ProtocolHarnessScenario {
  return wssScenario({
    attackCount: options.attackCount ?? 3, // 第三次攻击触发圣剑
  });
}

export function holySwordThirdAttackPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: WSS_PLAYER_ID,
    message: '【圣剑】第三次主动攻击，本次攻击强制命中。攻击结束后请选择摸弃牌数量（X<4）：',
    choice_type: 'holy_sword_draw',
    options: [
      { id: '0', label: 'X=0（不摸不弃）' },
      { id: '1', label: 'X=1（摸1弃1）' },
      { id: '2', label: 'X=2（摸2弃2）' },
      { id: '3', label: 'X=3（摸3弃3）' },
    ],
    min: 1, max: 1,
    presentation: { kind: 'numeric', numeric_base: 0 },
  } satisfies Prompt);
}

export function holySwordDiscardPrompt(x: number): WsMessage {
  return requireActionMessage({
    type: 'choose_cards',
    player_id: WSS_PLAYER_ID,
    message: `【圣剑】请弃置${x}张牌：`,
    choice_type: 'wss_holy_sword_discard',
    options: [
      { id: '0', label: '1: 风刃（风系 攻击）' },
      { id: '1', label: '2: 疾风斩（风系 攻击）' },
      { id: '2', label: '3: 风暴斩（风系 攻击）' },
      { id: '3', label: '4: 疾风技（风系 攻击）' },
      { id: '4', label: '5: 列风技（风系 攻击）' },
      { id: '5', label: '6: 火焰斩（火系 攻击）' },
    ],
    min: x, max: x,
  } satisfies Prompt);
}

// ============================================================
// Sword Shadow (剑影) - 响应技能(大招)
// ============================================================

export function swordShadowScenario(options: {
  crystal?: number;
} = {}): ProtocolHarnessScenario {
  return wssScenario({
    crystal: options.crystal ?? 1,
    attackCount: 1,
  });
}

export function swordShadowPrompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_skill',
    player_id: WSS_PLAYER_ID,
    message: '【剑影］攻击行动结束，是否消耗1个水晶额外+1［攻击行动］？',
    choice_type: 'response_skill_choice',
    options: [
      { id: 'sword_shadow', label: '发动剑影（消耗水晶）' },
      { id: 'skip', label: '跳过' },
    ],
    min: 1, max: 1,
    presentation: { kind: 'skill_choice', layout: 'overlay' },
  } satisfies Prompt);
}

// ============================================================
// Wind Fury + Sword Shadow Combo (互斥选择)
// ============================================================

export function wssComboScenario(options: {
  crystal?: number;
} = {}): ProtocolHarnessScenario {
  return wssScenario({
    crystal: options.crystal ?? 1,
    attackCount: 1,
  });
}

export function wssComboPrompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_skill',
    player_id: WSS_PLAYER_ID,
    message: '【风怒追击】与【剑影】均可发动，请选择：',
    choice_type: 'response_skill_choice',
    options: [
      { id: 'wind_fury', label: '发动风怒追击（风系攻击）' },
      { id: 'sword_shadow', label: '发动剑影（消耗水晶）' },
      { id: 'skip', label: '跳过' },
    ],
    min: 1, max: 1,
    presentation: { kind: 'skill_choice', layout: 'overlay' },
  } satisfies Prompt);
}

// ============================================================
// Gale Skill (疾风技) - 独有牌
// ============================================================

export function galeSkillScenario(): ProtocolHarnessScenario {
  return wssScenario({
    hand: [
      card({ id: WSS_GALE_SKILL_ID, name: '疾风技', type: 'Attack', element: 'Wind' }),
      card({ id: 'wss-wind-atk1', name: '风刃', type: 'Attack', element: 'Wind' }),
      card({ id: 'wss-fire-atk', name: '火焰斩', type: 'Attack', element: 'Fire' }),
    ],
  });
}

// ============================================================
// Wind Blade (列风技) - 独有牌
// ============================================================

export function windBladeScenario(): ProtocolHarnessScenario {
  return wssScenario({
    hand: [
      card({ id: WSS_WIND_BLADE_ID, name: '列风技', type: 'Attack', element: 'Wind' }),
      card({ id: 'wss-wind-atk1', name: '风刃', type: 'Attack', element: 'Wind' }),
      card({ id: 'wss-fire-atk', name: '火焰斩', type: 'Attack', element: 'Fire' }),
    ],
  });
}

export function windBladeShieldPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: WSS_PLAYER_ID,
    message: '【列风技】目标拥有圣盾，本次攻击无视圣盾且对手无法应战。',
    choice_type: 'wss_wind_blade_shield',
    options: [
      { id: 'confirm', label: '确认' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}