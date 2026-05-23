// ============================================================
// Magic Lancer (魔枪) Protocol Harness Scenarios
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

export const ML_PLAYER_ID = 'ml_player';
export const ENEMY_PLAYER_ID = 'enemy_1';
export const ALLY_PLAYER_ID = 'ally_1';

export const ML_DARK_RELEASE_SKILL_ID = 'ml_dark_release';
export const ML_PHANTOM_STARDUST_SKILL_ID = 'ml_phantom_stardust';
export const ML_DARK_BARRIER_SKILL_ID = 'ml_dark_barrier';
export const ML_FULLNESS_SKILL_ID = 'ml_fullness';
export const ML_BLACK_SPEAR_SKILL_ID = 'ml_black_spear';

export const ML_ATTACK_CARD_ID = 'ml-atk-fire';
export const ML_MAGIC_CARD_ID = 'ml-magic-light';
export const ML_THUNDER_CARD_ID = 'ml-thunder-atk';
export const ML_WATER_CARD_ID = 'ml-water-atk';

const magicLancerCharacter = characterView({
  id: 'magic_lancer',
  name: '魔枪',
  title: '幻',
  faction: '幻',
  skills: [
    {
      id: ML_DARK_RELEASE_SKILL_ID,
      title: '暗之解放',
      description: '转为幻影形态，手牌上限恒定为5；本回合下一次主动攻击伤害+1，且本回合不能发动漆黑之枪与充盈。',
      type: 1,
      min_targets: 0,
      max_targets: 0,
      target_type: 0,
      cost_gem: 0, cost_crystal: 0, cost_discards: 0,
    },
    {
      id: ML_PHANTOM_STARDUST_SKILL_ID,
      title: '幻影星尘',
      description: '仅幻影形态可发动：先对自己造成2点法术伤害，随后转正；若士气未下降，对目标角色造成2点法术伤害。',
      type: 1,
      min_targets: 1,
      max_targets: 1,
      target_type: 2,
      cost_gem: 0, cost_crystal: 0, cost_discards: 0,
    },
    {
      id: ML_DARK_BARRIER_SKILL_ID,
      title: '暗之障壁',
      description: '任何人对你造成伤害时可发动：弃X张法术牌或X张雷系牌。',
      type: 3,
      min_targets: 0,
      max_targets: 0,
      target_type: 0,
      cost_gem: 0, cost_crystal: 0, cost_discards: 0,
    },
    {
      id: ML_FULLNESS_SKILL_ID,
      title: '充盈',
      description: '弃1张法术牌或雷系牌：全体角色各弃1张牌，根据结果额外攻击伤害+1与额外+1次攻击行动。',
      type: 2,
      min_targets: 0,
      max_targets: 0,
      target_type: 0,
      cost_gem: 0, cost_crystal: 0, cost_discards: 0,
    },
    {
      id: ML_BLACK_SPEAR_SKILL_ID,
      title: '漆黑之枪',
      description: 'X水晶，仅幻影形态下主动攻击命中手牌为1或2的对手后，本次攻击伤害额外+(X+2)。',
      type: 3,
      min_targets: 0,
      max_targets: 0,
      target_type: 0,
      cost_gem: 0, cost_crystal: 0, cost_discards: 0,
    },
  ],
});

const enemyCharacter = characterView({
  id: 'enemy_char',
  name: '守卫',
  title: '测试目标',
  faction: '异端',
  skills: [],
});

const allyCharacter = characterView({
  id: 'ally_char',
  name: '勇者',
  title: '测试队友',
  faction: '星杯',
  skills: [],
});

const defaultCharacters = [magicLancerCharacter, enemyCharacter, allyCharacter];

