// ============================================================
// Prayer (祈祷师) Protocol Harness Scenarios
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

export const PRAYER_PLAYER_ID = 'prayer_player';
export const ENEMY_PLAYER_ID = 'enemy_1';
export const ALLY_PLAYER_ID = 'ally_1';

export const PRAYER_GLORY_BELIEF_ID = 'prayer_glory_belief';
export const PRAYER_DARK_CURSE_ID = 'prayer_dark_curse';
export const PRAYER_POWER_BLESSING_ID = 'prayer_power_blessing';
export const PRAYER_SWIFT_BLESSING_ID = 'prayer_swift_blessing';
export const PRAYER_PRAY_ID = 'prayer_pray';
export const PRAYER_MANA_TIDE_ID = 'prayer_mana_tide';

const prayerCharacter = characterView({
  id: 'prayer',
  name: '祈祷师',
  title: '咏',
  faction: '咏',
  skills: [
    {
      id: PRAYER_GLORY_BELIEF_ID,
      title: '光辉信仰',
      description: '（祈祷形态中）弃1张法术牌［展示］，令一名队友+2［治疗］。',
      type: 2, // 法术
      min_targets: 1, max_targets: 1, target_type: 1,
      cost_gem: 0, cost_crystal: 0, cost_discards: 0,
    },
    {
      id: PRAYER_DARK_CURSE_ID,
      title: '黑暗诅咒',
      description: '（祈祷形态中）你弃1张牌，对一名对手造成1点法术伤害③，该对手也弃1张牌。',
      type: 2, // 法术
      min_targets: 1, max_targets: 1, target_type: 2,
      cost_gem: 0, cost_crystal: 0, cost_discards: 0,
    },
    {
      id: PRAYER_POWER_BLESSING_ID,
      title: '威力赐福',
      description: '（将威力赐福放置于一名队友面前）该队友下次命中时，可以对目标造成1点法术伤害③。',
      type: 2, // 独有法术
      min_targets: 1, max_targets: 1, target_type: 1,
      cost_gem: 0, cost_crystal: 0, cost_discards: 0,
    },
    {
      id: PRAYER_SWIFT_BLESSING_ID,
      title: '迅捷赐福',
      description: '（将迅捷赐福放置于一名队友面前）该队友下次行动结束时，可以摸1张牌。',
      type: 2, // 独有法术
      min_targets: 1, max_targets: 1, target_type: 1,
      cost_gem: 0, cost_crystal: 0, cost_discards: 0,
    },
    {
      id: PRAYER_PRAY_ID,
      title: '祈祷',
      description: '［宝石］横置，进入祈祷形态。',
      type: 1, // 启动(大招)
      min_targets: 0, max_targets: 0, target_type: 0,
      cost_gem: 0, cost_crystal: 0, cost_discards: 0,
    },
    {
      id: PRAYER_MANA_TIDE_ID,
      title: '法力潮汐',
      description: '［水晶］（法术行动结束时发动）进行一次法术行动。',
      type: 3, // 响应(大招)
      min_targets: 0, max_targets: 0, target_type: 0,
      cost_gem: 0, cost_crystal: 0, cost_discards: 0,
    },
  ],
});

const enemyCharacter = characterView({
  id: 'enemy_char', name: '守卫', title: '测试目标', faction: '异端', skills: [],
});

const allyCharacter = characterView({
  id: 'ally_char', name: '勇者', title: '测试队友', faction: '咏', skills: [],
});

const defaultCharacters = [prayerCharacter, enemyCharacter, allyCharacter];

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function prayerHand(): Card[] {
  return [
    card({ id: 'prayer-magic-1', name: '圣光', type: 'Magic', element: 'Light' }),
    card({ id: 'prayer-magic-2', name: '治愈', type: 'Magic', element: 'Water' }),
    card({ id: 'prayer-magic-3', name: '祝福', type: 'Magic', element: 'Light' }),
    card({ id: 'prayer-attack-1', name: '光刃', type: 'Attack', element: 'Light' }),
    card({ id: PRAYER_POWER_BLESSING_ID, name: '威力赐福', type: 'Magic', element: 'Light' }),
    card({ id: PRAYER_SWIFT_BLESSING_ID, name: '迅捷赐福', type: 'Magic', element: 'Light' }),
  ];
}

