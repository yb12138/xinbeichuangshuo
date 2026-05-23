// ============================================================
// SoulSorcerer (灵魂术士) Protocol Harness Scenarios
// ============================================================

import type { Prompt, AvailableSkill } from '../../src/types/game';
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
export const SS_PLAYER_ID = 'ss_player';
export const ALLY_PLAYER_ID = 'ally_1';
export const ALLY_2_PLAYER_ID = 'ally_2';
export const ENEMY_PLAYER_ID = 'enemy_1';

// ---- Skill IDs ----
export const SS_SOUL_RECALL_SKILL_ID = 'ss_soul_recall';
export const SS_SOUL_CONVERT_SKILL_ID = 'ss_soul_convert';
export const SS_SOUL_MIRROR_SKILL_ID = 'ss_soul_mirror';
export const SS_SOUL_BLAST_SKILL_ID = 'ss_soul_blast';
export const SS_SOUL_GRANT_SKILL_ID = 'ss_soul_grant';
export const SS_SOUL_LINK_SKILL_ID = 'ss_soul_link';
export const SS_SOUL_AMP_SKILL_ID = 'ss_soul_amp';

// ---- SoulSorcerer character definition ----
const soulSorcererCharacter = characterView({
  id: 'soul_sorcerer',
  name: '灵魂术士',
  title: '幻',
  faction: '幻',
  skills: [
    {
      id: SS_SOUL_RECALL_SKILL_ID,
      title: '灵魂召还',
      description: '弃X张法术牌[展示]，你+X点[蓝色灵魂]（上限6）。',
      type: 1, // SkillTypeAction
      min_targets: 0,
      max_targets: 0,
      target_type: 0, // TargetNone
      cost_gem: 0,
      cost_crystal: 0,
      cost_discards: 0,
    },
    {
      id: SS_SOUL_CONVERT_SKILL_ID,
      title: '灵魂转换',
      description: '（你每发动1次主动攻击）可转换1点你拥有的[灵魂]颜色。',
      type: 2, // SkillTypeResponse
      min_targets: 0,
      max_targets: 0,
      target_type: 0,
      cost_gem: 0,
      cost_crystal: 0,
      cost_discards: 0,
    },
    {
      id: SS_SOUL_MIRROR_SKILL_ID,
      title: '灵魂镜像',
      description: '（移除2点[黄色灵魂]）你弃2张牌，目标角色摸2张牌[强制]，但最多补到其手牌上限。',
      type: 1,
      min_targets: 1,
      max_targets: 1,
      target_type: 3, // TargetAny
      cost_gem: 0,
      cost_crystal: 0,
      cost_discards: 2,
    },
    {
      id: SS_SOUL_BLAST_SKILL_ID,
      title: '灵魂震爆',
      description: '（移除3点[黄色灵魂]）对目标角色造成3点法术伤害；若其手牌<3且手牌上限>5，本次伤害额外+2。',
      type: 1,
      min_targets: 1,
      max_targets: 1,
      target_type: 3,
      cost_gem: 0,
      cost_crystal: 0,
      cost_discards: 0,
    },
    {
      id: SS_SOUL_GRANT_SKILL_ID,
      title: '灵魂赐予',
      description: '（移除3点[蓝色灵魂]）目标角色+2[宝石]。',
      type: 1,
      min_targets: 1,
      max_targets: 1,
      target_type: 3,
      cost_gem: 0,
      cost_crystal: 0,
      cost_discards: 0,
    },
    {
      id: SS_SOUL_LINK_SKILL_ID,
      title: '灵魂链接',
      description: '（仅你队友数>1时可发动，移除1黄魂+1蓝魂）将灵魂链接放置于目标队友面前；你或其承受伤害前可移除X蓝魂，将X点伤害转移给另一方。',
      type: 3, // SkillTypeStartup
      min_targets: 0,
      max_targets: 0,
      target_type: 0,
      cost_gem: 0,
      cost_crystal: 0,
      cost_discards: 0,
    },
    {
      id: SS_SOUL_AMP_SKILL_ID,
      title: '灵魂增幅',
      description: '[宝石]你+2[黄色灵魂]和+2[蓝色灵魂]。',
      type: 3, // SkillTypeStartup
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
  faction: '幻',
  skills: [],
});

const ally2Character = characterView({
  id: 'ally2_char',
  name: '天使',
  title: '圣',
  faction: '圣',
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
function ssHand(withMagic: boolean = true) {
  const cards = [
    card({ id: 'card_1', name: '火焰斩', type: 'Attack', element: 'Fire' }),
    card({ id: 'card_2', name: '火焰斩', type: 'Attack', element: 'Fire' }),
    card({ id: 'card_3', name: '水涟斩', type: 'Attack', element: 'Water' }),
  ];
  if (withMagic) {
    cards.push(
      card({ id: 'card_4', name: '寒冰箭', type: 'Magic', element: 'Water' }),
      card({ id: 'card_5', name: '圣光', type: 'Magic', element: 'Light' })
    );
  }
  return cards;
}

function ssPlayerView(options: {
  is_active?: boolean;
  blue_soul?: number;
  yellow_soul?: number;
  gems?: number;
  crystals?: number;
  with_magic?: boolean;
  exclusive_card_count?: number;
} = {}) {
  const hand = ssHand(options.with_magic ?? true);
  return playerView({
    id: SS_PLAYER_ID,
    name: 'E2E SoulSorcerer',
    camp: 'Red',
    role: 'soul_sorcerer',
    hand,
    hand_count: hand.length,
    heal: 2,
    max_heal: 4,
    is_active: options.is_active ?? true,
    gem: options.gems ?? 0,
    crystal: options.crystals ?? 0,
    exclusive_card_count: options.exclusive_card_count ?? 1,
    tokens: {
      ss_blue_soul: options.blue_soul ?? 0,
      ss_yellow_soul: options.yellow_soul ?? 0,
    },
  });
}

// ============================================================
// 灵魂召还 (ss_soul_recall) - Discard magic cards for blue souls
// ============================================================

export function soulRecallScenario(options: { with_magic?: boolean } = {}): ProtocolHarnessScenario {
  const characters = [soulSorcererCharacter, allyCharacter, enemyCharacter];
  const withMagic = options.with_magic ?? true;

  const ss = ssPlayerView({ with_magic: withMagic, is_active: true });

  const players = [
    ss,
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
    myPlayerId: SS_PLAYER_ID,
    myPlayerName: 'E2E SoulSorcerer',
    characters,
    players: [
      playerInfo({ id: SS_PLAYER_ID, name: 'E2E SoulSorcerer', camp: 'Red', char_role: 'soul_sorcerer', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy Bot', camp: 'Blue', char_role: 'villain' }),
    ],
    initialState: syncState({
      turn_player_id: SS_PLAYER_ID,
      turn_stage: 'ActionExecution',
      available_skills: [availableSkill({ id: SS_SOUL_RECALL_SKILL_ID, title: '灵魂召还' })],
      characters,
      players,
    }),
  };
}

export function soulRecallPickPrompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_cards',
    player_id: SS_PLAYER_ID,
    message: '【灵魂召还】请选择要弃置的法术牌（至少1张）：',
    choice_type: 'ss_recall_pick',
    skill_id: SS_SOUL_RECALL_SKILL_ID,
    options: [
      { id: 'card_4', label: '4: 寒冰箭 (法术·水)', button_label: '选择', card_id: 'card_4' },
      { id: 'card_5', label: '5: 圣光 (法术·光)', button_label: '选择', card_id: 'card_5' },
    ],
    min: 1,
    max: 2,
    presentation: { kind: 'card_picker', card_source: 'hand', card_filter: 'option_limited', numeric_base: 0 },
  } satisfies Prompt);
}

// ============================================================
// 灵魂转换 (ss_soul_convert) - Convert soul colors on attack
// ============================================================

export function soulConvertScenario(options: {
  blue_soul?: number;
  yellow_soul?: number;
} = {}): ProtocolHarnessScenario {
  const characters = [soulSorcererCharacter, allyCharacter, enemyCharacter];

  const ss = ssPlayerView({
    blue_soul: options.blue_soul ?? 2,
    yellow_soul: options.yellow_soul ?? 2,
    is_active: true,
  });

  const players = [
    ss,
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
    myPlayerId: SS_PLAYER_ID,
    myPlayerName: 'E2E SoulSorcerer',
    characters,
    players: [
      playerInfo({ id: SS_PLAYER_ID, name: 'E2E SoulSorcerer', camp: 'Red', char_role: 'soul_sorcerer', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy Bot', camp: 'Blue', char_role: 'villain' }),
    ],
    initialState: syncState({
      turn_player_id: SS_PLAYER_ID,
      turn_stage: 'AttackDeclaration',
      available_skills: [],
      characters,
      players,
    }),
  };
}

export function soulConvertColorPrompt(options: {
  can_y2b?: boolean;
  can_b2y?: boolean;
} = {}): WsMessage {
  const canY2B = options.can_y2b ?? true;
  const canB2Y = options.can_b2y ?? true;
  const optionsList: { id: string; label: string; button_label: string }[] = [];

  if (canY2B) {
    optionsList.push({ id: 'yellow_to_blue', label: '黄色灵魂转蓝色灵魂', button_label: '黄转蓝' });
  }
  if (canB2Y) {
    optionsList.push({ id: 'blue_to_yellow', label: '蓝色灵魂转黄色灵魂', button_label: '蓝转黄' });
  }
  optionsList.push({ id: 'cancel', label: '取消', button_label: '取消' });

  return requireActionMessage({
    type: 'confirm',
    player_id: SS_PLAYER_ID,
    message: '【灵魂转换】请选择转换方向：',
    choice_type: 'ss_convert_color',
    skill_id: SS_SOUL_CONVERT_SKILL_ID,
    options: optionsList,
    min: 1,
    max: 1,
    presentation: { kind: 'branch_select', layout: 'overlay', numeric_base: 0 },
  } satisfies Prompt);
}

// ============================================================
// 灵魂镜像 (ss_soul_mirror) - Remove 2 yellow souls, target draws 2
// ============================================================

export function soulMirrorScenario(options: { yellow_soul?: number } = {}): ProtocolHarnessScenario {
  const yellowSoul = options.yellow_soul ?? 2;
  const characters = [soulSorcererCharacter, allyCharacter, enemyCharacter];

  const ss = ssPlayerView({
    yellow_soul: yellowSoul,
    is_active: true,
  });

  const ally = playerView({
    id: ALLY_PLAYER_ID,
    name: 'Ally Bot',
    camp: 'Red',
    role: 'hero',
    hand: [],
    hand_count: 2,
    heal: 2,
    max_heal: 4,
    max_hand: 6,
    is_active: false,
  });

  const players = [ss, ally];

  return {
    roomCode: 'MOCK',
    myPlayerId: SS_PLAYER_ID,
    myPlayerName: 'E2E SoulSorcerer',
    characters,
    players: [
      playerInfo({ id: SS_PLAYER_ID, name: 'E2E SoulSorcerer', camp: 'Red', char_role: 'soul_sorcerer', is_host: true }),
      playerInfo({ id: ALLY_PLAYER_ID, name: 'Ally Bot', camp: 'Red', char_role: 'hero' }),
    ],
    initialState: syncState({
      turn_player_id: SS_PLAYER_ID,
      turn_stage: 'ActionExecution',
      available_skills: [availableSkill({ id: SS_SOUL_MIRROR_SKILL_ID, title: '灵魂镜像' })],
      characters,
      players,
    }),
  };
}

export function soulMirrorTargetPrompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_target',
    player_id: SS_PLAYER_ID,
    message: '【灵魂镜像】请选择目标角色：',
    choice_type: 'skill_target',
    skill_id: SS_SOUL_MIRROR_SKILL_ID,
    presentation: { kind: 'target_picker', target_filter: 'all', numeric_base: 0 },
    options: [
      { id: SS_PLAYER_ID, target_id: SS_PLAYER_ID, label: '灵魂术士', button_label: '选择' },
      { id: ALLY_PLAYER_ID, target_id: ALLY_PLAYER_ID, label: '圣女', button_label: '选择' },
    ],
    min: 1,
    max: 1,
  } satisfies Prompt);
}

// ============================================================
// 灵魂震爆 (ss_soul_blast) - Remove 3 yellow souls, deal magic damage
// ============================================================

export function soulBlastScenario(options: { yellow_soul?: number } = {}): ProtocolHarnessScenario {
  const yellowSoul = options.yellow_soul ?? 3;
  const characters = [soulSorcererCharacter, allyCharacter, enemyCharacter];

  const ss = ssPlayerView({
    yellow_soul: yellowSoul,
    is_active: true,
    exclusive_card_count: 1,
  });

  const enemy = playerView({
    id: ENEMY_PLAYER_ID,
    name: 'Enemy Bot',
    camp: 'Blue',
    role: 'villain',
    hand: [],
    hand_count: 2,
    heal: 1,
    max_heal: 4,
    max_hand: 6,
    is_active: false,
  });

  const players = [ss, enemy];

  return {
    roomCode: 'MOCK',
    myPlayerId: SS_PLAYER_ID,
    myPlayerName: 'E2E SoulSorcerer',
    characters,
    players: [
      playerInfo({ id: SS_PLAYER_ID, name: 'E2E SoulSorcerer', camp: 'Red', char_role: 'soul_sorcerer', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy Bot', camp: 'Blue', char_role: 'villain' }),
    ],
    initialState: syncState({
      turn_player_id: SS_PLAYER_ID,
      turn_stage: 'ActionExecution',
      available_skills: [availableSkill({ id: SS_SOUL_BLAST_SKILL_ID, title: '灵魂震爆' })],
      characters,
      players,
    }),
  };
}

export function soulBlastTargetPrompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_target',
    player_id: SS_PLAYER_ID,
    message: '【灵魂震爆】请选择目标角色：',
    choice_type: 'skill_target',
    skill_id: SS_SOUL_BLAST_SKILL_ID,
    presentation: { kind: 'target_picker', target_filter: 'all', numeric_base: 0 },
    options: [
      { id: SS_PLAYER_ID, target_id: SS_PLAYER_ID, label: '灵魂术士', button_label: '选择' },
      { id: ENEMY_PLAYER_ID, target_id: ENEMY_PLAYER_ID, label: '魔神', button_label: '选择' },
    ],
    min: 1,
    max: 1,
  } satisfies Prompt);
}

// ============================================================
// 灵魂赐予 (ss_soul_grant) - Remove 3 blue souls, target gains gems
// ============================================================

export function soulGrantScenario(options: { blue_soul?: number } = {}): ProtocolHarnessScenario {
  const blueSoul = options.blue_soul ?? 3;
  const characters = [soulSorcererCharacter, allyCharacter, enemyCharacter];

  const ss = ssPlayerView({
    blue_soul: blueSoul,
    is_active: true,
    exclusive_card_count: 1,
  });

  const ally = playerView({
    id: ALLY_PLAYER_ID,
    name: 'Ally Bot',
    camp: 'Red',
    role: 'hero',
    hand: [],
    hand_count: 2,
    heal: 2,
    max_heal: 4,
    gem: 0,
    crystal: 0,
    is_active: false,
  });

  const players = [ss, ally];

  return {
    roomCode: 'MOCK',
    myPlayerId: SS_PLAYER_ID,
    myPlayerName: 'E2E SoulSorcerer',
    characters,
    players: [
      playerInfo({ id: SS_PLAYER_ID, name: 'E2E SoulSorcerer', camp: 'Red', char_role: 'soul_sorcerer', is_host: true }),
      playerInfo({ id: ALLY_PLAYER_ID, name: 'Ally Bot', camp: 'Red', char_role: 'hero' }),
    ],
    initialState: syncState({
      turn_player_id: SS_PLAYER_ID,
      turn_stage: 'ActionExecution',
      available_skills: [availableSkill({ id: SS_SOUL_GRANT_SKILL_ID, title: '灵魂赐予' })],
      characters,
      players,
    }),
  };
}

export function soulGrantTargetPrompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_target',
    player_id: SS_PLAYER_ID,
    message: '【灵魂赐予】请选择目标角色：',
    choice_type: 'skill_target',
    skill_id: SS_SOUL_GRANT_SKILL_ID,
    presentation: { kind: 'target_picker', target_filter: 'all', numeric_base: 0 },
    options: [
      { id: SS_PLAYER_ID, target_id: SS_PLAYER_ID, label: '灵魂术士', button_label: '选择' },
      { id: ALLY_PLAYER_ID, target_id: ALLY_PLAYER_ID, label: '圣女', button_label: '选择' },
    ],
    min: 1,
    max: 1,
  } satisfies Prompt);
}

