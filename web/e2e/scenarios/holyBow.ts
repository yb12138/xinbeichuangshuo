// ============================================================
// Holy Bow Archer (圣弓) Protocol Harness Scenarios
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
export const HB_PLAYER_ID = 'hb_player';
export const ALLY_PLAYER_ID = 'ally_1';
export const ENEMY_PLAYER_ID = 'enemy_1';
export const ENEMY_2_PLAYER_ID = 'enemy_2';

// ---- Skill IDs (与 internal/data/characters.go holy_bow 段一致) ----
export const HB_SHARD_STORM_SKILL_ID = 'hb_holy_shard_storm';
export const HB_RADIANT_DESCENT_SKILL_ID = 'hb_radiant_descent';
export const HB_LIGHT_BURST_SKILL_ID = 'hb_light_burst';
export const HB_STAR_BULLET_SKILL_ID = 'hb_meteor_bullet';
export const HB_RADIANT_CANNON_SKILL_ID = 'hb_radiant_cannon';
export const HB_AUTO_FILL_SKILL_ID = 'hb_auto_fill';

// ---- Holy Bow character definition ----
const holyBowCharacter = characterView({
  id: 'holy_bow',
  name: '圣弓',
  title: '光之射手',
  faction: '星杯',
  skills: [
    {
      id: HB_SHARD_STORM_SKILL_ID,
      title: '圣屑飓暴',
      description: '弃2张同系攻击牌发动攻击，未命中时移除治疗令队友弃牌',
      type: 2,
      min_targets: 0,
      max_targets: 0,
      target_type: 0,
      cost_gem: 0,
      cost_crystal: 0,
      cost_discards: 2,
    },
    {
      id: HB_RADIANT_DESCENT_SKILL_ID,
      title: '圣煌降临',
      description: '法术技能，直接发动',
      type: 2,
      min_targets: 0,
      max_targets: 0,
      target_type: 0,
      cost_gem: 0,
      cost_crystal: 0,
      cost_discards: 0,
    },
    {
      id: HB_LIGHT_BURST_SKILL_ID,
      title: '圣光爆裂',
      description: '法术技能，分支选择后结算',
      type: 2,
      min_targets: 0,
      max_targets: 0,
      target_type: 0,
      cost_gem: 0,
      cost_crystal: 0,
      cost_discards: 0,
    },
    {
      id: HB_STAR_BULLET_SKILL_ID,
      title: '流星圣弹',
      description: '响应技能，攻击时发动',
      type: 0,
      min_targets: 0,
      max_targets: 0,
      target_type: 0,
      cost_gem: 0,
      cost_crystal: 0,
      cost_discards: 0,
    },
    {
      id: HB_RADIANT_CANNON_SKILL_ID,
      title: '圣煌辉光炮',
      description: '法术技能，发动后选择士气',
      type: 2,
      min_targets: 0,
      max_targets: 0,
      target_type: 0,
      cost_gem: 0,
      cost_crystal: 0,
      cost_discards: 0,
    },
    {
      id: HB_AUTO_FILL_SKILL_ID,
      title: '自动填充',
      description: '回合结束时，消耗水晶或宝石获得信仰或治疗',
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

// ---- Helper functions ----
function holyBowHand() {
  return [
    card({ id: 'card_1', name: '火焰斩', type: 'Attack', element: 'Fire' }),
    card({ id: 'card_2', name: '火焰斩', type: 'Attack', element: 'Fire' }),
    card({ id: 'card_3', name: '水涟斩', type: 'Attack', element: 'Water' }),
    card({ id: 'card_4', name: '寒冰箭', type: 'Magic', element: 'Water' }),
    card({ id: 'card_5', name: '圣光', type: 'Magic', element: 'Light' }),
  ];
}

function holyBowPlayerView(options: {
  is_active?: boolean;
  heal?: number;
  max_heal?: number;
  faith?: number;
  gems?: number;
  crystals?: number;
  cannon?: boolean;
} = {}) {
  return playerView({
    id: HB_PLAYER_ID,
    name: 'E2E Holy Bow',
    camp: 'Red',
    role: 'holy_bow',
    hand: holyBowHand(),
    hand_count: holyBowHand().length,
    heal: options.heal ?? 2,
    max_heal: options.max_heal ?? 4,
    is_active: options.is_active ?? true,
    gem: options.gems ?? 0,
    crystal: options.crystals ?? 0,
    tokens: {
      hb_faith: options.faith ?? 0,
    },
    exclusive_cards: options.cannon ? [
      card({ id: 'hb_cannon', name: '圣煌辉光炮', type: 'Magic', element: 'Light' }),
    ] : [],
  });
}

// ============================================================
// 圣屑飓暴 (hb_shard_storm) - Skill with discard then miss follow-up
// ============================================================

export function shardStormScenario(): ProtocolHarnessScenario {
  const characters = [holyBowCharacter, allyCharacter, enemyCharacter];

  const holyBow = holyBowPlayerView({ is_active: true });

  const players = [
    holyBow,
    playerView({
      id: ALLY_PLAYER_ID,
      name: 'Ally Hero',
      camp: 'Red',
      role: 'hero',
      hand: [card({ id: 'ally_card_1', name: '测试牌' })],
      hand_count: 1,
      heal: 2,
      max_heal: 3,
      is_active: false,
    }),
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
    myPlayerId: HB_PLAYER_ID,
    myPlayerName: 'E2E Holy Bow',
    characters,
    players: [
      playerInfo({ id: HB_PLAYER_ID, name: 'E2E Holy Bow', camp: 'Red', char_role: 'holy_bow', is_host: true }),
      playerInfo({ id: ALLY_PLAYER_ID, name: 'Ally Hero', camp: 'Red', char_role: 'hero' }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy Bot', camp: 'Blue', char_role: 'villain' }),
    ],
    initialState: syncState({
      turn_player_id: HB_PLAYER_ID,
      turn_stage: 'ActionExecution',
      available_skills: [availableSkill({ id: HB_SHARD_STORM_SKILL_ID, title: '圣屑飓暴', cost_discards: 2 })],
      characters,
      players,
    }),
  };
}

// 后端 buildHolyShardComboPrompt 为单选 confirm，option id 是 "Element:i,j" 组合字符串。
export function shardStormDiscardPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: HB_PLAYER_ID,
    message: '【圣屑飓暴】请选择要弃置的2张同系攻击牌：',
    choice_type: 'hb_holy_shard_combo',
    skill_id: HB_SHARD_STORM_SKILL_ID,
    options: [
      { id: 'fire:0,1', label: '火系：1:火焰斩 + 2:火焰斩' },
    ],
    min: 1,
    max: 1,
  } satisfies Prompt);
}

// 后端 buildHolyShardMissConfirmPrompt：未命中后先弹「是否走未命中分支」的 confirm。
// 选「否」直接结束未命中流程，选「是」才进入 miss_x 选择。
export function shardStormMissConfirmPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: HB_PLAYER_ID,
    message: '【圣屑飓暴】未命中：是否移除治疗并令1名队友弃牌？',
    choice_type: 'hb_holy_shard_miss_confirm',
    skill_id: HB_SHARD_STORM_SKILL_ID,
    options: [
      { id: '0', label: '是' },
      { id: '1', label: '否' },
    ],
    min: 1,
    max: 1,
  } satisfies Prompt);
}

