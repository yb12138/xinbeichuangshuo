// ============================================================
// Sealer (封印师) Protocol Harness Scenarios
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

export const SEALER_PLAYER_ID = 'sealer_player';
export const ENEMY_PLAYER_ID = 'enemy_1';
export const ALLY_PLAYER_ID = 'ally_1';

export const SEALER_MAGIC_SURGE_ID = 'sealer_magic_surge';
export const SEALER_SEAL_BREAK_ID = 'sealer_seal_break';
export const SEALER_FIVE_ELEMENTS_BIND_ID = 'sealer_five_elements_bind';
export const SEALER_WATER_SEAL_ID = 'sealer_water_seal';
export const SEALER_FIRE_SEAL_ID = 'sealer_fire_seal';
export const SEALER_EARTH_SEAL_ID = 'sealer_earth_seal';
export const SEALER_WIND_SEAL_ID = 'sealer_wind_seal';
export const SEALER_THUNDER_SEAL_ID = 'sealer_thunder_seal';

const sealerCharacter = characterView({
  id: 'sealer',
  name: '封印师',
  title: '幻',
  faction: '幻',
  skills: [
    {
      id: SEALER_MAGIC_SURGE_ID,
      title: '法术激荡',
      description: '（［法术行动］结束时发动）额外+1［攻击行动］。',
      type: 3, // 响应
      min_targets: 0, max_targets: 0, target_type: 0,
    },
    {
      id: SEALER_SEAL_BREAK_ID,
      title: '封印破碎',
      description: '［水晶］将场上任意一张基础效果牌收入自己手中。',
      type: 2, // 法术(大招)
      min_targets: 0, max_targets: 0, target_type: 0,
    },
    {
      id: SEALER_FIVE_ELEMENTS_BIND_ID,
      title: '五系束缚',
      description: '［水晶］将五系束缚放置于目标对手面前，该对手跳过其下个行动阶段。在其下个行动阶段开始前他可以选择摸（2+X）张牌来取消五系束缚的效果。X为场上封印的数量，X最高为2。',
      type: 2, // 法术(独有)
      min_targets: 1, max_targets: 1, target_type: 2,
    },
    {
      id: SEALER_WATER_SEAL_ID,
      title: '水之封印',
      description: '（将水之封印放置于目标对手面前）该对手获得（直到他从手中打出或展示出水系牌时强制触发）对他造成3点法术伤害③，触发后移除此牌。',
      type: 2, // 法术(独有)
      min_targets: 1, max_targets: 1, target_type: 2,
    },
    {
      id: SEALER_FIRE_SEAL_ID,
      title: '火之封印',
      description: '（将火之封印放置于目标对手面前）该对手获得（直到他从手中打出或展示出火系牌时强制触发）对他造成3点法术伤害③，触发后移除此牌。',
      type: 2, // 法术(独有)
      min_targets: 1, max_targets: 1, target_type: 2,
    },
    {
      id: SEALER_EARTH_SEAL_ID,
      title: '地之封印',
      description: '（将地之封印放置于目标对手面前）该对手获得（直到他从手中打出或展示出地系牌时强制触发）对他造成3点法术伤害③，触发后移除此牌。',
      type: 2, // 法术(独有)
      min_targets: 1, max_targets: 1, target_type: 2,
    },
    {
      id: SEALER_WIND_SEAL_ID,
      title: '风之封印',
      description: '（将风之封印放置于目标对手面前）该对手获得（直到他从手中打出或展示出风系牌时强制触发）对他造成3点法术伤害③，触发后移除此牌。',
      type: 2, // 法术(独有)
      min_targets: 1, max_targets: 1, target_type: 2,
    },
    {
      id: SEALER_THUNDER_SEAL_ID,
      title: '雷之封印',
      description: '（将雷之封印放置于目标对手面前）该对手获得（直到他从手中打出或展示出雷系牌时强制触发）对他造成3点法术伤害③，触发后移除此牌。',
      type: 2, // 法术(独有)
      min_targets: 1, max_targets: 1, target_type: 2,
    },
  ],
});

const enemyCharacter = characterView({
  id: 'enemy_char', name: '守卫', title: '测试目标', faction: '异端', skills: [],
});

