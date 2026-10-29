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
  'skill_activated',
  'special_action',
  'state_delta',
])

export function extractGameplayEventsFromTimeline(events: TimelineEvent[]): GameEvent[] {
  const output: GameEvent[] = []

  for (const event of events) {
    const gameplayType = event.gameplay_type as GameEvent['event_type'] | undefined
    if (!gameplayType || !SUPPORTED_GAMEPLAY_TIMELINE_TYPES.has(gameplayType)) {
      continue
    }

    const evt = buildGameEvent(event, gameplayType)
    if (evt) {
      output.push(evt)
    }
  }

  return output
}

function buildGameEvent(event: TimelineEvent, gameplayType: GameEvent['event_type']): GameEvent | null {
  switch (gameplayType) {
    case 'log':
      return { event_type: 'log', message: event.message ?? '' }
    case 'error':
      return { event_type: 'error', message: event.message ?? '' }
    case 'game_end':
      return { event_type: 'game_end', message: event.message ?? '' }
    case 'chat':
      return {
        event_type: 'chat',
        player_id: event.actor_user_id ?? '',
        player_name: event.actor_name ?? '',
        message: event.message ?? '',
      }
    case 'card_revealed':
      return {
        event_type: 'card_revealed',
        player_id: event.actor_user_id ?? '',
        player_name: event.actor_name ?? '',
        cards: event.cards || [],
        action_type: event.action_type ?? '',
        hidden: event.hidden ?? false,
      }
    case 'damage_dealt':
      return {
        event_type: 'damage_dealt',
        source_id: event.actor_user_id ?? '',
        source_name: event.actor_name ?? '',
        target_id: event.target_user_ids?.[0] ?? '',
        target_name: event.target_name ?? '',
        damage: event.damage ?? extractDeltaValue(event, 'TimelineDeltaDamage') ?? 0,
        damage_type: event.damage_type ?? '',
        message: event.message,
      }
    case 'action_step':
      return {
        event_type: 'action_step',
        line: event.message ?? '',
        kind: (event.detail_kind as 'detail' | 'summary') ?? 'detail',
      }
    case 'combat_cue':
      return {
        event_type: 'combat_cue',
        attacker_id: event.actor_user_id ?? '',
        target_id: event.target_user_ids?.[0] ?? '',
        phase: event.cue_phase ?? '',
      }
    case 'draw_cards':
      return {
        event_type: 'draw_cards',
        player_id: event.actor_user_id ?? '',
        player_name: event.actor_name ?? '',
        draw_count: event.draw_count ?? extractDeltaValue(event, 'TimelineDeltaHandCount') ?? 0,
        reason: event.reason ?? '',
      }
    case 'skill_activated':
      return {
        event_type: 'skill_activated',
        player_id: event.actor_user_id ?? '',
        player_name: event.actor_name ?? '',
        skill_id: event.skill_id ?? '',
        skill_name: event.skill_name ?? '',
        effect_text: event.effect_text ?? '',
        target_ids: event.target_user_ids || [],
      }
    case 'special_action':
      return {
        event_type: 'special_action',
        player_id: event.actor_user_id ?? '',
        player_name: event.actor_name ?? '',
        action_type: event.action_type ?? '',
        target_ids: event.target_user_ids || [],
        summary: event.summary || event.message || '',
      }
    case 'state_delta':
      return {
        event_type: 'state_delta',
        deltas: event.deltas || [],
        reason: event.reason,
      }
    default:
      return null
  }
}

function extractDeltaValue(event: TimelineEvent, deltaType: string): number | undefined {
  return event.deltas?.find(delta => delta.type === deltaType)?.value
}
