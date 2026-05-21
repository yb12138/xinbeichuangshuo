// ============================================================
// Blaze Witch (苍炎魔女) Protocol Harness Scenarios
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

export const BW_PLAYER_ID = 'bw_player';
export const ENEMY_PLAYER_ID = 'enemy_1';
export const ALLY_PLAYER_ID = 'ally_1';

export const BW_BLAZING_CODEX_ID = 'bw_blazing_codex';
export const BW_HEAVENFIRE_CLEAVE_ID = 'bw_heavenfire_cleave';
export const BW_WITCH_WRATH_ID = 'bw_witch_wrath';
export const BW_SUBSTITUTE_DOLL_ID = 'bw_substitute_doll';
export const BW_PAIN_LINK_ID = 'bw_pain_link';
export const BW_MANA_INVERSION_ID = 'bw_mana_inversion';

const blazeWitchCharacter = characterView({
  id: 'blaze_witch',
  name: '苍炎魔女',
  title: '血',
  faction: '血',
  skills: [
    {
      id: 'bw_rebirth_clock',
      title: '永生银时计',
      description: '［重生］上限4；当你因承受法术伤害导致士气下降时，你+1［重生］。',
      type: 0,
      min_targets: 0, max_targets: 0, target_type: 0,
      cost_gem: 0, cost_crystal: 0, cost_discards: 0,
    },
    {
      id: BW_BLAZING_CODEX_ID,
      title: '苍炎法典',
      description: '弃1张火系牌［展示］，对目标角色和自己各造成2点法术伤害（目标先结算）。',
      type: 2,
      min_targets: 1, max_targets: 1, target_type: 3,
      cost_gem: 0, cost_crystal: 0, cost_discards: 1, discard_element: 'Fire',
    },
    {
      id: BW_HEAVENFIRE_CLEAVE_ID,
      title: '天火断空',
      description: '弃2张火系牌并移除1点［重生］（烈焰形态下免移除），对目标角色和自己各造成3点法术伤害。',
      type: 2,
      min_targets: 1, max_targets: 1, target_type: 3,
      cost_gem: 0, cost_crystal: 0, cost_discards: 2, discard_element: 'Fire',
    },
    {
      id: BW_WITCH_WRATH_ID,
      title: '魔女之怒',
      description: '手牌<4时可发动：［横置］进入烈焰形态并选择摸0~2张牌。',
      type: 1,
      min_targets: 0, max_targets: 0, target_type: 0,
      cost_gem: 0, cost_crystal: 0, cost_discards: 0,
    },
    {
      id: BW_SUBSTITUTE_DOLL_ID,
      title: '替身玩偶',
      description: '任何人对你造成攻击伤害时可响应：弃1张法术牌［展示］，令1名队友摸1张牌。',
      type: 3,
      min_targets: 1, max_targets: 1, target_type: 1,
      cost_gem: 0, cost_crystal: 0, cost_discards: 0,
    },
    {
      id: BW_PAIN_LINK_ID,
      title: '痛苦链接',
      description: '［水晶］对目标对手和自己各造成1点法术伤害，然后你弃到3张手牌。',
      type: 2,
      min_targets: 1, max_targets: 1, target_type: 2,
      cost_gem: 0, cost_crystal: 1, cost_discards: 0,
    },
    {
      id: BW_MANA_INVERSION_ID,
      title: '魔能反转',
      description: '［水晶］任何人对你造成法术伤害时可响应：弃X张法术牌［展示］（X>1），对目标对手造成(X-1)点法术伤害。',
      type: 3,
      min_targets: 1, max_targets: 1, target_type: 2,
      cost_gem: 0, cost_crystal: 1, cost_discards: 0,
    },
  ],
});

const enemyCharacter = characterView({
  id: 'enemy_char', name: '守卫', title: '测试目标', faction: '异端', skills: [],
});

const allyCharacter = characterView({
  id: 'ally_char', name: '勇者', title: '测试队友', faction: '血', skills: [],
});

const defaultCharacters = [blazeWitchCharacter, enemyCharacter, allyCharacter];

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

