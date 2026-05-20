// ============================================================
// Magic Bow (魔弓手) Protocol Harness Scenarios
// ============================================================

import type { AvailableSkill, Card, FieldCard, Prompt } from '../../src/types/game';
import type { WsMessage } from '../../src/network/protocol';
import {
  availableSkill,
  card,
  characterView,
  playerInfo,
  playerView,
  fieldOptionIndexInteraction,
  requireActionMessage,
  syncState,
  type ProtocolHarnessScenario,
} from './builders';

export const MB_PLAYER_ID = 'mb_player';
export const ENEMY_PLAYER_ID = 'enemy_1';
export const ENEMY_2_PLAYER_ID = 'enemy_2';

export const MB_MAGIC_PIERCE_SKILL_ID = 'mb_magic_pierce';
export const MB_THUNDER_SCATTER_SKILL_ID = 'mb_thunder_scatter';
export const MB_MULTI_SHOT_SKILL_ID = 'mb_multi_shot';
export const MB_CHARGE_SKILL_ID = 'mb_charge';
export const MB_DEMON_EYE_SKILL_ID = 'mb_demon_eye';

export const MB_ATTACK_CARD_ID = 'mb-atk-fire';

const magicBowCharacter = characterView({
  id: 'magic_bow',
  name: '魔弓手',
  title: '充能魔弓',
  faction: '星杯',
  skills: [
    {
      id: MB_MAGIC_PIERCE_SKILL_ID,
      title: '魔贯冲击',
      description: '响应技能，攻击发起时移除火系充能。',
      type: 3,
      min_targets: 0,
      max_targets: 0,
      target_type: 0,
      cost_gem: 0, cost_crystal: 0, cost_discards: 0,
    },
    {
      id: MB_THUNDER_SCATTER_SKILL_ID,
      title: '雷光散射',
      description: '法术技能，先移除雷系充能，再决定额外移除数量。',
      type: 2,
      min_targets: 0,
      max_targets: 1,
      target_type: 2,
      cost_gem: 0, cost_crystal: 0, cost_discards: 0,
    },
    {
      id: MB_MULTI_SHOT_SKILL_ID,
      title: '多重射击',
      description: '响应技能，攻击行动结束时移除风系充能。',
      type: 3,
      min_targets: 0,
      max_targets: 0,
      target_type: 0,
      cost_gem: 0, cost_crystal: 0, cost_discards: 0,
    },
    {
      id: MB_CHARGE_SKILL_ID,
      title: '充能',
      description: '启动技能，弃至4张后选择摸牌并盖放充能。',
      type: 1,
      min_targets: 0,
      max_targets: 0,
      target_type: 0,
      cost_gem: 0, cost_crystal: 1, cost_discards: 0,
    },
    {
      id: MB_DEMON_EYE_SKILL_ID,
      title: '魔眼',
      description: '启动技能，选择弃牌或摸牌分支，再盖放1张充能。',
      type: 1,
      min_targets: 0,
      max_targets: 0,
      target_type: 0,
      cost_gem: 1, cost_crystal: 0, cost_discards: 0,
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

const enemy2Character = characterView({
  id: 'enemy_2_char',
  name: '游侠',
  title: '测试目标',
  faction: '异端',
  skills: [],
});

const defaultCharacters = [magicBowCharacter, enemyCharacter, enemy2Character];

function magicBowHand(): Card[] {
  return [
    card({ id: MB_ATTACK_CARD_ID, name: '火焰斩', type: 'Attack', element: 'Fire', damage: 2 }),
    card({ id: 'mb-hand-water', name: '水涟斩', type: 'Attack', element: 'Water' }),
    card({ id: 'mb-hand-thunder', name: '雷光斩', type: 'Attack', element: 'Thunder' }),
    card({ id: 'mb-hand-wind', name: '风神斩', type: 'Attack', element: 'Wind' }),
  ];
}

function magicBowLargeHand(): Card[] {
  return [
    ...magicBowHand(),
    card({ id: 'mb-hand-light', name: '圣光', type: 'Magic', element: 'Light' }),
    card({ id: 'mb-hand-dark', name: '魔弹', type: 'Magic', element: 'Dark' }),
  ];
}

function chargeCover(index: number, element: Card['element'], name: string): FieldCard {
  return {
    card: card({
      id: `mb-charge-${index}`,
      name,
      type: 'Attack',
      element,
    }),
    owner_id: MB_PLAYER_ID,
    source_id: MB_PLAYER_ID,
    mode: 'Cover',
    effect: 'MagicBowCharge',
    field_hook: 'Manual',
    locked: false,
    duration: 0,
  };
}

function magicBowAvailableSkill(skill: Partial<AvailableSkill> & { id: string; title: string }): AvailableSkill {
  return availableSkill({
    description: '',
    target_type: 0,
    min_targets: 0,
    max_targets: 0,
    ...skill,
  });
}

export function magicBowScenario(options: {
  hand?: Card[];
  field?: FieldCard[];
  availableSkills?: AvailableSkill[];
  crystal?: number;
  gem?: number;
  turnStage?: string;
} = {}): ProtocolHarnessScenario {
  const hand = options.hand ?? magicBowHand();
  const field = options.field ?? [
    chargeCover(0, 'Fire', '火焰充能'),
    chargeCover(1, 'Thunder', '雷光充能'),
    chargeCover(2, 'Wind', '风神充能'),
  ];
  const players = [
    playerView({
      id: MB_PLAYER_ID,
      name: 'E2E Magic Bow',
      camp: 'Red',
      role: 'magic_bow',
      hand,
      hand_count: hand.length,
      field,
      crystal: options.crystal ?? 0,
      gem: options.gem ?? 0,
      is_active: true,
      indicators: { mb_charge_count: field.length },
    }),
    playerView({
      id: ENEMY_PLAYER_ID,
      name: 'Enemy E1',
      camp: 'Blue',
      role: 'enemy_char',
      hand: [],
      hand_count: 3,
      max_hand: 6,
      heal: 2,
      max_heal: 4,
      is_active: false,
    }),
    playerView({
      id: ENEMY_2_PLAYER_ID,
      name: 'Enemy E2',
      camp: 'Blue',
      role: 'enemy_2_char',
      hand: [],
      hand_count: 2,
      max_hand: 6,
      heal: 2,
      max_heal: 4,
      is_active: false,
    }),
  ];

  return {
    roomCode: 'MOCK',
    myPlayerId: MB_PLAYER_ID,
    myPlayerName: 'E2E Magic Bow',
    characters: defaultCharacters,
    players: [
      playerInfo({ id: MB_PLAYER_ID, name: 'E2E Magic Bow', camp: 'Red', char_role: 'magic_bow', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy E1', camp: 'Blue', char_role: 'enemy_char' }),
      playerInfo({ id: ENEMY_2_PLAYER_ID, name: 'Enemy E2', camp: 'Blue', char_role: 'enemy_2_char' }),
    ],
    initialState: syncState({
      turn_player_id: MB_PLAYER_ID,
      turn_stage: options.turnStage ?? 'ActionExecution',
      available_skills: options.availableSkills ?? [],
      characters: defaultCharacters,
      players,
    }),
  };
}

function skillChoicePrompt(skillId: string, title: string, message: string): WsMessage {
  return requireActionMessage({
    type: 'choose_skill',
    player_id: MB_PLAYER_ID,
    message,
    options: [
      { id: skillId, label: title, button_label: '发动', hint: `发动【${title}】` },
      { id: 'skip', label: '跳过', button_label: '跳过', hint: '不发动响应技能' },
    ],
    presentation: { kind: 'skill_choice', layout: 'overlay', numeric_base: 0 },
    min: 1,
    max: 1,
  } satisfies Prompt);
}

export function magicPierceResponsePrompt(): WsMessage {
  return skillChoicePrompt(
    MB_MAGIC_PIERCE_SKILL_ID,
    '魔贯冲击',
    '你触发了响应技能【魔贯冲击】，请选择是否发动。'
  );
}

export function magicPierceChargePrompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_cards',
    player_id: MB_PLAYER_ID,
    message: '【魔贯冲击】请选择移除1个火系充能：',
    choice_type: 'mb_magic_pierce_charge',
    skill_id: MB_MAGIC_PIERCE_SKILL_ID,
    options: [
      { id: '0', label: '火焰充能', button_label: '移除充能', field_index: 0, card_id: 'mb-charge-0' },
    ],
    min: 1,
    max: 1,
    interaction: fieldOptionIndexInteraction,
    presentation: { kind: 'card_picker', layout: 'field_cover', card_source: 'field', card_filter: 'effect:MagicBowCharge', numeric_base: 0 },
  } satisfies Prompt);
}

export function magicPierceHitBonusPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: MB_PLAYER_ID,
    message: '【魔贯冲击】是否额外移除1个火系充能使伤害+1？',
    choice_type: 'mb_magic_pierce_hit_bonus',
    skill_id: MB_MAGIC_PIERCE_SKILL_ID,
    presentation: { kind: 'branch_select', layout: 'overlay', numeric_base: 0 },
    options: [
      { id: '0', label: '是', button_label: '是' },
      { id: '1', label: '否', button_label: '否' },
    ],
    min: 1,
    max: 1,
  } satisfies Prompt);
}

