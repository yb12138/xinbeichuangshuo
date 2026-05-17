// ============================================================
// Saintess (圣女) Protocol Harness Scenarios
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

export const SAINTESS_PLAYER_ID = 'saintess_player';
export const ENEMY_PLAYER_ID = 'enemy_1';
export const ALLY_PLAYER_ID = 'ally_1';

export const SAINTESS_HEALING_LIGHT_ID = 'saintess_healing_light'; // 独有
export const SAINTESS_HEAL_SKILL_ID = 'saintess_heal_skill'; // 独有
export const SAINTESS_HOLY_HEAL_ID = 'saintess_holy_heal'; // 大招
export const SAINTESS_MERCY_ID = 'saintess_mercy'; // 启动大招

const saintessCharacter = characterView({
  id: 'saintess',
  name: '圣女',
  title: '圣',
  faction: '圣',
  skills: [
    {
      id: 'saintess_frost_prayer',
      title: '冰霜祷言',
      description: '（每当你使用水系牌或圣光时发动）目标角色+1［治疗］。',
      type: 0, // 被动
      min_targets: 1, max_targets: 1, target_type: 0,
    },
    {
      id: SAINTESS_HEALING_LIGHT_ID,
      title: '治愈之光',
      description: '指定最多3名角色各+1［治疗］。',
      type: 2, // 法术(独有)
      min_targets: 1, max_targets: 3, target_type: 0,
    },
    {
      id: SAINTESS_HEAL_SKILL_ID,
      title: '治疗术',
      description: '目标角色+2［治疗］。',
      type: 2, // 法术(独有)
      min_targets: 1, max_targets: 1, target_type: 0,
    },
    {
      id: SAINTESS_HOLY_HEAL_ID,
      title: '圣疗',
      description: '［回合限定］［水晶］任意分配3点［治疗］给1~3名角色，额外+1［攻击行动］或［法术行动］。',
      type: 2, // 法术(大招)
      min_targets: 0, max_targets: 0, target_type: 0,
    },
    {
      id: SAINTESS_MERCY_ID,
      title: '怜悯',
      description: '［持续］［宝石］［横置］，你的手牌上限恒定为7［恒定］，你+1［水晶］。',
      type: 1, // 启动(大招)
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

const defaultCharacters = [saintessCharacter, enemyCharacter, allyCharacter];

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function saintessHand(): Card[] {
  return [
    card({ id: 'saintess-water-atk1', name: '水刃', type: 'Attack', element: 'Water' }),
    card({ id: 'saintess-water-atk2', name: '寒冰斩', type: 'Attack', element: 'Water' }),
    card({ id: 'saintess-water-magic', name: '冰冻', type: 'Magic', element: 'Water' }),
    card({ id: 'saintess-holy-light', name: '圣光', type: 'Magic', element: 'Light' }),
    card({ id: 'saintess-fire-magic', name: '火球', type: 'Magic', element: 'Fire' }),
    card({ id: SAINTESS_HEALING_LIGHT_ID, name: '治愈之光', type: 'Magic', element: 'Light' }),
    card({ id: SAINTESS_HEAL_SKILL_ID, name: '治疗术', type: 'Magic', element: 'Light' }),
  ];
}

function saintessAvailableSkill(skill: Partial<AvailableSkill> & { id: string; title: string }): AvailableSkill {
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

export function saintessScenario(options: {
  hand?: Card[];
  availableSkills?: AvailableSkill[];
  crystal?: number;
  gem?: number;
  turnStage?: string;
  heal?: number;
} = {}): ProtocolHarnessScenario {
  const hand = options.hand ?? saintessHand();
  const players = [
    playerView({
      id: SAINTESS_PLAYER_ID,
      name: 'E2E Saintess',
      camp: 'Red',
      role: 'saintess',
      hand,
      hand_count: hand.length,
      crystal: options.crystal ?? 0,
      gem: options.gem ?? 0,
      heal: options.heal ?? 0,
      max_heal: 4,
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
    myPlayerId: SAINTESS_PLAYER_ID,
    myPlayerName: 'E2E Saintess',
    characters: defaultCharacters,
    players: [
      playerInfo({ id: SAINTESS_PLAYER_ID, name: 'E2E Saintess', camp: 'Red', char_role: 'saintess', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy E1', camp: 'Blue', char_role: 'enemy_char' }),
      playerInfo({ id: ALLY_PLAYER_ID, name: 'Ally A1', camp: 'Red', char_role: 'ally_char' }),
    ],
    initialState: syncState({
      turn_player_id: SAINTESS_PLAYER_ID,
      turn_stage: options.turnStage ?? 'ActionExecution',
      available_skills: options.availableSkills ?? [],
      characters: defaultCharacters,
      players,
    }),
  };
}

// ============================================================
// Frost Prayer (冰霜祷言) - 被动技能
// ============================================================

export function frostPrayerScenario(): ProtocolHarnessScenario {
  return saintessScenario();
}

export function frostPrayerTargetPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: SAINTESS_PLAYER_ID,
    message: '【冰霜祷言】使用水系牌或圣光，请选择一名目标角色+1［治疗］：',
    choice_type: 'frost_prayer_target',
    options: [
      { id: SAINTESS_PLAYER_ID, label: '自己' },
      { id: ENEMY_PLAYER_ID, label: 'Enemy E1' },
      { id: ALLY_PLAYER_ID, label: 'Ally A1' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}

// ============================================================
// Healing Light (治愈之光) - 独有法术技能
// ============================================================

export function healingLightScenario(): ProtocolHarnessScenario {
  return saintessScenario({
    availableSkills: [
      saintessAvailableSkill({
        id: SAINTESS_HEALING_LIGHT_ID, title: '治愈之光',
        min_targets: 1, max_targets: 3, target_type: 0,
      }),
    ],
  });
}

export function healingLightMultiTargetPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: SAINTESS_PLAYER_ID,
    message: '【治愈之光】请选择1~3名目标角色，各+1［治疗］：',
    choice_type: 'saintess_healing_light_targets',
    options: [
      { id: SAINTESS_PLAYER_ID, label: '自己' },
      { id: ENEMY_PLAYER_ID, label: 'Enemy E1' },
      { id: ALLY_PLAYER_ID, label: 'Ally A1' },
    ],
    min: 1, max: 3,
  } satisfies Prompt);
}

// ============================================================
// Heal Skill (治疗术) - 独有法术技能
// ============================================================

export function healSkillScenario(): ProtocolHarnessScenario {
  return saintessScenario({
    availableSkills: [
      saintessAvailableSkill({
        id: SAINTESS_HEAL_SKILL_ID, title: '治疗术',
        min_targets: 1, max_targets: 1, target_type: 0,
      }),
    ],
  });
}

export function healSkillTargetPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: SAINTESS_PLAYER_ID,
    message: '【治疗术】请选择一名目标角色+2［治疗］：',
    choice_type: 'saintess_heal_skill_target',
    options: [
      { id: SAINTESS_PLAYER_ID, label: '自己' },
      { id: ENEMY_PLAYER_ID, label: 'Enemy E1' },
      { id: ALLY_PLAYER_ID, label: 'Ally A1' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}

// ============================================================
// Holy Heal (圣疗) - 法术技能(大招)
// ============================================================

export function holyHealScenario(): ProtocolHarnessScenario {
  return saintessScenario({
    crystal: 1,
    availableSkills: [
      saintessAvailableSkill({
        id: SAINTESS_HOLY_HEAL_ID, title: '圣疗',
      }),
    ],
  });
}

export function holyHealDistributePrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: SAINTESS_PLAYER_ID,
    message: '【圣疗】请分配3点治疗给1~3名角色（使用分配器）：',
    choice_type: 'saint_heal_allocate',
    options: [
      { id: '1_player', label: '分配给1名角色（+3）' },
      { id: '2_players', label: '分配给2名角色（各+1/+2）' },
      { id: '3_players', label: '分配给3名角色（各+1）' },
    ],
    min: 1, max: 1,
    presentation: { kind: 'numeric', numeric_base: 0 },
  } satisfies Prompt);
}

export function holyHealBranchPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: SAINTESS_PLAYER_ID,
    message: '【圣疗】请选择额外获得的行动：',
    choice_type: 'saintess_holy_heal_branch',
    options: [
      { id: 'attack', label: '额外+1［攻击行动］' },
      { id: 'magic', label: '额外+1［法术行动］' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}

// ============================================================
// Mercy (怜悯) - 启动技能(大招)
// ============================================================

export function mercyScenario(): ProtocolHarnessScenario {
  return saintessScenario({
    gem: 1,
    turnStage: 'StartupPhase',
    availableSkills: [
      saintessAvailableSkill({
        id: SAINTESS_MERCY_ID, title: '怜悯',
      }),
    ],
  });
}

export function mercyPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: SAINTESS_PLAYER_ID,
    message: '【怜悯】是否消耗1个红宝石发动？手牌上限恒定为7，+1水晶。',
    choice_type: 'saintess_mercy',
    options: [
      { id: 'confirm', label: '发动' },
      { id: 'skip', label: '跳过' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}