// 后端 buildHolyShardMissXPrompt：X 取值范围为 1..maxX（无 X=0 选项）。
export function shardStormMissHealPrompt(maxX = 2): WsMessage {
  const options: { id: string; label: string }[] = [];
  for (let i = 1; i <= maxX; i++) {
    options.push({ id: `${i}`, label: `移除${i}点治疗，并令队友弃${i}张牌` });
  }
  return requireActionMessage({
    type: 'confirm',
    player_id: HB_PLAYER_ID,
    message: '【圣屑飓暴】请选择移除治疗点数X：',
    choice_type: 'hb_holy_shard_miss_x',
    skill_id: HB_SHARD_STORM_SKILL_ID,
    options,
    presentation: { kind: 'numeric', numeric_base: 0 },
    min: 1,
    max: 1,
  } satisfies Prompt);
}

export function shardStormMissTargetPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: HB_PLAYER_ID,
    message: '【圣屑飓暴】请选择一名队友令其弃牌：',
    choice_type: 'hb_holy_shard_miss_ally_target',
    skill_id: HB_SHARD_STORM_SKILL_ID,
    options: [
      { id: ALLY_PLAYER_ID, label: '勇者' },
    ],
    min: 1,
    max: 1,
  } satisfies Prompt);
}

// ============================================================
// 圣煌降临 (hb_radiant_descent) - Simple magic skill
// ============================================================

