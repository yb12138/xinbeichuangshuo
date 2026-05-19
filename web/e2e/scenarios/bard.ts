import type { AvailableSkill, Prompt, PromptOption } from '../../src/types/game';
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

// ---- Player IDs ----
export const BARD_PLAYER_ID = 'bard_player';
export const ALLY_PLAYER_ID = 'ally_1';
export const ENEMY_PLAYER_ID = 'enemy_1';
export const ENEMY_2_PLAYER_ID = 'enemy_2';

// ---- Skill IDs ----
export const BD_HOPE_FUGUE_SKILL_ID = 'bd_hope_fugue';
export const BD_ROUSING_RHAPSODY_SKILL_ID = 'bd_rousing_rhapsody';
export const BD_VICTORY_SYMPHONY_SKILL_ID = 'bd_victory_symphony';
export const BD_DESCENT_CONCERTO_SKILL_ID = 'bd_descent_concerto';
export const BD_DISSONANCE_CHORD_SKILL_ID = 'bd_dissonance_chord';

// ---- Bard character definition ----
const bardCharacter = characterView({
  id: 'bard',
  name: '吟游诗人',
  title: '永恒乐章',
  faction: '星杯',
  skills: [
    {
      id: BD_HOPE_FUGUE_SKILL_ID,
      title: '希望赋格曲',
      description: '将永恒乐章放置于目标队友面前，或转移永恒乐章并弃1张牌+1治疗/灵感',
      type: 2,
      min_targets: 0,
      max_targets: 0,
      target_type: 0,
      cost_gem: 0,
      cost_crystal: 0,
      cost_discards: 0,
    },
    {
      id: BD_ROUSING_RHAPSODY_SKILL_ID,
      title: '激昂狂想曲',
      description: '对2名对手各造成1点法术伤害，或弃2张牌',
      type: 2,
      min_targets: 0,
      max_targets: 0,
      target_type: 0,
      cost_gem: 0,
      cost_crystal: 0,
      cost_discards: 0,
    },
    {
      id: BD_VICTORY_SYMPHONY_SKILL_ID,
      title: '胜利交响诗',
      description: '提炼1个星石为能量，或我方战绩区+1宝石你+1治疗',
      type: 2,
      min_targets: 0,
      max_targets: 0,
      target_type: 0,
      cost_gem: 0,
      cost_crystal: 0,
      cost_discards: 0,
    },
    {
      id: BD_DESCENT_CONCERTO_SKILL_ID,
      title: '沉沦协奏曲',
      description: '弃2张同系牌+1灵感，若含魔法牌追加1点法术伤害',
      type: 2,
      min_targets: 0,
      max_targets: 0,
      target_type: 0,
      cost_gem: 0,
      cost_crystal: 0,
      cost_discards: 0,
    },
    {
      id: BD_DISSONANCE_CHORD_SKILL_ID,
      title: '不谐和弦',
      description: '消耗X灵感，与目标各摸/弃X-1张牌',
      type: 2,
      min_targets: 0,
      max_targets: 1,
      target_type: 0,
      cost_gem: 0,
      cost_crystal: 0,
      cost_discards: 0,
    },
  ],
});

const allyCharacter = characterView({
  id: 'hero',
  name: '勇者',
  title: '基础角色',
  faction: '星杯',
  skills: [],
});

const enemyCharacter = characterView({
  id: 'villain',
  name: '恶徒',
  title: '基础角色',
  faction: '星杯',
  skills: [],
});

// ---- Shared card factory ----
function bardHand() {
  return [
    card({ id: 'fire-atk-1', name: '火焰斩', type: 'Attack', element: 'Fire' }),
    card({ id: 'fire-atk-2', name: '火焰斩', type: 'Attack', element: 'Fire' }),
    card({ id: 'water-mgc-1', name: '水涟斩', type: 'Magic', element: 'Water' }),
    card({ id: 'water-mgc-2', name: '寒冰箭', type: 'Magic', element: 'Water' }),
    card({ id: 'light-mgc-1', name: '圣光', type: 'Magic', element: 'Light' }),
  ];
}

function allyHand() {
  return [
    card({ id: 'ally-card-1', name: '测试牌', type: 'Attack', element: 'Fire' }),
    card({ id: 'ally-card-2', name: '测试牌', type: 'Attack', element: 'Water' }),
  ];
}