function magicLancerHand(): Card[] {
  return [
    card({ id: ML_ATTACK_CARD_ID, name: '火焰斩', type: 'Attack', element: 'Fire', damage: 2 }),
    card({ id: ML_THUNDER_CARD_ID, name: '雷光斩', type: 'Attack', element: 'Thunder' }),
    card({ id: ML_MAGIC_CARD_ID, name: '圣光', type: 'Magic', element: 'Light' }),
    card({ id: ML_WATER_CARD_ID, name: '水涟斩', type: 'Attack', element: 'Water' }),
  ];
}

function magicLancerAvailableSkill(skill: Partial<AvailableSkill> & { id: string; title: string }): AvailableSkill {
  return availableSkill({
    description: '',
    target_type: 0,
    min_targets: 0,
    max_targets: 0,
    ...skill,
  });
}

// ============================================================
// Scenario Factories
// ============================================================

export function magicLancerScenario(options: {
  hand?: Card[];
  availableSkills?: AvailableSkill[];
  crystal?: number;
  gem?: number;
  turnStage?: string;
  form?: string[];
  tokens?: Record<string, number>;
} = {}): ProtocolHarnessScenario {
  const hand = options.hand ?? magicLancerHand();
  const players = [
    playerView({
      id: ML_PLAYER_ID,
      name: 'E2E Magic Lancer',
      camp: 'Red',
      role: 'magic_lancer',
      hand,
      hand_count: hand.length,
      crystal: options.crystal ?? 0,
      gem: options.gem ?? 0,
      is_active: true,
      form: options.form?.[0] ?? undefined,
      tokens: options.tokens ?? {},
    }),
    playerView({
      id: ENEMY_PLAYER_ID,
      name: 'Enemy E1',
      camp: 'Blue',
      role: 'enemy_char',
      hand: [card({ id: 'enemy-card-1', name: '测试牌', type: 'Attack', element: 'Fire' })],
      hand_count: 1,
      max_hand: 6,
      heal: 1,
      max_heal: 4,
      is_active: false,
    }),
    playerView({
      id: ALLY_PLAYER_ID,
      name: 'Ally A1',
      camp: 'Red',
      role: 'ally_char',
      hand: [card({ id: 'ally-card-1', name: '测试牌', type: 'Attack', element: 'Water' })],
      hand_count: 1,
      max_hand: 6,
      heal: 0,
      max_heal: 4,
      is_active: false,
    }),
  ];

  return {
    roomCode: 'MOCK',
    myPlayerId: ML_PLAYER_ID,
    myPlayerName: 'E2E Magic Lancer',
    characters: defaultCharacters,
    players: [
      playerInfo({ id: ML_PLAYER_ID, name: 'E2E Magic Lancer', camp: 'Red', char_role: 'magic_lancer', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy E1', camp: 'Blue', char_role: 'enemy_char' }),
      playerInfo({ id: ALLY_PLAYER_ID, name: 'Ally A1', camp: 'Red', char_role: 'ally_char' }),
    ],
    initialState: syncState({
      turn_player_id: ML_PLAYER_ID,
      turn_stage: options.turnStage ?? 'ActionExecution',
      available_skills: options.availableSkills ?? [],
      characters: defaultCharacters,
      players,
    }),
  };
}

// ============================================================
// Dark Release (暗之解放) 场景
// ============================================================

export function darkReleaseScenario(): ProtocolHarnessScenario {
  return magicLancerScenario({
    turnStage: 'ActionStart',
    availableSkills: [
      magicLancerAvailableSkill({
        id: ML_DARK_RELEASE_SKILL_ID,
        title: '暗之解放',
        target_type: 0,
      }),
    ],
  });
}

// ============================================================
// Phantom Stardust (幻影星尘) 场景
// ============================================================

export function phantomStardustScenario(): ProtocolHarnessScenario {
  return magicLancerScenario({
    turnStage: 'ActionStart',
    form: ['magic_lance_phantom_form'],
    availableSkills: [
      magicLancerAvailableSkill({
        id: ML_PHANTOM_STARDUST_SKILL_ID,
        title: '幻影星尘',
        target_type: 2,
        min_targets: 1,
        max_targets: 1,
      }),
    ],
  });
}

export function stardustTargetPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: ML_PLAYER_ID,
    message: '【幻影星尘】请选择2点法术伤害目标：',
    choice_type: 'ml_stardust_target',
    skill_id: ML_PHANTOM_STARDUST_SKILL_ID,
    presentation: { kind: 'target_picker', target_filter: 'custom', numeric_base: 0 },
    options: [
      { id: ENEMY_PLAYER_ID, target_id: ENEMY_PLAYER_ID, label: 'Enemy E1', button_label: '选择' },
    ],
    min: 1,
    max: 1,
  } satisfies Prompt);
}

