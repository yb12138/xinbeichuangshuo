// ============================================================
// Spirit Caster (灵符师) Protocol Harness Scenarios
// ============================================================

import type { Element, FieldCard, Prompt } from '../../src/types/game';
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
export const SC_PLAYER_ID = 'sc_player';
export const ALLY_PLAYER_ID = 'ally_1';
export const ENEMY_PLAYER_ID = 'enemy_1';
export const ENEMY_2_PLAYER_ID = 'enemy_2';
export const ENEMY_3_PLAYER_ID = 'enemy_3';

// ---- Skill IDs ----
export const SC_TALISMAN_THUNDER_SKILL_ID = 'sc_talisman_thunder';
export const SC_TALISMAN_WIND_SKILL_ID = 'sc_talisman_wind';
export const SC_INCANTATION_SKILL_ID = 'sc_incantation';
export const SC_HUNDRED_NIGHT_SKILL_ID = 'sc_hundred_night';
export const SC_MANA_COLLAPSE_SKILL_ID = 'sc_mana_collapse';

// ---- Spirit Caster character definition ----
const spiritCasterCharacter = characterView({
  id: 'spirit_caster',
  name: '灵符师',
  title: '符咒之道',
  faction: '星杯',
  skills: [
    {
      id: SC_TALISMAN_THUNDER_SKILL_ID,
      title: '灵符-雷鸣',
      description: '法术技能，弃置雷系牌，选择2名目标',
      type: 2,
      min_targets: 0,
      max_targets: 0,
      target_type: 0,
      cost_gem: 0,
      cost_crystal: 0,
      cost_discards: 0,
    },
    {
      id: SC_TALISMAN_WIND_SKILL_ID,
      title: '灵符-风行',
      description: '法术技能，弃置风系牌，指定2名角色',
      type: 2,
      min_targets: 0,
      max_targets: 0,
      target_type: 0,
      cost_gem: 0,
      cost_crystal: 0,
      cost_discards: 0,
    },
    {
      id: SC_INCANTATION_SKILL_ID,
      title: '念咒',
      description: '响应技能，灵符发动后可盖伏妖力',
      type: 1,
      min_targets: 0,
      max_targets: 0,
      target_type: 0,
      cost_gem: 0,
      cost_crystal: 0,
      cost_discards: 0,
    },
    {
      id: SC_HUNDRED_NIGHT_SKILL_ID,
      title: '百鬼夜行',
      description: '响应技能，攻击命中后可移除妖力',
      type: 1,
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

const enemy3Character = characterView({
  id: 'enemy_3_char',
  name: '暗影',
  title: '深渊行者',
  faction: '异端',
  skills: [],
});

// ---- Helper functions ----
function spiritCasterHand() {
  return [
    card({ id: 'card_1', name: '雷光斩', type: 'Attack', element: 'Thunder' }),
    card({ id: 'card_2', name: '风神斩', type: 'Attack', element: 'Wind' }),
    card({ id: 'card_3', name: '火焰斩', type: 'Attack', element: 'Fire' }),
    card({ id: 'card_4', name: '水涟斩', type: 'Attack', element: 'Water' }),
    card({ id: 'card_5', name: '地裂斩', type: 'Attack', element: 'Earth' }),
  ];
}

function spiritCasterPowerCover(fieldIndex = 0, name = '火焰斩', element: Element = 'Fire'): FieldCard {
  return {
    card: card({
      id: `sc-youli-${fieldIndex}`,
      name,
      type: 'Attack',
      element,
    }),
    owner_id: SC_PLAYER_ID,
    source_id: SC_PLAYER_ID,
    mode: 'Cover',
    effect: 'SpiritCasterPower',
    field_hook: 'Manual',
    locked: false,
    duration: 0,
  };
}

function spiritCasterPlayerView(options: {
  is_active?: boolean;
  heal?: number;
  max_heal?: number;
  crystal?: number;
} = {}) {
  return playerView({
    id: SC_PLAYER_ID,
    name: 'E2E Spirit Caster',
    camp: 'Red',
    role: 'spirit_caster',
    hand: spiritCasterHand(),
    hand_count: spiritCasterHand().length,
    heal: options.heal ?? 2,
    max_heal: options.max_heal ?? 4,
    is_active: options.is_active ?? true,
    crystal: options.crystal ?? 0,
  });
}

// ============================================================
// 灵符-雷鸣 (sc_talisman_thunder) - Magic skill with optional mana collapse, thunder discard, 2 targets
// ============================================================

export function talismanThunderScenario(options: { hasCrystal?: boolean } = {}): ProtocolHarnessScenario {
  const characters = [spiritCasterCharacter, allyCharacter, enemyCharacter, enemy2Character];

  const spiritCaster = spiritCasterPlayerView({
    is_active: true,
    crystal: options.hasCrystal ? 1 : 0,
  });

  const players = [
    spiritCaster,
    playerView({
      id: ENEMY_PLAYER_ID,
      name: 'Enemy E1',
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
      name: 'Enemy E2',
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
    myPlayerId: SC_PLAYER_ID,
    myPlayerName: 'E2E Spirit Caster',
    characters,
    players: [
      playerInfo({ id: SC_PLAYER_ID, name: 'E2E Spirit Caster', camp: 'Red', char_role: 'spirit_caster', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy E1', camp: 'Blue', char_role: 'villain' }),
      playerInfo({ id: ENEMY_2_PLAYER_ID, name: 'Enemy E2', camp: 'Blue', char_role: 'villain' }),
    ],
    initialState: syncState({
      turn_player_id: SC_PLAYER_ID,
      turn_stage: 'ActionExecution',
      available_skills: [availableSkill({ id: SC_TALISMAN_THUNDER_SKILL_ID, title: '灵符-雷鸣' })],
      characters,
      players,
    }),
  };
}

// 灵力崩解附加选择弹框（有水晶时弹出）
export function manaCollapseConfirmPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: SC_PLAYER_ID,
    message: '【灵力崩解】是否消耗1点水晶（红宝石可替代），使本次每段伤害额外+1？',
    choice_type: 'sc_spiritual_collapse_confirm',
    skill_id: SC_TALISMAN_THUNDER_SKILL_ID,
    options: [
      { id: '0', label: '是' },
      { id: '1', label: '否' },
    ],
    min: 1,
    max: 1,
  } satisfies Prompt);
}

// 雷系牌弃置选择
export function talismanThunderDiscardPrompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_card',
    player_id: SC_PLAYER_ID,
    message: '请选择1张雷系牌弃置（并展示）：',
    choice_type: 'sc_talisman_thunder_discard',
    skill_id: SC_TALISMAN_THUNDER_SKILL_ID,
    options: [
      { id: '0', label: '1: 雷光斩 (雷 Attack)' },
    ],
    min: 1,
    max: 1,
  } satisfies Prompt);
}

// 目标选择（2名角色）
export function talismanThunderTargetPrompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_target',
    player_id: SC_PLAYER_ID,
    message: '请选择2名角色作为目标：',
    choice_type: 'sc_talisman_thunder_target',
    skill_id: SC_TALISMAN_THUNDER_SKILL_ID,
    options: [
      { id: ENEMY_PLAYER_ID, label: 'Enemy E1' },
      { id: ENEMY_2_PLAYER_ID, label: 'Enemy E2' },
    ],
    min: 2,
    max: 2,
  } satisfies Prompt);
}

