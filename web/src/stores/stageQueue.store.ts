import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import type { Card } from '../types/game'
import type { ActionFlowDTO, ActionFlowEdgeDTO, ActionFlowNodeDTO } from '../network/protocol'

/**
 * Unified battle-stage playback queue.
 *
 * Every visual beat (links, bubbles, chips, banners, narration rows) is an AtomicStep
 * pushed through one FIFO queue and paced by a single clock (`tick`, driven by the
 * caller's requestAnimationFrame loop). New arrivals are appended, never replacing
 * steps that have not been played yet.
 */

export type LinkColor = 'attack' | 'magic' | 'skill' | 'heal' | 'grey'
export type LinkEndStyle = 'blocked' | 'hit' | 'miss' | 'heal'
export type ValueChipColor = 'damage' | 'heal' | 'gain' | 'neutral'

/** Optional per-step pacing override: how long the queue waits before the next step (ms, at 1x). */
interface StepTiming {
  holdMs?: number
}

export type AtomicStep = StepTiming &
  (
    | {
        type: 'link'
        linkId: string
        fromPlayerId: string
        toPlayerId: string
        color: LinkColor
        label: string
        cardImageUrl?: string
      }
    | { type: 'link-animate'; linkId: string; active: boolean }
    | { type: 'link-end'; linkId: string; text: string; style: LinkEndStyle }
    | { type: 'skill-bubble'; playerId: string; skillName: string; effectText: string; duration: number }
    | { type: 'value-chip'; playerId: string; text: string; color: ValueChipColor }
    | { type: 'status-change'; playerId: string; statusKey: string; added: boolean }
    | { type: 'turn-banner'; text: string; duration: number }
    | { type: 'mark-previous' }
    | { type: 'clear-stage'; summary?: string }
    | { type: 'narration-row'; html: string; tag: string; tagColor: LinkColor }
  )

export type AtomicStepType = AtomicStep['type']

export interface LinkState {
  id: string
  fromPlayerId: string
  toPlayerId: string
  color: LinkColor
  label: string
  cardImageUrl?: string
  animating: boolean
  end: { text: string; style: LinkEndStyle } | null
  /** True once the link belongs to the "上一手" layer (rendered dimmed + dashed). */
  previous: boolean
  createdAt: number
}

export interface SkillBubbleState {
  id: number
  playerId: string
  skillName: string
  effectText: string
  startedAt: number
  expiresAt: number
}

export interface ValueChipState {
  id: number
  playerId: string
  text: string
  color: ValueChipColor
  expiresAt: number
}

export interface StatusChangeState {
  id: number
  playerId: string
  statusKey: string
  added: boolean
  expiresAt: number
}

export interface NarrationRow {
  id: number
  /** Pre-escaped HTML. Builders must escape any user-controlled text (see escapeHtml). */
  html: string
  tag: string
  tagColor: LinkColor
}

export interface PreviousActionState {
  rows: NarrationRow[]
  label: string
}

const PREVIOUS_LABEL = '上一手'
const MAX_HISTORY_CHIPS = 20
const AUTO_ACCELERATE_THRESHOLD = 3
const VALUE_CHIP_LIFETIME_MS = 1400
const STATUS_CHANGE_LIFETIME_MS = 1200
const DEFAULT_SKILL_BUBBLE_MS = 1600
const DEFAULT_TURN_BANNER_MS = 1200

/** Default time the queue waits after applying a step, at 1x speed. */
const DEFAULT_HOLD_MS: Record<AtomicStepType, number> = {
  link: 350,
  'link-animate': 450,
  'link-end': 400,
  'skill-bubble': 700,
  'value-chip': 200,
  'status-change': 150,
  'turn-banner': DEFAULT_TURN_BANNER_MS,
  'mark-previous': 200,
  'clear-stage': 0,
  'narration-row': 250,
}