function prayerAvailableSkill(skill: Partial<AvailableSkill> & { id: string; title: string }): AvailableSkill {
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

export function prayerScenario(options: {
  hand?: Card[];
  availableSkills?: AvailableSkill[];
  crystal?: number;
  gem?: number;
  turnStage?: string;
  buffs?: { id: string; name: string; duration: number; value: number; source_id: string }[];
} = {}): ProtocolHarnessScenario {
  const hand = options.hand ?? prayerHand();
  const players = [
    playerView({
      id: PRAYER_PLAYER_ID,
      name: 'E2E Prayer',
      camp: 'Red',
      role: 'prayer',
      hand,
      hand_count: hand.length,
      crystal: options.crystal ?? 0,
      gem: options.gem ?? 0,
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
    myPlayerId: PRAYER_PLAYER_ID,
    myPlayerName: 'E2E Prayer',
    characters: defaultCharacters,
    players: [
      playerInfo({ id: PRAYER_PLAYER_ID, name: 'E2E Prayer', camp: 'Red', char_role: 'prayer', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy E1', camp: 'Blue', char_role: 'enemy_char' }),
      playerInfo({ id: ALLY_PLAYER_ID, name: 'Ally A1', camp: 'Red', char_role: 'ally_char' }),
    ],
    initialState: syncState({
      turn_player_id: PRAYER_PLAYER_ID,
      turn_stage: options.turnStage ?? 'ActionExecution',
      available_skills: options.availableSkills ?? [],
      characters: defaultCharacters,
      players,
    }),
  };
}

// ============================================================
// Glory Belief (光辉信仰) - 法术技能
// ============================================================

export function gloryBeliefScenario(): ProtocolHarnessScenario {
  return prayerScenario({
    buffs: [{ id: 'prayer_form', name: '祈祷形态', duration: 0, value: 0, source_id: PRAYER_PRAY_ID }],
    availableSkills: [
      prayerAvailableSkill({
        id: PRAYER_GLORY_BELIEF_ID, title: '光辉信仰',
        min_targets: 1, max_targets: 1, target_type: 1,
      }),
    ],
  });
}

export function gloryBeliefDiscardPrompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_cards',
    player_id: PRAYER_PLAYER_ID,
    message: '【光辉信仰】请选择弃1张法术牌［展示］：',
    choice_type: 'prayer_glory_belief_discard',
    options: [
      { id: 'prayer-magic-1', label: '圣光（法术）', button_label: '选择', card_id: 'prayer-magic-1' },
      { id: 'prayer-magic-2', label: '治愈（法术）', button_label: '选择', card_id: 'prayer-magic-2' },
      { id: 'prayer-magic-3', label: '祝福（法术）', button_label: '选择', card_id: 'prayer-magic-3' },
    ],
    min: 1, max: 1,
    presentation: { kind: 'card_picker', card_source: 'hand', card_filter: 'option_limited', numeric_base: 0 },
  } satisfies Prompt);
}

export function gloryBeliefTargetPrompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_target',
    player_id: PRAYER_PLAYER_ID,
    message: '【光辉信仰】请选择一名队友+2治疗：',
    choice_type: 'prayer_glory_belief_target',
    presentation: { kind: 'target_picker', target_filter: 'allies', numeric_base: 0 },
    options: [
      { id: ALLY_PLAYER_ID, target_id: ALLY_PLAYER_ID, label: 'Ally A1', button_label: '选择' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}

// ============================================================
// Dark Curse (黑暗诅咒) - 法术技能
// ============================================================

export function darkCurseScenario(): ProtocolHarnessScenario {
  return prayerScenario({
    buffs: [{ id: 'prayer_form', name: '祈祷形态', duration: 0, value: 0, source_id: PRAYER_PRAY_ID }],
    availableSkills: [
      prayerAvailableSkill({
        id: PRAYER_DARK_CURSE_ID, title: '黑暗诅咒',
        min_targets: 1, max_targets: 1, target_type: 2,
      }),
    ],
  });
}

export function darkCurseDiscardPrompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_cards',
    player_id: PRAYER_PLAYER_ID,
    message: '【黑暗诅咒】请选择弃1张牌：',
    choice_type: 'prayer_dark_curse_discard',
    options: [
      { id: 'prayer-magic-1', label: '圣光', button_label: '选择', card_id: 'prayer-magic-1' },
      { id: 'prayer-magic-2', label: '治愈', button_label: '选择', card_id: 'prayer-magic-2' },
      { id: 'prayer-attack-1', label: '光刃', button_label: '选择', card_id: 'prayer-attack-1' },
    ],
    min: 1, max: 1,
    presentation: { kind: 'card_picker', card_source: 'hand', card_filter: 'option_limited', numeric_base: 0 },
  } satisfies Prompt);
}

