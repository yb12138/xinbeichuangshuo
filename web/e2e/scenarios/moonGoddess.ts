// ============================================================
// Moon Goddess (月之女神) Protocol Harness Scenarios
// ============================================================

import type { Prompt } from '../../src/types/game';
import type { AvailableSkill, FieldCard } from '../../src/types/game';
import type { WsMessage } from '../../src/network/protocol';
import {
  availableSkill,
  card,
  characterView,
  playerInfo,
  playerView,
  fieldOptionIndexInteraction,
  requireActionMessage,
  syncState,
  type ProtocolHarnessScenario,
} from './builders';

// ---- Player IDs ----
export const MG_PLAYER_ID = 'mg_player';
export const ENEMY_PLAYER_ID = 'enemy_1';
export const ENEMY_2_PLAYER_ID = 'enemy_2';

// ---- Skill IDs ----
export const MG_NEW_MOON_SHelter_SKILL_ID = 'mg_new_moon_shelter';
export const MG_MEDUSA_EYE_SKILL_ID = 'mg_medusa_eye';
export const MG_MOON_CYCLE_SKILL_ID = 'mg_moon_cycle';
export const MG_MOON_READ_SKILL_ID = 'mg_blasphemy';
export const MG_DARKMOON_SLASH_SKILL_ID = 'mg_darkmoon_slash';
export const MG_PALE_MOON_SKILL_ID = 'mg_pale_moon';

