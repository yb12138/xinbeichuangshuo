// ============================================================
// Berserker (狂战士) Protocol Harness Scenarios
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

export const BERSERKER_PLAYER_ID = 'berserker_player';
export const ENEMY_PLAYER_ID = 'enemy_1';
export const ALLY_PLAYER_ID = 'ally_1';

export const BERSERKER_TEAR_ID = 'berserker_tear';
export const BERSERKER_BLOODY_ROAR_ID = 'berserker_bloody_roar';
export const BERSERKER_BLOOD_SHADOW_ID = 'berserker_blood_shadow';

const berserkerCharacter = characterView({
  id: 'berserker',
  name: '狂战士',
  title: '血',
  faction: '血',
  skills: [
    {
      id: 'berserker_rage',
      title: '狂化',
      description: '你发动的所有攻击伤害额外+1。（攻击命中时②，若你的手牌>3）本次攻击伤害额外+1。',
      type: 0, // 被动
      min_targets: 0, max_targets: 0, target_type: 0,
    },
    {
      id: BERSERKER_TEAR_ID,
      title: '撕裂',
      description: '［宝石］攻击命中后发动②，本次攻击伤害额外+2。',
      type: 3, // 响应(大招)
      min_targets: 0, max_targets: 0, target_type: 0,
    },
    {
      id: BERSERKER_BLOODY_ROAR_ID,
      title: '血腥咆哮',
      description: '作为主动攻击打出时发动，若攻击的目标拥有的［治疗］为2，则本次攻击强制命中。',
      type: 3, // 响应(独有)
      min_targets: 0, max_targets: 0, target_type: 0,
    },
    {
      id: BERSERKER_BLOOD_SHADOW_ID,
      title: '血影狂刀',
      description: '作为主动攻击打出时发动●若命中后②对手的手牌为2，本次攻击伤害额外+2。●若命中后②对手的手牌为3，本次攻击伤害额外+1。',
      type: 3, // 响应(独有)
      min_targets: 0, max_targets: 0, target_type: 0,
    },
  ],
});

const enemyCharacter = characterView({
  id: 'enemy_char', name: '守卫', title: '测试目标', faction: '异端', skills: [],
});

const allyCharacter = characterView({
  id: 'ally_char', name: '勇者', title: '测试队友', faction: '血', skills: [],
});

const defaultCharacters = [berserkerCharacter, enemyCharacter, allyCharacter];

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function berserkerHand(): Card[] {
  return [
    card({ id: 'berserker-fire-atk1', name: '火焰斩', type: 'Attack', element: 'Fire' }),
    card({ id: 'berserker-fire-atk2', name: '烈焰斩', type: 'Attack', element: 'Fire' }),
    card({ id: 'berserker-water-atk', name: '水刃', type: 'Attack', element: 'Water' }),
    card({ id: 'berserker-magic', name: '火球', type: 'Magic', element: 'Fire' }),
    card({ id: 'berserker-thunder', name: '雷击', type: 'Magic', element: 'Thunder' }),
    card({ id: BERSERKER_BLOODY_ROAR_ID, name: '血腥咆哮', type: 'Attack', element: 'Fire' }),
    card({ id: BERSERKER_BLOOD_SHADOW_ID, name: '血影狂刀', type: 'Attack', element: 'Fire' }),
  ];
}

