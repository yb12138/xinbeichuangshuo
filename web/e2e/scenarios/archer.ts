// ============================================================
// Archer (神箭手) Protocol Harness Scenarios
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

export const ARCHER_PLAYER_ID = 'archer_player';
export const ENEMY_PLAYER_ID = 'enemy_1';
export const ALLY_PLAYER_ID = 'ally_1';

export const ARCHER_PIERCING_SHOT_ID = 'archer_piercing_shot';
export const ARCHER_SNipe_ID = 'archer_snipe';
export const ARCHER_PRECISE_SHOT_ID = 'archer_precise_shot'; // 独有牌
export const ARCHER_FLASH_TRAP_ID = 'archer_flash_trap'; // 独有牌

const archerCharacter = characterView({
  id: 'archer',
  name: '神箭手',
  title: '技',
  faction: '技',
  skills: [
    {
      id: ARCHER_PIERCING_SHOT_ID,
      title: '贯穿射击',
      description: '（主动攻击未命中时发动②，弃1张法术牌［展示］）对你所攻击的目标造成2点法术伤害③。',
      type: 3, // 响应
      min_targets: 0, max_targets: 0, target_type: 0,
    },
    {
      id: 'archer_lightning_arrow',
      title: '闪电箭',
      description: '你的雷系攻击对手无法应战。',
      type: 0, // 被动
      min_targets: 0, max_targets: 0, target_type: 0,
    },
    {
      id: ARCHER_SNipe_ID,
      title: '狙击',
      description: '［水晶］目标角色手牌补到5张［强制］，额外+1［攻击行动］。',
      type: 2, // 法术(大招)
      min_targets: 1, max_targets: 1, target_type: 3,
    },
    {
      id: ARCHER_PRECISE_SHOT_ID,
      title: '精准射击',
      description: '此攻击强制命中，但本次攻击伤害-1。',
      type: 3, // 响应(独有)
      min_targets: 0, max_targets: 0, target_type: 0,
    },
    {
      id: ARCHER_FLASH_TRAP_ID,
      title: '闪光陷阱',
      description: '对目标角色造成2点法术伤害③。',
      type: 2, // 法术(独有)
      min_targets: 1, max_targets: 1, target_type: 3,
    },
  ],
});

const enemyCharacter = characterView({
  id: 'enemy_char', name: '守卫', title: '测试目标', faction: '异端', skills: [],
});

const allyCharacter = characterView({
  id: 'ally_char', name: '勇者', title: '测试队友', faction: '技', skills: [],
});

const defaultCharacters = [archerCharacter, enemyCharacter, allyCharacter];

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function archerHand(): Card[] {
  return [
    card({ id: 'archer-thunder-atk1', name: '雷击斩', type: 'Attack', element: 'Thunder' }),
    card({ id: 'archer-thunder-atk2', name: '闪电斩', type: 'Attack', element: 'Thunder' }),
    card({ id: 'archer-thunder-magic', name: '雷击', type: 'Magic', element: 'Thunder' }),
    card({ id: 'archer-fire-magic', name: '火球', type: 'Magic', element: 'Fire' }),
    card({ id: ARCHER_PRECISE_SHOT_ID, name: '精准射击', type: 'Attack', element: 'Thunder' }),
    card({ id: ARCHER_FLASH_TRAP_ID, name: '闪光陷阱', type: 'Magic', element: 'Light' }),
  ];
}

function archerAvailableSkill(skill: Partial<AvailableSkill> & { id: string; title: string }): AvailableSkill {
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

export function archerScenario(options: {
  hand?: Card[];
  availableSkills?: AvailableSkill[];
  crystal?: number;
  gem?: number;
  turnStage?: string;
} = {}): ProtocolHarnessScenario {
  const hand = options.hand ?? archerHand();
  const players = [
    playerView({
      id: ARCHER_PLAYER_ID,
      name: 'E2E Archer',
      camp: 'Red',
      role: 'archer',
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
    myPlayerId: ARCHER_PLAYER_ID,
    myPlayerName: 'E2E Archer',
    characters: defaultCharacters,
    players: [
      playerInfo({ id: ARCHER_PLAYER_ID, name: 'E2E Archer', camp: 'Red', char_role: 'archer', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy E1', camp: 'Blue', char_role: 'enemy_char' }),
      playerInfo({ id: ALLY_PLAYER_ID, name: 'Ally A1', camp: 'Red', char_role: 'ally_char' }),
    ],
    initialState: syncState({
      turn_player_id: ARCHER_PLAYER_ID,
      turn_stage: options.turnStage ?? 'ActionExecution',
      available_skills: options.availableSkills ?? [],
      characters: defaultCharacters,
      players,
    }),
  };
}

// ============================================================
// Piercing Shot (贯穿射击) - 响应技能
// ============================================================

export function piercingShotScenario(): ProtocolHarnessScenario {
  return archerScenario();
}

export function piercingShotMissPrompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_skill',
    player_id: ARCHER_PLAYER_ID,
    message: '你触发了响应技能【贯穿射击】，请选择是否发动。',
    choice_type: 'response_skill_choice',
    options: [
      { id: 'piercing_shot', label: '发动贯穿射击', button_label: '发动', hint: '弃1张法术牌对目标造成2点法术伤害' },
      { id: 'skip', label: '跳过', button_label: '跳过', hint: '不发动响应技能' },
    ],
    min: 1, max: 1,
    presentation: { kind: 'skill_choice', layout: 'overlay' },
  } satisfies Prompt);
}

export function piercingShotDiscardPrompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_cards',
    player_id: ARCHER_PLAYER_ID,
    message: '【贯穿射击】请选择1张法术牌弃置（展示）：',
    choice_type: 'discard_cards',
    options: [
      { id: '2', label: '3: 雷击（雷系 法术）', button_label: '选择', card_id: 'archer-thunder-magic' },
      { id: '3', label: '4: 火球（火系 法术）', button_label: '选择', card_id: 'archer-fire-magic' },
      { id: '5', label: '6: 闪光陷阱（光系 法术）', button_label: '选择', card_id: ARCHER_FLASH_TRAP_ID },
    ],
    min: 1, max: 1,
    presentation: { kind: 'card_picker', card_source: 'hand', card_filter: 'magic_only' },
  } satisfies Prompt);
}