// ============================================================
// 灵符-风行 (sc_talisman_wind) - Magic skill with wind discard, 2 targets
// ============================================================

export function talismanWindScenario(): ProtocolHarnessScenario {
  const characters = [spiritCasterCharacter, allyCharacter, enemyCharacter, enemy2Character];

  const spiritCaster = spiritCasterPlayerView({ is_active: true });

  const players = [
    spiritCaster,
    playerView({
      id: ENEMY_PLAYER_ID,
      name: 'Enemy E1',
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
      name: 'Enemy E2',
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
    myPlayerId: SC_PLAYER_ID,
    myPlayerName: 'E2E Spirit Caster',
    characters,
    players: [
      playerInfo({ id: SC_PLAYER_ID, name: 'E2E Spirit Caster', camp: 'Red', char_role: 'spirit_caster', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy E1', camp: 'Blue', char_role: 'villain' }),
      playerInfo({ id: ENEMY_2_PLAYER_ID, name: 'Enemy E2', camp: 'Blue', char_role: 'villain' }),
    ],
    initialState: syncState({
      turn_player_id: SC_PLAYER_ID,
      turn_stage: 'ActionExecution',
      available_skills: [availableSkill({ id: SC_TALISMAN_WIND_SKILL_ID, title: '灵符-风行' })],
      characters,
      players,
    }),
  };
}

// 风系牌弃置选择
export function talismanWindDiscardPrompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_card',
    player_id: SC_PLAYER_ID,
    message: '请选择1张风系牌弃置（并展示）：',
    choice_type: 'sc_talisman_wind_discard',
    skill_id: SC_TALISMAN_WIND_SKILL_ID,
    options: [
      { id: '1', label: '2: 风神斩 (风 Attack)' },
    ],
    min: 1,
    max: 1,
  } satisfies Prompt);
}

// 目标选择（2名角色）
export function talismanWindTargetPrompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_target',
    player_id: SC_PLAYER_ID,
    message: '请指定2名角色：',
    choice_type: 'sc_talisman_wind_target',
    skill_id: SC_TALISMAN_WIND_SKILL_ID,
    options: [
      { id: ENEMY_PLAYER_ID, label: 'Enemy E1' },
      { id: ENEMY_2_PLAYER_ID, label: 'Enemy E2' },
    ],
    min: 2,
    max: 2,
  } satisfies Prompt);
}

