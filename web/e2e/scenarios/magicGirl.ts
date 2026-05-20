// ============================================================
// MagicGirl (魔法少女) Protocol Harness Scenarios
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

export const MAGIC_GIRL_PLAYER_ID = 'magic_girl_player';
export const ENEMY_PLAYER_ID = 'enemy_1';
export const ENEMY2_PLAYER_ID = 'enemy_2';
export const ALLY_PLAYER_ID = 'ally_1';

export const MAGIC_GIRL_MAGIC_BULLET_CONTROL_ID = 'magic_girl_magic_bullet_control';
export const MAGIC_GIRL_MAGIC_BULLET_FUSION_ID = 'magic_girl_magic_bullet_fusion';
export const MAGIC_GIRL_MAGIC_EXPLOSION_ID = 'magic_girl_magic_explosion';
export const MAGIC_GIRL_DESTRUCTION_STORM_ID = 'magic_girl_destruction_storm';

const magicGirlCharacter = characterView({
  id: 'magic_girl',
  name: '魔法少女',
  title: '咏',
  faction: '咏',
  skills: [
    {
      id: MAGIC_GIRL_MAGIC_BULLET_CONTROL_ID,
      title: '魔弹掌控',
      description: '你主动使用魔弹时可以选择逆向传递。',
      type: 3,
      min_targets: 0, max_targets: 0, target_type: 0,
      cost_gem: 0, cost_crystal: 0, cost_discards: 0,
    },
    {
      id: MAGIC_GIRL_MAGIC_BULLET_FUSION_ID,
      title: '魔弹融合',
      description: '你的地系或火系牌可以当魔弹使用。',
      type: 3,
      min_targets: 0, max_targets: 0, target_type: 0,
      cost_gem: 0, cost_crystal: 0, cost_discards: 0,
    },
    {
      id: MAGIC_GIRL_MAGIC_EXPLOSION_ID,
      title: '魔爆冲击',
      description: '（弃1张法术牌［展示］）我方战绩区+1颗［宝石］。2名目标对手各弃1张法术牌［展示］，每有人不如此做，你对他造成2点法术伤害③，你弃1张牌。',
      type: 2,
      min_targets: 2, max_targets: 2, target_type: 2,
      cost_gem: 0, cost_crystal: 0, cost_discards: 0,
    },
    {
      id: MAGIC_GIRL_DESTRUCTION_STORM_ID,
      title: '毁灭风暴',
      description: '［宝石］对任2名目标对手各造成2点法术伤害③。',
      type: 2,
      min_targets: 2, max_targets: 2, target_type: 2,
      cost_gem: 0, cost_crystal: 0, cost_discards: 0,
    },
  ],
});

const enemyCharacter = characterView({
  id: 'enemy_char', name: '守卫', title: '测试目标', faction: '异端', skills: [],
});

const enemy2Character = characterView({
  id: 'enemy_char2', name: '守卫B', title: '测试目标2', faction: '异端', skills: [],
});

const allyCharacter = characterView({
  id: 'ally_char', name: '勇者', title: '测试队友', faction: '咏', skills: [],
});

const defaultCharacters = [magicGirlCharacter, enemyCharacter, enemy2Character, allyCharacter];

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function magicGirlHand(): Card[] {
  return [
    card({ id: 'mg-earth-atk', name: '地刺斩', type: 'Attack', element: 'Earth' }),
    card({ id: 'mg-earth-magic', name: '地刺', type: 'Magic', element: 'Earth' }),
    card({ id: 'mg-fire-atk', name: '火焰斩', type: 'Attack', element: 'Fire' }),
    card({ id: 'mg-fire-magic', name: '火球', type: 'Magic', element: 'Fire' }),
    card({ id: 'mg-magic-bullet', name: '魔弹', type: 'Magic', element: 'Light' }),
    card({ id: 'mg-water-magic', name: '冰冻', type: 'Magic', element: 'Water' }),
  ];
}