// ---- Moon Goddess character definition ----
const moonGoddessCharacter = characterView({
  id: 'moon_goddess',
  name: '月之女神',
  title: '月光守护者',
  faction: '星杯',
  skills: [
    {
      id: MG_NEW_MOON_SHelter_SKILL_ID,
      title: '新月庇护',
      description: '响应技能，触发条件时弹出确认',
      type: 0,
      min_targets: 0,
      max_targets: 0,
      target_type: 0,
      cost_gem: 0,
      cost_crystal: 0,
      cost_discards: 0,
    },
    {
      id: MG_MEDUSA_EYE_SKILL_ID,
      title: '美杜莎之眼',
      description: '响应技能，闇月选择→弃牌→目标选择',
      type: 0,
      min_targets: 0,
      max_targets: 0,
      target_type: 0,
      cost_gem: 0,
      cost_crystal: 0,
      cost_discards: 0,
    },
    {
      id: MG_MOON_CYCLE_SKILL_ID,
      title: '月之轮回',
      description: '响应技能，回合结束时触发分支选择',
      type: 0,
      min_targets: 0,
      max_targets: 0,
      target_type: 0,
      cost_gem: 0,
      cost_crystal: 0,
      cost_discards: 0,
    },
    {
      id: MG_MOON_READ_SKILL_ID,
      title: '月渎',
      description: '响应技能，法术伤害后摸牌触发',
      type: 0,
      min_targets: 0,
      max_targets: 0,
      target_type: 0,
      cost_gem: 0,
      cost_crystal: 0,
      cost_discards: 0,
    },
    {
      id: MG_DARKMOON_SLASH_SKILL_ID,
      title: '闇月斩',
      description: '响应技能，确认→选择X值',
      type: 0,
      min_targets: 0,
      max_targets: 0,
      target_type: 0,
      cost_gem: 0,
      cost_crystal: 0,
      cost_discards: 0,
    },
    {
      id: MG_PALE_MOON_SKILL_ID,
      title: '苍白之月',
      description: '响应技能，确认→分支选择',
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
function moonGoddessHand() {
  return [
    card({ id: 'card_1', name: '火焰斩', type: 'Attack', element: 'Fire' }),
    card({ id: 'card_2', name: '水涟斩', type: 'Attack', element: 'Water' }),
    card({ id: 'card_3', name: '暗月法术', type: 'Magic', element: 'Dark' }),
    card({ id: 'card_4', name: '圣光', type: 'Magic', element: 'Light' }),
  ];
}

function moonDarkMoonCover(fieldIndex: number, name: string, type: 'Attack' | 'Magic', element: FieldCard['card']['element']): FieldCard {
  return {
    card: card({
      id: `mg-dark-moon-${fieldIndex}`,
      name,
      type,
      element,
    }),
    owner_id: MG_PLAYER_ID,
    source_id: MG_PLAYER_ID,
    mode: 'Cover',
    effect: 'MoonDarkMoon',
    field_hook: 'Manual',
    locked: false,
    duration: 0,
  };
}

function moonDarkMoonCovers(count: number): FieldCard[] {
  const fixtures = [
    { name: '暗月法术', type: 'Magic' as const, element: 'Dark' as const },
    { name: '火焰斩', type: 'Attack' as const, element: 'Fire' as const },
    { name: '水涟斩', type: 'Attack' as const, element: 'Water' as const },
    { name: '风刃', type: 'Attack' as const, element: 'Wind' as const },
  ];
  return fixtures.slice(0, count).map((item, index) => (
    moonDarkMoonCover(index, item.name, item.type, item.element)
  ));
}

function moonGoddessPlayerView(options: {
  is_active?: boolean;
  heal?: number;
  max_heal?: number;
  dark_moon_cards?: number;
  new_moon_tokens?: number;
  petrify_tokens?: number;
} = {}) {
  const darkMoonCards = options.dark_moon_cards ?? 0;
  return playerView({
    id: MG_PLAYER_ID,
    name: 'E2E Moon Goddess',
    camp: 'Red',
    role: 'moon_goddess',
    hand: moonGoddessHand(),
    hand_count: moonGoddessHand().length,
    heal: options.heal ?? 2,
    max_heal: options.max_heal ?? 4,
    is_active: options.is_active ?? true,
    field: moonDarkMoonCovers(darkMoonCards),
    tokens: {
      mg_dark_moon: darkMoonCards,
      mg_new_moon: options.new_moon_tokens ?? 0,
      mg_petrify: options.petrify_tokens ?? 0,
    },
  });
}

// ============================================================
// 新月庇护 (mg_new_moon_shelter) - 后端通过 response_skills 自动触发
// 爆牌转化时进入暗月形态，吸收爆牌为暗月
// ============================================================

export function newMoonShelterScenario(): ProtocolHarnessScenario {
  const characters = [moonGoddessCharacter, enemyCharacter];

  const moonGoddess = moonGoddessPlayerView({ is_active: false });

  const players = [
    moonGoddess,
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
    myPlayerId: MG_PLAYER_ID,
    myPlayerName: 'E2E Moon Goddess',
    characters,
    players: [
      playerInfo({ id: MG_PLAYER_ID, name: 'E2E Moon Goddess', camp: 'Red', char_role: 'moon_goddess', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy Bot', camp: 'Blue', char_role: 'villain' }),
    ],
    initialState: syncState({
      turn_player_id: ENEMY_PLAYER_ID,
      turn_stage: 'ActionExecution',
      available_skills: [],
      characters,
      players,
      // 后端会设置 response_skills 触发确认弹框
    }),
  };
}

export function newMoonShelterResponsePrompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_skill',
    player_id: MG_PLAYER_ID,
    message: '你触发了响应技能【新月庇护】，请选择是否发动。',
    options: [
      {
        id: MG_NEW_MOON_SHelter_SKILL_ID,
        label: '新月庇护',
        button_label: '发动',
        hint: '将本次爆牌改为暗月并防止士气下降',
      },
      { id: 'skip', label: '跳过', button_label: '跳过', hint: '不发动响应技能' },
    ],
    presentation: { kind: 'skill_choice', layout: 'overlay', numeric_base: 0 },
    min: 1,
    max: 1,
  } satisfies Prompt);
}

// 爆牌转化场景：from_damage_draw=true，discarded_cards有牌
export function newMoonShelterTransformScenario(options: {
  discarded_cards?: number;
} = {}): ProtocolHarnessScenario {
  const discarded_cards = options.discarded_cards ?? 2;
  const characters = [moonGoddessCharacter, enemyCharacter];

  const moonGoddess = moonGoddessPlayerView({ is_active: false });

  const players = [
    moonGoddess,
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
    myPlayerId: MG_PLAYER_ID,
    myPlayerName: 'E2E Moon Goddess',
    characters,
    players: [
      playerInfo({ id: MG_PLAYER_ID, name: 'E2E Moon Goddess', camp: 'Red', char_role: 'moon_goddess', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy Bot', camp: 'Blue', char_role: 'villain' }),
    ],
    initialState: syncState({
      turn_player_id: ENEMY_PLAYER_ID,
      turn_stage: 'MoraleLoss',
      available_skills: [],
      characters,
      players,
      // 模拟爆牌转化上下文
    }),
  };
}

// ============================================================
// 美杜莎之眼 (mg_medusa_eye) - 后端通过 response_skills 触发
// 闇月选择使用 choice_type: mg_medusa_darkmoon_pick
// 弃牌通过 system_discard_cards，目标通过 min_targets 处理
// ============================================================

export function medusaEyeScenario(options: { dark_moon_cards?: number } = {}): ProtocolHarnessScenario {
  const dark_moon_cards = options.dark_moon_cards ?? 2;
  const characters = [moonGoddessCharacter, enemyCharacter, enemy2Character];

  const moonGoddess = moonGoddessPlayerView({ dark_moon_cards, is_active: true });

  const players = [
    moonGoddess,
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
    myPlayerId: MG_PLAYER_ID,
    myPlayerName: 'E2E Moon Goddess',
    characters,
    players: [
      playerInfo({ id: MG_PLAYER_ID, name: 'E2E Moon Goddess', camp: 'Red', char_role: 'moon_goddess', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy Bot', camp: 'Blue', char_role: 'villain' }),
      playerInfo({ id: ENEMY_2_PLAYER_ID, name: 'Enemy Bot 2', camp: 'Blue', char_role: 'villain' }),
    ],
    initialState: syncState({
      turn_player_id: MG_PLAYER_ID,
      turn_stage: 'ActionExecution',
      available_skills: [],
      characters,
      players,
    }),
  };
}

export function medusaEyeDarkMoonPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: MG_PLAYER_ID,
    message: '【美杜莎之眼】请选择要展示并移除的同系闇月：',
    choice_type: 'mg_medusa_darkmoon_pick',
    skill_id: MG_MEDUSA_EYE_SKILL_ID,
    options: [
      { id: '0', label: '移除闇月[暗月法术/Magic/Dark]', button_label: '移除闇月[0]', field_index: 0, card_id: 'mg-dark-moon-0' },
      { id: '1', label: '移除闇月[火焰斩/Attack/Fire]', button_label: '移除闇月[1]', field_index: 1, card_id: 'mg-dark-moon-1' },
    ],
    presentation: { kind: 'card_picker', layout: 'field_cover', card_source: 'field', card_filter: 'effect:MoonDarkMoon', numeric_base: 0 },
    interaction: fieldOptionIndexInteraction,
    min: 1,
    max: 1,
  } satisfies Prompt);
}

export function medusaEyeResponsePrompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_skill',
    player_id: MG_PLAYER_ID,
    message: '你触发了响应技能【美杜莎之眼】，请选择是否发动。',
    options: [
      {
        id: MG_MEDUSA_EYE_SKILL_ID,
        label: '美杜莎之眼',
        button_label: '发动',
        hint: '移除同系闇月并获得治疗、石化',
      },
      { id: 'skip', label: '跳过', button_label: '跳过', hint: '不发动响应技能' },
    ],
    presentation: { kind: 'skill_choice', layout: 'overlay', numeric_base: 0 },
    min: 1,
    max: 1,
  } satisfies Prompt);
}

// 法术闇月弃牌后续：mg_medusa_magic_discard
export function medusaEyeMagicDiscardPrompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_cards',
    player_id: MG_PLAYER_ID,
    message: '【美杜莎之眼】因移除了法术闇月，请弃1张手牌：',
    choice_type: 'mg_medusa_magic_discard',
    skill_id: MG_MEDUSA_EYE_SKILL_ID,
    options: [
      { id: 'card_1', label: '1: 火焰斩 (火 Attack)', button_label: '选择', card_id: 'card_1' },
      { id: 'card_2', label: '2: 水涟斩 (水 Attack)', button_label: '选择', card_id: 'card_2' },
      { id: 'card_3', label: '3: 暗月法术 (暗 Magic)', button_label: '选择', card_id: 'card_3' },
      { id: 'card_4', label: '4: 圣光 (光 Magic)', button_label: '选择', card_id: 'card_4' },
    ],
    presentation: { kind: 'card_picker', card_source: 'hand', card_filter: 'option_limited', numeric_base: 0 },
    min: 1,
    max: 1,
  } satisfies Prompt);
}

