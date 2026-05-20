import type {
  AvailableSkill,
  Card,
  CharacterView,
  PlayerInfo,
  PlayerView,
  Prompt,
} from '../../src/types/game';
import type {
  RequireActionPayload,
  RoomEventPayload,
  SyncStatePayload,
  TimelineNotifyPayload,
  WsMessage,
} from '../../src/network/protocol';

export interface ProtocolHarnessScenario {
  roomCode: string;
  myPlayerId: string;
  myPlayerName: string;
  players: PlayerInfo[];
  characters: CharacterView[];
  initialState: SyncStatePayload;
}

export const handCardIDInteraction = {
  selection_source: 'hand',
  selection_value: 'card_id',
  confirm_mode: 'manual',
  submit_action: 'select',
} as const;

export const fieldOptionIndexInteraction = {
  selection_source: 'field',
  selection_value: 'option_index',
  confirm_mode: 'manual',
  submit_action: 'select',
} as const;

export const singleTargetOptionIndexInteraction = {
  selection_source: 'target',
  selection_value: 'option_index',
  confirm_mode: 'immediate',
  submit_action: 'select',
} as const;

export const multiTargetOptionIndexInteraction = {
  selection_source: 'target',
  selection_value: 'option_index',
  confirm_mode: 'manual',
  submit_action: 'select',
} as const;

export function card(overrides: Partial<Card> = {}): Card {
  return {
    id: 'card-1',
    name: '火焰斩',
    type: 'Attack',
    element: 'Fire',
    damage: 2,
    description: '测试用卡牌',
    ...overrides,
  };
}

export function availableSkill(overrides: Partial<AvailableSkill> = {}): AvailableSkill {
  return {
    id: 'skill-1',
    title: '测试技能',
    description: '测试用技能',
    min_targets: 0,
    max_targets: 0,
    target_type: 0,
    cost_gem: 0,
    cost_crystal: 0,
    cost_discards: 0,
    ...overrides,
  };
}

export function characterView(overrides: Partial<CharacterView> = {}): CharacterView {
  return {
    id: 'hero',
    name: '勇者',
    title: '基础角色',
    faction: '星杯',
    skills: [],
    ...overrides,
  };
}

export function playerInfo(overrides: Partial<PlayerInfo> = {}): PlayerInfo {
  return {
    id: 'player-1',
    name: '玩家',
    camp: 'Red',
    char_role: 'hero',
    ready: true,
    is_online: true,
    is_bot: false,
    is_host: false,
    ...overrides,
  };
}

export function playerView(overrides: Partial<PlayerView> = {}): PlayerView {
  const hand = overrides.hand ?? [];
  return {
    id: 'player-1',
    name: '玩家',
    camp: 'Red',
    role: 'hero',
    max_hand: 6,
    exclusive_card_count: 0,
    hand,
    exclusive_cards: [],
    field: [],
    buffs: [],
    heal: 0,
    max_heal: 0,
    gem: 0,
    crystal: 0,
    is_active: false,
    tokens: {},
    indicators: {},
    ...overrides,
    hand_count: overrides.hand_count ?? hand.length,
  };
}

export function syncState(overrides: Partial<SyncStatePayload> = {}): SyncStatePayload {
  return {
    room_state: 'Playing',
    turn_stage: 'ActionExecution',
    combat_stage: '',
    subflow: '',
    turn_player_id: '',
    has_performed_startup: false,
    morale_red: 15,
    morale_blue: 15,
    cups_red: 0,
    cups_blue: 0,
    stones_red: [0, 0],
    stones_blue: [0, 0],
    deck_count: 142,
    discard_count: 0,
    available_skills: [],
    characters: [],
    players: [],
    ...overrides,
  };
}

export function roomEvent(overrides: Partial<RoomEventPayload>): WsMessage<RoomEventPayload> {
  return {
    Cmd: 'RoomEvent',
    Data: {
      action: 'player_list',
      room_code: 'MOCK',
      ...overrides,
    } as RoomEventPayload,
  };
}

export function syncStateMessage(payload: SyncStatePayload): WsMessage<SyncStatePayload> {
  return {
    Cmd: 'SyncState',
    Data: payload,
  };
}

export function requireActionPayload(
  prompt: Prompt,
  overrides: Partial<RequireActionPayload> = {}
): RequireActionPayload {
  return {
    interrupt_type: 'prompt',
    target_user_id: prompt.player_id,
    timeout: 30,
    msg: prompt.message,
    prompt,
    ...overrides,
  };
}

export function requireActionMessage(
  prompt: Prompt,
  overrides: Partial<RequireActionPayload> = {}
): WsMessage<RequireActionPayload> {
  return {
    Cmd: 'RequireAction',
    Data: requireActionPayload(prompt, overrides),
  };
}

export function notifyTimelineMessage(
  overrides: Partial<TimelineNotifyPayload> = {}
): WsMessage<TimelineNotifyPayload> {
  return {
    Cmd: 'NotifyTimeline',
    Data: {
      room_id: 'MOCK',
      seq_start: 1,
      seq_end: 1,
      is_replay: false,
      events: [],
      ...overrides,
    },
  };
}
