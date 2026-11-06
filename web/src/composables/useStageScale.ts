import { computed, onBeforeUnmount, onMounted, ref, type Ref } from 'vue'

/**
 * 逻辑舞台缩放：界面在一个"布局表现良好"的逻辑尺寸内排版，再整体 transform 缩放铺满屏幕。
 * - 中等屏（逻辑范围内）缩放为 1，与原生排版完全一致；
 * - 小屏按最小逻辑尺寸缩小（常规 1280×720，手机横屏等矮屏 1024×560），保证 6 个座位、手牌与行动面板都放得下；
 * - 大屏（1080p/2K/4K）按 1600×900 等比放大，避免元素过小、四周大片留白。
 *
 * 舞台元素是 `container: stage / size`，组件样式里的断点写成 `@container stage (...)`，
 * vw / vh 写成 `var(--vw)` / `var(--vh)`，都按逻辑尺寸解析，而不是真实视口。
 */
export const STAGE_MIN_WIDTH = 1280
export const STAGE_MIN_HEIGHT = 720
/**
 * 手机横屏等矮屏（短边 ≤ STAGE_COMPACT_MAX_SHORT_EDGE）改用更小的最小逻辑尺寸：
 * 否则会按 1280×720 整体缩到一半左右，字基本看不清。逻辑高度落进 `@container stage (max-height: 620px)`，
 * 组件在那里切换紧凑排版。
 */
export const STAGE_COMPACT_MIN_WIDTH = 1024
export const STAGE_COMPACT_MIN_HEIGHT = 560
export const STAGE_COMPACT_MAX_SHORT_EDGE = 600
/** 超过 1600×900 就开始等比放大：1080p 显示器上元素随屏放大，而不是只拉大中间的空场。 */
export const STAGE_UPSCALE_WIDTH = 1600
export const STAGE_UPSCALE_HEIGHT = 900

/** full：小屏缩小 + 大屏放大（对局）；upscale-only：只在大屏放大（大厅自身有响应式布局）。 */
export type StageScaleMode = 'full' | 'upscale-only'

/** 当前舞台缩放比例（视觉像素 / 布局像素）。没有舞台（如单测）时恒为 1。 */
export const stageScale = ref(1)
/** 舞台逻辑尺寸（布局像素），替代 window.innerWidth / innerHeight。 */
export const stageViewport = ref({
  width: typeof window !== 'undefined' ? window.innerWidth : 1280,
  height: typeof window !== 'undefined' ? window.innerHeight : 720,
})
/** App 挂载了舞台时为 true；组件的 Teleport 改投到舞台内的 #stage-portal，才能跟着一起缩放。 */
export const stagePortalEnabled = ref(false)

export function stagePortalTarget(): string {
  return stagePortalEnabled.value ? '#stage-portal' : 'body'
}

/** 把 getBoundingClientRect / clientX 等视觉坐标（或其差值）换算成舞台内的布局像素。 */
export function toStagePx(visualPx: number): number {
  return visualPx / (stageScale.value || 1)
}

/** 缩小时允许的最小逻辑尺寸：矮屏（手机横屏）用紧凑尺寸，其余用常规尺寸。 */
export function stageMinSize(width: number, height: number): { width: number; height: number } {
  return Math.min(width, height) <= STAGE_COMPACT_MAX_SHORT_EDGE
    ? { width: STAGE_COMPACT_MIN_WIDTH, height: STAGE_COMPACT_MIN_HEIGHT }
    : { width: STAGE_MIN_WIDTH, height: STAGE_MIN_HEIGHT }
}

export function computeStageScale(width: number, height: number, mode: StageScaleMode): number {
  if (width <= 0 || height <= 0) return 1
  const up = Math.min(width / STAGE_UPSCALE_WIDTH, height / STAGE_UPSCALE_HEIGHT)
  if (up > 1) return up
  if (mode === 'upscale-only') return 1
  const min = stageMinSize(width, height)
  return Math.min(1, width / min.width, height / min.height)
}

function readViewport() {
  const vv = window.visualViewport
  return vv
    ? { width: Math.round(vv.width), height: Math.round(vv.height) }
    : { width: window.innerWidth, height: window.innerHeight }
}

export function useStageScale(mode: Ref<StageScaleMode>) {
  stagePortalEnabled.value = true
  const viewport = ref(typeof window !== 'undefined' ? readViewport() : { width: 1280, height: 720 })

  const sync = () => {
    viewport.value = readViewport()
  }

  const scale = computed(() => computeStageScale(viewport.value.width, viewport.value.height, mode.value))
  const logicalSize = computed(() => ({
    width: viewport.value.width / scale.value,
    height: viewport.value.height / scale.value,
  }))

  const stageStyle = computed(() => {
    const s = scale.value
    const { width, height } = logicalSize.value
    stageScale.value = s
    stageViewport.value = { width, height }
    return {
      width: `${width}px`,
      height: `${height}px`,
      transform: s === 1 ? 'none' : `scale(${s})`,
      '--vw': `${width / 100}px`,
      '--vh': `${height / 100}px`,
      '--app-vw': `${width}px`,
      '--app-vh': `${height}px`,
    }
  })

  onMounted(() => {
    sync()
    window.addEventListener('resize', sync)
    window.addEventListener('orientationchange', sync)
    window.visualViewport?.addEventListener('resize', sync)
  })

  onBeforeUnmount(() => {
    window.removeEventListener('resize', sync)
    window.removeEventListener('orientationchange', sync)
    window.visualViewport?.removeEventListener('resize', sync)
    stageScale.value = 1
    stagePortalEnabled.value = false
  })

  return { scale, stageStyle }
}
