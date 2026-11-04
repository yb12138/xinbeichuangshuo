import { SKILL_ANIMATION_MAP } from '../constants/skillAnimationMap'
import type { Card } from '../types/game'
import type { ActionFlowDTO, ActionFlowEdgeDTO, ActionFlowNodeDTO, TimelineNotifyPayload } from './protocol'
import {
  deltaSteps,
  edgeEnd,
  edgeResolveSteps,
  edgeStartSteps,
  escapeHtml,
  flowLinkId,
  isSelfResolveEdge,
  flowLogSteps,
  handFlightSteps,
  linkColorForPhase,
  nodeDamageSteps,
  nodeSkillResult,
  nodeSteps,
  type AtomicStep,
  type MoraleCauses,
  type useStageQueueStore,
} from '../stores/stageQueue.store'

/**
 * Bridges NotifyTimeline payloads into the stage queue.
 *
 * The backend re-sends the FULL rebuilt `action_flows` graph on every timeline message,
 * so this bridge is incremental: per flow_id it remembers which edges / nodes / outcomes
 * were already enqueued and only enqueues the delta.
 */

type StageQueue = Pick<ReturnType<typeof useStageQueueStore>, 'enqueue' | 'reset'>

export interface StageQueueBridgeDeps {
  stageQueue: StageQueue
  playerName: (id: string) => string
  roleName: (id: string) => string
  cardImage?: (card: Card) => string | undefined
  /** 玩家所属阵营（"Red" / "Blue"），用于把士气下降的原因对到正确的阵营。 */
  playerCamp?: (id: string) => string
}

interface EdgeMemory {
  /** Serialized edgeEnd() already shown ('' = none yet). */
  endKey: string
  /** Damage value already shown as a chip (0 = none yet). */
  damage: number
  /** Whether the "flowing" animation has been stopped for this link. */
  settled: boolean
}

interface FlowMemory {
  edges: Map<string, EdgeMemory>
  /** node id -> damage already shown as chips. */
  nodes: Map<string, number>
  /** node id -> skill result text already shown (cause-backed skill nodes). */
  results: Map<string, string>
  /** node id -> 已经飞过的手牌流动张数（from>to → count）。 */
  handFlights: Map<string, Map<string, number>>
  logOrders: Set<number>
  /** Acting player of the flow; state deltas carry no actor, so status icons fly from here. */
  actorId: string
  narrated: boolean
  /** True once the flow produced its first step (and its mark-previous was enqueued). */
  started: boolean
  /** 本行动里各玩家已到达的状态差值（每条时间线只带一个事件，特殊行动的后果要跨消息累计）。 */
  playerDeltas: Map<string, Map<string, number>>
  /** 本行动里已经出现过的阵营资源差值事件 id（特殊行动节点要排在它们前面）。 */
  campResourceEventIds: Set<number>
}

const SPECIAL_ACTION_LABEL: Record<string, string> = { buy: '购买', synthesize: '合成', extract: '提炼' }

/** 引擎给特殊行动的 action_id 加了类型后缀（"nw-t2-p1-a1-buy"）。 */
function isSpecialActionId(actionId: string): boolean {
  return /-(buy|synthesize|extract)$/i.test(actionId)
}

/** 特殊行动后果里只统计这几项（按服务端状态差值类型）。 */
const PLAYER_OUTCOME_LABEL: Record<string, string> = { hand_count: '手牌', player_gem: '宝石', player_crystal: '水晶' }

/** Keep memory for a bounded number of recent flows. */
const MAX_REMEMBERED_FLOWS = 16

/** Edges that wait for a response keep flowing until an outcome arrives. */
function awaitsResponse(edge: ActionFlowEdgeDTO): boolean {
  if (isSelfResolveEdge(edge)) return false
  const color = linkColorForPhase(edge.phase, edge.damage_type)
  return (color === 'attack' || color === 'magic') && !edgeEnd(edge) && !(edge.damage && edge.damage > 0)
}

function endKeyOf(edge: ActionFlowEdgeDTO): string {
  const end = edgeEnd(edge)
  return end ? `${end.style}|${end.text}` : ''
}