/** Hand with fire cards + magic cards for all skill testing. */
function bwHand(): Card[] {
  return [
    card({ id: 'bw-fire-atk1', name: '火焰斩A', type: 'Attack', element: 'Fire' }),
    card({ id: 'bw-fire-atk2', name: '火焰斩B', type: 'Attack', element: 'Fire' }),
    card({ id: 'bw-fire-magic1', name: '火球', type: 'Magic', element: 'Fire' }),
    card({ id: 'bw-fire-magic2', name: '烈焰风暴', type: 'Magic', element: 'Fire' }),
    card({ id: 'bw-thunder-magic', name: '雷击', type: 'Magic', element: 'Thunder' }),
    card({ id: 'bw-light-magic', name: '圣光', type: 'Magic', element: 'Light' }),
  ];
}

function bwAvailableSkill(skill: Partial<AvailableSkill> & { id: string; title: string }): AvailableSkill {
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

export function blazeWitchScenario(options: {
  hand?: Card[];
  availableSkills?: AvailableSkill[];
  crystal?: number;
  gem?: number;
  turnStage?: string;
  tokens?: Record<string, number>;
} = {}): ProtocolHarnessScenario {
  const hand = options.hand ?? bwHand();
  const players = [
    playerView({
      id: BW_PLAYER_ID,
      name: 'E2E Blaze Witch',
      camp: 'Red',
      role: 'blaze_witch',
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
      hand_count: 1, max_hand: 6,
      heal: 1, max_heal: 4,
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
    myPlayerId: BW_PLAYER_ID,
    myPlayerName: 'E2E Blaze Witch',
    characters: defaultCharacters,
    players: [
      playerInfo({ id: BW_PLAYER_ID, name: 'E2E Blaze Witch', camp: 'Red', char_role: 'blaze_witch', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy E1', camp: 'Blue', char_role: 'enemy_char' }),
      playerInfo({ id: ALLY_PLAYER_ID, name: 'Ally A1', camp: 'Red', char_role: 'ally_char' }),
    ],
    initialState: syncState({
      turn_player_id: BW_PLAYER_ID,
      turn_stage: options.turnStage ?? 'ActionExecution',
      available_skills: options.availableSkills ?? [],
      characters: defaultCharacters,
      players,
    }),
  };
}

// ============================================================
// Blazing Codex (苍炎法典) scenario
// ============================================================

export function blazingCodexScenario(): ProtocolHarnessScenario {
  return blazeWitchScenario({
    availableSkills: [
      bwAvailableSkill({
        id: BW_BLAZING_CODEX_ID, title: '苍炎法典',
        cost_discards: 1, discard_element: 'Fire',
        min_targets: 1, max_targets: 1, target_type: 3,
      }),
    ],
  });
}

// ============================================================
// Heavenfire Cleave (天火断空) scenario
// ============================================================

export function heavenfireCleaveScenario(options: {
  tokens?: Record<string, number>;
} = {}): ProtocolHarnessScenario {
  return blazeWitchScenario({
    tokens: options.tokens ?? { bw_rebirth: 1 },
    availableSkills: [
      bwAvailableSkill({
        id: BW_HEAVENFIRE_CLEAVE_ID, title: '天火断空',
        cost_discards: 2, discard_element: 'Fire',
        min_targets: 1, max_targets: 1, target_type: 3,
      }),
    ],
  });
}

// ============================================================
// Pain Link (痛苦链接) scenario
// ============================================================

export function painLinkScenario(): ProtocolHarnessScenario {
  return blazeWitchScenario({
    crystal: 1,
    availableSkills: [
      bwAvailableSkill({
        id: BW_PAIN_LINK_ID, title: '痛苦链接',
        cost_crystal: 1,
        min_targets: 1, max_targets: 1, target_type: 2,
      }),
    ],
  });
}

export function painLinkDiscardPrompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_cards',
    player_id: BW_PLAYER_ID,
    message: '【痛苦链接】请弃牌至3张手牌：',
    choice_type: 'system_discard_cards',
    options: [
      { id: '0', label: '1: 火焰斩A（火系 攻击）', button_label: '选择', card_id: 'bw-fire-atk1' },
      { id: '1', label: '2: 火焰斩B（火系 攻击）', button_label: '选择', card_id: 'bw-fire-atk2' },
      { id: '2', label: '3: 火球（火系 法术）', button_label: '选择', card_id: 'bw-fire-magic1' },
      { id: '3', label: '4: 烈焰风暴（火系 法术）', button_label: '选择', card_id: 'bw-fire-magic2' },
      { id: '4', label: '5: 雷击（雷系 法术）', button_label: '选择', card_id: 'bw-thunder-magic' },
      { id: '5', label: '6: 圣光（光系 法术）', button_label: '选择', card_id: 'bw-light-magic' },
    ],
    min: 3, max: 3,
    presentation: { kind: 'card_picker', card_source: 'hand', card_filter: 'option_limited', numeric_base: 0 },
  } satisfies Prompt);
}

// ============================================================
// Witch Wrath (魔女之怒) prompts
// ============================================================

export function witchWrathScenario(): ProtocolHarnessScenario {
  return blazeWitchScenario({
    availableSkills: [
      bwAvailableSkill({ id: BW_WITCH_WRATH_ID, title: '魔女之怒' }),
    ],
  });
}

export function witchWrathDrawPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: BW_PLAYER_ID,
    message: '【魔女之怒】请选择摸牌数量：',
    choice_type: 'bw_witch_wrath_draw',
    options: [
      { id: '0', label: '摸0张', button_label: '摸0张' },
      { id: '1', label: '摸1张', button_label: '摸1张' },
      { id: '2', label: '摸2张', button_label: '摸2张' },
    ],
    min: 1, max: 1,
    presentation: { kind: 'branch_select', layout: 'overlay', numeric_base: 0 },
  } satisfies Prompt);
}

