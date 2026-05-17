// ============================================================
// Butterfly Dancer (蝶舞者) Protocol Harness Scenarios
// ============================================================

import type { AvailableSkill, Card, FieldCard, Prompt } from '../../src/types/game';
import type { WsMessage } from '../../src/network/protocol';
import {
  availableSkill,
  card,
  characterView,
  playerInfo,
  playerView,
  requireActionMessage,
  syncState,
  syncStateMessage,
  type ProtocolHarnessScenario,
} from './builders';

export const BD_PLAYER_ID = 'bd_player';
export const ENEMY_PLAYER_ID = 'enemy_1';
export const ALLY_PLAYER_ID = 'ally_1';

export const BD_DANCE_SKILL_ID = 'bt_dance';
export const BD_CHRYSALIS_SKILL_ID = 'bt_chrysalis';
export const BD_LIFEFIRE_SKILL_ID = 'bt_life_fire';
export const BD_REVERSE_SKILL_ID = 'bt_reverse_butterfly';
export const BD_POISON_SKILL_ID = 'bt_poison_powder';
export const BD_PILGRIMAGE_SKILL_ID = 'bt_pilgrimage';
export const BD_MIRROR_SKILL_ID = 'bt_mirror';
export const BD_WITHER_SKILL_ID = 'bt_wither';

const butterflyDancerCharacter = characterView({
  id: 'butterfly_dancer',
  name: '蝶舞者',
  title: '蝶',
  faction: '星杯',
  skills: [
    {
      id: BD_DANCE_SKILL_ID,
      title: '舞动',
      description: '选择：摸1张牌（强制）或弃1张牌（强制）；然后将牌库顶1张牌面朝下放置为茧。',
      type: 2,
      min_targets: 0, max_targets: 0, target_type: 0,
    },
    {
      id: BD_CHRYSALIS_SKILL_ID,
      title: '蛹化',
      description: '你+1蛹，并将牌库顶4张牌面朝下放置为茧。',
      type: 2,
      min_targets: 0, max_targets: 0, target_type: 0,
      cost_gem: 1,
    },
    {
      id: BD_REVERSE_SKILL_ID,
      title: '倒逆之蝶',
      description: '弃2张牌，选择分支①或分支②。',
      type: 2,
      min_targets: 0, max_targets: 0, target_type: 0,
      cost_crystal: 1,
      cost_discards: 2,
    },
    {
      id: BD_POISON_SKILL_ID,
      title: '毒粉',
      description: '每当有角色产生1点实际法术伤害时，可移除1个茧，使该次伤害额外+1。',
      type: 3,
      min_targets: 0, max_targets: 0, target_type: 0,
    },
    {
      id: BD_PILGRIMAGE_SKILL_ID,
      title: '朝圣',
      description: '每当你承受伤害时，可移除1个茧，抵御1点该来源伤害。',
      type: 3,
      min_targets: 0, max_targets: 0, target_type: 0,
    },
    {
      id: BD_MIRROR_SKILL_ID,
      title: '镜花水月',
      description: '每当有角色产生2点实际法术伤害时，可移除2张同系茧并展示：抵御该次伤害。',
      type: 3,
      min_targets: 0, max_targets: 0, target_type: 0,
    },
    {
      id: BD_WITHER_SKILL_ID,
      title: '凋零',
      description: '你每次移除茧时，若该茧为法术牌，可展示并发动：对目标造成1点法术伤害。',
      type: 3,
      min_targets: 0, max_targets: 0, target_type: 0,
    },
  ],
});

const enemyCharacter = characterView({
  id: 'enemy_char', name: '守卫', title: '测试目标', faction: '异端', skills: [],
});

const allyCharacter = characterView({
  id: 'ally_char', name: '勇者', title: '测试队友', faction: '星杯', skills: [],
});