function stepHoldMs(step: AtomicStep): number {
  if (typeof step.holdMs === 'number') return Math.max(0, step.holdMs)
  if (step.type === 'turn-banner') return Math.max(0, step.duration)
  // Turning a link animation off is instant; only the "flowing" beat takes time.
  if (step.type === 'link-animate' && !step.active) return 0
  return DEFAULT_HOLD_MS[step.type]
}

function stripHtml(html: string): string {
  return html.replace(/<[^>]*>/g, '').replace(/\s+/g, ' ').trim()
}

export function escapeHtml(value: string): string {
  return value
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;')
}

export const useStageQueueStore = defineStore('stageQueue', () => {
  const queue = ref<AtomicStep[]>([])
  const activeLinks = ref<Map<string, LinkState>>(new Map())
  const activeSkillBubble = ref<SkillBubbleState | null>(null)
  const valueChips = ref<ValueChipState[]>([])
  const statusChanges = ref<StatusChangeState[]>([])
  const stageRows = ref<NarrationRow[]>([])
  const previousAction = ref<PreviousActionState | null>(null)
  const historyChips = ref<string[]>([])
  const turnBanner = ref<{ text: string } | null>(null)
  const isPlaying = ref(false)
  const speedMultiplier = ref(1)

  let nextReadyAt = 0
  let turnBannerExpiresAt = 0
  let previousSummary = ''
  let seq = 0

  const pendingCount = computed(() => queue.value.length)

  function now(): number {
    return typeof performance !== 'undefined' ? performance.now() : Date.now()
  }

  /** Scale factor applied to every duration: speed setting plus auto-acceleration on backlog. */
  function timeScale(): number {
    const speed = speedMultiplier.value > 0 ? speedMultiplier.value : 1
    const backlog = queue.value.length > AUTO_ACCELERATE_THRESHOLD ? 0.5 : 1
    return backlog / speed
  }

  function pushHistoryChip(text: string) {
    const chip = text.trim()
    if (!chip) return
    historyChips.value = [...historyChips.value, chip].slice(-MAX_HISTORY_CHIPS)
  }

  function summarizeRows(rows: NarrationRow[]): string {
    const first = rows[0]
    return first ? stripHtml(first.html) : ''
  }

  function enqueue(...steps: AtomicStep[]) {
    if (steps.length === 0) return
    queue.value.push(...steps)
    isPlaying.value = true
  }

  /** Demote the existing previous action to a history chip, then move the current stage into "上一手". */
  function archiveCurrentAction(summaryText: string) {
    if (previousAction.value) {
      pushHistoryChip(previousSummary || summarizeRows(previousAction.value.rows))
      previousAction.value = null
      previousSummary = ''
    }
    // Links from the already-demoted layer leave the stage.
    for (const [id, link] of activeLinks.value) {
      if (link.previous) activeLinks.value.delete(id)
    }
    if (stageRows.value.length === 0 && activeLinks.value.size === 0) return
    previousAction.value = { rows: stageRows.value, label: PREVIOUS_LABEL }
    previousSummary = summaryText.trim() || summarizeRows(stageRows.value)
    stageRows.value = []
    for (const link of activeLinks.value.values()) {
      link.previous = true
      link.animating = false
    }
  }

  function clearStage(summary?: string) {
    if (previousAction.value) {
      pushHistoryChip(previousSummary || summarizeRows(previousAction.value.rows))
    }
    if (stageRows.value.length > 0) {
      pushHistoryChip(summary || summarizeRows(stageRows.value))
    }
    previousAction.value = null
    previousSummary = ''
    stageRows.value = []
    activeLinks.value.clear()
    activeSkillBubble.value = null
    valueChips.value = []
    statusChanges.value = []
  }

  function applyStep(step: AtomicStep, at: number, scale: number) {
    switch (step.type) {
      case 'link':
        activeLinks.value.set(step.linkId, {
          id: step.linkId,
          fromPlayerId: step.fromPlayerId,
          toPlayerId: step.toPlayerId,
          color: step.color,
          label: step.label,
          cardImageUrl: step.cardImageUrl,
          animating: false,
          end: null,
          previous: false,
          createdAt: at,
        })
        break
      case 'link-animate': {
        const link = activeLinks.value.get(step.linkId)
        if (link) link.animating = step.active
        break
      }
      case 'link-end': {
        const link = activeLinks.value.get(step.linkId)
        if (link) {
          link.end = { text: step.text, style: step.style }
          link.animating = false
        }
        break
      }
      case 'skill-bubble': {
        const duration = step.duration > 0 ? step.duration : DEFAULT_SKILL_BUBBLE_MS
        activeSkillBubble.value = {
          id: ++seq,
          playerId: step.playerId,
          skillName: step.skillName,
          effectText: step.effectText,
          startedAt: at,
          expiresAt: at + duration * scale,
        }
        break
      }
      case 'value-chip':
        valueChips.value = [
          ...valueChips.value,
          { id: ++seq, playerId: step.playerId, text: step.text, color: step.color, expiresAt: at + VALUE_CHIP_LIFETIME_MS * scale },
        ]
        break
      case 'status-change':
        statusChanges.value = [
          ...statusChanges.value,
          {
            id: ++seq,
            playerId: step.playerId,
            statusKey: step.statusKey,
            added: step.added,
            expiresAt: at + STATUS_CHANGE_LIFETIME_MS * scale,
          },
        ]
        break
      case 'turn-banner':
        turnBanner.value = { text: step.text }
        turnBannerExpiresAt = at + (step.duration > 0 ? step.duration : DEFAULT_TURN_BANNER_MS) * scale
        break
      case 'mark-previous':
        archiveCurrentAction('')
        break
      case 'clear-stage':
        clearStage(step.summary)
        break
      case 'narration-row':
        stageRows.value = [...stageRows.value, { id: ++seq, html: step.html, tag: step.tag, tagColor: step.tagColor }]
        break
    }
  }

  function expireTransient(at: number) {
    if (activeSkillBubble.value && at >= activeSkillBubble.value.expiresAt) {
      activeSkillBubble.value = null
    }
    if (turnBanner.value && at >= turnBannerExpiresAt) {
      turnBanner.value = null
    }
    if (valueChips.value.some((chip) => at >= chip.expiresAt)) {
      valueChips.value = valueChips.value.filter((chip) => at < chip.expiresAt)
    }
    if (statusChanges.value.some((change) => at >= change.expiresAt)) {
      statusChanges.value = statusChanges.value.filter((change) => at < change.expiresAt)
    }
  }

  /**
   * Advance playback. Call once per animation frame with the rAF timestamp.
   * Applies at most one step per call, then waits for that step's (scaled) hold time.
   */
  function tick(timestamp?: number) {
    const at = timestamp ?? now()
    expireTransient(at)
    if (at < nextReadyAt) return

    const step = queue.value[0]
    if (!step) {
      isPlaying.value = false
      return
    }
    // Scale is computed before dequeue so a backlog of >3 pending steps accelerates this one too.
    const scale = timeScale()
    queue.value.shift()
    applyStep(step, at, scale)
    nextReadyAt = at + stepHoldMs(step) * scale
    isPlaying.value = true
  }

  function setSpeed(multiplier: number) {
    if (!Number.isFinite(multiplier) || multiplier <= 0) return
    speedMultiplier.value = multiplier
  }

  function reset() {
    queue.value = []
    activeLinks.value.clear()
    activeSkillBubble.value = null
    valueChips.value = []
    statusChanges.value = []
    stageRows.value = []
    previousAction.value = null
    historyChips.value = []
    turnBanner.value = null
    isPlaying.value = false
    nextReadyAt = 0
    turnBannerExpiresAt = 0
    previousSummary = ''
  }

  return {
    queue,
    activeLinks,
    activeSkillBubble,
    valueChips,
    statusChanges,
    stageRows,
    previousAction,
    historyChips,
    turnBanner,
    isPlaying,
    speedMultiplier,
    pendingCount,
    enqueue,
    tick,
    archiveCurrentAction,
    reset,
    setSpeed,
  }
})