// ============================================================
// Substitute Doll (替身玩偶) prompts
// ============================================================

export function substituteDollScenario(): ProtocolHarnessScenario {
  return blazeWitchScenario();
}

export function substituteDollCardPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: BW_PLAYER_ID,
    message: '【替身玩偶】请选择弃置1张法术牌：',
    choice_type: 'bw_substitute_doll_card',
    options: [
      { id: '2', label: '3: 火球（火系 法术）', button_label: '选择', card_id: 'bw-fire-magic1' },
      { id: '3', label: '4: 烈焰风暴（火系 法术）', button_label: '选择', card_id: 'bw-fire-magic2' },
      { id: '4', label: '5: 雷击（雷系 法术）', button_label: '选择', card_id: 'bw-thunder-magic' },
      { id: '5', label: '6: 圣光（光系 法术）', button_label: '选择', card_id: 'bw-light-magic' },
    ],
    min: 1, max: 1,
    presentation: { kind: 'card_picker', card_source: 'hand', card_filter: 'magic_only', numeric_base: 0 },
  } satisfies Prompt);
}

export function substituteDollTargetPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: BW_PLAYER_ID,
    message: '【替身玩偶】请选择摸1张牌的队友：',
    choice_type: 'bw_substitute_doll_target',
    options: [
      { id: ALLY_PLAYER_ID, target_id: ALLY_PLAYER_ID, label: 'Ally A1', button_label: '选择' },
    ],
    min: 1, max: 1,
    presentation: { kind: 'target_picker', target_filter: 'custom', numeric_base: 0 },
  } satisfies Prompt);
}

// ============================================================
// Mana Inversion (魔能反转) prompts
// ============================================================

export function manaInversionScenario(): ProtocolHarnessScenario {
  return blazeWitchScenario();
}

export function manaInversionCardsPrompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_cards',
    player_id: BW_PLAYER_ID,
    message: '【魔能反转】请选择要弃置的法术牌（X=选择数量，至少2张）：',
    choice_type: 'bw_mana_inversion_cards',
    options: [
      { id: '2', label: '3: 火球（火系 法术）', button_label: '选择', card_id: 'bw-fire-magic1' },
      { id: '3', label: '4: 烈焰风暴（火系 法术）', button_label: '选择', card_id: 'bw-fire-magic2' },
      { id: '4', label: '5: 雷击（雷系 法术）', button_label: '选择', card_id: 'bw-thunder-magic' },
      { id: '5', label: '6: 圣光（光系 法术）', button_label: '选择', card_id: 'bw-light-magic' },
    ],
    min: 2, max: 4,
    presentation: { kind: 'card_picker', card_source: 'hand', card_filter: 'magic_only', numeric_base: 0 },
  } satisfies Prompt);
}

export function manaInversionTargetPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: BW_PLAYER_ID,
    message: '【魔能反转】请选择法术伤害目标：',
    choice_type: 'bw_mana_inversion_target',
    options: [
      { id: ENEMY_PLAYER_ID, target_id: ENEMY_PLAYER_ID, label: 'Enemy E1', button_label: '选择' },
    ],
    min: 1, max: 1,
    presentation: { kind: 'target_picker', target_filter: 'custom', numeric_base: 0 },
  } satisfies Prompt);
}
