<script setup lang="ts">
/**
 * Battle "stage" drawn directly between player seats.
 *
 * Renders the state of useStageQueueStore (links, skill bubble, value chips, turn banner,
 * history chips, narration card) on top of the board and runs the store clock via rAF.
 * Visual reference: public/mockups/battle-readability.html.
 */
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import {
  useStageQueueStore,
  type LinkColor,
  type LinkEndStyle,
  type LinkState,
  type NarrationRow,
  type ValueChipColor,
} from '../stores/stageQueue.store'

const props = defineProps<{
  rootEl: HTMLElement | null
  /** Optional label for the skill bubble header (e.g. "圣女 · 测试玩家1"). */
  playerLabel?: (playerId: string) => string
}>()

const store = useStageQueueStore()

const LINK_COLORS: Record<LinkColor, string> = {
  attack: '#ff5d5d',
  magic: '#b48bff',
  skill: '#6fc3ff',
  heal: '#59e39a',
  grey: '#7d8799',
}
const TAG_BG: Record<LinkColor, string> = {
  attack: '#7a1f27',
  magic: '#4a2f80',
  skill: '#1d4f73',
  heal: '#1c6b45',
  grey: '#6b7690',
}
const END_STYLE: Record<LinkEndStyle, { bg: string; fg: string; prefix: string }> = {
  blocked: { bg: '#f0d27a', fg: '#2a1d00', prefix: '✕ ' },
  hit: { bg: '#a3262f', fg: '#fff', prefix: '' },
  miss: { bg: '#4a5368', fg: '#e8ecf4', prefix: '' },
  heal: { bg: '#1c7a49', fg: '#fff', prefix: '' },
}
const CHIP_BG: Record<ValueChipColor, string> = {
  damage: '#a3262f',
  heal: '#1c7a49',
  gain: '#2c6b2c',
  neutral: '#33425e',
}
const COLOR_KEYS = Object.keys(LINK_COLORS) as LinkColor[]
const MAX_ROWS = 5
const MAX_HISTORY = 4
const MEASURE_INTERVAL_MS = 250

interface SeatBox {
  left: number
  top: number
  right: number
  bottom: number
  cx: number
  cy: number
  /** Seat sits on the left half of the board (faces right toward the centre). */
  leftSide: boolean
}

const size = ref({ width: 0, height: 0 })
const seats = ref<Record<string, SeatBox>>({})

function involvedPlayerIds(): string[] {
  const ids = new Set<string>()
  for (const link of store.activeLinks.values()) {
    ids.add(link.fromPlayerId)
    ids.add(link.toPlayerId)
  }
  if (store.activeSkillBubble) ids.add(store.activeSkillBubble.playerId)
  for (const chip of store.valueChips) ids.add(chip.playerId)
  ids.delete('')
  return [...ids]
}

function measure() {
  const root = props.rootEl
  if (!root) {
    seats.value = {}
    return
  }
  const rootRect = root.getBoundingClientRect()
  const originX = rootRect.left + root.clientLeft
  const originY = rootRect.top + root.clientTop
  const width = root.clientWidth
  const height = root.clientHeight
  if (size.value.width !== width || size.value.height !== height) size.value = { width, height }

  const next: Record<string, SeatBox> = {}
  for (const id of involvedPlayerIds()) {
    const el = root.querySelector<HTMLElement>(`[data-player-anchor="${CSS.escape(id)}"]`)
    if (!el) continue
    const r = el.getBoundingClientRect()
    const left = r.left - originX
    const top = r.top - originY
    const box: SeatBox = {
      left,
      top,
      right: left + r.width,
      bottom: top + r.height,
      cx: left + r.width / 2,
      cy: top + r.height / 2,
      leftSide: left + r.width / 2 < width / 2,
    }
    next[id] = box
  }
  if (!sameSeats(seats.value, next)) seats.value = next
}

function sameSeats(a: Record<string, SeatBox>, b: Record<string, SeatBox>): boolean {
  const ak = Object.keys(a)
  if (ak.length !== Object.keys(b).length) return false
  return ak.every((k) => {
    const x = a[k]
    const y = b[k]
    return !!y && Math.abs(x!.left - y.left) < 0.5 && Math.abs(x!.top - y.top) < 0.5 && Math.abs(x!.right - y.right) < 0.5 && Math.abs(x!.bottom - y.bottom) < 0.5
  })
}

