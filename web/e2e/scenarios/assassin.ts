// ============================================================
// Assassin (暗杀者) Protocol Harness Scenarios
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

export const ASSASSIN_PLAYER_ID = 'assassin_player';
export const ENEMY_PLAYER_ID = 'enemy_1';
export const ALLY_PLAYER_ID = 'ally_1';

export const ASSASSIN_WATER_SHADOW_ID = 'assassin_water_shadow';
export const ASSASSIN_STEALTH_ID = 'assassin_stealth';

const assassinCharacter = characterView({
  id: 'assassin',
  name: '暗杀者',
  title: '技',
  faction: '技',
  skills: [
    {
      id: 'assassin_counter_bite',
      title: '反噬',
      description: '（承受攻击伤害时发动⑥）攻击你的对手摸1张牌［强制］。',
      type: 0, // 被动
      min_targets: 0, max_targets: 0, target_type: 0,
    },
    {
      id: ASSASSIN_WATER_SHADOW_ID,
      title: '水影',
      description: '（除［特殊行动］外，当你摸牌前发动）弃X张水系牌（展示）；（若你处于［潜行］效果下）你可额外弃1张法术牌（展示）。',
      type: 3, // 响应
      min_targets: 0, max_targets: 0, target_type: 0,
    },
    {
      id: ASSASSIN_STEALTH_ID,
      title: '潜行',
      description: '［宝石］你可选择摸1张牌，［横置］持续到你的下个行动阶段开始，你的手牌上限-1；你不能成为主动攻击的目标；你的主动攻击对方无法应战且伤害额外+X，X为你剩余的能量数。潜行的效果结束时角色［转正］。',
      type: 1, // 启动(大招)
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

const defaultCharacters = [assassinCharacter, enemyCharacter, allyCharacter];

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function assassinHand(): Card[] {
  return [
    card({ id: 'assassin-water-atk1', name: '水刃', type: 'Attack', element: 'Water' }),
    card({ id: 'assassin-water-atk2', name: '寒冰斩', type: 'Attack', element: 'Water' }),
    card({ id: 'assassin-water-magic', name: '冰冻', type: 'Magic', element: 'Water' }),
    card({ id: 'assassin-fire-magic', name: '火球', type: 'Magic', element: 'Fire' }),
    card({ id: 'assassin-thunder-magic', name: '雷击', type: 'Magic', element: 'Thunder' }),
    card({ id: 'assassin-wind-atk', name: '风刃', type: 'Attack', element: 'Wind' }),
  ];
}

function assassinAvailableSkill(skill: Partial<AvailableSkill> & { id: string; title: string }): AvailableSkill {
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

export function assassinScenario(options: {
  hand?: Card[];
  availableSkills?: AvailableSkill[];
  crystal?: number;
  gem?: number;
  turnStage?: string;
  tokens?: Record<string, number>;
  inStealth?: boolean;
} = {}): ProtocolHarnessScenario {
  const hand = options.hand ?? assassinHand();
  const buffs = options.inStealth ? [{ id: 'stealth', name: '潜行', duration: 1, value: 0, source_id: ASSASSIN_STEALTH_ID }] : [];
  const players = [
    playerView({
      id: ASSASSIN_PLAYER_ID,
      name: 'E2E Assassin',
      camp: 'Red',
      role: 'assassin',
      hand,
      hand_count: hand.length,
      crystal: options.crystal ?? 0,
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
    myPlayerId: ASSASSIN_PLAYER_ID,
    myPlayerName: 'E2E Assassin',
    characters: defaultCharacters,
    players: [
      playerInfo({ id: ASSASSIN_PLAYER_ID, name: 'E2E Assassin', camp: 'Red', char_role: 'assassin', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy E1', camp: 'Blue', char_role: 'enemy_char' }),
      playerInfo({ id: ALLY_PLAYER_ID, name: 'Ally A1', camp: 'Red', char_role: 'ally_char' }),
    ],
    initialState: syncState({
      turn_player_id: ASSASSIN_PLAYER_ID,
      turn_stage: options.turnStage ?? 'StartupPhase',
      available_skills: options.availableSkills ?? [],
      characters: defaultCharacters,
      players,
    }),
  };
}

// ============================================================
// Stealth (潜行) - 启动技能(大招)
// ============================================================

export function stealthScenario(options: {
  gem?: number;
} = {}): ProtocolHarnessScenario {
  return assassinScenario({
    gem: options.gem ?? 1,
    turnStage: 'StartupPhase',
    availableSkills: [
      assassinAvailableSkill({
        id: ASSASSIN_STEALTH_ID, title: '潜行',
      }),
    ],
  });
}

export function stealthConfirmPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: ASSASSIN_PLAYER_ID,
    message: '【潜行】是否消耗1个红宝石发动该技能？',
    choice_type: 'assassin_stealth_confirm',
    options: [
      { id: 'confirm', label: '发动' },
      { id: 'skip', label: '跳过' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}

export function stealthDrawPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: ASSASSIN_PLAYER_ID,
    message: '【潜行】发动成功，是否摸1张牌？',
    choice_type: 'assassin_stealth_draw',
    options: [
      { id: 'draw', label: '摸1张牌' },
      { id: 'skip', label: '不摸牌' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}

// ============================================================
// Water Shadow (水影) - 响应技能
// ============================================================

export function waterShadowScenario(options: {
  inStealth?: boolean;
} = {}): ProtocolHarnessScenario {
  return assassinScenario({
    inStealth: options.inStealth ?? false,
    turnStage: 'ActionExecution',
  });
}

export function waterShadowBeforeDrawPrompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_cards',
    player_id: ASSASSIN_PLAYER_ID,
    message: '【水影】摸牌前，请选择弃置X张水系牌（展示）：',
    choice_type: 'assassin_water_shadow_discard',
    options: [
      { id: '0', label: '1: 水刃（水系 攻击）' },
      { id: '1', label: '2: 寒冰斩（水系 攻击）' },
      { id: '2', label: '3: 冰冻（水系 法术）' },
    ],
    min: 0, max: 3,
  } satisfies Prompt);
}

export function waterShadowStealthExtraPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: ASSASSIN_PLAYER_ID,
    message: '【水影】你处于［潜行］效果下，是否额外弃1张法术牌？',
    choice_type: 'assassin_water_shadow_extra',
    options: [
      { id: 'yes', label: '弃1张法术牌' },
      { id: 'no', label: '不弃牌' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}

export function waterShadowExtraCardPrompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_cards',
    player_id: ASSASSIN_PLAYER_ID,
    message: '【水影】请选择1张法术牌弃置（展示）：',
    choice_type: 'assassin_water_shadow_extra_card',
    options: [
      { id: '2', label: '3: 冰冻（水系 法术）' },
      { id: '3', label: '4: 火球（火系 法术）' },
      { id: '4', label: '5: 雷击（雷系 法术）' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}