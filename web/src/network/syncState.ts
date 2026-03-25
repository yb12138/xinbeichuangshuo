import type { GameStateUpdate, PlayerView } from '../types/game'
import type { SyncStatePayload } from './protocol'

export function buildGameStateUpdateFromSyncState(payload: SyncStatePayload): GameStateUpdate {
  const players: Record<string, PlayerView> = {}

  for (const player of payload.players || []) {
    if (!player?.id) continue
    players[player.id] = player
  }

  return {
    turn_stage: payload.turn_stage,
    combat_stage: payload.combat_stage,
    subflow: payload.subflow,
    current_player: payload.turn_player_id,
    has_performed_startup: payload.has_performed_startup,
    players,
    red_morale: payload.morale_red,
    blue_morale: payload.morale_blue,
    red_cups: payload.cups_red,
    blue_cups: payload.cups_blue,
    red_gems: payload.stones_red?.[0] ?? 0,
    blue_gems: payload.stones_blue?.[0] ?? 0,
    red_crystals: payload.stones_red?.[1] ?? 0,
    blue_crystals: payload.stones_blue?.[1] ?? 0,
    deck_count: payload.deck_count,
    discard_count: payload.discard_count,
    available_skills: payload.available_skills ?? [],
    characters: payload.characters,
  }
}
