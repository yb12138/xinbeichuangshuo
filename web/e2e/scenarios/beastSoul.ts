// ============================================================
// Beast Samurai (兽灵武士) Protocol Harness Scenarios
// 与后端 internal/engine/player/beast_samurai/choices.go 对齐：
// - 响应技能（一击无念 / 兽魂警戒 / 兽返 / 逆反居合斩）走通用 choose_skill 弹框
// - 启动技能（御魂流居合式）走通用 choose_skill 弹框
// - X 选择 / 摸弃选择走 bs_* 专属 choice_type（option id 为数字字符串）
// - 武者残心 (bs_warrior_zanshin) 后端 ResponseSilent，无前端弹框，不在此文件 mock
// - "不屈意志" 属于剑帝技能，已从兽灵武士 e2e 移除
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
export const ENEMY_PLAYER_ID = 'enemy_1';

// ---- Skill IDs (与 internal/data/characters.go beast_samurai 段一致) ----
export const BSW_ONE_STRIKE_SKILL_ID = 'bs_one_strike_no_thought';
export const BSW_BEAST_SOUL_ALERT_SKILL_ID = 'bs_beast_soul_alert';
export const BSW_BEAST_RETURN_SKILL_ID = 'bs_beast_return';
export const BSW_REVERSAL_IAIJUTSU_SKILL_ID = 'bs_reversal_iaijutsu';
export const BSW_IAIJUTSU_STYLE_SKILL_ID = 'bs_iaijutsu_style';

// ---- Beast Samurai character definition ----
const beastSoulCharacter = characterView({
  id: 'beast_samurai',
  name: '兽灵武士',
  title: '野兽之魂',
  faction: '星杯',
  skills: [
    {
      id: BSW_ONE_STRIKE_SKILL_ID,
      title: '一击无念',
      description: '攻击行动结束时若残心≥4可发动',
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
      description: '响应技能：消耗1兽魂指定 1 名角色弃 1 张牌',
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
      description: '响应技能，选 X 移除兽魂',
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
      description: '响应技能，选 X 移除兽魂改写攻击为弃牌效果',
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
      title: '御魂流居合式',
      description: '启动技能，消耗宝石，摸 1 或弃 1',
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
    role: 'beast_samurai',
    hand: beastSoulHand(),
    hand_count: beastSoulHand().length,
    heal: options.heal ?? 2,
    max_heal: options.max_heal ?? 4,
    is_active: options.is_active ?? true,
    gem: options.gems ?? 0,
    crystal: options.crystals ?? 0,
    tokens: {
      bs_zanshin: options.zanshin ?? 0,
      bs_beast_soul: options.beast_souls ?? 0,
    },
  });
}

// 后端的响应技能 / 启动技能选择统一通过 choose_skill 入口下发
// （参见 internal/engine/interrupt_prompt_framework.go buildResponseSkillPrompt /
// buildStartupSkillPrompt）。"确认/跳过" 不是独立 choice_type，
// 而是 choose_skill 中的两个选项。
function beastSoulSkillChoicePrompt(skillId: string, title: string, message: string): WsMessage {
  return requireActionMessage({
    type: 'choose_skill',
    player_id: BSW_PLAYER_ID,
    message,
    options: [
      { id: skillId, label: title, button_label: title, hint: `发动【${title}】` },
      { id: 'skip', label: '跳过', button_label: '跳过', hint: '不发动响应技能' },
    ],
    presentation: { kind: 'skill_choice', layout: 'overlay', numeric_base: 0 },
    min: 1,
    max: 1,
  } satisfies Prompt);
}

// ============================================================
// 一击无念 (bs_one_strike_no_thought) - 残心≥4 时由响应技能链路发动
// ============================================================

export function oneStrikeScenario(options: { zanshin?: number } = {}): ProtocolHarnessScenario {
  const zanshin = options.zanshin ?? 4;
  const characters = [beastSoulCharacter, enemyCharacter];

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
      playerInfo({ id: BSW_PLAYER_ID, name: 'E2E Beast Soul', camp: 'Red', char_role: 'beast_samurai', is_host: true }),
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

export function oneStrikeResponsePrompt(): WsMessage {
  return beastSoulSkillChoicePrompt(
    BSW_ONE_STRIKE_SKILL_ID,
    '一击无念',
    '你触发了响应技能【一击无念】，请选择是否发动。',
  );
}

// ============================================================
// 兽魂警戒 (bs_beast_soul_alert) - 响应 → 选目标 → 让目标弃 1 张牌
// ============================================================

export function beastSoulAlertScenario(options: { beast_souls?: number } = {}): ProtocolHarnessScenario {
  const beast_souls = options.beast_souls ?? 2;
  const characters = [beastSoulCharacter, enemyCharacter];

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
      playerInfo({ id: BSW_PLAYER_ID, name: 'E2E Beast Soul', camp: 'Red', char_role: 'beast_samurai', is_host: true }),
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

export function beastSoulAlertResponsePrompt(): WsMessage {
  return beastSoulSkillChoicePrompt(
    BSW_BEAST_SOUL_ALERT_SKILL_ID,
    '兽魂警戒',
    '你触发了响应技能【兽魂警戒】，请选择是否发动。',
  );
}

export function beastSoulAlertTargetPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: BSW_PLAYER_ID,
    message: '【兽魂警戒】请选择 1 名让其弃 1 张牌的角色：',
    choice_type: 'bs_alert_target',
    options: [
      { id: ENEMY_PLAYER_ID, label: 'Enemy Bot', button_label: '选择' },
    ],
    min: 1,
    max: 1,
    presentation: { kind: 'target_picker', target_filter: 'custom', numeric_base: 0 },
  } satisfies Prompt);
}

export function beastSoulAlertDiscardPrompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_cards',
    player_id: ENEMY_PLAYER_ID,
    message: '【兽魂警戒】请选择并展示弃置1张手牌：',
    choice_type: 'bs_alert_source_discard',
    options: [
      { id: '0', label: '1: 神秘手牌', button_label: '选择', card_id: 'enemy-hidden-card-1' },
    ],
    min: 1,
    max: 1,
    presentation: { kind: 'card_picker', card_source: 'hand', card_filter: 'option_limited', numeric_base: 0 },
  } satisfies Prompt);
}

// ============================================================
// 兽返 (bs_beast_return) - 响应 → 选 X 移除兽魂 → 自己弃 X 张 → 来源弃 1 张
// ============================================================

export function beastReturnScenario(options: { beast_souls?: number } = {}): ProtocolHarnessScenario {
  const beast_souls = options.beast_souls ?? 3;
  const characters = [beastSoulCharacter, enemyCharacter];

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
      playerInfo({ id: BSW_PLAYER_ID, name: 'E2E Beast Soul', camp: 'Red', char_role: 'beast_samurai', is_host: true }),
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

export function beastReturnResponsePrompt(): WsMessage {
  return beastSoulSkillChoicePrompt(
    BSW_BEAST_RETURN_SKILL_ID,
    '兽返',
    '你触发了响应技能【兽返】，请选择是否发动。',
  );
}

// 后端 buildBeastReturnXPrompt 选项为 X=0..maxX（含「不移除兽魂」），option id 为字符串数字。
export function beastReturnXPrompt(xMax: number): WsMessage {
  const options: { id: string; label: string; button_label: string }[] = [];
  for (let i = 0; i <= xMax; i++) {
    const label = i === 0 ? 'X=0（不移除兽魂）' : `X=${i}`;
    options.push({ id: `${i}`, label, button_label: String(i) });
  }
  return requireActionMessage({
    type: 'confirm',
    player_id: BSW_PLAYER_ID,
    message: `【兽返】请选择要移除的兽魂数量（0-${xMax}）：`,
    choice_type: 'bs_beast_return_x',
    options,
    presentation: { kind: 'numeric', numeric_base: 0 },
    min: 1,
    max: 1,
  } satisfies Prompt);
}

// 兽返步骤3：自己弃 X 张牌（后端 buildDiscardPrompt）
export function beastReturnSelfDiscardPrompt(discardCount: number): WsMessage {
  const hand = beastSoulHand();
  const options = hand.map((c, i) => ({
    id: `${i}`,
    label: `${i + 1}: ${c.name} (${c.element} ${c.type})`,
    button_label: '选择',
    card_id: c.id,
  }));
  return requireActionMessage({
    type: 'choose_cards',
    player_id: BSW_PLAYER_ID,
    message: `【兽返】请选择弃置${discardCount}张手牌：`,
    choice_type: 'bs_beast_return_self_discard',
    options,
    min: discardCount,
    max: discardCount,
    presentation: { kind: 'card_picker', card_source: 'hand', card_filter: 'option_limited', numeric_base: 0 },
  } satisfies Prompt);
}

// 兽返步骤4：伤害来源弃 1 张牌（后端 buildDiscardPrompt）
export function beastReturnSourceDiscardPrompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_cards',
    player_id: ENEMY_PLAYER_ID,
    message: '【兽返】请选择弃置1张手牌：',
    choice_type: 'bs_beast_return_source_discard',
    options: [
      { id: '0', label: '1: 神秘手牌', button_label: '选择', card_id: 'enemy-hidden-card-1' },
    ],
    min: 1,
    max: 1,
    presentation: { kind: 'card_picker', card_source: 'hand', card_filter: 'option_limited', numeric_base: 0 },
  } satisfies Prompt);
}

// ============================================================
// 逆反居合斩 (bs_reversal_iaijutsu) - 响应 → 选 X → 攻击目标弃 X+2 张
// （后端直接以攻击目标为弃牌对象，不需额外目标选择）
// ============================================================