// ---------------------------------------------------------------------------
// ActionFlowDTO -> AtomicStep[]
// ---------------------------------------------------------------------------

const PHASE_VERB: Record<string, string> = {
  attack: '攻击',
  counter: '应战',
  defend: '防御',
  shield: '圣盾',
  magic: '法术',
  skill: '技能',
  effect: '效果',
  take: '承受',
  extra_action: '额外行动',
  special: '特殊',
}

function normalizePhase(phase: string | undefined): string {
  const value = (phase || '').trim().toLowerCase()
  if (value === 'respond') return 'counter'
  if (value === 'defense') return 'defend'
  if (value === 'damage') return 'take'
  if (value === 'field_effect') return 'effect'
  return value
}

export function linkColorForPhase(phase: string | undefined, damageType?: string): LinkColor {
  if ((damageType || '').toLowerCase() === 'heal') return 'heal'
  switch (normalizePhase(phase)) {
    case 'attack':
    case 'counter':
    case 'defend':
    case 'shield':
    case 'take':
      return 'attack'
    case 'magic':
      return 'magic'
    case 'skill':
    case 'effect':
      return 'skill'
    default:
      return 'grey'
  }
}

function phaseVerb(phase: string | undefined): string {
  return PHASE_VERB[normalizePhase(phase)] || '行动'
}

