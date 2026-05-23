import type { PlayerAction } from '../types/game'
import type { ClientActionRequest } from './protocol'

export function buildClientActionRequest(
  action: PlayerAction
): ClientActionRequest {
  const request: ClientActionRequest = {
    action_type: action.type,
  }

  if (action.skill_id) request.skill_id = action.skill_id

  if (action.card_id) {
    request.card_id = action.card_id
  }

  if (action.card_ids !== undefined) {
    request.card_ids = [...action.card_ids]
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

export function buildActionTargets(action: PlayerAction) {
  if (action.target_ids?.length) {
    return action.target_ids.map(targetId => ({ target_user_id: targetId }))
  }
  if (action.target_id) {
    return [{ target_user_id: action.target_id }]
  }
  return undefined
}
