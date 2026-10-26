<script setup lang="ts">
/**
 * Battle "stage" drawn directly between player seats.
 *
 * Renders the state of useStageQueueStore (links, skill bubble, value chips, turn banner,
 * history chips, narration card) on top of the board and runs the store clock via rAF.
 * Visual reference: public/mockups/battle-readability.html.
 */
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import ElementGlyph from './card/ElementGlyph.vue'
import StatusEffectIcon from './StatusIcons/StatusEffectIcon.vue'
import { cardArtSrc } from '../constants/cardArt'
import { characterArtSrc } from '../constants/characterArt'
import { useSnapshotStore } from '../stores/snapshot.store'
import { useInterruptStore } from '../stores/interrupt.store'
import { useSessionStore } from '../stores/session.store'
import type { Card } from '../types/game'
import {
  useStageQueueStore,
  type LinkColor,
  type LinkEndStyle,
  type LinkState,
  type NarrationChain,
  type NarrationRow,
  type NarrationSkill,
  type ValueChipColor,
} from '../stores/stageQueue.store'

const props = defineProps<{
  rootEl: HTMLElement | null
  /** Optional label for the skill bubble header (e.g. "圣女 · 测试玩家1"). */
  playerLabel?: (playerId: string) => string
}>()

const store = useStageQueueStore()
const snapshotStore = useSnapshotStore()
const interruptStore = useInterruptStore()
const sessionStore = useSessionStore()

/** The waiting-for id is empty for the player holding the prompt, so fall back to "me". */
function isAwaiting(playerId: string): boolean {
  if (interruptStore.waitingFor) return interruptStore.waitingFor === playerId
  return !!interruptStore.currentPrompt && sessionStore.myPlayerId === playerId
}

/** Arc colour by card element, so the current strike reads as "雷 / 火 / 水…" at a glance. */
const ELEMENT_STROKE: Record<string, string> = {
  Fire: '#ff7a4d',
  Water: '#4fb0ff',
  Wind: '#4fd39a',
  Earth: '#d8a85a',
  Thunder: '#a98bff',
  Light: '#f0d27a',
  Dark: '#a07ce0',
}

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
  if (store.activeSkillBubble) {
    ids.add(store.activeSkillBubble.playerId)
    for (const id of store.activeSkillBubble.targetIds) ids.add(id)
  }
  for (const capsule of store.skillCapsules) ids.add(capsule.playerId)
  for (const chip of store.valueChips) ids.add(chip.playerId)
  for (const change of store.statusChanges) {
    ids.add(change.playerId)
    if (change.fromPlayerId) ids.add(change.fromPlayerId)
  }
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
  endPos: Point
  endStyle: { bg: string; fg: string; prefix: string } | null
  /** Latest live link: the only one drawn bright, with the flying card and target reticle. */
  current: boolean
  cardImageUrl: string
  target: SeatBox
  source: SeatBox
}

/** Newest link that is not part of the "上一手" layer. */
const currentLinkId = computed(() => {
  let latest: LinkState | null = null
  for (const link of store.activeLinks.values()) {
    if (link.previous) continue
    if (!latest || link.createdAt >= latest.createdAt) latest = link
  }
  return latest?.id ?? ''
})

interface MergedChainNode {
  row: NarrationRow
  chain?: NarrationChain
  skill?: NarrationSkill
  /** Link carrying the result: a cardless follow-up edge (same from→to) replaces the original. */
  resultLinkId: string
  /** Trigger skills that fired during this step, shown as small tags under the node. */
  triggers: NarrationSkill[]
}

/**
 * The engine emits a cardless "攻击" edge after a counter (same from→to as the counter edge).
 * Fold it into the counter node so the chain shows one card per step.
 */
const mergedChain = computed(() => {
  const source = narration.value
  if (!source) return null
  const nodes: MergedChainNode[] = []
  for (const row of source.rows) {
    const prev = nodes[nodes.length - 1]
    if (row.skill) {
      // Passive / response skills hang under the step they reacted to instead of adding a card.
      if (row.skill.opening === 'trigger' && prev) {
        prev.triggers.push(row.skill)
        continue
      }
      nodes.push({ row, skill: row.skill, resultLinkId: '', triggers: [] })
      continue
    }
    const chain = row.chain
    if (!chain) continue
    if (prev?.chain && !chain.card && prev.chain.fromPlayerId === chain.fromPlayerId && prev.chain.toPlayerId === chain.toPlayerId) {
      prev.resultLinkId = chain.linkId
      continue
    }
    nodes.push({ row, chain, resultLinkId: chain.linkId, triggers: [] })
  }
  return nodes.length ? { nodes, source } : null
})

/** Element per link id, so a cardless follow-up arc keeps the colour of the card it continues. */
const linkElements = computed(() => {
  const map = new Map<string, string>()
  for (const node of mergedChain.value?.nodes || []) {
    const element = node.chain?.card?.element
    if (!element || !node.chain) continue
    map.set(node.chain.linkId, element)
    map.set(node.resultLinkId, element)
  }
  return map
})

const linkViews = computed<LinkView[]>(() => {
  const views: LinkView[] = []
  const { width, height } = size.value
  const pairKey = (link: LinkState) => `${link.fromPlayerId}>${link.toPlayerId}`
  // A cardless follow-up edge shares from→to with the card edge it continues: reuse its
  // element, and draw only the newest of the pair so the arc isn't overlaid by a dashed copy.
  const pairElement = new Map<string, string>()
  const pairImage = new Map<string, string>()
  const pairNewest = new Map<string, LinkState>()
  for (const link of store.activeLinks.values()) {
    if (link.element) pairElement.set(pairKey(link), link.element)
    if (link.cardImageUrl) pairImage.set(pairKey(link), link.cardImageUrl)
    const newest = pairNewest.get(pairKey(link))
    if (!newest || link.createdAt >= newest.createdAt) pairNewest.set(pairKey(link), link)
  }
  for (const link of store.activeLinks.values()) {
    if (pairNewest.get(pairKey(link)) !== link) continue
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
      // Bulge upward in proportion to the horizontal span so long cross-board links arc
      // above the centre narration card instead of running through it.
      c = { x: (a.x + b.x) / 2, y: Math.max(16, Math.min(a.y, b.y) - Math.max(70, Math.abs(a.x - b.x) * 0.35)) }
    }
    const end = link.end ? END_STYLE[link.end.style] : null
    const nearEnd = quadAt(a, c, b, 0.82)
    const key = strokeKey(link)
    const element = link.element || linkElements.value.get(link.id) || pairElement.get(pairKey(link))
    const elementStroke = key !== 'grey' && element ? ELEMENT_STROKE[element] : undefined
    views.push({
      link,
      d: `M${a.x.toFixed(1)},${a.y.toFixed(1)} Q${c.x.toFixed(1)},${c.y.toFixed(1)} ${b.x.toFixed(1)},${b.y.toFixed(1)}`,
      stroke: elementStroke || LINK_COLORS[key],
      marker: elementStroke ? '' : `url(#stage-ah-${key})`,
      endPos: { x: nearEnd.x, y: nearEnd.y - 24 },
      endStyle: end,
      cardImageUrl: link.cardImageUrl || pairImage.get(pairKey(link)) || '',
      current: link.id === currentLinkId.value,
      target: toBox,
      source: fromBox,
    })
  }
  return views
})

