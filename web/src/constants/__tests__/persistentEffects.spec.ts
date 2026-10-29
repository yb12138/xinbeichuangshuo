import { describe, expect, it } from 'vitest'
import { campPersistentEffects, seatPersistentEffects } from '../persistentEffects'
import { seatTokenMarks } from '../seatMarks'
import type { PlayerView } from '../../types/game'

function player(partial: Partial<PlayerView>): PlayerView {
  return { id: 'p1', name: 'A1', camp: 'Blue', role: 'butterfly_dancer', indicators: {}, tokens: {}, field: [], ...partial } as PlayerView
}

describe('persistentEffects', () => {
  it('shows wither on the opposing camp and as a seat chip on the dancer', () => {
    const dancer = player({ indicators: { bt_wither_active: 1 } })
    const enemy = player({ id: 'p2', camp: 'Red', role: 'berserker' })
    const players = { p1: dancer, p2: enemy }
    expect(campPersistentEffects(players, 'Red').map((e) => [e.label, e.sourceId])).toEqual([['凋零', 'p1']])
    expect(campPersistentEffects(players, 'Blue')).toEqual([])
    expect(seatPersistentEffects(dancer).map((e) => e.label)).toEqual(['凋零'])
  })

  it('formats counts and bonuses, and keeps them out of the plain token list', () => {
    const moon = player({ role: 'moon_goddess', indicators: { mg_extra_turn_pending: 2, ml_fullness_next_attack_bonus: 3 } })
    expect(seatPersistentEffects(moon).map((e) => `${e.label}${e.valueText}`)).toEqual(['额外回合×2', '下击加伤+3'])
    expect(seatTokenMarks(moon).map((t) => t.key)).toEqual([])
  })
})