/** Role name is the primary label; player name is secondary: "封印师（测试玩家5）". */
export function playerLabel(id: string | undefined, playerNameFn: (id: string) => string, roleNameFn: (id: string) => string): string {
  if (!id) return ''
  const name = playerNameFn(id) || id
  const role = roleNameFn(id)
  if (!role || role === name) return name
  return `${role}（${name}）`
}

function playerHtml(id: string | undefined, playerNameFn: (id: string) => string, roleNameFn: (id: string) => string): string {
  const label = playerLabel(id, playerNameFn, roleNameFn)
  return label ? `<b>${escapeHtml(label)}</b>` : ''
}

function firstCardName(cards: Card[] | undefined): string {
  return cards?.find((card) => card?.name)?.name || ''
}

export function edgeEnd(edge: ActionFlowEdgeDTO): { text: string; style: LinkEndStyle } | null {
  const outcome = (edge.outcome || '').toLowerCase()
  const phase = normalizePhase(edge.phase)
  const damage = edge.damage ?? 0
  if (outcome === 'miss' || edge.label === '未命中') return { text: '未命中', style: 'miss' }
  if ((edge.damage_type || '').toLowerCase() === 'heal') return { text: damage > 0 ? `+${damage}` : '治疗', style: 'heal' }
  if (phase === 'defend' || phase === 'shield') return { text: phase === 'shield' ? '圣盾抵挡' : '抵挡', style: 'blocked' }
  if (outcome === 'hit' || damage > 0) return { text: damage > 0 ? `${damage}伤害` : '命中', style: 'hit' }
  return null
}

export function nodeSteps(
  node: ActionFlowNodeDTO,
  playerNameFn: (id: string) => string,
  roleNameFn: (id: string) => string,
): AtomicStep[] {
  const kind = (node.kind || '').toLowerCase()
  const label = node.skill_name || node.label || node.effect_text || ''
  const actor = playerHtml(node.actor_user_id, playerNameFn, roleNameFn)
  const targets = (node.target_user_ids || [])
    .map((id) => playerHtml(id, playerNameFn, roleNameFn))
    .filter(Boolean)
    .join('、')
  const steps: AtomicStep[] = []

  if (kind === 'skill') {
    const skillName = node.skill_name || node.label || '技能'
    if (node.actor_user_id) {
      steps.push({
        type: 'skill-bubble',
        playerId: node.actor_user_id,
        skillName,
        effectText: node.effect_text || '',
        duration: DEFAULT_SKILL_BUBBLE_MS,
      })
    }
    const effect = node.effect_text ? `：${escapeHtml(node.effect_text)}` : ''
    const toward = targets ? ` → ${targets}` : ''
    steps.push({
      type: 'narration-row',
      html: `${actor} 发动【${escapeHtml(skillName)}】${toward}${effect}`.trim(),
      tag: '技能',
      tagColor: 'skill',
    })
  } else if ((kind === 'effect' || kind === 'extra_action') && label) {
    const tag = kind === 'extra_action' ? '额外行动' : '效果'
    const toward = targets ? ` → ${targets}` : ''
    steps.push({
      type: 'narration-row',
      html: `${actor ? `${actor} ` : ''}${escapeHtml(label)}${toward}`,
      tag,
      tagColor: kind === 'extra_action' ? 'grey' : 'skill',
    })
  }
  // 'card' / 'action' nodes are represented by their edges; nothing extra to show.

  steps.push(...nodeDamageSteps(node))
  return steps
}

