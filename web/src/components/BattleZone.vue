<script setup lang="ts">
import gsap from 'gsap'
import { storeToRefs } from 'pinia'
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useBattleFxStore } from '../stores/battlefx.store'
import type {
  ActionNarrativeCardView,
  ActionNarrativeEventView,
  NarrativeLinkKind,
  NarrativePlaybackStepKind,
  NarrativePlaybackStepStatus,
  NarrativePlaybackStepView,
} from '../stores/battlefx.store'
import { useSnapshotStore } from '../stores/snapshot.store'
import type { PlayerView } from '../types/game'
import CardComponent from './CardComponent.vue'
import RoseCourtyardIcon from './StatusIcons/RoseCourtyardIcon.vue'

const props = defineProps<{
  narrativeSuspended?: boolean
  actorSeatPositions?: Record<string, number>
}>()

const battleFxStore = useBattleFxStore()
const snapshotStore = useSnapshotStore()
const { actionNarrative, narrativePlayback, skillAnnouncements } = storeToRefs(battleFxStore)
const { characters, players } = storeToRefs(snapshotStore)

const featuredSkillAnnouncement = computed(() =>
  skillAnnouncements.value.find((item) => item.phase === 'featured') ?? null
)

const settledSkillAnnouncements = computed(() =>
  skillAnnouncements.value
    .filter((item) => item.phase === 'settled')
    .slice(-3)
)

const hasRoseCourtyard = computed(() => {
  return Object.values(players.value).some(p =>
    p.field?.some(fc => fc.mode === 'Effect' && fc.effect === 'RoseCourtyard')
  )
})

function isPlayerView(player: PlayerView | undefined): player is PlayerView {
  return !!player
}

const hasNarrativeLayer = computed(() => !!actionNarrative.value || hasRoseCourtyard.value)

const featuredPlayer = computed(() => {
  const id = actionNarrative.value?.featuredActorId
  return id ? players.value[id] : undefined
})

const opposedPlayers = computed(() =>
  (actionNarrative.value?.opposedActorIds || [])
    .map(id => players.value[id])
    .filter(isPlayerView)
)

const settledPlayers = computed(() =>
  (actionNarrative.value?.settledActorIds || [])
    .filter(id => id !== actionNarrative.value?.featuredActorId && !(actionNarrative.value?.opposedActorIds || []).includes(id))
    .map(id => players.value[id])
    .filter(isPlayerView)
)

const narrativeCards = computed(() => actionNarrative.value?.playedCards || [])
const narrativeEvents = computed(() => actionNarrative.value?.events || [])
const narrativeLinks = computed(() => actionNarrative.value?.links || [])

type NarrativeSide = 'left' | 'right'
type NarrativeRow = 'top' | 'middle' | 'bottom'
type NarrativeActorRole = 'featured' | 'opposed' | 'settled'
type NarrativeStackItemKind = 'card' | 'skill' | 'damage'

interface NarrativeStackItem {
  id: string
  kind: NarrativeStackItemKind
  sourcePlayerId: string
  targetIds: string[]
  actionKind: NarrativeLinkKind
  createdAt: number
  stackIndex: number
  stepId?: string
  stepKind?: NarrativePlaybackStepKind
  stepStatus?: NarrativePlaybackStepStatus
  isActiveStep?: boolean
  isContextStep?: boolean
  cardView?: ActionNarrativeCardView
  eventView?: ActionNarrativeEventView
  title?: string
  detail?: string
  damage?: number
}

interface NarrativeStepGroup {
  id: string
  kind: NarrativePlaybackStepKind
  label: string
  status: NarrativePlaybackStepStatus
  order: number
  items: NarrativeStackItem[]
  step?: NarrativePlaybackStepView
}

interface NarrativeMistSegment {
  id: string
  kind: NarrativeLinkKind
  fromType: 'actor' | 'stack'
  fromId: string
  toType: 'actor' | 'stack'
  toId: string
  sourcePlayerId: string
  targetPlayerId?: string
  stackItemId?: string
  damage?: number
  x1: number
  y1: number
  x2: number
  y2: number
  path: string
  damageX: number
  damageY: number
  particleSeeds: number[]
}

const narrativeLayerRef = ref<HTMLElement | null>(null)
const narrativeMistLayerRef = ref<SVGSVGElement | null>(null)
const measuredNarrativeMistSegments = ref<NarrativeMistSegment[]>([])
let narrativeMistResizeObserver: ResizeObserver | null = null
let narrativeMistMeasureFrame = 0
let narrativeGsapContext: gsap.Context | null = null
let narrativeGsapTimeline: gsap.core.Timeline | null = null
let lastAnimatedNarrativeStepId = ''

function roleNameForPlayer(playerId?: string) {
  if (!playerId) return ''
  const player = players.value[playerId]
  if (!player) return playerId
  const roleId = String(player.role || '').trim()
  return roleId ? (characters.value[roleId]?.name || roleId) : (player.name || player.id)
}

function portraitSrcForPlayer(playerId?: string) {
  const roleId = String(players.value[playerId || '']?.role || '').trim()
  return roleId ? `/characters/${roleId}.png` : ''
}

function latestDamageForPlayer(playerId?: string) {
  if (!playerId) return null
  return [...narrativeEvents.value].reverse().find(event => event.kind === 'damage' && event.targetId === playerId) ?? null
}

function actorSeatIndex(playerId?: string): number | null {
  if (!playerId) return null
  const index = props.actorSeatPositions?.[playerId]
  return typeof index === 'number' && Number.isInteger(index) && index >= 0 && index <= 5 ? index : null
}

function actorSideForSeat(index: number | null, fallback: NarrativeActorRole): NarrativeSide {
  if (index === null) return fallback === 'featured' ? 'left' : 'right'
  return index <= 2 ? 'left' : 'right'
}

function actorRowForSeat(index: number | null): NarrativeRow {
  if (index === null) return 'middle'
  if (index === 0 || index === 3) return 'top'
  if (index === 2 || index === 5) return 'bottom'
  return 'middle'
}

function narrativeSideForPlayer(playerId?: string, fallback: NarrativeActorRole = 'featured') {
  return actorSideForSeat(actorSeatIndex(playerId), fallback)
}

function narrativeRowForPlayer(playerId?: string) {
  return actorRowForSeat(actorSeatIndex(playerId))
}

function rowPercent(row: NarrativeRow) {
  if (row === 'top') return 22
  if (row === 'bottom') return 78
  return 50
}

function actorLineX(side: NarrativeSide) {
  return side === 'left' ? 14 : 86
}

function targetLineX(side: NarrativeSide) {
  return side === 'left' ? 16 : 84
}

function narrativeActorCardClasses(player: PlayerView, role: NarrativeActorRole) {
  const seatIndex = actorSeatIndex(player.id)
  const side = actorSideForSeat(seatIndex, role)
  const row = actorRowForSeat(seatIndex)

  return [
    `narrative-actor-card--${role}`,
    `narrative-actor-card--${player.camp === 'Red' ? 'red' : 'blue'}`,
    `narrative-actor-card--side-${side}`,
    `narrative-actor-card--row-${row}`,
    seatIndex === null ? `narrative-actor-card--fallback-${role}` : `narrative-actor-card--seat-${seatIndex}`,
  ]
}

function narrativePlayedCardClasses(item: ActionNarrativeCardView) {
  return [
    `narrative-played-card--${item.actionType}`,
  ]
}

function narrativeActionKindLabel(kind: NarrativeLinkKind) {
  if (kind === 'respond') return '应战'
  if (kind === 'skill') return '技能'
  if (kind === 'damage') return '伤害'
  return '发起攻击'
}

function narrativeCardActionLabel(actionType: string) {
  const normalized = String(actionType || '').trim().toLowerCase()
  if (normalized === 'counter') return '应战'
  if (normalized === 'defend') return '防御'
  if (normalized === 'shield') return '圣盾'
  if (normalized === 'magic') return '法术'
  if (normalized === 'field_effect') return '施加封印'
  if (normalized === 'discard' || normalized === 'skill_cost') return '弃牌'
  return '发起攻击'
}

function narrativeSettledCardClasses(player: PlayerView) {
  const side = narrativeSideForPlayer(player.id, 'settled')
  const row = narrativeRowForPlayer(player.id)
  return [
    `narrative-settled-card--${player.camp === 'Red' ? 'red' : 'blue'}`,
    `narrative-settled-card--side-${side}`,
    `narrative-settled-card--row-${row}`,
  ]
}