// ============================================================
// 月之轮回 (mg_moon_cycle) - Response skill at turn end with branch choice
// ============================================================

export function moonCycleScenario(options: { dark_moon_cards?: number; heal?: number } = {}): ProtocolHarnessScenario {
  const dark_moon_cards = options.dark_moon_cards ?? 1;
  const heal = options.heal ?? 2;
  const characters = [moonGoddessCharacter, enemyCharacter];

  const moonGoddess = moonGoddessPlayerView({ dark_moon_cards, heal, is_active: false });

  const players = [
    moonGoddess,
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
    myPlayerId: MG_PLAYER_ID,
    myPlayerName: 'E2E Moon Goddess',
    characters,
    players: [
      playerInfo({ id: MG_PLAYER_ID, name: 'E2E Moon Goddess', camp: 'Red', char_role: 'moon_goddess', is_host: true }),
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

export function moonCycleBranchPrompt(options: { branch1?: boolean; branch2?: boolean } = {}): WsMessage {
  const branch1 = options.branch1 ?? true
  const branch2 = options.branch2 ?? true
  const promptOptions: Prompt['options'] = []
  if (branch1) {
    promptOptions.push({ id: 'branch1', label: '分支①：移除1个闇月，令目标角色+1治疗', button_label: '分支①' })
  }
  if (branch2) {
    promptOptions.push({ id: 'branch2', label: '分支②：移除1点治疗，你+1新月', button_label: '分支②' })
  }
  promptOptions.push({ id: 'decline', label: '不发动', button_label: '不发动' })
  return requireActionMessage({
    type: 'confirm',
    player_id: MG_PLAYER_ID,
    message: '【月之轮回】请选择发动分支：',
    choice_type: 'mg_moon_cycle_mode',
    skill_id: MG_MOON_CYCLE_SKILL_ID,
    options: promptOptions,
    presentation: { kind: 'branch_select', layout: 'overlay', cancel_policy: 'decline', has_decline: true, decline_index: promptOptions.length - 1, numeric_base: 0 },
    min: 1,
    max: 1,
  } satisfies Prompt);
}

export function moonCycleTargetPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: MG_PLAYER_ID,
    message: '【月之轮回】分支一：请选择目标令其+1治疗：',
    choice_type: 'mg_moon_cycle_heal_target',
    skill_id: MG_MOON_CYCLE_SKILL_ID,
    options: [
      { id: ENEMY_PLAYER_ID, target_id: ENEMY_PLAYER_ID, label: '恶徒', button_label: '选择' },
    ],
    presentation: { kind: 'target_picker', target_filter: 'custom', numeric_base: 0 },
    min: 1,
    max: 1,
  } satisfies Prompt);
}

