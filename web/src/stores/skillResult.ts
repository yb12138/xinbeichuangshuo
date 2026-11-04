/**
 * 技能节点结果文案：把服务端挂在技能因果上的公开状态差值（deltas）与战斗语义结果（effects）
 * 压成几条短语，例如 消耗「宝石 1」、结果「蛹 +1 · 茧 ×4」「伤害 +1」「代为应战 → 精灵射手」。
 *
 * 规则：
 *   - 技能持有者自己的宝石 / 水晶减少视为“消耗”，与效果分开展示。
 *   - 同一玩家同一字段的多次变化合并为净值；净值为 0 的不展示。
 *   - 盖牌（茧、充能……）按放置 / 移除张数计，用 ×N / -N 表示。
 *   - 手牌数量单看是过程噪声；但技能让手牌在玩家之间流动（天使祝福：目标交牌给天使、
 *     水之神力：给队友牌）或让别人弃牌 / 摸牌时，这正是技能的结果，见 handTransfers。
 *     持有者自己因代价弃掉的牌（公开展示的那张）不算结果。
 */
import type { ActionFlowEffectDTO, ActionFlowNodeDTO, TimelineDelta } from '../types/generated'
import { COVER_MARKS, STATUS_MARKS, TOKEN_MARKS } from '../constants/battleMarks'
import { PERSISTENT_EFFECTS } from '../constants/persistentEffects'
import { formLabel } from '../constants/forms'

export interface SkillResultItem {
  text: string
  /** 'transfer' = 手牌流动（战斗链里单独画成头像行，不进结果短语）。 */
  kind?: 'transfer'
  /** 结果作用的玩家；持有者自身时为空（前端不必再写名字）。 */
  targetId?: string
  tone: 'gain' | 'loss' | 'neutral'
}

export interface SkillResult {
  costs: SkillResultItem[]
  effects: SkillResultItem[]
  /** 手牌流动：from → to 若干张（天使祝福每名交牌人一条）。 */
  transfers: HandTransfer[]
}

/** 一次手牌流动。from 为空 = 从牌库摸入；to 为空 = 弃入弃牌堆。 */
export interface HandTransfer {
  from: string
  to: string
  count: number
}

const PLAYER_FIELD_LABEL: Record<string, string> = {
  gem: '宝石',
  crystal: '水晶',
  heal: '治疗',
}

const TEAM_FIELD_LABEL: Record<string, string> = {
  morale: '士气',
  cup: '星杯',
  gem: '阵营宝石',
  crystal: '阵营水晶',
}

const EFFECT_LABEL = {
  attack_redirected: '攻击改向',
  attack_taken_over: '代为应战',
  damage_prevented: '伤害抵消',
} as const

function signed(value: number): string {
  return value > 0 ? `+${value}` : `${value}`
}

function markLabel(key: string): string {
  return TOKEN_MARKS[key]?.label || STATUS_MARKS[key]?.label || COVER_MARKS[key]?.label || ''
}

interface Bucket {
  label: string
  targetId: string
  camp: string
  value: number
  kind: 'count' | 'cover' | 'status' | 'form' | 'effect'
  text?: string
}