export function magicPierceHitChargePrompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_cards',
    player_id: MB_PLAYER_ID,
    message: '【魔贯冲击】请选择额外移除1个火系充能：',
    choice_type: 'mb_magic_pierce_hit_charge',
    skill_id: MB_MAGIC_PIERCE_SKILL_ID,
    options: [
      { id: '1', label: '备用火焰充能', button_label: '移除充能', field_index: 1, card_id: 'mb-charge-1' },
    ],
    min: 1,
    max: 1,
    interaction: fieldOptionIndexInteraction,
    presentation: { kind: 'card_picker', layout: 'field_cover', card_source: 'field', card_filter: 'effect:MagicBowCharge', numeric_base: 0 },
  } satisfies Prompt);
}

export function thunderScatterScenario(): ProtocolHarnessScenario {
  return magicBowScenario({
    field: [
      chargeCover(0, 'Thunder', '雷光充能A'),
      chargeCover(1, 'Thunder', '雷光充能B'),
      chargeCover(2, 'Thunder', '雷光充能C'),
    ],
    availableSkills: [
      magicBowAvailableSkill({
        id: MB_THUNDER_SCATTER_SKILL_ID,
        title: '雷光散射',
        target_type: 2,
        min_targets: 0,
        max_targets: 1,
      }),
    ],
  });
}

export function thunderScatterBaseChargePrompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_cards',
    player_id: MB_PLAYER_ID,
    message: '【雷光散射】请选择移除1个雷系充能：',
    choice_type: 'mb_thunder_scatter_base_charge',
    skill_id: MB_THUNDER_SCATTER_SKILL_ID,
    options: [
      { id: '0', label: '雷光充能A', button_label: '移除充能', field_index: 0, card_id: 'mb-charge-0' },
      { id: '1', label: '雷光充能B', button_label: '移除充能', field_index: 1, card_id: 'mb-charge-1' },
      { id: '2', label: '雷光充能C', button_label: '移除充能', field_index: 2, card_id: 'mb-charge-2' },
    ],
    min: 1,
    max: 1,
    interaction: fieldOptionIndexInteraction,
    presentation: { kind: 'card_picker', layout: 'field_cover', card_source: 'field', card_filter: 'effect:MagicBowCharge', numeric_base: 0 },
  } satisfies Prompt);
}

export function thunderScatterExtraPrompt(maxExtra = 2): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: MB_PLAYER_ID,
    message: '【雷光散射】请选择额外移除雷系充能数量X：',
    choice_type: 'mb_thunder_scatter_extra',
    skill_id: MB_THUNDER_SCATTER_SKILL_ID,
    options: Array.from({ length: maxExtra + 1 }, (_, x) => ({
      id: String(x),
      label: `额外移除${x}个雷系充能`,
      button_label: String(x),
    })),
    presentation: { kind: 'numeric', numeric_base: 0 },
    min: 1,
    max: 1,
  } satisfies Prompt);
}

export function thunderScatterTargetPrompt(extraX = 2): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: MB_PLAYER_ID,
    message: `【雷光散射】请选择额外受到${extraX}点法术伤害的目标：`,
    choice_type: 'mb_thunder_scatter_target',
    skill_id: MB_THUNDER_SCATTER_SKILL_ID,
    presentation: { kind: 'target_picker', target_filter: 'custom', numeric_base: 0 },
    options: [
      { id: ENEMY_PLAYER_ID, target_id: ENEMY_PLAYER_ID, label: 'Enemy E1', button_label: '选择' },
      { id: ENEMY_2_PLAYER_ID, target_id: ENEMY_2_PLAYER_ID, label: 'Enemy E2', button_label: '选择' },
    ],
    min: 1,
    max: 1,
  } satisfies Prompt);
}