/** Damage / heal chips for a node's targets (empty when the node carries no damage). */
export function nodeDamageSteps(node: ActionFlowNodeDTO): AtomicStep[] {
  const damage = node.damage ?? 0
  if (damage <= 0) return []
  const heal = (node.damage_type || '').toLowerCase() === 'heal'
  return (node.target_user_ids || []).map(
    (targetId): AtomicStep => ({
      type: 'value-chip',
      playerId: targetId,
      text: heal ? `+${damage}` : `-${damage}`,
      color: heal ? 'heal' : 'damage',
    }),
  )
}

export function flowLinkId(flowId: string, edgeId: string): string {
  return `${flowId}:${edgeId}`
}

/** Steps that introduce an edge: link + flowing animation + narration row. */
export function edgeStartSteps(
  flowId: string,
  edge: ActionFlowEdgeDTO,
  playerNameFn: (id: string) => string,
  roleNameFn: (id: string) => string,
  cardImageFn?: (card: Card) => string | undefined,
): AtomicStep[] {
  if (isSelfResolveEdge(edge)) {
    // 承受结算只记一行叙事，不画线；结果（伤害数值）由 edgeResolveSteps 标到目标身上
    const who = playerHtml(edge.to_user_id || edge.from_user_id, playerNameFn, roleNameFn)
    return who ? [{ type: 'narration-row', html: `${who} ${phaseVerb(edge.phase)}`, tag: phaseVerb(edge.phase), tagColor: 'grey' }] : []
  }
  const linkId = flowLinkId(flowId, edge.id)
  const color = linkColorForPhase(edge.phase, edge.damage_type)
  const verb = phaseVerb(edge.phase)
  const cardName = firstCardName(edge.cards)
  const firstCard = edge.cards?.[0]
  const label = edge.label || (cardName ? `${verb}·${cardName}` : verb)

  const link: AtomicStep = {
    type: 'link',
    linkId,
    fromPlayerId: edge.from_user_id,
    toPlayerId: edge.to_user_id,
    color,
    label,
  }
  const imageUrl = firstCard && cardImageFn ? cardImageFn(firstCard) : undefined
  if (imageUrl) link.cardImageUrl = imageUrl

  const from = playerHtml(edge.from_user_id, playerNameFn, roleNameFn)
  const to = playerHtml(edge.to_user_id, playerNameFn, roleNameFn)
  const cardHtml = cardName ? ` 【${escapeHtml(cardName)}】` : ''
  return [
    link,
    { type: 'link-animate', linkId, active: true },
    { type: 'narration-row', html: `${from} → ${to}${cardHtml}`, tag: verb, tagColor: color },
  ]
}

/** Steps that resolve an edge: stop flowing, result marker, damage/heal chip. */
/** 承受/自身结算类边（from === to 或 phase=take）：不单独画线，结果标到目标身上。 */
export function isSelfResolveEdge(edge: ActionFlowEdgeDTO): boolean {
  return normalizePhase(edge.phase) === 'take' || (!!edge.from_user_id && edge.from_user_id === edge.to_user_id)
}