// ============================================================
// Snipe (狙击) - 法术技能(大招)
// ============================================================

export function snipeScenario(): ProtocolHarnessScenario {
  return archerScenario({
    crystal: 1,
    availableSkills: [
      archerAvailableSkill({
        id: ARCHER_SNipe_ID, title: '狙击',
        min_targets: 1, max_targets: 1, target_type: 3,
      }),
    ],
  });
}

export function snipeTargetPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: ARCHER_PLAYER_ID,
    message: '【狙击】请选择一名目标角色，使其手牌补到5张：',
    choice_type: 'skill_target_selection',
    options: [
      { id: ENEMY_PLAYER_ID, label: 'Enemy E1（手牌1→5）', button_label: '选择' },
      { id: ALLY_PLAYER_ID, label: 'Ally A1（手牌1→5）', button_label: '选择' },
      { id: ARCHER_PLAYER_ID, label: '自己', button_label: '选择' },
    ],
    min: 1, max: 1,
    presentation: { kind: 'target_picker', target_filter: 'custom', numeric_base: 0 },
  } satisfies Prompt);
}

// ============================================================
// Precise Shot (精准射击) - 独有牌响应
// ============================================================

export function preciseShotScenario(): ProtocolHarnessScenario {
  return archerScenario({
    hand: [
      card({ id: ARCHER_PRECISE_SHOT_ID, name: '精准射击', type: 'Attack', element: 'Thunder' }),
      card({ id: 'archer-thunder-atk1', name: '雷击斩', type: 'Attack', element: 'Thunder' }),
      card({ id: 'archer-fire-magic', name: '火球', type: 'Magic', element: 'Fire' }),
    ],
  });
}

export function preciseShotConfirmPrompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_skill',
    player_id: ARCHER_PLAYER_ID,
    message: '你触发了响应技能【精准射击】，请选择是否发动。',
    choice_type: 'response_skill_choice',
    options: [
      { id: 'precise_shot', label: '发动精准射击', button_label: '发动', hint: '本次攻击强制命中，但伤害-1' },
      { id: 'skip', label: '跳过', button_label: '跳过', hint: '不发动响应技能' },
    ],
    min: 1, max: 1,
    presentation: { kind: 'skill_choice', layout: 'overlay' },
  } satisfies Prompt);
}

// ============================================================
// Flash Trap (闪光陷阱) - 独有法术技能
// ============================================================

export function flashTrapScenario(): ProtocolHarnessScenario {
  return archerScenario({
    hand: [
      card({ id: ARCHER_FLASH_TRAP_ID, name: '闪光陷阱', type: 'Magic', element: 'Light' }),
      card({ id: 'archer-thunder-atk1', name: '雷击斩', type: 'Attack', element: 'Thunder' }),
      card({ id: 'archer-fire-magic', name: '火球', type: 'Magic', element: 'Fire' }),
    ],
  });
}

export function flashTrapTargetPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: ARCHER_PLAYER_ID,
    message: '【闪光陷阱】请选择一名目标角色，对其造成2点法术伤害：',
    choice_type: 'skill_target_selection',
    options: [
      { id: ENEMY_PLAYER_ID, label: 'Enemy E1', button_label: '选择' },
      { id: ALLY_PLAYER_ID, label: 'Ally A1', button_label: '选择' },
    ],
    min: 1, max: 1,
    presentation: { kind: 'target_picker', target_filter: 'custom', numeric_base: 0 },
  } satisfies Prompt);
}
