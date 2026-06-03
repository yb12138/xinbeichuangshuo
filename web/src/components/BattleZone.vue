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
import type { ActionFlowDTO, ActionFlowEdgeDTO, ActionFlowNodeDTO } from '../network/protocol'
import { useSnapshotStore } from '../stores/snapshot.store'
import type { Card, PlayerView } from '../types/game'
import CardComponent from './CardComponent.vue'
import RoseCourtyardIcon from './StatusIcons/RoseCourtyardIcon.vue'

const props = defineProps<{
  narrativeSuspended?: boolean
  actorSeatPositions?: Record<string, number>
}>()

const battleFxStore = useBattleFxStore()
const snapshotStore = useSnapshotStore()
const { actionNarrative, latestActionFlow, narrativePlayback, skillAnnouncements } = storeToRefs(battleFxStore)
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

const hasNarrativeLayer = computed(() => !!latestActionFlow.value || !!actionNarrative.value || hasRoseCourtyard.value)

const narrativeCards = computed(() => actionNarrative.value?.playedCards || [])
const narrativeEvents = computed(() => actionNarrative.value?.events || [])
const narrativeLinks = computed(() => actionNarrative.value?.links || [])

function normalizedActionType(actionType?: string) {
  return String(actionType || '').trim().toLowerCase()
}

function isCombatLoopActionType(actionType?: string) {
  return ['attack', 'counter', 'defend', 'shield'].includes(normalizedActionType(actionType))
}

const narrativeLoopCardChain = computed(() => {
  const hasNonCombatSkillEvent = narrativeEvents.value.some((event) => {
    return event.kind === 'skill' && !String(event.label || '').includes('未命中')
  })
  if (hasNonCombatSkillEvent) return null

  const cards = narrativeCards.value
    .filter(card => isCombatLoopActionType(card.actionType) && !!card.targetId)
    .sort((a, b) => a.createdAt - b.createdAt || a.id - b.id)
  if (cards.length < 1) return null

  const actorIds: string[] = []
  const pushActor = (id?: string) => {
    if (id && !actorIds.includes(id)) actorIds.push(id)
  }

  pushActor(cards[0]?.playerId)
  for (const card of cards) {
    pushActor(card.playerId)
    pushActor(card.targetId)
  }

  if (actorIds.length < 2) return null
  return {
    actorIds,
    cards,
  }
})

const hasNarrativeCombatLoop = computed(() => !!latestActionFlow.value || !!narrativeLoopCardChain.value)

const primaryCounterChainCards = computed(() => {
  const attackCard = narrativeCards.value.find((card) => {
    return normalizedActionType(card.actionType) === 'attack' && !!card.targetId
  })
  if (!attackCard?.targetId) return null

  const counterCard = narrativeCards.value.find((card) => {
    return normalizedActionType(card.actionType) === 'counter' &&
      card.playerId === attackCard.targetId
  })
  if (!counterCard) return null

  const missEvent = narrativeEvents.value.find((event) => {
    return event.actorId === attackCard.playerId &&
      event.targetId === attackCard.targetId &&
      String(event.label || '').includes('未命中')
  })

  return {
    sourceId: attackCard.playerId,
    targetId: attackCard.targetId,
    attackCard,
    counterCard,
    missEvent,
  }
})

const primaryCounterChainActorIds = computed(() => {
  const chain = primaryCounterChainCards.value
  return new Set(chain ? [chain.sourceId, chain.targetId] : [])
})

function isPrimaryCounterChainActor(playerId?: string) {
  return !!playerId && primaryCounterChainActorIds.value.has(playerId)
}

const featuredPlayer = computed(() => {
  if (hasNarrativeCombatLoop.value) return undefined
  const id = actionNarrative.value?.featuredActorId
  if (isPrimaryCounterChainActor(id)) return undefined
  return id ? players.value[id] : undefined
})

const opposedPlayers = computed(() =>
  (actionNarrative.value?.opposedActorIds || [])
    .filter(() => !hasNarrativeCombatLoop.value)
    .filter(id => !isPrimaryCounterChainActor(id))
    .map(id => players.value[id])
    .filter(isPlayerView)
)

const settledPlayers = computed(() =>
  (actionNarrative.value?.settledActorIds || [])
    .filter(() => !hasNarrativeCombatLoop.value)
    .filter(id => id !== actionNarrative.value?.featuredActorId && !(actionNarrative.value?.opposedActorIds || []).includes(id))
    .filter(id => !isPrimaryCounterChainActor(id))
    .map(id => players.value[id])
    .filter(isPlayerView)
)

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

interface NarrativeLoopActor {
  id: string
  index: number
  x: number
  y: number
}

interface NarrativeLoopCard {
  id: string
  eventId?: number
  card: Card
}

interface NarrativeLoopAction {
  id: string
  index: number
  sourceId: string
  targetId: string
  actionKind: NarrativeLinkKind
  actionType: string
  tone: 'red' | 'gold'
  path: string
  startX: number
  startY: number
  controlX: number
  controlY: number
  endX: number
  endY: number
  cardX: number
  cardY: number
  labelX: number
  labelY: number
  normalX: number
  normalY: number
  cards: NarrativeLoopCard[]
  cardView?: ActionNarrativeCardView
  item?: NarrativeStackItem
  damage?: number
  damageType?: string
  missed?: boolean
  outcome?: string
  label?: string
  nodeIds: string[]
}

interface NarrativeLoopNode {
  id: string
  kind: string
  actorId?: string
  targetIds: string[]
  anchorEdgeId?: string
  x: number
  y: number
  title: string
  detail?: string
  outcome?: string
  damage?: number
  cards: NarrativeLoopCard[]
  width?: number
  height?: number
  rect?: LoopLayoutRect
  zIndex?: number
}

interface LoopLayoutSize {
  width: number
  height: number
}

interface LoopLayoutRect {
  x: number
  y: number
  width: number
  height: number
}

type NarrativeLoopPacketMode = 'full' | 'compact' | 'marker'
type NarrativeLoopNotePlacement = 'bottom' | 'side' | 'badge'

interface NarrativeLoopPacket {
  id: string
  action: NarrativeLoopAction
  cards: NarrativeLoopCard[]
  visibleCards: NarrativeLoopCard[]
  hiddenCardCount: number
  mode: NarrativeLoopPacketMode
  notePlacement: NarrativeLoopNotePlacement
  x: number
  y: number
  rect: LoopLayoutRect
  cardWidth: number
  cardScale: number
  noteTitle: string
  noteDetail: string
  noteResult: string
  zIndex: number
}

interface NarrativeLoopPacketPlacement {
  x: number
  y: number
  rect: LoopLayoutRect
  score: number
  mode: NarrativeLoopPacketMode
  notePlacement: NarrativeLoopNotePlacement
}

type NarrativeLoopNodeCandidateTier = 'line' | 'primary' | 'extended' | 'fallback'

interface NarrativeLoopNodePlacement {
  x: number
  y: number
  rect: LoopLayoutRect
  width: number
  height: number
  zIndex: number
}

interface NarrativeCombatLoopView {
  actors: NarrativeLoopActor[]
  actions: NarrativeLoopAction[]
  nodes: NarrativeLoopNode[]
  source: 'backend' | 'fallback'
}

interface NarrativeCombatLoopLayout extends NarrativeCombatLoopView {
  packets: NarrativeLoopPacket[]
}

const narrativeLayerRef = ref<HTMLElement | null>(null)
const narrativeLoopFieldRef = ref<HTMLElement | null>(null)
const narrativeMistLayerRef = ref<SVGSVGElement | null>(null)
const measuredNarrativeMistSegments = ref<NarrativeMistSegment[]>([])
let narrativeMistResizeObserver: ResizeObserver | null = null
let narrativeLoopFieldResizeObserver: ResizeObserver | null = null
let narrativeMistMeasureFrame = 0
let narrativeGsapContext: gsap.Context | null = null
let narrativeGsapTimeline: gsap.core.Timeline | null = null
let lastAnimatedNarrativeStepId = ''
const narrativeLoopFieldSize = ref<LoopLayoutSize>({ width: 900, height: 360 })

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
  const flowDamageEdge = [...(latestActionFlow.value?.edges || [])]
    .reverse()
    .find(edge => edge.to_user_id === playerId && (edge.damage || 0) > 0)
  if (flowDamageEdge?.damage) {
    return {
      id: flowDamageEdge.id,
      kind: 'damage' as const,
      label: `造成 ${flowDamageEdge.damage} 点伤害`,
      actorId: flowDamageEdge.from_user_id,
      targetId: playerId,
      damage: flowDamageEdge.damage,
      createdAt: flowDamageEdge.order,
    }
  }
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