const SKILL_ACCENT = '#f0d27a'

/**
 * Active skill opening: golden sigil ring on the caster seat, a plaque beside it, and
 * soft beams fanning out to each target (heal/status/strike skills with known targets).
 */
const castView = computed(() => {
  const bubble = store.activeSkillBubble
  if (!bubble) return null
  const box = seats.value[bubble.playerId]
  if (!box) return null
  const accent = (bubble.element && ELEMENT_STROKE[bubble.element]) || SKILL_ACCENT
  const gap = 16
  const plaqueStyle: Record<string, string> = { top: `${Math.max(4, box.top + 6)}px`, '--skill-accent': accent }
  if (box.leftSide) plaqueStyle.left = `${box.right + gap}px`
  else plaqueStyle.right = `${size.value.width - box.left + gap}px`
  const pad = 10
  const sigilStyle = {
    left: `${box.left - pad}px`,
    top: `${box.top - pad}px`,
    width: `${box.right - box.left + pad * 2}px`,
    height: `${box.bottom - box.top + pad * 2}px`,
    '--skill-accent': accent,
  }
  const beamColor = bubble.effects.includes('heal') ? '#bff5d2' : accent
  const from = { x: box.cx, y: box.cy }
  const beams = bubble.targetIds
    .filter((id) => id !== bubble.playerId)
    .map((id, index) => {
      const target = seats.value[id]
      if (!target) return null
      const to = { x: target.cx, y: target.cy }
      let c: Point
      if (Math.abs(to.x - from.x) < 40) {
        // Same column: bow toward the board centre instead of looping over the seats.
        const toward = size.value.width / 2 > from.x ? 1 : -1
        c = { x: from.x + toward * Math.max(110, Math.abs(to.y - from.y) * 0.5), y: (from.y + to.y) / 2 }
      } else {
        const lift = Math.max(60, Math.abs(to.x - from.x) * 0.22)
        c = { x: (from.x + to.x) / 2, y: Math.min(from.y, to.y) - lift }
      }
      return {
        id: `${bubble.id}-${id}`,
        d: `M${from.x.toFixed(1)},${from.y.toFixed(1)} Q${c.x.toFixed(1)},${c.y.toFixed(1)} ${to.x.toFixed(1)},${to.y.toFixed(1)}`,
        delay: `${0.35 + index * 0.12}s`,
      }
    })
    .filter((beam): beam is { id: string; d: string; delay: string } => !!beam)
  const effectHint = bubble.effects.length ? EFFECT_HINT[bubble.effects[0]!] : '主动技能'
  return {
    bubble,
    accent,
    beamColor,
    plaqueStyle,
    sigilStyle,
    beams,
    pointLeft: box.leftSide,
    avatar: characterArtSrc(snapshotStore.players[bubble.playerId]?.role),
    who: props.playerLabel ? props.playerLabel(bubble.playerId) : seatName(bubble.playerId),
    hint: bubble.targetIds.length > 1 ? `${effectHint} · ${bubble.targetIds.length} 名角色` : effectHint,
  }
})

const EFFECT_HINT: Record<string, string> = {
  strike: '伤害技能',
  heal: '治疗技能',
  status: '状态技能',
  resource: '资源技能',
  form: '形态技能',
  global: '全局技能',
}

/** Passive / response capsules: "技能名 · 触发" beside the seat, stacked when several fire. */
const capsuleViews = computed(() => {
  const perPlayer = new Map<string, number>()
  const views: { id: number; name: string; style: Record<string, string> }[] = []
  for (const capsule of store.skillCapsules) {
    const box = seats.value[capsule.playerId]
    if (!box) continue
    const index = perPlayer.get(capsule.playerId) ?? 0
    perPlayer.set(capsule.playerId, index + 1)
    const style: Record<string, string> = {
      top: `${box.bottom - 30 - index * 30}px`,
      '--skill-accent': (capsule.element && ELEMENT_STROKE[capsule.element]) || SKILL_ACCENT,
    }
    if (box.leftSide) style.left = `${box.right + 10}px`
    else style.right = `${size.value.width - box.left + 10}px`
    views.push({ id: capsule.id, name: capsule.skillName, style })
  }
  return views
})

/** Status icons: fly from the caster along a short arc, then settle on the target's lower edge. */
const statusViews = computed(() => {
  const views: { id: number; key: string; added: boolean; style: Record<string, string> }[] = []
  const perPlayer = new Map<string, number>()
  for (const change of store.statusChanges) {
    const box = seats.value[change.playerId]
    if (!box) continue
    const index = perPlayer.get(change.playerId) ?? 0
    perPlayer.set(change.playerId, index + 1)
    const end = { x: box.left + 22 + index * 30, y: box.bottom - 22 }
    const source = change.fromPlayerId ? seats.value[change.fromPlayerId] : undefined
    const start = source ? { x: source.cx, y: source.cy } : { x: end.x, y: end.y - 40 }
    const lift = source ? Math.max(50, Math.abs(end.x - start.x) * 0.25) : 0
    const c = { x: (start.x + end.x) / 2, y: Math.min(start.y, end.y) - lift }
    views.push({
      id: change.id,
      key: change.statusKey,
      added: change.added,
      style: {
        offsetPath: `path('M${start.x.toFixed(1)},${start.y.toFixed(1)} Q${c.x.toFixed(1)},${c.y.toFixed(1)} ${end.x.toFixed(1)},${end.y.toFixed(1)}')`,
        '--status-fly': source ? '0.7s' : '0s',
      },
    })
  }
  return views
})

interface ChipView {
  id: number
  text: string
  style: Record<string, string>
}

interface ImpactView {
  id: number
  /** Sign stripped: rendered as a separate glyph so the digits stay large. */
  sign: string
  amount: string
  heal: boolean
  /** Number position: centre of the target seat, stacked upward for repeated hits. */
  numberStyle: Record<string, string>
  /** Full seat rectangle for the impact flash. */
  flashStyle: Record<string, string>
  sparks: { id: number; style: Record<string, string> }[]
}

