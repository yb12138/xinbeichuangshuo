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
      description: '主动技能：选择1名敌方角色，弃同系牌将本次视为一次主动攻击（弃2张同系可选五系攻击〔不含暗灭〕；弃3张同系视为暗灭）。',
      type: 2, // 主动
      min_targets: 0, max_targets: 0, target_type: 0,
      cost_gem: 0, cost_crystal: 0, cost_discards: 0,
    },
    {
      id: 'adventurer_underground_rule',
      title: '地下法则',
      description: '购买时，你可以支付额外费用。',
      type: 3, // 被动
      min_targets: 0, max_targets: 0, target_type: 0,
      cost_gem: 0, cost_crystal: 0, cost_discards: 0,
    },
    {
      id: ADVENTURER_ADVENTURER_PARADISE_ID,
      title: '冒险者天堂',
      description: '（提炼时发动）将提炼的能量转移给一名队友，并移除对手的1［能量］。',
      type: 3, // 响应
      min_targets: 0, max_targets: 0, target_type: 0,
      cost_gem: 0, cost_crystal: 0, cost_discards: 0,
    },
    {
      id: ADVENTURER_STEAL_SKY_CHANGE_DAY_ID,
      title: '偷天换日',
      description: '［水晶］选择：Ⅰ、将一名对手的1［能量］转移给一名队友；Ⅱ、将一名队友的手牌与你的一张手牌交换。',
      type: 2, // 法术(大招)
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

const defaultCharacters = [adventurerCharacter, enemyCharacter, allyCharacter];

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function adventurerHand(): Card[] {
  return [
    card({ id: 'adv-attack-1', name: '冒险斩', type: 'Attack', element: 'Light' }),
    card({ id: 'adv-attack-2', name: '探索斩', type: 'Attack', element: 'Light' }),
    card({ id: 'adv-light-magic', name: '圣印', type: 'Magic', element: 'Light' }),
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
// Backend flow: multi-select same-element cards with PromptFlowState, then element if needed, then target
// ============================================================

export function fraudScenario(): ProtocolHarnessScenario {
  return adventurerScenario();
}

// Prompt to select 2-3 same-element hand cards in one submission.
export function fraudPickPrompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_cards',
    player_id: ADVENTURER_PLAYER_ID,
    message: '【欺诈】请选择2~3张同系手牌：',
    choice_type: 'adventurer_fraud_pick',
    // Backend uses hand indices as option.id.
    options: [
      { id: '0', label: '1: 冒险斩（光）', button_label: '选择', card_id: 'adv-attack-1' },
      { id: '1', label: '2: 探索斩（光）', button_label: '选择', card_id: 'adv-attack-2' },
      { id: '2', label: '3: 圣印（光）', button_label: '选择', card_id: 'adv-light-magic' },
      { id: '3', label: '4: 火球（火）', button_label: '选择', card_id: 'adv-magic-1' },
      { id: '4', label: '5: 冰冻（水）', button_label: '选择', card_id: 'adv-magic-2' },
      { id: '5', label: '6: 地刺（地）', button_label: '选择', card_id: 'adv-magic-3' },
      { id: '6', label: '7: 风刃（风）', button_label: '选择', card_id: 'adv-magic-4' },
      { id: '7', label: '8: 雷击（雷）', button_label: '选择', card_id: 'adv-magic-5' },
      { id: '8', label: '9: 暗影（暗）', button_label: '选择', card_id: 'adv-dark-magic' },
    ],
    presentation: { kind: 'card_picker', card_source: 'hand', card_filter: 'same_element_combo', numeric_base: 0 },
    min: 2, max: 3,
  } satisfies Prompt);
}