/** On replay, drop transient beats and zero every hold so the final state applies immediately. */
const REPLAY_DROPPED: ReadonlySet<AtomicStep['type']> = new Set(['skill-bubble', 'value-chip', 'status-change', 'turn-banner'])

function toReplaySteps(steps: AtomicStep[]): AtomicStep[] {
  return steps
    .filter((step) => !REPLAY_DROPPED.has(step.type) && !(step.type === 'link-animate' && step.active))
    .map((step) => ({ ...step, holdMs: 0 }))
}

export function createStageQueueBridge(deps: StageQueueBridgeDeps) {
  const { stageQueue, playerName, roleName, cardImage, playerCamp } = deps
  const flows = new Map<string, FlowMemory>()
  let currentFlowId: string | null = null
  /**
   * 当前战斗链所属的回合窗口（nw-t{回合}-{玩家}）。同一名玩家这一回合里的每一手（攻击、应战、
   * 回合结束的自动填充…）都接在同一条链上；换到下一名玩家的回合才把整条链收进"上一回合"。
   */
  let currentWindowId: string | null = null

  /** 从 flow / action id 取回合窗口：flow 带了就用，没带时按 "nw-t1-p1-a2-xxx" 截出前缀。 */
  function windowOf(windowId: string | undefined, actionId: string): string {
    if (windowId) return windowId
    const m = /^(nw-t\d+-[^-]+)-a\d+/.exec(actionId)
    return m ? m[1]! : actionId
  }

  /** 新的一手开始：同一回合窗口内继续累积，换了窗口才收起上一回合。 */
  function startAction(flowId: string, windowId: string, batch: AtomicStep[]) {
    if (flowId !== currentFlowId && windowId !== currentWindowId) batch.push({ type: 'mark-previous' })
    currentFlowId = flowId
    currentWindowId = windowId
  }
  /** 引擎在士气差值之前先发 morale_loss 标记说明原因；按阵营暂存，等差值到达时写进士气行。 */
  const pendingMoraleCauses: MoraleCauses = {}

  function memoryFor(flowId: string): FlowMemory {
    const existing = flows.get(flowId)
    if (existing) return existing
    const memory: FlowMemory = {
      edges: new Map(),
      nodes: new Map(),
      results: new Map(),
      handFlights: new Map(),
      logOrders: new Set(),
      actorId: '',
      narrated: false,
      started: false,
      playerDeltas: new Map(),
      campResourceEventIds: new Set(),
    }
    flows.set(flowId, memory)
    while (flows.size > MAX_REMEMBERED_FLOWS) {
      const oldest = flows.keys().next().value
      if (oldest === undefined || oldest === flowId) break
      flows.delete(oldest)
    }
    return memory
  }

  function nodeDelta(node: ActionFlowNodeDTO, memory: FlowMemory): AtomicStep[] {
    const shownDamage = memory.nodes.get(node.id)
    const damage = node.damage ?? 0
    const resultKey = node.cause_id ? JSON.stringify(nodeSkillResult(node, playerName, roleName)) : ''
    let shownFlights = memory.handFlights.get(node.id)
    if (!shownFlights) {
      shownFlights = new Map()
      memory.handFlights.set(node.id, shownFlights)
    }
    if (shownDamage === undefined) {
      memory.nodes.set(node.id, damage)
      if (resultKey) memory.results.set(node.id, resultKey)
      return [...nodeSteps(node, playerName, roleName), ...handFlightSteps(node, shownFlights)]
    }
    const steps: AtomicStep[] = []
    if (resultKey && resultKey !== memory.results.get(node.id)) {
      memory.results.set(node.id, resultKey)
      // 牌先飞、结果文案随后落定
      steps.push(...handFlightSteps(node, shownFlights))
      steps.push({ type: 'skill-result', causeId: node.cause_id!, ...nodeSkillResult(node, playerName, roleName) })
    }
    if (damage > 0 && damage !== shownDamage) {
      memory.nodes.set(node.id, damage)
      steps.push(...nodeDamageSteps(node))
    }
    return steps
  }

  function edgeResolveDelta(flowId: string, edge: ActionFlowEdgeDTO, edgeMemory: EdgeMemory): AtomicStep[] {
    const steps: AtomicStep[] = []
    const linkId = flowLinkId(flowId, edge.id)
    const endKey = endKeyOf(edge)
    const damage = edge.damage ?? 0
    const resolve = edgeResolveSteps(flowId, edge)

    if (!edgeMemory.settled && !awaitsResponse(edge)) {
      steps.push({ type: 'link-animate', linkId, active: false })
      edgeMemory.settled = true
    }
    if (endKey && endKey !== edgeMemory.endKey) {
      steps.push(...resolve.filter((step) => step.type === 'link-end'))
      edgeMemory.endKey = endKey
    }
    if (damage > 0 && damage !== edgeMemory.damage) {
      steps.push(...resolve.filter((step) => step.type === 'value-chip'))
      edgeMemory.damage = damage
    }
    return steps
  }

  /** Compute steps not yet enqueued for this flow, updating memory as a side effect. */
  function flowDelta(flow: ActionFlowDTO, memory: FlowMemory): AtomicStep[] {
    const steps: AtomicStep[] = []
    const edges = [...(flow.edges || [])].sort((a, b) => a.order - b.order)
    const nodes = [...(flow.nodes || [])].sort((a, b) => a.order - b.order)
    const edgeIds = new Set(edges.map((edge) => edge.id))
    const anchored = new Map<string, ActionFlowNodeDTO[]>()
    const free: ActionFlowNodeDTO[] = []
    for (const node of nodes) {
      if (node.anchor_edge_id && edgeIds.has(node.anchor_edge_id)) {
        const list = anchored.get(node.anchor_edge_id) || []
        list.push(node)
        anchored.set(node.anchor_edge_id, list)
      } else {
        free.push(node)
      }
    }

    for (const edge of edges) {
      let edgeMemory = memory.edges.get(edge.id)
      if (!edgeMemory) {
        edgeMemory = { endKey: '', damage: 0, settled: false }
        memory.edges.set(edge.id, edgeMemory)
        steps.push(...edgeStartSteps(flow.flow_id, edge, playerName, roleName, cardImage))
      }
      for (const node of anchored.get(edge.id) || []) {
        steps.push(...nodeDelta(node, memory))
      }
      steps.push(...edgeResolveDelta(flow.flow_id, edge, edgeMemory))
    }
    for (const node of free) {
      steps.push(...nodeDelta(node, memory))
    }

    if (steps.some((step) => step.type === 'narration-row')) memory.narrated = true
    if (!memory.narrated) {
      const freshLogs = (flow.logs || []).filter((log) => !memory.logOrders.has(log.order))
      for (const log of freshLogs) memory.logOrders.add(log.order)
      steps.push(...flowLogSteps(flow, freshLogs))
    }
    return steps
  }

  function onTimeline(payload: TimelineNotifyPayload) {
    const batch: AtomicStep[] = []
    for (const flow of payload.action_flows || []) {
      if (!flow?.flow_id) continue
      const memory = memoryFor(flow.flow_id)
      if (flow.actor_user_id) memory.actorId = flow.actor_user_id
      const delta = flowDelta(flow, memory)
      if (delta.length === 0) continue
      // 新的一手：同一回合内接在链后面；换了回合才把这一条链收进"上一回合"
      if (!memory.started) {
        memory.started = true
        startAction(flow.flow_id, windowOf(flow.narrative_window_id, flow.action_id || flow.flow_id), batch)
      }
      batch.push(...delta)
    }
    // Heal / status changes only arrive as state deltas; play them after the flow beats they follow.
    if (!payload.is_replay) {
      for (const event of payload.events || []) {
        // 特殊行动（购买 / 合成 / 提炼）没有攻击连线，流图里只有一个空的行动节点。
        // 它的摸牌、战绩区星石差值先于 special_action 事件到达，所以从该行动的第一个事件起
        // 就开新的一手，避免这些差值落进上一手。
        const actionId = event.action_id || ''
        if (isSpecialActionId(actionId)) {
          const memory = memoryFor(actionId)
          if (!memory.started) {
            memory.started = true
            startAction(actionId, windowOf(event.narrative_window_id, actionId), batch)
          }
        }
        const special = specialActionStep(event)
        if (special) {
          const memory = actionId ? memoryFor(actionId) : null
          if (!memory || !memory.narrated) {
            batch.push(special)
            if (memory) memory.narrated = true
          }
        }
        if (event.narrative_kind === 'morale_loss') rememberMoraleCause(event)
        if (!event.deltas?.length) continue
        const flowId = event.action_id || event.source_event_id || ''
        if (flowId && event.action_id) rememberPlayerDeltas(memoryFor(flowId), event)
        // 状态差值事件本身没有行动者：优先用技能因果的持有者（地下法则 +2 宝石 → 从冒险者飞出）。
        const actorId = event.actor_user_id || event.cause?.owner_id || (flowId ? flows.get(flowId)?.actorId : '') || ''
        const skillId = event.cause?.skill_id || ''
        batch.push(...deltaSteps(event.deltas, actorId, skillId ? SKILL_ANIMATION_MAP[skillId]?.element : undefined, pendingMoraleCauses))
        for (const delta of event.deltas) {
          if (delta.type === 'morale' && delta.camp) delete pendingMoraleCauses[delta.camp]
        }
      }
    }
    if (batch.length === 0) return
    stageQueue.enqueue(...(payload.is_replay ? toReplaySteps(batch) : batch))
  }

  /** special_action 事件 → 战斗链节点。同一事件随整张时间线反复下发，靠 flow 的 narrated 标记去重。 */
  function specialActionStep(event: NonNullable<TimelineNotifyPayload['events']>[number]): AtomicStep | null {
    if (event.gameplay_type !== 'special_action') return null
    const actorId = event.actor_user_id || ''
    const action = SPECIAL_ACTION_LABEL[String(event.action_type || '').toLowerCase()]
    if (!actorId || !action) return null
    const who = roleName(actorId) || playerName(actorId)
    return {
      type: 'narration-row',
      html: `${escapeHtml(who)} 执行特殊行动【${action}】`,
      tag: '特殊',
      tagColor: 'grey',
      special: { actorId, action, outcome: specialActionOutcome(event.action_id || '', actorId) },
    }
  }

  /** "吟游诗人爆牌" / "吟游诗人爆牌 · 凋零"：爆牌者 + 造成伤害的技能（若有）。 */
  function rememberMoraleCause(event: NonNullable<TimelineNotifyPayload['events']>[number]) {
    const victimId = event.actor_user_id || ''
    const camp = victimId && playerCamp ? playerCamp(victimId) : ''
    if (!victimId || !camp) return
    const who = roleName(victimId) || playerName(victimId)
    const skill = String(event.skill_name || '').trim()
    pendingMoraleCauses[camp] = skill ? `${who}爆牌 · ${skill}` : `${who}爆牌`
  }

  function rememberPlayerDeltas(memory: FlowMemory, event: NonNullable<TimelineNotifyPayload['events']>[number]) {
    for (const delta of event.deltas || []) {
      const label = PLAYER_OUTCOME_LABEL[delta.type]
      if (!label || !delta.target_user_id) continue
      const value = delta.value ?? (delta.after ?? 0) - (delta.before ?? 0)
      const perPlayer = memory.playerDeltas.get(delta.target_user_id) || new Map<string, number>()
      perPlayer.set(label, (perPlayer.get(label) ?? 0) + value)
      memory.playerDeltas.set(delta.target_user_id, perPlayer)
    }
  }

  /**
   * 行动者自身的后果（"手牌 +3" / "宝石 +1 · 水晶 +1"）：取同一行动里该玩家已到达的状态差值，
   * 不按行动名猜。阵营战绩区的变化由 campResourceSteps 另起一行。
   */
  function specialActionOutcome(actionId: string, actorId: string): string {
    const totals = actionId ? flows.get(actionId)?.playerDeltas.get(actorId) : undefined
    if (!totals) return ''
    return [...totals]
      .filter(([, value]) => value !== 0)
      .map(([label, value]) => `${label} ${value > 0 ? '+' : ''}${value}`)
      .join(' · ')
  }

  function reset() {
    for (const camp of Object.keys(pendingMoraleCauses)) delete pendingMoraleCauses[camp]
    flows.clear()
    currentFlowId = null
    currentWindowId = null
    stageQueue.reset()
  }

  return { onTimeline, reset }
}