function narrativeSettledCardStyle(player: PlayerView) {
  const row = narrativeRowForPlayer(player.id)
  return {
    '--settled-card-y': `${rowPercent(row)}%`,
  }
}

function linkKindForActionType(actionType: string): NarrativeLinkKind {
  const normalized = String(actionType || '').trim().toLowerCase()
  if (normalized === 'counter' || normalized === 'defend' || normalized === 'shield') return 'respond'
  if (normalized === 'skill' || normalized === 'magic' || normalized === 'discard' || normalized === 'skill_cost' || normalized === 'field_effect') return 'skill'
  return 'attack'
}

function skillDisplayFromEvent(event: ActionNarrativeEventView) {
  const label = String(event.label || '').trim()
  const quoted = label.match(/发动「(.+?)」/)
  if (quoted?.[1]) {
    return { title: quoted[1], detail: '' }
  }
  const separated = label.match(/^(.+?)[：:](.+)$/)
  if (separated?.[1] && separated?.[2]) {
    return { title: separated[1].trim(), detail: separated[2].trim() }
  }
  return {
    title: label || '技能发动',
    detail: '',
  }
}

function skillTargetIdsForEvent(event: ActionNarrativeEventView) {
  const ids = new Set<string>()
  if (event.targetId) ids.add(event.targetId)
  for (const link of narrativeLinks.value) {
    if (link.kind === 'skill' && link.fromType === 'actor' && link.fromId === event.actorId) {
      ids.add(link.toPlayerId)
    }
  }
  return [...ids]
}

function playbackStepForItemId(itemId: string) {
  return narrativePlayback.value?.steps.find(step => step.itemIds.includes(itemId))
}

function fallbackStepForItem(item: Omit<NarrativeStackItem, 'stackIndex'>): NarrativePlaybackStepView {
  const cardAction = item.cardView?.actionType || ''
  const normalized = String(cardAction).trim().toLowerCase()
  const kind: NarrativePlaybackStepKind =
    item.kind === 'damage' ? 'damage' :
    item.kind === 'card' && ['counter', 'defend', 'shield'].includes(normalized) ? 'response' :
    item.kind === 'card' && normalized === 'field_effect' ? 'skill' :
    item.kind === 'card' && ['discard', 'skill_cost'].includes(normalized) ? 'discard' :
    item.kind === 'card' ? 'combat' :
    'skill'
  return {
    id: item.id,
    kind,
    label: item.kind === 'card' ? narrativeCardActionLabel(cardAction) : item.title || item.eventView?.label || '叙事',
    actorId: item.sourcePlayerId,
    targetIds: item.targetIds,
    itemIds: [item.id],
    order: item.createdAt,
    durationMs: 940,
    status: 'active',
  }
}

function previousPlaybackStepId() {
  const playback = narrativePlayback.value
  if (!playback?.activeStepId) return undefined
  const activeIndex = playback.steps.findIndex(step => step.id === playback.activeStepId)
  if (activeIndex <= 0) return undefined
  return playback.steps[activeIndex - 1]?.id
}

const narrativeStackItems = computed<NarrativeStackItem[]>(() => {
  const items: Omit<NarrativeStackItem, 'stackIndex'>[] = [
    ...narrativeCards.value.map((cardView) => ({
      id: `card-${cardView.id}`,
      kind: 'card' as const,
      sourcePlayerId: cardView.playerId,
      targetIds: cardView.targetId ? [cardView.targetId] : [],
      actionKind: linkKindForActionType(cardView.actionType),
      createdAt: cardView.createdAt,
      cardView,
    })),
    ...narrativeEvents.value
      .filter(event => event.kind === 'skill')
      .map((eventView) => {
        const display = skillDisplayFromEvent(eventView)
        return {
          id: `skill-${eventView.id}`,
          kind: 'skill' as const,
          sourcePlayerId: eventView.actorId || actionNarrative.value?.featuredActorId || actionNarrative.value?.currentActionPlayerId || '',
          targetIds: skillTargetIdsForEvent(eventView),
          actionKind: 'skill' as const,
          createdAt: eventView.createdAt,
          eventView,
          title: display.title,
          detail: display.detail,
        }
      }),
    ...narrativeEvents.value
      .filter(event => event.kind === 'damage')
      .map((eventView) => ({
        id: `damage-${eventView.id}`,
        kind: 'damage' as const,
        sourcePlayerId: eventView.actorId || actionNarrative.value?.featuredActorId || actionNarrative.value?.currentActionPlayerId || '',
        targetIds: eventView.targetId ? [eventView.targetId] : [],
        actionKind: 'damage' as const,
        createdAt: eventView.createdAt,
        eventView,
        title: '伤害',
        detail: eventView.label,
        damage: eventView.damage,
      })),
  ]

  return items
    .filter(item => item.sourcePlayerId)
    .sort((a, b) => a.createdAt - b.createdAt)
    .map((item, stackIndex) => {
      const step = playbackStepForItemId(item.id) || fallbackStepForItem(item)
      const contextStepId = previousPlaybackStepId()
      return {
        ...item,
        stackIndex,
        stepId: step.id,
        stepKind: step.kind,
        stepStatus: step.status,
        isActiveStep: step.status === 'active',
        isContextStep: step.id === contextStepId,
      }
    })
})

const narrativeStepGroups = computed<NarrativeStepGroup[]>(() => {
  const groups = new Map<string, NarrativeStepGroup>()

  for (const item of narrativeStackItems.value) {
    if (isSealFieldEffectItem(item)) continue
    const step = item.stepId ? narrativePlayback.value?.steps.find(entry => entry.id === item.stepId) : undefined
    const id = item.stepId || item.id
    const existing = groups.get(id)
    if (existing) {
      existing.items.push(item)
      continue
    }
    groups.set(id, {
      id,
      kind: step?.kind || item.stepKind || 'skill',
      label: (step?.kind || item.stepKind) === 'skill'
        ? item.title || step?.label || '技能'
        : step?.label || item.title || item.eventView?.label || '叙事',
      status: step?.status || item.stepStatus || 'active',
      order: step?.order || item.createdAt,
      items: [item],
      step,
    })
  }

  return [...groups.values()]
    .filter(group => group.status !== 'pending' || narrativePlayback.value?.isReview)
    .sort((a, b) => a.order - b.order || a.id.localeCompare(b.id))
})

function isSealFieldEffectItem(item: NarrativeStackItem) {
  return item.kind === 'card' && item.cardView?.actionType === 'field_effect'
}

const narrativeSealFieldEffectItems = computed(() =>
  narrativeStackItems.value.filter((item) => {
    if (!isSealFieldEffectItem(item)) return false
    if (narrativePlayback.value?.isReview) return true
    return !narrativePlayback.value || item.stepStatus !== 'pending'
  })
)

function narrativeStackItemClasses(item: NarrativeStackItem) {
  const sourceSide = narrativeSideForPlayer(item.sourcePlayerId)
  return [
    `narrative-stack-item--${item.kind}`,
    `narrative-stack-item--${item.actionKind}`,
    `narrative-stack-item--from-${sourceSide}`,
    item.stepStatus ? `narrative-stack-item--step-${item.stepStatus}` : '',
    item.isActiveStep ? 'narrative-stack-item--active-step' : '',
    item.isContextStep ? 'narrative-stack-item--context-step' : '',
    item.stackIndex === narrativeStackItems.value.length - 1 ? 'narrative-stack-item--latest' : '',
  ]
}

function narrativeStackItemStyle(item: NarrativeStackItem) {
  return {
    '--stack-enter-x': narrativeSideForPlayer(item.sourcePlayerId) === 'left' ? '-42px' : '42px',
    '--stack-order': String(item.stackIndex),
    zIndex: String(20 + item.stackIndex),
  }
}

function narrativeStepGroupClasses(group: NarrativeStepGroup) {
  return [
    `narrative-step-group--${group.kind}`,
    `narrative-step-group--${group.status}`,
  ]
}

function findNarrativeElement(kind: 'actor' | 'stack', id: string | number) {
  const root = narrativeLayerRef.value
  if (!root) return null
  const attr = kind === 'actor' ? 'narrativeActorId' : 'narrativeStackId'
  const selector = kind === 'actor' ? '[data-narrative-actor-id]' : '[data-narrative-stack-id]'
  return Array.from(root.querySelectorAll<HTMLElement>(selector))
    .find(el => el.dataset[attr] === String(id)) ?? null
}

