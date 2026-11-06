import { describe, expect, it } from 'vitest'
import { computeStageScale, stageMinSize } from '../useStageScale'

describe('stageMinSize', () => {
  it('短边 ≤ 600 时使用紧凑最小尺寸', () => {
    expect(stageMinSize(750, 342)).toEqual({ width: 1024, height: 560 })
    expect(stageMinSize(568, 320)).toEqual({ width: 1024, height: 560 })
  })

  it('常规屏使用 1280×720', () => {
    expect(stageMinSize(1366, 768)).toEqual({ width: 1280, height: 720 })
    expect(stageMinSize(1024, 768)).toEqual({ width: 1280, height: 720 })
  })
})

describe('computeStageScale full', () => {
  it('逻辑范围内缩放为 1', () => {
    expect(computeStageScale(1280, 720, 'full')).toBeCloseTo(1)
    expect(computeStageScale(1600, 900, 'full')).toBeCloseTo(1)
    expect(computeStageScale(1366, 768, 'full')).toBeCloseTo(1)
  })

  it('大屏按 1600×900 等比放大', () => {
    expect(computeStageScale(1920, 1080, 'full')).toBeCloseTo(1.2)
    expect(computeStageScale(3840, 2160, 'full')).toBeCloseTo(2.4)
  })

  it('手机横屏按紧凑尺寸缩小', () => {
    const s = computeStageScale(750, 342, 'full')
    expect(s).toBeCloseTo(Math.min(750 / 1024, 342 / 560))
    expect(s).toBeCloseTo(0.6107, 3)
    expect(computeStageScale(568, 320, 'full')).toBeCloseTo(Math.min(568 / 1024, 320 / 560))
  })

  it('1024×768 非紧凑，按 1280 宽缩到 0.8', () => {
    expect(computeStageScale(1024, 768, 'full')).toBeCloseTo(0.8)
  })
})

describe('computeStageScale upscale-only', () => {
  it('只放大不缩小', () => {
    expect(computeStageScale(1280, 720, 'upscale-only')).toBeCloseTo(1)
    expect(computeStageScale(750, 342, 'upscale-only')).toBeCloseTo(1)
    expect(computeStageScale(1920, 1080, 'upscale-only')).toBeCloseTo(1.2)
  })
})

describe('computeStageScale 非法尺寸', () => {
  it('宽或高 ≤ 0 时返回 1', () => {
    expect(computeStageScale(0, 720, 'full')).toBe(1)
    expect(computeStageScale(1280, 0, 'full')).toBe(1)
    expect(computeStageScale(-1, -1, 'upscale-only')).toBe(1)
  })
})