// ============================================================
// 灵魂链接 (ss_soul_link) - Place soul link on ally
// ============================================================

export function soulLinkScenario(options: {
  blue_soul?: number;
  yellow_soul?: number;
} = {}): ProtocolHarnessScenario {
  const characters = [soulSorcererCharacter, allyCharacter, ally2Character, enemyCharacter];

  const ss = ssPlayerView({
    blue_soul: options.blue_soul ?? 1,
    yellow_soul: options.yellow_soul ?? 1,
    is_active: true,
    exclusive_card_count: 1,
  });

  const ally1 = playerView({
    id: ALLY_PLAYER_ID,
    name: 'Ally Bot 1',
    camp: 'Red',
    role: 'hero',
    hand: [],
    hand_count: 2,
    heal: 2,
    max_heal: 4,
    is_active: false,
  });

  const ally2 = playerView({
    id: ALLY_2_PLAYER_ID,
    name: 'Ally Bot 2',
    camp: 'Red',
    role: 'angel',
    hand: [],
    hand_count: 3,
    heal: 2,
    max_heal: 4,
    is_active: false,
  });

  const players = [ss, ally1, ally2];

  return {
    roomCode: 'MOCK',
    myPlayerId: SS_PLAYER_ID,
    myPlayerName: 'E2E SoulSorcerer',
    characters,
    players: [
      playerInfo({ id: SS_PLAYER_ID, name: 'E2E SoulSorcerer', camp: 'Red', char_role: 'soul_sorcerer', is_host: true }),
      playerInfo({ id: ALLY_PLAYER_ID, name: 'Ally Bot 1', camp: 'Red', char_role: 'hero' }),
      playerInfo({ id: ALLY_2_PLAYER_ID, name: 'Ally Bot 2', camp: 'Red', char_role: 'angel' }),
    ],
    initialState: syncState({
      turn_player_id: SS_PLAYER_ID,
      turn_stage: 'TurnStart',
      available_skills: [availableSkill({ id: SS_SOUL_LINK_SKILL_ID, title: '灵魂链接' })],
      characters,
      players,
    }),
  };
}

