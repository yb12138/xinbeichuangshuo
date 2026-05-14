// ============================================================
// Adventurer (冒险家) Protocol Harness Scenarios
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

export const ADVENTURER_PLAYER_ID = 'adventurer_player';
export const ENEMY_PLAYER_ID = 'enemy_1';
export const ALLY_PLAYER_ID = 'ally_1';

export const ADVENTURER_FRAUD_ID = 'adventurer_fraud';
export const ADVENTURER_ADVENTURER_PARADISE_ID = 'adventurer_adventurer_paradise';
export const ADVENTURER_STEAL_SKY_CHANGE_DAY_ID = 'adventurer_steal_sky_change_day';

const adventurerCharacter = characterView({
  id: 'adventurer',
  name: '冒险家',
  title: '咏',
  faction: '咏',
  skills: [
    {
      id: ADVENTURER_FRAUD_ID,
      title: '欺诈',
      description: '（购买时发动）弃2张牌，选择一个系别，本次购买视为该系别；或弃3张牌，本次购买视为暗系。',
      type: 3, // 响应
      min_targets: 0, max_targets: 0, target_type: 0,
    },
    {
      id: 'adventurer_underground_rule',
      title: '地下法则',
      description: '购买时，你可以支付额外费用。',
      type: 3, // 被动
      min_targets: 0, max_targets: 0, target_type: 0,
    },
    {
      id: ADVENTURER_ADVENTURER_PARADISE_ID,
      title: '冒险者天堂',
      description: '（提炼时发动）将提炼的能量转移给一名队友，并移除对手的1［能量］。',
      type: 3, // 响应
      min_targets: 0, max_targets: 0, target_type: 0,
    },
    {
      id: ADVENTURER_STEAL_SKY_CHANGE_DAY_ID,
      title: '偷天换日',
      description: '［水晶］选择：Ⅰ、将一名对手的1［能量］转移给一名队友；Ⅱ、将一名队友的手牌与你的一张手牌交换。',
      type: 2, // 法术(大招)
      min_targets: 0, max_targets: 0, target_type: 0,
    },
  ],
});

const enemyCharacter = characterView({
  id: 'enemy_char', name: '守卫', title: '测试目标', faction: '异端', skills: [],
});

const allyCharacter = characterView({
  id: 'ally_char', name: '勇者', title: '测试队友', faction: '咏', skills: [],
});

const defaultCharacters = [adventurerCharacter, enemyCharacter, allyCharacter];

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function adventurerHand(): Card[] {
  return [
    card({ id: 'adv-attack-1', name: '冒险斩', type: 'Attack', element: 'Light' }),
    card({ id: 'adv-attack-2', name: '探索斩', type: 'Attack', element: 'Light' }),
    card({ id: 'adv-magic-1', name: '火球', type: 'Magic', element: 'Fire' }),
    card({ id: 'adv-magic-2', name: '冰冻', type: 'Magic', element: 'Water' }),
    card({ id: 'adv-magic-3', name: '地刺', type: 'Magic', element: 'Earth' }),
    card({ id: 'adv-magic-4', name: '风刃', type: 'Magic', element: 'Wind' }),
    card({ id: 'adv-magic-5', name: '雷击', type: 'Magic', element: 'Thunder' }),
    card({ id: 'adv-dark-magic', name: '暗影', type: 'Magic', element: 'Dark' }),
  ];
}