function actorPointForSide(el: HTMLElement, layerRect: DOMRect, side: NarrativeSide) {
  const rect = el.getBoundingClientRect()
  return {
    x: (side === 'left' ? rect.right : rect.left) - layerRect.left,
    y: rect.top + rect.height / 2 - layerRect.top,
  }
}

function stackPointForToward(el: HTMLElement, layerRect: DOMRect, towardX: number) {
  const rect = el.getBoundingClientRect()
  const centerX = rect.left + rect.width / 2 - layerRect.left
  return {
    x: (towardX < centerX ? rect.left : rect.right) - layerRect.left,
    y: rect.top + rect.height / 2 - layerRect.top,
  }
}

function fallbackActorPoint(playerId: string, layerRect: DOMRect, targetSide = narrativeSideForPlayer(playerId)) {
  const row = narrativeRowForPlayer(playerId)
  const xPercent = actorLineX(targetSide)
  return {
    x: layerRect.width * xPercent / 100,
    y: layerRect.height * rowPercent(row) / 100,
  }
}

function fallbackTargetPoint(playerId: string, layerRect: DOMRect) {
  const side = narrativeSideForPlayer(playerId, 'opposed')
  const row = narrativeRowForPlayer(playerId)
  return {
    x: layerRect.width * targetLineX(side) / 100,
    y: layerRect.height * rowPercent(row) / 100,
  }
}

function fallbackStackPoint(layerRect: DOMRect) {
  return {
    x: layerRect.width * 0.5,
    y: layerRect.height * 0.5,
  }
}

function bezierPath(x1: number, y1: number, x2: number, y2: number) {
  const dx = Math.abs(x2 - x1)
  const curve = Math.max(42, dx * 0.42)
  const direction = x2 >= x1 ? 1 : -1
  const lift = Math.min(34, Math.abs(y2 - y1) * 0.22)
  const c1x = x1 + curve * direction
  const c2x = x2 - curve * direction
  const c1y = y1 - lift
  const c2y = y2 + lift
  return `M ${x1.toFixed(1)} ${y1.toFixed(1)} C ${c1x.toFixed(1)} ${c1y.toFixed(1)}, ${c2x.toFixed(1)} ${c2y.toFixed(1)}, ${x2.toFixed(1)} ${y2.toFixed(1)}`
}

function latestStackItemForDamage(event: ActionNarrativeEventView) {
  const actorItems = narrativeStackItems.value
    .filter(item => item.sourcePlayerId === event.actorId && item.createdAt <= event.createdAt)
  const byActor = actorItems[actorItems.length - 1]
  return byActor ?? narrativeStackItems.value[narrativeStackItems.value.length - 1]
}

const narrativeMistBlueprints = computed(() => {
  const segments: Array<Omit<NarrativeMistSegment, 'x1' | 'y1' | 'x2' | 'y2' | 'path' | 'damageX' | 'damageY' | 'particleSeeds'>> = []
  const visibleItems = narrativeStackItems.value.filter((item) => {
    if (narrativePlayback.value?.isReview) return false
    return !narrativePlayback.value || item.isActiveStep || item.isContextStep
  })

  for (const item of visibleItems) {
    segments.push({
      id: `${item.id}-actor-to-stack`,
      kind: item.actionKind,
      fromType: 'actor',
      fromId: item.sourcePlayerId,
      toType: 'stack',
      toId: item.id,
      sourcePlayerId: item.sourcePlayerId,
      stackItemId: item.id,
    })

    for (const targetId of item.targetIds) {
      segments.push({
        id: `${item.id}-stack-to-${targetId}`,
        kind: item.actionKind,
        fromType: 'stack',
        fromId: item.id,
        toType: 'actor',
        toId: targetId,
        sourcePlayerId: item.sourcePlayerId,
        targetPlayerId: targetId,
        stackItemId: item.id,
      })
    }
  }

  for (const event of visibleItems
    .filter(item => item.kind === 'damage' && item.eventView?.targetId && item.damage)
    .map(item => item.eventView)
    .filter((event): event is ActionNarrativeEventView => !!event)) {
    const stackItem = latestStackItemForDamage(event)
    if (!stackItem || !event.targetId) continue
    segments.push({
      id: `damage-${event.id}-${stackItem.id}-to-${event.targetId}`,
      kind: 'damage',
      fromType: 'stack',
      fromId: stackItem.id,
      toType: 'actor',
      toId: event.targetId,
      sourcePlayerId: stackItem.sourcePlayerId,
      targetPlayerId: event.targetId,
      stackItemId: stackItem.id,
      damage: event.damage,
    })
  }

  return segments
})

function pointForEndpoint(
  endpointType: 'actor' | 'stack',
  endpointId: string,
  layerRect: DOMRect,
  otherPoint?: { x: number; y: number },
) {
  if (endpointType === 'actor') {
    const side = narrativeSideForPlayer(endpointId, 'opposed')
    const element = findNarrativeElement('actor', endpointId)
    return element
      ? actorPointForSide(element, layerRect, side)
      : fallbackActorPoint(endpointId, layerRect, side)
  }

  const element = findNarrativeElement('stack', endpointId)
  return element
    ? stackPointForToward(element, layerRect, otherPoint?.x ?? layerRect.width / 2)
    : fallbackStackPoint(layerRect)
}

function particleSeedsForSegment(id: string) {
  const base = id.split('').reduce((sum, char) => sum + char.charCodeAt(0), 0)
  return Array.from({ length: 8 }, (_, index) => ((base + index * 17) % 100) / 100)
}

function measureNarrativeMist() {
  const layer = narrativeMistLayerRef.value
  if (!layer) {
    measuredNarrativeMistSegments.value = []
    return
  }

  const layerRect = layer.getBoundingClientRect()
  measuredNarrativeMistSegments.value = narrativeMistBlueprints.value.map((segment) => {
    const roughTarget = segment.toType === 'actor'
      ? fallbackTargetPoint(segment.toId, layerRect)
      : fallbackStackPoint(layerRect)
    const sourcePoint = pointForEndpoint(segment.fromType, segment.fromId, layerRect, roughTarget)
    const targetPoint = pointForEndpoint(segment.toType, segment.toId, layerRect, sourcePoint)

    return {
      ...segment,
      x1: sourcePoint.x,
      y1: sourcePoint.y,
      x2: targetPoint.x,
      y2: targetPoint.y,
      path: bezierPath(sourcePoint.x, sourcePoint.y, targetPoint.x, targetPoint.y),
      damageX: targetPoint.x,
      damageY: targetPoint.y,
      particleSeeds: particleSeedsForSegment(segment.id),
    }
  })

  scheduleNarrativeAnimation()
}

function scheduleNarrativeMistMeasure() {
  if (narrativeMistMeasureFrame && typeof cancelAnimationFrame === 'function') {
    cancelAnimationFrame(narrativeMistMeasureFrame)
  }
  void nextTick(() => {
    const runMeasure = () => {
      narrativeMistMeasureFrame = 0
      measureNarrativeMist()
    }
    if (typeof requestAnimationFrame === 'function') {
      narrativeMistMeasureFrame = requestAnimationFrame(runMeasure)
      return
    }
    runMeasure()
  })
}

function prefersReducedNarrativeMotion() {
  return typeof window !== 'undefined' &&
    typeof window.matchMedia === 'function' &&
    window.matchMedia('(prefers-reduced-motion: reduce)').matches
}