export function reversalIaijutsuScenario(options: { beast_souls?: number } = {}): ProtocolHarnessScenario {
  const beast_souls = options.beast_souls ?? 3;
  const characters = [beastSoulCharacter, enemyCharacter];

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
      playerInfo({ id: BSW_PLAYER_ID, name: 'E2E Beast Soul', camp: 'Red', char_role: 'beast_samurai', is_host: true }),
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

export function reversalIaijutsuResponsePrompt(): WsMessage {
  return beastSoulSkillChoicePrompt(
    BSW_REVERSAL_IAIJUTSU_SKILL_ID,
    '逆反居合斩',
    '你触发了响应技能【逆反居合斩】，请选择是否发动。',
  );
}

// 后端 buildReversalXPrompt：X=0..maxX，文案标注 "目标将弃置 X+2 张手牌"。
export function reversalIaijutsuXPrompt(xMax: number): WsMessage {
  const options: { id: string; label: string; button_label: string }[] = [];
  for (let i = 0; i <= xMax; i++) {
    options.push({ id: `${i}`, label: `X=${i}（目标将弃置${i + 2}张手牌）`, button_label: String(i) });
  }
  return requireActionMessage({
    type: 'confirm',
    player_id: BSW_PLAYER_ID,
    message: `【逆反居合斩】请选择要移除的兽魂数量（0-${xMax}）：`,
    choice_type: 'bs_reversal_x',
    options,
    presentation: { kind: 'numeric', numeric_base: 0 },
    min: 1,
    max: 1,
  } satisfies Prompt);
}

// 攻击目标弃牌：后端 buildDiscardPrompt 走 PromptChooseCards，每次 Max=1 迭代消耗。
export function reversalIaijutsuTargetDiscardPrompt(discardCount: number): WsMessage {
  return requireActionMessage({
    type: 'choose_cards',
    player_id: ENEMY_PLAYER_ID,
    message: `【逆反居合斩】请选择弃置${discardCount}张手牌：`,
    choice_type: 'bs_reversal_target_discard',
    options: [
      { id: '0', label: '1: 神秘手牌', button_label: '选择', card_id: 'enemy-hidden-card-1' },
    ],
    min: 1,
    max: 1,
    presentation: { kind: 'card_picker', card_source: 'hand', card_filter: 'option_limited', numeric_base: 0 },
  } satisfies Prompt);
}

// ============================================================
// 御魂流居合式 (bs_iaijutsu_style) - 启动技能：摸 1 / 弃 1
// ============================================================

export function iaijutsuStyleScenario(options: { gems?: number } = {}): ProtocolHarnessScenario {
  const gems = options.gems ?? 1;
  const characters = [beastSoulCharacter, enemyCharacter];

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
      playerInfo({ id: BSW_PLAYER_ID, name: 'E2E Beast Soul', camp: 'Red', char_role: 'beast_samurai', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy Bot', camp: 'Blue', char_role: 'villain' }),
    ],
    initialState: syncState({
      turn_player_id: BSW_PLAYER_ID,
      turn_stage: 'ActionExecution',
      available_skills: [availableSkill({ id: BSW_IAIJUTSU_STYLE_SKILL_ID, title: '御魂流居合式', cost_gem: 1 })],
      characters,
      players,
    }),
  };
}

export function iaijutsuStyleStartupPrompt(): WsMessage {
  return beastSoulSkillChoicePrompt(
    BSW_IAIJUTSU_STYLE_SKILL_ID,
    '御魂流居合式',
    '你可以发动启动技能，请选择 1 个发动，或跳过。',
  );
}

// 后端 buildIaijutsuStyleModePrompt：摸1张/弃1张，option id 为 "0"/"1"。
export function iaijutsuStyleModePrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: BSW_PLAYER_ID,
    message: '【御魂流居合式】请选择"摸1张牌"或"弃1张牌"：',
    choice_type: 'bs_iaijutsu_style_mode',
    options: [
      { id: '0', label: '摸1张牌', button_label: '摸牌' },
      { id: '1', label: '弃1张牌', button_label: '弃牌' },
    ],
    presentation: { kind: 'branch_select', layout: 'overlay', numeric_base: 0 },
    min: 1,
    max: 1,
  } satisfies Prompt);
}

// 后端 buildDiscardPrompt（御魂流分支②）：仅弃 1 张牌。
export function iaijutsuStyleDiscardPrompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_cards',
    player_id: BSW_PLAYER_ID,
    message: '【御魂流居合式】请选择弃置1张手牌：',
    choice_type: 'bs_iaijutsu_style_discard',
    options: [
      { id: '0', label: '1: 火焰斩 (火 Attack)', button_label: '选择', card_id: 'card_1' },
      { id: '1', label: '2: 水涟斩 (水 Attack)', button_label: '选择', card_id: 'card_2' },
      { id: '2', label: '3: 风刃 (风 Attack)', button_label: '选择', card_id: 'card_3' },
      { id: '3', label: '4: 寒冰箭 (水 Magic)', button_label: '选择', card_id: 'card_4' },
    ],
    min: 1,
    max: 1,
    presentation: { kind: 'card_picker', card_source: 'hand', card_filter: 'option_limited', numeric_base: 0 },
  } satisfies Prompt);
}