function magicGirlAvailableSkill(skill: Partial<AvailableSkill> & { id: string; title: string }): AvailableSkill {
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

export function magicGirlScenario(options: {
  hand?: Card[];
  availableSkills?: AvailableSkill[];
  crystal?: number;
  gem?: number;
  turnStage?: string;
} = {}): ProtocolHarnessScenario {
  const hand = options.hand ?? magicGirlHand();
  const players = [
    playerView({
      id: MAGIC_GIRL_PLAYER_ID,
      name: 'E2E MagicGirl',
      camp: 'Red',
      role: 'magic_girl',
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
      hand: [card({ id: 'en-card-1', name: '测试牌', type: 'Magic', element: 'Fire' })],
      hand_count: 1, max_hand: 6,
      heal: 0, max_heal: 4,
      is_active: false,
    }),
    playerView({
      id: ENEMY2_PLAYER_ID,
      name: 'Enemy E2',
      camp: 'Blue',
      role: 'enemy_char2',
      hand: [card({ id: 'en2-card-1', name: '测试牌', type: 'Magic', element: 'Thunder' })],
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
    myPlayerId: MAGIC_GIRL_PLAYER_ID,
    myPlayerName: 'E2E MagicGirl',
    characters: defaultCharacters,
    players: [
      playerInfo({ id: MAGIC_GIRL_PLAYER_ID, name: 'E2E MagicGirl', camp: 'Red', char_role: 'magic_girl', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy E1', camp: 'Blue', char_role: 'enemy_char' }),
      playerInfo({ id: ENEMY2_PLAYER_ID, name: 'Enemy E2', camp: 'Blue', char_role: 'enemy_char2' }),
      playerInfo({ id: ALLY_PLAYER_ID, name: 'Ally A1', camp: 'Red', char_role: 'ally_char' }),
    ],
    initialState: syncState({
      turn_player_id: MAGIC_GIRL_PLAYER_ID,
      turn_stage: options.turnStage ?? 'ActionExecution',
      available_skills: options.availableSkills ?? [],
      characters: defaultCharacters,
      players,
    }),
  };
}

// ============================================================
// Magic Bullet Control (魔弹掌控) - 响应技能
// ============================================================

export function magicBulletControlScenario(): ProtocolHarnessScenario {
  return magicGirlScenario();
}

export function magicBulletDirectionPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: MAGIC_GIRL_PLAYER_ID,
    message: '【魔弹掌控】请选择魔弹传递方向：',
    choice_type: 'mg_magic_bullet_direction',
    options: [
      { id: 'forward', label: '正向传递', button_label: '正向传递' },
      { id: 'reverse', label: '逆向传递', button_label: '逆向传递' },
    ],
    min: 1, max: 1,
    presentation: { kind: 'branch_select', layout: 'overlay', numeric_base: 0 },
  } satisfies Prompt);
}

// ============================================================
// Magic Bullet Fusion (魔弹融合) - 响应技能
// ============================================================

export function magicBulletFusionScenario(): ProtocolHarnessScenario {
  return magicGirlScenario({
    hand: [
      card({ id: 'mg-earth-atk', name: '地刺斩', type: 'Attack', element: 'Earth' }),
      card({ id: 'mg-fire-atk', name: '火焰斩', type: 'Attack', element: 'Fire' }),
      card({ id: 'mg-magic-bullet', name: '魔弹', type: 'Magic', element: 'Light' }),
      card({ id: 'mg-water-magic', name: '冰冻', type: 'Magic', element: 'Water' }),
    ],
  });
}

// ============================================================
// Magic Explosion (魔爆冲击) - 法术技能
// ============================================================

export function magicExplosionScenario(): ProtocolHarnessScenario {
  return magicGirlScenario({
    availableSkills: [
      magicGirlAvailableSkill({
        id: MAGIC_GIRL_MAGIC_EXPLOSION_ID, title: '魔爆冲击',
        cost_discards: 1, discard_type: 'Magic',
        min_targets: 2, max_targets: 2, target_type: 2,
      }),
    ],
  });
}

export function magicExplosionTargetPrompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_target',
    player_id: MAGIC_GIRL_PLAYER_ID,
    message: '【魔爆冲击】请选择2名目标对手：',
    choice_type: 'mg_magic_explosion_target',
    options: [
      { id: ENEMY_PLAYER_ID, target_id: ENEMY_PLAYER_ID, label: 'Enemy E1', button_label: '选择' },
      { id: ENEMY2_PLAYER_ID, target_id: ENEMY2_PLAYER_ID, label: 'Enemy E2', button_label: '选择' },
    ],
    min: 2, max: 2,
    presentation: { kind: 'target_picker', target_filter: 'enemies', multi_target: true, numeric_base: 0 },
  } satisfies Prompt);
}