function scheduleNarrativeAnimation() {
  void nextTick(() => {
    const root = narrativeLayerRef.value
    if (!root) return

    narrativeGsapTimeline?.kill()
    narrativeGsapContext?.revert()
    narrativeGsapContext = gsap.context(() => {
      const activeStepId = narrativePlayback.value?.activeStepId || 'fallback'
      const shouldAnimateStep = activeStepId !== lastAnimatedNarrativeStepId
      const stackItems = shouldAnimateStep
        ? gsap.utils.toArray<HTMLElement>('.narrative-step-group--active .narrative-stack-item')
        : []
      const skillTokens = shouldAnimateStep
        ? gsap.utils.toArray<HTMLElement>('.narrative-step-group--active .narrative-skill-token')
        : []
      const skillShines = shouldAnimateStep
        ? gsap.utils.toArray<HTMLElement>('.narrative-step-group--active .narrative-skill-token__shine')
        : []
      const paths = gsap.utils.toArray<SVGPathElement>('.narrative-mist__flow')
      const particles = gsap.utils.toArray<SVGCircleElement>('.narrative-mist-particle')
      lastAnimatedNarrativeStepId = activeStepId

      if (prefersReducedNarrativeMotion()) {
        gsap.set([...stackItems, ...skillTokens, ...skillShines, ...paths, ...particles], { clearProps: 'all' })
        gsap.set(skillShines, { opacity: 0 })
        gsap.set(particles, { opacity: 0.42 })
        return
      }

      narrativeGsapTimeline = gsap.timeline()
      if (stackItems.length) {
        narrativeGsapTimeline.from(stackItems, {
          opacity: 0,
          x: (_index, target) => Number.parseFloat(getComputedStyle(target).getPropertyValue('--stack-enter-x')) || 0,
          y: 18,
          scale: 0.82,
          rotate: -4,
          duration: 0.34,
          stagger: 0.045,
          ease: 'back.out(1.7)',
        }, 0)
      }

      if (skillTokens.length) {
        narrativeGsapTimeline.fromTo(skillTokens, {
          filter: 'brightness(1)',
        }, {
          filter: 'brightness(1.24) drop-shadow(0 0 12px rgba(255, 238, 180, 0.52))',
          duration: 0.18,
          yoyo: true,
          repeat: 1,
          ease: 'sine.inOut',
        }, 0.1)
      }

      if (skillShines.length) {
        gsap.set(skillShines, { xPercent: -135, opacity: 0 })
        narrativeGsapTimeline.to(skillShines, {
          xPercent: 135,
          opacity: 0.82,
          duration: 0.42,
          ease: 'power2.out',
          stagger: 0.04,
        }, 0.12)
        narrativeGsapTimeline.to(skillShines, {
          opacity: 0,
          duration: 0.16,
          ease: 'power1.out',
        }, 0.42)
      }

      for (const path of paths) {
        const length = typeof path.getTotalLength === 'function' ? path.getTotalLength() : 0
        if (length > 0) {
          gsap.set(path, { strokeDasharray: length, strokeDashoffset: length })
          narrativeGsapTimeline.to(path, {
            strokeDashoffset: 0,
            duration: 0.58,
            ease: 'power2.out',
          }, 0.06)
        }
      }

      for (const particle of particles) {
        const pathId = particle.dataset.mistPathId
        const path = pathId ? root.querySelector<SVGPathElement>(`#${pathId}`) : null
        const length = path && typeof path.getTotalLength === 'function' ? path.getTotalLength() : 0
        if (!path || length <= 0) continue
        const seed = Number.parseFloat(particle.dataset.seed || '0')
        const state = { progress: seed }
        gsap.to(state, {
          progress: seed + 1,
          duration: 1.6 + seed * 0.48,
          repeat: -1,
          ease: 'none',
          onUpdate: () => {
            const progress = state.progress % 1
            const point = path.getPointAtLength(length * progress)
            particle.setAttribute('cx', point.x.toFixed(1))
            particle.setAttribute('cy', point.y.toFixed(1))
            particle.style.opacity = String(0.24 + Math.sin(progress * Math.PI) * 0.58)
          },
        })
      }
    }, root)
  })
}

watch(
  () => [
    narrativeMistBlueprints.value.map(segment => `${segment.id}:${segment.kind}:${segment.fromId}:${segment.toId}:${segment.damage || ''}`).join('|'),
    narrativeStackItems.value.map(item => `${item.id}:${item.stackIndex}:${item.sourcePlayerId}:${item.targetIds.join(',')}`).join('|'),
    narrativeCards.value.map(card => `${card.id}:${card.playerId}:${card.targetId || ''}`).join('|'),
    narrativeEvents.value.map(event => `${event.id}:${event.kind}:${event.actorId || ''}:${event.targetId || ''}:${event.damage || ''}`).join('|'),
    opposedPlayers.value.map(player => player.id).join('|'),
    settledPlayers.value.map(player => player.id).join('|'),
    featuredPlayer.value?.id || '',
  ],
  () => scheduleNarrativeMistMeasure(),
  { flush: 'post' },
)

onMounted(() => {
  if (typeof ResizeObserver !== 'undefined' && narrativeMistLayerRef.value) {
    narrativeMistResizeObserver = new ResizeObserver(() => scheduleNarrativeMistMeasure())
    narrativeMistResizeObserver.observe(narrativeMistLayerRef.value)
  }
  window.addEventListener('resize', scheduleNarrativeMistMeasure)
  scheduleNarrativeMistMeasure()
})

onBeforeUnmount(() => {
  if (narrativeMistMeasureFrame && typeof cancelAnimationFrame === 'function') {
    cancelAnimationFrame(narrativeMistMeasureFrame)
    narrativeMistMeasureFrame = 0
  }
  narrativeMistResizeObserver?.disconnect()
  narrativeMistResizeObserver = null
  narrativeGsapTimeline?.kill()
  narrativeGsapTimeline = null
  narrativeGsapContext?.revert()
  narrativeGsapContext = null
  window.removeEventListener('resize', scheduleNarrativeMistMeasure)
})

function narrativeMistSegmentClasses(segment: NarrativeMistSegment) {
  return [
    `narrative-mist--${segment.kind}`,
    `narrative-mist--from-${segment.fromType}`,
    `narrative-mist--to-${segment.toType}`,
  ]
}

function narrativeMistPathDomId(segmentId: string) {
  return `narrative-mist-path-${segmentId.replace(/[^a-zA-Z0-9_-]/g, '-')}`
}
</script>