// ============================================================
// 念咒 (sc_incantation) - Response skill after talisman skill, cover a card as Youli
// ============================================================

export function incantationScenario(): ProtocolHarnessScenario {
  const characters = [spiritCasterCharacter, allyCharacter, enemyCharacter];

  const spiritCaster = spiritCasterPlayerView({ is_active: true });

  const players = [
    spiritCaster,
    playerView({
      id: ENEMY_PLAYER_ID,
      name: 'Enemy E1',
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
    myPlayerId: SC_PLAYER_ID,
    myPlayerName: 'E2E Spirit Caster',
    characters,
    players: [
      playerInfo({ id: SC_PLAYER_ID, name: 'E2E Spirit Caster', camp: 'Red', char_role: 'spirit_caster', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy E1', camp: 'Blue', char_role: 'villain' }),
    ],
    initialState: syncState({
      turn_player_id: SC_PLAYER_ID,
      turn_stage: 'ActionExecution',
      characters,
      players,
    }),
  };
}

// 念咒响应确认弹框
export function incantationConfirmPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: SC_PLAYER_ID,
    message: '【念咒】是否将1张手牌面朝下放置为妖力？',
    choice_type: 'sc_incant_confirm',
    skill_id: SC_INCANTATION_SKILL_ID,
    options: [
      { id: '0', label: '是' },
      { id: '1', label: '否' },
    ],
    min: 1,
    max: 1,
  } satisfies Prompt);
}

// 妖力盖牌选择（选择1张牌盖伏）- matches backend sc_incant_card
export function incantationCoverPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: SC_PLAYER_ID,
    message: '【念咒】请选择要作为妖力盖放的手牌：',
    choice_type: 'sc_incant_card',
    skill_id: SC_INCANTATION_SKILL_ID,
    options: [
      { id: '0', label: '1: 雷光斩 (雷 Attack)' },
      { id: '1', label: '2: 风神斩 (风 Attack)' },
      { id: '2', label: '3: 火焰斩 (火 Attack)' },
      { id: '3', label: '4: 水涟斩 (水 Attack)' },
      { id: '4', label: '5: 地裂斩 (地 Attack)' },
    ],
    min: 1,
    max: 1,
  } satisfies Prompt);
}

// ============================================================
// 百鬼夜行 (sc_hundred_night) - Response skill after attack hit, remove Youli with branching
// ============================================================

