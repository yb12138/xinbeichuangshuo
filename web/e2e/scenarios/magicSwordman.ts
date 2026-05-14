// ============================================================
// MagicSwordman (魔剑士) Protocol Harness Scenarios
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

export const MAGIC_SWORDMAN_PLAYER_ID = 'magic_swordman_player';
export const ENEMY_PLAYER_ID = 'enemy_1';
export const ALLY_PLAYER_ID = 'ally_1';

export const MAGIC_SWORDMAN_ASURA_COMBO_ID = 'magic_swordman_asura_combo';
export const MAGIC_SWORDMAN_SHADOW_GATHER_ID = 'magic_swordman_shadow_gather';
export const MAGIC_SWORDMAN_SHADOW_METEOR_ID = 'magic_swordman_shadow_meteor';
export const MAGIC_SWORDMAN_UNDERWORLD_TREMOR_ID = 'magic_swordman_underworld_tremor';

const magicSwordmanCharacter = characterView({
  id: 'magic_swordman',
  name: '魔剑士',
  title: '咏',
  faction: '咏',
  skills: [
    {
      id: MAGIC_SWORDMAN_ASURA_COMBO_ID,
      title: '修罗连斩',
      description: '（攻击结束时发动）若你本次攻击伤害≥2，你可以弃1张火系牌［展示］，进行一次火系攻击。',
      type: 3, // 响应
      min_targets: 0, max_targets: 0, target_type: 0,
    },
    {
      id: MAGIC_SWORDMAN_SHADOW_GATHER_ID,
      title: '暗影凝聚',
      description: '对自己造成1点法术伤害③，横置，获得暗影形态。',
      type: 1, // 启动
      min_targets: 0, max_targets: 0, target_type: 0,
    },
    {
      id: MAGIC_SWORDMAN_SHADOW_METEOR_ID,
      title: '暗影流星',
      description: '（暗影形态中）弃1张法术牌［展示］，对一名对手造成2点法术伤害③，脱离暗影形态。',
      type: 2, // 法术
      min_targets: 1, max_targets: 1, target_type: 2,
    },
    {
      id: MAGIC_SWORDMAN_UNDERWORLD_TREMOR_ID,
      title: '黄泉震颤',
      description: '［宝石］本次攻击无法应战。若命中，你对目标造成1点法术伤害③，并从其手中抽取一张牌加入自己手牌。',
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

const defaultCharacters = [magicSwordmanCharacter, enemyCharacter, allyCharacter];

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function magicSwordmanHand(): Card[] {
  return [
    card({ id: 'ms-fire-attack-1', name: '火焰斩', type: 'Attack', element: 'Fire' }),
    card({ id: 'ms-fire-attack-2', name: '炎刃', type: 'Attack', element: 'Fire' }),
    card({ id: 'ms-fire-magic', name: '火球', type: 'Magic', element: 'Fire' }),
    card({ id: 'ms-water-magic', name: '冰冻', type: 'Magic', element: 'Water' }),
    card({ id: 'ms-dark-magic', name: '暗影', type: 'Magic', element: 'Dark' }),
  ];
}

function magicSwordmanAvailableSkill(skill: Partial<AvailableSkill> & { id: string; title: string }): AvailableSkill {
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

export function magicSwordmanScenario(options: {
  hand?: Card[];
  availableSkills?: AvailableSkill[];
  crystal?: number;
  gem?: number;
  turnStage?: string;
  buffs?: { id: string; name: string; duration: number; value: number; source_id: string }[];
} = {}): ProtocolHarnessScenario {
  const hand = options.hand ?? magicSwordmanHand();
  const players = [
    playerView({
      id: MAGIC_SWORDMAN_PLAYER_ID,
      name: 'E2E MagicSwordman',
      camp: 'Red',
      role: 'magic_swordman',
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
    myPlayerId: MAGIC_SWORDMAN_PLAYER_ID,
    myPlayerName: 'E2E MagicSwordman',
    characters: defaultCharacters,
    players: [
      playerInfo({ id: MAGIC_SWORDMAN_PLAYER_ID, name: 'E2E MagicSwordman', camp: 'Red', char_role: 'magic_swordman', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy E1', camp: 'Blue', char_role: 'enemy_char' }),
      playerInfo({ id: ALLY_PLAYER_ID, name: 'Ally A1', camp: 'Red', char_role: 'ally_char' }),
    ],
    initialState: syncState({
      turn_player_id: MAGIC_SWORDMAN_PLAYER_ID,
      turn_stage: options.turnStage ?? 'ActionExecution',
      available_skills: options.availableSkills ?? [],
      characters: defaultCharacters,
      players,
    }),
  };
}

// ============================================================
// Asura Combo (修罗连斩) - 响应技能
// ============================================================

export function asuraComboScenario(): ProtocolHarnessScenario {
  return magicSwordmanScenario();
}

export function asuraComboPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: MAGIC_SWORDMAN_PLAYER_ID,
    message: '【修罗连斩】攻击伤害≥2，弃1张火系牌进行火系攻击？',
    choice_type: 'ms_asura_combo',
    options: [
      { id: 'confirm', label: '发动' },
      { id: 'skip', label: '跳过' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}

export function asuraComboDiscardPrompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_cards',
    player_id: MAGIC_SWORDMAN_PLAYER_ID,
    message: '【修罗连斩】请选择弃1张火系牌［展示］：',
    choice_type: 'ms_asura_combo_discard',
    options: [
      { id: 'ms-fire-attack-1', label: '火焰斩（火系）' },
      { id: 'ms-fire-attack-2', label: '炎刃（火系）' },
      { id: 'ms-fire-magic', label: '火球（火系）' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}

// ============================================================
// Shadow Gather (暗影凝聚) - 启动技能
// ============================================================

export function shadowGatherScenario(): ProtocolHarnessScenario {
  return magicSwordmanScenario({
    availableSkills: [
      magicSwordmanAvailableSkill({
        id: MAGIC_SWORDMAN_SHADOW_GATHER_ID, title: '暗影凝聚',
      }),
    ],
  });
}

export function shadowGatherPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: MAGIC_SWORDMAN_PLAYER_ID,
    message: '【暗影凝聚】对自己造成1点法术伤害，横置进入暗影形态？',
    choice_type: 'ms_shadow_gather',
    options: [
      { id: 'confirm', label: '发动' },
      { id: 'skip', label: '跳过' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}

// ============================================================
// Shadow Meteor (暗影流星) - 法术技能
// ============================================================

export function shadowMeteorScenario(): ProtocolHarnessScenario {
  return magicSwordmanScenario({
    buffs: [{ id: 'shadow_form', name: '暗影形态', duration: 0, value: 0, source_id: MAGIC_SWORDMAN_SHADOW_GATHER_ID }],
    availableSkills: [
      magicSwordmanAvailableSkill({
        id: MAGIC_SWORDMAN_SHADOW_METEOR_ID, title: '暗影流星',
        min_targets: 1, max_targets: 1, target_type: 2,
      }),
    ],
  });
}

export function shadowMeteorDiscardPrompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_cards',
    player_id: MAGIC_SWORDMAN_PLAYER_ID,
    message: '【暗影流星】请选择弃1张法术牌［展示］：',
    choice_type: 'ms_shadow_meteor_discard',
    options: [
      { id: 'ms-fire-magic', label: '火球（法术）' },
      { id: 'ms-water-magic', label: '冰冻（法术）' },
      { id: 'ms-dark-magic', label: '暗影（法术）' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}

export function shadowMeteorTargetPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: MAGIC_SWORDMAN_PLAYER_ID,
    message: '【暗影流星】请选择一名对手造成2点法术伤害：',
    choice_type: 'ms_shadow_meteor_target',
    options: [
      { id: ENEMY_PLAYER_ID, label: 'Enemy E1' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}

// ============================================================
// Underworld Tremor (黄泉震颤) - 响应技能(大招)
// ============================================================

export function underworldTremorScenario(): ProtocolHarnessScenario {
  return magicSwordmanScenario({
    gem: 1,
  });
}

export function underworldTremorPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: MAGIC_SWORDMAN_PLAYER_ID,
    message: '【黄泉震颤］消耗宝石，本次攻击无法应战，命中后+1法术伤害并抽牌？',
    choice_type: 'ms_underworld_tremor',
    options: [
      { id: 'confirm', label: '发动' },
      { id: 'skip', label: '跳过' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}