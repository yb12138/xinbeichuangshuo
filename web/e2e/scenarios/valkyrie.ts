// ============================================================
// Valkyrie (女武神) Protocol Harness Scenarios
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

export const VALKYRIE_PLAYER_ID = 'valkyrie_player';
export const ENEMY_PLAYER_ID = 'enemy_1';
export const ALLY_PLAYER_ID = 'ally_1';

export const VALKYRIE_HOLY_PURSUIT_ID = 'valkyrie_holy_pursuit';
export const VALKYRIE_ORDER_MARK_ID = 'valkyrie_order_mark';
export const VALKYRIE_PEACE_WALKER_ID = 'valkyrie_peace_walker';
export const VALKYRIE_MARTIAL_GOD_LIGHT_ID = 'valkyrie_martial_god_light';
export const VALKYRIE_HEROIC_SUMMON_ID = 'valkyrie_heroic_summon';

const valkyrieCharacter = characterView({
  id: 'valkyrie',
  name: '女武神',
  title: '咏',
  faction: '咏',
  skills: [
    {
      id: VALKYRIE_HOLY_PURSUIT_ID,
      title: '神圣追击',
      description: '（行动结束时发动）移除自己2点［治疗］，额外进行一次攻击行动。',
      type: 3, // 响应
      min_targets: 0, max_targets: 0, target_type: 0,
    },
    {
      id: VALKYRIE_ORDER_MARK_ID,
      title: '秩序之印',
      description: '作为攻击行动时，你可以额外摸1张牌。若如此做，该次攻击的伤害-1，且本次攻击行动结束时你+1［治疗］。',
      type: 2, // 法术
      min_targets: 0, max_targets: 0, target_type: 0,
    },
    {
      id: VALKYRIE_PEACE_WALKER_ID,
      title: '和平行者',
      description: '（英灵形态中）回合开始时，你可以对自己造成1点法术伤害③，令一名队友获得1［治疗］。',
      type: 3, // 响应（被动形态）
      min_targets: 0, max_targets: 0, target_type: 0,
    },
    {
      id: VALKYRIE_MARTIAL_GOD_LIGHT_ID,
      title: '军威神光',
      description: '（回合开始时发动）你选择：Ⅰ、摸2张牌；Ⅱ、对一名对手造成1点法术伤害③。',
      type: 3, // 响应
      min_targets: 0, max_targets: 0, target_type: 0,
    },
    {
      id: VALKYRIE_HEROIC_SUMMON_ID,
      title: '英灵召唤',
      description: '（命中后发动）［水晶］弃1张法术牌［展示］，对一名对手造成2点法术伤害③，并给自己或一名队友+1［治疗］。',
      type: 3, // 响应(大招)
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

const defaultCharacters = [valkyrieCharacter, enemyCharacter, allyCharacter];

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function valkyrieHand(): Card[] {
  return [
    card({ id: 'valk-attack-1', name: '圣枪', type: 'Attack', element: 'Light' }),
    card({ id: 'valk-attack-2', name: '光刃', type: 'Attack', element: 'Light' }),
    card({ id: 'valk-magic-1', name: '圣光', type: 'Magic', element: 'Light' }),
    card({ id: 'valk-magic-2', name: '神圣', type: 'Magic', element: 'Light' }),
    card({ id: 'valk-water-magic', name: '治愈', type: 'Magic', element: 'Water' }),
  ];
}

function valkyrieAvailableSkill(skill: Partial<AvailableSkill> & { id: string; title: string }): AvailableSkill {
  return availableSkill({
    description: '',
    target_type: 0,
    min_targets: 0, max_targets: 0,
    ...skill,
  });
}

// Export for use in tests that need to test response_skills trigger mechanism
export { valkyrieAvailableSkill };

// ---------------------------------------------------------------------------
// Scenario Factory
// ---------------------------------------------------------------------------

export function valkyrieScenario(options: {
  hand?: Card[];
  availableSkills?: AvailableSkill[];
  crystal?: number;
  gem?: number;
  turnStage?: string;
  heal?: number;
  buffs?: { id: string; name: string; duration: number; value: number; source_id: string }[];
} = {}): ProtocolHarnessScenario {
  const hand = options.hand ?? valkyrieHand();
  const players = [
    playerView({
      id: VALKYRIE_PLAYER_ID,
      name: 'E2E Valkyrie',
      camp: 'Red',
      role: 'valkyrie',
      hand,
      hand_count: hand.length,
      crystal: options.crystal ?? 0,
      gem: options.gem ?? 0,
      heal: options.heal ?? 2,
      max_heal: 4,
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
    myPlayerId: VALKYRIE_PLAYER_ID,
    myPlayerName: 'E2E Valkyrie',
    characters: defaultCharacters,
    players: [
      playerInfo({ id: VALKYRIE_PLAYER_ID, name: 'E2E Valkyrie', camp: 'Red', char_role: 'valkyrie', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy E1', camp: 'Blue', char_role: 'enemy_char' }),
      playerInfo({ id: ALLY_PLAYER_ID, name: 'Ally A1', camp: 'Red', char_role: 'ally_char' }),
    ],
    initialState: syncState({
      turn_player_id: VALKYRIE_PLAYER_ID,
      turn_stage: options.turnStage ?? 'ActionExecution',
      available_skills: options.availableSkills ?? [],
      characters: defaultCharacters,
      players,
    }),
  };
}

// ============================================================
// Holy Pursuit (神圣追击) - 响应技能
// 后端通过 response_skills 自动触发，无需 choice_type
// ============================================================

export function holyPursuitScenario(options: {
  heal?: number;
} = {}): ProtocolHarnessScenario {
  return valkyrieScenario({
    heal: options.heal ?? 2,
    // 后端会设置 response_skills 触发确认弹框
  });
}

// ============================================================
// Order Mark (秩序之印) - 法术技能
// ============================================================

export function orderMarkScenario(): ProtocolHarnessScenario {
  return valkyrieScenario();
}

// ============================================================
// Peace Walker (和平行者) - 响应技能(被动形态)
// 后端通过 response_skills 自动触发，目标通过 min_targets 处理
// ============================================================

export function peaceWalkerScenario(): ProtocolHarnessScenario {
  return valkyrieScenario({
    buffs: [{ id: 'heroic_form', name: '英灵形态', duration: 0, value: 0, source_id: VALKYRIE_HEROIC_SUMMON_ID }],
    turnStage: 'TurnStart',
    // 后端会设置 response_skills + min_targets 触发确认和目标选择
  });
}

// ============================================================
// Martial God Light (军威神光) - 响应技能
// ============================================================

export function martialGodLightScenario(): ProtocolHarnessScenario {
  return valkyrieScenario({
    turnStage: 'TurnStart',
  });
}

export function martialGodLightBranchPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: VALKYRIE_PLAYER_ID,
    message: '【军威神光】回合开始选择：',
    choice_type: 'valkyrie_military_glory_mode',
    options: [
      { id: 'draw', label: '摸2张牌' },
      { id: 'damage', label: '对对手造成1点法术伤害' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}

export function martialGodLightTargetPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: VALKYRIE_PLAYER_ID,
    message: '【军威神光】请选择一名对手造成1点法术伤害：',
    choice_type: 'valkyrie_military_glory_target',
    options: [
      { id: ENEMY_PLAYER_ID, label: 'Enemy E1' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}

// ============================================================
// Heroic Summon (英灵召唤) - 响应技能(大招)
// 后端通过 response_skills 自动触发，目标通过 min_targets 处理
// 弃牌通过后端 choice_type: valkyrie_heroic_discard_card
// ============================================================

export function heroicSummonScenario(): ProtocolHarnessScenario {
  return valkyrieScenario({
    crystal: 1,
    heal: 0,
    // 后端会设置 response_skills 触发确认弹框
  });
}

export function heroicSummonDiscardPrompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_cards',
    player_id: VALKYRIE_PLAYER_ID,
    message: '【英灵召唤】请选择弃1张法术牌［展示］：',
    choice_type: 'valkyrie_heroic_discard_card',
    options: [
      { id: 'valk-magic-1', label: '圣光（法术）' },
      { id: 'valk-magic-2', label: '神圣（法术）' },
      { id: 'valk-water-magic', label: '治愈（法术）' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}