function berserkerAvailableSkill(skill: Partial<AvailableSkill> & { id: string; title: string }): AvailableSkill {
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

export function berserkerScenario(options: {
  hand?: Card[];
  availableSkills?: AvailableSkill[];
  crystal?: number;
  gem?: number;
  turnStage?: string;
  tokens?: Record<string, number>;
} = {}): ProtocolHarnessScenario {
  const hand = options.hand ?? berserkerHand();
  const players = [
    playerView({
      id: BERSERKER_PLAYER_ID,
      name: 'E2E Berserker',
      camp: 'Red',
      role: 'berserker',
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
      hand: [
        card({ id: 'en-card-1', name: '测试牌', type: 'Attack', element: 'Fire' }),
        card({ id: 'en-card-2', name: '测试牌2', type: 'Attack', element: 'Water' }),
      ],
      hand_count: 2,
      max_hand: 6,
      heal: 2, // 治疗为2，触发血腥咆哮
      max_heal: 4,
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
    myPlayerId: BERSERKER_PLAYER_ID,
    myPlayerName: 'E2E Berserker',
    characters: defaultCharacters,
    players: [
      playerInfo({ id: BERSERKER_PLAYER_ID, name: 'E2E Berserker', camp: 'Red', char_role: 'berserker', is_host: true }),
      playerInfo({ id: ENEMY_PLAYER_ID, name: 'Enemy E1', camp: 'Blue', char_role: 'enemy_char' }),
      playerInfo({ id: ALLY_PLAYER_ID, name: 'Ally A1', camp: 'Red', char_role: 'ally_char' }),
    ],
    initialState: syncState({
      turn_player_id: BERSERKER_PLAYER_ID,
      turn_stage: options.turnStage ?? 'ActionExecution',
      available_skills: options.availableSkills ?? [],
      characters: defaultCharacters,
      players,
    }),
  };
}

// ============================================================
// Tear (撕裂) - 响应技能(大招)
// ============================================================

export function tearScenario(options: {
  gem?: number;
} = {}): ProtocolHarnessScenario {
  return berserkerScenario({
    gem: options.gem ?? 1,
    availableSkills: [
      berserkerAvailableSkill({
        id: BERSERKER_TEAR_ID, title: '撕裂',
      }),
    ],
  });
}

export function tearHitPrompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_skill',
    player_id: BERSERKER_PLAYER_ID,
    message: '你触发了响应技能【撕裂】，请选择是否发动。',
    choice_type: 'response_skill_choice',
    options: [
      { id: 'berserker_tear', label: '发动撕裂', hint: '消耗1宝石，本次攻击伤害额外+2' },
      { id: 'skip', label: '跳过', hint: '不发动响应技能' },
    ],
    min: 1, max: 1,
    presentation: { kind: 'skill_choice', layout: 'overlay' },
  } satisfies Prompt);
}

// ============================================================
// Bloody Roar (血腥咆哮) - 独有牌响应
// 注意：此技能为自动触发，后端不会发送选择弹框
// ============================================================

export function bloodyRoarScenario(): ProtocolHarnessScenario {
  return berserkerScenario({
    hand: [
      card({ id: BERSERKER_BLOODY_ROAR_ID, name: '血腥咆哮', type: 'Attack', element: 'Fire' }),
      card({ id: 'berserker-fire-atk1', name: '火焰斩', type: 'Attack', element: 'Fire' }),
      card({ id: 'berserker-water-atk', name: '水刃', type: 'Attack', element: 'Water' }),
      card({ id: 'berserker-magic', name: '火球', type: 'Magic', element: 'Fire' }),
    ],
  });
}

// bloodyRoarTargetHeal2Prompt 已删除 - 此技能自动触发，无弹窗交互

// ============================================================
// Blood Shadow Blade (血影狂刀) - 独有牌响应
// 注意：此技能为自动触发，后端不会发送选择弹框
// ============================================================

export function bloodShadowScenario(options: {
  enemyHandCount?: number;
} = {}): ProtocolHarnessScenario {
  const enemyHandCount = options.enemyHandCount ?? 2;
  const enemyHand: Card[] = [];
  for (let i = 0; i < enemyHandCount; i++) {
    enemyHand.push(card({ id: `en-card-${i}`, name: `测试牌${i}`, type: 'Attack', element: 'Fire' }));
  }

  return berserkerScenario({
    hand: [
      card({ id: BERSERKER_BLOOD_SHADOW_ID, name: '血影狂刀', type: 'Attack', element: 'Fire' }),
      card({ id: 'berserker-fire-atk1', name: '火焰斩', type: 'Attack', element: 'Fire' }),
      card({ id: 'berserker-water-atk', name: '水刃', type: 'Attack', element: 'Water' }),
      card({ id: 'berserker-magic', name: '火球', type: 'Magic', element: 'Fire' }),
    ],
  });
}

// bloodShadowHand2Prompt/bloodShadowHand3Prompt 已删除 - 此技能自动触发，无弹窗交互