interface Point {
  x: number
  y: number
}

/** Seat edge facing the board centre; dominant axis decides side vs top/bottom edge. */
function anchorPoint(box: SeatBox): Point {
  const { width, height } = size.value
  const dx = width / 2 - box.cx
  const dy = height / 2 - box.cy
  if (Math.abs(dx) / Math.max(width, 1) >= Math.abs(dy) / Math.max(height, 1)) {
    return { x: dx > 0 ? box.right : box.left, y: box.cy }
  }
  return { x: box.cx, y: dy > 0 ? box.bottom : box.top }
}

function quadAt(a: Point, c: Point, b: Point, t: number): Point {
  const u = 1 - t
  return { x: u * u * a.x + 2 * u * t * c.x + t * t * b.x, y: u * u * a.y + 2 * u * t * c.y + t * t * b.y }
}

/** A blocked / missed link fades to grey like the mockup; others keep their colour. */
function strokeKey(link: LinkState): LinkColor {
  const end = link.end
  if (!link.animating && end && (end.style === 'blocked' || end.style === 'miss')) return 'grey'
  return link.color
}

interface LinkView {
  link: LinkState
  d: string
  stroke: string
  marker: string
  mid: Point
  endPos: Point
  endStyle: { bg: string; fg: string; prefix: string } | null
}

const linkViews = computed<LinkView[]>(() => {
  const views: LinkView[] = []
  const { width, height } = size.value
  for (const link of store.activeLinks.values()) {
    const fromBox = seats.value[link.fromPlayerId]
    const toBox = seats.value[link.toPlayerId]
    if (!fromBox || !toBox) continue
    const a = anchorPoint(fromBox)
    let b = anchorPoint(toBox)
    if (link.fromPlayerId === link.toPlayerId) {
      // Self-target: short loop out toward the centre.
      b = { x: a.x, y: a.y + 28 }
    }
    let c: Point
    if (Math.abs(a.x - b.x) < 5) {
      // Same column: bulge toward the board centre.
      const toward = width / 2 > a.x ? 1 : -1
      c = { x: a.x + toward * Math.max(120, Math.abs(a.y - b.y) * 0.45), y: (a.y + b.y) / 2 }
    } else if (Math.abs(a.y - b.y) < 5 && height > 0 && Math.abs(a.x - b.x) < width * 0.25) {
      const toward = height / 2 > a.y ? 1 : -1
      c = { x: (a.x + b.x) / 2, y: a.y + toward * 120 }
    } else {
      c = { x: (a.x + b.x) / 2, y: Math.min(a.y, b.y) - 70 }
    }
    const end = link.end ? END_STYLE[link.end.style] : null
    const nearEnd = quadAt(a, c, b, 0.82)
    const key = strokeKey(link)
    views.push({
      link,
      d: `M${a.x.toFixed(1)},${a.y.toFixed(1)} Q${c.x.toFixed(1)},${c.y.toFixed(1)} ${b.x.toFixed(1)},${b.y.toFixed(1)}`,
      stroke: LINK_COLORS[key],
      marker: `url(#stage-ah-${key})`,
      mid: quadAt(a, c, b, 0.5),
      endPos: { x: nearEnd.x, y: nearEnd.y - 24 },
      endStyle: end,
    })
  }
  return views
})

const bubbleView = computed(() => {
  const bubble = store.activeSkillBubble
  if (!bubble) return null
  const box = seats.value[bubble.playerId]
  if (!box) return null
  const gap = 14
  const style: Record<string, string> = { top: `${Math.max(4, box.top + 4)}px` }
  if (box.leftSide) style.left = `${box.right + gap}px`
  else style.right = `${size.value.width - box.left + gap}px`
  return {
    bubble,
    style,
    pointLeft: box.leftSide,
    who: props.playerLabel ? props.playerLabel(bubble.playerId) : '',
  }
})

const chipViews = computed(() => {
  const perPlayer = new Map<string, number>()
  const views: { id: number; text: string; style: Record<string, string> }[] = []
  for (const chip of store.valueChips) {
    const box = seats.value[chip.playerId]
    if (!box) continue
    const index = perPlayer.get(chip.playerId) ?? 0
    perPlayer.set(chip.playerId, index + 1)
    const anchor = anchorPoint(box)
    const x = anchor.x + (box.leftSide ? 36 : -36)
    const y = box.cy - 18 - index * 34
    views.push({ id: chip.id, text: chip.text, style: { left: `${x}px`, top: `${y}px`, background: CHIP_BG[chip.color] } })
  }
  return views
})