export function darkCurseTargetPrompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_target',
    player_id: PRAYER_PLAYER_ID,
    message: '【黑暗诅咒】请选择一名对手造成1点法术伤害：',
    choice_type: 'prayer_dark_curse_target',
    presentation: { kind: 'target_picker', target_filter: 'enemies', numeric_base: 0 },
    options: [
      { id: ENEMY_PLAYER_ID, target_id: ENEMY_PLAYER_ID, label: 'Enemy E1', button_label: '选择' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}

// ============================================================
// Power Blessing (威力赐福) - 独有法术
// ============================================================

export function powerBlessingScenario(): ProtocolHarnessScenario {
  return prayerScenario({
    availableSkills: [
      prayerAvailableSkill({
        id: PRAYER_POWER_BLESSING_ID, title: '威力赐福',
        min_targets: 1, max_targets: 1, target_type: 1,
      }),
    ],
  });
}

export function powerBlessingTargetPrompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_target',
    player_id: PRAYER_PLAYER_ID,
    message: '【威力赐福】请选择一名队友放置赐福：',
    choice_type: 'prayer_power_blessing_target',
    presentation: { kind: 'target_picker', target_filter: 'allies', numeric_base: 0 },
    options: [
      { id: ALLY_PLAYER_ID, target_id: ALLY_PLAYER_ID, label: 'Ally A1', button_label: '选择' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}

// 威力赐福触发：队友命中后额外伤害确认
// 威力赐福触发场景：从队友视角
export function powerBlessingTriggerScenario(): ProtocolHarnessScenario {
  const hand = prayerHand();
  const characters = defaultCharacters;
  const players = [
    playerView({
      id: PRAYER_PLAYER_ID,
      name: 'E2E Prayer',
      camp: 'Red',
      role: 'prayer',
      hand,
      hand_count: hand.length,
      crystal: 0,
      gem: 0,
      is_active: false,
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
      is_active: true,
    }),
  ];

  return {
    roomCode: 'MOCK',
    myPlayerId: ALLY_PLAYER_ID,
    myPlayerName: 'Ally A1',
    characters,
    players: [
      playerInfo({ id: PRAYER_PLAYER_ID, name: 'E2E Prayer', camp: 'Red', char_role: 'prayer', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy E1', camp: 'Blue', char_role: 'enemy_char' }),
      playerInfo({ id: ALLY_PLAYER_ID, name: 'Ally A1', camp: 'Red', char_role: 'ally_char' }),
    ],
    initialState: syncState({
      turn_player_id: ALLY_PLAYER_ID,
      turn_stage: 'ActionExecution',
      available_skills: [],
      characters,
      players,
    }),
  };
}

// 迅捷赐福触发场景：从队友视角（同上）
export function swiftBlessingTriggerScenario(): ProtocolHarnessScenario {
  return powerBlessingTriggerScenario();
}

export function powerBlessingTriggerPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: ALLY_PLAYER_ID,
    message: '【威力赐福】命中触发：对目标造成1点法术伤害？',
    choice_type: 'prayer_power_blessing_response',
    options: [
      { id: '0', label: '发动', button_label: '发动' },
      { id: '1', label: '跳过', button_label: '跳过' },
    ],
    presentation: { kind: 'branch_select', layout: 'overlay', numeric_base: 0 },
    min: 1, max: 1,
  } satisfies Prompt);
}

// ============================================================
// Swift Blessing (迅捷赐福) - 独有法术
// ============================================================

export function swiftBlessingScenario(): ProtocolHarnessScenario {
  return prayerScenario({
    availableSkills: [
      prayerAvailableSkill({
        id: PRAYER_SWIFT_BLESSING_ID, title: '迅捷赐福',
        min_targets: 1, max_targets: 1, target_type: 1,
      }),
    ],
  });
}

export function swiftBlessingTargetPrompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_target',
    player_id: PRAYER_PLAYER_ID,
    message: '【迅捷赐福】请选择一名队友放置赐福：',
    choice_type: 'prayer_swift_blessing_target',
    presentation: { kind: 'target_picker', target_filter: 'allies', numeric_base: 0 },
    options: [
      { id: ALLY_PLAYER_ID, target_id: ALLY_PLAYER_ID, label: 'Ally A1', button_label: '选择' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}

// 迅捷赐福触发：队友行动结束后摸牌确认
export function swiftBlessingTriggerPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: ALLY_PLAYER_ID,
    message: '【迅捷赐福】行动结束触发：摸1张牌？',
    choice_type: 'prayer_swift_blessing_response',
    options: [
      { id: '0', label: '摸牌', button_label: '摸牌' },
      { id: '1', label: '跳过', button_label: '跳过' },
    ],
    presentation: { kind: 'branch_select', layout: 'overlay', numeric_base: 0 },
    min: 1, max: 1,
  } satisfies Prompt);
}