// 支持不同 heal/faith 配置的场景
export function radiantDescentScenario(options: { heal?: number; faith?: number } = {}): ProtocolHarnessScenario {
  const heal = options.heal ?? 2;
  const faith = options.faith ?? 0;
  const characters = [holyBowCharacter, allyCharacter, enemyCharacter];

  const holyBow = holyBowPlayerView({ heal, is_active: true });
  // 修改 tokens 以反映 faith 配置
  holyBow.tokens = { hb_faith: faith };

  const players = [
    holyBow,
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
    myPlayerId: HB_PLAYER_ID,
    myPlayerName: 'E2E Holy Bow',
    characters,
    players: [
      playerInfo({ id: HB_PLAYER_ID, name: 'E2E Holy Bow', camp: 'Red', char_role: 'holy_bow', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy Bot', camp: 'Blue', char_role: 'villain' }),
    ],
    initialState: syncState({
      turn_player_id: HB_PLAYER_ID,
      turn_stage: 'ActionExecution',
      available_skills: [availableSkill({ id: HB_RADIANT_DESCENT_SKILL_ID, title: '圣煌降临' })],
      characters,
      players,
    }),
  };
}

// 后端 buildRadiantDescentCostPrompt：根据 cost_modes 动态生成选项
// choice_type: 'hb_radiant_descent_cost', Message: '【圣煌降临】请选择支付方式：'
export function radiantDescentCostPrompt(costModes: ('heal' | 'faith')[]): WsMessage {
  const options: { id: string; label: string }[] = [];
  for (const mode of costModes) {
    if (mode === 'heal') {
      options.push({ id: `${options.length}`, label: '移除2点治疗' });
    } else if (mode === 'faith') {
      options.push({ id: `${options.length}`, label: '移除2点信仰' });
    }
  }
  return requireActionMessage({
    type: 'confirm',
    player_id: HB_PLAYER_ID,
    message: '【圣煌降临】请选择支付方式：',
    choice_type: 'hb_radiant_descent_cost',
    skill_id: HB_RADIANT_DESCENT_SKILL_ID,
    options,
    min: 1,
    max: 1,
  } satisfies Prompt);
}

// ============================================================
// 圣光爆裂 (hb_light_burst) - Branch skill with complex flow
// ============================================================

export function lightBurstScenario(options: { heal?: number } = {}): ProtocolHarnessScenario {
  const heal = options.heal ?? 2;
  const characters = [holyBowCharacter, allyCharacter, enemyCharacter];

  const holyBow = holyBowPlayerView({ heal, is_active: true });

  const players = [
    holyBow,
    playerView({
      id: ALLY_PLAYER_ID,
      name: 'Ally Hero',
      camp: 'Red',
      role: 'hero',
      hand: [card({ id: 'ally_card_1', name: '测试牌' })],
      hand_count: 1,
      heal: 2,
      max_heal: 3,
      is_active: false,
    }),
    playerView({
      id: ENEMY_PLAYER_ID,
      name: 'Enemy Bot',
      camp: 'Blue',
      role: 'villain',
      hand: [],
      hand_count: 2,
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
      hand_count: 1,
      heal: 0,
      max_heal: 2,
      is_active: false,
    }),
  ];

  return {
    roomCode: 'MOCK',
    myPlayerId: HB_PLAYER_ID,
    myPlayerName: 'E2E Holy Bow',
    characters,
    players: [
      playerInfo({ id: HB_PLAYER_ID, name: 'E2E Holy Bow', camp: 'Red', char_role: 'holy_bow', is_host: true }),
      playerInfo({ id: ALLY_PLAYER_ID, name: 'Ally Hero', camp: 'Red', char_role: 'hero' }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy Bot', camp: 'Blue', char_role: 'villain' }),
      playerInfo({ id: ENEMY_2_PLAYER_ID, name: 'Enemy Bot 2', camp: 'Blue', char_role: 'villain' }),
    ],
    initialState: syncState({
      turn_player_id: HB_PLAYER_ID,
      turn_stage: 'ActionExecution',
      available_skills: [availableSkill({ id: HB_LIGHT_BURST_SKILL_ID, title: '圣光爆裂' })],
      characters,
      players,
    }),
  };
}

export function lightBurstBranchPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: HB_PLAYER_ID,
    message: '【圣光爆裂】请选择分支：',
    choice_type: 'hb_light_burst_mode',
    skill_id: HB_LIGHT_BURST_SKILL_ID,
    options: [
      { id: '0', label: '分支一：摸牌+移除治疗+增信仰，队友+治疗' },
      { id: '1', label: '分支二：移除X治疗，选择最多X名对手，弃X牌造成伤害' },
    ],
    presentation: { kind: 'branch_select', layout: 'overlay' },
    min: 1,
    max: 1,
  } satisfies Prompt);
}