const narration = computed<{ rows: NarrationRow[]; previous: boolean } | null>(() => {
  if (store.stageRows.length > 0) return { rows: store.stageRows.slice(-MAX_ROWS), previous: false }
  if (store.previousAction && store.previousAction.rows.length > 0) {
    return { rows: store.previousAction.rows.slice(-MAX_ROWS), previous: true }
  }
  return null
})

const historyView = computed(() => store.historyChips.slice(-MAX_HISTORY))

// --- clock + measurement -------------------------------------------------

let rafId = 0
let lastMeasure = 0
let observer: ResizeObserver | null = null

function frame(timestamp: number) {
  store.tick(timestamp)
  // Seats can shift without a root resize (layout changes, hover scale); re-measure periodically
  // while there is something anchored to them.
  if (timestamp - lastMeasure >= MEASURE_INTERVAL_MS) {
    lastMeasure = timestamp
    if (store.activeLinks.size > 0 || store.activeSkillBubble || store.valueChips.length > 0) measure()
  }
  rafId = requestAnimationFrame(frame)
}

function observe(root: HTMLElement | null) {
  observer?.disconnect()
  observer = null
  if (root && typeof ResizeObserver !== 'undefined') {
    observer = new ResizeObserver(() => measure())
    observer.observe(root)
  }
  measure()
}

watch(() => props.rootEl, (root) => observe(root))
watch(
  () => involvedPlayerIds().join('|'),
  () => measure(),
  { flush: 'post' },
)

onMounted(() => {
  observe(props.rootEl)
  window.addEventListener('resize', measure)
  rafId = requestAnimationFrame(frame)
})

onBeforeUnmount(() => {
  cancelAnimationFrame(rafId)
  observer?.disconnect()
  observer = null
  window.removeEventListener('resize', measure)
})
</script>

<template>
  <div class="stage-overlay" aria-hidden="true" data-testid="stage-overlay">
    <svg class="stage-links" :width="size.width" :height="size.height" :viewBox="`0 0 ${size.width || 1} ${size.height || 1}`">
      <defs>
        <marker
          v-for="key in COLOR_KEYS"
          :id="`stage-ah-${key}`"
          :key="key"
          viewBox="0 0 10 10"
          refX="7"
          refY="5"
          markerWidth="7"
          markerHeight="7"
          orient="auto-start-reverse"
        >
          <path d="M0,0 L10,5 L0,10 z" :fill="LINK_COLORS[key]" />
        </marker>
      </defs>
      <g
        v-for="view in linkViews"
        :key="view.link.id"
        class="stage-link"
        :class="{ 'stage-link--previous': view.link.previous }"
      >
        <path
          class="stage-link__base"
          :class="{ 'stage-link__base--muted': view.link.animating }"
          :d="view.d"
          :stroke="view.stroke"
          stroke-width="4"
          pathLength="1"
          :marker-end="view.marker"
        />
        <path v-if="view.link.animating" class="stage-link__flow" :d="view.d" :stroke="LINK_COLORS[view.link.color]" stroke-width="4" />
      </g>
    </svg>

    <template v-for="view in linkViews" :key="`label-${view.link.id}`">
      <div
        class="stage-label"
        :class="{ 'stage-dim': view.link.previous }"
        :style="{ left: `${view.mid.x}px`, top: `${view.mid.y}px`, color: view.stroke }"
      >
        <img v-if="view.link.cardImageUrl" :src="view.link.cardImageUrl" alt="" />
        <span>{{ view.link.label }}</span>
      </div>
      <div
        v-if="view.link.end && view.endStyle"
        class="stage-end"
        :class="{ 'stage-dim': view.link.previous }"
        :style="{ left: `${view.endPos.x}px`, top: `${view.endPos.y}px`, background: view.endStyle.bg, color: view.endStyle.fg }"
      >
        {{ view.endStyle.prefix }}{{ view.link.end.text }}
      </div>
    </template>

    <div
      v-if="bubbleView"
      :key="bubbleView.bubble.id"
      class="stage-bubble"
      :class="bubbleView.pointLeft ? 'stage-bubble--point-left' : 'stage-bubble--point-right'"
      :style="bubbleView.style"
    >
      <div v-if="bubbleView.who" class="stage-bubble__who">{{ bubbleView.who }} 发动技能</div>
      <div class="stage-bubble__name">{{ bubbleView.bubble.skillName }}</div>
      <div v-if="bubbleView.bubble.effectText" class="stage-bubble__desc">{{ bubbleView.bubble.effectText }}</div>
    </div>

    <div v-for="chip in chipViews" :key="chip.id" class="stage-chip" :style="chip.style">{{ chip.text }}</div>

    <div v-if="store.turnBanner" :key="store.turnBanner.text" class="stage-banner">{{ store.turnBanner.text }}</div>

    <div v-if="historyView.length" class="stage-history">
      <span v-for="(chip, index) in historyView" :key="`${index}-${chip}`" class="stage-hchip">{{ chip }}</span>
    </div>

    <div v-if="narration" class="stage-story" :class="{ 'stage-story--prev': narration.previous }">
      <div class="stage-story__head">
        <span class="stage-story__tag">{{ narration.previous ? '上一手' : '当前行动' }}</span>
      </div>
      <div v-for="row in narration.rows" :key="row.id" class="stage-row">
        <!-- Row html is pre-escaped by the stage queue builders. -->
        <span class="stage-row__text" v-html="row.html" />
        <span class="stage-row__tag" :style="{ background: narration.previous ? TAG_BG.grey : TAG_BG[row.tagColor] }">{{ row.tag }}</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.stage-overlay {
  position: absolute;
  inset: 0;
  pointer-events: none;
  z-index: 30;
  overflow: hidden;
  color: #e8ecf4;
  font-family: 'PingFang SC', 'Microsoft YaHei', system-ui, sans-serif;
}