function enemyHand() {
  return [
    card({ id: 'enemy-card-1', name: '测试牌', type: 'Attack', element: 'Earth' }),
  ];
}

function enemy2Hand() {
  return [
    card({ id: 'enemy2-card-1', name: '测试牌', type: 'Attack', element: 'Wind' }),
  ];
}

// ---- Shared players ----
function bardPlayerView(overrides: Partial<ReturnType<typeof playerView>> = {}) {
  const hand = bardHand();
  return playerView({
    id: BARD_PLAYER_ID,
    name: 'E2E Bard',
    camp: 'Red',
    role: 'bard',
    hand,
    hand_count: hand.length,
    heal: 3,
    max_heal: 5,
    gem: 0,
    crystal: 0,
    is_active: true,
    tokens: { bd_inspiration: 0 },
    ...overrides,
  });
}

function allyPlayerView() {
  const hand = allyHand();
  return playerView({
    id: ALLY_PLAYER_ID,
    name: 'Ally Hero',
    camp: 'Red',
    role: 'hero',
    hand,
    hand_count: hand.length,
    heal: 2,
    max_heal: 3,
    is_active: false,
  });
}

function enemyPlayerView(overrides: Partial<ReturnType<typeof playerView>> = {}) {
  const hand = enemyHand();
  return playerView({
    id: ENEMY_PLAYER_ID,
    name: 'Enemy Bot',
    camp: 'Blue',
    role: 'villain',
    hand,
    hand_count: hand.length,
    heal: 0,
    max_heal: 2,
    is_active: false,
    ...overrides,
  });
}

function enemy2PlayerView() {
  const hand = enemy2Hand();
  return playerView({
    id: ENEMY_2_PLAYER_ID,
    name: 'Enemy Bot 2',
    camp: 'Blue',
    role: 'villain',
    hand,
    hand_count: hand.length,
    heal: 1,
    max_heal: 2,
    is_active: false,
  });
}

// ---- Shared player infos ----
function bardPlayerInfo(overrides: Partial<ReturnType<typeof playerInfo>> = {}) {
  return playerInfo({
    id: BARD_PLAYER_ID,
    name: 'E2E Bard',
    camp: 'Red',
    char_role: 'bard',
    is_host: true,
    ...overrides,
  });
}

function allyPlayerInfo() {
  return playerInfo({
    id: ALLY_PLAYER_ID,
    name: 'Ally Hero',
    camp: 'Red',
    char_role: 'hero',
    is_bot: false,
  });
}

function enemyPlayerInfo(overrides: Partial<ReturnType<typeof playerInfo>> = {}) {
  return playerInfo({
    id: ENEMY_PLAYER_ID,
    name: 'Enemy Bot',
    camp: 'Blue',
    char_role: 'villain',
    is_bot: true,
    ...overrides,
  });
}

function enemy2PlayerInfo() {
  return playerInfo({
    id: ENEMY_2_PLAYER_ID,
    name: 'Enemy Bot 2',
    camp: 'Blue',
    char_role: 'villain',
    is_bot: true,
  });
}

// ============================================================
// 希望赋格曲 (bd_hope_fugue)
// ============================================================

export function hopeFugueScenario(options: {
  /** Whether eternal movement is currently held by another player (enables transfer branches) */
  hasEternalHolder?: boolean;
  /** Whether the bard has enough hand cards for transfer discard */
  hasHandCards?: boolean;
} = {}): ProtocolHarnessScenario {
  const characters = [bardCharacter, allyCharacter, enemyCharacter];
  const bard = bardPlayerView();
  const ally = allyPlayerView();

  // If eternal is already held, ally holds it
  if (options.hasEternalHolder) {
    ally.exclusive_cards = [
      card({
        id: 'starter-bard-bd_eternal_movement',
        name: '永恒乐章',
        type: 'Magic',
        element: 'Dark',
        description: '吟游诗人开局自带专属牌',
      }),
    ];
  }

  const players = [bard, ally, enemyPlayerView()];
  const playerInfos = [bardPlayerInfo(), allyPlayerInfo(), enemyPlayerInfo()];

  return {
    roomCode: 'MOCK',
    myPlayerId: BARD_PLAYER_ID,
    myPlayerName: 'E2E Bard',
    characters,
    players: playerInfos,
    initialState: syncState({
      turn_player_id: BARD_PLAYER_ID,
      turn_stage: 'ActionExecution',
      available_skills: [hopeFugueSkill()],
      characters,
      players,
    }),
  };
}