const defaultCharacters = [butterflyDancerCharacter, enemyCharacter, allyCharacter];

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function bdHand(): Card[] {
  return [
    card({ id: 'bd-atk-fire', name: '火焰斩', type: 'Attack', element: 'Fire', damage: 2 }),
    card({ id: 'bd-atk-thunder', name: '雷光斩', type: 'Attack', element: 'Thunder' }),
    card({ id: 'bd-magic-light', name: '圣光', type: 'Magic', element: 'Light' }),
    card({ id: 'bd-atk-water', name: '水涟斩', type: 'Attack', element: 'Water' }),
  ];
}

function cocoonCover(fieldIndex: number, element: Card['element'], name: string, isMagic = false): FieldCard {
  return {
    card: card({
      id: `bd-cocoon-${fieldIndex}`,
      name,
      type: isMagic ? 'Magic' : 'Attack',
      element,
    }),
    owner_id: BD_PLAYER_ID,
    source_id: BD_PLAYER_ID,
    mode: 'Cover',
    effect: 'ButterflyCocoon',
    field_hook: 'Manual',
    locked: false,
    duration: 0,
  };
}

function makeCocoons(elements: Array<{ el: Card['element']; name: string; magic?: boolean }>): FieldCard[] {
  return elements.map((e, i) => cocoonCover(i, e.el, e.name, e.magic));
}

