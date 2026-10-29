import { skillResult, skillResultSummary } from './skillResult'
import { defineStore } from 'pinia'
import { computed, ref, toRaw } from 'vue'
import type { Card } from '../types/game'
import type { ActionFlowDTO, ActionFlowEdgeDTO, ActionFlowNodeDTO, TimelineDelta } from '../network/protocol'
import { SKILL_ANIMATION_MAP, type SkillEffectTemplate, type SkillOpening } from '../constants/skillAnimationMap'
import { COVER_MARKS, STATUS_MARKS, TOKEN_MARKS } from '../constants/battleMarks'
import { formLabel } from '../constants/forms'

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
        /** Element of the played card; drives the arc colour on the stage. */
        element?: string
      }
    | { type: 'link-animate'; linkId: string; active: boolean }
    | { type: 'link-end'; linkId: string; text: string; style: LinkEndStyle }
    | {
        type: 'skill-bubble'
        playerId: string
        skillName: string
        effectText: string
        duration: number
        /** 'active' = sigil + plaque + beams; 'trigger' = small capsule only. Defaults to 'active'. */
        opening?: SkillOpening
        skillId?: string
        element?: string
        effects?: SkillEffectTemplate[]
        targetIds?: string[]
        outcome?: string
        causeId?: string
        /** 招牌大招（模板 6 全局型）：开场改为横贯中央的全幅横幅。 */
        signature?: boolean
      }
    | { type: 'value-chip'; playerId: string; text: string; color: ValueChipColor }
    | {
        type: 'status-change'
        playerId: string
        statusKey: string
        added: boolean
        fromPlayerId?: string
        /** 'status' = 场上状态（默认）；'cover' = 盖牌；'token' = 指示物。 */
        markKind?: MarkKind
        /** 指示物增减量（显示为 +N / -N）。 */
        amount?: number
      }
    | { type: 'turn-banner'; text: string; duration: number }
    | { type: 'mark-previous' }
    | { type: 'clear-stage'; summary?: string }
    | { type: 'narration-row'; html: string; tag: string; tagColor: LinkColor; chain?: NarrationChain; skill?: NarrationSkill }
    /** 同一技能因果的结果后到（如茧在结算后才放置）：就地更新已有技能行。 */
    | { type: 'skill-result'; causeId: string; outcome: string; cost: string }
    /** 自身形态型（模板 5）：座位包裹光焰 + 形态名横幅。form 为空 = 脱离形态。 */
    | { type: 'form-change'; playerId: string; form: string; label: string; element?: string }
    /** 资源型（模板 4）：count 张牌背从中央牌堆飞入座位（盖牌放置：茧、充能……）。 */
    | { type: 'deck-to-seat'; playerId: string; count: number; coverKey: string }
    /**
     * 资源型（模板 4）：宝石 / 水晶在座位与阵营战绩区之间飞行。
     * toCamp = 座位 → 阵营栏（阵营获得星石）；否则阵营栏 → 座位（提炼等）。
     */
    | { type: 'stone-flight'; playerId: string; camp: string; stone: 'gem' | 'crystal'; count: number; toCamp: boolean }
  )

/** Structured form of a skill row, rendered as a skill node of the combat chain. */
export interface NarrationSkill {
  skillId: string
  skillName: string
  actorId: string
  targetIds: string[]
  opening: SkillOpening
  element?: string
  effects: SkillEffectTemplate[]
  /** Link drawn for this skill (actor → first target), whose result the node shows. */
  linkId?: string
  /** 结算结果短语（"蛹 +1 · 茧 ×4"、"伤害 +1"、"代为应战 → X"）；空串 = 尚无可见结果。 */
  outcome?: string
  /** 发动代价（"宝石 1"），与结果分开展示。 */
  cost?: string
  /** 完整技能描述，仅用于悬停提示。 */
  effectText?: string
  /** 同一因果的节点 ID：后续结果到达时就地更新同一节点。 */
  causeId?: string
}

/** Structured form of an edge row, rendered as one node of the combat chain. */
export interface NarrationChain {
  linkId: string
  fromPlayerId: string
  toPlayerId: string
  verb: string
  card?: Pick<Card, 'name' | 'element' | 'type'>
}

export type AtomicStepType = AtomicStep['type']

