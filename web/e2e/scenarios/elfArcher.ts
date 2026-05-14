// ============================================================
// ElfArcher (精灵射手) Protocol Harness Scenarios
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

export const ELF_ARCHER_PLAYER_ID = 'elf_archer_player';
export const ENEMY_PLAYER_ID = 'enemy_1';
export const ALLY_PLAYER_ID = 'ally_1';

export const ELF_ARCHER_ELEMENTAL_SHOT_ID = 'elf_archer_elemental_shot';
export const ELF_ARCHER_ANIMAL_COMPANION_ID = 'elf_archer_animal_companion';
export const ELF_ARCHER_ELF_RITUAL_ID = 'elf_archer_elf_ritual';
export const ELF_ARCHER_PET_ENHANCE_ID = 'elf_archer_pet_enhance';

const elfArcherCharacter = characterView({
  id: 'elf_archer',
  name: '精灵射手',
  title: '咏',
  faction: '咏',
  skills: [
    {
      id: ELF_ARCHER_ELEMENTAL_SHOT_ID,
      title: '元素射击',
      description: '（攻击行动开始前发动）弃1张法术牌或［祝福］，根据牌的系别执行：火系+1伤害；水系命中后目标弃牌；地系无法应战；风系目标无距离限制；雷系强制命中。',
      type: 3, // 响应
      min_targets: 0, max_targets: 0, target_type: 0,
    },
    {
      id: ELF_ARCHER_ANIMAL_COMPANION_ID,
      title: '动物伙伴',
      description: '（受到伤害后发动）摸1张牌，弃1张牌。',
      type: 3, // 响应
      min_targets: 0, max_targets: 0, target_type: 0,
    },
    {
      id: ELF_ARCHER_ELF_RITUAL_ID,
      title: '精灵密仪',
      description: '［宝石］横置，将手中的［祝福］牌盖放于自己面前作为祝福储备。回合结束时，你可以消耗1［祝福］来令一名队友+1［治疗］。',
      type: 1, // 启动(大招)
      min_targets: 0, max_targets: 0, target_type: 0,
    },
    {
      id: ELF_ARCHER_PET_ENHANCE_ID,
      title: '宠物强化',
      description: '［水晶］升级动物伙伴的效果：摸牌数+1，或弃牌数-1，或目标弃牌。',
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

const defaultCharacters = [elfArcherCharacter, enemyCharacter, allyCharacter];

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function elfArcherHand(): Card[] {
  return [
    card({ id: 'elf-attack-1', name: '精灵箭', type: 'Attack', element: 'Light' }),
    card({ id: 'elf-attack-2', name: '精准', type: 'Attack', element: 'Light' }),
    card({ id: 'elf-fire-magic', name: '火球', type: 'Magic', element: 'Fire' }),
    card({ id: 'elf-water-magic', name: '冰冻', type: 'Magic', element: 'Water' }),
    card({ id: 'elf-earth-magic', name: '地刺', type: 'Magic', element: 'Earth' }),
    card({ id: 'elf-wind-magic', name: '风刃', type: 'Magic', element: 'Wind' }),
    card({ id: 'elf-thunder-magic', name: '雷击', type: 'Magic', element: 'Thunder' }),
    card({ id: 'elf-blessing-1', name: '祝福', type: 'Magic', element: 'Light' }),
  ];
}

function elfArcherAvailableSkill(skill: Partial<AvailableSkill> & { id: string; title: string }): AvailableSkill {
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

export function elfArcherScenario(options: {
  hand?: Card[];
  availableSkills?: AvailableSkill[];
  crystal?: number;
  gem?: number;
  turnStage?: string;
  buffs?: { id: string; name: string; duration: number; value: number; source_id: string }[];
  blessings?: Card[]; // 精灵密仪盖放的祝福储备
} = {}): ProtocolHarnessScenario {
  const hand = options.hand ?? elfArcherHand();
  const players = [
    playerView({
      id: ELF_ARCHER_PLAYER_ID,
      name: 'E2E ElfArcher',
      camp: 'Red',
      role: 'elf_archer',
      hand,
      hand_count: hand.length,
      crystal: options.crystal ?? 0,
      gem: options.gem ?? 0,
      is_active: true,
      buffs: options.buffs ?? [],
      blessings: options.blessings ?? [], // 祝福储备区（盖放）
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
    myPlayerId: ELF_ARCHER_PLAYER_ID,
    myPlayerName: 'E2E ElfArcher',
    characters: defaultCharacters,
    players: [
      playerInfo({ id: ELF_ARCHER_PLAYER_ID, name: 'E2E ElfArcher', camp: 'Red', char_role: 'elf_archer', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy E1', camp: 'Blue', char_role: 'enemy_char' }),
      playerInfo({ id: ALLY_PLAYER_ID, name: 'Ally A1', camp: 'Red', char_role: 'ally_char' }),
    ],
    initialState: syncState({
      turn_player_id: ELF_ARCHER_PLAYER_ID,
      turn_stage: options.turnStage ?? 'ActionExecution',
      available_skills: options.availableSkills ?? [],
      characters: defaultCharacters,
      players,
    }),
  };
}

// ============================================================
// Elemental Shot (元素射击) - 响应技能
// 后端通过 response_skills 自动触发
// 弃牌使用 choice_type: elf_archer_elemental_shot_pick
// ============================================================

export function elementalShotScenario(): ProtocolHarnessScenario {
  return elfArcherScenario({
    // 后端会设置 response_skills 触发确认弹框
  });
}

export function elementalShotDiscardPrompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_cards',
    player_id: ELF_ARCHER_PLAYER_ID,
    message: '【元素射击】请选择弃1张法术牌或祝福：',
    choice_type: 'elf_archer_elemental_shot_pick',
    options: [
      { id: 'elf-fire-magic', label: '火球（火系，+1伤害）' },
      { id: 'elf-water-magic', label: '冰冻（水系，命中后弃牌）' },
      { id: 'elf-earth-magic', label: '地刺（地系，无法应战）' },
      { id: 'elf-wind-magic', label: '风刃（风系，无距离限制）' },
      { id: 'elf-thunder-magic', label: '雷击（雷系，强制命中）' },
      { id: 'elf-blessing-1', label: '祝福' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}

// ============================================================
// Animal Companion (动物伙伴) - 响应技能
// 后端通过 response_skills 自动触发，使用 choice_type: elf_animal_companion_confirm
// 弃牌通过 system_discard_cards 处理
// ============================================================

export function animalCompanionScenario(): ProtocolHarnessScenario {
  return elfArcherScenario({
    // 后端会设置 response_skills 触发确认弹框
  });
}

export function animalCompanionPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: ELF_ARCHER_PLAYER_ID,
    message: '【动物伙伴】受到伤害后，摸1张牌弃1张牌？',
    choice_type: 'elf_animal_companion_confirm',
    options: [
      { id: 'confirm', label: '发动' },
      { id: 'skip', label: '跳过' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}

// ============================================================
// Elf Ritual (精灵密仪) - 启动技能(大招)
// 后端通过 startup_skills 触发，目标通过 min_targets 处理
// ============================================================

export function elfRitualScenario(): ProtocolHarnessScenario {
  return elfArcherScenario({
    gem: 1,
    availableSkills: [
      elfArcherAvailableSkill({
        id: ELF_ARCHER_ELF_RITUAL_ID, title: '精灵密仪',
      }),
    ],
  });
}

// 精灵密仪已激活，祝福已盖放的回合结束场景
export function elfRitualWithBlessingScenario(): ProtocolHarnessScenario {
  return elfArcherScenario({
    gem: 1,
    turnStage: 'TurnEnd',
    // 祝福储备区已有盖放的祝福牌（回合结束时消耗）
    blessings: [
      card({ id: 'elf-blessing-reserve-1', name: '祝福储备', type: 'Magic', element: 'Light' }),
    ],
    // 后端会设置 response_skills 触发回合结束选择
  });
}

// ============================================================
// Pet Enhance (宠物强化) - 响应技能(大招)
// 后端通过 response_skills 自动触发
// 分支选择使用 choice_type: elf_pet_empower_confirm
// ============================================================

export function petEnhanceScenario(): ProtocolHarnessScenario {
  return elfArcherScenario({
    crystal: 1,
    // 后端会设置 response_skills 触发确认弹框
  });
}

export function petEnhanceBranchPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: ELF_ARCHER_PLAYER_ID,
    message: '【宠物强化】选择升级效果：',
    choice_type: 'elf_pet_empower_confirm',
    options: [
      { id: 'draw_plus', label: '摸牌数+1' },
      { id: 'discard_minus', label: '弃牌数-1' },
      { id: 'target_discard', label: '目标弃牌' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}