function hopeFugueSkill(): AvailableSkill {
  return availableSkill({
    id: BD_HOPE_FUGUE_SKILL_ID,
    title: '希望赋格曲',
    description: '将永恒乐章放置于目标队友面前，或转移永恒乐章并弃1张牌+1治疗/灵感',
    min_targets: 0,
    max_targets: 0,
    target_type: 0,
    cost_gem: 0,
    cost_crystal: 0,
    cost_discards: 0,
  });
}

export function hopeDrawConfirmPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: BARD_PLAYER_ID,
    message: '【希望赋格曲】是否先摸1张牌？',
    choice_type: 'bd_hope_draw_confirm',
    skill_id: BD_HOPE_FUGUE_SKILL_ID,
    options: [
      { id: '0', label: '是', button_label: '是' },
      { id: '1', label: '否', button_label: '否' },
    ],
    min: 1,
    max: 1,
    presentation: { kind: 'branch_select', layout: 'overlay', numeric_base: 0 },
  } satisfies Prompt);
}

export function hopeModePrompt(hasEternalHolder: boolean): WsMessage {
  const opts: PromptOption[] = [
    { id: '0', label: '将永恒乐章放置于目标队友面前', button_label: '放置' },
  ];
  if (hasEternalHolder) {
    opts.push(
      { id: '1', label: '转移永恒乐章，弃1张牌并+1治疗', button_label: '+治疗' },
      { id: '2', label: '转移永恒乐章，弃1张牌并+1灵感', button_label: '+灵感' },
    );
  }
  return requireActionMessage({
    type: 'confirm',
    player_id: BARD_PLAYER_ID,
    message: '【希望赋格曲】请选择分支：',
    choice_type: 'bd_hope_mode',
    skill_id: BD_HOPE_FUGUE_SKILL_ID,
    options: opts,
    presentation: { kind: 'branch_select', layout: 'overlay', numeric_base: 0 },
    min: 1,
    max: 1,
  } satisfies Prompt);
}

export function hopePlaceTargetPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: BARD_PLAYER_ID,
    message: '【希望赋格曲】请选择放置永恒乐章的目标队友：',
    choice_type: 'bd_hope_place_target',
    skill_id: BD_HOPE_FUGUE_SKILL_ID,
    options: [
      { id: ALLY_PLAYER_ID, label: '勇者', button_label: '选择' },
    ],
    min: 1,
    max: 1,
    presentation: { kind: 'target_picker', target_filter: 'custom', numeric_base: 0 },
  } satisfies Prompt);
}

export function hopeTransferTargetPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: BARD_PLAYER_ID,
    message: '【希望赋格曲】请选择转移永恒乐章的目标队友：',
    choice_type: 'bd_hope_transfer_target',
    skill_id: BD_HOPE_FUGUE_SKILL_ID,
    options: [
      { id: ALLY_PLAYER_ID, label: '勇者', button_label: '选择' },
    ],
    min: 1,
    max: 1,
    presentation: { kind: 'target_picker', target_filter: 'custom', numeric_base: 0 },
  } satisfies Prompt);
}

export function hopeTransferDiscardPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: BARD_PLAYER_ID,
    message: '【希望赋格曲】请选择弃置1张手牌：',
    choice_type: 'bd_hope_transfer_discard',
    skill_id: BD_HOPE_FUGUE_SKILL_ID,
    options: [
      { id: '0', label: '1: 火焰斩 (火 Attack)', button_label: '选择', card_id: 'fire-atk-1' },
      { id: '1', label: '2: 火焰斩 (火 Attack)', button_label: '选择', card_id: 'fire-atk-2' },
      { id: '2', label: '3: 水涟斩 (水 Magic)', button_label: '选择', card_id: 'water-mgc-1' },
      { id: '3', label: '4: 寒冰箭 (水 Magic)', button_label: '选择', card_id: 'water-mgc-2' },
      { id: '4', label: '5: 圣光 (光 Magic)', button_label: '选择', card_id: 'light-mgc-1' },
    ],
    min: 1,
    max: 1,
    presentation: { kind: 'card_picker', card_source: 'hand', card_filter: 'same_element', numeric_base: 0 },
  } satisfies Prompt);
}

// ============================================================
// 激昂狂想曲 (bd_rousing_rhapsody)
// ============================================================