// ============================================================
// 月渎 (mg_blasphemy) - 后端通过 response_skills 自动触发
// 目标选择通过 mg_blasphemy_target choice_type
// ============================================================

export function moonReadScenario(options: { heal?: number } = {}): ProtocolHarnessScenario {
  const heal = options.heal ?? 2;
  const characters = [moonGoddessCharacter, enemyCharacter, enemy2Character];

  const moonGoddess = moonGoddessPlayerView({ heal, is_active: true });

  const players = [
    moonGoddess,
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
    myPlayerId: MG_PLAYER_ID,
    myPlayerName: 'E2E Moon Goddess',
    characters,
    players: [
      playerInfo({ id: MG_PLAYER_ID, name: 'E2E Moon Goddess', camp: 'Red', char_role: 'moon_goddess', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy Bot', camp: 'Blue', char_role: 'villain' }),
      playerInfo({ id: ENEMY_2_PLAYER_ID, name: 'Enemy Bot 2', camp: 'Blue', char_role: 'villain' }),
    ],
    initialState: syncState({
      turn_player_id: MG_PLAYER_ID,
      turn_stage: 'ActionExecution',
      available_skills: [],
      characters,
      players,
      // 后端会设置 response_skills 触发确认弹框
    }),
  };
}

// 月渎确认：是否发动
export function moonReadConfirmPrompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_skill',
    player_id: MG_PLAYER_ID,
    message: '你触发了响应技能【月渎】，请选择是否发动。',
    options: [
      {
        id: MG_MOON_READ_SKILL_ID,
        label: '月渎',
        button_label: '发动',
        hint: '对受伤目标追加1点法术伤害',
      },
      { id: 'skip', label: '跳过', button_label: '跳过', hint: '不发动响应技能' },
    ],
    presentation: { kind: 'skill_choice', layout: 'overlay', numeric_base: 0 },
    min: 1,
    max: 1,
  } satisfies Prompt);
}

