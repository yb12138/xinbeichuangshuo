import { describe, expect, it } from 'vitest'
import { endCueFor } from '../endCue'

describe('endCueFor', () => {
  it('picks victory or defeat from my camp', () => {
    expect(endCueFor('游戏结束：红方胜利', 'Red')).toBe('victory')
    expect(endCueFor('游戏结束：红方胜利', 'Blue')).toBe('defeat')
    expect(endCueFor('蓝方胜利（士气归零）', 'Blue')).toBe('victory')
  })

  it('falls back to the neutral cue for spectators or unknown results', () => {
    expect(endCueFor('红方胜利', undefined)).toBe('game-end')
    expect(endCueFor('游戏结束', 'Red')).toBe('game-end')
  })
})
