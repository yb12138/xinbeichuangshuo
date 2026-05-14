import { describe, expect, it } from 'vitest'
import { buildActionTargets, buildClientActionRequest, cardUUIDByActionIndex } from '../actionRequestAdapter'
import type { Card, PlayerAction } from '../../types/game'

function buildCard(overrides: Partial<Card> = {}): Card {
  return {
    id: 'card-1',
    name: '烈焰斩',
    type: 'Attack',
    element: 'Fire',
    damage: 2,
    description: 'test',
    ...overrides,
  }
}

describe('actionRequestAdapter', () => {
  it('builds a submit payload for targeted attack actions', () => {
    const action: PlayerAction = {
      player_id: 'p1',
      type: 'Attack',
      target_id: 'p2',
      card_index: 0,
    }

    const payload = buildClientActionRequest(action, [
      { index: 0, card: buildCard({ id: 'attack-1' }) },
    ])

    expect(payload).toEqual({
      action_type: 'Attack',
      used_card_uuids: ['attack-1'],
      targets: [{ target_user_id: 'p2' }],
    })
  })

  it('passes response extra_args through unchanged', () => {
    const action: PlayerAction = {
      player_id: 'p1',
      type: 'Respond',
      card_index: 1,
      target_id: 'p3',
      extra_args: ['counter', 'p3'],
    }

    const payload = buildClientActionRequest(action, [
      { index: 0, card: buildCard({ id: 'attack-1' }) },
      { index: 1, card: buildCard({ id: 'counter-1', name: '逆风斩' }) },
    ])

    expect(payload).toEqual({
      action_type: 'Respond',
      used_card_uuids: ['counter-1'],
      targets: [{ target_user_id: 'p3' }],
      extra_args: ['counter', 'p3'],
    })
  })

  it('keeps targets and selections for skill-style multi target actions and tolerates stale indexes', () => {
    const action: PlayerAction = {
      player_id: 'p1',
      type: 'Skill',
      skill_id: 'skill-1',
      target_ids: ['p2', 'p3'],
      selections: [1, 2],
      card_index: 99,
    }

    expect(buildActionTargets(action)).toEqual([
      { target_user_id: 'p2' },
      { target_user_id: 'p3' },
    ])
    expect(cardUUIDByActionIndex(99, [{ index: 0, card: buildCard() }])).toBeUndefined()

    const payload = buildClientActionRequest(action, [{ index: 0, card: buildCard() }])

    expect(payload).toEqual({
      action_type: 'Skill',
      skill_id: 'skill-1',
      targets: [{ target_user_id: 'p2' }, { target_user_id: 'p3' }],
      option_indexes: [1, 2],
    })
  })
})