// 月渎目标选择：mg_blasphemy_target
export function moonReadTargetPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: MG_PLAYER_ID,
    message: '【月渎】请选择目标或不发动：',
    choice_type: 'mg_blasphemy_target',
    skill_id: MG_MOON_READ_SKILL_ID,
    presentation: { kind: 'branch_select', layout: 'overlay', has_decline: true, decline_index: 2, numeric_base: 0 },
    options: [
      { id: '0', label: '对 Enemy Bot 造成1点法术伤害', button_label: 'Enemy Bot' },
      { id: '1', label: '对 Enemy Bot 2 造成1点法术伤害', button_label: 'Enemy Bot 2' },
      { id: '2', label: '不发动', button_label: '不发动' },
    ],
    min: 1,
    max: 1,
  } satisfies Prompt);
}

// ============================================================
// 闇月斩 (mg_darkmoon_slash) - 后端通过 response_skills 自动触发
// X值选择通过 mg_darkmoon_slash_x choice_type
// ============================================================

export function darkmoonSlashScenario(options: {
  dark_moon_cards?: number;
} = {}): ProtocolHarnessScenario {
  const dark_moon_cards = options.dark_moon_cards ?? 2;
  const characters = [moonGoddessCharacter, enemyCharacter];

  const moonGoddess = moonGoddessPlayerView({ dark_moon_cards, is_active: true });

  const players = [
    moonGoddess,
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
    myPlayerId: MG_PLAYER_ID,
    myPlayerName: 'E2E Moon Goddess',
    characters,
    players: [
      playerInfo({ id: MG_PLAYER_ID, name: 'E2E Moon Goddess', camp: 'Red', char_role: 'moon_goddess', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy Bot', camp: 'Blue', char_role: 'villain' }),
    ],
    initialState: syncState({
      turn_player_id: MG_PLAYER_ID,
      turn_stage: 'ActionExecution',
      available_skills: [],
      characters,
      players,
      // 后端会设置 response_skills 触发确认弹框
    }),
  };
}

// 闇月斩X值选择：mg_darkmoon_slash_x
export function darkmoonSlashResponsePrompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_skill',
    player_id: MG_PLAYER_ID,
    message: '你触发了响应技能【闇月斩】，请选择是否发动。',
    options: [
      {
        id: MG_DARKMOON_SLASH_SKILL_ID,
        label: '闇月斩',
        button_label: '发动',
        hint: '消耗1点蓝水晶（红宝石可替代），移除X个闇月使本次攻击伤害+X',
      },
      { id: 'skip', label: '跳过', button_label: '跳过', hint: '不发动响应技能' },
    ],
    presentation: { kind: 'skill_choice', layout: 'overlay', numeric_base: 0 },
    min: 1,
    max: 1,
  } satisfies Prompt);
}

export function darkmoonSlashXPrompt(maxX: number): WsMessage {
  const options: { id: string; label: string; button_label: string }[] = [];
  for (let x = 1; x <= maxX; x++) {
    options.push({ id: `${x}`, label: `移除${x}个闇月，本次攻击伤害额外+${x}`, button_label: `${x}` });
  }
  return requireActionMessage({
    type: 'confirm',
    player_id: MG_PLAYER_ID,
    message: '【闇月斩】请选择X值：',
    choice_type: 'mg_darkmoon_slash_x',
    skill_id: MG_DARKMOON_SLASH_SKILL_ID,
    options,
    presentation: { kind: 'numeric', numeric_base: 0 },
    min: 1,
    max: 1,
  } satisfies Prompt);
}

// ============================================================
// 苍白之月 (mg_pale_moon) - 后端通过 response_skills 自动触发
// 分支选择使用 choice_type: mg_pale_moon_mode
// X值选择使用 choice_type: mg_pale_moon_x
// 弃牌通过 system_discard_cards，目标通过 min_targets 处理
// ============================================================

