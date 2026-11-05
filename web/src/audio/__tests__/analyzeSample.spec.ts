import { describe, expect, it } from 'vitest'
import { analyzeSample } from '../sfx'

/** 最小化的 AudioBuffer 替身：单声道，sampleRate 1000，前 100 个采样为静音。 */
function fakeBuffer(samples: number[], sampleRate = 1000): AudioBuffer {
  const data = Float32Array.from(samples)
  return { numberOfChannels: 1, length: data.length, sampleRate, getChannelData: () => data } as unknown as AudioBuffer
}

describe('analyzeSample', () => {
  it('skips leading silence and normalizes quiet samples toward the target', () => {
    const buffer = fakeBuffer([...new Array(100).fill(0), ...new Array(100).fill(0.01)])
    const { offset, gain } = analyzeSample(buffer, -20)
    expect(offset).toBeCloseTo(0.1 - 0.004, 5)
    // 目标 RMS 0.1 / 实际 0.01 = 10 倍，峰值 0.1 不触顶
    expect(gain).toBeCloseTo(10, 3)
  })

  it('caps gain so the peak stays below the ceiling', () => {
    const buffer = fakeBuffer([0.5, ...new Array(99).fill(0.01)])
    const { gain } = analyzeSample(buffer, -10)
    expect(0.5 * gain).toBeLessThanOrEqual(0.89 + 1e-9)
  })

  it('leaves silent buffers untouched', () => {
    expect(analyzeSample(fakeBuffer(new Array(10).fill(0)), -20)).toMatchObject({ offset: 0, gain: 1 })
  })
})