export function rousingRhapsodyScenario(options: {
  playerHandCount?: number;
} = {}): ProtocolHarnessScenario {
  const characters = [bardCharacter, allyCharacter, enemyCharacter, enemyCharacter];

  const bard = bardPlayerView({ is_active: false });

  // Eternal holder is the "my player" who triggers rousing
  const eternalHolder = playerView({
    id: ALLY_PLAYER_ID,
    name: 'E2E Host',
    camp: 'Red',
    role: 'hero',
    hand: allyHand(),
    hand_count: options.playerHandCount ?? allyHand().length,
    heal: 2,
    max_heal: 3,
    is_active: true,
    exclusive_cards: [
      card({
        id: 'starter-bard-bd_eternal_movement',
        name: '永恒乐章',
        type: 'Magic',
        element: 'Dark',
        description: '永恒乐章专属牌',
      }),
    ],
  });

  const enemy = enemyPlayerView();
  const enemy2 = enemy2PlayerView();

  const players = [bard, eternalHolder, enemy, enemy2];
  const playerInfos = [
    bardPlayerInfo({ is_host: false }),
    playerInfo({
      id: ALLY_PLAYER_ID,
      name: 'E2E Host',
      camp: 'Red',
      char_role: 'hero',
      is_host: true,
    }),
    enemyPlayerInfo(),
    enemy2PlayerInfo(),
  ];

  return {
    roomCode: 'MOCK',
    myPlayerId: ALLY_PLAYER_ID,
    myPlayerName: 'E2E Host',
    characters,
    players: playerInfos,
    initialState: syncState({
      turn_player_id: ALLY_PLAYER_ID,
      turn_stage: 'ActionExecution',
      available_skills: [],
      characters,
      players,
    }),
  };
}

// 吟游诗人视角的目标选择场景（用于测试 bd_rousing_targets 流程）
export function rousingTargetSelectionScenario(): ProtocolHarnessScenario {
  const characters = [bardCharacter, allyCharacter, enemyCharacter, enemyCharacter];

  const bard = bardPlayerView({ is_active: true });

  // 永恒乐章持有者（队友）
  const eternalHolder = playerView({
    id: ALLY_PLAYER_ID,
    name: 'E2E Ally',
    camp: 'Red',
    role: 'hero',
    hand: allyHand(),
    hand_count: allyHand().length,
    heal: 2,
    max_heal: 3,
    is_active: false,
    exclusive_cards: [
      card({
        id: 'starter-bard-bd_eternal_movement',
        name: '永恒乐章',
        type: 'Magic',
        element: 'Dark',
        description: '永恒乐章专属牌',
      }),
    ],
  });

  const enemy = enemyPlayerView();
  const enemy2 = enemy2PlayerView();

  const players = [bard, eternalHolder, enemy, enemy2];
  const playerInfos = [
    bardPlayerInfo({ is_host: true }),
    playerInfo({
      id: ALLY_PLAYER_ID,
      name: 'E2E Ally',
      camp: 'Red',
      char_role: 'hero',
      is_host: false,
    }),
    enemyPlayerInfo(),
    enemy2PlayerInfo(),
  ];

  return {
    roomCode: 'MOCK',
    myPlayerId: BARD_PLAYER_ID,
    myPlayerName: 'E2E Bard',
    characters,
    players: playerInfos,
    initialState: syncState({
      turn_player_id: ALLY_PLAYER_ID,
      turn_stage: 'ActionExecution',
      available_skills: [],
      characters,
      players,
    }),
  };
}

export function rousingModePrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: ALLY_PLAYER_ID,
    message: '【激昂狂想曲】请选择效果：',
    choice_type: 'bd_rousing_mode',
    skill_id: BD_ROUSING_RHAPSODY_SKILL_ID,
    options: [
      { id: '0', label: '对2名对手各造成1点法术伤害', button_label: '伤害' },
      { id: '1', label: '弃2张牌', button_label: '弃牌' },
      { id: '2', label: '跳过', button_label: '跳过' },
    ],
    presentation: { kind: 'branch_select', layout: 'overlay', numeric_base: 0 },
    min: 1,
    max: 1,
  } satisfies Prompt);
}

