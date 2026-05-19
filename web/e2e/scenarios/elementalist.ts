// ============================================================
// Elementalist (元素师) Protocol Harness Scenarios
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

export const ELEMENTALIST_PLAYER_ID = 'elementalist_player';
export const ENEMY_PLAYER_ID = 'enemy_1';
export const ALLY_PLAYER_ID = 'ally_1';

export const ELEMENTALIST_ELEMENT_IGNITE_ID = 'elementalist_element_ignite';
export const ELEMENTALIST_THUNDER_STRIKE_ID = 'elementalist_thunder_strike';
export const ELEMENTALIST_FREEZE_ID = 'elementalist_freeze';
export const ELEMENTALIST_WIND_BLADE_ID = 'elementalist_wind_blade';
export const ELEMENTALIST_METEOR_ID = 'elementalist_meteor';
export const ELEMENTALIST_FIREBALL_ID = 'elementalist_fireball';
export const ELEMENTALIST_MOONLIGHT_ID = 'elementalist_moonlight';

const elementalistCharacter = characterView({
  id: 'elementalist',
  name: '元素师',
  title: '咏',
  faction: '咏',
  skills: [
    {
      id: 'elementalist_element_absorb',
      title: '元素吸收',
      description: '（对目标角色造成法术伤害时发动③）你+1［元素］。',
      type: 3, // 响应
      min_targets: 0, max_targets: 0, target_type: 0,
      cost_gem: 0, cost_crystal: 0, cost_discards: 0,
    },
    {
      id: ELEMENTALIST_ELEMENT_IGNITE_ID,
      title: '元素点燃',
      description: '（移除3点［元素］）对目标角色造成2点法术伤害③，额外+1［法术行动］；不能和［元素吸收］同时发动。',
      type: 2, // 法术
      min_targets: 1, max_targets: 1, target_type: 3,
      cost_gem: 0, cost_crystal: 0, cost_discards: 0,
    },
    {
      id: ELEMENTALIST_THUNDER_STRIKE_ID,
      title: '雷击',
      description: '对目标角色造成1点法术伤害③，我方战绩区+1宝石，（若你额外弃1张雷系牌［展示］）本次法术伤害额外+1。',
      type: 2, // 法术(独有)
      min_targets: 1, max_targets: 1, target_type: 3,
      cost_gem: 0, cost_crystal: 0, cost_discards: 0,
    },
    {
      id: ELEMENTALIST_FREEZE_ID,
      title: '冰冻',
      description: '对目标角色造成1点法术伤害③，并指定1名角色+1［治疗］，（若你额外弃1张水系牌［展示］）本次法术伤害额外+1。',
      type: 2, // 法术(独有)
      min_targets: 2, max_targets: 2, target_type: 0,
      cost_gem: 0, cost_crystal: 0, cost_discards: 0,
    },
    {
      id: ELEMENTALIST_WIND_BLADE_ID,
      title: '风刃',
      description: '对目标角色造成1点法术伤害③，额外+1［攻击行动］，（若你额外弃1张风系牌［展示］）本次法术伤害额外+1。',
      type: 2, // 法术(独有)
      min_targets: 1, max_targets: 1, target_type: 3,
      cost_gem: 0, cost_crystal: 0, cost_discards: 0,
    },
    {
      id: ELEMENTALIST_METEOR_ID,
      title: '陨石',
      description: '对目标角色造成1点法术伤害③，额外+1［法术行动］，（若你额外弃1张地系牌［展示］）本次法术伤害额外+1。',
      type: 2, // 法术(独有)
      min_targets: 1, max_targets: 1, target_type: 3,
      cost_gem: 0, cost_crystal: 0, cost_discards: 0,
    },
    {
      id: ELEMENTALIST_FIREBALL_ID,
      title: '火球',
      description: '对目标角色造成2点法术伤害③，（若你额外弃1张火系牌［展示］）本次法术伤害额外+1。',
      type: 2, // 法术(独有)
      min_targets: 1, max_targets: 1, target_type: 3,
      cost_gem: 0, cost_crystal: 0, cost_discards: 0,
    },
    {
      id: ELEMENTALIST_MOONLIGHT_ID,
      title: '月光',
      description: '［宝石］对目标角色造成（X+1）点法术伤害③，X为你剩余的能量数。',
      type: 2, // 法术(大招)
      min_targets: 1, max_targets: 1, target_type: 3,
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

const defaultCharacters = [elementalistCharacter, enemyCharacter, allyCharacter];

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function elementalistHand(): Card[] {
  return [
    card({ id: 'el-thunder-atk', name: '雷击斩', type: 'Attack', element: 'Thunder' }),
    card({ id: 'el-thunder-magic', name: '雷电', type: 'Magic', element: 'Thunder' }),
    card({ id: 'el-water-magic', name: '冰冻', type: 'Magic', element: 'Water' }),
    card({ id: 'el-wind-magic', name: '风刃', type: 'Magic', element: 'Wind' }),
    card({ id: 'el-earth-magic', name: '地刺', type: 'Magic', element: 'Earth' }),
    card({ id: 'el-fire-magic', name: '火球', type: 'Magic', element: 'Fire' }),
  ];
}

function elementalistAvailableSkill(skill: Partial<AvailableSkill> & { id: string; title: string }): AvailableSkill {
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

export function elementalistScenario(options: {
  hand?: Card[];
  availableSkills?: AvailableSkill[];
  crystal?: number;
  gem?: number;
  turnStage?: string;
  energy?: number; // 能量数
  tokens?: Record<string, number>;
} = {}): ProtocolHarnessScenario {
  const hand = options.hand ?? elementalistHand();
  const players = [
    playerView({
      id: ELEMENTALIST_PLAYER_ID,
      name: 'E2E Elementalist',
      camp: 'Red',
      role: 'elementalist',
      hand,
      hand_count: hand.length,
      crystal: options.crystal ?? 0,
      gem: options.gem ?? 0,
      is_active: true,
      tokens: options.tokens ?? { element: options.energy ?? 3 },
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
    myPlayerId: ELEMENTALIST_PLAYER_ID,
    myPlayerName: 'E2E Elementalist',
    characters: defaultCharacters,
    players: [
      playerInfo({ id: ELEMENTALIST_PLAYER_ID, name: 'E2E Elementalist', camp: 'Red', char_role: 'elementalist', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy E1', camp: 'Blue', char_role: 'enemy_char' }),
      playerInfo({ id: ALLY_PLAYER_ID, name: 'Ally A1', camp: 'Red', char_role: 'ally_char' }),
    ],
    initialState: syncState({
      turn_player_id: ELEMENTALIST_PLAYER_ID,
      turn_stage: options.turnStage ?? 'ActionExecution',
      available_skills: options.availableSkills ?? [],
      characters: defaultCharacters,
      players,
    }),
  };
}

// ============================================================
// Element Ignite (元素点燃) - 法术技能
// ============================================================

export function elementIgniteScenario(options: {
  elementCount?: number;
} = {}): ProtocolHarnessScenario {
  return elementalistScenario({
    tokens: { element: options.elementCount ?? 3 },
    availableSkills: [
      elementalistAvailableSkill({
        id: ELEMENTALIST_ELEMENT_IGNITE_ID, title: '元素点燃',
        min_targets: 1, max_targets: 1, target_type: 3,
      }),
    ],
  });
}

export function elementIgniteTargetPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: ELEMENTALIST_PLAYER_ID,
    message: '【元素点燃】请选择一名目标角色，造成2点法术伤害：',
    choice_type: 'elementalist_element_ignite_target',
    presentation: { kind: 'target_picker', target_filter: 'custom', numeric_base: 0 },
    options: [
      { id: ENEMY_PLAYER_ID, label: 'Enemy E1', button_label: '选择' },
      { id: ALLY_PLAYER_ID, label: 'Ally A1', button_label: '选择' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}

// ============================================================
// Thunder Strike (雷击) - 独有法术技能
// ============================================================

export function thunderStrikeScenario(): ProtocolHarnessScenario {
  return elementalistScenario({
    availableSkills: [
      elementalistAvailableSkill({
        id: ELEMENTALIST_THUNDER_STRIKE_ID, title: '雷击',
        min_targets: 1, max_targets: 1, target_type: 3,
      }),
    ],
  });
}

export function thunderStrikeTargetPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: ELEMENTALIST_PLAYER_ID,
    message: '【雷击】请选择一名目标角色，造成1点法术伤害：',
    choice_type: 'elementalist_thunder_strike_target',
    presentation: { kind: 'target_picker', target_filter: 'custom', numeric_base: 0 },
    options: [
      { id: ENEMY_PLAYER_ID, label: 'Enemy E1', button_label: '选择' },
      { id: ALLY_PLAYER_ID, label: 'Ally A1', button_label: '选择' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}

export function thunderStrikeExtraDiscardPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: ELEMENTALIST_PLAYER_ID,
    message: '【雷击】是否额外弃1张雷系牌使伤害+1？',
    choice_type: 'elementalist_bonus_card',
    options: [
      { id: 'yes', label: '弃牌+1伤害', button_label: '弃牌+1' },
      { id: 'no', label: '不弃牌', button_label: '不弃牌' },
    ],
    presentation: { kind: 'branch_select', layout: 'overlay', numeric_base: 0 },
    min: 1, max: 1,
  } satisfies Prompt);
}

// ============================================================
// Freeze (冰冻) - 独有法术技能
// ============================================================

export function freezeScenario(): ProtocolHarnessScenario {
  return elementalistScenario({
    availableSkills: [
      elementalistAvailableSkill({
        id: ELEMENTALIST_FREEZE_ID, title: '冰冻',
        min_targets: 2, max_targets: 2, target_type: 0,
      }),
    ],
  });
}

export function freezeDamageTargetPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: ELEMENTALIST_PLAYER_ID,
    message: '【冰冻】请选择造成法术伤害的目标：',
    choice_type: 'elementalist_freeze_damage_target',
    presentation: { kind: 'target_picker', target_filter: 'custom', numeric_base: 0 },
    options: [
      { id: ENEMY_PLAYER_ID, label: 'Enemy E1', button_label: '选择' },
      { id: ALLY_PLAYER_ID, label: 'Ally A1', button_label: '选择' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}

export function freezeHealTargetPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: ELEMENTALIST_PLAYER_ID,
    message: '【冰冻】请选择治疗目标（可选择自己）：',
    choice_type: 'elementalist_freeze_heal_target',
    presentation: { kind: 'target_picker', target_filter: 'custom', numeric_base: 0 },
    options: [
      { id: ELEMENTALIST_PLAYER_ID, label: '自己', button_label: '选择' },
      { id: ENEMY_PLAYER_ID, label: 'Enemy E1', button_label: '选择' },
      { id: ALLY_PLAYER_ID, label: 'Ally A1', button_label: '选择' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}

export function freezeExtraDiscardPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: ELEMENTALIST_PLAYER_ID,
    message: '【冰冻】是否额外弃1张水系牌使伤害+1？',
    choice_type: 'elementalist_bonus_card',
    options: [
      { id: 'yes', label: '弃牌+1伤害', button_label: '弃牌+1' },
      { id: 'no', label: '不弃牌', button_label: '不弃牌' },
    ],
    presentation: { kind: 'branch_select', layout: 'overlay', numeric_base: 0 },
    min: 1, max: 1,
  } satisfies Prompt);
}

// ============================================================
// Wind Blade (风刃) - 独有法术技能
// ============================================================

export function windBladeScenario(): ProtocolHarnessScenario {
  return elementalistScenario({
    availableSkills: [
      elementalistAvailableSkill({
        id: ELEMENTALIST_WIND_BLADE_ID, title: '风刃',
        min_targets: 1, max_targets: 1, target_type: 3,
      }),
    ],
  });
}

export function windBladeExtraDiscardPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: ELEMENTALIST_PLAYER_ID,
    message: '【风刃】是否额外弃1张风系牌使伤害+1？',
    choice_type: 'elementalist_bonus_card',
    options: [
      { id: 'yes', label: '弃牌+1伤害', button_label: '弃牌+1' },
      { id: 'no', label: '不弃牌', button_label: '不弃牌' },
    ],
    presentation: { kind: 'branch_select', layout: 'overlay', numeric_base: 0 },
    min: 1, max: 1,
  } satisfies Prompt);
}

// ============================================================
// Meteor (陨石) - 独有法术技能
// ============================================================

export function meteorScenario(): ProtocolHarnessScenario {
  return elementalistScenario({
    availableSkills: [
      elementalistAvailableSkill({
        id: ELEMENTALIST_METEOR_ID, title: '陨石',
        min_targets: 1, max_targets: 1, target_type: 3,
      }),
    ],
  });
}

export function meteorExtraDiscardPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: ELEMENTALIST_PLAYER_ID,
    message: '【陨石】是否额外弃1张地系牌使伤害+1？',
    choice_type: 'elementalist_bonus_card',
    options: [
      { id: 'yes', label: '弃牌+1伤害', button_label: '弃牌+1' },
      { id: 'no', label: '不弃牌', button_label: '不弃牌' },
    ],
    presentation: { kind: 'branch_select', layout: 'overlay', numeric_base: 0 },
    min: 1, max: 1,
  } satisfies Prompt);
}

// ============================================================
// Fireball (火球) - 独有法术技能
// ============================================================

export function fireballScenario(): ProtocolHarnessScenario {
  return elementalistScenario({
    availableSkills: [
      elementalistAvailableSkill({
        id: ELEMENTALIST_FIREBALL_ID, title: '火球',
        min_targets: 1, max_targets: 1, target_type: 3,
      }),
    ],
  });
}

export function fireballExtraDiscardPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: ELEMENTALIST_PLAYER_ID,
    message: '【火球】是否额外弃1张火系牌使伤害+1？',
    choice_type: 'elementalist_bonus_card',
    options: [
      { id: 'yes', label: '弃牌+1伤害', button_label: '弃牌+1' },
      { id: 'no', label: '不弃牌', button_label: '不弃牌' },
    ],
    presentation: { kind: 'branch_select', layout: 'overlay', numeric_base: 0 },
    min: 1, max: 1,
  } satisfies Prompt);
}

// ============================================================
// Moonlight (月光) - 法术技能(大招)
// ============================================================

export function moonlightScenario(options: {
  energy?: number;
} = {}): ProtocolHarnessScenario {
  return elementalistScenario({
    gem: 1,
    tokens: { element: options.energy ?? 3 },
    availableSkills: [
      elementalistAvailableSkill({
        id: ELEMENTALIST_MOONLIGHT_ID, title: '月光',
        min_targets: 1, max_targets: 1, target_type: 3,
      }),
    ],
  });
}

export function moonlightTargetPrompt(x: number): WsMessage {
  const damage = x + 1;
  return requireActionMessage({
    type: 'confirm',
    player_id: ELEMENTALIST_PLAYER_ID,
    message: `【月光】请选择一名目标角色，造成${damage}点法术伤害：`,
    choice_type: 'elementalist_moonlight_target',
    presentation: { kind: 'target_picker', target_filter: 'custom', numeric_base: 0 },
    options: [
      { id: ENEMY_PLAYER_ID, label: 'Enemy E1', button_label: '选择' },
      { id: ALLY_PLAYER_ID, label: 'Ally A1', button_label: '选择' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}
