// ============================================================
// Beast Soul Warrior (兽灵武士) Protocol Harness Scenarios
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
export const BSW_PLAYER_ID = 'bsw_player';
export const ALLY_PLAYER_ID = 'ally_1';
export const ENEMY_PLAYER_ID = 'enemy_1';

// ---- Skill IDs ----
export const BSW_INDOMITABLE_WILL_SKILL_ID = 'bsw_indomitable_will';
export const BSW_WARRIOR_ZANSHIN_SKILL_ID = 'bsw_warrior_zanshin';
export const BSW_ONE_STRIKE_SKILL_ID = 'bsw_one_strike_no_thought';
export const BSW_BEAST_SOUL_ALERT_SKILL_ID = 'bsw_beast_soul_alert';
export const BSW_BEAST_RETURN_SKILL_ID = 'bsw_beast_return';
export const BSW_REVERSAL_IAIJUTSU_SKILL_ID = 'bsw_reversal_iaijutsu';
export const BSW_IAIJUTSU_STYLE_SKILL_ID = 'bsw_iaijutsu_style';

// ---- Beast Soul Warrior character definition ----
const beastSoulCharacter = characterView({
  id: 'beast_soul_warrior',
  name: '兽灵武士',
  title: '野兽之魂',
  faction: '星杯',
  skills: [
    {
      id: BSW_INDOMITABLE_WILL_SKILL_ID,
      title: '不屈意志',
      description: '攻击行动结束时，若有水晶则可发动',
      type: 0,
      min_targets: 0,
      max_targets: 0,
      target_type: 0,
      cost_gem: 0,
      cost_crystal: 0,
      cost_discards: 0,
    },
    {
      id: BSW_WARRIOR_ZANSHIN_SKILL_ID,
      title: '武者残心',
      description: '首次攻击行动结束时触发',
      type: 0,
      min_targets: 0,
      max_targets: 0,
      target_type: 0,
      cost_gem: 0,
      cost_crystal: 0,
      cost_discards: 0,
    },
    {
      id: BSW_ONE_STRIKE_SKILL_ID,
      title: '一击无念',
      description: '残心≥4时可发动',
      type: 0,
      min_targets: 0,
      max_targets: 0,
      target_type: 0,
      cost_gem: 0,
      cost_crystal: 0,
      cost_discards: 0,
    },
    {
      id: BSW_BEAST_SOUL_ALERT_SKILL_ID,
      title: '兽魂警戒',
      description: '触发后确认，选目标，弃牌',
      type: 0,
      min_targets: 0,
      max_targets: 0,
      target_type: 0,
      cost_gem: 0,
      cost_crystal: 0,
      cost_discards: 0,
    },
    {
      id: BSW_BEAST_RETURN_SKILL_ID,
      title: '兽返',
      description: '响应技能，确认后移除兽魂',
      type: 0,
      min_targets: 0,
      max_targets: 0,
      target_type: 0,
      cost_gem: 0,
      cost_crystal: 0,
      cost_discards: 0,
    },
    {
      id: BSW_REVERSAL_IAIJUTSU_SKILL_ID,
      title: '逆反居合斩',
      description: '响应技能，确认后移除兽魂，目标弃牌',
      type: 0,
      min_targets: 0,
      max_targets: 0,
      target_type: 0,
      cost_gem: 0,
      cost_crystal: 0,
      cost_discards: 0,
    },
    {
      id: BSW_IAIJUTSU_STYLE_SKILL_ID,
      title: '徙魂流居合式',
      description: '启动技能，消耗宝石，选择摸牌或弃牌',
      type: 2,
      min_targets: 0,
      max_targets: 0,
      target_type: 0,
      cost_gem: 1,
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
function beastSoulHand() {
  return [
    card({ id: 'card_1', name: '火焰斩', type: 'Attack', element: 'Fire' }),
    card({ id: 'card_2', name: '水涟斩', type: 'Attack', element: 'Water' }),
    card({ id: 'card_3', name: '风刃', type: 'Attack', element: 'Wind' }),
    card({ id: 'card_4', name: '寒冰箭', type: 'Magic', element: 'Water' }),
  ];
}

function beastSoulPlayerView(options: {
  is_active?: boolean;
  heal?: number;
  max_heal?: number;
  gems?: number;
  crystals?: number;
  zanshin?: number;
  beast_souls?: number;
} = {}) {
  return playerView({
    id: BSW_PLAYER_ID,
    name: 'E2E Beast Soul',
    camp: 'Red',
    role: 'beast_soul_warrior',
    hand: beastSoulHand(),
    hand_count: beastSoulHand().length,
    heal: options.heal ?? 2,
    max_heal: options.max_heal ?? 4,
    is_active: options.is_active ?? true,
    gem: options.gems ?? 0,
    crystal: options.crystals ?? 0,
    tokens: {
      bsw_zanshin: options.zanshin ?? 0,
      bsw_beast_souls: options.beast_souls ?? 0,
    },
  });
}

// ============================================================
// 不屈意志 (bsw_indomitable_will) - Attack action end, requires crystal
// ============================================================

export function indomitableWillScenario(options: { crystals?: number } = {}): ProtocolHarnessScenario {
  const crystals = options.crystals ?? 1;
  const characters = [beastSoulCharacter, allyCharacter, enemyCharacter];

  const beastSoul = beastSoulPlayerView({ crystals, is_active: true });

  const players = [
    beastSoul,
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
    myPlayerId: BSW_PLAYER_ID,
    myPlayerName: 'E2E Beast Soul',
    characters,
    players: [
      playerInfo({ id: BSW_PLAYER_ID, name: 'E2E Beast Soul', camp: 'Red', char_role: 'beast_soul_warrior', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy Bot', camp: 'Blue', char_role: 'villain' }),
    ],
    initialState: syncState({
      turn_player_id: BSW_PLAYER_ID,
      turn_stage: 'ActionExecution',
      available_skills: [],
      characters,
      players,
    }),
  };
}

export function indomitableWillConfirmPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: BSW_PLAYER_ID,
    message: '【不屈意志】攻击行动结束，是否发动？（消耗1水晶）',
    choice_type: 'bsw_indomitable_will_confirm',
    skill_id: BSW_INDOMITABLE_WILL_SKILL_ID,
    options: [
      { id: '0', label: '发动' },
      { id: '1', label: '不发动' },
    ],
    min: 1,
    max: 1,
  } satisfies Prompt);
}

// ============================================================
// 武者残心 (bsw_warrior_zanshin) - First attack action end
// ============================================================

export function warriorZanshinScenario(options: { zanshin?: number } = {}): ProtocolHarnessScenario {
  const zanshin = options.zanshin ?? 0;
  const characters = [beastSoulCharacter, allyCharacter, enemyCharacter];

  const beastSoul = beastSoulPlayerView({ zanshin, is_active: true });

  const players = [
    beastSoul,
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
    myPlayerId: BSW_PLAYER_ID,
    myPlayerName: 'E2E Beast Soul',
    characters,
    players: [
      playerInfo({ id: BSW_PLAYER_ID, name: 'E2E Beast Soul', camp: 'Red', char_role: 'beast_soul_warrior', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy Bot', camp: 'Blue', char_role: 'villain' }),
    ],
    initialState: syncState({
      turn_player_id: BSW_PLAYER_ID,
      turn_stage: 'ActionExecution',
      available_skills: [],
      characters,
      players,
    }),
  };
}

export function warriorZanshinConfirmPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: BSW_PLAYER_ID,
    message: '【武者残心】首次攻击行动结束，是否发动？',
    choice_type: 'bsw_warrior_zanshin_confirm',
    skill_id: BSW_WARRIOR_ZANSHIN_SKILL_ID,
    options: [
      { id: '0', label: '发动' },
      { id: '1', label: '不发动' },
    ],
    min: 1,
    max: 1,
  } satisfies Prompt);
}

// ============================================================
// 一击无念 (bsw_one_strike_no_thought) - Requires 4+ zanshin tokens
// ============================================================

export function oneStrikeScenario(options: { zanshin?: number } = {}): ProtocolHarnessScenario {
  const zanshin = options.zanshin ?? 4;
  const characters = [beastSoulCharacter, allyCharacter, enemyCharacter];

  const beastSoul = beastSoulPlayerView({ zanshin, is_active: true });

  const players = [
    beastSoul,
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
    myPlayerId: BSW_PLAYER_ID,
    myPlayerName: 'E2E Beast Soul',
    characters,
    players: [
      playerInfo({ id: BSW_PLAYER_ID, name: 'E2E Beast Soul', camp: 'Red', char_role: 'beast_soul_warrior', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy Bot', camp: 'Blue', char_role: 'villain' }),
    ],
    initialState: syncState({
      turn_player_id: BSW_PLAYER_ID,
      turn_stage: 'ActionExecution',
      available_skills: [],
      characters,
      players,
    }),
  };
}

export function oneStrikeConfirmPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: BSW_PLAYER_ID,
    message: '【一击无念】残心≥4，是否发动？（清空残心，+1兽魂标记）',
    choice_type: 'bsw_one_strike_confirm',
    skill_id: BSW_ONE_STRIKE_SKILL_ID,
    options: [
      { id: '0', label: '发动' },
      { id: '1', label: '不发动' },
    ],
    min: 1,
    max: 1,
  } satisfies Prompt);
}

// ============================================================
// Mutual exclusion scenario - All three skills trigger simultaneously
// ============================================================

export function mutualExclusionScenario(options: { crystals?: number; zanshin?: number } = {}): ProtocolHarnessScenario {
  const crystals = options.crystals ?? 1;
  const zanshin = options.zanshin ?? 4;
  const characters = [beastSoulCharacter, allyCharacter, enemyCharacter];

  const beastSoul = beastSoulPlayerView({ crystals, zanshin, is_active: true });

  const players = [
    beastSoul,
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
    myPlayerId: BSW_PLAYER_ID,
    myPlayerName: 'E2E Beast Soul',
    characters,
    players: [
      playerInfo({ id: BSW_PLAYER_ID, name: 'E2E Beast Soul', camp: 'Red', char_role: 'beast_soul_warrior', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy Bot', camp: 'Blue', char_role: 'villain' }),
    ],
    initialState: syncState({
      turn_player_id: BSW_PLAYER_ID,
      turn_stage: 'ActionExecution',
      available_skills: [],
      characters,
      players,
    }),
  };
}

export function mutualExclusionPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: BSW_PLAYER_ID,
    message: '攻击行动结束，多个技能可发动，请选择：',
    choice_type: 'bsw_attack_end_choice',
    options: [
      { id: 'indomitable_will', label: '不屈意志（消耗1水晶）' },
      { id: 'warrior_zanshin', label: '武者残心' },
      { id: 'one_strike', label: '一击无念（清空残心）' },
    ],
    presentation: { kind: 'branch_select', layout: 'overlay' },
    cancelable: true,
    min: 1,
    max: 1,
  } satisfies Prompt);
}