export function paleMoonScenario(options: {
  petrify_tokens?: number;
  new_moon_tokens?: number;
  availableSkills?: AvailableSkill[];
} = {}): ProtocolHarnessScenario {
  const petrify_tokens = options.petrify_tokens ?? 3;
  const new_moon_tokens = options.new_moon_tokens ?? 2;
  const characters = [moonGoddessCharacter, enemyCharacter, enemy2Character];

  const moonGoddess = moonGoddessPlayerView({ petrify_tokens, new_moon_tokens, is_active: true });

  const players = [
    moonGoddess,
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
    myPlayerId: MG_PLAYER_ID,
    myPlayerName: 'E2E Moon Goddess',
    characters,
    players: [
      playerInfo({ id: MG_PLAYER_ID, name: 'E2E Moon Goddess', camp: 'Red', char_role: 'moon_goddess', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy Bot', camp: 'Blue', char_role: 'villain' }),
      playerInfo({ id: ENEMY_2_PLAYER_ID, name: 'Enemy Bot 2', camp: 'Blue', char_role: 'villain' }),
    ],
    initialState: syncState({
      turn_player_id: MG_PLAYER_ID,
      turn_stage: 'ActionExecution',
      available_skills: options.availableSkills ?? [],
      characters,
      players,
      // 后端会设置 response_skills 触发确认弹框
    }),
  };
}

export function paleMoonAvailableSkill(): AvailableSkill {
  return availableSkill({
    id: MG_PALE_MOON_SKILL_ID,
    title: '苍白之月',
    cost_gem: 1,
  });
}

export function paleMoonBranchPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: MG_PLAYER_ID,
    message: '【苍白之月】请选择分支：',
    choice_type: 'mg_pale_moon_mode',
    skill_id: MG_PALE_MOON_SKILL_ID,
    options: [
      { id: '0', label: '分支①：移除3石化，强化下次主动攻击并获得额外回合', button_label: '分支①' },
      { id: '1', label: '分支②：移除X新月，弃1张牌并造成(X+1)法术伤害', button_label: '分支②' },
    ],
    presentation: { kind: 'branch_select', layout: 'overlay', numeric_base: 0 },
    min: 1,
    max: 1,
  } satisfies Prompt);
}

export function paleMoonXPrompt(xMax: number): WsMessage {
  const options: { id: string; label: string; button_label: string }[] = [];
  for (let i = 1; i <= xMax; i++) {
    options.push({ id: `${i}`, label: `X=${i}（目标法术伤害=${i + 1}）`, button_label: `${i}` });
  }
  return requireActionMessage({
    type: 'confirm',
    player_id: MG_PLAYER_ID,
    message: `【苍白之月】分支②请选择X值：`,
    choice_type: 'mg_pale_moon_x',
    skill_id: MG_PALE_MOON_SKILL_ID,
    options,
    presentation: { kind: 'numeric', numeric_base: 0 },
    min: 1,
    max: 1,
  } satisfies Prompt);
}

// 苍白之月目标选择：mg_pale_moon_target
export function paleMoonTargetPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: MG_PLAYER_ID,
    message: '【苍白之月】分支②请选择目标对手：',
    choice_type: 'mg_pale_moon_target',
    skill_id: MG_PALE_MOON_SKILL_ID,
    presentation: { kind: 'target_picker', target_filter: 'enemies', numeric_base: 0 },
    options: [
      { id: ENEMY_PLAYER_ID, target_id: ENEMY_PLAYER_ID, label: 'Enemy Bot', button_label: '选择' },
      { id: ENEMY_2_PLAYER_ID, target_id: ENEMY_2_PLAYER_ID, label: 'Enemy Bot 2', button_label: '选择' },
    ],
    min: 1,
    max: 1,
  } satisfies Prompt);
}

// 苍白之月弃牌：mg_pale_moon_discard
export function paleMoonDiscardPrompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_cards',
    player_id: MG_PLAYER_ID,
    message: '【苍白之月】分支②请弃1张牌：',
    choice_type: 'mg_pale_moon_discard',
    skill_id: MG_PALE_MOON_SKILL_ID,
    options: [
      { id: '0', label: '1: 火焰斩 (火 Attack)', button_label: '选择', card_id: 'card_1' },
      { id: '1', label: '2: 水涟斩 (水 Attack)', button_label: '选择', card_id: 'card_2' },
      { id: '2', label: '3: 暗月法术 (暗 Magic)', button_label: '选择', card_id: 'card_3' },
      { id: '3', label: '4: 圣光 (光 Magic)', button_label: '选择', card_id: 'card_4' },
    ],
    min: 1,
    max: 1,
    presentation: { kind: 'card_picker', card_source: 'hand', card_filter: 'option_limited', numeric_base: 0 },
  } satisfies Prompt);
}