function bdAvailableSkill(skill: Partial<AvailableSkill> & { id: string; title: string }): AvailableSkill {
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

export function butterflyDancerScenario(options: {
  hand?: Card[];
  field?: FieldCard[];
  availableSkills?: AvailableSkill[];
  crystal?: number;
  gem?: number;
  turnStage?: string;
  tokens?: Record<string, number>;
} = {}): ProtocolHarnessScenario {
  const hand = options.hand ?? bdHand();
  const players = [
    playerView({
      id: BD_PLAYER_ID,
      name: 'E2E Butterfly Dancer',
      camp: 'Red',
      role: 'butterfly_dancer',
      hand,
      hand_count: hand.length,
      field: options.field ?? [],
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
    myPlayerId: BD_PLAYER_ID,
    myPlayerName: 'E2E Butterfly Dancer',
    characters: defaultCharacters,
    players: [
      playerInfo({ id: BD_PLAYER_ID, name: 'E2E Butterfly Dancer', camp: 'Red', char_role: 'butterfly_dancer', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy E1', camp: 'Blue', char_role: 'enemy_char' }),
      playerInfo({ id: ALLY_PLAYER_ID, name: 'Ally A1', camp: 'Red', char_role: 'ally_char' }),
    ],
    initialState: syncState({
      turn_player_id: BD_PLAYER_ID,
      turn_stage: options.turnStage ?? 'ActionExecution',
      available_skills: options.availableSkills ?? [],
      characters: defaultCharacters,
      players,
    }),
  };
}

// ============================================================
// Dance (舞动) prompts
// ============================================================

export function danceScenario(options: {
  canDiscard?: boolean;
  hand?: Card[];
} = {}): ProtocolHarnessScenario {
  return butterflyDancerScenario({
    hand: options.hand ?? bdHand(),
    availableSkills: [
      bdAvailableSkill({ id: BD_DANCE_SKILL_ID, title: '舞动' }),
    ],
  });
}

export function danceModePrompt(canDiscard = true): WsMessage {
  const options: Array<{ id: string; label: string }> = [
    { id: '0', label: '摸1张牌' },
  ];
  if (canDiscard) {
    options.push({ id: '1', label: '弃1张牌' });
  }
  return requireActionMessage({
    type: 'confirm',
    player_id: BD_PLAYER_ID,
    message: '【舞动】请选择先执行的动作：',
    choice_type: 'bt_dance_mode',
    skill_id: BD_DANCE_SKILL_ID,
    options,
    presentation: { kind: 'branch_select', layout: 'overlay' },
    min: 1, max: 1,
  } satisfies Prompt);
}

export function danceDiscardPrompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_cards',
    player_id: BD_PLAYER_ID,
    message: '【舞动】请选择要弃置的1张手牌：',
    choice_type: 'bt_dance_discard',
    skill_id: BD_DANCE_SKILL_ID,
    options: [
      { id: '0', label: '1: 火焰斩（火系 攻击）' },
      { id: '1', label: '2: 雷光斩（雷系 攻击）' },
      { id: '2', label: '3: 圣光（光系 法术）' },
      { id: '3', label: '4: 水涟斩（水系 攻击）' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}

export function cocoonOverflowDiscardPrompt(discardCount = 1): WsMessage {
  return requireActionMessage({
    type: 'choose_cards',
    player_id: BD_PLAYER_ID,
    message: `【茧上限】请选择要舍弃的${discardCount}个茧：`,
    choice_type: 'bt_cocoon_overflow_discard',
    options: [
      { id: '0', label: '茧[0]: 茧牌A（火系 攻击）' },
    ],
    min: discardCount, max: discardCount,
  } satisfies Prompt);
}

// ============================================================
// Reverse Butterfly (倒逆之蝶) prompts
// ============================================================

export function reverseScenario(options: {
  canBranch2?: boolean;
  crystal?: number;
  cocoons?: FieldCard[];
} = {}): ProtocolHarnessScenario {
  return butterflyDancerScenario({
    hand: [
      ...bdHand(),
      card({ id: 'bd-extra-1', name: '备牌A', type: 'Attack', element: 'Earth' }),
      card({ id: 'bd-extra-2', name: '备牌B', type: 'Attack', element: 'Wind' }),
    ],
    field: options.cocoons ?? makeCocoons([
      { el: 'Fire', name: '茧牌A' },
      { el: 'Water', name: '茧牌B' },
    ]),
    crystal: options.crystal ?? 1,
    tokens: options.canBranch2 ? { bt_pupa: 1, bt_cocoon_count: 2 } : { bt_cocoon_count: 2 },
    availableSkills: [
      bdAvailableSkill({ id: BD_REVERSE_SKILL_ID, title: '倒逆之蝶', cost_crystal: 1, cost_discards: 2 }),
    ],
  });
}

/**
 * 倒逆之蝶的弃 2 张费用前置弹框。
 * 后端契约：技能确认后引擎统一通过 InteractionDiscard 流程下发
 * choice_type: 'system_discard_cards' 的弃牌中断，玩家完成弃 2 张后
 * 才进入 bt_reverse_mode 分支选择
 * （见 butterfly_dancer_regression_test.go TestButterflyReverse_UsesUnifiedDiscardCostBeforeBranchChoice）。
 */
export function reverseDiscardPrompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_cards',
    player_id: BD_PLAYER_ID,
    message: '【倒逆之蝶】请选择要弃置的2张手牌：',
    choice_type: 'system_discard_cards',
    skill_id: BD_REVERSE_SKILL_ID,
    options: [
      { id: '0', label: '1: 火焰斩（火系 攻击）' },
      { id: '1', label: '2: 雷光斩（雷系 攻击）' },
      { id: '2', label: '3: 圣光（光系 法术）' },
      { id: '3', label: '4: 水涟斩（水系 攻击）' },
      { id: '4', label: '5: 备牌A（地系 攻击）' },
      { id: '5', label: '6: 备牌B（风系 攻击）' },
    ],
    min: 2,
    max: 2,
  } satisfies Prompt);
}

export function reverseModePrompt(canBranch2 = true): WsMessage {
  const options: Array<{ id: string; label: string }> = [
    { id: '0', label: '分支①：对目标造成1点不可治疗抵御的法术伤害' },
  ];
  if (canBranch2) {
    options.push({ id: '1', label: '分支②：移除2个茧或自伤4，然后移除1个蛹' });
  }
  return requireActionMessage({
    type: 'confirm',
    player_id: BD_PLAYER_ID,
    message: '【倒逆之蝶】请选择发动分支：',
    choice_type: 'bt_reverse_mode',
    skill_id: BD_REVERSE_SKILL_ID,
    options,
    presentation: { kind: 'branch_select', layout: 'overlay' },
    min: 1, max: 1,
  } satisfies Prompt);
}

export function reverseTargetPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: BD_PLAYER_ID,
    message: '【倒逆之蝶】请选择分支①伤害目标：',
    choice_type: 'bt_reverse_target',
    skill_id: BD_REVERSE_SKILL_ID,
    options: [
      { id: ENEMY_PLAYER_ID, label: 'Enemy E1' },
      { id: ALLY_PLAYER_ID, label: 'Ally A1' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}

export function reverseBranch2CostPrompt(canRemoveCocoon = true): WsMessage {
  const options: Array<{ id: string; label: string }> = [];
  if (canRemoveCocoon) {
    options.push({ id: '0', label: '移除2个茧' });
  }
  options.push({ id: '1', label: '对自己造成4点法术伤害' });
  return requireActionMessage({
    type: 'confirm',
    player_id: BD_PLAYER_ID,
    message: '【倒逆之蝶】请选择分支②代价：',
    choice_type: 'bt_reverse_branch2_cost',
    skill_id: BD_REVERSE_SKILL_ID,
    options,
    min: 1, max: 1,
  } satisfies Prompt);
}

export function reverseBranch2PickPrompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_cards',
    player_id: BD_PLAYER_ID,
    message: '【倒逆之蝶】分支②请选择要移除的2个茧：',
    choice_type: 'bt_reverse_branch2_pick',
    skill_id: BD_REVERSE_SKILL_ID,
    options: [
      { id: '0', label: '茧[0]: 茧牌A（火系 攻击）' },
      { id: '1', label: '茧[1]: 茧牌B（水系 魔术）' },
    ],
    min: 2, max: 2,
  } satisfies Prompt);
}

// ============================================================
// Pilgrimage / Poison (朝圣/毒粉) cocoon pick prompts
// ============================================================

export function pilgrimagePickPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: BD_PLAYER_ID,
    message: '【朝圣】是否移除1个茧抵御1点伤害？',
    choice_type: 'bt_pilgrimage_pick',
    options: [
      { id: '-1', label: '不发动' },
      {
        id: '1',
        label: '移除茧[0]: 茧牌A（火系 攻击）',
        button_label: '移除茧[0]',
        hint: '茧牌A（火系 攻击）',
        field_index: 0,
      },
    ],
    min: 1, max: 1,
    presentation: { kind: 'branch_select', layout: 'overlay', numeric_base: 0 },
  } satisfies Prompt);
}