// ============================================================
// Pray (祈祷) - 启动技能(大招)
// ============================================================

export function prayScenario(): ProtocolHarnessScenario {
  return prayerScenario({
    gem: 1,
    availableSkills: [
      prayerAvailableSkill({
        id: PRAYER_PRAY_ID, title: '祈祷',
      }),
    ],
  });
}

export function prayPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: PRAYER_PLAYER_ID,
    message: '【祈祷］消耗宝石，横置进入祈祷形态？',
    choice_type: 'prayer_pray',
    options: [
      { id: '0', label: '发动', button_label: '发动' },
      { id: '1', label: '跳过', button_label: '跳过' },
    ],
    presentation: { kind: 'branch_select', layout: 'overlay', numeric_base: 0 },
    min: 1, max: 1,
  } satisfies Prompt);
}

// ============================================================
// Mana Tide (法力潮汐) - 响应技能(大招)
// ============================================================

export function manaTideScenario(): ProtocolHarnessScenario {
  return prayerScenario({
    crystal: 1,
  });
}

export function manaTidePrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: PRAYER_PLAYER_ID,
    message: '【法力潮汐］消耗水晶，法术行动结束后再进行一次法术行动？',
    choice_type: 'prayer_mana_tide',
    options: [
      { id: '0', label: '发动', button_label: '发动' },
      { id: '1', label: '跳过', button_label: '跳过' },
    ],
    presentation: { kind: 'branch_select', layout: 'overlay', numeric_base: 0 },
    min: 1, max: 1,
  } satisfies Prompt);
}