export function multiShotResponsePrompt(): WsMessage {
  return skillChoicePrompt(
    MB_MULTI_SHOT_SKILL_ID,
    '多重射击',
    '你触发了响应技能【多重射击】，请选择是否发动。'
  );
}

export function multiShotChargePrompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_cards',
    player_id: MB_PLAYER_ID,
    message: '【多重射击】请选择移除1个风系充能：',
    choice_type: 'mb_multi_shot_charge',
    skill_id: MB_MULTI_SHOT_SKILL_ID,
    options: [
      { id: '2', label: '风神充能', button_label: '移除充能', field_index: 2, card_id: 'mb-charge-2' },
    ],
    min: 1,
    max: 1,
    interaction: fieldOptionIndexInteraction,
    presentation: { kind: 'card_picker', layout: 'field_cover', card_source: 'field', card_filter: 'effect:MagicBowCharge', numeric_base: 0 },
  } satisfies Prompt);
}

export function multiShotTargetPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: MB_PLAYER_ID,
    message: '【多重射击】请选择暗系追加攻击目标：',
    choice_type: 'mb_multi_shot_target',
    skill_id: MB_MULTI_SHOT_SKILL_ID,
    presentation: { kind: 'target_picker', target_filter: 'custom', numeric_base: 0 },
    options: [
      { id: ENEMY_2_PLAYER_ID, target_id: ENEMY_2_PLAYER_ID, label: 'Enemy E2', button_label: '选择' },
    ],
    min: 1,
    max: 1,
  } satisfies Prompt);
}

export function chargeScenario(): ProtocolHarnessScenario {
  return magicBowScenario({
    hand: magicBowLargeHand(),
    crystal: 1,
    turnStage: 'ActionStart',
    availableSkills: [
      magicBowAvailableSkill({
        id: MB_CHARGE_SKILL_ID,
        title: '充能',
        cost_crystal: 1,
      }),
    ],
  });
}

export function chargeDiscardPrompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_cards',
    player_id: MB_PLAYER_ID,
    message: '【充能】请先弃置手牌至4张：',
    choice_type: 'system_discard_cards',
    options: [
      { id: 'mb-hand-light', label: '5: 圣光（光系 法术）', button_label: '选择', card_id: 'mb-hand-light' },
      { id: 'mb-hand-dark', label: '6: 魔弹（暗灭 法术）', button_label: '选择', card_id: 'mb-hand-dark' },
    ],
    presentation: { kind: 'card_picker', card_source: 'hand', card_filter: 'overflow_discard', numeric_base: 0 },
    min: 2,
    max: 2,
  } satisfies Prompt);
}

export function chargeDrawPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: MB_PLAYER_ID,
    message: '【充能】请选择摸牌数量X（0~4）：',
    choice_type: 'mb_charge_draw_x',
    skill_id: MB_CHARGE_SKILL_ID,
    options: Array.from({ length: 5 }, (_, x) => ({
      id: String(x),
      label: `X=${x}（摸${x}张）`,
      button_label: String(x),
    })),
    presentation: { kind: 'numeric', numeric_base: 0 },
    min: 1,
    max: 1,
  } satisfies Prompt);
}

// 新流程：直接进入盖牌多选（跳过数量选择步骤）
// maxPlace = X值（摸牌数量），玩家可选择 0~maxPlace 张手牌作为充能
export function chargePlaceCardsMultiSelectPrompt(maxPlace = 4): WsMessage {
  return requireActionMessage({
    type: 'choose_cards',
    player_id: MB_PLAYER_ID,
    message: `【充能】请选择要放置为充能的手牌（最多${maxPlace}张，可不选）：`,
    choice_type: 'mb_charge_place_cards',
    skill_id: MB_CHARGE_SKILL_ID,
    options: [
      { id: 'mb-atk-fire', label: '1: 火焰斩（火系 攻击）', button_label: '选择', card_id: 'mb-atk-fire' },
      { id: 'mb-hand-water', label: '2: 水涟斩（水系 攻击）', button_label: '选择', card_id: 'mb-hand-water' },
      { id: 'mb-hand-thunder', label: '3: 雷光斩（雷系 攻击）', button_label: '选择', card_id: 'mb-hand-thunder' },
      { id: 'mb-hand-wind', label: '4: 风神斩（风系 攻击）', button_label: '选择', card_id: 'mb-hand-wind' },
    ],
    presentation: { kind: 'card_picker', card_source: 'hand', card_filter: 'option_limited', numeric_base: 0 },
    min: 0, // 允许不选
    max: maxPlace, // 最多可选 maxPlace 张
  } satisfies Prompt);
}