export function lightBurstBranch1TargetPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: HB_PLAYER_ID,
    message: '【圣光爆裂】分支一：请选择一名队友令其+1治疗：',
    choice_type: 'hb_light_burst_mode_a_target',
    skill_id: HB_LIGHT_BURST_SKILL_ID,
    options: [
      { id: ALLY_PLAYER_ID, label: '勇者' },
    ],
    min: 1,
    max: 1,
  } satisfies Prompt);
}

// 后端 buildLightBurstModeBXPrompt 生成的选项格式：X={x}（移除{x}治疗并弃{x}张牌）
export function lightBurstBranch2HealPrompt(maxX = 2): WsMessage {
  const options: { id: string; label: string }[] = [];
  for (let x = 1; x <= maxX; x++) {
    options.push({ id: `${x}`, label: `X=${x}（移除${x}治疗并弃${x}张牌）` });
  }
  return requireActionMessage({
    type: 'confirm',
    player_id: HB_PLAYER_ID,
    message: '【圣光爆裂】分支②请选择X值：',
    choice_type: 'hb_light_burst_mode_b_x',
    skill_id: HB_LIGHT_BURST_SKILL_ID,
    options,
    presentation: { kind: 'numeric', numeric_base: 0 },
    min: 1,
    max: 1,
  } satisfies Prompt);
}

// 后端 buildLightBurstModeBTargetsPrompt 为「每次单选 + finish 按钮」迭代选择：
// Min=Max=1，玩家可分多次提交，最多 X 名目标。已选目标会从下次 options 移除，
// 第二次及之后会附加 "finish" 选项让玩家结束选择。
// 注意：前端对于有玩家选项和非玩家选项的 confirm prompt，会显示 overlay（只含非玩家选项）阻挡 player-area 点击。
// 所以第二次选择时，如果需要继续选择玩家目标，应该不包含 finish 选项，让 overlay 不显示。
export function lightBurstBranch2TargetPrompt(args: {
  xValue: number;
  selectedCount?: number;
  withFinish?: boolean;
  selectedIds?: string[];
} = { xValue: 2 }): WsMessage {
  const xValue = args.xValue;
  const selectedCount = args.selectedCount ?? 0;
  const selectedIds = args.selectedIds ?? [];
  const allCandidates: { id: string; label: string }[] = [
    { id: ENEMY_PLAYER_ID, label: '恶徒' },
    { id: ENEMY_2_PLAYER_ID, label: '恶徒2' },
  ];
  // 过滤掉已选中的目标
  const options = allCandidates.filter(c => !selectedIds.includes(c.id));
  if (args.withFinish ?? selectedCount > 0) {
    options.push({ id: 'finish', label: '完成目标选择' });
  }
  return requireActionMessage({
    type: 'confirm',
    player_id: HB_PLAYER_ID,
    message: `【圣光爆裂】分支②请点击角色立绘选择目标（已选${selectedCount}/最多${xValue}）：`,
    choice_type: 'hb_light_burst_mode_b_targets',
    skill_id: HB_LIGHT_BURST_SKILL_ID,
    options,
    min: 1,
    max: 1,
  } satisfies Prompt);
}

