import type { GameEvent } from '../types/game'
import type { TimelineEvent } from './protocol'

const SUPPORTED_GAMEPLAY_TIMELINE_TYPES = new Set<GameEvent['event_type']>([
  'log',
  'error',
  'game_end',
  'chat',
  'card_revealed',
  'damage_dealt',
  'action_step',
  'combat_cue',
  'draw_cards',
])

export function extractGameplayEventsFromTimeline(events: TimelineEvent[]): GameEvent[] {
  const output: GameEvent[] = []

  for (const event of events) {
    const gameplayType = event.gameplay_type as GameEvent['event_type'] | undefined
    if (!gameplayType || !SUPPORTED_GAMEPLAY_TIMELINE_TYPES.has(gameplayType)) {
      continue
    }

    const payload = buildGameEventPayload(event, gameplayType)

    output.push({
      event_type: gameplayType,
      ...(payload as Omit<GameEvent, 'event_type'>),
    } as GameEvent)
  }

  return output
}

function buildGameEventPayload(
  event: TimelineEvent,
  gameplayType: GameEvent['event_type']
): Record<string, unknown> {
  switch (gameplayType) {
    case 'log':
    case 'error':
    case 'game_end':
      return {
        message: event.message,
      }
    case 'chat':
      return {
        player_id: event.actor_user_id,
        player_name: event.actor_name,
        message: event.message,
      }
    case 'card_revealed':
      return {
        player_id: event.actor_user_id,
        player_name: event.actor_name,
        cards: event.cards || [],
        action_type: event.action_type,
        hidden: event.hidden,
      }
    case 'damage_dealt':
      return {
        source_id: event.actor_user_id,
        source_name: event.actor_name,
        target_id: event.target_user_ids?.[0],
        target_name: event.target_name,
        damage: event.damage ?? extractDeltaValue(event, 'TimelineDeltaDamage'),
        damage_type: event.damage_type,
        message: event.message,
      }
    case 'action_step':
      return {
        line: event.message,
        kind: event.detail_kind,
      }
    case 'combat_cue':
      return {
        attacker_id: event.actor_user_id,
        target_id: event.target_user_ids?.[0],
        phase: event.cue_phase,
      }
    case 'draw_cards':
      return {
        player_id: event.actor_user_id,
        player_name: event.actor_name,
        draw_count: event.draw_count ?? extractDeltaValue(event, 'TimelineDeltaHandCount'),
        reason: event.reason,
      }
    default:
      return event.message ? { message: event.message } : {}
  }
}

function extractDeltaValue(event: TimelineEvent, deltaType: string): number | undefined {
  return event.deltas?.find(delta => delta.type === deltaType)?.value
}
