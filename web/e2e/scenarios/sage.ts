// ============================================================
// Sage (贤者) Protocol Harness Scenarios
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

export const SAGE_PLAYER_ID = 'sage_player';
export const ENEMY_PLAYER_ID = 'enemy_1';
export const ALLY_PLAYER_ID = 'ally_1';
const HOLY_TARGET_ID_BY_NAME: Record<string, string> = {
  'E2E Sage': SAGE_PLAYER_ID,
  'Enemy E1': ENEMY_PLAYER_ID,
  'Ally A1': ALLY_PLAYER_ID,
};

export const SAGE_WISDOM_CODEX_ID = 'sage_wisdom_codex';
export const SAGE_MAGIC_REBOUND_ID = 'sage_magic_rebound';
export const SAGE_ARCANE_CODEX_ID = 'sage_arcane_codex';
export const SAGE_HOLY_CODEX_ID = 'sage_holy_codex';

const sageCharacter = characterView({
  id: 'sage',
  name: '贤者',
  title: '咏',
  faction: '咏',
  skills: [
    {
      id: SAGE_WISDOM_CODEX_ID,
      title: '智慧法典',
      description: '你的能量上限+1；你每次承受法术伤害时，若该伤害>3：你+2红宝石并弃1张牌。',
      type: 0,
      min_targets: 0, max_targets: 0, target_type: 0,
    },
    {
      id: SAGE_MAGIC_REBOUND_ID,
      title: '法术反弹',
      description: '你每次承受法术伤害时，若该伤害仅为1点：可弃X张同系牌（X>1），对目标角色造成(X-1)点法术伤害，并对自己造成X点法术伤害。',
      type: 3,
      min_targets: 1, max_targets: 1, target_type: 3,
    },
    {
      id: SAGE_ARCANE_CODEX_ID,
      title: '魔道法典',
      description: '［宝石］弃X张异系牌（X>1），对目标角色与自己各造成(X-1)点法术伤害。',
      type: 2,
      min_targets: 1, max_targets: 1, target_type: 3,
      cost_gem: 1,
    },
    {
      id: SAGE_HOLY_CODEX_ID,
      title: '圣洁法典',
      description: '［宝石］弃X张异系牌（X>2），最多(X-2)名角色各+2治疗，然后对自己造成(X-1)点法术伤害。',
      type: 2,
      min_targets: 1, max_targets: 6, target_type: 3,
      cost_gem: 1,
    },
  ],
});

const enemyCharacter = characterView({
  id: 'enemy_char', name: '守卫', title: '测试目标', faction: '异端', skills: [],
});

const allyCharacter = characterView({
  id: 'ally_char', name: '勇者', title: '测试队友', faction: '咏', skills: [],
});

const defaultCharacters = [sageCharacter, enemyCharacter, allyCharacter];

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

/** Hand with 6 cards of different elements for arcane/holy codex testing. */
function sageDiverseHand(): Card[] {
  return [
    card({ id: 'sg-fire-atk', name: '火焰斩', type: 'Attack', element: 'Fire' }),
    card({ id: 'sg-water-magic', name: '水流', type: 'Magic', element: 'Water' }),
    card({ id: 'sg-thunder-atk', name: '雷光斩', type: 'Attack', element: 'Thunder' }),
    card({ id: 'sg-earth-magic', name: '地裂', type: 'Magic', element: 'Earth' }),
    card({ id: 'sg-wind-magic', name: '风刃', type: 'Magic', element: 'Wind' }),
    card({ id: 'sg-light-magic', name: '圣光', type: 'Magic', element: 'Light' }),
  ];
}

/** Hand with multiple same-element (Fire) cards for magic rebound testing. */
function sageReboundHand(): Card[] {
  return [
    card({ id: 'sg-fire-atk1', name: '火焰斩A', type: 'Attack', element: 'Fire' }),
    card({ id: 'sg-fire-atk2', name: '火焰斩B', type: 'Attack', element: 'Fire' }),
    card({ id: 'sg-fire-magic', name: '火球', type: 'Magic', element: 'Fire' }),
    card({ id: 'sg-water-magic', name: '水流', type: 'Magic', element: 'Water' }),
    card({ id: 'sg-thunder-atk', name: '雷光斩', type: 'Attack', element: 'Thunder' }),
  ];
}

