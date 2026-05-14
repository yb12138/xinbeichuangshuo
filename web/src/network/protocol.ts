import type { AvailableSkill, Card, CharacterView, PlayerView, Prompt, RoomEvent } from '../types/game'

export interface WsMessage<T = unknown> {
  Cmd: string
  Data: T
}

export interface TargetNode {
  target_user_id?: string
  selected_field_cards?: string[]
  selected_hand_cards?: string[]
  selected_tokens?: string[]
}

export interface ClientActionRequest {
  action_type: string
  used_card_uuids?: string[]
  targets?: TargetNode[]
  skill_id?: string
  option_indexes?: number[]
  extra_args?: string[]
}

export interface RoomActionRequest {
  action: string
  camp?: string
  char_role?: string
  target_id?: string
  bot_name?: string
}

export interface SyncStatePayload {
  room_state: string
  turn_stage?: string
  combat_stage?: string
  subflow?: string
  turn_player_id: string
  has_performed_startup: boolean
  morale_red: number
  morale_blue: number
  cups_red: number
  cups_blue: number
  stones_red: [number, number] | number[]
  stones_blue: [number, number] | number[]
  deck_count: number
  discard_count: number
  available_skills?: AvailableSkill[]
  characters?: CharacterView[]
  players: PlayerView[]
}

export interface RequireActionPayload {
  interrupt_type: string
  target_user_id: string
  timeout: number
  msg: string
  valid_actions?: string[]
  require_count?: number
  prompt_type?: string
  prompt?: Prompt
}

export interface TimelineDelta {
  type: string
  target_user_id?: string
  value?: number
}

export interface TimelineEvent {
  event_id: number
  turn_id: number
  turn_stage?: string
  combat_stage?: string
  subflow?: string
  timing?: string
  chain_id: string
  parent_event_id?: number
  type: string
  outcome: string
  visibility: string
  actor_user_id?: string
  actor_name?: string
  target_user_ids?: string[]
  target_name?: string
  action_type?: string
  skill_id?: string
  card_ids?: string[]
  cards?: Card[]
  hidden?: boolean
  damage?: number
  damage_type?: string
  detail_kind?: string
  cue_phase?: string
  draw_count?: number
  reason?: string
  deltas?: TimelineDelta[]
  message?: string
  gameplay_type?: string
}

export interface TimelineNotifyPayload {
  room_id: string
  seq_start: number
  seq_end: number
  is_replay: boolean
  events: TimelineEvent[]
}

export type RoomEventPayload = RoomEvent
