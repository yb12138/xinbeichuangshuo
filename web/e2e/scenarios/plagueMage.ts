import type { AvailableSkill, Prompt } from '../../src/types/game';
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

export const PLAGUE_PLAYER_ID = 'plague_player';
export const ENEMY_PLAYER_ID = 'enemy_1';
export const PLAGUE_DEATH_TOUCH_SKILL_ID = 'plague_death_touch';

export const plagueDeathTouchSkill = availableSkill({
  id: PLAGUE_DEATH_TOUCH_SKILL_ID,
  title: '死亡之触',
  description: '失去治疗并弃置同系牌，对目标造成法术伤害。',
  min_targets: 0,
  max_targets: 0,
  target_type: 0,
  cost_discards: 0,
});

export function plagueMageDeathTouchScenario(options: {
  availableSkills?: AvailableSkill[];
} = {}): ProtocolHarnessScenario {
  const characters = [
    characterView({
      id: 'plague_mage',
      name: '瘟疫法师',
      title: '死亡触媒',
      faction: '瘟疫',
      skills: [
        {
          id: PLAGUE_DEATH_TOUCH_SKILL_ID,
          title: '死亡之触',
          description: '失去治疗并弃置同系牌，对目标造成法术伤害。',
          type: 2,
          min_targets: 0,
          max_targets: 0,
          target_type: 0,
          cost_gem: 0,
          cost_crystal: 0,
          cost_discards: 0,
        },
      ],
    }),
    characterView({
      id: 'hero',
      name: '勇者',
      title: '基础角色',
      faction: '星杯',
      skills: [],
    }),
  ];

  const plagueHand = [
    card({
      id: 'fire-attack-1',
      name: '火焰斩',
      type: 'Attack',
      element: 'Fire',
      description: '测试用火系攻击牌 1',
    }),
    card({
      id: 'fire-attack-2',
      name: '火焰斩',
      type: 'Attack',
      element: 'Fire',
      description: '测试用火系攻击牌 2',
    }),
    card({
      id: 'water-attack-1',
      name: '水涟斩',
      type: 'Attack',
      element: 'Water',
      description: '测试用水系攻击牌',
    }),
    card({
      id: 'magic-light-1',
      name: '圣光',
      type: 'Magic',
      element: 'Light',
      description: '测试用圣光',
    }),
  ];

  const players = [
    playerView({
      id: PLAGUE_PLAYER_ID,
      name: 'E2E Host',
      camp: 'Red',
      role: 'plague_mage',
      hand: plagueHand,
      hand_count: plagueHand.length,
      heal: 3,
      max_heal: 5,
      is_active: true,
    }),
    playerView({
      id: ENEMY_PLAYER_ID,
      name: 'Enemy Bot',
      camp: 'Blue',
      role: 'hero',
      hand: [],
      hand_count: 4,
      heal: 0,
      max_heal: 2,
      is_active: false,
    }),
  ];

  return {
    roomCode: 'MOCK',
    myPlayerId: PLAGUE_PLAYER_ID,
    myPlayerName: 'E2E Host',
    characters,
    players: [
      playerInfo({
        id: PLAGUE_PLAYER_ID,
        name: 'E2E Host',
        camp: 'Red',
        char_role: 'plague_mage',
        is_host: true,
      }),
      playerInfo({
        id: ENEMY_PLAYER_ID,
        name: 'Enemy Bot',
        camp: 'Blue',
        char_role: 'hero',
        is_bot: true,
      }),
    ],
    initialState: syncState({
      turn_player_id: PLAGUE_PLAYER_ID,
      turn_stage: 'ActionExecution',
      available_skills: options.availableSkills ?? [plagueDeathTouchSkill],
      characters,
      players,
    }),
  };
}

export function deathTouchElementPrompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_cards',
    player_id: PLAGUE_PLAYER_ID,
    message: '请选择【死亡之触】消耗手牌的系别',
    choice_type: 'plague_death_touch_element',
    skill_id: PLAGUE_DEATH_TOUCH_SKILL_ID,
    options: [
      { id: 'Fire', label: '火 Fire', button_label: '火', element: 'Fire' },
      { id: 'Water', label: '水 Water', button_label: '水', element: 'Water' },
      { id: 'Thunder', label: '雷 Thunder', button_label: '雷', element: 'Thunder' },
    ],
    presentation: { kind: 'card_picker', layout: 'inline', card_source: 'hand', card_filter: 'plague_death_touch_element', cancel_policy: 'abort', numeric_base: 0 },
    min: 1,
    max: 1,
  } satisfies Prompt);
}

export function deathTouchXPrompt(): WsMessage {
  return requireActionMessage({
    type: 'confirm',
    player_id: PLAGUE_PLAYER_ID,
    message: '请选择【死亡之触】失去的治疗点数 X',
    choice_type: 'plague_death_touch_x',
    skill_id: PLAGUE_DEATH_TOUCH_SKILL_ID,
    options: [
      { id: 'x1', label: 'X=1', button_label: '1' },
      { id: 'x2', label: 'X=2', button_label: '2' },
      { id: 'x3', label: 'X=3', button_label: '3' },
    ],
    presentation: { kind: 'numeric', numeric_base: 0 },
    min: 1,
    max: 1,
  } satisfies Prompt);
}

export function deathTouchCardsPrompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_cards',
    player_id: PLAGUE_PLAYER_ID,
    message: '请选择要弃置的同系手牌',
    choice_type: 'plague_death_touch_cards',
    skill_id: PLAGUE_DEATH_TOUCH_SKILL_ID,
    options: [
      { id: '0', label: '1: 火焰斩', button_label: '选择', card_id: 'fire-attack-1' },
      { id: '1', label: '2: 火焰斩', button_label: '选择', card_id: 'fire-attack-2' },
    ],
    presentation: { kind: 'card_picker', card_source: 'hand', card_filter: 'same_element', cancel_policy: 'abort', numeric_base: 0 },
    min: 2,
    max: 2,
  } satisfies Prompt);
}

export function deathTouchTargetPrompt(): WsMessage {
  return requireActionMessage({
    type: 'choose_target',
    player_id: PLAGUE_PLAYER_ID,
    message: '请选择【死亡之触】目标',
    choice_type: 'plague_death_touch_target',
    skill_id: PLAGUE_DEATH_TOUCH_SKILL_ID,
    options: [
      { id: ENEMY_PLAYER_ID, label: 'Enemy Bot hero', button_label: '选择' },
    ],
    presentation: { kind: 'target_picker', numeric_base: 0 },
    min: 1,
    max: 1,
  } satisfies Prompt);
}
