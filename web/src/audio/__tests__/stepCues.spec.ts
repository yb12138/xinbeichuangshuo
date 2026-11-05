import { describe, expect, it } from 'vitest'
import { cueForStep } from '../stepCues'

describe('cueForStep', () => {
  it('plays a sound when a card is played', () => {
    expect(cueForStep({ type: 'link', linkId: 'l', fromPlayerId: 'a', toPlayerId: 'b', color: 'attack' } as never)).toBe('attack')
    expect(cueForStep({ type: 'link', linkId: 'l', fromPlayerId: 'a', toPlayerId: 'b', color: 'magic' } as never)).toBe('magic')
  })

  it('stays silent for hits, morale, turns, draws, skills and replays', () => {
    expect(cueForStep({ type: 'narration-row', html: '红方士气 15 → 13', tag: '士气', tagColor: 'attack' } as never)).toBeNull()
    expect(cueForStep({ type: 'link-end', linkId: 'l', text: '', style: 'hit' })).toBeNull()
    expect(cueForStep({ type: 'turn-banner', text: '', duration: 800 })).toBeNull()
    expect(cueForStep({ type: 'deck-to-seat', playerId: 'a', count: 2, coverKey: '' })).toBeNull()
    expect(cueForStep({ type: 'link', linkId: 'l', fromPlayerId: 'a', toPlayerId: 'b', color: 'skill' } as never)).toBeNull()
    expect(cueForStep({ type: 'link', linkId: 'l', fromPlayerId: 'a', toPlayerId: 'b', color: 'attack', holdMs: 0 } as never)).toBeNull()
  })
})