// ============================================================
// 兽魂警戒 (bsw_beast_soul_alert) - Trigger → confirm → target → discard
// ============================================================

export function beastSoulAlertScenario(options: { beast_souls?: number } = {}): ProtocolHarnessScenario {
  const beast_souls = options.beast_souls ?? 2;
  const characters = [beastSoulCharacter, allyCharacter, enemyCharacter];

  const beastSoul = beastSoulPlayerView({ beast_souls, is_active: true });

  const players = [
    beastSoul,
    playerView({
      id: ENEMY_PLAYER_ID,
      name: 'Enemy Bot',
      camp: 'Blue',
      role: 'villain',
      hand: [],
      hand_count: 4,
      heal: 1,
      max_heal: 2,
      is_active: false,
    }),
  ];

  return {
    roomCode: 'MOCK',
    myPlayerId: BSW_PLAYER_ID,
    myPlayerName: 'E2E Beast Soul',
    characters,
    players: [
      playerInfo({ id: BSW_PLAYER_ID, name: 'E2E Beast Soul', camp: 'Red', char_role: 'beast_soul_warrior', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy Bot', camp: 'Blue', char_role: 'villain' }),
    ],
    initialState: syncState({
      turn_player_id: BSW_PLAYER_ID,
      turn_stage: 'ActionExecution',
      available_skills: [],
      characters,
      players,
    }),
  };
}

export function beastSoulAlertConfirmPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: BSW_PLAYER_ID,
    message: '【兽魂警戒】触发条件满足，是否发动？',
    choice_type: 'bsw_beast_soul_alert_confirm',
    skill_id: BSW_BEAST_SOUL_ALERT_SKILL_ID,
    options: [
      { id: '0', label: '发动' },
      { id: '1', label: '不发动' },
    ],
    min: 1,
    max: 1,
  } satisfies Prompt);
}

