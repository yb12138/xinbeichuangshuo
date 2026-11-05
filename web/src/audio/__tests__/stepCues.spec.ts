import { describe, expect, it } from 'vitest'
import { cueForStep } from '../stepCues'

describe('cueForStep', () => {
  it('maps combat beats to cues', () => {
    expect(cueForStep({ type: 'link', linkId: 'l', fromPlayerId: 'a', toPlayerId: 'b', color: 'attack' } as never)).toBe('attack')
    expect(cueForStep({ type: 'link-end', linkId: 'l', text: '', style: 'hit' })).toBe('hit')
    expect(cueForStep({ type: 'link-end', linkId: 'l', text: '', style: 'blocked' })).toBe('block')
    expect(cueForStep({ type: 'value-chip', playerId: 'a', text: '+1', color: 'heal' })).toBe('heal')
    expect(cueForStep({ type: 'turn-banner', text: '', duration: 800 })).toBe('turn')
  })

  it('stays silent for damage chips, bookkeeping steps and replays', () => {
    expect(cueForStep({ type: 'value-chip', playerId: 'a', text: '-2', color: 'damage' })).toBeNull()
    expect(cueForStep({ type: 'mark-previous' })).toBeNull()
    expect(cueForStep({ type: 'link-end', linkId: 'l', text: '', style: 'hit', holdMs: 0 })).toBeNull()
  })
})
