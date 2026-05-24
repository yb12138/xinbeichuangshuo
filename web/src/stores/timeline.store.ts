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

  function buildFeedType(event: TimelineEvent): TimelineFeedType {
    const actionType = normalizeActionType(event.action_type)
    if (event.gameplay_type === 'log' && isSkillTimelineLog(event.message)) {
      return 'skill'
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
    if (message) return message

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
      case 'game_end':
        return '游戏结束'
      case 'chat':
        return '聊天消息'
      default:
        return event.type || '时间线事件'
    }
  }

  function isSkillTimelineLog(message?: string): boolean {
    const text = String(message || '').trim()
    if (!text) return false
    if (/^\[(Debug|System|Damage|Draw|Interrupt|Warn|Action)\]/.test(text)) return false
    if (/^\[Skill\]\s*.+?\s*使用了技能[:：]/.test(text)) return false
    return (
      /^.+?\s*发动\s*\[[^\]]+\]/.test(text) ||
      /^.+?\s*的\s*\[[^\]]+\]\s*生效/.test(text) ||
      /^\[Skill\]\s*.+?\s*发动\s*\[[^\]]+\]/.test(text)
    )
  }

  function shouldDisplayTimelineEvent(event: TimelineEvent): boolean {
    switch (event.gameplay_type) {
      case 'log':
        return isSkillTimelineLog(event.message)
      case 'action_step':
        return event.detail_kind === 'summary'
      case 'card_revealed':
        return true
      case 'draw_cards':
        return event.reason !== 'damage_draw'
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