export function poisonPickPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: BD_PLAYER_ID,
    message: '【毒粉】是否移除1个茧使该次法术伤害+1？',
    choice_type: 'bt_poison_pick',
    options: [
      { id: '-1', label: '不发动' },
      {
        id: '1',
        label: '移除茧[0]: 茧牌A（火系 攻击）',
        button_label: '移除茧[0]',
        hint: '茧牌A（火系 攻击）',
        field_index: 0,
      },
    ],
    min: 1, max: 1,
    presentation: { kind: 'branch_select', layout: 'overlay', numeric_base: 0 },
  } satisfies Prompt);
}

// ============================================================
// Mirror (镜花水月) pair selection prompt
// ============================================================

export function mirrorPairPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: BD_PLAYER_ID,
    message: '【镜花水月】是否发动并改写该次2点法术伤害？',
    choice_type: 'bt_mirror_pair',
    options: [
      { id: '-1', label: '不发动' },
      { id: '1', label: '移除并展示：火系茧：茧牌A（火系 攻击） + 茧牌B（火系 攻击）' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}

// ============================================================
// Wither (凋零) prompts
// ============================================================

export function witherScenario(options: {
  tokens?: Record<string, number>;
} = {}): ProtocolHarnessScenario {
  return butterflyDancerScenario({
    tokens: options.tokens ?? {},
  });
}

export function witherConfirmPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: BD_PLAYER_ID,
    message: '【凋零】可发动：是否对目标造成1点法术伤害并对自己造成2点法术伤害？',
    choice_type: 'bt_wither_confirm',
    options: [
      { id: '0', label: '发动凋零' },
      { id: '1', label: '不发动' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}

export function witherTargetPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: BD_PLAYER_ID,
    message: '【凋零】请选择1名目标角色：',
    choice_type: 'bt_wither_target',
    options: [
      { id: ENEMY_PLAYER_ID, label: 'Enemy E1' },
      { id: ALLY_PLAYER_ID, label: 'Ally A1' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}

// ============================================================
// Chrysalis (蛹化) scenario
// ============================================================

export function chrysalisScenario(): ProtocolHarnessScenario {
  return butterflyDancerScenario({
    gem: 1,
    availableSkills: [
      bdAvailableSkill({ id: BD_CHRYSALIS_SKILL_ID, title: '蛹化', cost_gem: 1 }),
    ],
  });
}

/**
 * 蛹化结算后的状态同步消息。
 * 后端契约：技能激活后，后端自动执行以下操作：
 * 1. AddPupa(user, 1) → tokens.bt_pupa = 1
 * 2. DrawRawCards(4) 并放置为茧 → field 包含4张茧牌
 * 3. 如果茧未溢出（<=8），无后续交互弹框
 */
export function chrysalisResolvedState(options: {
  pupaCount?: number;
  cocoonCards?: Array<{ element: Card['element']; name: string; isMagic?: boolean }>;
  cocoonOverflow?: number;
} = {}): WsMessage {
  const pupaCount = options.pupaCount ?? 1;
  const cocoonCards = options.cocoonCards ?? [
    { element: 'Fire', name: '茧牌A' },
    { element: 'Water', name: '茧牌B' },
    { element: 'Wind', name: '茧牌C' },
    { element: 'Earth', name: '茧牌D' },
  ];
  const cocoonCount = cocoonCards.length;

  const field: FieldCard[] = cocoonCards.map((c, i) => cocoonCover(i, c.element, c.name, c.isMagic));

  const tokens: Record<string, number> = {
    bt_pupa: pupaCount,
    bt_cocoon_count: cocoonCount,
  };

  // 更新后的玩家状态
  const updatedPlayers = [
    playerView({
      id: BD_PLAYER_ID,
      name: 'E2E Butterfly Dancer',
      camp: 'Red',
      role: 'butterfly_dancer',
      hand: bdHand(),
      hand_count: bdHand().length,
      field,
      gem: 0, // 消耗1宝石
      crystal: 0,
      is_active: true,
      tokens,
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

  return syncStateMessage(syncState({
    turn_player_id: BD_PLAYER_ID,
    turn_stage: 'ActionExecution',
    available_skills: [],
    characters: defaultCharacters,
    players: updatedPlayers,
    deck_count: 138, // 减少4张
  }));
}

/**
 * 蛹化茧溢出时的弃茧弹框。
 * 后端契约：当茧数量超过上限（8张），推送 bt_cocoon_overflow_discard 弹框。
 */
export function chrysalisOverflowDiscardPrompt(overflowCount: number, cocoonLabels: Array<{ id: string; label: string }>): WsMessage {
  return requireActionMessage({
    type: 'choose_cards',
    player_id: BD_PLAYER_ID,
    message: `【茧上限】请选择要舍弃的${overflowCount}个茧：`,
    choice_type: 'bt_cocoon_overflow_discard',
    options: cocoonLabels,
    min: overflowCount,
    max: overflowCount,
  } satisfies Prompt);
}