<template>
  <div class="battle-zone battle-zone-shell min-h-[90px]">
    <div
      v-if="hasNarrativeLayer"
      ref="narrativeLayerRef"
      class="action-narrative-layer"
      :class="{ 'action-narrative-layer--suspended': props.narrativeSuspended }"
    >
      <div v-if="actionNarrative" class="narrative-actors">
        <div
          v-if="featuredPlayer"
          class="narrative-actor-card narrative-actor-card--featured"
          :class="narrativeActorCardClasses(featuredPlayer, 'featured')"
          :data-narrative-actor-id="featuredPlayer.id"
        >
          <img
            v-if="portraitSrcForPlayer(featuredPlayer.id)"
            class="narrative-actor-card__portrait"
            :src="portraitSrcForPlayer(featuredPlayer.id)"
            :alt="roleNameForPlayer(featuredPlayer.id)"
          >
          <div class="narrative-actor-card__overlay">
            <span class="narrative-actor-card__camp">{{ featuredPlayer.camp === 'Red' ? '红方' : '蓝方' }}</span>
            <strong>{{ roleNameForPlayer(featuredPlayer.id) }}</strong>
            <span>{{ featuredPlayer.name || featuredPlayer.id }}</span>
          </div>
          <div
            v-if="latestDamageForPlayer(featuredPlayer.id)"
            class="narrative-damage-pop"
          >
            -{{ latestDamageForPlayer(featuredPlayer.id)?.damage }}
          </div>
        </div>

        <div class="narrative-opposed-stack">
          <div
            v-for="player in opposedPlayers"
            :key="`opposed-${player.id}`"
            class="narrative-actor-card narrative-actor-card--opposed"
            :class="narrativeActorCardClasses(player, 'opposed')"
            :data-narrative-actor-id="player.id"
          >
            <img
              v-if="portraitSrcForPlayer(player.id)"
              class="narrative-actor-card__portrait"
              :src="portraitSrcForPlayer(player.id)"
              :alt="roleNameForPlayer(player.id)"
            >
            <div class="narrative-actor-card__overlay">
              <strong>{{ roleNameForPlayer(player.id) }}</strong>
              <span>{{ player.name || player.id }}</span>
            </div>
            <div
              v-if="latestDamageForPlayer(player.id)"
              class="narrative-damage-pop"
            >
              -{{ latestDamageForPlayer(player.id)?.damage }}
            </div>
          </div>
        </div>
      </div>

      <div
        v-if="actionNarrative && narrativeSealFieldEffectItems.length"
        class="narrative-seal-card-stage"
      >
        <div
          v-for="item in narrativeSealFieldEffectItems"
          :key="`narrative-seal-card-${item.id}`"
          class="narrative-stack-item narrative-stack-item--seal-card"
          :class="narrativeStackItemClasses(item)"
          :style="narrativeStackItemStyle(item)"
          :data-narrative-stack-id="item.id"
          :data-narrative-card-id="item.cardView?.id"
        >
          <div
            v-if="item.cardView"
            class="narrative-played-card"
            :class="narrativePlayedCardClasses(item.cardView)"
          >
            <CardComponent :card="item.cardView.card" battle-mini />
          </div>
        </div>
      </div>

      <div
        v-if="actionNarrative && narrativeStepGroups.length"
        class="narrative-stack-lane"
        :class="{ 'narrative-stack-lane--review': narrativePlayback?.isReview }"
      >
        <div
          v-for="group in narrativeStepGroups"
          :key="`narrative-step-${group.id}`"
          class="narrative-step-group"
          :class="narrativeStepGroupClasses(group)"
          :data-narrative-step-id="group.id"
        >
          <div
            v-if="group.kind !== 'skill'"
            class="narrative-step-group__label"
          >
            {{ group.label }}
          </div>
          <div class="narrative-step-group__items">
            <div
              v-for="item in group.items"
              :key="`narrative-stack-${item.id}`"
              class="narrative-stack-item"
              :class="narrativeStackItemClasses(item)"
              :style="narrativeStackItemStyle(item)"
              :data-narrative-stack-id="item.id"
              :data-narrative-card-id="item.cardView?.id"
              :data-narrative-skill-id="item.kind === 'skill' ? item.id : undefined"
            >
              <div
                v-if="item.cardView"
                class="narrative-played-card"
                :class="narrativePlayedCardClasses(item.cardView)"
              >
                <CardComponent :card="item.cardView.card" battle-mini />
                <div class="narrative-stack-item__caption">
                  <span>{{ roleNameForPlayer(item.sourcePlayerId) }}</span>
                  <strong>{{ narrativeActionKindLabel(item.actionKind) }}</strong>
                  <span v-if="item.targetIds[0]">→ {{ roleNameForPlayer(item.targetIds[0]) }}</span>
                </div>
              </div>

              <div
                v-else-if="item.kind === 'damage'"
                class="narrative-damage-token"
              >
                <span>伤害</span>
                <strong>-{{ item.damage }}</strong>
                <small v-if="item.targetIds[0]">{{ roleNameForPlayer(item.targetIds[0]) }}</small>
              </div>

              <div
                v-else
                class="narrative-skill-token"
              >
                <div class="narrative-skill-token__shine"></div>
                <div class="narrative-skill-token__body">
                  <strong>{{ item.title }}</strong>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <svg
        v-if="actionNarrative && narrativeMistBlueprints.length"
        ref="narrativeMistLayerRef"
        class="narrative-mist-layer"
        aria-hidden="true"
      >
        <defs>
          <filter id="narrative-mist-blur" x="-25%" y="-80%" width="150%" height="260%">
            <feGaussianBlur stdDeviation="2.6" />
          </filter>
        </defs>
        <g
          v-for="segment in measuredNarrativeMistSegments"
          :key="`narrative-mist-${segment.id}`"
          class="narrative-mist"
          :class="narrativeMistSegmentClasses(segment)"
          :data-narrative-mist-id="segment.id"
        >
          <path
            class="narrative-mist__aura"
            :d="segment.path"
          />
          <path
            :id="narrativeMistPathDomId(segment.id)"
            class="narrative-mist__flow"
            :d="segment.path"
          />
          <circle
            v-for="(seed, index) in segment.particleSeeds"
            :key="`particle-${segment.id}-${index}`"
            class="narrative-mist-particle"
            :class="index % 3 === 0 ? 'narrative-mist-particle--bright' : ''"
            :cx="segment.x1"
            :cy="segment.y1"
            :r="index % 3 === 0 ? 2.2 : 1.35"
            :data-mist-path-id="narrativeMistPathDomId(segment.id)"
            :data-seed="seed"
          />
          <text
            v-if="segment.damage"
            class="narrative-mist__damage-label"
            :x="segment.damageX"
            :y="segment.damageY - 12"
            text-anchor="middle"
            dominant-baseline="middle"
          >
            -{{ segment.damage }}
          </text>
        </g>
      </svg>

      <div v-if="actionNarrative" class="narrative-settled-row">
        <div
          v-for="player in settledPlayers"
          :key="`settled-${player.id}`"
          class="narrative-settled-card"
          :class="narrativeSettledCardClasses(player)"
          :style="narrativeSettledCardStyle(player)"
          :data-narrative-actor-id="player.id"
        >
          <img
            v-if="portraitSrcForPlayer(player.id)"
            :src="portraitSrcForPlayer(player.id)"
            :alt="roleNameForPlayer(player.id)"
          >
          <span>{{ roleNameForPlayer(player.id) }}</span>
        </div>
      </div>

      <div
        v-if="hasRoseCourtyard"
        class="narrative-field-effect narrative-field-effect--rose"
        :class="{ 'narrative-field-effect--ambient': !actionNarrative }"
      >
        <div class="narrative-field-effect__icon">
          <RoseCourtyardIcon />
        </div>
        <div class="narrative-field-effect__body">
          <strong>血蔷薇庭院</strong>
          <span>玩家无法使用治疗抵消伤害</span>
        </div>
      </div>
    </div>

    <Transition name="skill-plaque-featured">
      <div
        v-if="featuredSkillAnnouncement"
        :key="featuredSkillAnnouncement.id"
        class="skill-plaque skill-plaque--featured"
      >
        <div class="skill-plaque__glow"></div>
        <div class="skill-plaque__body">
          <div class="skill-plaque__eyebrow">{{ featuredSkillAnnouncement.actorName }} 发动技能</div>
          <div class="skill-plaque__title">{{ featuredSkillAnnouncement.skillName }}</div>
          <div class="skill-plaque__effect">{{ featuredSkillAnnouncement.effectText }}</div>
        </div>
      </div>
    </Transition>

    <TransitionGroup name="skill-plaque-settled" tag="div" class="skill-plaque-stack">
      <div
        v-for="item in settledSkillAnnouncements"
        :key="item.id"
        class="skill-plaque skill-plaque--settled"
      >
        <div class="skill-plaque__body">
          <div class="skill-plaque__eyebrow">{{ item.actorName }}</div>
          <div class="skill-plaque__title">{{ item.skillName }}</div>
          <div class="skill-plaque__effect">{{ item.effectText }}</div>
        </div>
      </div>
    </TransitionGroup>
  </div>
</template>

<style scoped>
.battle-zone-shell {
  width: 100%;
  height: 100%;
  min-height: 0;
  position: relative;
  overflow: hidden;
  border-radius: 0;
  border: 0;
  background: transparent;
  padding: 10px 10px 8px;
}

.action-narrative-layer {
  position: absolute;
  inset: 0;
  z-index: 6;
  pointer-events: none;
  transition: opacity 0.18s ease, filter 0.18s ease;
  --narrative-actor-card-edge: clamp(0px, 1.2vw, 16px);
  --narrative-actor-card-width: clamp(106px, 9.5vw, 142px);
  --narrative-actor-card-height: clamp(148px, 13.4vw, 198px);
}

.action-narrative-layer--suspended {
  opacity: 0.34;
  filter: saturate(0.72) brightness(0.72);
}

.action-narrative-layer--suspended * {
  animation-play-state: paused !important;
}

.narrative-actors {
  position: absolute;
  inset: 14px 18px 54px;
  display: block;
}

.narrative-opposed-stack {
  display: contents;
}

.narrative-actor-card {
  position: absolute;
  --actor-enter-x: 0px;
  width: var(--narrative-actor-card-width);
  height: var(--narrative-actor-card-height);
  overflow: hidden;
  border-radius: 12px;
  background: rgba(9, 18, 31, 0.88);
  box-shadow:
    inset 0 1px 0 rgba(237, 248, 255, 0.12),
    0 14px 34px rgba(0, 0, 0, 0.36);
  isolation: isolate;
}

.narrative-actor-card--featured {
  z-index: 3;
  animation: narrativeActorIn 0.34s cubic-bezier(0.2, 0.86, 0.24, 1) both;
}

.narrative-actor-card--opposed {
  z-index: 2;
  animation: narrativeActorIn 0.28s cubic-bezier(0.2, 0.86, 0.24, 1) both;
}

.narrative-actor-card--side-left {
  left: var(--narrative-actor-card-edge);
  --actor-enter-x: -24px;
  transform-origin: left center;
}

.narrative-actor-card--side-right {
  right: var(--narrative-actor-card-edge);
  --actor-enter-x: 24px;
  transform-origin: right center;
}

.narrative-actor-card--row-top {
  top: 0;
}

.narrative-actor-card--row-middle {
  top: calc(50% - var(--narrative-actor-card-height) / 2);
}