/** Element colour of the newest link landing on this player, so the impact glows like the strike. */
function impactColor(playerId: string, heal: boolean): string {
  if (heal) return '#59e39a'
  let latest: LinkState | null = null
  for (const link of store.activeLinks.values()) {
    if (link.toPlayerId !== playerId) continue
    if (!latest || link.createdAt >= latest.createdAt) latest = link
  }
  const element = latest?.element || (latest ? linkElements.value.get(latest.id) : undefined)
  return (element && ELEMENT_STROKE[element]) || '#ff5d5d'
}

const SPARK_COUNT = 8

/** Damage / heal become large impact numbers on the seat; other chips stay small badges. */
const chipLayers = computed(() => {
  const perPlayer = new Map<string, number>()
  const chips: ChipView[] = []
  const impacts: ImpactView[] = []
  for (const chip of store.valueChips) {
    const box = seats.value[chip.playerId]
    if (!box) continue
    const index = perPlayer.get(chip.playerId) ?? 0
    perPlayer.set(chip.playerId, index + 1)
    const isImpact = chip.color === 'damage' || chip.color === 'heal'
    if (!isImpact) {
      const anchor = anchorPoint(box)
      const x = anchor.x + (box.leftSide ? 36 : -36)
      const y = box.cy - 18 - index * 34
      chips.push({ id: chip.id, text: chip.text, style: { left: `${x}px`, top: `${y}px`, background: CHIP_BG[chip.color] } })
      continue
    }
    const heal = chip.color === 'heal'
    const match = chip.text.match(/^([+-]?)(.*)$/)
    const color = impactColor(chip.playerId, heal)
    const sparks = heal
      ? []
      : Array.from({ length: SPARK_COUNT }, (_, i) => {
          const angle = (i / SPARK_COUNT) * Math.PI * 2 + (chip.id % 7) * 0.3
          const dist = 46 + ((chip.id + i * 13) % 5) * 9
          return {
            id: i,
            style: {
              '--spark-x': `${(Math.cos(angle) * dist).toFixed(1)}px`,
              '--spark-y': `${(Math.sin(angle) * dist).toFixed(1)}px`,
              '--spark-rot': `${((angle * 180) / Math.PI).toFixed(0)}deg`,
            },
          }
        })
    impacts.push({
      id: chip.id,
      sign: match?.[1] || '',
      amount: match?.[2] || chip.text,
      heal,
      numberStyle: {
        left: `${box.cx}px`,
        top: `${box.cy - index * 44}px`,
        '--impact-color': color,
        // Later hits on the same seat wait a beat so each number reads on its own.
        animationDelay: `${index * 120}ms`,
      },
      flashStyle: {
        left: `${box.left}px`,
        top: `${box.top}px`,
        width: `${box.right - box.left}px`,
        height: `${box.bottom - box.top}px`,
        '--impact-color': color,
      },
      sparks,
    })
  }
  return { chips, impacts }
})

const narration = computed<{ rows: NarrationRow[]; previous: boolean } | null>(() => {
  if (store.stageRows.length > 0) return { rows: store.stageRows.slice(-MAX_ROWS), previous: false }
  if (store.previousAction && store.previousAction.rows.length > 0) {
    return { rows: store.previousAction.rows.slice(-MAX_ROWS), previous: true }
  }
  return null
})

interface ChainNodeView {
  id: number
  verb: string
  fromAvatar: string
  toAvatar: string
  fromName: string
  toName: string
  cardName: string
  cardArt: string
  element: string
  stroke: string
  state: 'past' | 'current' | 'waiting'
  result: string
  resultStyle: LinkEndStyle | ''
  /** Skill node: card face is the caster portrait + skill emblem instead of card art. */
  skill: boolean
  /** Extra targets beyond the one avatar shown ("+2"). */
  moreTargets: number
  triggers: { name: string; stroke: string }[]
}

function seatName(id: string): string {
  const player = snapshotStore.players[id]
  if (!player) return id
  return snapshotStore.characters[player.role || '']?.name || player.name || id
}

/**
 * Edge rows become nodes of a horizontal combat chain (card + 出手方→目标 avatars + verb);
 * skill/effect rows without a chain stay as short text lines under it.
 */
