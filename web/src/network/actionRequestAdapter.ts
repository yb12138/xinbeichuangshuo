import type { Card, PlayerAction } from '../types/game'
import type { ClientActionRequest } from './protocol'

export interface PlayableCardEntry {
  card: Card
  index: number
}

export function buildClientActionRequest(
  action: PlayerAction,
  playableCards: PlayableCardEntry[]
): ClientActionRequest {
  const request: ClientActionRequest = {
    action_type: action.type,
  }

  if (action.skill_id) request.skill_id = action.skill_id

  const usedCardUUID = cardUUIDByActionIndex(action.card_index, playableCards)
  if (usedCardUUID) {
    request.used_card_uuids = [usedCardUUID]
  }

  const targets = buildActionTargets(action)
  if (targets?.length) {
    request.targets = targets
  }

  if (action.selections?.length) {
    request.option_indexes = action.selections
  }

  if (action.extra_args?.length) {
    request.extra_args = action.extra_args
  }

  return request
}

export function cardUUIDByActionIndex(cardIndex: number | undefined, playableCards: PlayableCardEntry[]): string | undefined {
  if (cardIndex === undefined || cardIndex === null || cardIndex < 0) return undefined
  const entry = playableCards[cardIndex]
  return entry?.card?.id
}

export function buildActionTargets(action: PlayerAction) {
  if (action.target_ids?.length) {
    return action.target_ids.map(targetId => ({ target_user_id: targetId }))
  }
  if (action.target_id) {
    return [{ target_user_id: action.target_id }]
  }
  return undefined
}
