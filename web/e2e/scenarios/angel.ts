// ============================================================
// Angel (天使) Protocol Harness Scenarios
// ============================================================

import type { AvailableSkill, Card, Prompt, FieldCard } from '../../src/types/game';
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

export const ANGEL_PLAYER_ID = 'angel_player';
export const ENEMY_PLAYER_ID = 'enemy_1';
export const ALLY_PLAYER_ID = 'ally_1';

export const ANGEL_BLESSING_ID = 'angel_blessing';
export const ANGEL_WIND_CLEANSE_ID = 'angel_wind_cleanse';
export const ANGEL_SONG_ID = 'angel_song';
export const ANGEL_PROTECTION_ID = 'angel_protection';
export const ANGEL_WALL_ID = 'angel_wall'; // 独有牌

const angelCharacter = characterView({
  id: 'angel',
  name: '天使',
  title: '圣',
  faction: '圣',
  skills: [
    {
      id: 'angel_bond',
      title: '天使羁绊',
      description: '（每当你移除一个基础效果或是使用［圣盾］时）目标角色+1［治疗］。',
      type: 0, // 被动
      min_targets: 1, max_targets: 1, target_type: 0,
    },
    {
      id: ANGEL_BLESSING_ID,
      title: '天使祝福',
      description: '（弃1张水系牌［展示］）指定目标玩家给你2张牌或指定2名角色各给你1张牌。',
      type: 2, // 法术
      min_targets: 0, max_targets: 0, target_type: 0,
      cost_discards: 1, discard_element: 'Water',
    },
    {
      id: ANGEL_WIND_CLEANSE_ID,
      title: '风之洁净',
      description: '（弃1张风系牌［展示］）移除场上任意1个基础效果。',
      type: 2, // 法术
      min_targets: 0, max_targets: 0, target_type: 0,
      cost_discards: 1, discard_element: 'Wind',
    },
    {
      id: ANGEL_SONG_ID,
      title: '天使之歌',
      description: '［回合限定］［水晶］（在你的回合开始前发动）移除场上任意1个基础效果。',
      type: 3, // 响应(大招)
      min_targets: 0, max_targets: 0, target_type: 0,
    },
    {
      id: ANGEL_PROTECTION_ID,
      title: '神之庇护',
      description: 'X个［水晶］为我方抵御X点因法术伤害而造成的士气下降。',
      type: 3, // 响应(大招)
      min_targets: 0, max_targets: 0, target_type: 0,
    },
  ],
});

const enemyCharacter = characterView({
  id: 'enemy_char', name: '守卫', title: '测试目标', faction: '异端', skills: [],
});

const allyCharacter = characterView({
  id: 'ally_char', name: '勇者', title: '测试队友', faction: '圣', skills: [],
});

const defaultCharacters = [angelCharacter, enemyCharacter, allyCharacter];

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function angelHand(): Card[] {
  return [
    card({ id: 'angel-water-atk1', name: '水刃', type: 'Attack', element: 'Water' }),
    card({ id: 'angel-water-atk2', name: '寒冰斩', type: 'Attack', element: 'Water' }),
    card({ id: 'angel-water-magic', name: '冰冻', type: 'Magic', element: 'Water' }),
    card({ id: 'angel-wind-atk', name: '风刃', type: 'Attack', element: 'Wind' }),
    card({ id: 'angel-wind-magic', name: '风行', type: 'Magic', element: 'Wind' }),
    card({ id: ANGEL_WALL_ID, name: '天使之墙', type: 'Magic', element: 'Light' }),
  ];
}

function angelAvailableSkill(skill: Partial<AvailableSkill> & { id: string; title: string }): AvailableSkill {
  return availableSkill({
    description: '',
    target_type: 0,
    min_targets: 0, max_targets: 0,
    ...skill,
  });
}