function narrativeEndpointSideForPlayer(playerId?: string, fallback: NarrativeActorRole = 'featured') {
  const chain = primaryCounterChainCards.value
  if (chain && playerId === chain.sourceId) return 'left'
  if (chain && playerId === chain.targetId) return 'right'
  return narrativeSideForPlayer(playerId, fallback)
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

const primaryCounterChain = computed(() => {
  const chain = primaryCounterChainCards.value
  if (!chain) return null

  const attackItem = narrativeStackItems.value.find((item) => {
    return item.kind === 'card' && item.cardView?.id === chain.attackCard.id
  })
  const counterItem = narrativeStackItems.value.find((item) => {
    return item.kind === 'card' && item.cardView?.id === chain.counterCard.id
  })
  const missItem = chain.missEvent
    ? narrativeStackItems.value.find((item) => item.kind === 'skill' && item.eventView?.id === chain.missEvent?.id)
    : undefined

  if (!attackItem) return null
  return {
    ...chain,
    attackItem,
    counterItem,
    missItem,
  }
})

function loopAngleForIndex(index: number, total: number) {
  return -90 + (360 / Math.max(total, 1)) * index
}

function loopPointForAngle(angle: number, scale = 1) {
  const radians = angle * Math.PI / 180
  return {
    x: 50 + Math.cos(radians) * 31 * scale,
    y: 52 + Math.sin(radians) * 32 * scale,
  }
}

function loopPointForActor(index: number, total: number) {
  if (total <= 1) {
    return { x: 50, y: 52 }
  }
  return loopPointForAngle(loopAngleForIndex(index, total))
}

function loopClockwiseDiff(fromAngle: number, toAngle: number) {
  let diff = toAngle - fromAngle
  while (diff <= 0) diff += 360
  return diff
}

function clampLoopCoord(value: number, min: number, max: number) {
  return Math.max(min, Math.min(max, value))
}

function loopQuadraticPoint(
  start: { x: number; y: number },
  control: { x: number; y: number },
  end: { x: number; y: number },
  t: number,
) {
  const inverse = 1 - t
  return {
    x: inverse * inverse * start.x + 2 * inverse * t * control.x + t * t * end.x,
    y: inverse * inverse * start.y + 2 * inverse * t * control.y + t * t * end.y,
  }
}

function loopNormalForCurve(start: { x: number; y: number }, end: { x: number; y: number }) {
  const dx = end.x - start.x
  const dy = end.y - start.y
  const length = Math.hypot(dx, dy) || 1
  return {
    x: -dy / length,
    y: dx / length,
  }
}

function loopActionGeometry(sourceIndex: number, targetIndex: number, total: number) {
  const sourceAngle = loopAngleForIndex(sourceIndex, total)
  const targetAngle = loopAngleForIndex(targetIndex, total)
  const diff = loopClockwiseDiff(sourceAngle, targetAngle)
  const midAngle = sourceAngle + diff / 2
  const start = loopPointForAngle(sourceAngle, 0.9)
  const end = loopPointForAngle(targetAngle, 0.9)
  const control = loopPointForAngle(midAngle, 1.22)
  const card = loopQuadraticPoint(start, control, end, 0.52)
  const label = card
  const normal = loopNormalForCurve(start, end)
  return {
    path: `M ${start.x.toFixed(2)} ${start.y.toFixed(2)} Q ${control.x.toFixed(2)} ${control.y.toFixed(2)} ${end.x.toFixed(2)} ${end.y.toFixed(2)}`,
    startX: start.x,
    startY: start.y,
    controlX: control.x,
    controlY: control.y,
    endX: end.x,
    endY: end.y,
    cardX: clampLoopCoord(card.x, 8, 92),
    cardY: clampLoopCoord(card.y, 12, 88),
    labelX: clampLoopCoord(label.x, 8, 92),
    labelY: clampLoopCoord(label.y, 12, 88),
    normalX: normal.x,
    normalY: normal.y,
  }
}

function damageForLoopAction(card: ActionNarrativeCardView) {
  const damageEvent = narrativeEvents.value
    .filter(event =>
      event.kind === 'damage' &&
      event.actorId === card.playerId &&
      event.targetId === card.targetId &&
      event.createdAt >= card.createdAt
    )
    .sort((a, b) => a.createdAt - b.createdAt)[0]
  return damageEvent?.damage ?? card.card.damage
}

function missForLoopAction(card: ActionNarrativeCardView) {
  return narrativeEvents.value.some((event) => {
    return event.actorId === card.playerId &&
      event.targetId === card.targetId &&
      event.createdAt >= card.createdAt &&
      String(event.label || '').includes('未命中')
  })
}

function pushUniqueActor(actorIds: string[], id?: string) {
  const normalized = String(id || '').trim()
  if (normalized && !actorIds.includes(normalized)) actorIds.push(normalized)
}

function flowActionKind(phase?: string): NarrativeLinkKind {
  const normalized = normalizedActionType(phase)
  if (normalized === 'counter' || normalized === 'defend' || normalized === 'shield') return 'respond'
  if (normalized === 'magic' || normalized === 'skill' || normalized === 'effect') return 'skill'
  if (normalized === 'take') return 'damage'
  return 'attack'
}

function loopCardsFromFlow(cards: Card[] | undefined, prefix: string, eventId?: number): NarrativeLoopCard[] {
  return (cards || []).map((card, index) => ({
    id: `${prefix}-card-${eventId || 0}-${card.id || index}`,
    eventId,
    card,
  }))
}

function loopCardIdentity(loopCard: NarrativeLoopCard) {
  if (loopCard.eventId) return `event:${loopCard.eventId}:${loopCard.card.id || loopCard.card.name || loopCard.id}`
  const cardKey = String(loopCard.card.id || loopCard.card.name || '').trim()
  if (cardKey) return `card:${cardKey}`
  return `card-instance:${loopCard.id}`
}

function appendUniqueLoopCards(target: NarrativeLoopCard[], cards: NarrativeLoopCard[]) {
  const seen = new Set(target.map(loopCardIdentity))
  for (const card of cards) {
    const key = loopCardIdentity(card)
    if (seen.has(key)) continue
    seen.add(key)
    target.push(card)
  }
}

function edgeIdForFlowCardNode(node: ActionFlowNodeDTO, edges: ActionFlowEdgeDTO[]) {
  if (node.anchor_edge_id) return node.anchor_edge_id

  const nodeEdge = edges.find(edge => (edge.node_ids || []).includes(node.id))
  if (nodeEdge) return nodeEdge.id

  if (node.event_id) {
    const eventEdge = edges.find(edge => edge.card_event_id === node.event_id)
    if (eventEdge) return eventEdge.id
  }

  const targets = node.target_user_ids || []
  const candidates = edges.filter(edge => {
    if (edge.from_user_id !== node.actor_user_id) return false
    return targets.length < 1 || targets.includes(edge.to_user_id)
  })
  if (candidates.length === 1) return candidates[0]?.id || ''

  const emptyCandidate = candidates.find(edge => (edge.cards || []).length < 1)
  if (emptyCandidate) return emptyCandidate.id

  if (candidates.length > 1) {
    const sortedByOrder = [...candidates].sort((a, b) => {
      const aDistance = Math.abs((a.order || 0) - (node.order || 0))
      const bDistance = Math.abs((b.order || 0) - (node.order || 0))
      return aDistance - bDistance || a.order - b.order
    })
    return sortedByOrder[0]?.id || ''
  }

  return ''
}

function cardsForFlowEdge(edge: ActionFlowEdgeDTO, nodes: ActionFlowNodeDTO[], edges: ActionFlowEdgeDTO[]) {
  const cards = loopCardsFromFlow(edge.cards, edge.id, edge.card_event_id)
  const cardNodes = nodes.filter(node => normalizedActionType(node.kind) === 'card')
  for (const node of cardNodes) {
    if (edgeIdForFlowCardNode(node, edges) !== edge.id) continue
    appendUniqueLoopCards(cards, loopCardsFromFlow(node.cards, node.id, node.event_id))
  }
  return cards
}

function edgeHasVisualPayload(
  edge: ActionFlowEdgeDTO,
  cards: NarrativeLoopCard[],
  renderableNodeIds: Set<string>,
) {
  const outcome = normalizedActionType(edge.outcome)
  const hasVisualOutcome = outcome !== '' && outcome !== 'pending' && outcome !== 'resolved'
  return cards.length > 0 ||
    (edge.damage || 0) > 0 ||
    String(edge.damage_type || '').trim() !== '' ||
    String(edge.label || '').trim() !== '' ||
    hasVisualOutcome ||
    (edge.node_ids || []).some(nodeId => renderableNodeIds.has(nodeId))
}

function shouldRenderFlowEdge(
  edge: ActionFlowEdgeDTO,
  cards: NarrativeLoopCard[],
  edges: ActionFlowEdgeDTO[],
  cardsByEdgeId: Map<string, NarrativeLoopCard[]>,
  renderableNodeIds: Set<string>,
) {
  const phase = normalizedActionType(edge.phase)
  if (phase !== 'attack') return true
  if (edgeHasVisualPayload(edge, cards, renderableNodeIds)) return true

  return !edges.some((candidate) => {
    if (candidate.id === edge.id) return false
    if (candidate.from_user_id !== edge.from_user_id || candidate.to_user_id !== edge.to_user_id) return false
    const candidatePhase = normalizedActionType(candidate.phase)
    if (!['counter', 'defend', 'shield'].includes(candidatePhase)) return false
    return edgeHasVisualPayload(candidate, cardsByEdgeId.get(candidate.id) || [], renderableNodeIds)
  })
}

function isResponseFlowPhase(phase: string) {
  return ['counter', 'defend', 'shield'].includes(normalizedActionType(phase))
}

function isEmptyAttackResolutionEdge(
  edge: ActionFlowEdgeDTO,
  cardsByEdgeId: Map<string, NarrativeLoopCard[]>,
  renderableNodeIds: Set<string>,
) {
  if (normalizedActionType(edge.phase) !== 'attack') return false
  if ((cardsByEdgeId.get(edge.id) || []).length > 0) return false
  if ((edge.node_ids || []).some(nodeId => renderableNodeIds.has(nodeId))) return false
  const outcome = normalizedActionType(edge.outcome)
  return outcome !== '' && outcome !== 'pending'
}

function mergeResponseResolutionEdges(
  edges: ActionFlowEdgeDTO[],
  cardsByEdgeId: Map<string, NarrativeLoopCard[]>,
  renderableNodeIds: Set<string>,
) {
  const mergedEdges = edges.map(edge => ({
    ...edge,
    cards: [...(edge.cards || [])],
    node_ids: [...(edge.node_ids || [])],
  }))
  const hiddenEdgeIds = new Set<string>()

  for (const edge of mergedEdges) {
    if (!isEmptyAttackResolutionEdge(edge, cardsByEdgeId, renderableNodeIds)) continue
    const candidates = mergedEdges.filter(candidate =>
      candidate.id !== edge.id &&
      !hiddenEdgeIds.has(candidate.id) &&
      candidate.from_user_id === edge.from_user_id &&
      candidate.to_user_id === edge.to_user_id &&
      isResponseFlowPhase(candidate.phase) &&
      edgeHasVisualPayload(candidate, cardsByEdgeId.get(candidate.id) || [], renderableNodeIds)
    )
    if (candidates.length < 1) continue
    candidates.sort((a, b) => Math.abs((a.order || 0) - (edge.order || 0)) - Math.abs((b.order || 0) - (edge.order || 0)))
    const target = candidates[0]
    if (!target) continue
    if (edge.outcome && normalizedActionType(edge.outcome) !== 'pending') {
      target.outcome = edge.outcome
    }
    if (edge.label && !target.label) {
      target.label = edge.label
    }
    if ((edge.damage || 0) > 0 && !(target.damage || 0)) {
      target.damage = edge.damage
    }
    if (edge.damage_type && !target.damage_type) {
      target.damage_type = edge.damage_type
    }
    hiddenEdgeIds.add(edge.id)
  }

  return mergedEdges.filter(edge => !hiddenEdgeIds.has(edge.id))
}

function offsetLoopActionPoint(action: NarrativeLoopAction, amount: number): NarrativeLoopAction {
  if (amount === 0) return action
  const cardX = clampLoopCoord(action.cardX + action.normalX * amount, 8, 92)
  const cardY = clampLoopCoord(action.cardY + action.normalY * amount, 12, 88)
  return {
    ...action,
    cardX,
    cardY,
    labelX: cardX,
    labelY: cardY,
  }
}

function offsetLoopActionCurve(action: NarrativeLoopAction, amount: number): NarrativeLoopAction {
  if (amount === 0) return action
  const controlX = clampLoopCoord(action.controlX + action.normalX * amount, 4, 96)
  const controlY = clampLoopCoord(action.controlY + action.normalY * amount, 6, 94)
  const cardX = clampLoopCoord(action.cardX + action.normalX * amount * 0.72, 8, 92)
  const cardY = clampLoopCoord(action.cardY + action.normalY * amount * 0.72, 12, 88)
  return {
    ...action,
    controlX,
    controlY,
    cardX,
    cardY,
    labelX: cardX,
    labelY: cardY,
    path: `M ${action.startX.toFixed(2)} ${action.startY.toFixed(2)} Q ${controlX.toFixed(2)} ${controlY.toFixed(2)} ${action.endX.toFixed(2)} ${action.endY.toFixed(2)}`,
  }
}

function withLoopActionLayoutLanes(actions: NarrativeLoopAction[]) {
  const laneGroups = new Map<string, NarrativeLoopAction[]>()
  for (const action of actions) {
    const key = `${action.sourceId}->${action.targetId}`
    laneGroups.set(key, [...(laneGroups.get(key) || []), action])
  }
  const laneOffsets = [0, -5, 5, -9, 9, -13, 13]
  const laneOffsetById = new Map<string, number>()
  for (const group of laneGroups.values()) {
    if (group.length < 2) continue
    group.forEach((action, index) => {
      const cycle = Math.floor(index / laneOffsets.length)
      const base = laneOffsets[index % laneOffsets.length] || 0
      laneOffsetById.set(action.id, base + Math.sign(base) * cycle * 4)
    })
  }
  const routedActions = laneOffsetById.size
    ? actions.map(action => offsetLoopActionCurve(action, laneOffsetById.get(action.id) || 0))
    : actions

  const groups = new Map<string, NarrativeLoopAction[]>()
  for (const action of routedActions) {
    if (action.cards.length < 1) continue
    const key = `${action.cardX.toFixed(1)}:${action.cardY.toFixed(1)}`
    groups.set(key, [...(groups.get(key) || []), action])
  }

  const offsetByActionId = new Map<string, number>()
  const offsets = [0, -5.5, 5.5, -9.5, 9.5]
  for (const group of groups.values()) {
    if (group.length < 2) continue
    group.forEach((action, index) => {
      const cycle = Math.floor(index / offsets.length)
      const base = offsets[index % offsets.length] || 0
      offsetByActionId.set(action.id, base + Math.sign(base) * cycle * 4)
    })
  }

  if (offsetByActionId.size < 1) return routedActions
  return routedActions.map(action => offsetLoopActionPoint(action, offsetByActionId.get(action.id) || 0))
}

function loopNodeTitle(node: ActionFlowNodeDTO) {
  return String(node.skill_name || node.label || node.cards?.[0]?.name || node.kind || '结算').trim()
}

function loopNodeDetail(node: ActionFlowNodeDTO) {
  const effectText = String(node.effect_text || '').trim()
  if (normalizedActionType(node.kind) === 'skill') {
    return /^额外\+\d+/.test(effectText) && effectText.includes('行动') ? effectText : ''
  }
  if (node.outcome === 'miss') return '未命中'
  if (node.damage) return `伤害 ${node.damage}`
  return effectText
}

function flowNodeDedupeKey(node: ActionFlowNodeDTO) {
  const kind = normalizedActionType(node.kind)
  if (kind !== 'skill') return `${node.id || node.event_id || node.order}`
  const targets = [...(node.target_user_ids || [])].sort().join(',')
  return [
    kind,
    node.actor_user_id || '',
    node.skill_id || node.skill_name || node.label || '',
    node.anchor_edge_id || '',
    targets,
  ].join('|')
}

function uniqueRenderableFlowNodes(nodes: ActionFlowNodeDTO[]) {
  const seen = new Set<string>()
  return nodes.filter((node) => {
    if (!shouldRenderFlowNode(node)) return false
    const key = flowNodeDedupeKey(node)
    if (seen.has(key)) return false
    seen.add(key)
    return true
  })
}

function nodePointForFlowNode(
  node: ActionFlowNodeDTO,
  actionById: Map<string, NarrativeLoopAction>,
  actorIndexById: Map<string, number>,
  total: number,
) {
  if (node.anchor_edge_id) {
    const action = actionById.get(node.anchor_edge_id)
    if (action) {
      const offset = ((node.order || 1) % 3) - 1
      return {
        x: Math.max(8, Math.min(92, action.labelX + offset * 7)),
        y: Math.max(12, Math.min(88, action.labelY + 10)),
      }
    }
  }
  const actorIndex = actorIndexById.get(node.actor_user_id || '')
  if (actorIndex !== undefined) {
    return loopPointForAngle(loopAngleForIndex(actorIndex, total), 0.64)
  }
  return { x: 50, y: 52 }
}

function buildBackendNarrativeLoop(flow: ActionFlowDTO): NarrativeCombatLoopView | null {
  const actorIds: string[] = []
  const sortedActors = [...(flow.actors || [])].sort((a, b) => a.order - b.order)
  for (const actor of sortedActors) pushUniqueActor(actorIds, actor.player_id)
  pushUniqueActor(actorIds, flow.actor_user_id)
  for (const edge of flow.edges || []) {
    pushUniqueActor(actorIds, edge.from_user_id)
    pushUniqueActor(actorIds, edge.to_user_id)
  }
  for (const node of flow.nodes || []) {
    pushUniqueActor(actorIds, node.actor_user_id)
    for (const targetId of node.target_user_ids || []) pushUniqueActor(actorIds, targetId)
  }
  if (!actorIds.length) return null

  const total = actorIds.length
  const actorIndexById = new Map(actorIds.map((id, index) => [id, index]))
  const actors = actorIds.map((id, index) => {
    const point = loopPointForActor(index, total)
    return {
      id,
      index,
      x: point.x,
      y: point.y,
    }
  })

  const sortedEdges = [...(flow.edges || [])].sort((a, b) => a.order - b.order)
  const flowNodes = flow.nodes || []
  const renderableNodeIds = new Set(flowNodes.filter(shouldRenderFlowNode).map(node => node.id))
  const cardsByEdgeId = new Map(sortedEdges.map(edge => [
    edge.id,
    cardsForFlowEdge(edge, flowNodes, sortedEdges),
  ]))
  const normalizedEdges = mergeResponseResolutionEdges(sortedEdges, cardsByEdgeId, renderableNodeIds)
  const renderableEdges = normalizedEdges.filter(edge => shouldRenderFlowEdge(
    edge,
    cardsByEdgeId.get(edge.id) || [],
    normalizedEdges,
    cardsByEdgeId,
    renderableNodeIds,
  ))
  const actions = withLoopActionLayoutLanes(renderableEdges
    .map((edge: ActionFlowEdgeDTO, index): NarrativeLoopAction | null => {
      const sourceIndex = actorIndexById.get(edge.from_user_id)
      const targetIndex = actorIndexById.get(edge.to_user_id)
      if (sourceIndex === undefined || targetIndex === undefined) return null
      const actionType = normalizedActionType(edge.phase)
      const missed = edge.outcome === 'miss' || String(edge.label || '').includes('未命中')
      return {
        id: edge.id,
        index,
        sourceId: edge.from_user_id,
        targetId: edge.to_user_id,
        actionKind: flowActionKind(actionType),
        actionType,
        tone: index === 0 && actionType === 'attack' ? 'red' : 'gold',
        cards: cardsByEdgeId.get(edge.id) || [],
        damage: missed ? undefined : edge.damage,
        damageType: edge.damage_type,
        missed,
        outcome: edge.outcome,
        label: edge.label,
        nodeIds: edge.node_ids || [],
        ...loopActionGeometry(sourceIndex, targetIndex, total),
      }
    })
    .filter((action): action is NarrativeLoopAction => !!action))

  const actionById = new Map(actions.map(action => [action.id, action]))
  const nodes = uniqueRenderableFlowNodes(flowNodes)
    .sort((a, b) => a.order - b.order)
    .map((node): NarrativeLoopNode => {
      const point = nodePointForFlowNode(node, actionById, actorIndexById, total)
      return {
        id: node.id,
        kind: node.kind,
        actorId: node.actor_user_id,
        targetIds: node.target_user_ids || [],
        anchorEdgeId: node.anchor_edge_id,
        x: point.x,
        y: point.y,
        title: loopNodeTitle(node),
        detail: loopNodeDetail(node),
        outcome: node.outcome,
        damage: node.damage,
        cards: loopCardsFromFlow(node.cards, node.id, node.event_id),
      }
    })

  if (actions.length < 1 && nodes.length < 1) return null
  return {
    actors,
    actions,
    nodes,
    source: 'backend',
  }
}

function shouldRenderFlowNode(node: ActionFlowNodeDTO): boolean {
  const kind = normalizedActionType(node.kind)
  if (kind === 'card') return false
  if (kind === 'damage' || kind === 'resolution') return false
  if (kind === 'effect' && node.anchor_edge_id) return false
  return true
}

function buildFallbackNarrativeLoop(): NarrativeCombatLoopView | null {
  const chain = narrativeLoopCardChain.value
  if (!chain) return null

  const actorIds = chain.actorIds
  const total = actorIds.length
  const actorIndexById = new Map(actorIds.map((id, index) => [id, index]))
  const actors = actorIds.map((id, index) => {
    const point = loopPointForActor(index, total)
    return {
      id,
      index,
      x: point.x,
      y: point.y,
    }
  })

  const actions = withLoopActionLayoutLanes(chain.cards
    .map((cardView, index): NarrativeLoopAction | null => {
      const sourceIndex = actorIndexById.get(cardView.playerId)
      const targetId = cardView.targetId || ''
      const targetIndex = actorIndexById.get(targetId)
      if (sourceIndex === undefined || targetIndex === undefined) return null
      const geometry = loopActionGeometry(sourceIndex, targetIndex, total)
      const actionType = normalizedActionType(cardView.actionType)
      const item = narrativeStackItems.value.find(candidate =>
        candidate.kind === 'card' && candidate.cardView?.id === cardView.id
      )
      const missed = missForLoopAction(cardView)
      return {
        id: `loop-action-${cardView.id}`,
        index,
        sourceId: cardView.playerId,
        targetId,
        actionKind: linkKindForActionType(actionType),
        actionType,
        tone: index === 0 && actionType === 'attack' ? 'red' : 'gold',
        cards: [{
          id: `legacy-card-${cardView.id}`,
          eventId: cardView.timelineEventId,
          card: cardView.card,
        }],
        cardView,
        item,
        damage: missed ? undefined : damageForLoopAction(cardView),
        missed,
        outcome: missed ? 'miss' : undefined,
        nodeIds: [],
        ...geometry,
      }
    })
    .filter((action): action is NarrativeLoopAction => !!action))

  if (actors.length < 2 || actions.length < 1) return null
  return {
    actors,
    actions,
    nodes: [],
    source: 'fallback',
  }
}

const narrativeCombatLoop = computed<NarrativeCombatLoopView | null>(() => {
  const backendLoop = latestActionFlow.value ? buildBackendNarrativeLoop(latestActionFlow.value) : null
  return backendLoop || buildFallbackNarrativeLoop()
})

function percentToPx(point: { x: number; y: number }, size: LoopLayoutSize) {
  return {
    x: point.x / 100 * size.width,
    y: point.y / 100 * size.height,
  }
}

function pxToPercent(point: { x: number; y: number }, size: LoopLayoutSize) {
  return {
    x: size.width > 0 ? point.x / size.width * 100 : 50,
    y: size.height > 0 ? point.y / size.height * 100 : 50,
  }
}

function rectFromCenter(x: number, y: number, width: number, height: number): LoopLayoutRect {
  return {
    x: x - width / 2,
    y: y - height / 2,
    width,
    height,
  }
}

function inflateRect(rect: LoopLayoutRect, amount: number): LoopLayoutRect {
  return {
    x: rect.x - amount,
    y: rect.y - amount,
    width: rect.width + amount * 2,
    height: rect.height + amount * 2,
  }
}

function rectIntersects(a: LoopLayoutRect, b: LoopLayoutRect) {
  return a.x < b.x + b.width &&
    a.x + a.width > b.x &&
    a.y < b.y + b.height &&
    a.y + a.height > b.y
}

function rectHalfExtentAlongVector(rect: LoopLayoutRect, x: number, y: number) {
  return Math.abs(x) * rect.width / 2 + Math.abs(y) * rect.height / 2
}

function rectCollisionCount(rect: LoopLayoutRect, obstacles: LoopLayoutRect[], padding = 4) {
  const inflated = inflateRect(rect, padding)
  return obstacles.filter(obstacle => rectIntersects(inflated, obstacle)).length
}

function rectOutOfBoundsPenalty(rect: LoopLayoutRect, size: LoopLayoutSize) {
  const left = Math.max(0, -rect.x)
  const top = Math.max(0, -rect.y)
  const right = Math.max(0, rect.x + rect.width - size.width)
  const bottom = Math.max(0, rect.y + rect.height - size.height)
  return left + top + right + bottom
}

function clampRectCenter(x: number, y: number, width: number, height: number, size: LoopLayoutSize) {
  return {
    x: Math.max(width / 2 + 4, Math.min(size.width - width / 2 - 4, x)),
    y: Math.max(height / 2 + 4, Math.min(size.height - height / 2 - 4, y)),
  }
}

function loopLayoutProfile(size: LoopLayoutSize) {
  const isSmall = size.width < 640 || size.height < 300
  const isMedium = !isSmall && (size.width < 900 || size.height < 360)
  if (isSmall) {
    return {
      mode: 'small' as const,
      cardWidth: 52,
      cardScale: 52 / 74,
      noteWidth: 102,
      noteHeight: 22,
      compactNoteHeight: 22,
      actorWidth: 66,
      actorHeight: 92,
      gap: 5,
      laneStep: 34,
      maxVisibleCards: 1,
    }
  }
  if (isMedium) {
    return {
      mode: 'medium' as const,
      cardWidth: 62,
      cardScale: 62 / 74,
      noteWidth: 132,
      noteHeight: 34,
      compactNoteHeight: 26,
      actorWidth: 76,
      actorHeight: 106,
      gap: 6,
      laneStep: 42,
      maxVisibleCards: 2,
    }
  }
  return {
    mode: 'desktop' as const,
    cardWidth: 74,
    cardScale: 1,
    noteWidth: 158,
    noteHeight: 46,
    compactNoteHeight: 30,
    actorWidth: 90,
    actorHeight: 122,
    gap: 7,
    laneStep: 52,
    maxVisibleCards: 3,
  }
}

function cardStackWidth(cardWidth: number, visibleCount: number, mode: NarrativeLoopPacketMode) {
  if (visibleCount <= 1) return cardWidth
  const overlap = mode === 'full' ? 18 : mode === 'compact' ? 13 : 8
  return cardWidth + (visibleCount - 1) * overlap
}

function packetNoteText(action: NarrativeLoopAction) {
  if (action.missed) return '未命中'
  if (action.damage) return `伤害 ${action.damage}`
  if (action.outcome === 'resolved') return '已结算'
  if (action.label) return action.label
  return ''
}

function packetRectForMode(
  x: number,
  y: number,
  cardCount: number,
  mode: NarrativeLoopPacketMode,
  notePlacement: NarrativeLoopNotePlacement,
  profile: ReturnType<typeof loopLayoutProfile>,
) {
  if (mode === 'marker') {
    return rectFromCenter(x, y, 42, 24)
  }
  const visibleCount = Math.max(1, Math.min(cardCount, profile.maxVisibleCards))
  const cardWidth = profile.cardWidth
  const cardHeight = cardWidth * 1.5
  const stackWidth = cardStackWidth(cardWidth, visibleCount, mode)
  const noteHeight = mode === 'full' ? profile.noteHeight : profile.compactNoteHeight
  if (notePlacement === 'side') {
    return rectFromCenter(x, y, stackWidth + profile.gap + profile.noteWidth, Math.max(cardHeight, noteHeight))
  }
  if (notePlacement === 'badge') {
    return rectFromCenter(x, y, Math.max(stackWidth, profile.noteWidth * 0.72), cardHeight + profile.gap + profile.compactNoteHeight)
  }
  return rectFromCenter(x, y, Math.max(stackWidth, profile.noteWidth), cardHeight + profile.gap + noteHeight)
}

function actionPointAt(action: NarrativeLoopAction, t: number, size: LoopLayoutSize) {
  const start = percentToPx({ x: action.startX, y: action.startY }, size)
  const control = percentToPx({ x: action.controlX, y: action.controlY }, size)
  const end = percentToPx({ x: action.endX, y: action.endY }, size)
  return loopQuadraticPoint(start, control, end, t)
}

function actionTangentAt(action: NarrativeLoopAction, t: number, size: LoopLayoutSize) {
  const start = percentToPx({ x: action.startX, y: action.startY }, size)
  const control = percentToPx({ x: action.controlX, y: action.controlY }, size)
  const end = percentToPx({ x: action.endX, y: action.endY }, size)
  const x = 2 * (1 - t) * (control.x - start.x) + 2 * t * (end.x - control.x)
  const y = 2 * (1 - t) * (control.y - start.y) + 2 * t * (end.y - control.y)
  const length = Math.hypot(x, y) || 1
  return {
    x: x / length,
    y: y / length,
  }
}

function actorObstacles(actors: NarrativeLoopActor[], size: LoopLayoutSize, profile: ReturnType<typeof loopLayoutProfile>) {
  return actors.map(actor => {
    const point = percentToPx(actor, size)
    return inflateRect(rectFromCenter(point.x, point.y, profile.actorWidth, profile.actorHeight), 10)
  })
}

function packetCandidatePoints(action: NarrativeLoopAction, size: LoopLayoutSize, profile: ReturnType<typeof loopLayoutProfile>) {
  const curvedMid = actionPointAt(action, 0.5, size)
  const straightMid = percentToPx({
    x: (action.startX + action.endX) / 2,
    y: (action.startY + action.endY) / 2,
  }, size)
  const hasStrongCurveBias = Math.hypot(curvedMid.x - straightMid.x, curvedMid.y - straightMid.y) > profile.laneStep * 0.55
  const preferredT = action.index === 0 && action.actionKind === 'attack' && hasStrongCurveBias ? 0.4 : 0.5
  const tValues = preferredT === 0.4
    ? [0.4, 0.32, 0.3, 0.46, 0.34, 0.54, 0.28, 0.66, 0.72]
    : [0.5, 0.38, 0.62, 0.28, 0.72]
  const normalValues = [0, -1, 1, -2, 2, -3, 3]
  const points: Array<{ x: number; y: number; score: number }> = []
  for (const t of tValues) {
    const base = actionPointAt(action, t, size)
    for (const normalIndex of normalValues) {
      points.push({
        x: base.x + action.normalX * normalIndex * profile.laneStep,
        y: base.y + action.normalY * normalIndex * profile.laneStep,
        score: Math.abs(t - preferredT) * 120 + Math.abs(normalIndex) * 18,
      })
    }
  }
  if (preferredT === 0.4) {
    for (const t of [0.29, 0.32]) {
      const base = actionPointAt(action, t, size)
      points.push({
        x: base.x,
        y: base.y - profile.laneStep * 0.55,
        score: Math.abs(t - 0.3) * 80 + 4,
      })
    }
  }
  return points
}

function choosePacketPlacement(
  action: NarrativeLoopAction,
  cardCount: number,
  index: number,
  actionCount: number,
  size: LoopLayoutSize,
  obstacles: LoopLayoutRect[],
  profile: ReturnType<typeof loopLayoutProfile>,
) {
  const forceMarker = profile.mode === 'small' && actionCount > 4 && index < actionCount - 1
  const modeOptions: NarrativeLoopPacketMode[] = forceMarker
    ? ['marker']
    : profile.mode === 'desktop'
      ? ['full', 'compact']
      : ['compact', 'full']
  const placementOptions: NarrativeLoopNotePlacement[] = profile.mode === 'small'
    ? ['badge', 'side', 'bottom']
    : size.height < 330
      ? ['side', 'bottom', 'badge']
      : ['bottom', 'side', 'badge']

  let best: NarrativeLoopPacketPlacement | undefined

  for (const mode of modeOptions) {
    for (const notePlacement of placementOptions) {
      for (const candidate of packetCandidatePoints(action, size, profile)) {
        const rawRect = packetRectForMode(candidate.x, candidate.y, cardCount, mode, notePlacement, profile)
        const center = clampRectCenter(candidate.x, candidate.y, rawRect.width, rawRect.height, size)
        const rect = packetRectForMode(center.x, center.y, cardCount, mode, notePlacement, profile)
        const collisionCount = obstacles.filter(obstacle => rectIntersects(inflateRect(rect, 4), obstacle)).length
        const overflow = rectOutOfBoundsPenalty(rect, size)
        const score = candidate.score +
          collisionCount * 1000 +
          overflow * 20 +
          (mode === 'full' ? 0 : mode === 'compact' ? 120 : 260) +
          (notePlacement === 'bottom' ? 0 : notePlacement === 'side' ? 28 : 66)
        if (!best || score < best.score) {
          best = {
            x: center.x,
            y: center.y,
            rect,
            score,
            mode,
            notePlacement,
          }
        }
      }
    }
    if (best && best.score < 900) break
  }

  const fallback = actionPointAt(action, 0.5, size)
  return best || {
    x: fallback.x,
    y: fallback.y,
    rect: packetRectForMode(fallback.x, fallback.y, cardCount, 'compact', 'badge', profile),
    score: 0,
    mode: 'compact' as const,
    notePlacement: 'badge' as const,
  }
}

function separatePacketPlacement(
  placement: NarrativeLoopPacketPlacement,
  action: NarrativeLoopAction,
  cardCount: number,
  size: LoopLayoutSize,
  profile: ReturnType<typeof loopLayoutProfile>,
  packets: NarrativeLoopPacket[],
) {
  let next = placement
  let attempt = 0
  while (attempt < 8 && packets.some(packet => Math.hypot(
    packet.rect.x + packet.rect.width / 2 - next.x,
    packet.rect.y + packet.rect.height / 2 - next.y,
  ) < Math.min(58, profile.laneStep))) {
    attempt += 1
    const direction = attempt % 2 === 0 ? 1 : -1
    const distance = profile.laneStep * (1 + Math.floor((attempt - 1) / 2) * 0.45)
    const rawX = placement.x + action.normalX * direction * distance
    const rawY = placement.y + action.normalY * direction * distance
    const rawRect = packetRectForMode(rawX, rawY, cardCount, placement.mode, placement.notePlacement, profile)
    const center = clampRectCenter(rawX, rawY, rawRect.width, rawRect.height, size)
    next = {
      ...placement,
      x: center.x,
      y: center.y,
      rect: packetRectForMode(center.x, center.y, cardCount, placement.mode, placement.notePlacement, profile),
    }
  }
  return next
}

function buildActionPackets(loop: NarrativeCombatLoopView, size: LoopLayoutSize) {
  const profile = loopLayoutProfile(size)
  const obstacles = actorObstacles(loop.actors, size, profile)
  const packets: NarrativeLoopPacket[] = []
  for (const action of loop.actions) {
    if (action.cards.length < 1) continue
    const placement = separatePacketPlacement(
      choosePacketPlacement(action, action.cards.length, action.index, loop.actions.length, size, obstacles, profile),
      action,
      action.cards.length,
      size,
      profile,
      packets,
    )
    obstacles.push(inflateRect(placement.rect, 8))
    const percent = pxToPercent({ x: placement.x, y: placement.y }, size)
    const visibleLimit = placement.mode === 'marker' ? 0 : Math.min(action.cards.length, profile.maxVisibleCards)
    const visibleCards = action.cards.slice(0, visibleLimit)
    packets.push({
      id: `packet-${action.id}`,
      action,
      cards: action.cards,
      visibleCards,
      hiddenCardCount: Math.max(0, action.cards.length - visibleCards.length),
      mode: placement.mode,
      notePlacement: placement.notePlacement,
      x: percent.x,
      y: percent.y,
      rect: placement.rect,
      cardWidth: profile.cardWidth,
      cardScale: profile.cardScale,
      noteTitle: `${narrativeLoopActionVerb(action)} ${action.index + 1}`,
      noteDetail: `${roleNameForPlayer(action.sourceId)} → ${roleNameForPlayer(action.targetId)}`,
      noteResult: packetNoteText(action),
      zIndex: 48 + action.index,
    })
  }
  return {
    packets,
    obstacles,
    profile,
  }
}

function nodeLayoutSizes(node: NarrativeLoopNode, profile: ReturnType<typeof loopLayoutProfile>) {
  if (node.kind === 'skill') {
    const fullWidth = profile.mode === 'desktop' ? 126 : profile.mode === 'medium' ? 112 : 96
    const compactWidth = profile.mode === 'desktop' ? 92 : 84
    const fullHeight = node.detail ? 42 : 34
    const compactHeight = node.detail ? 38 : 30
    return [
      { width: fullWidth, height: fullHeight, compact: false },
      { width: compactWidth, height: compactHeight, compact: true },
    ]
  }
  return [{
    width: node.cards.length ? 86 : 108,
    height: node.cards.length ? 116 : node.detail ? 48 : 38,
    compact: false,
  }]
}

function edgeAnchoredNodeCandidates(
  anchorAction: NarrativeLoopAction,
  anchorPacket: NarrativeLoopPacket | undefined,
  width: number,
  height: number,
  localAnchorIndex: number,
  size: LoopLayoutSize,
  profile: ReturnType<typeof loopLayoutProfile>,
) {
  const candidates: Array<{ x: number; y: number; score: number; tier: 'primary' | 'extended' }> = []
  const tValues = [0.5, 0.42, 0.58, 0.34, 0.66]
  const tangentOffsets = [0, -0.44, 0.44]
  const sidePattern = [1, -1, 2, -2, 3, -3]
  const preferredSide = sidePattern[localAnchorIndex % sidePattern.length] || 1
  const packetNormalExtent = anchorPacket
    ? rectHalfExtentAlongVector(anchorPacket.rect, anchorAction.normalX, anchorAction.normalY)
    : 0
  const baseClearance = packetNormalExtent + height / 2 + profile.gap + 14
  const packetCenter = anchorPacket
    ? {
      x: anchorPacket.rect.x + anchorPacket.rect.width / 2,
      y: anchorPacket.rect.y + anchorPacket.rect.height / 2,
    }
    : undefined

  if (packetCenter) {
    const tangent = actionTangentAt(anchorAction, 0.5, size)
    const sideMagnitude = Math.max(1, Math.abs(preferredSide))
    const directions = [
      preferredSide,
      -Math.sign(preferredSide || 1),
      Math.sign(preferredSide || 1) * (sideMagnitude + 1),
    ]
    for (const [directionIndex, direction] of directions.entries()) {
      const normalSign = Math.sign(direction || 1)
      const normalDistance = baseClearance + (Math.abs(direction) - 1) * (height + profile.gap + 8)
      for (const [offsetIndex, tangentOffset] of tangentOffsets.entries()) {
        candidates.push({
          x: packetCenter.x + anchorAction.normalX * normalSign * normalDistance + tangent.x * tangentOffset * width,
          y: packetCenter.y + anchorAction.normalY * normalSign * normalDistance + tangent.y * tangentOffset * width,
          score: directionIndex * 28 + Math.abs(tangentOffset) * 34,
          tier: directionIndex === 0 && offsetIndex === 0 ? 'primary' : 'extended',
        })
      }
    }
  }

  for (const [tIndex, t] of tValues.entries()) {
    const point = actionPointAt(anchorAction, t, size)
    const tangent = actionTangentAt(anchorAction, t, size)
    const sideMagnitude = Math.max(1, Math.abs(preferredSide))
    const directions = [
      preferredSide,
      -Math.sign(preferredSide || 1),
      Math.sign(preferredSide || 1) * (sideMagnitude + 1),
    ]
    for (const [directionIndex, direction] of directions.entries()) {
      const normalSign = Math.sign(direction || 1)
      const normalDistance = baseClearance + (Math.abs(direction) - 1) * (height + profile.gap + 8)
      for (const [offsetIndex, tangentOffset] of tangentOffsets.entries()) {
        const tier = directionIndex === 0 && offsetIndex === 0 && tIndex < 3 ? 'primary' : 'extended'
        candidates.push({
          x: point.x + anchorAction.normalX * normalSign * normalDistance + tangent.x * tangentOffset * width,
          y: point.y + anchorAction.normalY * normalSign * normalDistance + tangent.y * tangentOffset * width,
          score: tIndex * 20 + directionIndex * 42 + Math.abs(tangentOffset) * 36,
          tier,
        })
      }
    }
  }
  return candidates
}

function edgeLineNodeCandidates(
  anchorAction: NarrativeLoopAction,
  width: number,
  height: number,
  localAnchorIndex: number,
  size: LoopLayoutSize,
  profile: ReturnType<typeof loopLayoutProfile>,
) {
  const candidates: Array<{ x: number; y: number; score: number; tier: NarrativeLoopNodeCandidateTier }> = []
  const firstSlots = [0.24, 0.76, 0.16, 0.84, 0.34, 0.66, 0.44, 0.56, 0.2, 0.8]
  const alternateSlots = [0.84, 0.66, 0.92, 0.76, 0.56, 0.44, 0.34, 0.24, 0.16, 0.8]
  const tValues = localAnchorIndex % 2 === 0 ? firstSlots : alternateSlots
  const tangentOffsets = [0, -0.45, 0.45, -0.9, 0.9]
  const sidePattern = localAnchorIndex % 2 === 0
    ? [0, -0.85, 0.85, -1.35, 1.35, -1.85, 1.85, -2.55, 2.55]
    : [0, 0.85, -0.85, 1.35, -1.35, 1.85, -1.85, 2.55, -2.55]
  const tangentUnit = Math.max(width, height) * 0.62
  const normalUnit = Math.min(profile.laneStep, Math.max(width, height) * 0.42)

  for (const [tIndex, t] of tValues.entries()) {
    const point = actionPointAt(anchorAction, t, size)
    const tangent = actionTangentAt(anchorAction, t, size)
    for (const [normalIndex, normalOffset] of sidePattern.entries()) {
      for (const [offsetIndex, tangentOffset] of tangentOffsets.entries()) {
        candidates.push({
          x: point.x + tangent.x * tangentOffset * tangentUnit + anchorAction.normalX * normalOffset * normalUnit,
          y: point.y + tangent.y * tangentOffset * tangentUnit + anchorAction.normalY * normalOffset * normalUnit,
          score: tIndex * 34 + normalIndex * 52 + Math.abs(tangentOffset) * 46 + offsetIndex * 2,
          tier: 'line',
        })
      }
    }
  }
  return candidates
}

function relatedPacketForNode(
  node: NarrativeLoopNode,
  anchorAction: NarrativeLoopAction | undefined,
  packets: NarrativeLoopPacket[],
) {
  if (!anchorAction) return undefined
  const targetIds = new Set(node.targetIds)
  const sameEdgePacket = packets.find(packet => packet.action.id === anchorAction.id)
  if (sameEdgePacket) return sameEdgePacket

  const samePairPackets = packets.filter((packet) => {
    if (packet.action.sourceId !== anchorAction.sourceId || packet.action.targetId !== anchorAction.targetId) return false
    return packet.cards.length > 0
  })
  if (samePairPackets.length > 0) return samePairPackets[samePairPackets.length - 1]

  const nodePairPackets = packets.filter((packet) => {
    if (node.actorId && packet.action.sourceId !== node.actorId) return false
    if (targetIds.size > 0 && !targetIds.has(packet.action.targetId)) return false
    return packet.cards.length > 0
  })
  if (nodePairPackets.length > 0) return nodePairPackets[nodePairPackets.length - 1]

  return undefined
}

function fallbackNodeCandidates(
  node: NarrativeLoopNode,
  actorById: Map<string, NarrativeLoopActor>,
  anchorAction: NarrativeLoopAction | undefined,
  anchorPacket: NarrativeLoopPacket | undefined,
  size: LoopLayoutSize,
) {
  const candidates: Array<{ x: number; y: number; score: number; tier: 'fallback' }> = []
  const actor = actorById.get(node.actorId || '')
  if (actor) {
    const point = percentToPx(actor, size)
    const fallbackScore = anchorAction || anchorPacket ? 760 : 80
    candidates.push(
      { x: point.x, y: point.y + 88, score: fallbackScore, tier: 'fallback' },
      { x: point.x + 104, y: point.y, score: fallbackScore + 24, tier: 'fallback' },
      { x: point.x - 104, y: point.y, score: fallbackScore + 24, tier: 'fallback' },
    )
  }
  const original = percentToPx(node, size)
  candidates.push({
    x: original.x,
    y: original.y,
    score: anchorAction || anchorPacket ? 820 : 120,
    tier: 'fallback',
  })
  return candidates
}

function chooseNodePlacement(
  node: NarrativeLoopNode,
  packets: NarrativeLoopPacket[],
  packetByEdgeId: Map<string, NarrativeLoopPacket>,
  actorById: Map<string, NarrativeLoopActor>,
  actionById: Map<string, NarrativeLoopAction>,
  actions: NarrativeLoopAction[],
  size: LoopLayoutSize,
  obstacles: LoopLayoutRect[],
  profile: ReturnType<typeof loopLayoutProfile>,
  localAnchorIndex: number,
) {
  const anchorAction = node.anchorEdgeId
    ? actionById.get(node.anchorEdgeId)
    : actions.find(action =>
      action.sourceId === node.actorId &&
      (node.targetIds.length < 1 || node.targetIds.includes(action.targetId))
    )
  const directAnchorPacket = node.anchorEdgeId ? packetByEdgeId.get(node.anchorEdgeId) : undefined
  const anchorPacket = directAnchorPacket || relatedPacketForNode(node, anchorAction, packets)
  const sizes = nodeLayoutSizes(node, profile)
  let overlapFallback: (NarrativeLoopNodePlacement & { score: number }) | undefined
  const preferAttackLinePlacement = node.kind === 'skill' && anchorAction?.actionType === 'attack'

  for (const [sizeIndex, nodeSize] of sizes.entries()) {
    const lineCandidates = preferAttackLinePlacement && anchorAction
      ? edgeLineNodeCandidates(anchorAction, nodeSize.width, nodeSize.height, localAnchorIndex, size, profile)
      : []
    const edgeCandidates = anchorAction
      ? edgeAnchoredNodeCandidates(anchorAction, anchorPacket, nodeSize.width, nodeSize.height, localAnchorIndex, size, profile)
      : []
    const packetCandidates = !anchorAction && anchorPacket
      ? edgeAnchoredNodeCandidates(anchorPacket.action, anchorPacket, nodeSize.width, nodeSize.height, localAnchorIndex, size, profile)
      : []
    const fallbackCandidates = fallbackNodeCandidates(node, actorById, anchorAction, anchorPacket, size)
    const candidateBatches = [
      lineCandidates,
      edgeCandidates.filter(candidate => candidate.tier === 'primary'),
      packetCandidates.filter(candidate => candidate.tier === 'primary'),
      edgeCandidates.filter(candidate => candidate.tier === 'extended'),
      packetCandidates.filter(candidate => candidate.tier === 'extended'),
      fallbackCandidates,
    ]

    for (const [batchIndex, batch] of candidateBatches.entries()) {
      let best: (NarrativeLoopNodePlacement & { score: number }) | undefined
      for (const candidate of batch) {
        const rawRect = rectFromCenter(candidate.x, candidate.y, nodeSize.width, nodeSize.height)
        const center = clampRectCenter(candidate.x, candidate.y, nodeSize.width, nodeSize.height, size)
        const rect = rectFromCenter(center.x, center.y, nodeSize.width, nodeSize.height)
        const collisionCount = rectCollisionCount(rect, obstacles)
        const overflow = rectOutOfBoundsPenalty(rawRect, size)
        const score = candidate.score +
          batchIndex * 140 +
          sizeIndex * 260 +
          overflow * 20 +
          collisionCount * 2500
        const placement = {
          x: center.x,
          y: center.y,
          rect,
          width: nodeSize.width,
          height: nodeSize.height,
          zIndex: 78 + localAnchorIndex,
          score,
        }
        if (collisionCount === 0 && (!best || score < best.score)) {
          best = placement
        }
        if (!overlapFallback || score < overlapFallback.score) {
          overlapFallback = placement
        }
      }
      if (best) return best
    }
  }

  const fallback = overlapFallback || (() => {
    const nodeSize = sizes[sizes.length - 1] || { width: 108, height: 38 }
    const original = percentToPx(node, size)
    const center = clampRectCenter(original.x, original.y, nodeSize.width, nodeSize.height, size)
    return {
      x: center.x,
      y: center.y,
      rect: rectFromCenter(center.x, center.y, nodeSize.width, nodeSize.height),
      width: nodeSize.width,
      height: nodeSize.height,
      zIndex: 88 + localAnchorIndex,
      score: Number.POSITIVE_INFINITY,
    }
  })()
  return {
    ...fallback,
    zIndex: Math.max(fallback.zIndex, 88 + localAnchorIndex),
  }
}

function resolveNodeCollisions(
  loop: NarrativeCombatLoopView,
  packets: NarrativeLoopPacket[],
  obstacles: LoopLayoutRect[],
  size: LoopLayoutSize,
  profile: ReturnType<typeof loopLayoutProfile>,
) {
  const packetByEdgeId = new Map(packets.map(packet => [packet.action.id, packet]))
  const actorById = new Map(loop.actors.map(actor => [actor.id, actor]))
  const actionById = new Map(loop.actions.map(action => [action.id, action]))
  const anchorCounts = new Map<string, number>()
  return loop.nodes.map((node) => {
    const anchorKey = node.anchorEdgeId || ''
    const localAnchorIndex = anchorKey ? anchorCounts.get(anchorKey) || 0 : 0
    if (anchorKey) anchorCounts.set(anchorKey, localAnchorIndex + 1)
    const placement = chooseNodePlacement(node, packets, packetByEdgeId, actorById, actionById, loop.actions, size, obstacles, profile, localAnchorIndex)
    obstacles.push(inflateRect(placement.rect, 8))
    const point = pxToPercent({ x: placement.x, y: placement.y }, size)
    return {
      ...node,
      x: point.x,
      y: point.y,
      width: placement.width,
      height: placement.height,
      rect: placement.rect,
      zIndex: placement.zIndex,
    }
  })
}

const narrativeCombatLoopLayout = computed<NarrativeCombatLoopLayout | null>(() => {
  const loop = narrativeCombatLoop.value
  if (!loop) return null
  const size = narrativeLoopFieldSize.value
  const { packets, obstacles, profile } = buildActionPackets(loop, size)
  return {
    ...loop,
    packets,
    nodes: resolveNodeCollisions(loop, packets, obstacles, size, profile),
  }
})

function narrativeLoopActorClasses(actor: NarrativeLoopActor) {
  const player = players.value[actor.id]
  return [
    `narrative-loop-actor--${player?.camp === 'Red' ? 'red' : 'blue'}`,
  ]
}

function narrativeLoopActorStyle(actor: NarrativeLoopActor) {
  return {
    '--loop-x': `${actor.x}%`,
    '--loop-y': `${actor.y}%`,
    zIndex: String(30 + actor.index),
  }
}

function narrativeLoopActionClasses(action: NarrativeLoopAction) {
  return [
    `narrative-loop-action--${action.tone}`,
    `narrative-loop-action--${action.actionKind}`,
    `narrative-loop-action--phase-${action.actionType || 'unknown'}`,
    action.outcome ? `narrative-loop-action--outcome-${action.outcome}` : '',
    action.missed ? 'narrative-loop-action--missed' : '',
    action.item?.stepStatus ? `narrative-loop-action--step-${action.item.stepStatus}` : '',
  ]
}

function narrativeLoopPacketClasses(packet: NarrativeLoopPacket) {
  return [
    ...narrativeLoopActionClasses(packet.action),
    `narrative-loop-packet--${packet.mode}`,
    `narrative-loop-packet--note-${packet.notePlacement}`,
  ]
}

function narrativeLoopPacketStyle(packet: NarrativeLoopPacket) {
  return {
    '--loop-packet-x': `${packet.x}%`,
    '--loop-packet-y': `${packet.y}%`,
    '--loop-card-width': `${packet.cardWidth}px`,
    '--loop-card-scale': String(packet.cardScale),
    '--loop-packet-width': `${packet.rect.width}px`,
    '--loop-packet-height': `${packet.rect.height}px`,
    zIndex: String(packet.zIndex),
  }
}

function narrativeLoopPacketCardStyle(packet: NarrativeLoopPacket, cardIndex = 0) {
  const visibleCount = Math.max(packet.visibleCards.length, 1)
  const offsetStep = packet.mode === 'full' ? 18 : packet.mode === 'compact' ? 13 : 8
  const offset = (cardIndex - (visibleCount - 1) / 2) * offsetStep
  return {
    '--loop-card-x': `${packet.x}%`,
    '--loop-card-y': `${packet.y}%`,
    '--loop-card-offset-x': `${offset}px`,
    '--loop-card-rotate': `${(cardIndex - (visibleCount - 1) / 2) * 3}deg`,
    zIndex: String(packet.zIndex + cardIndex),
  }
}

function narrativeLoopPacketNoteStyle(packet: NarrativeLoopPacket) {
  return {
    '--loop-label-x': `${packet.x}%`,
    '--loop-label-y': `${packet.y}%`,
    zIndex: String(packet.zIndex + 20),
  }
}

function narrativeLoopMarkerUrl(action: NarrativeLoopAction) {
  return `url(#narrative-loop-arrow-${action.tone})`
}

function narrativeLoopActionVerb(action: NarrativeLoopAction) {
  if (action.actionType === 'magic') return '法术'
  if (action.actionType === 'skill') return '技能'
  if (action.actionType === 'effect') return '效果'
  if (action.actionType === 'take') return '承受'
  if (action.actionType === 'defend') return '防御'
  if (action.actionType === 'shield') return '圣盾'
  if (action.index === 0 && action.actionType === 'attack') return '攻击'
  return '应战'
}

function narrativeLoopNodeClasses(node: NarrativeLoopNode) {
  return [
    `narrative-loop-node--${node.kind}`,
    node.outcome ? `narrative-loop-node--outcome-${node.outcome}` : '',
    node.anchorEdgeId ? 'narrative-loop-node--anchored' : '',
  ]
}

function narrativeLoopNodeStyle(node: NarrativeLoopNode) {
  return {
    '--loop-node-x': `${node.x}%`,
    '--loop-node-y': `${node.y}%`,
    '--loop-node-width': `${node.width || 108}px`,
    '--loop-node-height': `${node.height || 38}px`,
    zIndex: String(node.zIndex || 50 + (node.damage || 0)),
  }
}

function isPrimaryCounterChainDetachedItem(item: NarrativeStackItem) {
  const chain = primaryCounterChain.value
  if (!chain) return false
  return item.id === chain.attackItem.id || item.id === chain.missItem?.id
}

const narrativeStepGroups = computed<NarrativeStepGroup[]>(() => {
  const groups = new Map<string, NarrativeStepGroup>()

  for (const item of narrativeStackItems.value) {
    if (isSealFieldEffectItem(item)) continue
    if (isPrimaryCounterChainDetachedItem(item)) continue
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
  const sourceSide = narrativeEndpointSideForPlayer(item.sourcePlayerId)
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
    '--stack-enter-x': narrativeEndpointSideForPlayer(item.sourcePlayerId) === 'left' ? '-42px' : '42px',
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

function narrativeCounterChainActorClasses(playerId: string, endpoint: 'source' | 'target') {
  const player = players.value[playerId]
  return [
    `narrative-counter-chain__actor--${endpoint}`,
    `narrative-counter-chain__actor--${player?.camp === 'Red' ? 'red' : 'blue'}`,
  ]
}

function narrativeCounterChainMissClasses() {
  const status = primaryCounterChain.value?.missItem?.stepStatus
  return [
    status ? `narrative-counter-chain__miss--${status}` : '',
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

function fallbackActorPoint(playerId: string, layerRect: DOMRect, targetSide = narrativeEndpointSideForPlayer(playerId)) {
  const row = narrativeRowForPlayer(playerId)
  const xPercent = actorLineX(targetSide)
  return {
    x: layerRect.width * xPercent / 100,
    y: layerRect.height * rowPercent(row) / 100,
  }
}

function fallbackTargetPoint(playerId: string, layerRect: DOMRect) {
  const side = narrativeEndpointSideForPlayer(playerId, 'opposed')
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
    if (isPrimaryCounterChainDetachedItem(item)) return false
    if (narrativePlayback.value?.isReview) return true
    return !narrativePlayback.value || item.stepStatus !== 'pending'
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
    const side = narrativeEndpointSideForPlayer(endpointId, 'opposed')
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

function measureNarrativeLoopField() {
  const field = narrativeLoopFieldRef.value
  if (!field) return
  const rect = field.getBoundingClientRect()
  if (rect.width >= 10 && rect.height >= 10) {
    narrativeLoopFieldSize.value = {
      width: rect.width,
      height: rect.height,
    }
  }
}

function observeNarrativeLoopField(field: HTMLElement | null) {
  narrativeLoopFieldResizeObserver?.disconnect()
  narrativeLoopFieldResizeObserver = null
  if (!field) return
  measureNarrativeLoopField()
  if (typeof ResizeObserver !== 'undefined') {
    narrativeLoopFieldResizeObserver = new ResizeObserver(() => measureNarrativeLoopField())
    narrativeLoopFieldResizeObserver.observe(field)
  }
}

watch(
  narrativeLoopFieldRef,
  field => observeNarrativeLoopField(field),
  { flush: 'post' },
)

onMounted(() => {
  if (typeof ResizeObserver !== 'undefined' && narrativeMistLayerRef.value) {
    narrativeMistResizeObserver = new ResizeObserver(() => scheduleNarrativeMistMeasure())
    narrativeMistResizeObserver.observe(narrativeMistLayerRef.value)
  }
  observeNarrativeLoopField(narrativeLoopFieldRef.value)
  window.addEventListener('resize', measureNarrativeLoopField)
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
  narrativeLoopFieldResizeObserver?.disconnect()
  narrativeLoopFieldResizeObserver = null
  narrativeGsapTimeline?.kill()
  narrativeGsapTimeline = null
  narrativeGsapContext?.revert()
  narrativeGsapContext = null
  window.removeEventListener('resize', measureNarrativeLoopField)
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
      <div
        v-if="narrativeCombatLoopLayout"
        class="narrative-loop-stage"
        data-testid="narrative-combat-loop"
      >
        <div
          ref="narrativeLoopFieldRef"
          class="narrative-loop-field"
        >
          <svg
            class="narrative-loop-svg"
            viewBox="0 0 100 100"
            preserveAspectRatio="none"
            aria-hidden="true"
          >
            <defs>
              <marker
                id="narrative-loop-arrow-red"
                markerWidth="8"
                markerHeight="8"
                refX="7"
                refY="4"
                orient="auto"
                markerUnits="strokeWidth"
              >
                <path d="M 0 0 L 8 4 L 0 8 z" class="narrative-loop-arrow-head narrative-loop-arrow-head--red" />
              </marker>
              <marker
                id="narrative-loop-arrow-gold"
                markerWidth="8"
                markerHeight="8"
                refX="7"
                refY="4"
                orient="auto"
                markerUnits="strokeWidth"
              >
                <path d="M 0 0 L 8 4 L 0 8 z" class="narrative-loop-arrow-head narrative-loop-arrow-head--gold" />
              </marker>
              <filter id="narrative-loop-glow" x="-30%" y="-70%" width="160%" height="240%">
                <feGaussianBlur stdDeviation="1.4" />
              </filter>
            </defs>
            <g
              v-for="action in narrativeCombatLoopLayout.actions"
              :key="`loop-link-${action.id}`"
              class="narrative-loop-link"
              :class="narrativeLoopActionClasses(action)"
            >
              <path
                class="narrative-loop-link__aura"
                :d="action.path"
              />
              <path
                class="narrative-loop-link__flow"
                :d="action.path"
                :marker-end="narrativeLoopMarkerUrl(action)"
              />
            </g>
          </svg>

          <div
            v-for="actor in narrativeCombatLoopLayout.actors"
            :key="`loop-actor-${actor.id}`"
            class="narrative-loop-actor"
            :class="narrativeLoopActorClasses(actor)"
            :style="narrativeLoopActorStyle(actor)"
            :data-narrative-actor-id="actor.id"
            :data-narrative-loop-actor-id="actor.id"
          >
            <img
              v-if="portraitSrcForPlayer(actor.id)"
              :src="portraitSrcForPlayer(actor.id)"
              :alt="roleNameForPlayer(actor.id)"
            >
            <div class="narrative-loop-actor__name">
              {{ roleNameForPlayer(actor.id) }}
            </div>
            <div
              v-if="latestDamageForPlayer(actor.id)"
              class="narrative-loop-actor__damage"
            >
              -{{ latestDamageForPlayer(actor.id)?.damage }}
            </div>
          </div>

          <div
            v-for="packet in narrativeCombatLoopLayout.packets"
            :key="packet.id"
            class="narrative-loop-packet"
            :class="narrativeLoopPacketClasses(packet)"
            :style="narrativeLoopPacketStyle(packet)"
            :data-narrative-packet-id="packet.id"
          >
            <div
              v-if="packet.mode === 'marker'"
              class="narrative-loop-packet-marker"
            >
              {{ narrativeLoopActionVerb(packet.action) }}
            </div>
            <div
              v-else
              class="narrative-loop-card-stack"
            >
              <div
                v-for="(loopCard, cardIndex) in packet.visibleCards"
                :key="loopCard.id"
                class="narrative-loop-card"
                :class="narrativeLoopActionClasses(packet.action)"
                :style="narrativeLoopPacketCardStyle(packet, cardIndex)"
                :data-narrative-stack-id="packet.action.item?.id || loopCard.id"
                :data-narrative-card-id="packet.action.cardView?.id || loopCard.id"
              >
                <div class="narrative-loop-card__label">
                  {{ narrativeLoopActionVerb(packet.action) }}
                </div>
                <CardComponent :card="loopCard.card" battle-mini />
              </div>
              <div
                v-if="packet.hiddenCardCount > 0"
                class="narrative-loop-card-overflow"
              >
                +{{ packet.hiddenCardCount }}
              </div>
            </div>
            <div
              class="narrative-loop-note"
              :class="narrativeLoopActionClasses(packet.action)"
              :style="narrativeLoopPacketNoteStyle(packet)"
            >
              <strong>{{ packet.noteTitle }}</strong>
              <span v-if="packet.mode === 'full'">{{ packet.noteDetail }}</span>
              <small v-if="packet.noteResult">{{ packet.noteResult }}</small>
            </div>
          </div>

          <div
            v-for="node in narrativeCombatLoopLayout.nodes"
            :key="`loop-node-${node.id}`"
            class="narrative-loop-node"
            :class="narrativeLoopNodeClasses(node)"
            :style="narrativeLoopNodeStyle(node)"
            :data-narrative-flow-node-id="node.id"
          >
            <template v-if="node.cards.length">
              <template
                v-for="loopCard in node.cards.slice(0, 1)"
                :key="loopCard.id"
              >
                <div class="narrative-loop-node__label">
                  {{ node.title }}
                </div>
                <CardComponent :card="loopCard.card" battle-mini />
              </template>
            </template>
            <template v-else>
              <strong>{{ node.title }}</strong>
              <span v-if="node.detail">{{ node.detail }}</span>
            </template>
          </div>

        </div>

      </div>

      <div v-if="actionNarrative && !narrativeCombatLoop" class="narrative-actors">
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
        v-if="actionNarrative && !narrativeCombatLoop && narrativeSealFieldEffectItems.length"
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
        v-if="actionNarrative && !narrativeCombatLoop && primaryCounterChainCards"
        class="narrative-counter-chain"
        data-testid="narrative-counter-chain"
      >
        <div class="narrative-counter-chain__track">
          <div
            class="narrative-counter-chain__actor"
            :class="narrativeCounterChainActorClasses(primaryCounterChainCards.sourceId, 'source')"
            :data-narrative-actor-id="primaryCounterChainCards.sourceId"
          >
            <img
              v-if="portraitSrcForPlayer(primaryCounterChainCards.sourceId)"
              :src="portraitSrcForPlayer(primaryCounterChainCards.sourceId)"
              :alt="roleNameForPlayer(primaryCounterChainCards.sourceId)"
            >
            <span>{{ roleNameForPlayer(primaryCounterChainCards.sourceId) }}</span>
          </div>

          <div
            class="narrative-counter-chain__line"
            aria-hidden="true"
          ></div>

          <div
            class="narrative-counter-chain__actor"
            :class="narrativeCounterChainActorClasses(primaryCounterChainCards.targetId, 'target')"
            :data-narrative-actor-id="primaryCounterChainCards.targetId"
          >
            <img
              v-if="portraitSrcForPlayer(primaryCounterChainCards.targetId)"
              :src="portraitSrcForPlayer(primaryCounterChainCards.targetId)"
              :alt="roleNameForPlayer(primaryCounterChainCards.targetId)"
            >
            <span>{{ roleNameForPlayer(primaryCounterChainCards.targetId) }}</span>
          </div>

          <div
            v-if="primaryCounterChain"
            class="narrative-stack-item narrative-stack-item--chain-card"
            :class="narrativeStackItemClasses(primaryCounterChain.attackItem)"
            :style="narrativeStackItemStyle(primaryCounterChain.attackItem)"
            :data-narrative-stack-id="primaryCounterChain.attackItem.id"
            :data-narrative-card-id="primaryCounterChain.attackItem.cardView?.id"
          >
            <div class="narrative-counter-chain__card-label">
              发起攻击
            </div>
            <div
              v-if="primaryCounterChain.attackItem.cardView"
              class="narrative-played-card"
              :class="narrativePlayedCardClasses(primaryCounterChain.attackItem.cardView)"
            >
              <CardComponent :card="primaryCounterChain.attackItem.cardView.card" battle-mini />
            </div>
          </div>

          <button
            v-if="primaryCounterChain?.missItem || primaryCounterChainCards.missEvent"
            type="button"
            class="narrative-counter-chain__miss"
            :class="narrativeCounterChainMissClasses()"
            tabindex="-1"
          >
            未命中
          </button>
        </div>
      </div>

      <div
        v-if="actionNarrative && !narrativeCombatLoop && narrativeStepGroups.length"
        class="narrative-stack-lane"
        :class="{
          'narrative-stack-lane--review': narrativePlayback?.isReview,
          'narrative-stack-lane--with-counter-chain': primaryCounterChainCards,
        }"
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
        v-if="actionNarrative && !narrativeCombatLoop && narrativeMistBlueprints.length"
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

      <div v-if="actionNarrative && !narrativeCombatLoop" class="narrative-settled-row">
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
  --narrative-actor-card-edge: clamp(42px, 4.2vw, 72px);
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

.narrative-loop-stage {
  position: absolute;
  inset: 10px 18px 48px;
  z-index: 14;
  display: block;
  pointer-events: none;
}

.narrative-loop-field {
  position: relative;
  min-width: 0;
  width: 100%;
  height: 100%;
  min-height: 260px;
  isolation: isolate;
}

.narrative-loop-svg {
  position: absolute;
  inset: 0;
  z-index: 1;
  width: 100%;
  height: 100%;
  overflow: visible;
}

.narrative-loop-arrow-head--red {
  fill: #ff7568;
  filter: drop-shadow(0 0 4px rgba(255, 94, 82, 0.72));
}

.narrative-loop-arrow-head--gold {
  fill: #ffe28a;
  filter: drop-shadow(0 0 4px rgba(255, 214, 112, 0.72));
}

.narrative-loop-link {
  --loop-flow: rgba(255, 225, 138, 0.94);
  --loop-aura: rgba(255, 198, 88, 0.26);
}

.narrative-loop-action--red {
  --loop-flow: rgba(255, 106, 92, 0.96);
  --loop-aura: rgba(255, 80, 72, 0.28);
}

.narrative-loop-link__aura,
.narrative-loop-link__flow {
  fill: none;
  stroke-linecap: round;
  stroke-linejoin: round;
  vector-effect: non-scaling-stroke;
}

.narrative-loop-link__aura {
  stroke: var(--loop-aura);
  stroke-width: 8px;
  filter: url(#narrative-loop-glow);
}

.narrative-loop-link__flow {
  stroke: var(--loop-flow);
  stroke-width: 2.6px;
  stroke-dasharray: 8 7;
  filter: drop-shadow(0 0 7px var(--loop-flow));
  animation: narrativeLoopCurrent 1.15s linear infinite;
}

.narrative-loop-actor {
  position: absolute;
  left: var(--loop-x);
  top: var(--loop-y);
  width: clamp(58px, 6.4vw, 82px);
  height: clamp(82px, 9vw, 112px);
  overflow: hidden;
  border-radius: 8px;
  border: 1px solid rgba(142, 183, 215, 0.48);
  background: rgba(7, 17, 29, 0.88);
  box-shadow:
    inset 0 1px 0 rgba(240, 248, 255, 0.12),
    0 0 18px rgba(255, 226, 150, 0.16),
    0 12px 24px rgba(0, 0, 0, 0.42);
  transform: translate(-50%, -50%);
  animation: narrativeLoopActorIn 0.3s cubic-bezier(0.2, 0.86, 0.24, 1) both;
}

.narrative-loop-actor--blue {
  border-color: rgba(82, 190, 250, 0.64);
  box-shadow:
    inset 0 1px 0 rgba(240, 248, 255, 0.12),
    0 0 18px rgba(56, 189, 248, 0.24),
    0 12px 24px rgba(0, 0, 0, 0.42);
}

.narrative-loop-actor--red {
  border-color: rgba(248, 113, 113, 0.66);
  box-shadow:
    inset 0 1px 0 rgba(255, 240, 240, 0.12),
    0 0 18px rgba(248, 113, 113, 0.24),
    0 12px 24px rgba(0, 0, 0, 0.42);
}

.narrative-loop-actor img {
  width: 100%;
  height: 100%;
  border-radius: inherit;
  object-fit: cover;
  object-position: 50% 12%;
}

.narrative-loop-actor__name {
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  overflow: hidden;
  padding: 5px 4px;
  color: rgba(240, 248, 255, 0.96);
  font-size: 11px;
  font-weight: 900;
  line-height: 1;
  text-align: center;
  text-overflow: ellipsis;
  white-space: nowrap;
  background: linear-gradient(180deg, rgba(4, 10, 18, 0), rgba(4, 10, 18, 0.94) 38%);
  text-shadow: 0 1px 3px rgba(0, 0, 0, 0.88);
}

.narrative-loop-actor__damage {
  position: absolute;
  left: 50%;
  top: 44%;
  z-index: 4;
  color: #ffb199;
  font-size: 28px;
  font-weight: 950;
  line-height: 1;
  text-shadow:
    0 2px 4px rgba(0, 0, 0, 0.9),
    0 0 16px rgba(248, 113, 113, 0.7);
  transform: translate(-50%, -50%);
  animation: narrativeDamagePop 0.86s ease-out both;
}

.narrative-loop-packet {
  position: absolute;
  left: var(--loop-packet-x);
  top: var(--loop-packet-y);
  width: var(--loop-packet-width);
  height: var(--loop-packet-height);
  min-width: 42px;
  min-height: 24px;
  box-sizing: border-box;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 7px;
  transform: translate(-50%, -50%);
  animation: narrativeLoopCardIn 0.32s cubic-bezier(0.2, 0.86, 0.24, 1) both;
}

.narrative-loop-packet--note-bottom,
.narrative-loop-packet--note-badge {
  flex-direction: column;
}

.narrative-loop-packet--note-side {
  flex-direction: row;
}

.narrative-loop-card-stack {
  position: relative;
  width: 100%;
  min-height: calc(var(--loop-card-width) * 1.5 + 14px);
  display: flex;
  align-items: flex-start;
  justify-content: center;
}

.narrative-loop-card {
  position: relative;
  width: var(--loop-card-width);
  flex: 0 0 auto;
  margin-inline: calc(var(--loop-card-width) * -0.05);
  transform:
    translateX(var(--loop-card-offset-x, 0))
    rotate(var(--loop-card-rotate, 0deg));
  filter:
    drop-shadow(0 12px 20px rgba(0, 0, 0, 0.46))
    drop-shadow(0 0 11px rgba(255, 226, 150, 0.24));
}

.narrative-loop-card.narrative-loop-action--red {
  filter:
    drop-shadow(0 12px 20px rgba(0, 0, 0, 0.46))
    drop-shadow(0 0 13px rgba(255, 106, 92, 0.34));
}

.narrative-loop-card :deep(.card-battle-mini) {
  width: var(--loop-card-width) !important;
  height: calc(var(--loop-card-width) * 1.5) !important;
  transform: none !important;
}

.narrative-loop-card__label {
  margin-bottom: 3px;
  color: rgba(255, 239, 202, 0.96);
  font-size: 11px;
  font-weight: 950;
  line-height: 1;
  text-align: center;
  text-shadow:
    0 1px 3px rgba(0, 0, 0, 0.9),
    0 0 10px rgba(220, 166, 80, 0.28);
}

.narrative-loop-action--red .narrative-loop-card__label {
  color: #ffd0bd;
}

.narrative-loop-card-overflow {
  position: absolute;
  right: 0;
  bottom: 4px;
  min-width: 24px;
  padding: 2px 5px;
  border-radius: 999px;
  border: 1px solid rgba(255, 239, 202, 0.62);
  background: rgba(5, 14, 25, 0.86);
  color: #fff1bc;
  font-size: 11px;
  font-weight: 950;
  line-height: 1;
  text-align: center;
  box-shadow: 0 0 10px rgba(255, 226, 150, 0.22);
}

.narrative-loop-packet-marker {
  min-width: 42px;
  padding: 5px 8px;
  border-radius: 999px;
  border: 1px solid var(--loop-flow, rgba(255, 225, 138, 0.94));
  background: rgba(4, 13, 24, 0.78);
  color: rgba(255, 239, 202, 0.96);
  font-size: 11px;
  font-weight: 950;
  line-height: 1;
  text-align: center;
  text-shadow: 0 1px 3px rgba(0, 0, 0, 0.88);
  box-shadow: 0 0 12px var(--loop-aura, rgba(255, 198, 88, 0.26));
}

.narrative-loop-node {
  position: absolute;
  left: var(--loop-node-x);
  top: var(--loop-node-y);
  width: var(--loop-node-width);
  height: var(--loop-node-height);
  box-sizing: border-box;
  padding: 6px 8px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 3px;
  border-radius: 8px;
  border: 1px solid rgba(255, 226, 150, 0.48);
  background: rgba(5, 15, 27, 0.78);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.1),
    0 0 14px rgba(255, 219, 128, 0.2),
    0 10px 18px rgba(0, 0, 0, 0.34);
  color: rgba(241, 247, 255, 0.96);
  text-align: center;
  text-shadow: 0 1px 3px rgba(0, 0, 0, 0.88);
  transform: translate(-50%, -50%);
  animation: narrativeLoopCardIn 0.32s cubic-bezier(0.2, 0.86, 0.24, 1) both;
}

.narrative-loop-node--skill {
  border-color: rgba(190, 219, 255, 0.52);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.1),
    0 0 16px rgba(125, 211, 252, 0.24),
    0 10px 18px rgba(0, 0, 0, 0.34);
}

.narrative-loop-node--outcome-miss {
  border-color: rgba(255, 118, 103, 0.68);
  color: #ffd2bd;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.12),
    0 0 18px rgba(255, 99, 88, 0.34),
    0 10px 18px rgba(0, 0, 0, 0.36);
}

.narrative-loop-node strong,
.narrative-loop-node__label {
  max-width: 100%;
  overflow: hidden;
  color: #fff1bc;
  font-size: 12px;
  font-weight: 950;
  line-height: 1.05;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.narrative-loop-node span {
  max-width: 100%;
  overflow: hidden;
  color: rgba(223, 235, 247, 0.9);
  font-size: 10px;
  line-height: 1.15;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.narrative-loop-node :deep(.card-battle-mini) {
  width: 68px;
  transform: none !important;
}

.narrative-loop-note {
  position: relative;
  width: max-content;
  max-width: 158px;
  padding: 5px 7px;
  display: flex;
  flex-direction: column;
  gap: 2px;
  border-left: 2px solid var(--loop-flow, rgba(255, 225, 138, 0.94));
  background: rgba(4, 13, 24, 0.76);
  box-shadow:
    inset 0 1px 0 rgba(235, 248, 255, 0.08),
    0 8px 16px rgba(0, 0, 0, 0.28);
  color: rgba(230, 242, 250, 0.94);
  text-shadow: 0 1px 3px rgba(0, 0, 0, 0.88);
}

.narrative-loop-packet--note-badge .narrative-loop-note {
  max-width: 118px;
  padding: 4px 7px;
  flex-direction: row;
  align-items: center;
}

.narrative-loop-note strong {
  color: #fff1bc;
  font-size: 11px;
  font-weight: 950;
  line-height: 1;
  white-space: nowrap;
}

.narrative-loop-note span,
.narrative-loop-note small {
  overflow: hidden;
  font-size: 10px;
  line-height: 1.15;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.narrative-loop-note small {
  color: #ffd7a3;
  font-weight: 800;
}

.narrative-loop-packet--compact .narrative-loop-note span,
.narrative-loop-packet--marker .narrative-loop-note,
.narrative-loop-packet--note-badge .narrative-loop-note span {
  display: none;
}

@media (max-width: 900px) {
  .narrative-loop-stage {
    inset: 6px 8px 42px;
  }

  .narrative-loop-field {
    min-height: 250px;
  }

  .narrative-loop-packet {
    max-width: 132px;
  }
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

.narrative-counter-chain {
  position: absolute;
  left: 50%;
  top: 8px;
  z-index: 12;
  width: min(430px, calc(100% - 260px));
  min-width: 310px;
  height: 148px;
  transform: translateX(-50%);
  pointer-events: none;
}

.narrative-counter-chain__track {
  position: relative;
  width: 100%;
  height: 100%;
}

.narrative-counter-chain__line {
  position: absolute;
  left: 88px;
  right: 88px;
  top: 92px;
  height: 3px;
  border-radius: 999px;
  background: linear-gradient(
    90deg,
    rgba(127, 211, 255, 0.28),
    rgba(255, 232, 168, 0.9) 48%,
    rgba(127, 211, 255, 0.28)
  );
  box-shadow:
    0 0 14px rgba(125, 211, 252, 0.32),
    0 0 18px rgba(255, 232, 168, 0.26);
}

.narrative-counter-chain__line::before,
.narrative-counter-chain__line::after {
  content: '';
  position: absolute;
  top: 50%;
  width: 8px;
  height: 8px;
  border-radius: 999px;
  background: #fff2bc;
  box-shadow: 0 0 12px rgba(255, 236, 184, 0.72);
  transform: translate(-50%, -50%);
}

.narrative-counter-chain__line::before {
  left: 0;
}

.narrative-counter-chain__line::after {
  left: 100%;
}

.narrative-counter-chain__actor {
  position: absolute;
  top: 48px;
  z-index: 2;
  width: 64px;
  height: 86px;
  overflow: hidden;
  border-radius: 8px;
  background: rgba(7, 17, 29, 0.86);
  border: 1px solid rgba(132, 172, 207, 0.48);
  box-shadow:
    inset 0 1px 0 rgba(240, 248, 255, 0.12),
    0 10px 20px rgba(1, 7, 14, 0.36);
  animation: narrativeCounterActorIn 0.28s cubic-bezier(0.2, 0.86, 0.24, 1) both;
}

.narrative-counter-chain__actor--source {
  left: 18px;
}

.narrative-counter-chain__actor--target {
  right: 18px;
}

.narrative-counter-chain__actor--blue {
  border-color: rgba(82, 190, 250, 0.66);
  box-shadow:
    inset 0 1px 0 rgba(240, 248, 255, 0.12),
    0 0 18px rgba(56, 189, 248, 0.24),
    0 10px 20px rgba(1, 7, 14, 0.36);
}

.narrative-counter-chain__actor--red {
  border-color: rgba(248, 113, 113, 0.66);
  box-shadow:
    inset 0 1px 0 rgba(255, 240, 240, 0.12),
    0 0 18px rgba(248, 113, 113, 0.24),
    0 10px 20px rgba(1, 7, 14, 0.36);
}

.narrative-counter-chain__actor img {
  width: 100%;
  height: 100%;
  border-radius: inherit;
  object-fit: cover;
  object-position: 50% 12%;
}

.narrative-counter-chain__actor span {
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  overflow: hidden;
  padding: 4px 3px;
  color: rgba(240, 248, 255, 0.96);
  font-size: 10px;
  font-weight: 800;
  line-height: 1;
  text-align: center;
  text-overflow: ellipsis;
  white-space: nowrap;
  background: linear-gradient(180deg, rgba(4, 10, 18, 0), rgba(4, 10, 18, 0.92) 34%);
  text-shadow: 0 1px 3px rgba(0, 0, 0, 0.86);
}

.narrative-counter-chain .narrative-stack-item--chain-card {
  position: absolute;
  left: 50%;
  top: 0;
  z-index: 4;
  width: 74px;
  transform: translateX(-50%);
  animation: narrativeCounterChainCardIn 0.32s cubic-bezier(0.2, 0.86, 0.24, 1) both;
}

.narrative-counter-chain .narrative-stack-item--chain-card .narrative-played-card {
  width: 74px;
}

.narrative-counter-chain__card-label {
  margin-bottom: 3px;
  color: rgba(255, 239, 202, 0.96);
  font-size: 11px;
  font-weight: 900;
  line-height: 1;
  text-align: center;
  text-shadow:
    0 1px 3px rgba(0, 0, 0, 0.88),
    0 0 10px rgba(220, 166, 80, 0.24);
}

.narrative-counter-chain__miss {
  position: absolute;
  left: 50%;
  top: 106px;
  z-index: 5;
  appearance: none;
  min-width: 74px;
  height: 28px;
  padding: 0 12px;
  border: 1px solid rgba(255, 232, 168, 0.74);
  border-radius: 999px;
  background:
    radial-gradient(circle at 50% 0%, rgba(255, 246, 208, 0.34), transparent 52%),
    rgba(8, 18, 30, 0.86);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.16),
    0 0 18px rgba(255, 226, 150, 0.34),
    0 10px 20px rgba(0, 0, 0, 0.34);
  color: #fff2bc;
  cursor: default;
  font-family: inherit;
  font-size: 13px;
  font-weight: 950;
  line-height: 1;
  text-align: center;
  text-shadow: 0 1px 3px rgba(0, 0, 0, 0.86);
  transform: translateX(-50%);
  animation: narrativeCounterMissIn 0.32s cubic-bezier(0.2, 0.86, 0.24, 1) both;
}

.narrative-counter-chain__miss--active {
  border-color: rgba(255, 245, 190, 0.94);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.18),
    0 0 24px rgba(255, 232, 168, 0.48),
    0 12px 22px rgba(0, 0, 0, 0.38);
}

@media (max-width: 820px) {
  .narrative-counter-chain {
    width: min(360px, calc(100% - 24px));
    min-width: 0;
  }

  .narrative-counter-chain__actor {
    top: 54px;
    width: 54px;
    height: 74px;
  }

  .narrative-counter-chain__actor--source {
    left: 4px;
  }

  .narrative-counter-chain__actor--target {
    right: 4px;
  }

  .narrative-counter-chain__line {
    left: 62px;
    right: 62px;
    top: 91px;
  }

  .narrative-counter-chain .narrative-stack-item--chain-card,
  .narrative-counter-chain .narrative-stack-item--chain-card .narrative-played-card {
    width: 66px;
  }

  .narrative-counter-chain__miss {
    top: 104px;
  }
}

.narrative-stack-lane--review {
  align-content: flex-start;
  justify-content: center;
  padding: 8px 4px;
}

.narrative-stack-lane--with-counter-chain {
  align-content: flex-start;
  padding-top: 148px;
}

.narrative-stack-lane--review.narrative-stack-lane--with-counter-chain {
  padding-top: 148px;
}

.narrative-step-group {
  position: relative;
  flex: 0 1 auto;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 5px;
  max-width: 192px;
  padding: 4px 6px 6px;
  border: 0;
  background: transparent;
  transition:
    opacity 0.24s ease,
    transform 0.24s ease,
    filter 0.24s ease;
}

.narrative-step-group--active {
  z-index: 4;
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
  padding: 0 5px;
  border: 0;
  background: transparent;
  color: rgba(255, 239, 202, 0.94);
  font-size: 11px;
  font-weight: 900;
  line-height: 1;
  text-overflow: ellipsis;
  white-space: nowrap;
  text-align: center;
  text-shadow:
    0 1px 3px rgba(0, 0, 0, 0.88),
    0 0 10px rgba(220, 166, 80, 0.22);
}

.narrative-step-group--active .narrative-step-group__label {
  color: #fff1bc;
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

@keyframes narrativeLoopCurrent {
  from { stroke-dashoffset: 0; }
  to { stroke-dashoffset: -30; }
}

@keyframes narrativeLoopActorIn {
  from { opacity: 0; transform: translate(-50%, calc(-50% + 10px)) scale(0.88); }
  to { opacity: 1; transform: translate(-50%, -50%) scale(1); }
}

@keyframes narrativeLoopCardIn {
  from { opacity: 0; transform: translate(-50%, calc(-50% + 14px)) scale(0.84); }
  to { opacity: 1; transform: translate(-50%, -50%) scale(1); }
}

@keyframes narrativeCounterActorIn {
  from { opacity: 0; transform: translateY(-8px) scale(0.9); }
  to { opacity: 1; transform: translateY(0) scale(1); }
}

@keyframes narrativeCounterChainCardIn {
  from { opacity: 0; transform: translate(-50%, 16px) scale(0.86); }
  to { opacity: 1; transform: translate(-50%, 0) scale(1); }
}

@keyframes narrativeCounterMissIn {
  from { opacity: 0; transform: translate(-50%, -4px) scale(0.9); }
  to { opacity: 1; transform: translate(-50%, 0) scale(1); }
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
