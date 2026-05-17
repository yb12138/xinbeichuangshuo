// ============================================================
// BloodSwordSpirit (血色剑灵) Protocol Harness Scenarios
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

export const BLOOD_SWORD_SPIRIT_PLAYER_ID = 'blood_sword_spirit_player';
export const ENEMY_PLAYER_ID = 'enemy_1';
export const ALLY_PLAYER_ID = 'ally_1';

export const BLOOD_SWORD_SPIRIT_RED_FLASH_ID = 'blood_sword_spirit_red_flash';
export const BLOOD_SWORD_SPIRIT_BLOOD_DYE_ROSE_ID = 'blood_sword_spirit_blood_dye_rose';
export const BLOOD_SWORD_SPIRIT_BLOOD_BARRIER_ID = 'blood_sword_spirit_blood_barrier';
export const BLOOD_SWORD_SPIRIT_SCATTERING_DANCE_ID = 'blood_sword_spirit_scattering_dance';

const bloodSwordSpiritCharacter = characterView({
  id: 'blood_sword_spirit',
  name: '血色剑灵',
  title: '咏',
  faction: '咏',
  skills: [
    {
      id: BLOOD_SWORD_SPIRIT_RED_FLASH_ID,
      title: '赤色一闪',
      description: '（攻击前发动）若你有［鲜血］，你可以对自己造成1点法术伤害③，本次攻击伤害+1，可连续发动。',
      type: 3, // 响应
      min_targets: 0, max_targets: 0, target_type: 0,
    },
    {
      id: BLOOD_SWORD_SPIRIT_BLOOD_DYE_ROSE_ID,
      title: '血染蔷薇',
      description: '（［鲜血］≥3时发动）指定一名对手移除其所有［治疗］，并令一名队友获得等量的［治疗］。',
      type: 2, // 法术
      min_targets: 2, max_targets: 2, target_type: 0,
    },
    {
      id: BLOOD_SWORD_SPIRIT_BLOOD_BARRIER_ID,
      title: '血气屏障',
      description: '（受到法术伤害时发动）若你有［鲜血］，移除1［鲜血］，令该次伤害-1。',
      type: 3, // 响应
      min_targets: 0, max_targets: 0, target_type: 0,
    },
    {
      id: BLOOD_SWORD_SPIRIT_SCATTERING_DANCE_ID,
      title: '散华轮舞',
      description: '选择：Ⅰ、对一名对手造成2点法术伤害③，在其面前放置一个［血色庭院］；Ⅱ、移除自己1［治疗］，令一名队友+1［治疗］，在其面前放置一个［血色庭院］。',
      type: 1, // 启动
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

const defaultCharacters = [bloodSwordSpiritCharacter, enemyCharacter, allyCharacter];

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function bloodSwordSpiritHand(): Card[] {
  return [
    card({ id: 'bss-attack-1', name: '血刃', type: 'Attack', element: 'Fire' }),
    card({ id: 'bss-attack-2', name: '血斩', type: 'Attack', element: 'Fire' }),
    card({ id: 'bss-magic-1', name: '血术', type: 'Magic', element: 'Fire' }),
    card({ id: 'bss-magic-2', name: '血咒', type: 'Magic', element: 'Dark' }),
  ];
}

function bloodSwordSpiritAvailableSkill(skill: Partial<AvailableSkill> & { id: string; title: string }): AvailableSkill {
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

export function bloodSwordSpiritScenario(options: {
  hand?: Card[];
  availableSkills?: AvailableSkill[];
  crystal?: number;
  gem?: number;
  turnStage?: string;
  heal?: number;
  tokens?: Record<string, number>;
} = {}): ProtocolHarnessScenario {
  const hand = options.hand ?? bloodSwordSpiritHand();
  const players = [
    playerView({
      id: BLOOD_SWORD_SPIRIT_PLAYER_ID,
      name: 'E2E BloodSwordSpirit',
      camp: 'Red',
      role: 'blood_sword_spirit',
      hand,
      hand_count: hand.length,
      crystal: options.crystal ?? 0,
      gem: options.gem ?? 0,
      heal: options.heal ?? 0,
      max_heal: 4,
      is_active: true,
      tokens: options.tokens ?? { blood: 2 },
    }),
    playerView({
      id: ENEMY_PLAYER_ID,
      name: 'Enemy E1',
      camp: 'Blue',
      role: 'enemy_char',
      hand: [card({ id: 'en-card-1', name: '测试牌', type: 'Attack', element: 'Fire' })],
      hand_count: 1, max_hand: 6,
      heal: 2, max_heal: 4,
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
    myPlayerId: BLOOD_SWORD_SPIRIT_PLAYER_ID,
    myPlayerName: 'E2E BloodSwordSpirit',
    characters: defaultCharacters,
    players: [
      playerInfo({ id: BLOOD_SWORD_SPIRIT_PLAYER_ID, name: 'E2E BloodSwordSpirit', camp: 'Red', char_role: 'blood_sword_spirit', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy E1', camp: 'Blue', char_role: 'enemy_char' }),
      playerInfo({ id: ALLY_PLAYER_ID, name: 'Ally A1', camp: 'Red', char_role: 'ally_char' }),
    ],
    initialState: syncState({
      turn_player_id: BLOOD_SWORD_SPIRIT_PLAYER_ID,
      turn_stage: options.turnStage ?? 'ActionExecution',
      available_skills: options.availableSkills ?? [],
      characters: defaultCharacters,
      players,
    }),
  };
}

// ============================================================
// Red Flash (赤色一闪) - 响应技能
// ============================================================

export function redFlashScenario(options: {
  bloodCount?: number;
} = {}): ProtocolHarnessScenario {
  return bloodSwordSpiritScenario({
    tokens: { blood: options.bloodCount ?? 2 },
  });
}

export function redFlashPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: BLOOD_SWORD_SPIRIT_PLAYER_ID,
    message: '【赤色一闪】对自己造成1点法术伤害，本次攻击伤害+1？（可连续发动）',
    choice_type: 'css_red_flash',
    options: [
      { id: 'confirm', label: '发动' },
      { id: 'skip', label: '跳过' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}

// ============================================================
// Blood Dye Rose (血染蔷薇) - 法术技能
// ============================================================

export function bloodDyeRoseScenario(options: {
  bloodCount?: number;
} = {}): ProtocolHarnessScenario {
  return bloodSwordSpiritScenario({
    tokens: { blood: options.bloodCount ?? 3 },
    availableSkills: [
      bloodSwordSpiritAvailableSkill({
        id: BLOOD_SWORD_SPIRIT_BLOOD_DYE_ROSE_ID, title: '血染蔷薇',
        min_targets: 2, max_targets: 2, target_type: 0,
      }),
    ],
  });
}

export function bloodDyeRoseTargetPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: BLOOD_SWORD_SPIRIT_PLAYER_ID,
    message: '【血染蔷薇】请选择移除治疗的对手和获得治疗的队友：',
    choice_type: 'css_blood_rose_gain_heal_target',
    options: [
      { id: ENEMY_PLAYER_ID, label: 'Enemy E1（移除治疗）' },
      { id: ALLY_PLAYER_ID, label: 'Ally A1（获得治疗）' },
    ],
    min: 2, max: 2,
  } satisfies Prompt);
}

// ============================================================
// Blood Barrier (血气屏障) - 响应技能
// ============================================================

export function bloodBarrierScenario(options: {
  bloodCount?: number;
} = {}): ProtocolHarnessScenario {
  return bloodSwordSpiritScenario({
    tokens: { blood: options.bloodCount ?? 2 },
  });
}

export function bloodBarrierPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: BLOOD_SWORD_SPIRIT_PLAYER_ID,
    message: '【血气屏障】移除1鲜血，令法术伤害-1？',
    choice_type: 'css_blood_barrier',
    options: [
      { id: 'confirm', label: '发动' },
      { id: 'skip', label: '跳过' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}

// ============================================================
// Scattering Dance (散华轮舞) - 启动技能
// ============================================================

export function scatteringDanceScenario(): ProtocolHarnessScenario {
  return bloodSwordSpiritScenario({
    heal: 1,
    availableSkills: [
      bloodSwordSpiritAvailableSkill({
        id: BLOOD_SWORD_SPIRIT_SCATTERING_DANCE_ID, title: '散华轮舞',
      }),
    ],
  });
}

export function scatteringDanceBranchPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: BLOOD_SWORD_SPIRIT_PLAYER_ID,
    message: '【散华轮舞】请选择分支：',
    choice_type: 'css_dance_mode',
    options: [
      { id: 'damage', label: '对对手造成2点法术伤害，放置血色庭院' },
      { id: 'heal_transfer', label: '移除自己1治疗给队友，放置血色庭院' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}

export function scatteringDanceDamageTargetPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: BLOOD_SWORD_SPIRIT_PLAYER_ID,
    message: '【散华轮舞】请选择造成伤害的对手：',
    choice_type: 'css_dance_damage_target',
    options: [
      { id: ENEMY_PLAYER_ID, label: 'Enemy E1' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}

export function scatteringDanceHealTargetPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: BLOOD_SWORD_SPIRIT_PLAYER_ID,
    message: '【散华轮舞】请选择获得治疗的队友：',
    choice_type: 'css_dance_heal_target',
    options: [
      { id: ALLY_PLAYER_ID, label: 'Ally A1' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}