// 模拟场上基础效果
function createFieldEffect(cardId: string, name: string, ownerId: string): FieldCard {
  return {
    card: card({ id: cardId, name, type: 'Magic', element: 'Light' }),
    owner_id: ownerId,
    source_id: cardId,
    mode: 'Effect',
    effect: '基础效果',
    field_hook: '',
    locked: false,
    duration: 0,
  };
}

// ---------------------------------------------------------------------------
// Scenario Factory
// ---------------------------------------------------------------------------

export function angelScenario(options: {
  hand?: Card[];
  availableSkills?: AvailableSkill[];
  crystal?: number;
  gem?: number;
  turnStage?: string;
  fieldCards?: FieldCard[];
} = {}): ProtocolHarnessScenario {
  const hand = options.hand ?? angelHand();
  const players = [
    playerView({
      id: ANGEL_PLAYER_ID,
      name: 'E2E Angel',
      camp: 'Red',
      role: 'angel',
      hand,
      hand_count: hand.length,
      crystal: options.crystal ?? 0,
      gem: options.gem ?? 0,
      is_active: true,
      field: options.fieldCards ?? [],
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
      field: [createFieldEffect('enemy_shield', '敌方圣盾', ENEMY_PLAYER_ID)],
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
      field: [createFieldEffect('ally_buff', '队友增益', ALLY_PLAYER_ID)],
    }),
  ];

  return {
    roomCode: 'MOCK',
    myPlayerId: ANGEL_PLAYER_ID,
    myPlayerName: 'E2E Angel',
    characters: defaultCharacters,
    players: [
      playerInfo({ id: ANGEL_PLAYER_ID, name: 'E2E Angel', camp: 'Red', char_role: 'angel', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy E1', camp: 'Blue', char_role: 'enemy_char' }),
      playerInfo({ id: ALLY_PLAYER_ID, name: 'Ally A1', camp: 'Red', char_role: 'ally_char' }),
    ],
    initialState: syncState({
      turn_player_id: ANGEL_PLAYER_ID,
      turn_stage: options.turnStage ?? 'ActionExecution',
      available_skills: options.availableSkills ?? [],
      characters: defaultCharacters,
      players,
    }),
  };
}

// ============================================================
// Angel Blessing (天使祝福) - 法术技能
// ============================================================

export function angelBlessingScenario(): ProtocolHarnessScenario {
  return angelScenario({
    availableSkills: [
      angelAvailableSkill({
        id: ANGEL_BLESSING_ID, title: '天使祝福',
        cost_discards: 1, discard_element: 'Water',
      }),
    ],
  });
}

export function angelBlessingBranchPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: ANGEL_PLAYER_ID,
    message: '【天使祝福】请选择分支：',
    choice_type: 'angel_blessing_branch',
    options: [
      { id: 'branch1', label: '指定1名玩家给你2张牌', button_label: '指定1名玩家给你2张牌' },
      { id: 'branch2', label: '指定2名角色各给你1张牌', button_label: '指定2名角色各给你1张牌' },
    ],
    min: 1, max: 1,
    presentation: { kind: 'branch_select', layout: 'overlay', numeric_base: 0 },
  } satisfies Prompt);
}

export function angelBlessingSingleTargetPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: ANGEL_PLAYER_ID,
    message: '【天使祝福】请选择1名目标玩家：',
    choice_type: 'angel_blessing_single_target',
    options: [
      { id: ENEMY_PLAYER_ID, label: 'Enemy E1', button_label: '选择' },
      { id: ALLY_PLAYER_ID, label: 'Ally A1', button_label: '选择' },
    ],
    min: 1, max: 1,
    presentation: { kind: 'target_picker', target_filter: 'custom', numeric_base: 0 },
  } satisfies Prompt);
}