export function rousingTargetsPrompt(step: number): WsMessage {
  const options: PromptOption[] = [];
  // Step 1: show all enemies; Step 2: show remaining
  if (step === 1) {
    options.push({ id: ENEMY_PLAYER_ID, label: '恶徒', button_label: '选择' });
    options.push({ id: ENEMY_2_PLAYER_ID, label: '恶徒2', button_label: '选择' });
  } else {
    options.push({ id: ENEMY_2_PLAYER_ID, label: '恶徒2', button_label: '选择' });
  }
  // 目标选择由吟游诗人执行（伤害来源是吟游诗人）
  return requireActionMessage({
    type: 'confirm',
    player_id: BARD_PLAYER_ID,
    message: `【激昂狂想曲】请选择第 ${step}/2 名目标：`,
    choice_type: 'bd_rousing_targets',
    skill_id: BD_ROUSING_RHAPSODY_SKILL_ID,
    options,
    min: 1,
    max: 1,
    presentation: { kind: 'target_picker', target_filter: 'custom', numeric_base: 0 },
  } satisfies Prompt);
}

export function rousingDiscardCardsPrompt(pickRemaining: number): WsMessage {
  return requireActionMessage({
    type: 'choose_cards',
    player_id: ALLY_PLAYER_ID,
    message: `【激昂狂想曲】请选择要弃置的${pickRemaining}张手牌：`,
    choice_type: 'bd_rousing_discard_cards',
    skill_id: BD_ROUSING_RHAPSODY_SKILL_ID,
    options: [
      { id: '0', label: '1: 测试牌 (火 Attack)', button_label: '选择', card_id: 'ally-card-1' },
      { id: '1', label: '2: 测试牌 (水 Attack)', button_label: '选择', card_id: 'ally-card-2' },
    ],
    min: pickRemaining,
    max: pickRemaining,
    presentation: { kind: 'card_picker', card_source: 'hand', card_filter: 'option_limited', numeric_base: 0 },
  } satisfies Prompt);
}

// ============================================================
// 胜利交响诗 (bd_victory_symphony)
// ============================================================

export function victorySymphonyScenario(options: {
  campHasGems?: boolean;
  campHasCrystals?: boolean;
} = {}): ProtocolHarnessScenario {
  const characters = [bardCharacter, allyCharacter, enemyCharacter];

  const bard = bardPlayerView({ is_active: false });

  const eternalHolder = playerView({
    id: ALLY_PLAYER_ID,
    name: 'E2E Host',
    camp: 'Red',
    role: 'hero',
    hand: allyHand(),
    hand_count: allyHand().length,
    heal: 1,
    max_heal: 3,
    is_active: true,
    exclusive_cards: [
      card({
        id: 'starter-bard-bd_eternal_movement',
        name: '永恒乐章',
        type: 'Magic',
        element: 'Dark',
        description: '永恒乐章专属牌',
      }),
    ],
  });

  const players = [bard, eternalHolder, playerView({
    id: ENEMY_PLAYER_ID,
    name: 'Enemy Bot',
    camp: 'Blue',
    role: 'villain',
    hand: [],
    hand_count: 3,
    heal: 0,
    max_heal: 2,
    is_active: false,
  })];

  return {
    roomCode: 'MOCK',
    myPlayerId: ALLY_PLAYER_ID,
    myPlayerName: 'E2E Host',
    characters,
    players: [
      bardPlayerInfo({ is_host: false }),
      playerInfo({ id: ALLY_PLAYER_ID, name: 'E2E Host', camp: 'Red', char_role: 'hero', is_host: true }),
      enemyPlayerInfo(),
    ],
    initialState: syncState({
      turn_player_id: ALLY_PLAYER_ID,
      turn_stage: 'ActionExecution',
      available_skills: [],
      characters,
      players,
      stones_red: [options.campHasGems ?? true ? 2 : 0, options.campHasCrystals ?? true ? 1 : 0],
      stones_blue: [1, 0],
    }),
  };
}

export function victoryConfirmPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: ALLY_PLAYER_ID,
    message: '【胜利交响诗】请选择效果：',
    choice_type: 'bd_victory_confirm',
    skill_id: BD_VICTORY_SYMPHONY_SKILL_ID,
    options: [
      { id: '0', label: '将我方战绩区1个星石提炼为你的能量', button_label: '提炼' },
      { id: '1', label: '我方战绩区+1宝石，你+1治疗', button_label: '+宝石' },
      { id: '2', label: '取消', button_label: '取消' },
    ],
    presentation: { kind: 'branch_select', layout: 'overlay', numeric_base: 0, cancel_policy: 'decline', has_decline: true, decline_index: 2 },
    min: 1,
    max: 1,
  } satisfies Prompt);
}