// ============================================================
// Dark Barrier (暗之障壁) 场景
// ============================================================

export function darkBarrierScenario(options: {
  hand?: Card[];
} = {}): ProtocolHarnessScenario {
  const hand = options.hand ?? [
    card({ id: 'ml-magic-1', name: '圣光', type: 'Magic', element: 'Light' }),
    card({ id: 'ml-magic-2', name: '魔弹', type: 'Magic', element: 'Dark' }),
    card({ id: 'ml-thunder-1', name: '雷光斩', type: 'Attack', element: 'Thunder' }),
  ];
  return magicLancerScenario({
    hand,
  });
}

export function darkBarrierCardsPrompt(options?: {
  cardOptions?: Array<{ id: string; label: string; button_label: string; card_id: string }>;
}): WsMessage {
  const opts = options?.cardOptions ?? [
    { id: 'ml-magic-1', label: '1: 圣光（光系 法术）', button_label: '选择', card_id: 'ml-magic-1' },
    { id: 'ml-magic-2', label: '2: 魔弹（暗灭 法术）', button_label: '选择', card_id: 'ml-magic-2' },
    { id: 'ml-thunder-1', label: '3: 雷光斩（雷系 攻击）', button_label: '选择', card_id: 'ml-thunder-1' },
  ];
  return requireActionMessage({
    type: 'choose_cards',
    player_id: ML_PLAYER_ID,
    message: '【暗之障壁】请选择要弃置的牌（法术牌或雷系牌）：',
    choice_type: 'ml_dark_barrier_cards',
    skill_id: '',
    options: opts,
    presentation: { kind: 'card_picker', card_source: 'hand', card_filter: 'magic_or_thunder_chain', numeric_base: 0 },
    min: 1,
    max: opts.length,
  } satisfies Prompt);
}

// ============================================================
// Fullness (充盈) 场景
// ============================================================

export function fullnessScenario(): ProtocolHarnessScenario {
  return magicLancerScenario({
    availableSkills: [
      magicLancerAvailableSkill({
        id: ML_FULLNESS_SKILL_ID,
        title: '充盈',
        target_type: 0,
      }),
    ],
  });
}

