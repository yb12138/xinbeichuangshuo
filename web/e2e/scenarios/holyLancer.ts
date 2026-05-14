// ============================================================
// HolyLancer (圣枪骑士) Protocol Harness Scenarios
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

export const HOLY_LANCER_PLAYER_ID = 'holy_lancer_player';
export const ENEMY_PLAYER_ID = 'enemy_1';
export const ALLY_PLAYER_ID = 'ally_1';

export const HOLY_LANCER_GLORY_ID = 'holy_lancer_glory';
export const HOLY_LANCER_PUNISHMENT_ID = 'holy_lancer_punishment';
export const HOLY_LANCER_SKY_SPEAR_ID = 'holy_lancer_sky_spear';
export const HOLY_LANCER_EARTH_SPEAR_ID = 'holy_lancer_earth_spear';
export const HOLY_LANCER_HOLY_LIGHT_HEAL_ID = 'holy_lancer_holy_light_heal';

const holyLancerCharacter = characterView({
  id: 'holy_lancer',
  name: '圣枪骑士',
  title: '圣',
  faction: '圣',
  skills: [
    {
      id: 'holy_lancer_holy_revelation',
      title: '神圣启示',
      description: '（我方［星杯区］的［星杯］数不小于对方时）你的［治疗］上限+1。',
      type: 0, // 被动
      min_targets: 0, max_targets: 0, target_type: 0,
    },
    {
      id: HOLY_LANCER_GLORY_ID,
      title: '辉耀',
      description: '（弃1张水系牌［展示］）所有人各+1［治疗］，额外+1［攻击行动］。',
      type: 2, // 法术
      min_targets: 0, max_targets: 0, target_type: 0,
      cost_discards: 1, discard_element: 'Water',
    },
    {
      id: HOLY_LANCER_PUNISHMENT_ID,
      title: '惩戒',
      description: '（弃1张法术牌［展示］）将其他角色的1点［治疗］转移给你，额外+1［攻击行动］。',
      type: 2, // 法术
      min_targets: 1, max_targets: 1, target_type: 3,
      cost_discards: 1, discard_type: 'Magic',
    },
    {
      id: 'holy_lancer_holy_strike',
      title: '圣击',
      description: '（攻击命中后发动②）你+1［治疗］。',
      type: 3, // 响应
      min_targets: 0, max_targets: 0, target_type: 0,
    },
    {
      id: HOLY_LANCER_SKY_SPEAR_ID,
      title: '天枪',
      description: '（主动攻击前发动①）移除你的2点［治疗］，本次攻击对手无法应战；不能和［圣击］同时发动。',
      type: 3, // 响应
      min_targets: 0, max_targets: 0, target_type: 0,
    },
    {
      id: HOLY_LANCER_EARTH_SPEAR_ID,
      title: '地枪',
      description: '（主动攻击命中后发动②）移除你的X点［治疗］，本次攻击伤害额外+X，X最高为4；不能和［圣击］同时发动。',
      type: 3, // 响应
      min_targets: 0, max_targets: 0, target_type: 0,
    },
    {
      id: HOLY_LANCER_HOLY_LIGHT_HEAL_ID,
      title: '圣光祈愈',
      description: '［宝石］无视你的［治疗］上限为你+2［治疗］，但你的［治疗］数最高为5，额外+1［攻击行动］；本回合你不能再发动［天枪］。',
      type: 2, // 法术(大招)
      min_targets: 0, max_targets: 0, target_type: 0,
    },
  ],
});

const enemyCharacter = characterView({
  id: 'enemy_char', name: '守卫', title: '测试目标', faction: '异端', skills: [],
});

const allyCharacter = characterView({
  id: 'ally_char', name: '勇者', title: '测试队友', faction: '圣', skills: [],
});

const defaultCharacters = [holyLancerCharacter, enemyCharacter, allyCharacter];

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function holyLancerHand(): Card[] {
  return [
    card({ id: 'hl-water-atk1', name: '水刃', type: 'Attack', element: 'Water' }),
    card({ id: 'hl-water-atk2', name: '寒冰斩', type: 'Attack', element: 'Water' }),
    card({ id: 'hl-water-magic', name: '冰冻', type: 'Magic', element: 'Water' }),
    card({ id: 'hl-fire-magic', name: '火球', type: 'Magic', element: 'Fire' }),
    card({ id: 'hl-thunder', name: '雷击', type: 'Magic', element: 'Thunder' }),
    card({ id: 'hl-fire-atk', name: '火焰斩', type: 'Attack', element: 'Fire' }),
  ];
}