const allyCharacter = characterView({
  id: 'ally_char', name: '勇者', title: '测试队友', faction: '幻', skills: [],
});

const defaultCharacters = [sealerCharacter, enemyCharacter, allyCharacter];

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function sealerHand(): Card[] {
  return [
    card({ id: 'sealer-water-atk', name: '水刃', type: 'Attack', element: 'Water' }),
    card({ id: 'sealer-water-magic', name: '冰冻', type: 'Magic', element: 'Water' }),
    card({ id: 'sealer-fire-magic', name: '火球', type: 'Magic', element: 'Fire' }),
    card({ id: 'sealer-earth-magic', name: '地刺', type: 'Magic', element: 'Earth' }),
    card({ id: 'sealer-wind-magic', name: '风刃', type: 'Magic', element: 'Wind' }),
    card({ id: 'sealer-thunder-magic', name: '雷击', type: 'Magic', element: 'Thunder' }),
    card({ id: SEALER_WATER_SEAL_ID, name: '水之封印', type: 'Magic', element: 'Water' }),
    card({ id: SEALER_FIRE_SEAL_ID, name: '火之封印', type: 'Magic', element: 'Fire' }),
  ];
}

function sealerAvailableSkill(skill: Partial<AvailableSkill> & { id: string; title: string }): AvailableSkill {
  return availableSkill({
    description: '',
    target_type: 0,
    min_targets: 0, max_targets: 0,
    ...skill,
  });
}

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