export function fullnessAllyDiscardScenario(): ProtocolHarnessScenario {
  const allyHand = [
    card({ id: 'ally-card-0', name: '水涟斩', type: 'Attack', element: 'Water' }),
    card({ id: 'ally-card-1', name: '圣光', type: 'Magic', element: 'Light' }),
  ];
  const mlHand = magicLancerHand();

  return {
    roomCode: 'MOCK',
    myPlayerId: ALLY_PLAYER_ID,
    myPlayerName: 'Ally A1',
    characters: defaultCharacters,
    players: [
      playerInfo({ id: ML_PLAYER_ID, name: 'E2E Magic Lancer', camp: 'Red', char_role: 'magic_lancer', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy E1', camp: 'Blue', char_role: 'enemy_char' }),
      playerInfo({ id: ALLY_PLAYER_ID, name: 'Ally A1', camp: 'Red', char_role: 'ally_char' }),
    ],
    initialState: syncState({
      turn_player_id: ML_PLAYER_ID,
      turn_stage: 'ActionExecution',
      available_skills: [],
      characters: defaultCharacters,
      players: [
        playerView({
          id: ML_PLAYER_ID,
          name: 'E2E Magic Lancer',
          camp: 'Red',
          role: 'magic_lancer',
          hand: mlHand,
          hand_count: mlHand.length,
          is_active: true,
        }),
        playerView({
          id: ENEMY_PLAYER_ID,
          name: 'Enemy E1',
          camp: 'Blue',
          role: 'enemy_char',
          hand_count: 1,
          max_hand: 6,
          heal: 1,
          max_heal: 4,
          is_active: false,
        }),
        playerView({
          id: ALLY_PLAYER_ID,
          name: 'Ally A1',
          camp: 'Red',
          role: 'ally_char',
          hand: allyHand,
          hand_count: allyHand.length,
          max_hand: 6,
          heal: 0,
          max_heal: 4,
          is_active: false,
        }),
      ],
    }),
  };
}

export function fullnessCostCardPrompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_cards',
    player_id: ML_PLAYER_ID,
    message: '【充盈】请选择要弃置的1张法术牌或雷系牌：',
    choice_type: 'ml_fullness_cost_card',
    skill_id: ML_FULLNESS_SKILL_ID,
    options: [
      { id: ML_MAGIC_CARD_ID, label: '3: 圣光（光系 法术）', button_label: '选择', card_id: ML_MAGIC_CARD_ID },
      { id: ML_THUNDER_CARD_ID, label: '2: 雷光斩（雷系 攻击）', button_label: '选择', card_id: ML_THUNDER_CARD_ID },
    ],
    presentation: { kind: 'card_picker', card_source: 'hand', card_filter: 'magic_or_thunder', numeric_base: 0 },
    min: 1,
    max: 1,
  } satisfies Prompt);
}

export function fullnessAllyDiscardPrompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_cards',
    player_id: ALLY_PLAYER_ID,
    message: '【充盈】请选择是否弃置1张手牌：',
    choice_type: 'ml_fullness_discard_step',
    skill_id: ML_FULLNESS_SKILL_ID,
    options: [
      { id: '-1', label: '不弃置', button_label: '不弃置' },
      { id: 'ally-card-0', label: '水涟斩（水系 攻击）', button_label: '选择', card_id: 'ally-card-0' },
      { id: 'ally-card-1', label: '圣光（光系 法术）', button_label: '选择', card_id: 'ally-card-1' },
    ],
    presentation: { kind: 'card_picker', card_source: 'hand', card_filter: 'option_limited', has_decline: true, decline_index: 0, cancel_policy: 'decline', numeric_base: 0 },
    min: 1,
    max: 1,
  } satisfies Prompt);
}

// ============================================================
// Black Spear (漆黑之枪) 场景
// ============================================================

export function blackSpearScenario(options: {
  crystal?: number;
} = {}): ProtocolHarnessScenario {
  return magicLancerScenario({
    crystal: options.crystal ?? 2,
    form: ['magic_lance_phantom_form'],
    indicators: { ml_dark_release_next_attack_bonus: 1 },
  });
}

export function blackSpearXPrompt(maxX = 2): WsMessage {
  const opts: Array<{ id: string; label: string; button_label: string }> = [];
  for (let x = 1; x <= maxX; x++) {
    opts.push({
      id: String(x),
      label: `X=${x}（消耗${x}蓝水晶，伤害额外+${x + 2}）`,
      button_label: String(x),
    });
  }
  return requireActionMessage({
    type: 'confirm',
    player_id: ML_PLAYER_ID,
    message: '【漆黑之枪】请选择X值：',
    choice_type: 'ml_black_spear_x',
    skill_id: ML_BLACK_SPEAR_SKILL_ID,
    options: opts,
    presentation: { kind: 'numeric', numeric_base: 0 },
    min: 1,
    max: 1,
  } satisfies Prompt);
}