function adventurerAvailableSkill(skill: Partial<AvailableSkill> & { id: string; title: string }): AvailableSkill {
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

export function adventurerScenario(options: {
  hand?: Card[];
  availableSkills?: AvailableSkill[];
  crystal?: number;
  gem?: number;
  turnStage?: string;
} = {}): ProtocolHarnessScenario {
  const hand = options.hand ?? adventurerHand();
  const players = [
    playerView({
      id: ADVENTURER_PLAYER_ID,
      name: 'E2E Adventurer',
      camp: 'Red',
      role: 'adventurer',
      hand,
      hand_count: hand.length,
      crystal: options.crystal ?? 0,
      gem: options.gem ?? 0,
      is_active: true,
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
      tokens: { energy: 1 },
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
    myPlayerId: ADVENTURER_PLAYER_ID,
    myPlayerName: 'E2E Adventurer',
    characters: defaultCharacters,
    players: [
      playerInfo({ id: ADVENTURER_PLAYER_ID, name: 'E2E Adventurer', camp: 'Red', char_role: 'adventurer', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy E1', camp: 'Blue', char_role: 'enemy_char' }),
      playerInfo({ id: ALLY_PLAYER_ID, name: 'Ally A1', camp: 'Red', char_role: 'ally_char' }),
    ],
    initialState: syncState({
      turn_player_id: ADVENTURER_PLAYER_ID,
      turn_stage: options.turnStage ?? 'ActionExecution',
      available_skills: options.availableSkills ?? [],
      characters: defaultCharacters,
      players,
    }),
  };
}

// ============================================================
// Fraud (欺诈) - 响应技能
// ============================================================

export function fraudScenario(): ProtocolHarnessScenario {
  return adventurerScenario();
}

export function fraudDiscardCountPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: ADVENTURER_PLAYER_ID,
    message: '【欺诈】购买时发动，选择弃牌数量：',
    choice_type: 'adventurer_fraud_pick',
    options: [
      { id: '2', label: '弃2张牌，选择系别' },
      { id: '3', label: '弃3张牌，视为暗系' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}

export function fraudDiscard2Prompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_cards',
    player_id: ADVENTURER_PLAYER_ID,
    message: '【欺诈】请选择弃2张牌：',
    choice_type: 'adventurer_fraud_pick',
    options: [
      { id: 'adv-attack-1', label: '冒险斩' },
      { id: 'adv-attack-2', label: '探索斩' },
      { id: 'adv-magic-1', label: '火球' },
      { id: 'adv-magic-2', label: '冰冻' },
    ],
    min: 2, max: 2,
  } satisfies Prompt);
}

export function fraudElementSelectPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: ADVENTURER_PLAYER_ID,
    message: '【欺诈】请选择本次购买视为的系别：',
    choice_type: 'adventurer_fraud_attack_element',
    options: [
      { id: 'Fire', label: '火系' },
      { id: 'Water', label: '水系' },
      { id: 'Earth', label: '地系' },
      { id: 'Wind', label: '风系' },
      { id: 'Thunder', label: '雷系' },
      { id: 'Light', label: '光系' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}

export function fraudDiscard3Prompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_cards',
    player_id: ADVENTURER_PLAYER_ID,
    message: '【欺诈】请选择弃3张牌（视为暗系）：',
    choice_type: 'adventurer_fraud_pick',
    options: [
      { id: 'adv-attack-1', label: '冒险斩' },
      { id: 'adv-attack-2', label: '探索斩' },
      { id: 'adv-magic-1', label: '火球' },
      { id: 'adv-magic-2', label: '冰冻' },
      { id: 'adv-magic-3', label: '地刺' },
    ],
    min: 3, max: 3,
  } satisfies Prompt);
}

// ============================================================
// Adventurer Paradise (冒险者天堂) - 响应技能
// ============================================================

export function adventurerParadiseScenario(): ProtocolHarnessScenario {
  return adventurerScenario();
}

export function adventurerParadisePrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: ADVENTURER_PLAYER_ID,
    message: '【冒险者天堂】提炼时发动，将能量转移给队友并移除对手能量？',
    choice_type: 'adventurer_extract_paradise_check',
    options: [
      { id: 'confirm', label: '发动' },
      { id: 'skip', label: '跳过' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}

export function adventurerParadiseTransferTargetPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: ADVENTURER_PLAYER_ID,
    message: '【冒险者天堂】请选择获得能量的队友：',
    choice_type: 'adventurer_paradise_pick',
    options: [
      { id: ALLY_PLAYER_ID, label: 'Ally A1' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}

export function adventurerParadiseRemoveTargetPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: ADVENTURER_PLAYER_ID,
    message: '【冒险者天堂】请选择移除能量的对手：',
    choice_type: 'adv_adventurer_paradise_remove_target',
    options: [
      { id: ENEMY_PLAYER_ID, label: 'Enemy E1' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}

// ============================================================
// Steal Sky Change Day (偷天换日) - 法术技能(大招)
// ============================================================

export function stealSkyChangeDayScenario(): ProtocolHarnessScenario {
  return adventurerScenario({
    crystal: 1,
    availableSkills: [
      adventurerAvailableSkill({
        id: ADVENTURER_STEAL_SKY_CHANGE_DAY_ID, title: '偷天换日',
      }),
    ],
  });
}

export function stealSkyChangeDayBranchPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: ADVENTURER_PLAYER_ID,
    message: '【偷天换日】请选择分支：',
    choice_type: 'adventurer_steal_sky_mode',
    options: [
      { id: 'energy_transfer', label: '转移对手能量给队友' },
      { id: 'card_swap', label: '与队友交换手牌' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}

// 能量转移分支
export function stealSkyChangeDayEnergySourcePrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: ADVENTURER_PLAYER_ID,
    message: '【偷天换日】请选择移除能量的对手：',
    choice_type: 'adv_steal_sky_change_day_energy_source',
    options: [
      { id: ENEMY_PLAYER_ID, label: 'Enemy E1' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}

export function stealSkyChangeDayEnergyTargetPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: ADVENTURER_PLAYER_ID,
    message: '【偷天换日】请选择获得能量的队友：',
    choice_type: 'adv_steal_sky_change_day_energy_target',
    options: [
      { id: ALLY_PLAYER_ID, label: 'Ally A1' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}

// 手牌交换分支
export function stealSkyChangeDayCardSwapTargetPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: ADVENTURER_PLAYER_ID,
    message: '【偷天换日】请选择交换手牌的队友：',
    choice_type: 'adv_steal_sky_change_day_card_swap_target',
    options: [
      { id: ALLY_PLAYER_ID, label: 'Ally A1' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}

export function stealSkyChangeDayMyCardPrompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_cards',
    player_id: ADVENTURER_PLAYER_ID,
    message: '【偷天换日】请选择你要给出的手牌：',
    choice_type: 'adv_steal_sky_change_day_my_card',
    options: [
      { id: 'adv-attack-1', label: '冒险斩' },
      { id: 'adv-attack-2', label: '探索斩' },
      { id: 'adv-magic-1', label: '火球' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}