.stage-links {
  position: absolute;
  inset: 0;
  overflow: visible;
}
.stage-links path {
  fill: none;
  stroke-linecap: round;
}
.stage-link {
  transition: opacity 0.6s;
}
.stage-link--previous {
  opacity: 0.25;
}
.stage-link--previous .stage-link__base {
  stroke-dasharray: 0.02 0.015;
}
.stage-link__base {
  stroke-dasharray: 1;
  stroke-dashoffset: 0;
  animation: stage-draw 0.6s ease both;
  transition: stroke 0.3s, opacity 0.3s;
}
.stage-link__base--muted {
  opacity: 0.35;
}
.stage-link__flow {
  stroke-dasharray: 12 10;
  animation: stage-flow 0.7s linear infinite;
}

.stage-label {
  position: absolute;
  transform: translate(-50%, -50%);
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  font-weight: 600;
  padding: 4px 10px;
  border-radius: 12px;
  background: rgba(10, 14, 24, 0.92);
  border: 1px solid currentColor;
  white-space: nowrap;
  animation: stage-label-in 0.3s ease both;
  transition: opacity 0.6s, color 0.3s;
}
.stage-label span {
  color: #fff;
}
.stage-label img {
  width: 22px;
  height: 22px;
  border-radius: 4px;
  object-fit: cover;
}
.stage-end {
  position: absolute;
  transform: translate(-50%, -50%);
  font-size: 13px;
  font-weight: 700;
  padding: 3px 9px;
  border-radius: 10px;
  white-space: nowrap;
  animation: stage-pop 0.35s ease both;
  transition: opacity 0.6s;
}
.stage-dim {
  opacity: 0.25;
}

.stage-bubble {
  position: absolute;
  width: 250px;
  max-width: 34%;
  padding: 10px 14px;
  border-radius: 12px;
  background: linear-gradient(135deg, rgba(28, 64, 92, 0.96), rgba(16, 30, 52, 0.96));
  border: 1px solid #6fc3ff;
  box-shadow: 0 0 24px rgba(111, 195, 255, 0.35);
  animation: stage-bubble-in 0.35s ease both;
}
.stage-bubble__who {
  font-size: 12px;
  color: #9fd4ff;
}
.stage-bubble__name {
  font-size: 20px;
  font-weight: 800;
  letter-spacing: 2px;
  margin: 2px 0;
}
.stage-bubble__desc {
  font-size: 12px;
  color: #c6d6ea;
}
.stage-bubble::after {
  content: '';
  position: absolute;
  top: 24px;
  border: 9px solid transparent;
}
.stage-bubble--point-right::after {
  right: -9px;
  border-left-color: #6fc3ff;
  border-right: 0;
}
.stage-bubble--point-left::after {
  left: -9px;
  border-right-color: #6fc3ff;
  border-left: 0;
}