export function beastSoulAlertTargetPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: BSW_PLAYER_ID,
    message: '【兽魂警戒】请选择目标：',
    choice_type: 'bsw_beast_soul_alert_target',
    skill_id: BSW_BEAST_SOUL_ALERT_SKILL_ID,
    options: [
      { id: ENEMY_PLAYER_ID, label: '恶徒' },
    ],
    min: 1,
    max: 1,
  } satisfies Prompt);
}

export function beastSoulAlertDiscardPrompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_cards',
    player_id: BSW_PLAYER_ID,
    message: '【兽魂警戒】请选择弃置1张牌：',
    choice_type: 'bsw_beast_soul_alert_discard',
    skill_id: BSW_BEAST_SOUL_ALERT_SKILL_ID,
    options: [
      { id: '0', label: '1: 火焰斩 (火 Attack)' },
      { id: '1', label: '2: 水涟斩 (水 Attack)' },
    ],
    min: 1,
    max: 1,
  } satisfies Prompt);
}

// ============================================================
// 兽返 (bsw_beast_return) - Response → confirm → remove X beast souls
// ============================================================

export function beastReturnScenario(options: { beast_souls?: number } = {}): ProtocolHarnessScenario {
  const beast_souls = options.beast_souls ?? 3;
  const characters = [beastSoulCharacter, allyCharacter, enemyCharacter];

  const beastSoul = beastSoulPlayerView({ beast_souls, is_active: false });

  const players = [
    beastSoul,
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
    myPlayerId: BSW_PLAYER_ID,
    myPlayerName: 'E2E Beast Soul',
    characters,
    players: [
      playerInfo({ id: BSW_PLAYER_ID, name: 'E2E Beast Soul', camp: 'Red', char_role: 'beast_soul_warrior', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy Bot', camp: 'Blue', char_role: 'villain' }),
    ],
    initialState: syncState({
      turn_player_id: ENEMY_PLAYER_ID,
      turn_stage: 'ActionExecution',
      available_skills: [],
      characters,
      players,
    }),
  };
}

export function beastReturnConfirmPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: BSW_PLAYER_ID,
    message: '【兽返】响应时机触发，是否发动？',
    choice_type: 'bsw_beast_return_confirm',
    skill_id: BSW_BEAST_RETURN_SKILL_ID,
    options: [
      { id: '0', label: '发动' },
      { id: '1', label: '不发动' },
    ],
    min: 1,
    max: 1,
  } satisfies Prompt);
}