export function victoryModePrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: ALLY_PLAYER_ID,
    message: '【胜利交响诗】请选择效果：',
    choice_type: 'bd_victory_mode',
    skill_id: BD_VICTORY_SYMPHONY_SKILL_ID,
    options: [
      { id: '0', label: '将我方战绩区1个星石提炼为你的能量', button_label: '提炼' },
      { id: '1', label: '我方战绩区+1宝石，你+1治疗', button_label: '+宝石' },
    ],
    presentation: { kind: 'branch_select', layout: 'overlay', numeric_base: 0 },
    min: 1,
    max: 1,
  } satisfies Prompt);
}

export function victoryExtractStonePrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: ALLY_PLAYER_ID,
    message: '【胜利交响诗】请选择要提炼的星石：',
    choice_type: 'bd_victory_extract_stone',
    skill_id: BD_VICTORY_SYMPHONY_SKILL_ID,
    options: [
      { id: '0', label: '提炼1个宝石', button_label: '宝石' },
      { id: '1', label: '提炼1个水晶', button_label: '水晶' },
    ],
    min: 1,
    max: 1,
    presentation: { kind: 'branch_select', layout: 'overlay', numeric_base: 0 },
  } satisfies Prompt);
}

// ============================================================
// 沉沦协奏曲 (bd_descent_concerto)
// ============================================================

export function descentConcertoScenario(): ProtocolHarnessScenario {
  const characters = [bardCharacter, allyCharacter, enemyCharacter];

  // Descent prompt goes to the bard (not eternal holder). Bard needs >=2 same-element cards.
  const hand = bardHand();
  const bard = playerView({
    id: BARD_PLAYER_ID,
    name: 'E2E Bard',
    camp: 'Red',
    role: 'bard',
    hand,
    hand_count: hand.length,
    heal: 3,
    max_heal: 5,
    is_active: true,
  });

  const players = [bard, allyPlayerView(), enemyPlayerView()];

  return {
    roomCode: 'MOCK',
    myPlayerId: BARD_PLAYER_ID,
    myPlayerName: 'E2E Bard',
    characters,
    players: [bardPlayerInfo(), allyPlayerInfo(), enemyPlayerInfo()],
    initialState: syncState({
      turn_player_id: BARD_PLAYER_ID,
      turn_stage: 'ActionExecution',
      available_skills: [],
      characters,
      players,
    }),
  };
}

export function descentElementPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: BARD_PLAYER_ID,
    message: '【沉沦协奏曲】请选择要弃置的同系元素：',
    choice_type: 'bd_descent_element',
    skill_id: BD_DESCENT_CONCERTO_SKILL_ID,
    options: [
      { id: 'Fire', label: '火系', button_label: '火系' },
      { id: 'Water', label: '水系', button_label: '水系' },
    ],
    min: 1,
    max: 1,
    presentation: { kind: 'branch_select', layout: 'overlay', numeric_base: 0 },
  } satisfies Prompt);
}

// 新流程：直接推送弃牌选择，无需元素选择步骤
// candidateIndices 为所有有至少2张牌的元素系的手牌索引
export function descentCardsDirectPrompt(pickRemaining: number, candidateIndices: number[] = [0, 1, 2, 3]): WsMessage {
  // Build options from candidate indices (火系索引0-1，水系索引2-3)
  const hand = bardHand();
  const options: PromptOption[] = candidateIndices
    .filter((idx) => idx >= 0 && idx < hand.length)
    .map((idx) => ({
      id: String(idx),
      label: `${idx + 1}: ${hand[idx].name}`,
      button_label: '选择',
      card_id: hand[idx].id,
    }));

  return requireActionMessage({
    type: 'choose_cards',
    player_id: BARD_PLAYER_ID,
    message: `【沉沦协奏曲】请选择要弃置的${pickRemaining}张同系牌：`,
    choice_type: 'bd_descent_cards',
    skill_id: BD_DESCENT_CONCERTO_SKILL_ID,
    options,
    min: pickRemaining,
    max: pickRemaining,
    presentation: { kind: 'card_picker', card_source: 'hand', card_filter: 'same_element', numeric_base: 0 },
  } satisfies Prompt);
}