export interface LinkState {
  id: string
  fromPlayerId: string
  toPlayerId: string
  color: LinkColor
  label: string
  cardImageUrl?: string
  element?: string
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
  skillId: string
  element?: string
  effects: SkillEffectTemplate[]
  targetIds: string[]
  /** 技能因果名下的结算结果（"蛹 +1 · 茧 ×4"）；结果后到时由 skill-result 更新。 */
  outcome: string
  causeId: string
  signature: boolean
}

/** Small "技能名 · 触发" capsule beside a seat (passive / response skills). */
export interface SkillCapsuleState {
  id: number
  playerId: string
  skillName: string
  element?: string
  expiresAt: number
}

export interface ValueChipState {
  id: number
  playerId: string
  text: string
  color: ValueChipColor
  expiresAt: number
}

export type MarkKind = 'status' | 'cover' | 'token'

export interface StatusChangeState {
  id: number
  playerId: string
  statusKey: string
  added: boolean
  markKind: MarkKind
  amount?: number
  /** Caster the status flies from; absent = the icon just pops on the target. */
  fromPlayerId?: string
  expiresAt: number
}

export interface FormChangeState {
  id: number
  playerId: string
  form: string
  label: string
  element?: string
  expiresAt: number
}

export interface DeckFlightState {
  id: number
  playerId: string
  count: number
  coverKey: string
  expiresAt: number
}

export interface StoneFlightState {
  id: number
  playerId: string
  camp: string
  stone: 'gem' | 'crystal'
  count: number
  toCamp: boolean
  expiresAt: number
}

export interface NarrationRow {
  id: number
  /** Pre-escaped HTML. Builders must escape any user-controlled text (see escapeHtml). */
  html: string
  tag: string
  tagColor: LinkColor
  chain?: NarrationChain
  skill?: NarrationSkill
}

export interface PreviousActionState {
  rows: NarrationRow[]
  label: string
}

const PREVIOUS_LABEL = '上一手'
const MAX_HISTORY_CHIPS = 20
const AUTO_ACCELERATE_THRESHOLD = 3
/** 切回标签页时，入队超过这么久的积压视为已过时，直接结算不再播放。 */
const STALE_BACKLOG_MS = 3000
const VALUE_CHIP_LIFETIME_MS = 2000
const STATUS_CHANGE_LIFETIME_MS = 1600
const DEFAULT_SKILL_BUBBLE_MS = 1900
/** 招牌大招的全幅横幅停留更久，留出观看时间。 */
const SIGNATURE_SKILL_MS = 2600
const SKILL_CAPSULE_LIFETIME_MS = 1600
/** Trigger capsules don't hold the queue as long as a full active-skill opening. */
const TRIGGER_HOLD_MS = 350
const DEFAULT_TURN_BANNER_MS = 1200
const FORM_CHANGE_LIFETIME_MS = 1800
const DECK_FLIGHT_LIFETIME_MS = 1300
/** 一次最多飞几张牌背（茧 ×4 全飞，更多则截断，避免刷屏）。 */
const MAX_DECK_FLIGHT_CARDS = 5

/** Default time the queue waits after applying a step, at 1x speed. */
const DEFAULT_HOLD_MS: Record<AtomicStepType, number> = {
  link: 350,
  'link-animate': 450,
  'link-end': 400,
  'skill-bubble': 700,
  'value-chip': 450,
  'status-change': 300,
  'turn-banner': DEFAULT_TURN_BANNER_MS,
  'skill-result': 0,
  'form-change': 650,
  'deck-to-seat': 500,
  'stone-flight': 400,
  'mark-previous': 200,
  'clear-stage': 0,
  'narration-row': 250,
}

