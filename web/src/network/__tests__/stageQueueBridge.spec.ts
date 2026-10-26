import { beforeEach, describe, expect, it } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { createStageQueueBridge } from '../stageQueueBridge'
import { useStageQueueStore, type AtomicStep } from '../../stores/stageQueue.store'
import type { ActionFlowDTO, ActionFlowEdgeDTO, TimelineNotifyPayload } from '../protocol'

const NAMES: Record<string, string> = { p1: '测试玩家1', p5: '测试玩家5' }
const ROLES: Record<string, string> = { p1: '圣女', p5: '封印师' }

function attackEdge(overrides: Partial<ActionFlowEdgeDTO> = {}): ActionFlowEdgeDTO {
  return { id: 'e1', order: 1, from_user_id: 'p5', to_user_id: 'p1', phase: 'attack', label: '攻击·暗灭', ...overrides }
}

function flow(flowId: string, edges: ActionFlowEdgeDTO[], extra: Partial<ActionFlowDTO> = {}): ActionFlowDTO {
  return { flow_id: flowId, action_id: flowId, action_type: 'attack', actors: [], edges, ...extra }
}

function payload(flows: ActionFlowDTO[], isReplay = false): TimelineNotifyPayload {
  return { room_id: 'r1', seq_start: 1, seq_end: 1, is_replay: isReplay, events: [], action_flows: flows }
}