export function sealerScenario(options: {
  hand?: Card[];
  availableSkills?: AvailableSkill[];
  crystal?: number;
  gem?: number;
  turnStage?: string;
  fieldCards?: FieldCard[];
} = {}): ProtocolHarnessScenario {
  const hand = options.hand ?? sealerHand();
  const players = [
    playerView({
      id: SEALER_PLAYER_ID,
      name: 'E2E Sealer',
      camp: 'Red',
      role: 'sealer',
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
      field: [
        createFieldEffect('enemy_shield', '敌方圣盾', ENEMY_PLAYER_ID),
      ],
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
    myPlayerId: SEALER_PLAYER_ID,
    myPlayerName: 'E2E Sealer',
    characters: defaultCharacters,
    players: [
      playerInfo({ id: SEALER_PLAYER_ID, name: 'E2E Sealer', camp: 'Red', char_role: 'sealer', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy E1', camp: 'Blue', char_role: 'enemy_char' }),
      playerInfo({ id: ALLY_PLAYER_ID, name: 'Ally A1', camp: 'Red', char_role: 'ally_char' }),
    ],
    initialState: syncState({
      turn_player_id: SEALER_PLAYER_ID,
      turn_stage: options.turnStage ?? 'ActionExecution',
      available_skills: options.availableSkills ?? [],
      characters: defaultCharacters,
      players,
    }),
  };
}

// ============================================================
// Magic Surge (法术激荡) - 响应技能
// ============================================================

export function magicSurgeScenario(): ProtocolHarnessScenario {
  return sealerScenario();
}

export function magicSurgePrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: SEALER_PLAYER_ID,
    message: '【法术激荡］法术行动结束，是否额外+1［攻击行动］？',
    choice_type: 'sealer_magic_surge',
    options: [
      { id: 'confirm', label: '发动' },
      { id: 'skip', label: '跳过' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}

// ============================================================
// Seal Break (封印破碎) - 法术技能(大招)
// ============================================================

export function sealBreakScenario(): ProtocolHarnessScenario {
  return sealerScenario({
    crystal: 1,
    availableSkills: [
      sealerAvailableSkill({
        id: SEALER_SEAL_BREAK_ID, title: '封印破碎',
      }),
    ],
  });
}

export function sealBreakFieldSelectPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: SEALER_PLAYER_ID,
    message: '【封印破碎】请选择场上一张基础效果牌收入手中：',
    choice_type: 'basic_effect_pick',
    options: [
      { id: 'enemy_shield', label: '敌方圣盾（Enemy E1）' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}

// ============================================================
// Five Elements Bind (五系束缚) - 法术技能(独有)
// ============================================================

// Note: sealCount is used in fiveElementsBindCancelPrompt, not in scenario setup
export function fiveElementsBindScenario(_options?: {
  sealCount?: number;
}): ProtocolHarnessScenario {
  return sealerScenario({
    crystal: 1,
    availableSkills: [
      sealerAvailableSkill({
        id: SEALER_FIVE_ELEMENTS_BIND_ID, title: '五系束缚',
        min_targets: 1, max_targets: 1, target_type: 2,
      }),
    ],
  });
}

export function fiveElementsBindTargetPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: SEALER_PLAYER_ID,
    message: '【五系束缚】请选择一名目标对手：',
    choice_type: 'five_elements_bind',
    options: [
      { id: ENEMY_PLAYER_ID, label: 'Enemy E1' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}

// 对手视角：五系束缚取消选择
export function fiveElementsBindCancelPrompt(x: number): WsMessage {
  const drawCount = 2 + x;
  return requireActionMessage({
    type: 'confirm',
    player_id: ENEMY_PLAYER_ID,
    message: `【五系束缚】下个行动阶段前，是否摸${drawCount}张牌取消束缚？`,
    choice_type: 'five_elements_bind',
    options: [
      { id: 'draw', label: `摸${drawCount}张牌取消` },
      { id: 'skip', label: '跳过行动阶段' },
    ],
    min: 1, max: 1,
    presentation: { kind: 'branch_select', layout: 'overlay' },
  } satisfies Prompt);
}

// ============================================================
// Elemental Seals (五系封印) - 法术技能(独有)
// ============================================================

export function elementalSealScenario(sealId: string): ProtocolHarnessScenario {
  return sealerScenario({
    availableSkills: [
      sealerAvailableSkill({
        id: sealId, title: getSealName(sealId),
        min_targets: 1, max_targets: 1, target_type: 2,
      }),
    ],
  });
}

function getSealName(sealId: string): string {
  switch (sealId) {
    case SEALER_WATER_SEAL_ID: return '水之封印';
    case SEALER_FIRE_SEAL_ID: return '火之封印';
    case SEALER_EARTH_SEAL_ID: return '地之封印';
    case SEALER_WIND_SEAL_ID: return '风之封印';
    case SEALER_THUNDER_SEAL_ID: return '雷之封印';
    default: return '封印';
  }
}

export function elementalSealTargetPrompt(sealId: string): WsMessage {
  const sealName = getSealName(sealId);
  return requireActionMessage({
    type: 'confirm',
    player_id: SEALER_PLAYER_ID,
    message: `【${sealName}】请选择一名目标对手放置封印：`,
    choice_type: 'sealer_elemental_seal_target',
    options: [
      { id: ENEMY_PLAYER_ID, label: 'Enemy E1' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}

// 封印触发伤害提示（对手视角）
export function sealTriggerPrompt(sealId: string): WsMessage {
  const sealName = getSealName(sealId);
  return requireActionMessage({
    type: 'confirm',
    player_id: ENEMY_PLAYER_ID,
    message: `【${sealName}】触发：受到3点法术伤害③`,
    choice_type: 'sealer_seal_trigger',
    options: [
      { id: 'confirm', label: '确认' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}

// 水之封印场景
export function waterSealScenario(): ProtocolHarnessScenario {
  return elementalSealScenario(SEALER_WATER_SEAL_ID);
}

export function waterSealTargetPrompt(): WsMessage {
  return elementalSealTargetPrompt(SEALER_WATER_SEAL_ID);
}

// 火之封印场景
export function fireSealScenario(): ProtocolHarnessScenario {
  return elementalSealScenario(SEALER_FIRE_SEAL_ID);
}

export function fireSealTargetPrompt(): WsMessage {
  return elementalSealTargetPrompt(SEALER_FIRE_SEAL_ID);
}

// 地之封印场景
export function earthSealScenario(): ProtocolHarnessScenario {
  return elementalSealScenario(SEALER_EARTH_SEAL_ID);
}

// 风之封印场景
export function windSealScenario(): ProtocolHarnessScenario {
  return elementalSealScenario(SEALER_WIND_SEAL_ID);
}

// 雷之封印场景
export function thunderSealScenario(): ProtocolHarnessScenario {
  return elementalSealScenario(SEALER_THUNDER_SEAL_ID);
}