// Count prompt fixture kept for targeted protocol-shape coverage; the main flow uses chargePlaceCardsMultiSelectPrompt.
export function chargePlaceCountPrompt(maxPlace = 3): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: MB_PLAYER_ID,
    message: '【充能】请选择要放置为充能的手牌数量：',
    choice_type: 'mb_charge_place_count',
    skill_id: MB_CHARGE_SKILL_ID,
    options: Array.from({ length: maxPlace + 1 }, (_, count) => ({
      id: String(count),
      label: `放置${count}张充能`,
      button_label: String(count),
    })),
    presentation: { kind: 'numeric', numeric_base: 0 },
    min: 1,
    max: 1,
  } satisfies Prompt);
}

export function chargePlaceCardsPrompt(count = 2): WsMessage {
  return requireActionMessage({
    type: 'choose_cards',
    player_id: MB_PLAYER_ID,
    message: `【充能】请选择${count}张作为充能的手牌：`,
    choice_type: 'mb_charge_place_cards',
    skill_id: MB_CHARGE_SKILL_ID,
    options: [
      { id: '0', label: '1: 火焰斩（火系 攻击）', button_label: '选择' },
      { id: '1', label: '2: 水涟斩（水系 攻击）', button_label: '选择' },
    ],
    presentation: { kind: 'card_picker', card_source: 'hand', card_filter: 'option_limited', numeric_base: 0 },
    min: count,
    max: count,
  } satisfies Prompt);
}

export function demonEyeScenario(): ProtocolHarnessScenario {
  return magicBowScenario({
    gem: 1,
    turnStage: 'ActionStart',
    availableSkills: [
      magicBowAvailableSkill({
        id: MB_DEMON_EYE_SKILL_ID,
        title: '魔眼',
        cost_gem: 1,
      }),
    ],
  });
}

export function demonEyeModePrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: MB_PLAYER_ID,
    message: '【魔眼】请选择发动分支：',
    choice_type: 'mb_demon_eye_mode',
    skill_id: MB_DEMON_EYE_SKILL_ID,
    options: [
      { id: '0', label: '分支①：令1名角色弃1张牌', button_label: '分支①' },
      { id: '1', label: '分支②：你摸3张牌', button_label: '分支②' },
    ],
    presentation: { kind: 'branch_select', layout: 'overlay', numeric_base: 0 },
    min: 1,
    max: 1,
  } satisfies Prompt);
}

export function demonEyeTargetPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: MB_PLAYER_ID,
    message: '【魔眼·分支①】请选择弃1张牌的目标角色：',
    choice_type: 'mb_demon_eye_target',
    skill_id: MB_DEMON_EYE_SKILL_ID,
    presentation: { kind: 'target_picker', target_filter: 'custom', numeric_base: 0 },
    options: [
      { id: MB_PLAYER_ID, target_id: MB_PLAYER_ID, label: 'E2E Magic Bow', button_label: '选择' },
      { id: ENEMY_PLAYER_ID, target_id: ENEMY_PLAYER_ID, label: 'Enemy E1', button_label: '选择' },
      { id: ENEMY_2_PLAYER_ID, target_id: ENEMY_2_PLAYER_ID, label: 'Enemy E2', button_label: '选择' },
    ],
    min: 1,
    max: 1,
  } satisfies Prompt);
}

export function demonEyeChargeCardPrompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_cards',
    player_id: MB_PLAYER_ID,
    message: '【魔眼】请选择1张手牌作为充能：',
    choice_type: 'mb_demon_eye_charge_card',
    skill_id: MB_DEMON_EYE_SKILL_ID,
    options: [
      { id: 'mb-atk-fire', label: '1: 火焰斩（火系 攻击）', button_label: '选择', card_id: 'mb-atk-fire' },
      { id: 'mb-hand-water', label: '2: 水涟斩（水系 攻击）', button_label: '选择', card_id: 'mb-hand-water' },
    ],
    presentation: { kind: 'card_picker', card_source: 'hand', card_filter: 'option_limited', numeric_base: 0 },
    min: 1,
    max: 1,
  } satisfies Prompt);
}
