import { describe, expect, it } from 'vitest'
import { buildActionTargets, buildClientActionRequest } from '../actionRequestAdapter'
import type { PlayerAction } from '../../types/game'

describe('actionRequestAdapter', () => {
  it('builds a submit payload for targeted attack actions', () => {
    const action: PlayerAction = {
      player_id: 'p1',
      type: 'Attack',
      target_id: 'p2',
      card_id: 'attack-1',
    }

    const payload = buildClientActionRequest(action)

    expect(payload).toEqual({
      action_type: 'Attack',
      card_id: 'attack-1',
      targets: [{ target_user_id: 'p2' }],
    })
  })

  it('passes response extra_args through unchanged', () => {
    const action: PlayerAction = {
      player_id: 'p1',
      type: 'Respond',
      card_id: 'counter-1',
      target_id: 'p3',
      extra_args: ['counter', 'p3'],
    }

    const payload = buildClientActionRequest(action)

    expect(payload).toEqual({
      action_type: 'Respond',
      card_id: 'counter-1',
      targets: [{ target_user_id: 'p3' }],
      extra_args: ['counter', 'p3'],
    })
  })

  it('keeps targets and selections for skill-style multi target actions', () => {
    const action: PlayerAction = {
      player_id: 'p1',
      type: 'Skill',
      skill_id: 'skill-1',
      target_ids: ['p2', 'p3'],
      selections: [1, 2],
    }

    expect(buildActionTargets(action)).toEqual([
      { target_user_id: 'p2' },
      { target_user_id: 'p3' },
    ])
    const payload = buildClientActionRequest(action)

    expect(payload).toEqual({
      action_type: 'Skill',
      skill_id: 'skill-1',
      targets: [{ target_user_id: 'p2' }, { target_user_id: 'p3' }],
      option_indexes: [1, 2],
    })
  })

  it('passes selected card ids for card-picker select actions', () => {
    const action: PlayerAction = {
      player_id: 'p1',
      type: 'Select',
      card_ids: ['card-a', 'card-b'],
    }

    expect(buildClientActionRequest(action)).toEqual({
      action_type: 'Select',
      card_ids: ['card-a', 'card-b'],
    })
  })
})