export function soulLinkTargetPrompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_target',
    player_id: SS_PLAYER_ID,
    message: '【灵魂链接】请选择要放置灵魂链接的队友：',
    choice_type: 'ss_link_target',
    skill_id: SS_SOUL_LINK_SKILL_ID,
    presentation: { kind: 'target_picker', target_filter: 'allies', numeric_base: 0 },
    options: [
      { id: ALLY_PLAYER_ID, target_id: ALLY_PLAYER_ID, label: 'Ally Bot 1', button_label: '选择' },
      { id: ALLY_2_PLAYER_ID, target_id: ALLY_2_PLAYER_ID, label: 'Ally Bot 2', button_label: '选择' },
    ],
    min: 1,
    max: 1,
  } satisfies Prompt);
}

// ============================================================
// 灵魂增幅 (ss_soul_amp) - Consume gem, gain souls
// ============================================================

export function soulAmpScenario(options: { gems?: number } = {}): ProtocolHarnessScenario {
  const gems = options.gems ?? 1;
  const characters = [soulSorcererCharacter, allyCharacter, enemyCharacter];

  const ss = ssPlayerView({
    gems,
    blue_soul: 0,
    yellow_soul: 0,
    is_active: true,
  });

  const players = [ss];

  return {
    roomCode: 'MOCK',
    myPlayerId: SS_PLAYER_ID,
    myPlayerName: 'E2E SoulSorcerer',
    characters,
    players: [
      playerInfo({ id: SS_PLAYER_ID, name: 'E2E SoulSorcerer', camp: 'Red', char_role: 'soul_sorcerer', is_host: true }),
    ],
    initialState: syncState({
      turn_player_id: SS_PLAYER_ID,
      turn_stage: 'TurnStart',
      available_skills: [availableSkill({ id: SS_SOUL_AMP_SKILL_ID, title: '灵魂增幅' })],
      characters,
      players,
    }),
  };
}

export function soulAmpConfirmPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: SS_PLAYER_ID,
    message: '【灵魂增幅】是否发动？消耗1个宝石，黄色灵魂+2，蓝色灵魂+2',
    choice_type: 'ss_soul_amp_confirm',
    skill_id: SS_SOUL_AMP_SKILL_ID,
    options: [
      { id: '0', label: '发动', button_label: '发动' },
      { id: '1', label: '不发动', button_label: '不发动' },
    ],
    presentation: { kind: 'branch_select', layout: 'overlay', numeric_base: 0 },
    min: 1,
    max: 1,
  } satisfies Prompt);
}