function sageAvailableSkill(skill: Partial<AvailableSkill> & { id: string; title: string }): AvailableSkill {
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

export function sageScenario(options: {
  hand?: Card[];
  availableSkills?: AvailableSkill[];
  gem?: number;
  crystal?: number;
  turnStage?: string;
  tokens?: Record<string, number>;
} = {}): ProtocolHarnessScenario {
  const hand = options.hand ?? sageDiverseHand();
  const players = [
    playerView({
      id: SAGE_PLAYER_ID,
      name: 'E2E Sage',
      camp: 'Red',
      role: 'sage',
      hand,
      hand_count: hand.length,
      crystal: options.crystal ?? 0,
      gem: options.gem ?? 0,
      is_active: true,
      tokens: options.tokens ?? {},
    }),
    playerView({
      id: ENEMY_PLAYER_ID,
      name: 'Enemy E1',
      camp: 'Blue',
      role: 'enemy_char',
      hand: [card({ id: 'en-card-1', name: '测试牌', type: 'Attack', element: 'Fire' })],
      hand_count: 1,
      heal: 1, max_heal: 4,
      is_active: false,
    }),
    playerView({
      id: ALLY_PLAYER_ID,
      name: 'Ally A1',
      camp: 'Red',
      role: 'ally_char',
      hand: [card({ id: 'al-card-1', name: '测试牌', type: 'Attack', element: 'Water' })],
      hand_count: 1,
      heal: 0, max_heal: 4,
      is_active: false,
    }),
  ];

  return {
    roomCode: 'MOCK',
    myPlayerId: SAGE_PLAYER_ID,
    myPlayerName: 'E2E Sage',
    characters: defaultCharacters,
    players: [
      playerInfo({ id: SAGE_PLAYER_ID, name: 'E2E Sage', camp: 'Red', char_role: 'sage', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy E1', camp: 'Blue', char_role: 'enemy_char' }),
      playerInfo({ id: ALLY_PLAYER_ID, name: 'Ally A1', camp: 'Red', char_role: 'ally_char' }),
    ],
    initialState: syncState({
      turn_player_id: SAGE_PLAYER_ID,
      turn_stage: options.turnStage ?? 'ActionExecution',
      available_skills: options.availableSkills ?? [],
      characters: defaultCharacters,
      players,
    }),
  };
}

// ============================================================
// Wisdom Codex (智慧法典) — standard system discard
// ============================================================

export function wisdomCodexDiscardPrompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_cards',
    player_id: SAGE_PLAYER_ID,
    message: '【智慧法典】请选择弃置1张手牌：',
    choice_type: 'system_discard_cards',
    options: [
      { id: '0', label: '1: 火焰斩（火系 攻击）' },
      { id: '1', label: '2: 水流（水系 法术）' },
      { id: '2', label: '3: 雷光斩（雷系 攻击）' },
      { id: '3', label: '4: 地裂（地系 法术）' },
      { id: '4', label: '5: 风刃（风系 法术）' },
      { id: '5', label: '6: 圣光（光系 法术）' },
    ],
    min: 1, max: totalCount,
  } satisfies Prompt);
}

// ============================================================
// Magic Rebound (法术反弹) prompts
// ============================================================

export function magicReboundScenario(options: {
  hand?: Card[];
} = {}): ProtocolHarnessScenario {
  return sageScenario({
    hand: options.hand ?? sageReboundHand(),
  });
}

export function magicReboundConfirmPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: SAGE_PLAYER_ID,
    message: '【法术反弹】是否发动？',
    choice_type: 'sage_magic_rebound_confirm',
    options: [
      { id: '0', label: '是' },
      { id: '1', label: '否' },
    ],
    min: 1, max: totalCount,
  } satisfies Prompt);
}

export function magicReboundElementPrompt(elementCount = 2): WsMessage {
  const elementOptions = [
    { id: '0', label: '火系' },
    { id: '1', label: '水系' },
  ].slice(0, elementCount);

  return requireActionMessage({
    type: 'confirm',
    player_id: SAGE_PLAYER_ID,
    message: '【法术反弹】请选择弃置同系牌的元素：',
    choice_type: 'sage_magic_rebound_element',
    options: elementOptions,
    min: 1, max: 1,
  } satisfies Prompt);
}

export function magicReboundCardsPrompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_cards',
    player_id: SAGE_PLAYER_ID,
    message: '【法术反弹】请选择同系牌（选几张X即为几）：',
    choice_type: 'sage_magic_rebound_cards',
    options: [
      { id: '0', label: '1: 火焰斩A（火系 攻击）' },
      { id: '1', label: '2: 火焰斩B（火系 攻击）' },
      { id: '2', label: '3: 火球（火系 法术）' },
    ],
    min: 2, max: 3,
  } satisfies Prompt);
}