// 旧流程的保留函数（用于兼容性或回退测试）
export function descentCardsPrompt(chosenElement: string, pickRemaining: number): WsMessage {
  // Build options filtered by chosen element
  const allCards: { idx: number; name: string; element: string }[] = [];
  const hand = bardHand();
  for (let i = 0; i < hand.length; i++) {
    if (hand[i].element === chosenElement) {
      allCards.push({ idx: i, name: hand[i].name, element: hand[i].element });
    }
  }
  const options: PromptOption[] = allCards.map((c) => ({
    id: String(c.idx),
    label: `${c.idx + 1}: ${c.name}`,
    button_label: '选择',
    card_id: hand[c.idx].id,
  }));

  return requireActionMessage({
    type: 'choose_cards',
    player_id: BARD_PLAYER_ID,
    message: `【沉沦协奏曲】请选择要弃置的${pickRemaining}张${chosenElement === 'Fire' ? '火' : '水'}系牌：`,
    choice_type: 'bd_descent_cards',
    skill_id: BD_DESCENT_CONCERTO_SKILL_ID,
    options,
    min: pickRemaining,
    max: pickRemaining,
    presentation: { kind: 'card_picker', card_source: 'hand', card_filter: 'same_element', numeric_base: 0 },
  } satisfies Prompt);
}

export function descentTargetPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: BARD_PLAYER_ID,
    message: '【沉沦协奏曲】请选择1点法术伤害目标：',
    choice_type: 'bd_descent_target',
    skill_id: BD_DESCENT_CONCERTO_SKILL_ID,
    options: [
      { id: ENEMY_PLAYER_ID, label: '恶徒', button_label: '选择' },
    ],
    min: 1,
    max: 1,
    presentation: { kind: 'target_picker', target_filter: 'custom', numeric_base: 0 },
  } satisfies Prompt);
}

// ============================================================
// 不谐和弦 (bd_dissonance_chord)
// ============================================================

export function dissonanceChordScenario(options: {
  inspirationCount?: number;
} = {}): ProtocolHarnessScenario {
  const characters = [bardCharacter, allyCharacter, enemyCharacter];

  const inspiration = options.inspirationCount ?? 3; // Cap is 3
  const bard = bardPlayerView({
    tokens: { bd_inspiration: inspiration },
    is_active: true,
  });

  const enemy = enemyPlayerView({
    hand: [card({ id: 'enemy-card-1', name: '测试牌', type: 'Attack', element: 'Earth' })],
    hand_count: 1,
  });

  const players = [bard, allyPlayerView(), enemy];

  return {
    roomCode: 'MOCK',
    myPlayerId: BARD_PLAYER_ID,
    myPlayerName: 'E2E Bard',
    characters,
    players: [bardPlayerInfo(), allyPlayerInfo(), enemyPlayerInfo()],
    initialState: syncState({
      turn_player_id: BARD_PLAYER_ID,
      turn_stage: 'ActionExecution',
      available_skills: [dissonanceChordSkill()],
      characters,
      players,
    }),
  };
}

function dissonanceChordSkill(): AvailableSkill {
  return availableSkill({
    id: BD_DISSONANCE_CHORD_SKILL_ID,
    title: '不谐和弦',
    description: '消耗X灵感，与目标各摸/弃X-1张牌',
    min_targets: 0,
    max_targets: 1,
    target_type: 0,
    cost_gem: 0,
    cost_crystal: 0,
    cost_discards: 0,
  });
}

export function dissonanceXPrompt(maxX: number): WsMessage {
  const options: PromptOption[] = [];
  for (let x = 2; x <= maxX; x++) {
    options.push({ id: String(x), label: `X=${x}`, button_label: String(x) });
  }
  return requireActionMessage({
    type: 'confirm',
    player_id: BARD_PLAYER_ID,
    message: '【不谐和弦】请选择X值：',
    choice_type: 'bd_dissonance_x',
    skill_id: BD_DISSONANCE_CHORD_SKILL_ID,
    options,
    min: 1,
    max: 1,
    presentation: { kind: 'numeric', numeric_base: 0 },
  } satisfies Prompt);
}

export function dissonanceModePrompt(xValue: number): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: BARD_PLAYER_ID,
    message: `【不谐和弦】请选择分支（X=${xValue}）：`,
    choice_type: 'bd_dissonance_mode',
    skill_id: BD_DISSONANCE_CHORD_SKILL_ID,
    options: [
      { id: '0', label: `你与目标各摸${xValue - 1}张牌`, button_label: '摸牌' },
      { id: '1', label: `你与目标各弃${xValue - 1}张牌`, button_label: '弃牌' },
    ],
    presentation: { kind: 'branch_select', layout: 'overlay', numeric_base: 0 },
    min: 1,
    max: 1,
  } satisfies Prompt);
}