function holyLancerAvailableSkill(skill: Partial<AvailableSkill> & { id: string; title: string }): AvailableSkill {
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

export function holyLancerScenario(options: {
  hand?: Card[];
  availableSkills?: AvailableSkill[];
  crystal?: number;
  gem?: number;
  turnStage?: string;
  heal?: number;
} = {}): ProtocolHarnessScenario {
  const hand = options.hand ?? holyLancerHand();
  const players = [
    playerView({
      id: HOLY_LANCER_PLAYER_ID,
      name: 'E2E HolyLancer',
      camp: 'Red',
      role: 'holy_lancer',
      hand,
      hand_count: hand.length,
      crystal: options.crystal ?? 0,
      gem: options.gem ?? 0,
      heal: options.heal ?? 2,
      max_heal: 5,
      is_active: true,
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
    myPlayerId: HOLY_LANCER_PLAYER_ID,
    myPlayerName: 'E2E HolyLancer',
    characters: defaultCharacters,
    players: [
      playerInfo({ id: HOLY_LANCER_PLAYER_ID, name: 'E2E HolyLancer', camp: 'Red', char_role: 'holy_lancer', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy E1', camp: 'Blue', char_role: 'enemy_char' }),
      playerInfo({ id: ALLY_PLAYER_ID, name: 'Ally A1', camp: 'Red', char_role: 'ally_char' }),
    ],
    initialState: syncState({
      turn_player_id: HOLY_LANCER_PLAYER_ID,
      turn_stage: options.turnStage ?? 'ActionExecution',
      available_skills: options.availableSkills ?? [],
      characters: defaultCharacters,
      players,
    }),
  };
}

// ============================================================
// Glory (辉耀) - 法术技能
// ============================================================

export function gloryScenario(): ProtocolHarnessScenario {
  return holyLancerScenario({
    availableSkills: [
      holyLancerAvailableSkill({
        id: HOLY_LANCER_GLORY_ID, title: '辉耀',
        cost_discards: 1, discard_element: 'Water',
      }),
    ],
  });
}

// ============================================================
// Punishment (惩戒) - 法术技能
// ============================================================

export function punishmentScenario(): ProtocolHarnessScenario {
  return holyLancerScenario({
    availableSkills: [
      holyLancerAvailableSkill({
        id: HOLY_LANCER_PUNISHMENT_ID, title: '惩戒',
        cost_discards: 1, discard_type: 'Magic',
        min_targets: 1, max_targets: 1, target_type: 3,
      }),
    ],
  });
}

export function punishmentTargetPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: HOLY_LANCER_PLAYER_ID,
    message: '【惩戒】请选择一名目标角色，将其1点治疗转移给你：',
    choice_type: 'holy_lancer_punishment_target',
    options: [
      { id: ENEMY_PLAYER_ID, label: 'Enemy E1' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}

// ============================================================
// Sky Spear (天枪) - 响应技能
// ============================================================

export function skySpearScenario(options: {
  heal?: number;
} = {}): ProtocolHarnessScenario {
  return holyLancerScenario({
    heal: options.heal ?? 3, // 至少2点治疗
  });
}

export function skySpearBeforeAttackPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: HOLY_LANCER_PLAYER_ID,
    message: '【天枪】是否移除2点［治疗］，使本次攻击对手无法应战？',
    choice_type: 'holy_lancer_sky_spear',
    options: [
      { id: 'confirm', label: '发动' },
      { id: 'skip', label: '跳过' },
    ],
    min: 1, max: 1,
  } satisfies Prompt);
}

// ============================================================
// Earth Spear (地枪) - 响应技能
// ============================================================

export function earthSpearScenario(options: {
  heal?: number;
} = {}): ProtocolHarnessScenario {
  return holyLancerScenario({
    heal: options.heal ?? 4, // 至少1点治疗
  });
}

export function earthSpearAfterHitPrompt(maxX: number): WsMessage {
  const options: Array<{ id: string; label: string }> = [];
  // X can be 0 to min(heal, 4)
  for (let x = 0; x <= Math.min(maxX, 4); x++) {
    if (x === 0) {
      options.push({ id: '0', label: '不发动' });
    } else {
      options.push({ id: String(x), label: `移除${x}点治疗，伤害+${x}` });
    }
  }
  return requireActionMessage({
    type: 'confirm',
    player_id: HOLY_LANCER_PLAYER_ID,
    message: '【地枪】命中后，请选择移除X点［治疗］：',
    choice_type: 'holy_lancer_earth_spear',
    options,
    min: 1, max: 1,
  } satisfies Prompt);
}

// ============================================================
// Holy Light Heal (圣光祈愈) - 法术技能(大招)
// ============================================================

export function holyLightHealScenario(): ProtocolHarnessScenario {
  return holyLancerScenario({
    gem: 1,
    availableSkills: [
      holyLancerAvailableSkill({
        id: HOLY_LANCER_HOLY_LIGHT_HEAL_ID, title: '圣光祈愈',
      }),
    ],
  });
}