export function beastReturnRemovePrompt(xMax: number): WsMessage {
  const options: { id: string; label: string }[] = [];
  for (let i = 1; i <= xMax; i++) {
    options.push({ id: `${i}`, label: `移除${i}个兽魂` });
  }
  return requireActionMessage({
    type: 'confirm',
    player_id: BSW_PLAYER_ID,
    message: `【兽返】请选择移除X个兽魂（1≤X≤${xMax}）：`,
    choice_type: 'bsw_beast_return_remove',
    skill_id: BSW_BEAST_RETURN_SKILL_ID,
    options,
    presentation: { kind: 'branch_select', layout: 'overlay' },
    min: 1,
    max: 1,
  } satisfies Prompt);
}

// ============================================================
// 逆反居合斩 (bsw_reversal_iaijutsu) - Response → confirm → remove X beast souls → target discard
// ============================================================

export function reversalIaijutsuScenario(options: { beast_souls?: number } = {}): ProtocolHarnessScenario {
  const beast_souls = options.beast_souls ?? 3;
  const characters = [beastSoulCharacter, allyCharacter, enemyCharacter];

  const beastSoul = beastSoulPlayerView({ beast_souls, is_active: false });

  const players = [
    beastSoul,
    playerView({
      id: ENEMY_PLAYER_ID,
      name: 'Enemy Bot',
      camp: 'Blue',
      role: 'villain',
      hand: [],
      hand_count: 6,
      heal: 1,
      max_heal: 2,
      is_active: true,
    }),
  ];

  return {
    roomCode: 'MOCK',
    myPlayerId: BSW_PLAYER_ID,
    myPlayerName: 'E2E Beast Soul',
    characters,
    players: [
      playerInfo({ id: BSW_PLAYER_ID, name: 'E2E Beast Soul', camp: 'Red', char_role: 'beast_soul_warrior', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy Bot', camp: 'Blue', char_role: 'villain' }),
    ],
    initialState: syncState({
      turn_player_id: ENEMY_PLAYER_ID,
      turn_stage: 'ActionExecution',
      available_skills: [],
      characters,
      players,
    }),
  };
}

export function reversalIaijutsuConfirmPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: BSW_PLAYER_ID,
    message: '【逆反居合斩】响应时机触发，是否发动？',
    choice_type: 'bsw_reversal_iaijutsu_confirm',
    skill_id: BSW_REVERSAL_IAIJUTSU_SKILL_ID,
    options: [
      { id: '0', label: '发动' },
      { id: '1', label: '不发动' },
    ],
    min: 1,
    max: 1,
  } satisfies Prompt);
}

