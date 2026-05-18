import { describe, expect, it } from 'vitest'
import { buildGameStateUpdateFromSyncState } from '../syncState'
import type { SyncStatePayload } from '../protocol'

describe('buildGameStateUpdateFromSyncState', () => {
  it('converts structured SyncState payloads into the battle snapshot shape', () => {
    const payload: SyncStatePayload = {
      room_state: 'Playing',
      turn_stage: 'Main',
      turn_player_id: 'p1',
      has_performed_startup: true,
      morale_red: 14,
      morale_blue: 12,
      cups_red: 1,
      cups_blue: 2,
      stones_red: [3, 1],
      stones_blue: [2, 4],
      deck_count: 20,
      discard_count: 5,
      available_skills: [
        {
          id: 'skill-1',
          title: '烈焰爆发',
          description: 'test',
          min_targets: 1,
          max_targets: 1,
          target_type: 2,
          cost_gem: 1,
          cost_crystal: 0,
          cost_discards: 0,
        },
      ],
      characters: [
        {
          id: 'hero',
          name: '英雄',
          title: '测试角色',
          faction: 'fire',
          skills: [],
        },
      ],
      players: [
        {
          id: 'p1',
          name: 'Alice',
          camp: 'Red',
          role: 'hero',
          hand_count: 2,
          max_hand: 6,
          exclusive_card_count: 0,
          hand: [],
          exclusive_cards: [],
          field: [],
          heal: 1,
          max_heal: 5,
          gem: 1,
          crystal: 2,
          is_active: true,
          buffs: [],
          tokens: { flame: 1 },
        },
      ],
    }

    expect(buildGameStateUpdateFromSyncState(payload)).toEqual({
      turn_stage: 'Main',
      current_player: 'p1',
      has_performed_startup: true,
      players: {
        p1: payload.players[0],
      },
      red_morale: 14,
      blue_morale: 12,
      red_cups: 1,
      blue_cups: 2,
      red_gems: 3,
      blue_gems: 2,
      red_crystals: 1,
      blue_crystals: 4,
      deck_count: 20,
      discard_count: 5,
      available_skills: payload.available_skills,
      characters: payload.characters,
    })
  })
})