// Element selection prompt (backend uses element string as option.id)
// Backend options: Water, Fire, Earth, Wind, Thunder (no Light/Dark)
export function fraudElementPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: ADVENTURER_PLAYER_ID,
    message: '【欺诈】请选择攻击系别（不含光/暗）：',
    choice_type: 'adventurer_fraud_attack_element',
    options: [
      { id: 'Water', label: '水', button_label: '水' },
      { id: 'Fire', label: '火', button_label: '火' },
      { id: 'Earth', label: '地', button_label: '地' },
      { id: 'Wind', label: '风', button_label: '风' },
      { id: 'Thunder', label: '雷', button_label: '雷' },
    ],
    presentation: { kind: 'branch_select', layout: 'fraud_attack_element', numeric_base: 0 },
    min: 1, max: 1,
  } satisfies Prompt);
}

// Target selection prompt: click enemy avatar to submit the single target.
export function fraudTargetPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: ADVENTURER_PLAYER_ID,
    message: '【欺诈】请选择攻击目标：',
    choice_type: 'adventurer_fraud_target',
    options: [
      { id: '0', target_id: ENEMY_PLAYER_ID, label: 'Enemy E1', button_label: '选择' },
    ],
    presentation: { kind: 'target_picker', target_filter: 'custom', numeric_base: 0 },
    min: 1, max: 1,
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
    message: '是否发动[冒险者天堂]，让队友代为提炼？',
    choice_type: 'adventurer_extract_paradise_check',
    options: [
      { id: 'yes', label: '是，发动冒险者天堂', button_label: '是' },
      { id: 'no', label: '否，自行提炼', button_label: '否' },
    ],
    min: 1, max: 1,
    presentation: { kind: 'branch_select', layout: 'overlay', numeric_base: 0 },
  } satisfies Prompt);
}

// Backend uses ally ID as option.id
export function adventurerParadiseAllyPickPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: ADVENTURER_PLAYER_ID,
    message: '【冒险者天堂】选择队友代为提炼：',
    choice_type: 'adventurer_paradise_pick',
    options: [
      { id: ALLY_PLAYER_ID, target_id: ALLY_PLAYER_ID, label: 'Ally A1', button_label: '选择' },
    ],
    min: 1, max: 1,
    presentation: { kind: 'target_picker', target_filter: 'allies_exclude_self', numeric_base: 0 },
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
    message: '【偷天换日】请选择效果：',
    choice_type: 'adventurer_steal_sky_mode',
    options: [
      { id: '0', label: '转移对方战绩区1红宝石到我方', button_label: '转移宝石' },
      { id: '1', label: '将我方战绩区全部蓝水晶转换成红宝石', button_label: '水晶转宝石' },
    ],
    min: 1, max: 1,
    presentation: { kind: 'branch_select', layout: 'overlay', numeric_base: 0 },
  } satisfies Prompt);
}

// Scenario builders for removed exploratory branches; retained only for explicit fixture coverage.
export function stealSkyChangeDayEnergyTargetPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: ADVENTURER_PLAYER_ID,
    message: '【偷天换日】请选择获得能量的队友：',
    choice_type: 'adv_steal_sky_change_day_energy_target',
    options: [
      { id: ALLY_PLAYER_ID, target_id: ALLY_PLAYER_ID, label: 'Ally A1', button_label: '选择' },
    ],
    min: 1, max: 1,
    presentation: { kind: 'target_picker', target_filter: 'custom', numeric_base: 0 },
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
      { id: ALLY_PLAYER_ID, target_id: ALLY_PLAYER_ID, label: 'Ally A1', button_label: '选择' },
    ],
    min: 1, max: 1,
    presentation: { kind: 'target_picker', target_filter: 'custom', numeric_base: 0 },
  } satisfies Prompt);
}

export function stealSkyChangeDayMyCardPrompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_cards',
    player_id: ADVENTURER_PLAYER_ID,
    message: '【偷天换日】请选择你要给出的手牌：',
    choice_type: 'adv_steal_sky_change_day_my_card',
    options: [
      { id: 'adv-attack-1', label: '冒险斩', button_label: '选择' },
      { id: 'adv-attack-2', label: '探索斩', button_label: '选择' },
      { id: 'adv-magic-1', label: '火球', button_label: '选择' },
    ],
    min: 1, max: 1,
    presentation: { kind: 'card_picker', card_source: 'hand', card_filter: 'option_limited', numeric_base: 0 },
  } satisfies Prompt);
}