/** 合并同一（玩家 / 阵营, 字段）的变化。 */
function collectBuckets(deltas: TimelineDelta[]): Bucket[] {
  const order: string[] = []
  const buckets = new Map<string, Bucket>()
  const bump = (key: string, init: Omit<Bucket, 'value'>, delta: number) => {
    let bucket = buckets.get(key)
    if (!bucket) {
      bucket = { ...init, value: 0 }
      buckets.set(key, bucket)
      order.push(key)
    }
    bucket.value += delta
  }

  for (const delta of deltas) {
    const target = delta.target_user_id || ''
    const value = delta.value ?? 0
    const field = delta.field || ''
    switch (delta.type) {
      case 'player_gem':
      case 'player_crystal':
      case 'heal': {
        const label = PLAYER_FIELD_LABEL[field]
        if (label && value) bump(`p:${target}:${field}`, { label, targetId: target, camp: '', kind: 'count' }, value)
        break
      }
      case 'morale':
      case 'team_cup':
      case 'team_gem':
      case 'team_crystal': {
        const label = TEAM_FIELD_LABEL[field]
        if (label && value) bump(`t:${delta.camp}:${field}`, { label, targetId: '', camp: delta.camp || '', kind: 'count' }, value)
        break
      }
      case 'token': {
        const key = field.slice(field.indexOf(':') + 1)
        // 持续效果（凋零、额外回合……）：开始 / 结束本身就是结果，写成“凋零生效 / 凋零结束”。
        const effect = PERSISTENT_EFFECTS[key]
        if (effect && value) {
          bump(`p:${target}:effect:${key}`, { label: effect.label, targetId: target, camp: '', kind: 'effect' }, value)
          break
        }
        const label = markLabel(key)
        if (label && value) bump(`p:${target}:token:${key}`, { label, targetId: target, camp: '', kind: 'count' }, value)
        break
      }
      case 'status': {
        const key = field.startsWith('status:') ? field.slice('status:'.length) : field
        const label = markLabel(key)
        if (label && value) bump(`p:${target}:status:${key}`, { label, targetId: target, camp: '', kind: 'status' }, value)
        break
      }
      case 'field_card_added':
      case 'field_card_removed': {
        const card = delta.field_card
        const key = card?.effect || ''
        const label = markLabel(key) || card?.card?.name || ''
        if (!label) break
        const kind = card?.mode === 'Cover' ? 'cover' : 'status'
        bump(`p:${target}:field:${key || label}`, { label, targetId: target, camp: '', kind }, delta.type === 'field_card_added' ? 1 : -1)
        break
      }
      case 'form': {
        // 进入（after_text 为新形态）或脱离（after_text 为空）形态；同一因果内多次变化取最终形态。
        if (delta.after_text || delta.before_text) {
          const key = `p:${target}:form`
          buckets.set(key, { label: '形态', targetId: target, camp: '', kind: 'form', value: 1, text: delta.after_text || '' })
          if (!order.includes(key)) order.push(key)
        }
        break
      }
      default:
        break
    }
  }
  return order.map((key) => buckets.get(key)!).filter((b) => b.kind === 'form' || b.value !== 0)
}

function bucketText(bucket: Bucket): string {
  switch (bucket.kind) {
    case 'cover':
      return bucket.value > 0 ? `${bucket.label} ×${bucket.value}` : `${bucket.label} ${bucket.value}`
    case 'status':
      return bucket.value > 0 ? `${bucket.label}` : `移除${bucket.label}`
    case 'form':
      return bucket.text ? `进入${formLabel(bucket.text) || '新形态'}` : '脱离形态'
    case 'effect':
      return bucket.value > 0 ? `${bucket.label}生效` : `${bucket.label}结束`
    default:
      return `${bucket.label} ${signed(bucket.value)}`
  }
}

function effectItem(effect: ActionFlowEffectDTO, nameFn: (id: string) => string): SkillResultItem | null {
  const ids = effect.target_ids || []
  switch (effect.kind) {
    case 'damage_modified':
      return effect.summary ? { text: effect.summary, targetId: ids[0], tone: effect.summary.includes('+') ? 'loss' : 'gain' } : null
    case 'damage_prevented':
      return { text: EFFECT_LABEL[effect.kind], targetId: ids[0], tone: 'gain' }
    case 'attack_redirected':
    case 'attack_taken_over': {
      const to = ids[ids.length - 1]
      const toName = to ? nameFn(to) : ''
      return { text: toName ? `${EFFECT_LABEL[effect.kind]} → ${toName}` : EFFECT_LABEL[effect.kind], tone: 'neutral' }
    }
    default:
      return null
  }
}

/**
 * 把同一因果里各玩家的手牌数变化配对成「谁 → 谁 N 张」。
 *   - 每名玩家分别累计增加与减少（不取净值：天使先弃 1 张代价、再收 2 张，净值 +1 会丢信息）。
 *   - 牌库减少 = 摸牌；其余增加都来自别人手牌。减少里流向别人手牌的张数 = 来自别人的增加，
 *     剩下的进了弃牌堆。流向手牌的优先取非持有者的减少（持有者的减少多是代价弃牌）。
 *   - 持有者自己的弃牌（代价，已作为展示牌画在节点上）与摸牌不报；只有持有者参与时整体不报。
 */