// ============================================================
// Magic Explosion — Enemy discard (对手弃牌视角)
// ============================================================

// Enemy perspective scenario: the enemy player is "me"
export function magicExplosionEnemyDiscardScenario(): ProtocolHarnessScenario {
  const mgHand = magicGirlHand();
  const enemyHand = [card({ id: 'en-card-1', name: '测试牌', type: 'Magic', element: 'Fire' })];
  const enemy2Hand = [card({ id: 'en2-card-1', name: '测试牌', type: 'Magic', element: 'Thunder' })];
  const allyHand = [card({ id: 'al-card-1', name: '测试牌', type: 'Attack', element: 'Water' })];

  return {
    roomCode: 'MOCK',
    myPlayerId: ENEMY_PLAYER_ID,
    myPlayerName: 'Enemy E1',
    characters: defaultCharacters,
    players: [
      playerInfo({ id: MAGIC_GIRL_PLAYER_ID, name: 'E2E MagicGirl', camp: 'Red', char_role: 'magic_girl', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy E1', camp: 'Blue', char_role: 'enemy_char' }),
      playerInfo({ id: ENEMY2_PLAYER_ID, name: 'Enemy E2', camp: 'Blue', char_role: 'enemy_char2' }),
      playerInfo({ id: ALLY_PLAYER_ID, name: 'Ally A1', camp: 'Red', char_role: 'ally_char' }),
    ],
    initialState: syncState({
      turn_player_id: MAGIC_GIRL_PLAYER_ID,
      turn_stage: 'ActionExecution',
      available_skills: [],
      characters: defaultCharacters,
      players: [
        playerView({
          id: MAGIC_GIRL_PLAYER_ID,
          name: 'E2E MagicGirl',
          camp: 'Red',
          role: 'magic_girl',
          hand: mgHand,
          hand_count: mgHand.length,
          is_active: true,
        }),
        playerView({
          id: ENEMY_PLAYER_ID,
          name: 'Enemy E1',
          camp: 'Blue',
          role: 'enemy_char',
          hand: enemyHand,
          hand_count: 1, max_hand: 6,
          heal: 0, max_heal: 4,
          is_active: false,
        }),
        playerView({
          id: ENEMY2_PLAYER_ID,
          name: 'Enemy E2',
          camp: 'Blue',
          role: 'enemy_char2',
          hand: enemy2Hand,
          hand_count: 1, max_hand: 6,
          heal: 0, max_heal: 4,
          is_active: false,
        }),
        playerView({
          id: ALLY_PLAYER_ID,
          name: 'Ally A1',
          camp: 'Red',
          role: 'ally_char',
          hand: allyHand,
          hand_count: 1, max_hand: 6,
          heal: 0, max_heal: 4,
          is_active: false,
        }),
      ],
    }),
  };
}

// Enemy E2 perspective scenario
export function magicExplosionEnemy2DiscardScenario(): ProtocolHarnessScenario {
  const mgHand = magicGirlHand();
  const enemyHand = [card({ id: 'en-card-1', name: '测试牌', type: 'Magic', element: 'Fire' })];
  const enemy2Hand = [card({ id: 'en2-card-1', name: '测试牌', type: 'Magic', element: 'Thunder' })];
  const allyHand = [card({ id: 'al-card-1', name: '测试牌', type: 'Attack', element: 'Water' })];

  return {
    roomCode: 'MOCK',
    myPlayerId: ENEMY2_PLAYER_ID,
    myPlayerName: 'Enemy E2',
    characters: defaultCharacters,
    players: [
      playerInfo({ id: MAGIC_GIRL_PLAYER_ID, name: 'E2E MagicGirl', camp: 'Red', char_role: 'magic_girl', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy E1', camp: 'Blue', char_role: 'enemy_char' }),
      playerInfo({ id: ENEMY2_PLAYER_ID, name: 'Enemy E2', camp: 'Blue', char_role: 'enemy_char2' }),
      playerInfo({ id: ALLY_PLAYER_ID, name: 'Ally A1', camp: 'Red', char_role: 'ally_char' }),
    ],
    initialState: syncState({
      turn_player_id: MAGIC_GIRL_PLAYER_ID,
      turn_stage: 'ActionExecution',
      available_skills: [],
      characters: defaultCharacters,
      players: [
        playerView({
          id: MAGIC_GIRL_PLAYER_ID,
          name: 'E2E MagicGirl',
          camp: 'Red',
          role: 'magic_girl',
          hand: mgHand,
          hand_count: mgHand.length,
          is_active: true,
        }),
        playerView({
          id: ENEMY_PLAYER_ID,
          name: 'Enemy E1',
          camp: 'Blue',
          role: 'enemy_char',
          hand: enemyHand,
          hand_count: 1, max_hand: 6,
          heal: 0, max_heal: 4,
          is_active: false,
        }),
        playerView({
          id: ENEMY2_PLAYER_ID,
          name: 'Enemy E2',
          camp: 'Blue',
          role: 'enemy_char2',
          hand: enemy2Hand,
          hand_count: 1, max_hand: 6,
          heal: 0, max_heal: 4,
          is_active: false,
        }),
        playerView({
          id: ALLY_PLAYER_ID,
          name: 'Ally A1',
          camp: 'Red',
          role: 'ally_char',
          hand: allyHand,
          hand_count: 1, max_hand: 6,
          heal: 0, max_heal: 4,
          is_active: false,
        }),
      ],
    }),
  };
}

export function magicExplosionEnemyDiscardPrompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_cards',
    player_id: ENEMY_PLAYER_ID,
    message: '【魔爆冲击】请弃1张法术牌［展示］，否则受到2点法术伤害：',
    choice_type: 'mg_magic_explosion_enemy_discard',
    options: [
      { id: 'en-card-1', label: '测试牌（法术）', button_label: '弃置', card_id: 'en-card-1' },
    ],
    min: 0, max: 1,
    presentation: { kind: 'card_picker', card_source: 'hand', card_filter: 'magic', has_decline: true, decline_index: 0, cancel_policy: 'decline', numeric_base: 0 },
  } satisfies Prompt);
}

export function magicExplosionEnemy2DiscardPrompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_cards',
    player_id: ENEMY2_PLAYER_ID,
    message: '【魔爆冲击】请弃1张法术牌［展示］，否则受到2点法术伤害：',
    choice_type: 'mg_magic_explosion_enemy2_discard',
    options: [
      { id: 'en2-card-1', label: '测试牌（法术）', button_label: '弃置', card_id: 'en2-card-1' },
    ],
    min: 0, max: 1,
    presentation: { kind: 'card_picker', card_source: 'hand', card_filter: 'magic', has_decline: true, decline_index: 0, cancel_policy: 'decline', numeric_base: 0 },
  } satisfies Prompt);
}

// ============================================================
// Destruction Storm (毁灭风暴) - 法术技能(大招)
// ============================================================

export function destructionStormScenario(): ProtocolHarnessScenario {
  return magicGirlScenario({
    gem: 1,
    availableSkills: [
      magicGirlAvailableSkill({
        id: MAGIC_GIRL_DESTRUCTION_STORM_ID, title: '毁灭风暴',
        min_targets: 2, max_targets: 2, target_type: 2,
      }),
    ],
  });
}

export function destructionStormTargetPrompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_target',
    player_id: MAGIC_GIRL_PLAYER_ID,
    message: '【毁灭风暴】请选择2名目标对手，各造成2点法术伤害：',
    choice_type: 'mg_destruction_storm_target',
    options: [
      { id: ENEMY_PLAYER_ID, target_id: ENEMY_PLAYER_ID, label: 'Enemy E1', button_label: '选择' },
      { id: ENEMY2_PLAYER_ID, target_id: ENEMY2_PLAYER_ID, label: 'Enemy E2', button_label: '选择' },
    ],
    min: 2, max: 2,
    presentation: { kind: 'target_picker', target_filter: 'enemies', multi_target: true, numeric_base: 0 },
  } satisfies Prompt);
}