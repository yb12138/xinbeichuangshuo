import { SKILL_ANIMATION_MAP } from '../constants/skillAnimationMap'
import type { Card } from '../types/game'
import type { ActionFlowDTO, ActionFlowEdgeDTO, ActionFlowNodeDTO, TimelineNotifyPayload } from './protocol'
import {
  deltaSteps,
  edgeEnd,
  edgeResolveSteps,
  edgeStartSteps,
  flowLinkId,
  isSelfResolveEdge,
  flowLogSteps,
  linkColorForPhase,
  nodeDamageSteps,
  nodeSkillResult,
  nodeSteps,
  type AtomicStep,
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
  logOrders: Set<number>
  /** Acting player of the flow; state deltas carry no actor, so status icons fly from here. */
  actorId: string
  narrated: boolean
  /** True once the flow produced its first step (and its mark-previous was enqueued). */
  started: boolean
}

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
  const { stageQueue, playerName, roleName, cardImage } = deps
  const flows = new Map<string, FlowMemory>()
  let currentFlowId: string | null = null

  function memoryFor(flowId: string): FlowMemory {
    const existing = flows.get(flowId)
    if (existing) return existing
    const memory: FlowMemory = { edges: new Map(), nodes: new Map(), results: new Map(), logOrders: new Set(), actorId: '', narrated: false, started: false }
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
    if (shownDamage === undefined) {
      memory.nodes.set(node.id, damage)
      if (resultKey) memory.results.set(node.id, resultKey)
      return nodeSteps(node, playerName, roleName)
    }
    const steps: AtomicStep[] = []
    if (resultKey && resultKey !== memory.results.get(node.id)) {
      memory.results.set(node.id, resultKey)
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
      // A new action: demote the current one to "上一手" before its first beat.
      if (!memory.started) {
        memory.started = true
        if (flow.flow_id !== currentFlowId) batch.push({ type: 'mark-previous' })
        currentFlowId = flow.flow_id
      }
      batch.push(...delta)
    }
    // Heal / status changes only arrive as state deltas; play them after the flow beats they follow.
    if (!payload.is_replay) {
      for (const event of payload.events || []) {
        if (!event.deltas?.length) continue
        const flowId = event.action_id || event.source_event_id || ''
        const actorId = event.actor_user_id || (flowId ? flows.get(flowId)?.actorId : '') || ''
        const skillId = event.cause?.skill_id || ''
        batch.push(...deltaSteps(event.deltas, actorId, skillId ? SKILL_ANIMATION_MAP[skillId]?.element : undefined))
      }
    }
    if (batch.length === 0) return
    stageQueue.enqueue(...(payload.is_replay ? toReplaySteps(batch) : batch))
  }

  function reset() {
    flows.clear()
    currentFlowId = null
    stageQueue.reset()
  }

  return { onTimeline, reset }
}