export function angelBlessingDualTargetPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: ANGEL_PLAYER_ID,
    message: '【天使祝福】请选择2名目标角色：',
    choice_type: 'angel_blessing_dual_target',
    options: [
      { id: ENEMY_PLAYER_ID, label: 'Enemy E1', button_label: '选择' },
      { id: ALLY_PLAYER_ID, label: 'Ally A1', button_label: '选择' },
    ],
    min: 2, max: 2,
    presentation: { kind: 'target_picker', target_filter: 'custom', multi_target: true, numeric_base: 0 },
  } satisfies Prompt);
}

// ============================================================
// Wind Cleanse (风之洁净) - 法术技能
// ============================================================

export function windCleanseScenario(): ProtocolHarnessScenario {
  return angelScenario({
    fieldCards: [
      createFieldEffect('field_effect_1', '场上效果A', ENEMY_PLAYER_ID),
    ],
    availableSkills: [
      angelAvailableSkill({
        id: ANGEL_WIND_CLEANSE_ID, title: '风之洁净',
        cost_discards: 1, discard_element: 'Wind',
      }),
    ],
  });
}

export function windCleanseFieldSelectPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: ANGEL_PLAYER_ID,
    message: '【风之洁净】请选择场上1个基础效果移除：',
    choice_type: 'basic_effect_pick',
    options: [
      { id: 'enemy_shield', label: '敌方圣盾（Enemy E1）', button_label: '敌方圣盾' },
      { id: 'ally_buff', label: '队友增益（Ally A1）', button_label: '队友增益' },
    ],
    min: 1, max: 1,
    presentation: { kind: 'branch_select', layout: 'overlay', numeric_base: 0 },
  } satisfies Prompt);
}

// ============================================================
// Angel Song (天使之歌) - 响应技能(大招)
// ============================================================

export function angelSongScenario(options: {
  crystal?: number;
} = {}): ProtocolHarnessScenario {
  return angelScenario({
    crystal: options.crystal ?? 1,
    turnStage: 'StartupPhase',
  });
}

export function angelSongBeforeTurnPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: ANGEL_PLAYER_ID,
    message: '【天使之歌】回合开始前，是否消耗1个水晶移除场上1个基础效果？',
    choice_type: 'angel_song_confirm',
    options: [
      { id: 'confirm', label: '发动', button_label: '发动' },
      { id: 'skip', label: '跳过', button_label: '跳过' },
    ],
    min: 1, max: 1,
    presentation: { kind: 'branch_select', layout: 'overlay', numeric_base: 0 },
  } satisfies Prompt);
}

export function angelSongFieldSelectPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: ANGEL_PLAYER_ID,
    message: '【天使之歌】请选择场上1个基础效果移除：',
    choice_type: 'basic_effect_pick',
    options: [
      { id: 'enemy_shield', label: '敌方圣盾', button_label: '敌方圣盾' },
      { id: 'ally_buff', label: '队友增益', button_label: '队友增益' },
    ],
    min: 1, max: 1,
    presentation: { kind: 'branch_select', layout: 'overlay', numeric_base: 0 },
  } satisfies Prompt);
}

// ============================================================
// God Protection (神之庇护) - 响应技能(大招)
// ============================================================

export function godProtectionScenario(options: {
  crystal?: number;
} = {}): ProtocolHarnessScenario {
  return angelScenario({
    crystal: options.crystal ?? 3,
  });
}

export function godProtectionPrompt(maxX: number): WsMessage {
  const options: Array<{ id: string; label: string; button_label: string }> = [];
  for (let x = 1; x <= maxX; x++) {
    options.push({ id: String(x - 1), label: `消耗${x}水晶抵御${x}点士气下降`, button_label: String(x) });
  }
  return requireActionMessage({
    type: 'confirm',
    player_id: ANGEL_PLAYER_ID,
    message: '【神之庇护】法术伤害导致士气下降，请选择消耗水晶数量：',
    choice_type: 'god_protection_x',
    options,
    min: 1, max: 1,
    presentation: { kind: 'numeric', numeric_base: 1 },
  } satisfies Prompt);
}