const chainView = computed(() => {
  const merged = mergedChain.value
  if (!merged) return null
  const { nodes: chainNodes, source } = merged
  const avatar = (id: string) => characterArtSrc(snapshotStore.players[id]?.role)
  const strokeOf = (element: string | undefined, fallback: LinkColor) => (element && ELEMENT_STROKE[element]) || LINK_COLORS[fallback]
  const nodes: ChainNodeView[] = chainNodes.map((entry, index) => {
    const { chain, skill, row } = entry
    const next = chainNodes[index + 1]
    const last = !next
    const triggers = entry.triggers.map((t) => ({ name: t.skillName, stroke: strokeOf(t.element, 'skill') }))
    if (skill) {
      const targets = skill.targetIds
      const firstTarget = targets.find((id) => id !== skill.actorId) || targets[0] || ''
      return {
        id: row.id,
        verb: '技能',
        fromAvatar: avatar(skill.actorId),
        toAvatar: firstTarget ? avatar(firstTarget) : '',
        fromName: seatName(skill.actorId),
        toName: targets.length ? targets.map(seatName).join('、') : '',
        cardName: skill.skillName,
        cardArt: avatar(skill.actorId),
        element: skill.element || '',
        stroke: strokeOf(skill.element, 'skill'),
        state: last && !source.previous ? 'current' : 'past',
        result: '',
        resultStyle: '',
        skill: true,
        moreTargets: Math.max(0, targets.length - 1),
        triggers,
      }
    }
    const c = chain!
    const link = store.activeLinks.get(entry.resultLinkId)
    const element = c.card?.element || ''
    const unresolved = !link?.end
    let state: ChainNodeView['state'] = 'past'
    if (last && !source.previous) {
      state = unresolved && isAwaiting(c.toPlayerId) ? 'waiting' : 'current'
    }
    let result = link?.end?.text || ''
    let resultStyle: ChainNodeView['resultStyle'] = link?.end?.style || ''
    // A countered attack resolves as "miss" for the original target; say what actually happened.
    if (next?.chain && next.chain.fromPlayerId === c.toPlayerId && next.chain.verb === '应战') {
      result = '被应战'
      resultStyle = 'miss'
    }
    return {
      id: row.id,
      verb: c.verb,
      fromAvatar: avatar(c.fromPlayerId),
      toAvatar: avatar(c.toPlayerId),
      fromName: seatName(c.fromPlayerId),
      toName: seatName(c.toPlayerId),
      cardName: c.card?.name || c.verb,
      cardArt: c.card?.name ? cardArtSrc(c.card as Card) : '',
      element,
      stroke: strokeOf(element, row.tagColor),
      state,
      result,
      resultStyle,
      skill: false,
      moreTargets: 0,
      triggers,
    }
  })
  // "X 承受" closes the chain: show it as the last card's result instead of a separate line.
  const lastNode = nodes[nodes.length - 1]
  const lastChainRowId = chainNodes[chainNodes.length - 1]?.row.id ?? 0
  const takeRow = source.rows.find((row) => !row.chain && !row.skill && row.tag === '承受' && row.id > lastChainRowId)
  if (lastNode && takeRow && !lastNode.skill) {
    if (!lastNode.result) {
      lastNode.result = '命中'
      lastNode.resultStyle = 'hit'
    }
    if (lastNode.state === 'waiting') lastNode.state = 'current'
  }
  const notes = source.rows.filter((row) => !row.chain && !row.skill && row !== takeRow).slice(-2)
  return { nodes: nodes.slice(-4), notes, previous: source.previous }
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
    if (
      store.activeLinks.size > 0 ||
      store.activeSkillBubble ||
      store.skillCapsules.length > 0 ||
      store.valueChips.length > 0 ||
      store.statusChanges.length > 0
    )
      measure()
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
        :class="{ 'stage-link--previous': view.link.previous, 'stage-link--past': !view.current && !view.link.previous }"
      >
        <path v-if="view.current" class="stage-link__glow" :d="view.d" :stroke="view.stroke" stroke-width="10" />
        <path
          class="stage-link__base"
          :class="{ 'stage-link__base--muted': view.link.animating && !view.current }"
          :d="view.d"
          :stroke="view.stroke"
          :stroke-width="view.current ? 3 : 2"
          pathLength="1"
          :marker-end="view.marker || undefined"
        />
        <path v-if="view.link.animating" class="stage-link__flow" :d="view.d" stroke="#fff" :stroke-width="view.current ? 4 : 3" pathLength="1" :style="{ filter: `drop-shadow(0 0 4px ${view.stroke})` }" />
      </g>
    </svg>

    <template v-for="view in linkViews" :key="`fx-${view.link.id}`">
      <!-- Card flies along the current arc once, then the reticle locks onto the target. -->
      <div
        v-if="view.current && view.cardImageUrl"
        :key="`fly-${view.link.id}`"
        class="stage-fly-card"
        :style="{ offsetPath: `path('${view.d}')`, '--fly-color': view.stroke }"
      >
        <img :src="view.cardImageUrl" alt="" />
      </div>
      <div
        v-if="view.current"
        class="stage-reticle"
        :class="{ 'stage-reticle--resolved': !!view.link.end }"
        :style="{
          left: `${view.target.left - 4}px`,
          top: `${view.target.top - 4}px`,
          width: `${view.target.right - view.target.left + 8}px`,
          height: `${view.target.bottom - view.target.top + 8}px`,
          '--fly-color': view.stroke,
        }"
      />
      <div
        v-if="view.current"
        class="stage-source"
        :style="{
          left: `${view.source.left - 3}px`,
          top: `${view.source.top - 3}px`,
          width: `${view.source.right - view.source.left + 6}px`,
          height: `${view.source.bottom - view.source.top + 6}px`,
          '--fly-color': view.stroke,
        }"
      />
      <div
        v-if="view.link.end && view.endStyle && view.current && view.link.end.style !== 'hit'"
        class="stage-end"
        :class="{ 'stage-dim': view.link.previous || !view.current }"
        :style="{ left: `${view.endPos.x}px`, top: `${view.endPos.y}px`, background: view.endStyle.bg, color: view.endStyle.fg }"
      >
        {{ view.endStyle.prefix }}{{ view.link.end.text }}
      </div>
    </template>

    <template v-if="castView">
      <svg
        class="stage-skill-beams"
        :width="size.width"
        :height="size.height"
        :viewBox="`0 0 ${size.width || 1} ${size.height || 1}`"
      >
        <g v-for="beam in castView.beams" :key="beam.id">
          <path class="stage-skill-beam__glow" :d="beam.d" :stroke="castView.beamColor" pathLength="1" :style="{ animationDelay: beam.delay }" />
          <path class="stage-skill-beam" :d="beam.d" :stroke="castView.beamColor" pathLength="1" :style="{ animationDelay: beam.delay }" />
        </g>
      </svg>
      <div :key="`sigil-${castView.bubble.id}`" class="stage-sigil" :style="castView.sigilStyle">
        <span class="stage-sigil__ring" />
        <span class="stage-sigil__ring stage-sigil__ring--inner" />
      </div>
      <div
        :key="`plaque-${castView.bubble.id}`"
        class="stage-plaque"
        :class="castView.pointLeft ? 'stage-plaque--from-left' : 'stage-plaque--from-right'"
        :style="castView.plaqueStyle"
        data-testid="stage-skill-plaque"
      >
        <img v-if="castView.avatar" class="stage-plaque__avatar" :src="castView.avatar" alt="" />
        <div class="stage-plaque__body">
          <div class="stage-plaque__who">{{ castView.who }} 发动</div>
          <div class="stage-plaque__name">{{ castView.bubble.skillName }}</div>
          <div class="stage-plaque__hint">{{ castView.hint }}</div>
        </div>
        <span class="stage-plaque__emblem">
          <ElementGlyph v-if="castView.bubble.element" :element="castView.bubble.element" />
          <span v-else class="stage-plaque__star">✦</span>
        </span>
      </div>
    </template>

    <div
      v-for="capsule in capsuleViews"
      :key="`cap-${capsule.id}`"
      class="stage-capsule"
      :style="capsule.style"
      data-testid="stage-skill-capsule"
    >
      <span class="stage-capsule__dot">✦</span>{{ capsule.name }}<span class="stage-capsule__tag">· 触发</span>
    </div>

    <div
      v-for="status in statusViews"
      :key="`status-${status.id}`"
      class="stage-status"
      :class="status.added ? 'stage-status--added' : 'stage-status--removed'"
      :style="status.style"
      data-testid="stage-status-fly"
    >
      <StatusEffectIcon :effect="status.key" />
    </div>

    <div v-for="chip in chipLayers.chips" :key="chip.id" class="stage-chip" :style="chip.style">{{ chip.text }}</div>

    <template v-for="impact in chipLayers.impacts" :key="`impact-${impact.id}`">
      <div class="stage-impact-flash" :class="{ 'stage-impact-flash--heal': impact.heal }" :style="impact.flashStyle" />
      <div
        class="stage-impact"
        :class="{ 'stage-impact--heal': impact.heal }"
        :style="impact.numberStyle"
        data-testid="stage-impact"
      >
        <span class="stage-impact__halo" />
        <span class="stage-impact__ring" />
        <span v-for="spark in impact.sparks" :key="spark.id" class="stage-impact__spark" :style="spark.style" />
        <span class="stage-impact__num"><small>{{ impact.sign }}</small>{{ impact.amount }}</span>
      </div>
    </template>

    <div v-if="store.turnBanner" :key="store.turnBanner.text" class="stage-banner">{{ store.turnBanner.text }}</div>

    <div v-if="historyView.length" class="stage-history">
      <span v-for="(chip, index) in historyView" :key="`${index}-${chip}`" class="stage-hchip">{{ chip }}</span>
    </div>

    <div
      v-if="chainView"
      class="stage-chain"
      :class="{ 'stage-chain--prev': chainView.previous }"
      data-testid="stage-chain"
    >
      <div class="stage-chain__tag">{{ chainView.previous ? '上一手' : '战斗链' }}</div>
      <div class="stage-chain__track">
        <template v-for="(node, index) in chainView.nodes" :key="node.id">
          <div v-if="index > 0" class="stage-chain__arrow" :class="{ 'stage-chain__arrow--live': node.state !== 'past' }" :style="{ '--node-color': node.stroke }" />
          <div
            class="stage-chain__node"
            :class="`stage-chain__node--${node.state}`"
            :style="{ '--node-color': node.stroke }"
            data-testid="stage-chain-node"
          >
            <div class="stage-chain__who" :title="node.toName ? `${node.fromName} → ${node.toName}` : node.fromName">
              <img v-if="node.fromAvatar" :src="node.fromAvatar" alt="" />
              <template v-if="node.toAvatar">
                <span class="stage-chain__who-arrow">›</span>
                <img :src="node.toAvatar" alt="" />
              </template>
              <span v-if="node.moreTargets > 0" class="stage-chain__more">+{{ node.moreTargets }}</span>
            </div>
            <div class="stage-chain__card" :class="{ 'stage-chain__card--skill': node.skill }">
              <template v-if="node.skill">
                <img v-if="node.cardArt" class="stage-chain__skill-art" :src="node.cardArt" alt="" />
                <span class="stage-chain__skill-name">{{ node.cardName }}</span>
              </template>
              <img v-else-if="node.cardArt" :src="node.cardArt" alt="" />
              <span v-else-if="!node.skill" class="stage-chain__card-text">{{ node.cardName }}</span>
              <span v-if="node.element" class="stage-chain__elem"><ElementGlyph :element="node.element" /></span>
              <span v-if="node.state === 'waiting'" class="stage-chain__timer" />
              <span v-if="node.result" class="stage-chain__result" :class="`stage-chain__result--${node.resultStyle}`">{{ node.result }}</span>
            </div>
            <div class="stage-chain__verb">{{ node.state === 'waiting' ? '等待响应' : node.verb }}</div>
            <div v-for="trigger in node.triggers" :key="trigger.name" class="stage-chain__trigger" :style="{ '--node-color': trigger.stroke }">
              ✦ {{ trigger.name }}
            </div>
            <div class="stage-chain__names">{{ node.fromName }} → {{ node.toName }}</div>
          </div>
        </template>
      </div>
      <div v-for="row in chainView.notes" :key="row.id" class="stage-chain__note">
        <span class="stage-row__text" v-html="row.html" />
      </div>
    </div>

    <div v-else-if="narration" class="stage-story" :class="{ 'stage-story--prev': narration.previous }">
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
  /* One short comet streak travelling source → target along the arc. */
  stroke-dasharray: 0.07 0.93;
  stroke-dashoffset: 0.07;
  opacity: 0.9;
  animation: stage-flow 1.1s cubic-bezier(0.4, 0, 0.6, 1) infinite;
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