export function reversalIaijutsuRemovePrompt(xMax: number): WsMessage {
  const options: { id: string; label: string }[] = [];
  for (let i = 1; i <= xMax; i++) {
    options.push({ id: `${i}`, label: `移除${i}个兽魂` });
  }
  return requireActionMessage({
    type: 'confirm',
    player_id: BSW_PLAYER_ID,
    message: `【逆反居合斩】请选择移除X个兽魂（1≤X≤${xMax}）：`,
    choice_type: 'bsw_reversal_iaijutsu_remove',
    skill_id: BSW_REVERSAL_IAIJUTSU_SKILL_ID,
    options,
    presentation: { kind: 'branch_select', layout: 'overlay' },
    min: 1,
    max: 1,
  } satisfies Prompt);
}

export function reversalIaijutsuTargetPrompt(xValue: number): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: BSW_PLAYER_ID,
    message: `【逆反居合斩】请选择目标令其弃置${xValue + 2}张牌：`,
    choice_type: 'bsw_reversal_iaijutsu_target',
    skill_id: BSW_REVERSAL_IAIJUTSU_SKILL_ID,
    options: [
      { id: ENEMY_PLAYER_ID, label: '恶徒' },
    ],
    min: 1,
    max: 1,
  } satisfies Prompt);
}

// ============================================================
// 徙魂流居合式 (bsw_iaijutsu_style) - Start skill with gem cost
// ============================================================

export function iaijutsuStyleScenario(options: { gems?: number } = {}): ProtocolHarnessScenario {
  const gems = options.gems ?? 1;
  const characters = [beastSoulCharacter, allyCharacter, enemyCharacter];

  const beastSoul = beastSoulPlayerView({ gems, is_active: true });

  const players = [
    beastSoul,
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
    myPlayerId: BSW_PLAYER_ID,
    myPlayerName: 'E2E Beast Soul',
    characters,
    players: [
      playerInfo({ id: BSW_PLAYER_ID, name: 'E2E Beast Soul', camp: 'Red', char_role: 'beast_soul_warrior', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy Bot', camp: 'Blue', char_role: 'villain' }),
    ],
    initialState: syncState({
      turn_player_id: BSW_PLAYER_ID,
      turn_stage: 'ActionExecution',
      available_skills: [availableSkill({ id: BSW_IAIJUTSU_STYLE_SKILL_ID, title: '徙魂流居合式', cost_gem: 1 })],
      characters,
      players,
    }),
  };
}

export function iaijutsuStyleConfirmPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: BSW_PLAYER_ID,
    message: '【徙魂流居合式】是否发动？（消耗1宝石）',
    choice_type: 'bsw_iaijutsu_style_confirm',
    skill_id: BSW_IAIJUTSU_STYLE_SKILL_ID,
    options: [
      { id: '0', label: '发动' },
      { id: '1', label: '不发动' },
    ],
    min: 1,
    max: 1,
  } satisfies Prompt);
}

export function iaijutsuStyleChoicePrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: BSW_PLAYER_ID,
    message: '【徙魂流居合式】请选择：',
    choice_type: 'bsw_iaijutsu_style_choice',
    skill_id: BSW_IAIJUTSU_STYLE_SKILL_ID,
    options: [
      { id: 'draw', label: '摸2张牌' },
      { id: 'discard', label: '弃2张牌' },
    ],
    presentation: { kind: 'branch_select', layout: 'overlay' },
    min: 1,
    max: 1,
  } satisfies Prompt);
}

export function iaijutsuStyleDiscardPrompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_cards',
    player_id: BSW_PLAYER_ID,
    message: '【徙魂流居合式】请选择弃置2张牌：',
    choice_type: 'bsw_iaijutsu_style_discard',
    skill_id: BSW_IAIJUTSU_STYLE_SKILL_ID,
    options: [
      { id: '0', label: '1: 火焰斩 (火 Attack)' },
      { id: '1', label: '2: 水涟斩 (水 Attack)' },
      { id: '2', label: '3: 风刃 (风 Attack)' },
    ],
    min: 2,
    max: 2,
  } satisfies Prompt);
}