.stage-chip {
  position: absolute;
  transform: translate(-50%, -50%);
  font-weight: 700;
  font-size: 16px;
  padding: 4px 12px;
  border-radius: 14px;
  white-space: nowrap;
  color: #fff;
  box-shadow: 0 4px 14px rgba(0, 0, 0, 0.5);
  animation: stage-pop 0.3s ease both;
}

.stage-banner {
  position: absolute;
  left: 50%;
  top: 26%;
  transform: translateX(-50%);
  padding: 10px 40px;
  font-size: 24px;
  font-weight: 800;
  letter-spacing: 3px;
  border-radius: 8px;
  white-space: nowrap;
  background: linear-gradient(90deg, transparent, rgba(120, 20, 30, 0.9) 20%, rgba(120, 20, 30, 0.9) 80%, transparent);
  animation: stage-banner-in 0.35s ease both;
}

.stage-history {
  position: absolute;
  top: 64px;
  left: 18%;
  right: 18%;
  display: flex;
  gap: 8px;
  justify-content: center;
  align-items: center;
  overflow: hidden;
}
.stage-hchip {
  font-size: 12px;
  padding: 4px 10px;
  border-radius: 14px;
  background: rgba(16, 24, 40, 0.85);
  border: 1px solid #33425e;
  color: #b9c4d8;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 320px;
  animation: stage-fade-in 0.3s ease both;
}

.stage-story {
  position: absolute;
  left: 50%;
  top: 46%;
  transform: translateX(-50%);
  width: min(560px, 44%);
  border-radius: 14px;
  background: rgba(9, 14, 26, 0.9);
  border: 1px solid #3b4d6e;
  padding: 10px 14px 8px;
  transition: filter 0.5s, border-color 0.5s;
}
.stage-story--prev {
  filter: saturate(0.35) brightness(0.8);
  border-style: dashed;
}
.stage-story__head {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
}
.stage-story__tag {
  font-size: 12px;
  font-weight: 700;
  padding: 1px 8px;
  border-radius: 8px;
  background: #c88a1e;
  color: #1b1205;
}
.stage-story--prev .stage-story__tag {
  background: #6b7690;
  color: #fff;
}
.stage-row {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 14px;
  padding: 4px 0;
  border-top: 1px solid rgba(255, 255, 255, 0.06);
  animation: stage-row-in 0.35s ease both;
}
.stage-row:first-of-type {
  border-top: 0;
}
.stage-row__text {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.stage-row__text :deep(b) {
  color: #fff;
  font-weight: 700;
}
.stage-row__tag {
  font-size: 11px;
  padding: 1px 6px;
  border-radius: 6px;
  margin-left: auto;
  flex: none;
}

@keyframes stage-draw {
  from {
    stroke-dashoffset: 1;
  }
  to {
    stroke-dashoffset: 0;
  }
}
@keyframes stage-flow {
  from {
    stroke-dashoffset: 44;
  }
  to {
    stroke-dashoffset: 0;
  }
}
@keyframes stage-label-in {
  from {
    opacity: 0;
    transform: translate(-50%, -30%);
  }
}
@keyframes stage-fade-in {
  from {
    opacity: 0;
    transform: scale(0.6);
  }
}
@keyframes stage-pop {
  from {
    opacity: 0;
    transform: translate(-50%, -50%) scale(1.5);
  }
}
@keyframes stage-bubble-in {
  from {
    opacity: 0;
    transform: scale(0.9);
  }
}
@keyframes stage-banner-in {
  from {
    opacity: 0;
    transform: translateX(-50%) scaleX(0.6);
  }
}
@keyframes stage-row-in {
  from {
    opacity: 0;
    transform: translateX(-12px);
  }
}

@media (prefers-reduced-motion: reduce) {
  .stage-overlay *,
  .stage-overlay *::after {
    animation: none !important;
    transition: none !important;
  }
}
</style>