describe('createStageQueueBridge', () => {
  let store: ReturnType<typeof useStageQueueStore>
  let bridge: ReturnType<typeof createStageQueueBridge>

  const queued = (): AtomicStep[] => [...store.queue]
  const types = (steps: AtomicStep[]) => steps.map((step) => step.type)

  beforeEach(() => {
    setActivePinia(createPinia())
    store = useStageQueueStore()
    bridge = createStageQueueBridge({
      stageQueue: store,
      playerName: (id) => NAMES[id] || id,
      roleName: (id) => ROLES[id] || '',
    })
  })

  it('enqueues the first flow with mark-previous first and role-first labels', () => {
    bridge.onTimeline(payload([flow('f1', [attackEdge()])]))
    const steps = queued()
    expect(steps[0]).toEqual({ type: 'mark-previous' })
    expect(types(steps)).toEqual(['mark-previous', 'link', 'link-animate', 'narration-row'])
    const row = steps.find((step) => step.type === 'narration-row')
    expect(row && row.type === 'narration-row' && row.html).toContain('封印师（测试玩家5）')
  })

  it('attaches structured chain data and the card element to edge steps', () => {
    const card = { id: 'c1', name: '雷光斩', type: 'Attack', element: 'Thunder', damage: 2 } as NonNullable<ActionFlowEdgeDTO['cards']>[number]
    bridge.onTimeline(payload([flow('f1', [attackEdge({ cards: [card] })])]))
    const steps = queued()
    const link = steps.find((step) => step.type === 'link')
    expect(link && link.type === 'link' && link.element).toBe('Thunder')
    const row = steps.find((step) => step.type === 'narration-row')
    expect(row && row.type === 'narration-row' && row.chain).toMatchObject({
      linkId: 'f1:e1',
      fromPlayerId: 'p5',
      toPlayerId: 'p1',
      verb: '攻击',
      card: { name: '雷光斩', element: 'Thunder' },
    })
  })

  it('uses the skill animation map for the skill opening', () => {
    bridge.onTimeline(
      payload([
        flow('f1', [{ id: 'e1', order: 1, from_user_id: 'p1', to_user_id: 'p1', phase: 'skill' } as ActionFlowEdgeDTO], {
          action_type: 'skill',
          nodes: [
            { id: 'n1', order: 1, kind: 'skill', actor_user_id: 'p1', skill_id: 'healing_light', skill_name: '治愈之光', target_user_ids: ['p1', 'p5'], anchor_edge_id: 'e1' },
            { id: 'n2', order: 2, kind: 'skill', actor_user_id: 'p1', skill_id: 'frost_prayer', skill_name: '冰霜祷言', target_user_ids: ['p1'], anchor_edge_id: 'e1' },
          ],
        }),
      ]),
    )
    const steps = queued()
    // A self skill edge is only an anchor: no "圣女 → 圣女" row and no link.
    expect(steps.some((step) => step.type === 'link')).toBe(false)
    const bubbles = steps.filter((step) => step.type === 'skill-bubble')
    expect(bubbles).toMatchObject([
      { skillName: '治愈之光', opening: 'active', effects: ['heal'], targetIds: ['p1', 'p5'] },
      { skillName: '冰霜祷言', opening: 'trigger' },
    ])
    const rows = steps.filter((step) => step.type === 'narration-row')
    expect(rows.map((row) => row.type === 'narration-row' && row.skill?.opening)).toEqual(['active', 'trigger'])
  })

  it('turns heal and status state deltas into heal numbers and flying status icons', () => {
    bridge.onTimeline(payload([flow('f1', [attackEdge()], { actor_user_id: 'p5' })]))
    const before = queued().length
    bridge.onTimeline({
      ...payload([flow('f1', [attackEdge()], { actor_user_id: 'p5' })]),
      events: [
        {
          event_id: 9,
          turn_id: 1,
          chain_id: 'c',
          type: 'TimelineEffectResolved',
          outcome: '',
          visibility: '',
          action_id: 'f1',
          deltas: [
            { type: 'heal', target_user_id: 'p1', value: 1 },
            { type: 'hand_count', target_user_id: 'p1', value: -1 },
            { type: 'status', field: 'status:Poison', target_user_id: 'p1', value: 1 },
          ],
        },
      ],
    })
    expect(queued().slice(before)).toEqual([
      { type: 'value-chip', playerId: 'p1', text: '+1', color: 'heal' },
      { type: 'status-change', playerId: 'p1', statusKey: 'Poison', added: true, fromPlayerId: 'p5' },
    ])
  })

  it('enqueues nothing new when identical action_flows are re-sent', () => {
    const flows = [flow('f1', [attackEdge()])]
    bridge.onTimeline(payload(flows))
    const before = queued().length
    bridge.onTimeline(payload(flows))
    bridge.onTimeline(payload(JSON.parse(JSON.stringify(flows))))
    expect(queued().length).toBe(before)
  })

  it('enqueues only the new edge (and resolution of the updated one) when a flow is re-sent', () => {
    bridge.onTimeline(payload([flow('f1', [attackEdge()])]))
    const before = queued().length

    const defend = attackEdge({ id: 'e2', order: 2, from_user_id: 'p1', to_user_id: 'p5', phase: 'defend', label: '圣光' })
    bridge.onTimeline(payload([flow('f1', [attackEdge(), defend])]))
    const added = queued().slice(before)

    expect(added.some((step) => step.type === 'mark-previous')).toBe(false)
    const links = added.filter((step) => step.type === 'link')
    expect(links).toHaveLength(1)
    expect(links[0]).toMatchObject({ linkId: 'f1:e2' })
    expect(added.every((step) => !('linkId' in step) || step.linkId === 'f1:e2')).toBe(true)

    // Outcome set later on the original edge: end mark, then the impact number on the target.
    const before2 = queued().length
    bridge.onTimeline(payload([flow('f1', [attackEdge({ outcome: 'hit', damage: 2 }), defend])]))
    const added2 = queued().slice(before2)
    expect(types(added2)).toEqual(['link-animate', 'link-end', 'value-chip'])
    expect(added2[1]).toMatchObject({ linkId: 'f1:e1', style: 'hit', text: '2伤害' })
    expect(added2[2]).toMatchObject({ playerId: 'p1', text: '-2', color: 'damage' })
  })

  it('folds a "take" self-resolve edge into a narration row and a damage chip, without a link', () => {
    const take = { id: 'e3', order: 3, from_user_id: 'p1', to_user_id: 'p1', phase: 'take' } as ActionFlowEdgeDTO
    bridge.onTimeline(payload([flow('f1', [attackEdge(), take])]))
    const before = queued().length
    bridge.onTimeline(payload([flow('f1', [attackEdge(), { ...take, damage: 2 }])]))
    const added = queued().slice(before)
    const all = queued()
    expect(all.filter((step) => step.type === 'link')).toHaveLength(1)
    expect(all.some((step) => step.type === 'narration-row' && step.html.includes('承受'))).toBe(true)
    expect(added).toEqual([{ type: 'value-chip', playerId: 'p1', text: '-2', color: 'damage' }])
  })

  it('enqueues mark-previous before the first step of a new flow_id', () => {
    const first = flow('f1', [attackEdge()])
    bridge.onTimeline(payload([first]))
    const before = queued().length
    bridge.onTimeline(payload([first, flow('f2', [attackEdge({ id: 'e9', from_user_id: 'p1', to_user_id: 'p5' })])]))
    const added = queued().slice(before)
    expect(added[0]).toEqual({ type: 'mark-previous' })
    expect(added[1]).toMatchObject({ type: 'link', linkId: 'f2:e9' })
    expect(added.filter((step) => step.type === 'mark-previous')).toHaveLength(1)
  })

  it('applies replays with zero holds and no transient beats', () => {
    bridge.onTimeline(
      payload(
        [
          flow('f1', [attackEdge({ outcome: 'hit', damage: 2 })], {
            nodes: [{ id: 'n1', order: 1, kind: 'skill', actor_user_id: 'p1', skill_name: '冰霜祷言', anchor_edge_id: 'e1' }],
          }),
        ],
        true,
      ),
    )
    const steps = queued()
    expect(steps.length).toBeGreaterThan(0)
    expect(steps.every((step) => step.holdMs === 0)).toBe(true)
    expect(steps.some((step) => step.type === 'skill-bubble' || step.type === 'value-chip')).toBe(false)
    expect(steps.some((step) => step.type === 'link-end')).toBe(true)

    // Draining at a single timestamp applies everything because holds are zero.
    for (let i = 0; i < steps.length; i++) store.tick(1000)
    expect(store.queue).toHaveLength(0)
    expect(store.activeLinks.get('f1:e1')?.end).toMatchObject({ style: 'hit' })
  })

  it('reset forgets flows so the same flow plays again', () => {
    const flows = [flow('f1', [attackEdge()])]
    bridge.onTimeline(payload(flows))
    bridge.reset()
    expect(store.queue).toHaveLength(0)
    bridge.onTimeline(payload(flows))
    expect(queued()[0]).toEqual({ type: 'mark-previous' })
  })
})