export function dissonanceTargetPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: BARD_PLAYER_ID,
    message: '【不谐和弦】请选择目标角色：',
    choice_type: 'bd_dissonance_target',
    skill_id: BD_DISSONANCE_CHORD_SKILL_ID,
    options: [
      { id: BARD_PLAYER_ID, label: '吟游诗人', button_label: '选择' },   // self
      { id: ALLY_PLAYER_ID, label: '勇者', button_label: '选择' },   // ally
      { id: ENEMY_PLAYER_ID, label: '恶徒', button_label: '选择' },   // enemy
    ],
    min: 1,
    max: 1,
    presentation: { kind: 'target_picker', target_filter: 'custom', numeric_base: 0 },
  } satisfies Prompt);
}

export function dissonanceDiscardStepPrompt(
  actorId: string,
  actorName: string,
  needCount: number,
): WsMessage {
  // Proxy card selection: when selecting cards for another player (enemy discard),
  // use a different choice_type that is excluded from hand card picker rendering
  const isProxySelection = actorId !== BARD_PLAYER_ID;
  const choiceType = isProxySelection ? 'bd_dissonance_discard_proxy' : 'bd_dissonance_discard_step';
  const promptType = isProxySelection ? 'confirm' : 'choose_cards';

  return requireActionMessage({
    type: promptType,
    player_id: BARD_PLAYER_ID,
    message: `【不谐和弦】请 ${actorName} 选择要弃置的${needCount}张手牌：`,
    choice_type: choiceType,
    skill_id: BD_DISSONANCE_CHORD_SKILL_ID,
    options: actorId === BARD_PLAYER_ID
      ? [
          { id: '0', label: '1: 火焰斩 (火 Attack)', button_label: '选择', card_id: 'fire-atk-1' },
          { id: '1', label: '2: 火焰斩 (火 Attack)', button_label: '选择', card_id: 'fire-atk-2' },
          { id: '2', label: '3: 水涟斩 (水 Magic)', button_label: '选择', card_id: 'water-mgc-1' },
          { id: '3', label: '4: 寒冰箭 (水 Magic)', button_label: '选择', card_id: 'water-mgc-2' },
          { id: '4', label: '5: 圣光 (光 Magic)', button_label: '选择', card_id: 'light-mgc-1' },
        ]
      : [
          { id: '0', label: '1: 测试牌 (地 Attack)', button_label: '选择' },
        ],
    min: needCount,
    max: needCount,
    presentation: actorId === BARD_PLAYER_ID
      ? { kind: 'card_picker', card_source: 'hand', card_filter: 'option_limited', numeric_base: 0 }
      : { kind: 'branch_select', layout: 'overlay', numeric_base: 0 },
  } satisfies Prompt);
}

// ============================================================
// 禁断诗篇 (bd_forbidden_verse) - passive, triggered after other songs
// No direct user interaction; test covers the card picker flow when selecting cards
// ============================================================

export function forbiddenVerseScenario(): ProtocolHarnessScenario {
  const characters = [bardCharacter, enemyCharacter];

  const hand = [
    card({ id: 'card-1', name: '火焰斩', type: 'Attack', element: 'Fire' }),
    card({ id: 'card-2', name: '冰霜箭', type: 'Magic', element: 'Water' }),
  ];

  const bard = playerView({
    id: BARD_PLAYER_ID,
    name: 'E2E Bard',
    camp: 'Red',
    role: 'bard',
    hand,
    hand_count: hand.length,
    heal: 3,
    max_heal: 5,
    is_active: true,
    tokens: { bd_inspiration: 0 },
    exclusive_cards: [
      card({
        id: 'starter-bard-bd_eternal_movement',
        name: '永恒乐章',
        type: 'Magic',
        element: 'Dark',
        description: '永恒乐章专属牌',
      }),
    ],
  });

  const players = [bard, enemyPlayerView()];

  return {
    roomCode: 'MOCK',
    myPlayerId: BARD_PLAYER_ID,
    myPlayerName: 'E2E Bard',
    characters,
    players: [bardPlayerInfo(), enemyPlayerInfo()],
    initialState: syncState({
      turn_player_id: BARD_PLAYER_ID,
      turn_stage: 'ActionExecution',
      available_skills: [],
      characters,
      players,
    }),
  };
}
