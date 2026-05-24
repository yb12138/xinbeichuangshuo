import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import type { TimelineEvent, TimelineNotifyPayload } from '../network/protocol'

export type TimelineFeedType =
  | 'turn'
  | 'skill'
  | 'attack'
  | 'magic'
  | 'respond'
  | 'damage'
  | 'resource'
  | 'system'

export interface TimelineFeedEntry {
  id: string
  eventId: number
  type: TimelineFeedType
  title: string
  actorUserId?: string
  targetUserIds: string[]
  phase?: string
  actionType?: string
  gameplayType?: string
  timestamp: number
  rawEvent: TimelineEvent
}

export const useTimelineStore = defineStore('timeline', () => {
  const payloads = ref<TimelineNotifyPayload[]>([])
  const entries = ref<TimelineFeedEntry[]>([])

  const lastSeq = computed(() => {
    const tail = payloads.value[payloads.value.length - 1]
    return tail?.seq_end ?? 0
  })

  const latestEntry = computed(() => entries.value[entries.value.length - 1] ?? null)

  const historyCount = computed(() => entries.value.length)

  function normalizeActionType(actionType?: string) {
    return String(actionType || '').trim().toLowerCase()
  }

  function campLabel(camp?: string) {
    if (camp === 'Red') return '红方'
    if (camp === 'Blue') return '蓝方'
    return camp || ''
  }

  function signed(value?: number) {
    const numeric = Number(value || 0)
    return numeric > 0 ? `+${numeric}` : String(numeric)
  }

  function firstDeltaTitle(event: TimelineEvent): string {
    const delta = event.deltas?.[0]
    if (!delta) return '状态变化'
    const actor = event.actor_name || delta.target_user_id || campLabel(delta.camp) || '状态'
    const fieldCardName = delta.field_card?.card?.name || delta.after_text || delta.before_text || '场上牌'
    switch (delta.type) {
      case 'morale':
        return `${campLabel(delta.camp)}士气 ${signed(delta.value)}`
      case 'team_gem':
        return `${campLabel(delta.camp)}阵营宝石 ${signed(delta.value)}`
      case 'team_crystal':
        return `${campLabel(delta.camp)}阵营水晶 ${signed(delta.value)}`
      case 'team_cup':
        return `${campLabel(delta.camp)}星杯 ${signed(delta.value)}`
      case 'player_gem':
        return `${actor} 宝石 ${signed(delta.value)}`
      case 'player_crystal':
        return `${actor} 水晶 ${signed(delta.value)}`
      case 'heal':
        return `${actor} 治疗 ${signed(delta.value)}`
      case 'hand_count':
        return `${actor} 手牌数 ${signed(delta.value)}`
      case 'discard_count':
        return `弃牌堆 ${signed(delta.value)}`
      case 'field_card_added':
        return `${actor} 放置 ${fieldCardName}`
      case 'field_card_removed':
        return `${actor} 移除 ${fieldCardName}`
      case 'form':
        return `${actor} 形态变化`
      case 'orientation':
        return `${actor} 朝向变化`
      default:
        return `${actor} ${delta.type} ${signed(delta.value)}`
    }
  }

  function buildFeedType(event: TimelineEvent): TimelineFeedType {
    const actionType = normalizeActionType(event.action_type)
    if (event.gameplay_type === 'skill_activated') return 'skill'
    if (event.gameplay_type === 'special_action') return 'resource'
    if (event.gameplay_type === 'state_delta') {
      const firstType = event.deltas?.[0]?.type || ''
      if (firstType === 'morale') return 'damage'
      if (firstType.includes('gem') || firstType.includes('crystal') || firstType.includes('cup') || firstType.endsWith('_count')) return 'resource'
      return 'system'
    }
    switch (event.type) {
      case 'TimelineCombatResolved':
        return 'damage'
      case 'TimelineResponseSelected':
        return 'respond'
      case 'TimelineActionDeclared':
        if (actionType === 'attack') return 'attack'
        if (actionType === 'magic') return 'magic'
        if (actionType === 'skill') return 'skill'
        if (actionType === 'respond' || actionType === 'counter' || actionType === 'defend') return 'respond'
        return event.gameplay_type === 'combat_cue' ? 'attack' : 'system'
      case 'TimelineEffectResolved':
        if (event.gameplay_type === 'draw_cards') return 'resource'
        if (event.gameplay_type === 'damage_dealt') return 'damage'
        if (event.gameplay_type === 'card_revealed') {
          if (actionType === 'attack') return 'attack'
          if (actionType === 'magic') return 'magic'
          if (actionType === 'skill') return 'skill'
          if (actionType === 'respond' || actionType === 'counter' || actionType === 'defend') return 'respond'
        }
        return 'system'
      default:
        if (event.gameplay_type === 'draw_cards') return 'resource'
        return 'system'
    }
  }

  function buildFeedTitle(event: TimelineEvent): string {
    const message = String(event.message || '').trim()
    if (message && event.gameplay_type !== 'state_delta') return message

    switch (event.gameplay_type) {
      case 'prompt':
        return '等待玩家行动'
      case 'card_revealed':
        return '亮出卡牌'
      case 'damage_dealt':
        return '伤害结算'
      case 'action_step':
        return '行动步骤'
      case 'combat_cue':
        return '进入战斗交互'
      case 'draw_cards':
        return '摸牌'
      case 'skill_activated':
        return `${event.actor_name || '玩家'} 发动「${event.skill_name || event.skill_id || '技能'}」`
      case 'special_action':
        return event.summary || `${event.actor_name || '玩家'} 执行特殊行动`
      case 'state_delta':
        return firstDeltaTitle(event)
      case 'game_end':
        return '游戏结束'
      case 'chat':
        return '聊天消息'
      default:
        return event.type || '时间线事件'
    }
  }

  function shouldDisplayTimelineEvent(event: TimelineEvent): boolean {
    switch (event.gameplay_type) {
      case 'log':
        return false
      case 'action_step':
        return event.detail_kind === 'summary'
      case 'card_revealed':
        return true
      case 'draw_cards':
        return event.reason !== 'damage_draw'
      case 'skill_activated':
      case 'special_action':
        return true
      case 'state_delta':
        return (event.deltas?.length || 0) > 0
      case 'game_end':
        return true
      default:
        return false
    }
  }

  function push(payload: TimelineNotifyPayload) {
    payloads.value.push(payload)
    if (payloads.value.length > 80) {
      payloads.value = payloads.value.slice(-80)
    }

    const timestamp = Date.now()
    for (const event of payload.events || []) {
      if (!shouldDisplayTimelineEvent(event)) {
        continue
      }
      entries.value.push({
        id: `${payload.room_id}:${event.event_id}`,
        eventId: event.event_id,
        type: buildFeedType(event),
        title: buildFeedTitle(event),
        actorUserId: event.actor_user_id,
        targetUserIds: [...(event.target_user_ids || [])],
        phase: event.cue_phase,
        actionType: event.action_type,
        gameplayType: event.gameplay_type,
        timestamp,
        rawEvent: event,
      })
    }
    if (entries.value.length > 120) {
      entries.value = entries.value.slice(-120)
    }
  }

  function clear() {
    payloads.value = []
    entries.value = []
  }

  return {
    payloads,
    entries,
    lastSeq,
    latestEntry,
    historyCount,
    push,
    clear,
  }
})