export function magicReboundTargetPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: SAGE_PLAYER_ID,
    message: '【法术反弹】请选择目标角色：',
    choice_type: 'sage_magic_rebound_target',
    options: [
      { id: ENEMY_PLAYER_ID, label: 'Enemy E1' },
      { id: ALLY_PLAYER_ID, label: 'Ally A1' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}

// ============================================================
// Arcane Codex (魔道法典) prompts
// ============================================================

export function arcaneCodexScenario(): ProtocolHarnessScenario {
  return sageScenario({
    gem: 1,
    availableSkills: [
      sageAvailableSkill({ id: SAGE_ARCANE_CODEX_ID, title: '魔道法典', cost_gem: 1 }),
    ],
  });
}

export function arcaneCardsPrompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_cards',
    player_id: SAGE_PLAYER_ID,
    message: '【魔道法典】请选择异系牌（选几张X即为几）：',
    choice_type: 'sage_arcane_cards',
    options: [
      { id: '0', label: '1: 火焰斩（火系 攻击）' },
      { id: '1', label: '2: 水流（水系 法术）' },
      { id: '2', label: '3: 雷光斩（雷系 攻击）' },
      { id: '3', label: '4: 地裂（地系 法术）' },
      { id: '4', label: '5: 风刃（风系 法术）' },
      { id: '5', label: '6: 圣光（光系 法术）' },
    ],
    min: 2, max: 6,
  } satisfies Prompt);
}

export function arcaneTargetPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: SAGE_PLAYER_ID,
    message: '【魔道法典】请选择目标角色：',
    choice_type: 'sage_arcane_target',
    options: [
      { id: ENEMY_PLAYER_ID, label: 'Enemy E1' },
      { id: ALLY_PLAYER_ID, label: 'Ally A1' },
      { id: SAGE_PLAYER_ID, label: 'E2E Sage' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}

// ============================================================
// Holy Codex (圣洁法典) prompts
// ============================================================

export function holyCodexScenario(): ProtocolHarnessScenario {
  return sageScenario({
    gem: 1,
    availableSkills: [
      sageAvailableSkill({ id: SAGE_HOLY_CODEX_ID, title: '圣洁法典', cost_gem: 1 }),
    ],
  });
}

export function holyCardsPrompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_cards',
    player_id: SAGE_PLAYER_ID,
    message: '【圣洁法典】请选择异系牌（选几张X即为几）：',
    choice_type: 'sage_holy_cards',
    options: [
      { id: '0', label: '1: 火焰斩（火系 攻击）' },
      { id: '1', label: '2: 水流（水系 法术）' },
      { id: '2', label: '3: 雷光斩（雷系 攻击）' },
      { id: '3', label: '4: 地裂（地系 法术）' },
      { id: '4', label: '5: 风刃（风系 法术）' },
      { id: '5', label: '6: 圣光（光系 法术）' },
    ],
    min: 3, max: 6,
  } satisfies Prompt);
}

export function holyTargetCountPrompt(maxTargets = 2): WsMessage {
  const options: Array<{ id: string; label: string }> = [];
  for (let count = 1; count <= maxTargets; count++) {
    options.push({ id: String(count - 1), label: `选择${count}名角色` });
  }
  return requireActionMessage({
    type: 'confirm',
    player_id: SAGE_PLAYER_ID,
    message: '【圣洁法典】请选择要获得治疗的角色数量：',
    choice_type: 'sage_holy_target_count',
    options,
    min: 1, max: 1,
  } satisfies Prompt);
}

export function holyTargetsStepPrompt(
  _step: number,
  totalCount: number,
  remainingNames: string[],
): WsMessage {
  const options = remainingNames.map((name, idx) => ({
    id: HOLY_TARGET_ID_BY_NAME[name] ?? String(idx),
    label: name,
  }));
  return requireActionMessage({
    type: 'confirm',
    player_id: SAGE_PLAYER_ID,
    message: `【圣洁法典】请选择治疗目标（1-${totalCount}名）：`,
    choice_type: 'sage_holy_targets',
    options,
    presentation: { kind: 'target_picker' },
    min: 1, max: 1,
  } satisfies Prompt);
}