.narrative-actor-card--row-bottom {
  bottom: 0;
}

.narrative-actor-card--blue {
  border: 1px solid rgba(82, 190, 250, 0.72);
  box-shadow:
    inset 0 1px 0 rgba(237, 248, 255, 0.12),
    0 0 24px rgba(56, 189, 248, 0.28),
    0 14px 34px rgba(0, 0, 0, 0.36);
}

.narrative-actor-card--red {
  border: 1px solid rgba(248, 113, 113, 0.72);
  box-shadow:
    inset 0 1px 0 rgba(255, 237, 237, 0.12),
    0 0 24px rgba(248, 113, 113, 0.28),
    0 14px 34px rgba(0, 0, 0, 0.36);
}

.narrative-actor-card__portrait {
  width: 100%;
  height: 100%;
  border-radius: inherit;
  object-fit: cover;
  object-position: 50% 12%;
}

.narrative-actor-card__overlay {
  position: absolute;
  inset: auto 0 0;
  z-index: 2;
  min-height: 46%;
  padding: 8px;
  display: flex;
  flex-direction: column;
  justify-content: flex-end;
  align-items: center;
  gap: 3px;
  text-align: center;
  color: rgba(238, 246, 255, 0.96);
  background: linear-gradient(180deg, rgba(3, 8, 14, 0), rgba(4, 10, 18, 0.86) 48%, rgba(3, 8, 14, 0.96));
  text-shadow: 0 1px 3px rgba(0, 0, 0, 0.84);
}