.stage-sigil {
  position: absolute;
  pointer-events: none;
}
.stage-sigil__ring {
  position: absolute;
  inset: 0;
  border-radius: 16px;
  border: 2px solid var(--skill-accent);
  box-shadow: 0 0 18px color-mix(in srgb, var(--skill-accent) 70%, transparent), inset 0 0 14px color-mix(in srgb, var(--skill-accent) 45%, transparent);
  animation: stage-sigil-in 0.5s ease-out both, stage-sigil-pulse 1.4s 0.5s ease-in-out infinite;
}
.stage-sigil__ring--inner {
  inset: 6px;
  border-width: 1px;
  border-style: dashed;
  opacity: 0.7;
  animation: stage-sigil-in 0.5s 0.1s ease-out both;
}

.stage-plaque {
  position: absolute;
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 230px;
  max-width: 34%;
  padding: 8px 12px 8px 8px;
  border-radius: 999px;
  background: linear-gradient(180deg, rgba(22, 30, 48, 0.97), rgba(9, 14, 26, 0.97));
  border: 1px solid var(--bt-chrome-mid, #8090a8);
  box-shadow:
    0 0 0 1px rgba(0, 0, 0, 0.6),
    inset 0 1px 0 rgba(255, 255, 255, 0.08),
    0 0 22px color-mix(in srgb, var(--skill-accent) 35%, transparent),
    0 8px 18px rgba(0, 0, 0, 0.55);
}
.stage-plaque--from-left {
  animation: stage-plaque-in-left 0.45s cubic-bezier(0.2, 0.9, 0.3, 1.1) both;
}
.stage-plaque--from-right {
  animation: stage-plaque-in-right 0.45s cubic-bezier(0.2, 0.9, 0.3, 1.1) both;
}
.stage-plaque__avatar {
  width: 44px;
  height: 44px;
  flex: none;
  border-radius: 999px;
  object-fit: cover;
  object-position: center 20%;
  border: 2px solid var(--skill-accent);
  box-shadow: 0 0 10px color-mix(in srgb, var(--skill-accent) 60%, transparent);
}
.stage-plaque__body {
  min-width: 0;
  flex: 1;
}
.stage-plaque__who {
  font-size: 11px;
  color: var(--bt-text-muted, #9aa6bb);
}
.stage-plaque__name {
  font-size: 20px;
  font-weight: 800;
  letter-spacing: 3px;
  line-height: 1.2;
  color: #fff;
  text-shadow: 0 0 10px color-mix(in srgb, var(--skill-accent) 60%, transparent);
}
.stage-plaque__hint {
  font-size: 11px;
  color: #c6d6ea;
}
.stage-plaque__emblem {
  width: 34px;
  height: 34px;
  flex: none;
  border-radius: 999px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--skill-accent);
  background: radial-gradient(circle, rgba(255, 255, 255, 0.08), rgba(0, 0, 0, 0.4));
  border: 1px solid var(--skill-accent);
}
.stage-plaque__emblem :deep(svg) {
  width: 18px;
  height: 18px;
}
.stage-plaque__star {
  font-size: 16px;
}

.stage-skill-beams {
  position: absolute;
  inset: 0;
  overflow: visible;
  pointer-events: none;
}
.stage-skill-beam,
.stage-skill-beam__glow {
  fill: none;
  stroke-linecap: round;
  stroke-dasharray: 1;
  stroke-dashoffset: 1;
  animation: stage-beam 1.3s ease-out both;
}
.stage-skill-beam {
  stroke-width: 2;
}
.stage-skill-beam__glow {
  stroke-width: 8;
  opacity: 0.25;
  filter: blur(3px);
}

.stage-capsule {
  position: absolute;
  display: flex;
  align-items: center;
  gap: 5px;
  padding: 3px 12px 3px 8px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 700;
  white-space: nowrap;
  color: #fff;
  background: rgba(9, 14, 26, 0.94);
  border: 1px solid var(--skill-accent);
  box-shadow: 0 0 10px color-mix(in srgb, var(--skill-accent) 40%, transparent);
  animation: stage-capsule-life 1.6s ease both;
}
.stage-capsule__dot {
  color: var(--skill-accent);
}
.stage-capsule__tag {
  font-weight: 400;
  color: var(--bt-text-muted, #9aa6bb);
}

.stage-status {
  position: absolute;
  left: 0;
  top: 0;
  width: 26px;
  height: 26px;
  margin: -13px 0 0 -13px;
  offset-rotate: 0deg;
  pointer-events: none;
}
.stage-status :deep(.status-effect-icon) {
  width: 100%;
  height: 100%;
}
.stage-status--added {
  animation: stage-status-fly var(--status-fly) cubic-bezier(0.45, 0.05, 0.4, 1) both, stage-status-land 0.9s var(--status-fly) ease-out both;
}
.stage-status--removed {
  offset-distance: 100%;
  animation: stage-status-out 0.9s ease-in both;
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
    stroke-dashoffset: 0.07;
  }
  to {
    stroke-dashoffset: -0.93;
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

/* ===== 当前弧线：发光 + 飞牌 + 瞄准框 ===== */
.stage-link--past {
  opacity: 0.32;
}
.stage-link--past .stage-link__base {
  stroke-dasharray: 0.012 0.012;
  animation: none;
}
.stage-link__glow {
  opacity: 0.22;
  filter: blur(4px);
  animation: stage-draw 0.6s ease both;
  stroke-dasharray: 1;
}

.stage-fly-card {
  position: absolute;
  left: 0;
  top: 0;
  width: 34px;
  height: 48px;
  offset-rotate: 0deg;
  offset-distance: 100%;
  opacity: 0;
  animation: stage-fly 0.75s cubic-bezier(0.45, 0.05, 0.4, 1) both;
}
.stage-fly-card img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  border-radius: 4px;
  border: 1px solid var(--bt-chrome-hi, #dfe7f2);
  box-shadow: 0 0 14px var(--fly-color), 0 4px 10px rgba(0, 0, 0, 0.6);
}

.stage-reticle,
.stage-source {
  position: absolute;
  border-radius: 12px;
  pointer-events: none;
}
.stage-reticle {
  border: 2px solid var(--fly-color);
  box-shadow: 0 0 18px var(--fly-color), inset 0 0 16px color-mix(in srgb, var(--fly-color) 35%, transparent);
  animation: stage-reticle-in 0.45s 0.6s ease both, stage-reticle-pulse 1.6s 1.05s ease-in-out infinite;
}
.stage-reticle--resolved {
  animation: stage-impact 0.5s ease both;
  opacity: 0.55;
}
.stage-source {
  border: 1px solid color-mix(in srgb, var(--fly-color) 70%, white);
  box-shadow: 0 0 12px color-mix(in srgb, var(--fly-color) 60%, transparent);
  opacity: 0.8;
  animation: stage-fade-in 0.3s ease both;
}

/* ===== 战斗链：横向卡牌节点 ===== */
.stage-chain {
  position: absolute;
  left: 50%;
  top: 44%;
  transform: translate(-50%, -50%);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  max-width: 52%;
  transition: filter 0.5s, opacity 0.5s;
}
.stage-chain--prev {
  filter: saturate(0.3) brightness(0.75);
  opacity: 0.8;
}
.stage-chain__tag {
  font-size: 11px;
  letter-spacing: 3px;
  color: var(--bt-text-muted, #9aa6bb);
  padding: 1px 12px;
  border-top: 1px solid var(--bt-hairline-soft, rgba(255, 255, 255, 0.12));
  border-bottom: 1px solid var(--bt-hairline-soft, rgba(255, 255, 255, 0.12));
}
.stage-chain__track {
  display: flex;
  align-items: center;
  gap: 4px;
}
.stage-chain__arrow {
  width: 46px;
  height: 2px;
  margin-bottom: 30px;
  position: relative;
  background: repeating-linear-gradient(90deg, rgba(180, 196, 220, 0.45) 0 6px, transparent 6px 11px);
}
.stage-chain__arrow::after {
  content: '';
  position: absolute;
  right: -2px;
  top: -4px;
  border: 5px solid transparent;
  border-left: 7px solid rgba(180, 196, 220, 0.55);
  border-right: 0;
}
.stage-chain__arrow--live {
  background: linear-gradient(90deg, transparent, var(--node-color));
  box-shadow: 0 0 8px var(--node-color);
  animation: stage-arrow-in 0.4s ease both;
}
.stage-chain__arrow--live::after {
  border-left-color: var(--node-color);
}

.stage-chain__node {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 5px;
  transition: transform 0.4s ease, opacity 0.4s ease, filter 0.4s ease;
  animation: stage-node-in 0.45s ease both;
}
.stage-chain__node--past {
  transform: scale(0.82);
  opacity: 0.6;
  filter: saturate(0.5);
}
.stage-chain__node--current,
.stage-chain__node--waiting {
  transform: scale(1.08);
}
.stage-chain__who {
  display: flex;
  align-items: center;
  gap: 2px;
}
.stage-chain__who img {
  width: 26px;
  height: 26px;
  border-radius: 999px;
  object-fit: cover;
  object-position: center 20%;
  border: 1px solid var(--bt-chrome, #b8c4d6);
  box-shadow: 0 0 0 2px var(--bt-bg-0, #070b14);
}
.stage-chain__who-arrow {
  font-size: 14px;
  color: var(--bt-text-muted, #9aa6bb);
}
.stage-chain__card {
  position: relative;
  width: 64px;
  height: 90px;
  border-radius: 7px;
  overflow: visible;
  background: var(--bt-bg-0, #070b14);
  border: 1px solid var(--bt-chrome-mid, #8090a8);
  box-shadow: 0 6px 14px rgba(0, 0, 0, 0.55);
}
.stage-chain__card > img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  border-radius: 6px;
  display: block;
}
.stage-chain__card-text {
  display: flex;
  height: 100%;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  writing-mode: vertical-rl;
  letter-spacing: 2px;
}
.stage-chain__node--current .stage-chain__card,
.stage-chain__node--waiting .stage-chain__card {
  border-color: var(--node-color);
  box-shadow: 0 0 0 1px var(--node-color), 0 0 22px color-mix(in srgb, var(--node-color) 65%, transparent), 0 6px 14px rgba(0, 0, 0, 0.55);
}
.stage-chain__elem {
  position: absolute;
  left: 4px;
  top: 4px;
  width: 18px;
  height: 18px;
  border-radius: 999px;
  background: var(--bt-bg-0, #070b14);
  border: 1px solid var(--node-color);
  color: var(--node-color);
  display: flex;
  align-items: center;
  justify-content: center;
}
.stage-chain__elem :deep(svg) {
  width: 11px;
  height: 11px;
}
.stage-chain__timer {
  position: absolute;
  right: -9px;
  bottom: -9px;
  width: 22px;
  height: 22px;
  border-radius: 999px;
  background: var(--bt-bg-0, #070b14);
  border: 2px solid color-mix(in srgb, var(--node-color) 35%, transparent);
  border-top-color: var(--node-color);
  animation: stage-spin 1.2s linear infinite;
}
.stage-chain__result {
  position: absolute;
  left: 50%;
  bottom: -8px;
  transform: translateX(-50%);
  font-size: 11px;
  font-weight: 700;
  padding: 1px 7px;
  border-radius: 8px;
  white-space: nowrap;
  background: #4a5368;
  color: #fff;
  animation: stage-pop 0.35s ease both;
}
.stage-chain__result--hit {
  background: #a3262f;
}
.stage-chain__result--blocked {
  background: #f0d27a;
  color: #2a1d00;
}
.stage-chain__result--heal {
  background: #1c7a49;
}
.stage-chain__verb {
  margin-top: 4px;
  min-width: 64px;
  text-align: center;
  font-size: 12px;
  font-weight: 700;
  padding: 2px 12px;
  border-radius: 999px;
  background: rgba(9, 14, 26, 0.92);
  border: 1px solid var(--bt-hairline-soft, rgba(255, 255, 255, 0.14));
  color: var(--bt-text-muted, #b9c4d8);
  white-space: nowrap;
}
.stage-chain__node--current .stage-chain__verb,
.stage-chain__node--waiting .stage-chain__verb {
  color: #fff;
  border-color: var(--node-color);
  box-shadow: 0 0 10px color-mix(in srgb, var(--node-color) 45%, transparent);
}
.stage-chain__names {
  font-size: 11px;
  color: var(--bt-text-muted, #9aa6bb);
  white-space: nowrap;
  max-width: 120px;
  overflow: hidden;
  text-overflow: ellipsis;
}
.stage-chain__more {
  margin-left: 2px;
  font-size: 11px;
  font-weight: 700;
  color: var(--bt-text-muted, #b9c4d8);
}
.stage-chain__card--skill {
  overflow: hidden;
}
.stage-chain__skill-art {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
  object-position: center 18%;
  border-radius: 6px;
  filter: saturate(0.85) brightness(0.8);
}
.stage-chain__card--skill::after {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: 6px;
  background: linear-gradient(180deg, transparent 35%, rgba(5, 8, 16, 0.92) 100%);
}
.stage-chain__skill-name {
  position: absolute;
  left: 2px;
  right: 2px;
  bottom: 5px;
  z-index: 1;
  text-align: center;
  font-size: 11px;
  font-weight: 800;
  line-height: 1.15;
  letter-spacing: 1px;
  color: #fff;
  text-shadow: 0 0 6px var(--node-color);
}
.stage-chain__trigger {
  font-size: 11px;
  padding: 1px 8px;
  border-radius: 999px;
  white-space: nowrap;
  color: #dfe7f2;
  background: rgba(9, 14, 26, 0.9);
  border: 1px solid color-mix(in srgb, var(--node-color) 60%, transparent);
  animation: stage-node-in 0.35s ease both;
}
.stage-chain__note {
  font-size: 12px;
  max-width: 420px;
  padding: 3px 12px;
  border-radius: 10px;
  background: rgba(9, 14, 26, 0.8);
  color: #c6d6ea;
  display: flex;
}

@keyframes stage-sigil-in {
  from {
    opacity: 0;
    transform: scale(1.25) rotate(-4deg);
  }
}
@keyframes stage-sigil-pulse {
  50% {
    box-shadow: 0 0 8px color-mix(in srgb, var(--skill-accent) 40%, transparent), inset 0 0 6px transparent;
  }
}
@keyframes stage-plaque-in-left {
  from {
    opacity: 0;
    transform: translateX(-24px) scaleX(0.6);
    transform-origin: left center;
  }
}
@keyframes stage-plaque-in-right {
  from {
    opacity: 0;
    transform: translateX(24px) scaleX(0.6);
    transform-origin: right center;
  }
}
@keyframes stage-beam {
  0% {
    stroke-dashoffset: 1;
    opacity: 0;
  }
  10% {
    opacity: 1;
  }
  55% {
    stroke-dashoffset: 0;
    opacity: 1;
  }
  100% {
    stroke-dashoffset: 0;
    opacity: 0;
  }
}
@keyframes stage-capsule-life {
  0% {
    opacity: 0;
    transform: translateY(6px) scale(0.9);
  }
  15%,
  80% {
    opacity: 1;
    transform: none;
  }
  100% {
    opacity: 0;
  }
}
@keyframes stage-status-fly {
  from {
    offset-distance: 0%;
    transform: scale(0.6);
  }
  to {
    offset-distance: 100%;
    transform: scale(1.4);
  }
}
@keyframes stage-status-land {
  0% {
    offset-distance: 100%;
    transform: scale(1.4);
    filter: drop-shadow(0 0 10px #fff);
  }
  40% {
    transform: scale(1);
  }
  80% {
    opacity: 1;
  }
  100% {
    offset-distance: 100%;
    opacity: 0;
    transform: scale(1);
  }
}
@keyframes stage-status-out {
  0% {
    opacity: 1;
    transform: scale(1);
  }
  100% {
    opacity: 0;
    transform: scale(0.4) translateY(-12px);
    filter: grayscale(1);
  }
}

@keyframes stage-fly {
  0% {
    offset-distance: 0%;
    opacity: 0;
    transform: scale(0.6);
  }
  12% {
    opacity: 1;
    transform: scale(1);
  }
  85% {
    opacity: 1;
  }
  100% {
    offset-distance: 100%;
    opacity: 0;
    transform: scale(0.7);
  }
}
@keyframes stage-reticle-in {
  from {
    opacity: 0;
    transform: scale(1.18);
  }
}
@keyframes stage-reticle-pulse {
  50% {
    box-shadow: 0 0 6px var(--fly-color), inset 0 0 6px transparent;
  }
}
@keyframes stage-impact {
  0% {
    opacity: 1;
    transform: scale(1);
    box-shadow: 0 0 40px var(--fly-color);
  }
  30% {
    transform: translateX(-3px);
  }
  60% {
    transform: translateX(3px);
  }
}
@keyframes stage-node-in {
  from {
    opacity: 0;
    transform: translateY(10px) scale(0.85);
  }
}
@keyframes stage-arrow-in {
  from {
    transform: scaleX(0);
    transform-origin: left;
  }
}
@keyframes stage-spin {
  to {
    transform: rotate(360deg);
  }
}

/* ===== 伤害 / 治疗冲击数字 ===== */
.stage-impact-flash {
  position: absolute;
  border-radius: 10px;
  pointer-events: none;
  background: radial-gradient(circle at 50% 50%, rgba(255, 255, 255, 0.85) 0%, color-mix(in srgb, var(--impact-color) 70%, transparent) 30%, transparent 75%);
  box-shadow: inset 0 0 0 3px var(--impact-color), 0 0 34px var(--impact-color);
  mix-blend-mode: screen;
  animation: stage-impact-flash 0.7s ease-out both;
}
.stage-impact-flash--heal {
  animation-duration: 1s;
}
.stage-impact {
  position: absolute;
  width: 0;
  height: 0;
  pointer-events: none;
  z-index: 2;
  animation: stage-impact-life 2s ease both;
}
.stage-impact__num {
  position: absolute;
  left: 0;
  top: 0;
  transform: translate(-50%, -50%);
  font-family: 'Cinzel', 'Times New Roman', 'Songti SC', serif;
  font-weight: 900;
  line-height: 1;
  white-space: nowrap;
  color: #fff;
  font-size: 72px;
  background: linear-gradient(180deg, #fff 0%, #fff2ee 30%, #ff6a5a 58%, #c0182a 100%);
  -webkit-background-clip: text;
  background-clip: text;
  -webkit-text-fill-color: transparent;
  -webkit-text-stroke: 2.5px #2a0006;
  paint-order: stroke fill;
  filter: drop-shadow(0 0 2px #000) drop-shadow(0 0 12px var(--impact-color)) drop-shadow(0 4px 3px rgba(0, 0, 0, 0.9));
  animation: stage-impact-num 2s cubic-bezier(0.2, 0.9, 0.3, 1.2) both;
  animation-delay: inherit;
}
.stage-impact__num small {
  font-size: 0.62em;
  margin-right: 2px;
  vertical-align: 0.2em;
}
.stage-impact--heal .stage-impact__num {
  background: linear-gradient(180deg, #fff 0%, #d6ffe6 40%, #3ed27f 70%, #11663a 100%);
  -webkit-background-clip: text;
  background-clip: text;
  -webkit-text-stroke-color: rgba(0, 30, 12, 0.9);
  font-size: 52px;
}
.stage-impact__halo {
  position: absolute;
  left: 0;
  top: 0;
  width: 150px;
  height: 110px;
  margin: -55px 0 0 -75px;
  border-radius: 999px;
  background: radial-gradient(ellipse at center, rgba(4, 6, 14, 0.78) 0%, rgba(4, 6, 14, 0.5) 45%, transparent 72%);
  animation: stage-impact-halo 2s ease both;
  animation-delay: inherit;
}
.stage-impact__ring {
  position: absolute;
  left: 0;
  top: 0;
  width: 130px;
  height: 130px;
  margin: -65px 0 0 -65px;
  border-radius: 999px;
  border: 4px solid #fff;
  box-shadow: 0 0 0 3px var(--impact-color), 0 0 26px var(--impact-color), inset 0 0 18px var(--impact-color);
  opacity: 0;
  animation: stage-impact-ring 0.6s ease-out both;
  animation-delay: inherit;
}
.stage-impact__spark {
  position: absolute;
  left: 0;
  top: 0;
  width: 30px;
  height: 4px;
  margin: -2px 0 0 -15px;
  border-radius: 2px;
  background: linear-gradient(90deg, transparent, #fff 35%, var(--impact-color));
  box-shadow: 0 0 8px var(--impact-color), 0 0 2px #fff;
  opacity: 0;
  transform: rotate(var(--spark-rot));
  animation: stage-impact-spark 0.6s cubic-bezier(0.1, 0.7, 0.3, 1) both;
  animation-delay: inherit;
}

@keyframes stage-impact-life {
  0%,
  92% {
    opacity: 1;
  }
  100% {
    opacity: 0;
  }
}
@keyframes stage-impact-num {
  0% {
    opacity: 0;
    transform: translate(-50%, -50%) scale(2.4);
  }
  14% {
    opacity: 1;
    transform: translate(-50%, -50%) scale(0.9);
  }
  22% {
    transform: translate(-50%, -50%) scale(1.08);
  }
  30% {
    transform: translate(-50%, -50%) scale(1);
  }
  78% {
    opacity: 1;
    transform: translate(-50%, -62%) scale(1);
  }
  100% {
    opacity: 0;
    transform: translate(-50%, -95%) scale(0.92);
  }
}
@keyframes stage-impact-halo {
  0% {
    opacity: 0;
    transform: scale(0.4);
  }
  12%,
  80% {
    opacity: 1;
    transform: scale(1);
  }
  100% {
    opacity: 0;
    transform: scale(1);
  }
}
@keyframes stage-impact-ring {
  0% {
    opacity: 0.95;
    transform: scale(0.2);
  }
  100% {
    opacity: 0;
    transform: scale(1.5);
  }
}
@keyframes stage-impact-spark {
  0% {
    opacity: 1;
    transform: translate(0, 0) rotate(var(--spark-rot)) scaleX(0.4);
  }
  100% {
    opacity: 0;
    transform: translate(calc(var(--spark-x) * 1.4), calc(var(--spark-y) * 1.4)) rotate(var(--spark-rot)) scaleX(1);
  }
}
@keyframes stage-impact-flash {
  0% {
    opacity: 0;
    transform: translateX(0);
  }
  12% {
    opacity: 1;
    transform: translateX(-5px);
  }
  24% {
    transform: translateX(5px);
  }
  36% {
    transform: translateX(-3px);
  }
  48% {
    transform: translateX(2px);
  }
  100% {
    opacity: 0;
    transform: translateX(0);
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