/**
 * Steps that resolve an edge: stop flowing + result marker at the link end.
 * Damage is shown once, as the end marker; a separate chip is only used when there is no
 * link to anchor it on (self-resolve edges).
 */
export function edgeResolveSteps(flowId: string, edge: ActionFlowEdgeDTO): AtomicStep[] {
  const damage = edge.damage ?? 0
  const heal = (edge.damage_type || '').toLowerCase() === 'heal'
  if (isSelfResolveEdge(edge)) {
    if (damage > 0 && edge.to_user_id) {
      return [{ type: 'value-chip', playerId: edge.to_user_id, text: heal ? `+${damage}` : `-${damage}`, color: heal ? 'heal' : 'damage' }]
    }
    return []
  }
  const linkId = flowLinkId(flowId, edge.id)
  const steps: AtomicStep[] = [{ type: 'link-animate', linkId, active: false }]
  const end = edgeEnd(edge)
  if (end) steps.push({ type: 'link-end', linkId, text: end.text, style: end.style })
  return steps
}

/** Narration rows from flow logs; used only when edges/nodes produce no narration. */
export function flowLogSteps(flow: ActionFlowDTO, logs: ActionFlowDTO['logs'] = flow.logs): AtomicStep[] {
  const steps: AtomicStep[] = []
  for (const log of [...(logs || [])].sort((a, b) => a.order - b.order)) {
    if (!log.text?.trim()) continue
    steps.push({ type: 'narration-row', html: escapeHtml(log.text), tag: phaseVerb(flow.action_type), tagColor: linkColorForPhase(flow.action_type) })
  }
  return steps
}

/**
 * Convert one ActionFlowDTO into ordered stage steps.
 *
 * Order: mark the previous action, then for each edge (by `order`):
 * link -> flowing animation -> anchored node beats -> result marker -> damage chip.
 * Unanchored nodes play after all edges. Flow logs are only used as a fallback when
 * edges/nodes produce no narration (the backend builds logs from the same edges/nodes).
 *
 * `cardImageFn` is optional: links get `cardImageUrl` only when the caller supplies one.
 * For incremental playback of re-sent flows see network/stageQueueBridge.ts.
 */
export function actionFlowToSteps(
  flow: ActionFlowDTO,
  playerNameFn: (id: string) => string,
  roleNameFn: (id: string) => string,
  cardImageFn?: (card: Card) => string | undefined,
): AtomicStep[] {
  const steps: AtomicStep[] = [{ type: 'mark-previous' }]
  const edges = [...(flow.edges || [])].sort((a, b) => a.order - b.order)
  const nodes = [...(flow.nodes || [])].sort((a, b) => a.order - b.order)
  const edgeIds = new Set(edges.map((edge) => edge.id))
  const anchoredNodes = new Map<string, ActionFlowNodeDTO[]>()
  const freeNodes: ActionFlowNodeDTO[] = []
  for (const node of nodes) {
    if (node.anchor_edge_id && edgeIds.has(node.anchor_edge_id)) {
      const list = anchoredNodes.get(node.anchor_edge_id) || []
      list.push(node)
      anchoredNodes.set(node.anchor_edge_id, list)
    } else {
      freeNodes.push(node)
    }
  }

  for (const edge of edges) {
    steps.push(...edgeStartSteps(flow.flow_id, edge, playerNameFn, roleNameFn, cardImageFn))
    for (const node of anchoredNodes.get(edge.id) || []) {
      steps.push(...nodeSteps(node, playerNameFn, roleNameFn))
    }
    steps.push(...edgeResolveSteps(flow.flow_id, edge))
  }
  for (const node of freeNodes) {
    steps.push(...nodeSteps(node, playerNameFn, roleNameFn))
  }
  if (!steps.some((step) => step.type === 'narration-row')) {
    steps.push(...flowLogSteps(flow))
  }

  // Only the leading mark-previous: nothing to play.
  return steps.length > 1 ? steps : []
}