.narrative-actor-card__overlay strong {
  max-width: 100%;
  overflow: hidden;
  font-size: clamp(12px, 1.3vw, 16px);
  line-height: 1.1;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.narrative-actor-card__overlay span {
  max-width: 100%;
  overflow: hidden;
  font-size: 10px;
  line-height: 1.1;
  color: rgba(196, 218, 235, 0.92);
  white-space: nowrap;
  text-overflow: ellipsis;
}

.narrative-actor-card__camp {
  padding: 2px 6px;
  border-radius: 999px;
  border: 1px solid rgba(226, 241, 255, 0.34);
  background: rgba(5, 14, 25, 0.68);
}

.narrative-damage-pop {
  position: absolute;
  top: 42%;
  left: 50%;
  z-index: 5;
  color: #ffb199;
  font-size: clamp(28px, 4vw, 46px);
  font-weight: 900;
  line-height: 1;
  text-shadow:
    0 2px 4px rgba(0, 0, 0, 0.9),
    0 0 18px rgba(248, 113, 113, 0.7);
  transform: translate(-50%, -50%);
  animation: narrativeDamagePop 0.86s ease-out both;
}

.narrative-stack-lane {
  position: absolute;
  inset: 14px calc(var(--narrative-actor-card-edge) + var(--narrative-actor-card-width) + 18px) 54px;
  z-index: 8;
  display: flex;
  flex-wrap: wrap;
  align-content: center;
  justify-content: center;
  align-items: flex-start;
  gap: 10px 12px;
  min-width: 0;
  overflow: hidden auto;
  pointer-events: none;
  scrollbar-width: thin;
  scrollbar-color: rgba(162, 190, 214, 0.32) transparent;
}

.narrative-seal-card-stage {
  position: absolute;
  inset: 14px calc(var(--narrative-actor-card-edge) + var(--narrative-actor-card-width) + 18px) 54px;
  z-index: 8;
  display: flex;
  flex-wrap: wrap;
  align-content: center;
  justify-content: center;
  align-items: center;
  gap: 10px;
  pointer-events: none;
}

.narrative-stack-item--seal-card {
  width: 82px;
}

.narrative-stack-lane--review {
  align-content: flex-start;
  justify-content: center;
  padding: 8px 4px;
}

.narrative-step-group {
  position: relative;
  flex: 0 1 auto;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 5px;
  max-width: 192px;
  padding: 6px 7px 7px;
  border-radius: 10px;
  border: 1px solid rgba(136, 173, 203, 0.2);
  background: rgba(4, 12, 22, 0.34);
  transition:
    opacity 0.24s ease,
    transform 0.24s ease,
    filter 0.24s ease,
    border-color 0.24s ease,
    background 0.24s ease;
}

.narrative-step-group--active {
  z-index: 4;
  border-color: rgba(254, 226, 150, 0.58);
  background: rgba(13, 25, 38, 0.58);
  filter:
    drop-shadow(0 12px 22px rgba(0, 0, 0, 0.4))
    drop-shadow(0 0 15px rgba(246, 220, 153, 0.2));
  transform: scale(1);
}

.narrative-step-group--completed {
  z-index: 1;
  opacity: 0.58;
  filter: saturate(0.7);
  transform: scale(0.84);
  transform-origin: center center;
}

.narrative-stack-lane--review .narrative-step-group--completed {
  opacity: 0.9;
  transform: scale(0.9);
}

.narrative-step-group__label {
  max-width: 100%;
  overflow: hidden;
  padding: 3px 8px;
  border-radius: 999px;
  border: 1px solid rgba(177, 209, 235, 0.22);
  background: rgba(3, 10, 19, 0.7);
  color: rgba(231, 241, 249, 0.9);
  font-size: 10px;
  font-weight: 900;
  line-height: 1;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.narrative-step-group--active .narrative-step-group__label {
  color: #ffe8a8;
  border-color: rgba(251, 226, 153, 0.48);
}

.narrative-step-group__items {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  align-items: flex-start;
  gap: 8px 10px;
}

.narrative-stack-item {
  position: relative;
  flex: 0 0 auto;
  width: 78px;
  transform: translateY(calc(var(--stack-order, 0) * -1px));
  transform-origin: center center;
  filter: drop-shadow(0 10px 20px rgba(0, 0, 0, 0.45));
  animation: narrativeStackItemIn 0.32s cubic-bezier(0.2, 0.86, 0.24, 1) both;
}

.narrative-stack-item--step-completed {
  animation: none;
}

.narrative-stack-item--latest {
  filter:
    drop-shadow(0 12px 24px rgba(0, 0, 0, 0.54))
    drop-shadow(0 0 12px rgba(246, 220, 153, 0.28));
}

.narrative-played-card {
  width: 78px;
  transform-origin: center center;
}

.narrative-stack-item--skill {
  width: 112px;
}

.narrative-stack-item--damage {
  width: 92px;
}

.narrative-stack-item--respond .narrative-played-card,
.narrative-stack-item--skill .narrative-played-card {
  filter: drop-shadow(0 0 10px rgba(125, 211, 252, 0.34));
}

.narrative-played-card--field_effect {
  filter:
    drop-shadow(0 0 10px rgba(190, 219, 255, 0.36))
    drop-shadow(0 0 16px rgba(167, 139, 250, 0.2));
}

.narrative-stack-item--damage .narrative-played-card {
  filter: drop-shadow(0 0 10px rgba(248, 113, 113, 0.36));
}

.narrative-played-card :deep(.card-battle-mini) {
  transform: none !important;
}

.narrative-stack-item__caption {
  margin-top: 4px;
  width: max-content;
  max-width: 132px;
  transform: translateX(calc(39px - 50%));
  overflow: hidden;
  padding: 3px 6px;
  border-radius: 999px;
  border: 1px solid rgba(150, 185, 214, 0.36);
  background: rgba(5, 14, 25, 0.7);
  color: rgba(225, 238, 248, 0.94);
  font-size: 10px;
  line-height: 1;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.narrative-stack-item__caption strong {
  color: #ffe8a8;
  font-weight: 900;
}

.narrative-skill-token {
  position: relative;
  width: 112px;
  min-height: 42px;
  padding: 4px 6px;
  display: grid;
  place-items: center;
  overflow: hidden;
  border: 0;
  border-radius: 0;
  background: transparent;
  box-shadow: none;
}

.narrative-skill-token__shine {
  position: absolute;
  inset: -8px -18px;
  z-index: 0;
  opacity: 0;
  background: linear-gradient(
    100deg,
    transparent 25%,
    rgba(255, 246, 190, 0.76) 48%,
    rgba(199, 210, 254, 0.48) 54%,
    transparent 76%
  );
  mix-blend-mode: screen;
  transform: translateX(-135%);
  pointer-events: none;
}

.narrative-skill-token__body {
  position: relative;
  z-index: 1;
  max-width: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  color: rgba(241, 245, 255, 0.96);
  text-shadow:
    0 1px 4px rgba(0, 0, 0, 0.92),
    0 0 12px rgba(167, 139, 250, 0.36);
}

.narrative-skill-token__body strong {
  max-width: 100%;
  overflow: hidden;
  color: #fff0bd;
  font-size: 17px;
  font-weight: 950;
  line-height: 1.12;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.narrative-damage-token {
  width: 92px;
  min-height: 78px;
  padding: 10px 9px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 4px;
  border-radius: 10px;
  border: 1px solid rgba(255, 170, 132, 0.58);
  background:
    radial-gradient(circle at 50% 18%, rgba(255, 214, 164, 0.2), transparent 38%),
    linear-gradient(145deg, rgba(89, 30, 28, 0.92), rgba(24, 14, 24, 0.94) 62%, rgba(62, 24, 40, 0.88));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.12),
    0 0 16px rgba(248, 113, 113, 0.34),
    0 12px 22px rgba(0, 0, 0, 0.44);
  color: rgba(255, 236, 225, 0.96);
  text-align: center;
}

.narrative-damage-token span {
  font-size: 10px;
  font-weight: 900;
  line-height: 1;
  color: rgba(254, 205, 211, 0.86);
}

.narrative-damage-token strong {
  font-size: 28px;
  font-weight: 950;
  line-height: 0.92;
  color: #ffbea8;
  text-shadow: 0 0 12px rgba(248, 113, 113, 0.66);
}

.narrative-damage-token small {
  max-width: 100%;
  overflow: hidden;
  font-size: 10px;
  line-height: 1.1;
  color: rgba(255, 232, 218, 0.84);
  text-overflow: ellipsis;
  white-space: nowrap;
}

.narrative-mist-layer {
  position: absolute;
  left: 18px;
  top: 14px;
  width: calc(100% - 36px);
  height: calc(100% - 68px);
  z-index: 9;
  overflow: visible;
  pointer-events: none;
}

.narrative-mist {
  --mist-core: rgba(255, 199, 113, 0.84);
  --mist-aura: rgba(255, 134, 82, 0.24);
  --mist-particle: rgba(255, 239, 190, 0.92);
  animation: narrativeMistIn 0.36s ease-out both;
}

.narrative-mist__aura,
.narrative-mist__flow {
  fill: none;
  stroke-linecap: round;
  stroke-linejoin: round;
  vector-effect: non-scaling-stroke;
}

.narrative-mist__aura {
  stroke: var(--mist-aura);
  stroke-width: 15px;
  filter: url(#narrative-mist-blur);
  opacity: 0.9;
}

.narrative-mist__flow {
  stroke: var(--mist-core);
  stroke-width: 3px;
  stroke-dasharray: 8 12;
  filter: drop-shadow(0 0 7px var(--mist-core));
  opacity: 0.88;
}

.narrative-mist-particle {
  fill: var(--mist-particle);
  opacity: 0.66;
  filter: drop-shadow(0 0 5px var(--mist-particle));
}

.narrative-mist-particle--bright {
  opacity: 0.9;
}

.narrative-mist__damage-label {
  paint-order: stroke;
  stroke: rgba(5, 12, 22, 0.9);
  stroke-width: 5px;
  fill: #ffbea8;
  font-size: 26px;
  font-weight: 900;
  letter-spacing: 0;
  filter: drop-shadow(0 0 10px rgba(248, 113, 113, 0.7));
  vector-effect: non-scaling-stroke;
  animation: narrativeDamagePopSvg 0.92s ease-out both;
}

.narrative-mist--respond {
  --mist-core: rgba(158, 226, 255, 0.9);
  --mist-aura: rgba(96, 165, 250, 0.28);
  --mist-particle: rgba(224, 247, 255, 0.94);
}

.narrative-mist--skill {
  --mist-core: rgba(226, 214, 255, 0.92);
  --mist-aura: rgba(168, 85, 247, 0.3);
  --mist-particle: rgba(255, 239, 190, 0.94);
}

.narrative-mist--damage {
  --mist-core: rgba(255, 159, 127, 0.94);
  --mist-aura: rgba(239, 68, 68, 0.34);
  --mist-particle: rgba(255, 215, 184, 0.96);
}

.narrative-settled-row {
  position: absolute;
  inset: 14px 18px 54px;
  z-index: 7;
  pointer-events: none;
}

.narrative-settled-card {
  position: absolute;
  top: var(--settled-card-y);
  width: clamp(46px, 5.8vw, 70px);
  height: clamp(58px, 7.5vw, 86px);
  overflow: hidden;
  border-radius: 8px;
  background: rgba(8, 17, 29, 0.86);
  border: 1px solid rgba(132, 172, 207, 0.32);
  box-shadow: 0 8px 16px rgba(2, 8, 16, 0.32);
  transform: translate(-50%, -50%);
  animation: narrativeSettledIn 0.28s ease-out both;
}

.narrative-settled-card--side-left {
  left: calc(var(--narrative-actor-card-edge) + var(--narrative-actor-card-width) / 2);
}

.narrative-settled-card--side-right {
  left: calc(100% - var(--narrative-actor-card-edge) - var(--narrative-actor-card-width) / 2);
}

.narrative-settled-card--blue {
  border-color: rgba(56, 189, 248, 0.54);
}

.narrative-settled-card--red {
  border-color: rgba(248, 113, 113, 0.54);
}

.narrative-settled-card img {
  width: 100%;
  height: 100%;
  border-radius: inherit;
  object-fit: cover;
  object-position: 50% 12%;
}

.narrative-settled-card span {
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  overflow: hidden;
  padding: 3px 4px;
  color: rgba(237, 247, 255, 0.94);
  font-size: 9px;
  line-height: 1;
  text-align: center;
  text-overflow: ellipsis;
  white-space: nowrap;
  background: rgba(4, 10, 18, 0.82);
}

.narrative-field-effect {
  position: absolute;
  right: 12px;
  bottom: 10px;
  z-index: 14;
  display: flex;
  align-items: center;
  gap: 8px;
  max-width: min(280px, calc(100% - 24px));
  padding: 8px 10px;
  border-radius: 12px;
  border: 1px solid rgba(251, 113, 133, 0.34);
  background: linear-gradient(135deg, rgba(47, 10, 20, 0.82), rgba(10, 16, 28, 0.78));
  box-shadow:
    inset 0 1px 0 rgba(255, 225, 230, 0.08),
    0 10px 24px rgba(10, 2, 8, 0.32),
    0 0 22px rgba(190, 18, 60, 0.18);
  animation: narrativeEventIn 0.28s ease-out both;
}

.narrative-field-effect--ambient {
  left: 50%;
  right: auto;
  top: 50%;
  bottom: auto;
  flex-direction: column;
  width: min(190px, 72%);
  padding: 12px 14px;
  text-align: center;
  transform: translate(-50%, -50%);
}

.narrative-field-effect__icon {
  width: 42px;
  height: 42px;
  flex: 0 0 auto;
  padding: 4px;
  border-radius: 10px;
  border: 1px solid rgba(251, 113, 133, 0.36);
  background: rgba(0, 0, 0, 0.42);
  box-shadow: 0 0 16px rgba(190, 18, 60, 0.28);
  animation: roseCourtyardPulse 3s ease-in-out infinite;
}

.narrative-field-effect--ambient .narrative-field-effect__icon {
  width: 62px;
  height: 62px;
}

.narrative-field-effect__body {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.narrative-field-effect__body strong {
  overflow: hidden;
  color: #fb7185;
  font-size: 13px;
  line-height: 1.15;
  text-overflow: ellipsis;
  white-space: nowrap;
  text-shadow: 0 1px 3px rgba(0, 0, 0, 0.82);
}

.narrative-field-effect__body span {
  color: #fecdd3;
  font-size: 10px;
  line-height: 1.25;
}

@keyframes narrativeActorIn {
  from { opacity: 0; transform: translateX(var(--actor-enter-x)) scale(0.88); }
  to { opacity: 1; transform: translateX(0) scale(1); }
}

@keyframes narrativeStackItemIn {
  from {
    opacity: 0;
    transform:
      translate(var(--stack-enter-x, 0px), calc(18px + var(--stack-order, 0) * -1px))
      scale(0.86);
  }
  to {
    opacity: 1;
    transform: translateY(calc(var(--stack-order, 0) * -1px)) scale(1);
  }
}

@keyframes narrativeMistIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

@keyframes narrativeSkillPulse {
  0%, 100% { opacity: 0.58; transform: scale(0.94); }
  50% { opacity: 1; transform: scale(1.04); }
}

@keyframes narrativeEventIn {
  from { opacity: 0; transform: translateY(-6px) scale(0.94); }
  to { opacity: 1; transform: translateY(0) scale(1); }
}

@keyframes narrativeSettledIn {
  from { opacity: 0; transform: translate(-50%, calc(-50% + 12px)) scale(0.88); }
  to { opacity: 1; transform: translate(-50%, -50%) scale(1); }
}

@keyframes narrativeDamagePop {
  0% { opacity: 0; transform: translate(-50%, -50%) scale(0.62); }
  24% { opacity: 1; transform: translate(-50%, -62%) scale(1.12); }
  100% { opacity: 0; transform: translate(-50%, -92%) scale(0.94); }
}

@keyframes narrativeDamagePopSvg {
  0% { opacity: 0; transform: translateY(6px) scale(0.68); }
  24% { opacity: 1; transform: translateY(-2px) scale(1.12); }
  100% { opacity: 0; transform: translateY(-24px) scale(0.96); }
}

.skill-plaque {
  position: absolute;
  pointer-events: none;
  color: #f8ecd1;
  isolation: isolate;
}

.skill-plaque::before {
  content: '';
  position: absolute;
  inset: 0;
  z-index: -1;
  background-image:
    linear-gradient(90deg, rgba(4, 10, 20, 0.28), rgba(4, 10, 20, 0.06), rgba(4, 10, 20, 0.28)),
    url('/assets/ui/skill-plaque-bg.png');
  background-size: cover;
  background-position: center;
  border-radius: inherit;
}

.skill-plaque__body {
  position: relative;
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
  overflow: hidden;
}

.skill-plaque__eyebrow {
  max-width: 100%;
  color: rgba(244, 219, 165, 0.86);
  font-size: 11px;
  font-weight: 800;
  line-height: 1.15;
  text-shadow: 0 1px 3px rgba(0, 0, 0, 0.85);
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.skill-plaque__title {
  max-width: 100%;
  color: #fff2c8;
  font-weight: 900;
  line-height: 1.05;
  text-shadow:
    0 2px 4px rgba(0, 0, 0, 0.88),
    0 0 18px rgba(235, 179, 88, 0.42);
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.skill-plaque__effect {
  max-width: 100%;
  color: rgba(235, 240, 246, 0.9);
  line-height: 1.25;
  text-shadow: 0 1px 3px rgba(0, 0, 0, 0.86);
  display: -webkit-box;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.skill-plaque--featured {
  left: 50%;
  top: 50%;
  width: min(520px, calc(100% - 52px));
  min-height: 126px;
  padding: 18px 46px;
  border-radius: 12px;
  transform: translate(-50%, -50%);
  z-index: 16;
  filter: drop-shadow(0 20px 34px rgba(0, 0, 0, 0.44));
}

.skill-plaque--featured::before {
  box-shadow:
    inset 0 0 0 1px rgba(251, 231, 176, 0.42),
    inset 0 0 36px rgba(255, 208, 116, 0.16),
    0 0 0 1px rgba(24, 12, 4, 0.42);
}

.skill-plaque--featured .skill-plaque__body {
  gap: 8px;
}

.skill-plaque--featured .skill-plaque__eyebrow {
  font-size: clamp(11px, 1.35vw, 13px);
}

.skill-plaque--featured .skill-plaque__title {
  font-size: clamp(25px, 4.2vw, 42px);
}

.skill-plaque--featured .skill-plaque__effect {
  width: min(390px, 100%);
  font-size: clamp(12px, 1.55vw, 15px);
  -webkit-line-clamp: 2;
}

.skill-plaque__glow {
  position: absolute;
  inset: -18px 16%;
  z-index: -2;
  border-radius: 999px;
  background: radial-gradient(ellipse at center, rgba(237, 181, 91, 0.3), rgba(237, 181, 91, 0));
  filter: blur(10px);
  animation: skillPlaqueGlow 1.1s ease-out both;
}

.skill-plaque-stack {
  position: absolute;
  right: 12px;
  top: 12px;
  z-index: 12;
  width: min(300px, calc(100% - 24px));
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 8px;
  pointer-events: none;
}

.skill-plaque--settled {
  position: relative;
  width: min(284px, 100%);
  min-height: 58px;
  padding: 8px 14px;
  border-radius: 8px;
  filter: drop-shadow(0 10px 16px rgba(0, 0, 0, 0.36));
}

.skill-plaque--settled::before {
  opacity: 0.92;
  box-shadow:
    inset 0 0 0 1px rgba(236, 211, 154, 0.28),
    inset 0 0 22px rgba(223, 168, 90, 0.08);
}

.skill-plaque--settled .skill-plaque__body {
  align-items: flex-start;
  text-align: left;
  gap: 2px;
}

.skill-plaque--settled .skill-plaque__eyebrow {
  max-width: 100%;
  font-size: 10px;
}

.skill-plaque--settled .skill-plaque__title {
  max-width: 100%;
  font-size: 15px;
}

.skill-plaque--settled .skill-plaque__effect {
  max-width: 100%;
  color: rgba(229, 236, 243, 0.78);
  font-size: 11px;
  -webkit-line-clamp: 1;
}

.skill-plaque-featured-enter-active,
.skill-plaque-featured-leave-active {
  transition:
    opacity 0.28s ease,
    transform 0.28s cubic-bezier(0.2, 0.8, 0.2, 1),
    filter 0.28s ease;
}

.skill-plaque-featured-enter-from,
.skill-plaque-featured-leave-to {
  opacity: 0;
  transform: translate(-50%, -50%) scale(0.86);
  filter: blur(2px) drop-shadow(0 10px 18px rgba(0, 0, 0, 0.3));
}

.skill-plaque-settled-enter-active,
.skill-plaque-settled-leave-active {
  transition:
    opacity 0.24s ease,
    transform 0.24s cubic-bezier(0.2, 0.8, 0.2, 1);
}

.skill-plaque-settled-enter-from,
.skill-plaque-settled-leave-to {
  opacity: 0;
  transform: translateX(16px) scale(0.96);
}

@keyframes skillPlaqueGlow {
  0% { opacity: 0; transform: scaleX(0.7); }
  30% { opacity: 1; transform: scaleX(1.05); }
  100% { opacity: 0.72; transform: scaleX(1); }
}

.battle-zone-shell > * {
  position: relative;
  z-index: 1;
}

.battle-zone-shell > .action-narrative-layer {
  position: absolute;
  inset: 0;
  z-index: 6;
}

.battle-zone-shell::before {
  content: '';
  position: absolute;
  width: min(320px, 80%);
  height: min(320px, 80%);
  left: 50%;
  top: 48%;
  transform: translate(-50%, -50%);
  border-radius: 999px;
  border: 0;
  box-shadow: none;
  background: transparent;
  pointer-events: none;
  z-index: 0;
  animation: battleRingBreath 6s ease-in-out infinite;
}

.battle-zone-shell::after {
  content: '';
  position: absolute;
  width: min(180px, 50%);
  height: min(180px, 50%);
  left: 50%;
  top: 48%;
  transform: translate(-50%, -50%) rotate(18deg);
  border: 0;
  border-radius: 0;
  background: transparent;
  pointer-events: none;
  z-index: 0;
  animation: battleSquareBreath 8s ease-in-out infinite reverse;
}

@keyframes battleRingBreath {
  0%, 100% { opacity: 0.7; transform: translate(-50%, -50%) scale(1); }
  50% { opacity: 1; transform: translate(-50%, -50%) scale(1.03); }
}

@keyframes battleSquareBreath {
  0%, 100% { opacity: 0.6; transform: translate(-50%, -50%) rotate(18deg) scale(1); }
  50% { opacity: 0.9; transform: translate(-50%, -50%) rotate(22deg) scale(1.04); }
}

@media (min-width: 1600px) {
  .battle-zone-shell {
    padding: 12px 12px 10px;
  }

}

@media (min-width: 2000px) {
  .battle-zone-shell {
    padding: 14px 14px 12px;
  }

}

@keyframes roseCourtyardPulse {
  0%, 100% {
    box-shadow: 0 4px 16px rgba(190, 18, 60, 0.35);
  }
  50% {
    box-shadow: 0 6px 22px rgba(190, 18, 60, 0.55);
  }
}

@media (max-width: 640px) {
  .battle-zone-shell {
    padding: 8px 8px 6px;
  }

}
</style>