export function hundredNightScenario(options: {
  hasCrystal?: boolean;
} = {}): ProtocolHarnessScenario {
  const characters = [spiritCasterCharacter, allyCharacter, enemyCharacter, enemy2Character, enemy3Character];

  const spiritCaster = spiritCasterPlayerView({
    is_active: true,
    crystal: options.hasCrystal ? 1 : 0,
  });
  spiritCaster.field = [spiritCasterPowerCover(0)];

  const players = [
    spiritCaster,
    playerView({
      id: ENEMY_PLAYER_ID,
      name: 'Enemy E1',
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
      name: 'Enemy E2',
      camp: 'Blue',
      role: 'villain',
      hand: [],
      hand_count: 2,
      heal: 1,
      max_heal: 2,
      is_active: false,
    }),
    playerView({
      id: ENEMY_3_PLAYER_ID,
      name: 'Enemy E3',
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
    myPlayerId: SC_PLAYER_ID,
    myPlayerName: 'E2E Spirit Caster',
    characters,
    players: [
      playerInfo({ id: SC_PLAYER_ID, name: 'E2E Spirit Caster', camp: 'Red', char_role: 'spirit_caster', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy E1', camp: 'Blue', char_role: 'villain' }),
      playerInfo({ id: ENEMY_2_PLAYER_ID, name: 'Enemy E2', camp: 'Blue', char_role: 'villain' }),
      playerInfo({ id: ENEMY_3_PLAYER_ID, name: 'Enemy E3', camp: 'Blue', char_role: 'villain' }),
    ],
    initialState: syncState({
      turn_player_id: SC_PLAYER_ID,
      turn_stage: 'ActionExecution',
      characters,
      players,
    }),
  };
}

// 百鬼夜行响应确认弹框（模拟响应技能触发）
export function hundredNightConfirmPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: SC_PLAYER_ID,
    message: '【百鬼夜行】是否发动？',
    choice_type: 'sc_hundred_night_confirm',
    skill_id: SC_HUNDRED_NIGHT_SKILL_ID,
    options: [
      { id: '0', label: '是' },
      { id: '1', label: '否' },
    ],
    min: 1,
    max: 1,
  } satisfies Prompt);
}

// 灵力崩解附加选择（百鬼夜行流程中，与雷鸣共用 sc_spiritual_collapse_confirm）
export function hundredNightManaCollapsePrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: SC_PLAYER_ID,
    message: '【灵力崩解】是否消耗1点水晶（红宝石可替代），使本次每段伤害额外+1？',
    choice_type: 'sc_spiritual_collapse_confirm',
    skill_id: SC_HUNDRED_NIGHT_SKILL_ID,
    options: [
      { id: '0', label: '是' },
      { id: '1', label: '否' },
    ],
    min: 1,
    max: 1,
  } satisfies Prompt);
}

// 妖力移除选择（盖牌选择 UI，不在弹框中显示按钮）
export function hundredNightRemoveYouliPrompt(options: {
  youliCards?: { name: string; element: string }[];
} = {}): WsMessage {
  const youliCards = options.youliCards ?? [
    { name: '火焰斩', element: 'Fire' },
  ];
  return requireActionMessage({
    type: 'confirm',
    player_id: SC_PLAYER_ID,
    message: '【百鬼夜行】请选择要移除的1个妖力：',
    choice_type: 'sc_hundred_night_power',
    skill_id: SC_HUNDRED_NIGHT_SKILL_ID,
    options: youliCards.map((c, idx) => ({
      id: `${idx}`,
      label: `妖力[${idx}] ${c.name}（${c.element}系）`,
      field_index: idx,
    })),
    presentation: { kind: 'card_picker', layout: 'field_cover' },
    min: 1,
    max: 1,
  } satisfies Prompt);
}

// 火系妖力分支选择（是否展示并发动群体伤害）
export function hundredNightFireBranchPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: SC_PLAYER_ID,
    message: '【百鬼夜行】移除的是火系妖力，是否展示并改为范围伤害？',
    choice_type: 'sc_hundred_night_fire_reveal',
    skill_id: SC_HUNDRED_NIGHT_SKILL_ID,
    options: [
      { id: '0', label: '展示并改为范围伤害' },
      { id: '1', label: '不展示，改为单体伤害' },
    ],
    presentation: { kind: 'branch_select', layout: 'overlay' },
    min: 1,
    max: 1,
  } satisfies Prompt);
}

// 单体目标选择（非火系/不展示）
export function hundredNightSingleTargetPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: SC_PLAYER_ID,
    message: '【百鬼夜行】请选择1点法术伤害目标：',
    choice_type: 'sc_hundred_night_target',
    skill_id: SC_HUNDRED_NIGHT_SKILL_ID,
    options: [
      { id: '0', label: 'Enemy E1' },
      { id: '1', label: 'Enemy E2' },
      { id: '2', label: 'Enemy E3' },
    ],
    min: 1,
    max: 1,
  } satisfies Prompt);
}

// 群体豁免目标选择（火系展示）- option IDs are sequential indices matching backend BuildTargetChoicePrompt
export function hundredNightAoeExemptTargetPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: SC_PLAYER_ID,
    message: '【百鬼夜行】请选择第 1/2 名排除目标：',
    choice_type: 'sc_hundred_night_exclude_pick',
    skill_id: SC_HUNDRED_NIGHT_SKILL_ID,
    options: [
      { id: '0', label: 'E2E Spirit Caster' },
      { id: '1', label: 'Enemy E1' },
      { id: '2', label: 'Enemy E2' },
      { id: '3', label: 'Enemy E3' },
    ],
    min: 1,
    max: 1,
  } satisfies Prompt);
}
