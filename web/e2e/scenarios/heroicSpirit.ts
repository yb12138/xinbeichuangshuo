// ============================================================
// HeroicSpirit (英灵人形) Protocol Harness Scenarios
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

export const HEROIC_SPIRIT_PLAYER_ID = 'heroic_spirit_player';
export const ENEMY_PLAYER_ID = 'enemy_1';
export const ALLY_PLAYER_ID = 'ally_1';

export const HEROIC_SPIRIT_RAGE_SUPPRESS_ID = 'heroic_spirit_rage_suppress';
export const HEROIC_SPIRIT_SEAL_STRIKE_ID = 'heroic_spirit_seal_strike';
export const HEROIC_SPIRIT_MAGIC_FUSION_ID = 'heroic_spirit_magic_fusion';
export const HEROIC_SPIRIT_RUNE_MODIFICATION_ID = 'heroic_spirit_rune_modification';
export const HEROIC_SPIRIT_DOUBLE_ECHO_ID = 'heroic_spirit_double_echo';

const heroicSpiritCharacter = characterView({
  id: 'heroic_spirit',
  name: '英灵人形',
  title: '咏',
  faction: '咏',
  skills: [
    {
      id: HEROIC_SPIRIT_RAGE_SUPPRESS_ID,
      title: '怒火压制',
      description: '（未命中时发动）翻转1个［战纹］，本次攻击强制命中。',
      type: 3, // 响应
      min_targets: 0, max_targets: 0, target_type: 0,
    },
    {
      id: HEROIC_SPIRIT_SEAL_STRIKE_ID,
      title: '战纹碎击',
      description: '（命中时发动）弃1张与攻击牌相同系别的牌［展示］，翻转1个［战纹］。若你处于魔纹形态，本次攻击伤害+1。',
      type: 3, // 响应
      min_targets: 0, max_targets: 0, target_type: 0,
    },
    {
      id: HEROIC_SPIRIT_MAGIC_FUSION_ID,
      title: '魔纹融合',
      description: '（未命中时发动）弃1张与攻击牌系别不同的牌［展示］，翻转1个［战纹］。若你处于魔纹形态，本次攻击伤害+1。',
      type: 3, // 响应
      min_targets: 0, max_targets: 0, target_type: 0,
    },
    {
      id: HEROIC_SPIRIT_RUNE_MODIFICATION_ID,
      title: '符文改造',
      description: '［宝石］横置，进入魔纹形态。调整你的［战纹］。',
      type: 1, // 启动(大招)
      min_targets: 0, max_targets: 0, target_type: 0,
    },
    {
      id: HEROIC_SPIRIT_DOUBLE_ECHO_ID,
      title: '双重回响',
      description: '［水晶］（命中后发动）对另一名目标造成本次攻击同等伤害。',
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

const defaultCharacters = [heroicSpiritCharacter, enemyCharacter, allyCharacter];

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function heroicSpiritHand(): Card[] {
  return [
    card({ id: 'hs-fire-attack', name: '火焰斩', type: 'Attack', element: 'Fire' }),
    card({ id: 'hs-fire-magic', name: '火球', type: 'Magic', element: 'Fire' }),
    card({ id: 'hs-water-attack', name: '水刃斩', type: 'Attack', element: 'Water' }),
    card({ id: 'hs-water-magic', name: '冰冻', type: 'Magic', element: 'Water' }),
    card({ id: 'hs-thunder-magic', name: '雷击', type: 'Magic', element: 'Thunder' }),
  ];
}

function heroicSpiritAvailableSkill(skill: Partial<AvailableSkill> & { id: string; title: string }): AvailableSkill {
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

export function heroicSpiritScenario(options: {
  hand?: Card[];
  availableSkills?: AvailableSkill[];
  crystal?: number;
  gem?: number;
  turnStage?: string;
  buffs?: { id: string; name: string; duration: number; value: number; source_id: string }[];
  tokens?: Record<string, number>;
} = {}): ProtocolHarnessScenario {
  const hand = options.hand ?? heroicSpiritHand();
  const players = [
    playerView({
      id: HEROIC_SPIRIT_PLAYER_ID,
      name: 'E2E HeroicSpirit',
      camp: 'Red',
      role: 'heroic_spirit',
      hand,
      hand_count: hand.length,
      crystal: options.crystal ?? 0,
      gem: options.gem ?? 0,
      is_active: true,
      buffs: options.buffs ?? [],
      tokens: options.tokens ?? { battle_seal: 3 },
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
    myPlayerId: HEROIC_SPIRIT_PLAYER_ID,
    myPlayerName: 'E2E HeroicSpirit',
    characters: defaultCharacters,
    players: [
      playerInfo({ id: HEROIC_SPIRIT_PLAYER_ID, name: 'E2E HeroicSpirit', camp: 'Red', char_role: 'heroic_spirit', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy E1', camp: 'Blue', char_role: 'enemy_char' }),
      playerInfo({ id: ALLY_PLAYER_ID, name: 'Ally A1', camp: 'Red', char_role: 'ally_char' }),
    ],
    initialState: syncState({
      turn_player_id: HEROIC_SPIRIT_PLAYER_ID,
      turn_stage: options.turnStage ?? 'ActionExecution',
      available_skills: options.availableSkills ?? [],
      characters: defaultCharacters,
      players,
    }),
  };
}

// ============================================================
// Rage Suppress (怒火压制) - 响应技能
// ============================================================

export function rageSuppressScenario(): ProtocolHarnessScenario {
  return heroicSpiritScenario({
    tokens: { battle_seal: 3 },
  });
}

export function rageSuppressPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: HEROIC_SPIRIT_PLAYER_ID,
    message: '【怒火压制】翻转1个战纹，本次攻击强制命中？',
    choice_type: 'hs_rage_suppress',
    options: [
      { id: 'confirm', label: '发动' },
      { id: 'skip', label: '跳过' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}

export function rageSuppressSealSelectPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: HEROIC_SPIRIT_PLAYER_ID,
    message: '【怒火压制】请选择翻转的战纹：',
    choice_type: 'hs_rage_suppress_seal',
    options: [
      { id: 'seal_1', label: '战纹1' },
      { id: 'seal_2', label: '战纹2' },
      { id: 'seal_3', label: '战纹3' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}

// ============================================================
// Seal Strike (战纹碎击) - 响应技能
// ============================================================

// Note: attackElement is used to filter discard options, passed to discardPrompt
export function sealStrikeScenario(_options?: {
  attackElement?: string;
}): ProtocolHarnessScenario {
  return heroicSpiritScenario({
    buffs: [{ id: 'magic_seal_form', name: '魔纹形态', duration: 0, value: 0, source_id: HEROIC_SPIRIT_RUNE_MODIFICATION_ID }],
  });
}

export function sealStrikePrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: HEROIC_SPIRIT_PLAYER_ID,
    message: '【战纹碎击】弃同系牌，翻转1个战纹？（魔纹形态+1伤害）',
    choice_type: 'hs_seal_strike',
    options: [
      { id: 'confirm', label: '发动' },
      { id: 'skip', label: '跳过' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}

export function sealStrikeDiscardPrompt(attackElement: string): WsMessage {
  const elementLabel = attackElement === 'Fire' ? '火系' : '水系';
  return requireActionMessage({
    type: 'choose_cards',
    player_id: HEROIC_SPIRIT_PLAYER_ID,
    message: `【战纹碎击】请选择弃1张${elementLabel}牌［展示］：`,
    choice_type: 'hs_seal_strike_discard',
    options: attackElement === 'Fire' ? [
      { id: 'hs-fire-attack', label: '火焰斩（火系）' },
      { id: 'hs-fire-magic', label: '火球（火系）' },
    ] : [
      { id: 'hs-water-attack', label: '水刃斩（水系）' },
      { id: 'hs-water-magic', label: '冰冻（水系）' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}

export function sealStrikeSealSelectPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: HEROIC_SPIRIT_PLAYER_ID,
    message: '【战纹碎击】请选择翻转的战纹：',
    choice_type: 'hs_seal_strike_seal',
    options: [
      { id: 'seal_1', label: '战纹1' },
      { id: 'seal_2', label: '战纹2' },
      { id: 'seal_3', label: '战纹3' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}

// ============================================================
// Magic Fusion (魔纹融合) - 响应技能
// ============================================================

// Note: attackElement is used to filter discard options, passed to discardPrompt
export function magicFusionScenario(_options?: {
  attackElement?: string;
}): ProtocolHarnessScenario {
  return heroicSpiritScenario({
    buffs: [{ id: 'magic_seal_form', name: '魔纹形态', duration: 0, value: 0, source_id: HEROIC_SPIRIT_RUNE_MODIFICATION_ID }],
  });
}

export function magicFusionPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: HEROIC_SPIRIT_PLAYER_ID,
    message: '【魔纹融合】弃异系牌，翻转1个战纹？（魔纹形态+1伤害）',
    choice_type: 'hs_magic_fusion',
    options: [
      { id: 'confirm', label: '发动' },
      { id: 'skip', label: '跳过' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}

export function magicFusionDiscardPrompt(attackElement: string): WsMessage {
  // 弃异系牌：如果攻击是火系，则弃非火系牌
  const excludeElement = attackElement === 'Fire' ? '火系' : '水系';
  return requireActionMessage({
    type: 'choose_cards',
    player_id: HEROIC_SPIRIT_PLAYER_ID,
    message: `【魔纹融合】请选择弃1张非${excludeElement}牌［展示］：`,
    choice_type: 'hs_magic_fusion_discard',
    options: attackElement === 'Fire' ? [
      { id: 'hs-water-attack', label: '水刃斩（水系）' },
      { id: 'hs-water-magic', label: '冰冻（水系）' },
      { id: 'hs-thunder-magic', label: '雷击（雷系）' },
    ] : [
      { id: 'hs-fire-attack', label: '火焰斩（火系）' },
      { id: 'hs-fire-magic', label: '火球（火系）' },
      { id: 'hs-thunder-magic', label: '雷击（雷系）' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}

export function magicFusionSealSelectPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: HEROIC_SPIRIT_PLAYER_ID,
    message: '【魔纹融合】请选择翻转的战纹：',
    choice_type: 'hs_magic_fusion_seal',
    options: [
      { id: 'seal_1', label: '战纹1' },
      { id: 'seal_2', label: '战纹2' },
      { id: 'seal_3', label: '战纹3' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}

// ============================================================
// Seal Suppress Combo (怒火压制与魔纹融合互斥)
// ============================================================

export function sealSuppressComboScenario(): ProtocolHarnessScenario {
  return heroicSpiritScenario();
}

// 互斥提示：选择发动哪个技能
export function sealSuppressComboPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: HEROIC_SPIRIT_PLAYER_ID,
    message: '未命中时，请选择发动哪个技能：',
    choice_type: 'hs_seal_suppress_combo',
    options: [
      { id: HEROIC_SPIRIT_RAGE_SUPPRESS_ID, label: '怒火压制（强制命中）' },
      { id: HEROIC_SPIRIT_MAGIC_FUSION_ID, label: '魔纹融合（魔纹形态+1伤害）' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}

// ============================================================
// Rune Modification (符文改造) - 启动技能(大招)
// ============================================================

export function runeModificationScenario(): ProtocolHarnessScenario {
  return heroicSpiritScenario({
    gem: 1,
    availableSkills: [
      heroicSpiritAvailableSkill({
        id: HEROIC_SPIRIT_RUNE_MODIFICATION_ID, title: '符文改造',
      }),
    ],
  });
}

export function runeModificationPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: HEROIC_SPIRIT_PLAYER_ID,
    message: '【符文改造］消耗宝石，横置进入魔纹形态，调整战纹？',
    choice_type: 'hs_rune_modification',
    options: [
      { id: 'confirm', label: '发动' },
      { id: 'skip', label: '跳过' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}

export function runeModificationSealAdjustPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: HEROIC_SPIRIT_PLAYER_ID,
    message: '【符文改造】请调整战纹：',
    choice_type: 'hs_rune_modification_seal_adjust',
    options: [
      { id: 'flip_1', label: '翻转战纹1' },
      { id: 'flip_2', label: '翻转战纹2' },
      { id: 'flip_3', label: '翻转战纹3' },
    ],
    min: 0, max: 3,
  } satisfies Prompt);
}

// ============================================================
// Double Echo (双重回响) - 响应技能(大招)
// ============================================================

export function doubleEchoScenario(): ProtocolHarnessScenario {
  return heroicSpiritScenario({
    crystal: 1,
  });
}

export function doubleEchoPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: HEROIC_SPIRIT_PLAYER_ID,
    message: '【双重回响］消耗水晶，命中后对另一名目标造成同等伤害？',
    choice_type: 'hs_double_echo',
    options: [
      { id: 'confirm', label: '发动' },
      { id: 'skip', label: '跳过' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}

export function doubleEchoTargetPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: HEROIC_SPIRIT_PLAYER_ID,
    message: '【双重回响】请选择另一名目标造成同等伤害：',
    choice_type: 'hs_double_echo_target',
    options: [
      { id: ALLY_PLAYER_ID, label: 'Ally A1' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}