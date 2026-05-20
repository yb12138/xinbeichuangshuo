// ============================================================
// Fighter (格斗家) Protocol Harness Scenarios
// ============================================================

import type { Card, Prompt } from '../../src/types/game';
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
export const FIGHTER_PLAYER_ID = 'fighter_player';
export const ALLY_PLAYER_ID = 'ally_1';
export const ENEMY_PLAYER_ID = 'enemy_1';

// ---- Skill IDs (与 internal/data/characters.go fighter 段一致) ----
export const FIGHTER_CHARGE_ATTACK_SKILL_ID = 'fighter_charge_strike';
export const FIGHTER_BURST_CRASH_SKILL_ID = 'fighter_burst_crash';
export const FIGHTER_BULLET_SKILL_ID = 'fighter_psi_bullet';
export const FIGHTER_HUNDRED_DRAGON_SKILL_ID = 'fighter_hundred_dragon';
export const FIGHTER_HEAVEN_DRIVE_SKILL_ID = 'fighter_war_god_drive';

// ---- Fighter character definition ----
const fighterCharacter = characterView({
  id: 'fighter',
  name: '格斗家',
  title: '斗气战士',
  faction: '星杯',
  skills: [
    {
      id: FIGHTER_CHARGE_ATTACK_SKILL_ID,
      title: '蓄力一击',
      description: '攻击时发动，若斗气未到上限，伤害+X',
      type: 0,
      min_targets: 0,
      max_targets: 0,
      target_type: 0,
      cost_gem: 0,
      cost_crystal: 0,
      cost_discards: 0,
    },
    {
      id: FIGHTER_BURST_CRASH_SKILL_ID,
      title: '气绝崩击',
      description: '攻击时发动，消耗斗气，效果由后端结算',
      type: 0,
      min_targets: 0,
      max_targets: 0,
      target_type: 0,
      cost_gem: 0,
      cost_crystal: 0,
      cost_discards: 0,
    },
    {
      id: FIGHTER_BULLET_SKILL_ID,
      title: '念弹',
      description: '法术行动后发动',
      type: 0,
      min_targets: 0,
      max_targets: 0,
      target_type: 0,
      cost_gem: 0,
      cost_crystal: 0,
      cost_discards: 0,
    },
    {
      id: FIGHTER_HUNDRED_DRAGON_SKILL_ID,
      title: '百式幻龙拳',
      description: '启动技能，消耗3点斗气',
      type: 2,
      min_targets: 0,
      max_targets: 0,
      target_type: 0,
      cost_gem: 0,
      cost_crystal: 0,
      cost_discards: 0,
    },
    {
      id: FIGHTER_HEAVEN_DRIVE_SKILL_ID,
      title: '斗神天驱',
      description: '启动技能，消耗水晶',
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

// ---- Helper functions ----
function fighterHand() {
  return [
    card({ id: 'card_1', name: '火焰斩', type: 'Attack', element: 'Fire' }),
    card({ id: 'card_2', name: '火焰斩', type: 'Attack', element: 'Fire' }),
    card({ id: 'card_3', name: '水涟斩', type: 'Attack', element: 'Water' }),
    card({ id: 'card_4', name: '寒冰箭', type: 'Magic', element: 'Water' }),
    card({ id: 'card_5', name: '圣光', type: 'Magic', element: 'Light' }),
  ];
}

function fighterPlayerView(options: {
  is_active?: boolean;
  qi?: number;
  max_qi?: number;
  gems?: number;
  crystals?: number;
} = {}) {
  return playerView({
    id: FIGHTER_PLAYER_ID,
    name: 'E2E Fighter',
    camp: 'Red',
    role: 'fighter',
    hand: fighterHand(),
    hand_count: fighterHand().length,
    heal: 2,
    max_heal: 4,
    is_active: options.is_active ?? true,
    gem: options.gems ?? 0,
    crystal: options.crystals ?? 0,
    tokens: {
      fighter_qi: options.qi ?? 0,
    },
  });
}

// 格斗家的响应技能 / 启动技能选择统一通过 choose_skill 入口下发
// （参见 internal/engine/interrupt_prompt_framework.go buildResponseSkillPrompt /
// buildStartupSkillPrompt）。"发动 / 跳过" 不是独立 confirm，
// 而是 choose_skill 中的选项。
function fighterSkillChoicePrompt(
  skills: { id: string; title: string }[],
  message: string,
): WsMessage {
  const options = skills.map((s) => ({
    id: s.id,
    label: s.title,
    button_label: '发动',
    hint: `发动【${s.title}】`,
  }));
  options.push({ id: 'skip', label: '跳过', button_label: '跳过', hint: '不发动响应技能' });
  return requireActionMessage({
    type: 'choose_skill',
    player_id: FIGHTER_PLAYER_ID,
    message,
    options,
    presentation: { kind: 'skill_choice', layout: 'overlay', numeric_base: 0 },
    min: 1,
    max: 1,
  } satisfies Prompt);
}

// ============================================================
// 蓄力一击 (fighter_charge_strike) - Attack response when qi not maxed
// ============================================================

export function chargeAttackScenario(options: { qi?: number } = {}): ProtocolHarnessScenario {
  const qi = options.qi ?? 1;
  const characters = [fighterCharacter, allyCharacter, enemyCharacter];

  const fighter = fighterPlayerView({ qi, is_active: true });

  const players = [
    fighter,
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
    myPlayerId: FIGHTER_PLAYER_ID,
    myPlayerName: 'E2E Fighter',
    characters,
    players: [
      playerInfo({ id: FIGHTER_PLAYER_ID, name: 'E2E Fighter', camp: 'Red', char_role: 'fighter', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy Bot', camp: 'Blue', char_role: 'villain' }),
    ],
    initialState: syncState({
      turn_player_id: FIGHTER_PLAYER_ID,
      turn_stage: 'ActionExecution',
      available_skills: [],
      characters,
      players,
    }),
  };
}

export function chargeAttackConfirmPrompt(): WsMessage {
  return fighterSkillChoicePrompt(
    [{ id: FIGHTER_CHARGE_ATTACK_SKILL_ID, title: '蓄力一击' }],
    '你触发了响应技能【蓄力一击】，请选择是否发动。',
  );
}

// ============================================================
// 气绝崩击 (fighter_burst_crash) - Attack response when qi >= 1
// ============================================================

export function burstCrashScenario(options: { qi?: number } = {}): ProtocolHarnessScenario {
  const qi = options.qi ?? 2;
  const characters = [fighterCharacter, allyCharacter, enemyCharacter];

  const fighter = fighterPlayerView({ qi, is_active: true });

  const players = [
    fighter,
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
    myPlayerId: FIGHTER_PLAYER_ID,
    myPlayerName: 'E2E Fighter',
    characters,
    players: [
      playerInfo({ id: FIGHTER_PLAYER_ID, name: 'E2E Fighter', camp: 'Red', char_role: 'fighter', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy Bot', camp: 'Blue', char_role: 'villain' }),
    ],
    initialState: syncState({
      turn_player_id: FIGHTER_PLAYER_ID,
      turn_stage: 'ActionExecution',
      available_skills: [],
      characters,
      players,
    }),
  };
}

export function burstCrashConfirmPrompt(): WsMessage {
  return fighterSkillChoicePrompt(
    [{ id: FIGHTER_BURST_CRASH_SKILL_ID, title: '气绝崩击' }],
    '你触发了响应技能【气绝崩击】，请选择是否发动。',
  );
}

// ============================================================
// 蓄力一击 & 气绝崩击 互斥选择 - Attack response when both can trigger
// ============================================================

export function attackSkillChoiceScenario(options: { qi?: number } = {}): ProtocolHarnessScenario {
  const qi = options.qi ?? 2;
  const characters = [fighterCharacter, allyCharacter, enemyCharacter];

  const fighter = fighterPlayerView({ qi, is_active: true });

  const players = [
    fighter,
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
    myPlayerId: FIGHTER_PLAYER_ID,
    myPlayerName: 'E2E Fighter',
    characters,
    players: [
      playerInfo({ id: FIGHTER_PLAYER_ID, name: 'E2E Fighter', camp: 'Red', char_role: 'fighter', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy Bot', camp: 'Blue', char_role: 'villain' }),
    ],
    initialState: syncState({
      turn_player_id: FIGHTER_PLAYER_ID,
      turn_stage: 'ActionExecution',
      available_skills: [],
      characters,
      players,
    }),
  };
}

export function attackSkillChoicePrompt(): WsMessage {
  return fighterSkillChoicePrompt(
    [
      { id: FIGHTER_CHARGE_ATTACK_SKILL_ID, title: '蓄力一击' },
      { id: FIGHTER_BURST_CRASH_SKILL_ID, title: '气绝崩击' },
    ],
    '你触发了 2 个响应技能，请选择 1 个发动，或跳过。',
  );
}

// ============================================================
// 念弹 (fighter_bullet) - Post-magic action response
// ============================================================

export function bulletScenario(): ProtocolHarnessScenario {
  const characters = [fighterCharacter, allyCharacter, enemyCharacter];

  const fighter = fighterPlayerView({ qi: 1, is_active: true });

  const players = [
    fighter,
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
    myPlayerId: FIGHTER_PLAYER_ID,
    myPlayerName: 'E2E Fighter',
    characters,
    players: [
      playerInfo({ id: FIGHTER_PLAYER_ID, name: 'E2E Fighter', camp: 'Red', char_role: 'fighter', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy Bot', camp: 'Blue', char_role: 'villain' }),
    ],
    initialState: syncState({
      turn_player_id: FIGHTER_PLAYER_ID,
      turn_stage: 'ActionExecution',
      available_skills: [],
      characters,
      players,
    }),
  };
}

export function bulletConfirmPrompt(): WsMessage {
  return fighterSkillChoicePrompt(
    [{ id: FIGHTER_BULLET_SKILL_ID, title: '念弹' }],
    '你触发了响应技能【念弹】，请选择是否发动。',
  );
}

// 念弹目标选择prompt - 发动确认后后端推送 fighter_psi_bullet_target
export function bulletTargetPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: FIGHTER_PLAYER_ID,
    message: '【念弹】请选择1名目标对手：',
    choice_type: 'fighter_psi_bullet_target',
    presentation: { kind: 'target_picker', target_filter: 'custom', numeric_base: 0 },
    options: [
      { id: ENEMY_PLAYER_ID, target_id: ENEMY_PLAYER_ID, label: 'Enemy Bot', button_label: '选择' },
    ],
    min: 1,
    max: 1,
  } satisfies Prompt);
}

// ============================================================
// 百式幻龙拳 (fighter_hundred_dragon) - Start skill with 3+ qi
// ============================================================

export function hundredDragonScenario(options: { qi?: number } = {}): ProtocolHarnessScenario {
  const qi = options.qi ?? 3;
  const characters = [fighterCharacter, allyCharacter, enemyCharacter];

  const fighter = fighterPlayerView({ qi, is_active: true });

  const players = [
    fighter,
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
    myPlayerId: FIGHTER_PLAYER_ID,
    myPlayerName: 'E2E Fighter',
    characters,
    players: [
      playerInfo({ id: FIGHTER_PLAYER_ID, name: 'E2E Fighter', camp: 'Red', char_role: 'fighter', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy Bot', camp: 'Blue', char_role: 'villain' }),
    ],
    initialState: syncState({
      turn_player_id: FIGHTER_PLAYER_ID,
      turn_stage: 'StartPhase',
      available_skills: [],
      characters,
      players,
    }),
  };
}

function fighterStartupSkillPrompt(
  skills: { id: string; title: string }[],
): WsMessage {
  const options = skills.map((s) => ({
    id: s.id,
    label: s.title,
    button_label: '发动',
    hint: `发动【${s.title}】`,
  }));
  options.push({ id: 'skip', label: '跳过', button_label: '跳过', hint: '本回合不发动启动技能' });
  return requireActionMessage({
    type: 'choose_skill',
    player_id: FIGHTER_PLAYER_ID,
    message: '你可以发动启动技能，请选择 1 个发动，或跳过。',
    options,
    presentation: { kind: 'skill_choice', layout: 'overlay', numeric_base: 0 },
    min: 1,
    max: 1,
  } satisfies Prompt);
}

export function hundredDragonConfirmPrompt(): WsMessage {
  return fighterStartupSkillPrompt([
    { id: FIGHTER_HUNDRED_DRAGON_SKILL_ID, title: '百式幻龙拳' },
  ]);
}

// 百式幻龙拳目标锁定prompt - 发动确认后后端推送 fighter_hundred_dragon_target
export function hundredDragonTargetPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: FIGHTER_PLAYER_ID,
    message: '【百式幻龙拳】请选择本行动阶段锁定的目标角色：',
    choice_type: 'fighter_hundred_dragon_target',
    presentation: { kind: 'target_picker', target_filter: 'custom', numeric_base: 0 },
    options: [
      { id: ENEMY_PLAYER_ID, target_id: ENEMY_PLAYER_ID, label: 'Enemy Bot', button_label: '选择' },
    ],
    min: 1,
    max: 1,
  } satisfies Prompt);
}

// ============================================================
// 斗神天驱 (fighter_heaven_drive) - Start skill with crystal
// ============================================================

export function heavenDriveScenario(options: { crystals?: number } = {}): ProtocolHarnessScenario {
  const crystals = options.crystals ?? 1;
  const characters = [fighterCharacter, allyCharacter, enemyCharacter];

  const fighter = fighterPlayerView({ qi: 1, crystals, is_active: true });

  const players = [
    fighter,
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
    myPlayerId: FIGHTER_PLAYER_ID,
    myPlayerName: 'E2E Fighter',
    characters,
    players: [
      playerInfo({ id: FIGHTER_PLAYER_ID, name: 'E2E Fighter', camp: 'Red', char_role: 'fighter', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy Bot', camp: 'Blue', char_role: 'villain' }),
    ],
    initialState: syncState({
      turn_player_id: FIGHTER_PLAYER_ID,
      turn_stage: 'StartPhase',
      available_skills: [],
      characters,
      players,
    }),
  };
}

export function heavenDriveConfirmPrompt(): WsMessage {
  return fighterStartupSkillPrompt([
    { id: FIGHTER_HEAVEN_DRIVE_SKILL_ID, title: '斗神天驱' },
  ]);
}

// 斗神天驱弃牌场景 - 手牌>3张需要弃牌至3张
export function heavenDriveDiscardScenario(options: { crystals?: number; handCount?: number } = {}): ProtocolHarnessScenario {
  const crystals = options.crystals ?? 1;
  const handCount = options.handCount ?? 5;
  const characters = [fighterCharacter, allyCharacter, enemyCharacter];

  // 创建多于3张的手牌
  const hand: Card[] = [];
  for (let i = 1; i <= handCount; i++) {
    hand.push(card({ id: `card_${i}`, name: `手牌${i}`, type: 'Attack', element: 'Fire' }));
  }

  const fighter = playerView({
    id: FIGHTER_PLAYER_ID,
    name: 'E2E Fighter',
    camp: 'Red',
    role: 'fighter',
    hand,
    hand_count: hand.length,
    heal: 2,
    max_heal: 4,
    is_active: true,
    gem: 0,
    crystal: crystals,
    tokens: {
      fighter_qi: 1,
    },
  });

  const players = [
    fighter,
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
    myPlayerId: FIGHTER_PLAYER_ID,
    myPlayerName: 'E2E Fighter',
    characters,
    players: [
      playerInfo({ id: FIGHTER_PLAYER_ID, name: 'E2E Fighter', camp: 'Red', char_role: 'fighter', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy Bot', camp: 'Blue', char_role: 'villain' }),
    ],
    initialState: syncState({
      turn_player_id: FIGHTER_PLAYER_ID,
      turn_stage: 'StartPhase',
      available_skills: [],
      characters,
      players,
    }),
  };
}

// 斗神天驱弃牌prompt - 手牌>3张时后端推送 system_discard_cards
export function heavenDriveDiscardPrompt(handCount: number = 5): WsMessage {
  const options: Array<{ id: string; label: string; button_label: string; card_id: string }> = [];
  for (let i = 0; i < handCount; i++) {
    const cardID = `card_${i + 1}`;
    options.push({
      id: cardID,
      label: `${i + 1}: 手牌${i + 1}（火系 攻击）`,
      button_label: '选择',
      card_id: cardID,
    });
  }
  const discardCount = handCount - 3;
  return requireActionMessage({
    type: 'choose_cards',
    player_id: FIGHTER_PLAYER_ID,
    message: '【斗神天驱】请选择需要弃置的手牌：',
    choice_type: 'system_discard_cards',
    skill_id: FIGHTER_HEAVEN_DRIVE_SKILL_ID,
    options,
    min: discardCount,
    max: discardCount,
    presentation: { kind: 'card_picker', card_source: 'hand', card_filter: 'overflow_discard', numeric_base: 0 },
  } satisfies Prompt);
}

// ============================================================
// 百式幻龙拳 & 斗神天驱 互斥选择 - Start skill choice when both can trigger
// ============================================================

export function startSkillChoiceScenario(options: { qi?: number; crystals?: number } = {}): ProtocolHarnessScenario {
  const qi = options.qi ?? 3;
  const crystals = options.crystals ?? 1;
  const characters = [fighterCharacter, allyCharacter, enemyCharacter];

  const fighter = fighterPlayerView({ qi, crystals, is_active: true });

  const players = [
    fighter,
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
    myPlayerId: FIGHTER_PLAYER_ID,
    myPlayerName: 'E2E Fighter',
    characters,
    players: [
      playerInfo({ id: FIGHTER_PLAYER_ID, name: 'E2E Fighter', camp: 'Red', char_role: 'fighter', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy Bot', camp: 'Blue', char_role: 'villain' }),
    ],
    initialState: syncState({
      turn_player_id: FIGHTER_PLAYER_ID,
      turn_stage: 'StartPhase',
      available_skills: [],
      characters,
      players,
    }),
  };
}

export function startSkillChoicePrompt(): WsMessage {
  return fighterStartupSkillPrompt([
    { id: FIGHTER_HUNDRED_DRAGON_SKILL_ID, title: '百式幻龙拳' },
    { id: FIGHTER_HEAVEN_DRIVE_SKILL_ID, title: '斗神天驱' },
  ]);
}