function stepHoldMs(step: AtomicStep): number {
  if (typeof step.holdMs === 'number') return Math.max(0, step.holdMs)
  if (step.type === 'turn-banner') return Math.max(0, step.duration)
  // Turning a link animation off is instant; only the "flowing" beat takes time.
  if (step.type === 'link-animate' && !step.active) return 0
  if (step.type === 'skill-bubble' && step.opening === 'trigger') return TRIGGER_HOLD_MS
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
  const skillCapsules = ref<SkillCapsuleState[]>([])
  const valueChips = ref<ValueChipState[]>([])
  const statusChanges = ref<StatusChangeState[]>([])
  const formChanges = ref<FormChangeState[]>([])
  const deckFlights = ref<DeckFlightState[]>([])
  const stoneFlights = ref<StoneFlightState[]>([])
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

  /** 每个步骤的入队时间（按原始对象记录，出队后自动回收），用于判断切回标签页时哪些已过时。 */
  const arrivedAt = new WeakMap<AtomicStep, number>()

  function enqueue(...steps: AtomicStep[]) {
    if (steps.length === 0) return
    const t = now()
    for (const step of steps) arrivedAt.set(toRaw(step), t)
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
    skillCapsules.value = []
    valueChips.value = []
    statusChanges.value = []
    formChanges.value = []
    deckFlights.value = []
    stoneFlights.value = []
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
          element: step.element,
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
        if (step.opening === 'trigger') {
          skillCapsules.value = [
            ...skillCapsules.value,
            { id: ++seq, playerId: step.playerId, skillName: step.skillName, element: step.element, expiresAt: at + SKILL_CAPSULE_LIFETIME_MS * scale },
          ]
          break
        }
        const duration = step.duration > 0 ? step.duration : DEFAULT_SKILL_BUBBLE_MS
        activeSkillBubble.value = {
          id: ++seq,
          playerId: step.playerId,
          skillName: step.skillName,
          effectText: step.effectText,
          startedAt: at,
          expiresAt: at + duration * scale,
          skillId: step.skillId || '',
          element: step.element,
          effects: step.effects || [],
          targetIds: step.targetIds || [],
          outcome: step.outcome || '',
          causeId: step.causeId || '',
          signature: !!step.signature,
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
            markKind: step.markKind ?? 'status',
            amount: step.amount,
            fromPlayerId: step.fromPlayerId,
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
      case 'skill-result': {
        if (activeSkillBubble.value?.causeId === step.causeId) {
          activeSkillBubble.value = { ...activeSkillBubble.value, outcome: step.outcome }
        }
        const index = stageRows.value.findIndex((row) => row.skill?.causeId === step.causeId)
        if (index < 0) break
        const row = stageRows.value[index]!
        const next = stageRows.value.slice()
        next[index] = { ...row, skill: { ...row.skill!, outcome: step.outcome, cost: step.cost } }
        stageRows.value = next
        break
      }
      case 'form-change':
        formChanges.value = [
          ...formChanges.value.filter((change) => change.playerId !== step.playerId),
          {
            id: ++seq,
            playerId: step.playerId,
            form: step.form,
            label: step.label,
            element: step.element,
            expiresAt: at + FORM_CHANGE_LIFETIME_MS * scale,
          },
        ]
        break
      case 'deck-to-seat':
        deckFlights.value = [
          ...deckFlights.value,
          {
            id: ++seq,
            playerId: step.playerId,
            count: Math.max(1, Math.min(step.count, MAX_DECK_FLIGHT_CARDS)),
            coverKey: step.coverKey,
            expiresAt: at + DECK_FLIGHT_LIFETIME_MS * scale,
          },
        ]
        break
      case 'stone-flight':
        stoneFlights.value = [
          ...stoneFlights.value,
          {
            id: ++seq,
            playerId: step.playerId,
            camp: step.camp,
            stone: step.stone,
            count: Math.max(1, Math.min(step.count, MAX_DECK_FLIGHT_CARDS)),
            toCamp: step.toCamp,
            expiresAt: at + DECK_FLIGHT_LIFETIME_MS * scale,
          },
        ]
        break
      case 'narration-row':
        stageRows.value = [
          ...stageRows.value,
          { id: ++seq, html: step.html, tag: step.tag, tagColor: step.tagColor, chain: step.chain, skill: step.skill },
        ]
        break
    }
  }

  function expireTransient(at: number) {
    if (activeSkillBubble.value && at >= activeSkillBubble.value.expiresAt) {
      activeSkillBubble.value = null
    }
    if (skillCapsules.value.some((capsule) => at >= capsule.expiresAt)) {
      skillCapsules.value = skillCapsules.value.filter((capsule) => at < capsule.expiresAt)
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
    if (formChanges.value.some((change) => at >= change.expiresAt)) {
      formChanges.value = formChanges.value.filter((change) => at < change.expiresAt)
    }
    if (deckFlights.value.some((flight) => at >= flight.expiresAt)) {
      deckFlights.value = deckFlights.value.filter((flight) => at < flight.expiresAt)
    }
    if (stoneFlights.value.some((flight) => at >= flight.expiresAt)) {
      stoneFlights.value = stoneFlights.value.filter((flight) => at < flight.expiresAt)
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

  /**
   * 跳过积压：标签页在后台时 rAF 暂停而 WebSocket 照常入队，切回来会把离开期间的动作逐个补播。
   * 已过时（入队超过 staleMs）的步骤与最近一个动作之前的步骤立即结算：留场 / 连线 / 历史照常更新，
   * 瞬时特效不播放。只有刚到达的最近一个动作留在队列里正常播放。
   */
  function fastForward(at: number = now(), staleMs: number = STALE_BACKLOG_MS) {
    const items = queue.value
    let lastMark = -1
    for (let i = items.length - 1; i >= 0; i -= 1) {
      if (items[i]!.type === 'mark-previous') {
        lastMark = i
        break
      }
    }
    let firstFresh = items.length
    for (let i = 0; i < items.length; i += 1) {
      const arrived = arrivedAt.get(toRaw(items[i]!))
      if (arrived === undefined || at - arrived < staleMs) {
        firstFresh = i
        break
      }
    }
    const keepFrom = Math.max(lastMark, firstFresh)
    if (keepFrom <= 0) return
    const skipped = items.slice(0, keepFrom)
    queue.value = items.slice(keepFrom)
    // scale = 0：瞬时特效的过期时间就是 at，下面立即清掉，不会闪出来
    for (const step of skipped) applyStep(step, at, 0)
    expireTransient(at)
    nextReadyAt = at
    if (queue.value.length === 0) isPlaying.value = false
  }

  function setSpeed(multiplier: number) {
    if (!Number.isFinite(multiplier) || multiplier <= 0) return
    speedMultiplier.value = multiplier
  }

  function reset() {
    queue.value = []
    activeLinks.value.clear()
    activeSkillBubble.value = null
    skillCapsules.value = []
    valueChips.value = []
    statusChanges.value = []
    formChanges.value = []
    deckFlights.value = []
    stoneFlights.value = []
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
    skillCapsules,
    valueChips,
    statusChanges,
    formChanges,
    deckFlights,
    stoneFlights,
    stageRows,
    previousAction,
    historyChips,
    turnBanner,
    isPlaying,
    speedMultiplier,
    pendingCount,
    enqueue,
    tick,
    fastForward,
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
    const skillId = node.skill_id || ''
    const spec = skillId ? SKILL_ANIMATION_MAP[skillId] : undefined
    const opening: SkillOpening = spec?.opening ?? 'active'
    const targetIds = (node.target_user_ids || []).filter(Boolean)
    if (node.actor_user_id) {
      steps.push({
        type: 'skill-bubble',
        playerId: node.actor_user_id,
        skillName,
        effectText: node.effect_text || '',
        duration: DEFAULT_SKILL_BUBBLE_MS,
        opening,
        skillId,
        element: spec?.element,
        effects: spec?.effects ?? [],
        targetIds,
        outcome: nodeSkillResult(node, playerNameFn, roleNameFn).outcome,
        causeId: node.cause_id || '',
        ...(spec?.signature ? { signature: true, duration: SIGNATURE_SKILL_MS } : {}),
      })
    }
    const result = skillResult(node, (id) => playerLabel(id, playerNameFn, roleNameFn))
    const outcome = skillResultSummary(result)
    const cost = result.costs.map((item) => item.text).join(' · ')
    const toward = targets ? ` → ${targets}` : ''
    const detail = outcome ? `：${escapeHtml(outcome)}` : node.effect_text ? `：${escapeHtml(node.effect_text)}` : ''
    steps.push({
      type: 'narration-row',
      html: `${actor} 发动【${escapeHtml(skillName)}】${toward}${detail}`.trim(),
      tag: '技能',
      tagColor: 'skill',
      skill: {
        skillId,
        skillName,
        actorId: node.actor_user_id || '',
        targetIds,
        opening,
        element: spec?.element,
        effects: spec?.effects ?? [],
        outcome,
        cost,
        effectText: node.effect_text || '',
        causeId: node.cause_id || '',
      },
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

/** 技能节点的结果文案（供结果后到时就地更新）。 */
export function nodeSkillResult(
  node: ActionFlowNodeDTO,
  playerNameFn: (id: string) => string,
  roleNameFn: (id: string) => string,
): { outcome: string; cost: string } {
  const result = skillResult(node, (id) => playerLabel(id, playerNameFn, roleNameFn))
  return { outcome: skillResultSummary(result), cost: result.costs.map((item) => item.text).join(' · ') }
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
/**
 * Public state deltas → stage beats: +N 治疗 lands as a heal number; a status, cover card or
 * token change becomes a mark that flies from the acting player into the target's seat
 * (or fades out of it when removed). Gems, hand counts etc. are left to the board itself.
 */
const TEAM_STONE: Record<string, 'gem' | 'crystal'> = { team_gem: 'gem', team_crystal: 'crystal' }
const PLAYER_STONE: Record<string, 'gem' | 'crystal'> = { player_gem: 'gem', player_crystal: 'crystal' }

/**
 * 阵营战绩区的宝石 / 水晶变化 → 飞行动画。
 *   - 阵营增加：从行动者座位飞向该阵营栏（攻击命中、地下法则……）。
 *   - 阵营减少且同批有玩家同种星石增加：从阵营栏飞向该玩家（提炼）。
 * 只是阵营减少（技能消耗阵营星石）不飞，避免和座位上的消耗提示重复。
 */
function stoneFlightSteps(deltas: TimelineDelta[], actorId?: string): AtomicStep[] {
  const steps: AtomicStep[] = []
  for (const delta of deltas) {
    const stone = TEAM_STONE[delta.type]
    const value = delta.value ?? 0
    if (!stone || !delta.camp || value === 0) continue
    if (value > 0) {
      if (actorId) steps.push({ type: 'stone-flight', playerId: actorId, camp: delta.camp, stone, count: value, toCamp: true })
      continue
    }
    const receiver = deltas.find((d) => PLAYER_STONE[d.type] === stone && (d.value ?? 0) > 0 && d.camp === delta.camp && d.target_user_id)
    if (receiver?.target_user_id) {
      steps.push({ type: 'stone-flight', playerId: receiver.target_user_id, camp: delta.camp, stone, count: -value, toCamp: false })
    }
  }
  return steps
}

/** @param element 引起这批变化的技能元素（来自因果技能的动画映射），给形态光焰着色。 */
export function deltaSteps(deltas: TimelineDelta[], actorId?: string, element?: string): AtomicStep[] {
  const steps: AtomicStep[] = []
  // 模板 4（资源型）：同一批里从牌库放到某座位的盖牌（牌库数量同时减少）合成一次“牌背飞入”。
  const deckDrawn = deltas.some((delta) => delta.type === 'deck_count' && (delta.value ?? 0) < 0)
  const coverPlaced = new Map<string, { count: number; coverKey: string }>()
  steps.push(...stoneFlightSteps(deltas, actorId))
  for (const delta of deltas) {
    const target = delta.target_user_id || ''
    if (!target) continue
    const value = delta.value ?? 0
    if (delta.type === 'heal' && value > 0) {
      steps.push({ type: 'value-chip', playerId: target, text: `+${value}`, color: 'heal' })
      continue
    }
    // 模板 5（自身形态型）：进入 / 脱离形态。
    if (delta.type === 'form') {
      const form = delta.after_text || ''
      steps.push({ type: 'form-change', playerId: target, form, label: form ? formLabel(form) || '新形态' : formLabel(delta.before_text) || '形态', ...(element ? { element } : {}) })
      continue
    }
    if (deckDrawn && delta.type === 'field_card_added' && delta.field_card?.mode === 'Cover' && COVER_MARKS[delta.field_card.effect || '']) {
      const entry = coverPlaced.get(target) ?? { count: 0, coverKey: delta.field_card.effect || '' }
      entry.count += 1
      coverPlaced.set(target, entry)
    }
    let statusKey = ''
    let added = false
    let markKind: MarkKind = 'status'
    let amount: number | undefined
    if (delta.type === 'field_card_added' || delta.type === 'field_card_removed') {
      const card = delta.field_card
      statusKey = card?.effect || ''
      added = delta.type === 'field_card_added'
      markKind = card?.mode === 'Cover' ? 'cover' : 'status'
      if (markKind === 'status' && !STATUS_MARKS[statusKey]) continue
      if (markKind === 'cover' && !COVER_MARKS[statusKey]) continue
    } else if (delta.type === 'status' && delta.field?.startsWith('status:')) {
      statusKey = delta.field.slice('status:'.length)
      added = value > 0
      if (!STATUS_MARKS[statusKey]) continue
    } else if (delta.type === 'token' && delta.field && value !== 0) {
      // field = "token:<key>" or "indicator:<key>"
      statusKey = delta.field.slice(delta.field.indexOf(':') + 1)
      if (!TOKEN_MARKS[statusKey]) continue
      markKind = 'token'
      added = value > 0
      amount = value
    }
    if (!statusKey) continue
    steps.push({
      type: 'status-change',
      playerId: target,
      statusKey,
      added,
      ...(markKind !== 'status' ? { markKind } : {}),
      ...(amount !== undefined ? { amount } : {}),
      ...(added && actorId && actorId !== target ? { fromPlayerId: actorId } : {}),
    })
  }
  // 牌背先飞入、徽章随后落定：把 deck-to-seat 放在同座位第一个盖牌徽章之前。
  for (const [playerId, { count, coverKey }] of coverPlaced) {
    const flight: AtomicStep = { type: 'deck-to-seat', playerId, count, coverKey }
    const index = steps.findIndex((step) => step.type === 'status-change' && step.playerId === playerId && step.markKind === 'cover')
    if (index < 0) steps.push(flight)
    else steps.splice(index, 0, flight)
  }
  return steps
}

export function edgeStartSteps(
  flowId: string,
  edge: ActionFlowEdgeDTO,
  playerNameFn: (id: string) => string,
  roleNameFn: (id: string) => string,
  cardImageFn?: (card: Card) => string | undefined,
): AtomicStep[] {
  if (isSelfResolveEdge(edge)) {
    // A self skill edge is only the anchor of the skill node, which narrates itself.
    if (normalizePhase(edge.phase) === 'skill') return []
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
  if (firstCard?.element) link.element = firstCard.element

  const from = playerHtml(edge.from_user_id, playerNameFn, roleNameFn)
  const to = playerHtml(edge.to_user_id, playerNameFn, roleNameFn)
  const cardHtml = cardName ? ` 【${escapeHtml(cardName)}】` : ''
  // 技能连线只是技能节点的锚点：画线，但不在战斗链里另起一张“技能”牌（技能节点自己会叙述）。
  if (normalizePhase(edge.phase) === 'skill' && !firstCard) {
    return [link, { type: 'link-animate', linkId, active: true }]
  }
  return [
    link,
    { type: 'link-animate', linkId, active: true },
    {
      type: 'narration-row',
      html: `${from} → ${to}${cardHtml}`,
      tag: verb,
      tagColor: color,
      chain: {
        linkId,
        fromPlayerId: edge.from_user_id,
        toPlayerId: edge.to_user_id,
        verb,
        card: firstCard ? { name: firstCard.name, element: firstCard.element, type: firstCard.type } : undefined,
      },
    },
  ]
}

/** Steps that resolve an edge: stop flowing, result marker, damage/heal chip. */
/** 承受/自身结算类边（from === to 或 phase=take）：不单独画线，结果标到目标身上。 */
export function isSelfResolveEdge(edge: ActionFlowEdgeDTO): boolean {
  return normalizePhase(edge.phase) === 'take' || (!!edge.from_user_id && edge.from_user_id === edge.to_user_id)
}

/**
 * Steps that resolve an edge: stop flowing + result marker at the link end, then the
 * damage/heal impact number on the target seat.
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
  // The impact number lands on the target seat right after the strike resolves.
  if (damage > 0 && edge.to_user_id && end && (end.style === 'hit' || end.style === 'heal')) {
    steps.push({ type: 'value-chip', playerId: edge.to_user_id, text: heal ? `+${damage}` : `-${damage}`, color: heal ? 'heal' : 'damage' })
  }
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