export function lightBurstBranch2DiscardPrompt(xValue: number): WsMessage {
  return requireActionMessage({
    type: 'choose_cards',
    player_id: HB_PLAYER_ID,
    message: `【圣光爆裂】分支二：请弃置${xValue}张牌：`,
    choice_type: 'hb_light_burst_mode_b_discard',
    skill_id: HB_LIGHT_BURST_SKILL_ID,
    options: [
      { id: '0', label: '1: 火焰斩 (火 Attack)' },
      { id: '1', label: '2: 火焰斩 (火 Attack)' },
      { id: '2', label: '3: 水涟斩 (水 Attack)' },
    ],
    min: xValue,
    max: xValue,
  } satisfies Prompt);
}

// ============================================================
// 流星圣弹 (hb_star_bullet) - Attack response skill
// ============================================================

export function starBulletScenario(): ProtocolHarnessScenario {
  const characters = [holyBowCharacter, allyCharacter, enemyCharacter];

  const holyBow = holyBowPlayerView({ is_active: true });

  const players = [
    holyBow,
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
    myPlayerId: HB_PLAYER_ID,
    myPlayerName: 'E2E Holy Bow',
    characters,
    players: [
      playerInfo({ id: HB_PLAYER_ID, name: 'E2E Holy Bow', camp: 'Red', char_role: 'holy_bow', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy Bot', camp: 'Blue', char_role: 'villain' }),
    ],
    initialState: syncState({
      turn_player_id: HB_PLAYER_ID,
      turn_stage: 'ActionExecution',
      available_skills: [],
      characters,
      players,
    }),
  };
}

// 后端流星圣弹真实流程：
// 1) 响应技能 PromptChooseSkill 入口（由通用 buildResponseSkillPrompt 处理，无独立 choice_type）
// 2) hb_meteor_bullet_cost：选择消耗 1 点治疗或 1 点信仰
// 3) hb_meteor_bullet_target：选择获得治疗的我方队友
export function starBulletResponsePrompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_skill',
    player_id: HB_PLAYER_ID,
    message: '你触发了响应技能【流星圣弹】，请选择是否发动。',
    options: [
      { id: HB_STAR_BULLET_SKILL_ID, label: '流星圣弹', hint: '发动【流星圣弹】' },
      { id: 'skip', label: '跳过', hint: '不发动响应技能' },
    ],
    presentation: { kind: 'skill_choice', layout: 'overlay' },
    min: 1,
    max: 1,
  } satisfies Prompt);
}

export function starBulletCostPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: HB_PLAYER_ID,
    message: '【流星圣弹】请选择要移除的资源：',
    choice_type: 'hb_meteor_bullet_cost',
    skill_id: HB_STAR_BULLET_SKILL_ID,
    options: [
      { id: '0', label: '移除1点治疗' },
      { id: '1', label: '移除1点信仰' },
    ],
    min: 1,
    max: 1,
  } satisfies Prompt);
}