export function handTransfers(node: ActionFlowNodeDTO): HandTransfer[] {
  const owner = node.actor_user_id || ''
  const gains = new Map<string, number>()
  const losses = new Map<string, number>()
  const order: string[] = []
  let drawn = 0
  for (const delta of node.deltas || []) {
    const change = delta.value ?? ((delta.after ?? 0) - (delta.before ?? 0))
    if (delta.type === 'deck_count') {
      if (change < 0) drawn -= change
      continue
    }
    if (delta.type !== 'hand_count' || !change) continue
    const id = delta.target_user_id || ''
    if (!id) continue
    if (!gains.has(id) && !losses.has(id)) order.push(id)
    const bucket = change > 0 ? gains : losses
    bucket.set(id, (bucket.get(id) ?? 0) + Math.abs(change))
  }
  if (!order.length || order.every((id) => id === owner)) return []

  const totalGain = [...gains.values()].reduce((sum, n) => sum + n, 0)
  let handToHand = Math.max(0, totalGain - drawn)

  // 交出方：非持有者优先；接收方：持有者优先（天使祝福、水之神力最常见的方向）
  const givers = order
    .filter((id) => losses.has(id))
    .sort((a, b) => Number(a === owner) - Number(b === owner))
    .map((id) => ({ id, left: losses.get(id)! }))
  const takers = order
    .filter((id) => gains.has(id))
    .sort((a, b) => Number(b === owner) - Number(a === owner))
    .map((id) => ({ id, left: gains.get(id)! }))

  const transfers: HandTransfer[] = []
  for (const giver of givers) {
    for (const taker of takers) {
      if (!giver.left || !handToHand) break
      if (!taker.left || taker.id === giver.id) continue
      const count = Math.min(giver.left, taker.left, handToHand)
      transfers.push({ from: giver.id, to: taker.id, count })
      giver.left -= count
      taker.left -= count
      handToHand -= count
    }
  }
  for (const giver of givers) if (giver.left && giver.id !== owner) transfers.push({ from: giver.id, to: '', count: giver.left })
  for (const taker of takers) if (taker.left && taker.id !== owner) transfers.push({ from: '', to: taker.id, count: taker.left })
  return transfers
}

function transferText(transfer: HandTransfer, owner: string, nameFn: (id: string) => string): string {
  const name = (id: string) => (id === owner ? '' : nameFn(id))
  if (transfer.from && transfer.to) {
    if (transfer.to === owner) return `${nameFn(transfer.from)} 交来 ${transfer.count} 张`
    if (transfer.from === owner) return `交给 ${nameFn(transfer.to)} ${transfer.count} 张`
    return `${nameFn(transfer.from)} → ${nameFn(transfer.to)} ${transfer.count} 张`
  }
  if (transfer.from) return `${name(transfer.from) ? `${name(transfer.from)} ` : ''}弃 ${transfer.count} 张`
  return `${name(transfer.to) ? `${name(transfer.to)} ` : ''}摸 ${transfer.count} 张`
}

/**
 * @param nameFn 玩家显示名（用于“代为应战 → X”）
 */
export function skillResult(node: ActionFlowNodeDTO, nameFn: (id: string) => string = (id) => id): SkillResult {
  const owner = node.actor_user_id || ''
  const costs: SkillResultItem[] = []
  const effects: SkillResultItem[] = []

  for (const bucket of collectBuckets(node.deltas || [])) {
    const isOwner = bucket.targetId === owner
    const item: SkillResultItem = {
      text: bucketText(bucket),
      targetId: isOwner || !bucket.targetId ? undefined : bucket.targetId,
      tone: bucket.value > 0 ? 'gain' : bucket.value < 0 ? 'loss' : 'neutral',
    }
    // 持有者自己的宝石 / 水晶减少 = 发动代价。
    if (isOwner && bucket.kind === 'count' && bucket.value < 0 && (bucket.label === '宝石' || bucket.label === '水晶')) {
      costs.push({ ...item, text: `${bucket.label} ${-bucket.value}`, tone: 'neutral' })
      continue
    }
    effects.push(item)
  }
  for (const effect of node.effects || []) {
    const item = effectItem(effect, nameFn)
    if (item) effects.push(item)
  }
  const transfers = handTransfers(node)
  for (const transfer of transfers) {
    const target = transfer.from === owner ? transfer.to : transfer.from || transfer.to
    effects.push({
      kind: 'transfer',
      text: transferText(transfer, owner, nameFn),
      targetId: target && target !== owner ? target : undefined,
      tone: transfer.to === owner ? 'gain' : transfer.from === owner || !transfer.to ? 'loss' : 'neutral',
    })
  }
  return { costs, effects, transfers }
}

/** 单行摘要：「蛹 +1 · 茧 ×4」；无结果时为空串。 */
export function skillResultSummary(result: SkillResult, options: { skipTransfers?: boolean } = {}): string {
  return result.effects
    .filter((item) => !(options.skipTransfers && item.kind === 'transfer'))
    .map((item) => item.text)
    .join(' · ')
}