export function starBulletTargetPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: HB_PLAYER_ID,
    message: '【流星圣弹】请选择获得治疗的我方角色：',
    choice_type: 'hb_meteor_bullet_target',
    skill_id: HB_STAR_BULLET_SKILL_ID,
    options: [
      { id: ALLY_PLAYER_ID, label: '勇者' },
    ],
    min: 1,
    max: 1,
  } satisfies Prompt);
}

// ============================================================
// 圣煌辉光炮 (hb_radiant_cannon) - Skill with morale choice
// ============================================================

export function radiantCannonScenario(): ProtocolHarnessScenario {
  const characters = [holyBowCharacter, allyCharacter, enemyCharacter];

  const holyBow = holyBowPlayerView({ cannon: true, is_active: true });

  const players = [
    holyBow,
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
    myPlayerId: HB_PLAYER_ID,
    myPlayerName: 'E2E Holy Bow',
    characters,
    players: [
      playerInfo({ id: HB_PLAYER_ID, name: 'E2E Holy Bow', camp: 'Red', char_role: 'holy_bow', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy Bot', camp: 'Blue', char_role: 'villain' }),
    ],
    initialState: syncState({
      turn_player_id: HB_PLAYER_ID,
      turn_stage: 'ActionExecution',
      available_skills: [availableSkill({ id: HB_RADIANT_CANNON_SKILL_ID, title: '圣煌辉光炮' })],
      characters,
      players,
    }),
  };
}

export function radiantCannonMoralePrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: HB_PLAYER_ID,
    message: '【圣煌辉光炮】请选择将两方士气调整为：',
    choice_type: 'hb_radiant_cannon_side',
    skill_id: HB_RADIANT_CANNON_SKILL_ID,
    options: [
      { id: 'red', label: '红方士气' },
      { id: 'blue', label: '蓝方士气' },
    ],
    presentation: { kind: 'branch_select', layout: 'overlay' },
    min: 1,
    max: 1,
  } satisfies Prompt);
}

// ============================================================
// 自动填充 (hb_auto_fill) - End of turn passive with branch choice
// ============================================================

export function autoFillScenario(options: { gems?: number; crystals?: number } = {}): ProtocolHarnessScenario {
  const gems = options.gems ?? 1;
  const crystals = options.crystals ?? 1;
  const characters = [holyBowCharacter, allyCharacter, enemyCharacter];

  const holyBow = holyBowPlayerView({ gems, crystals, is_active: false });

  const players = [
    holyBow,
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
    myPlayerId: HB_PLAYER_ID,
    myPlayerName: 'E2E Holy Bow',
    characters,
    players: [
      playerInfo({ id: HB_PLAYER_ID, name: 'E2E Holy Bow', camp: 'Red', char_role: 'holy_bow', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy Bot', camp: 'Blue', char_role: 'villain' }),
    ],
    initialState: syncState({
      turn_player_id: ENEMY_PLAYER_ID,
      turn_stage: 'EndPhase',
      available_skills: [],
      characters,
      players,
    }),
  };
}

export function autoFillBranchPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: HB_PLAYER_ID,
    message: '【自动填充】回合结束，请选择分支：',
    choice_type: 'hb_auto_fill_resource',
    skill_id: HB_AUTO_FILL_SKILL_ID,
    options: [
      { id: 'crystal', label: '消耗水晶，选择增加信仰或治疗' },
      { id: 'gem', label: '消耗红宝石获得蓝水晶，选择增加信仰或治疗' },
    ],
    presentation: { kind: 'branch_select', layout: 'overlay' },
    min: 1,
    max: 1,
  } satisfies Prompt);
}

export function autoFillRewardPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: HB_PLAYER_ID,
    message: '【自动填充】请选择获得：',
    choice_type: 'hb_auto_fill_gain',
    skill_id: HB_AUTO_FILL_SKILL_ID,
    options: [
      { id: 'faith', label: '+1信仰' },
      { id: 'heal', label: '+1治疗' },
    ],
    presentation: { kind: 'branch_select', layout: 'overlay' },
    min: 1,
    max: 1,
  } satisfies Prompt);
}