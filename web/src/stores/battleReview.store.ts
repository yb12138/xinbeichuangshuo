import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

export type BattleFeedType =
  | 'turn'
  | 'skill'
  | 'attack'
  | 'magic'
  | 'respond'
  | 'damage'
  | 'resource'
  | 'system'

export interface BattleFeedEntry {
  id: number
  type: BattleFeedType
  title: string
  detail?: string
  actorId?: string
  actorName?: string
  targetId?: string
  targetName?: string
  timestamp: number
}

export type MoraleCamp = 'Red' | 'Blue'

export interface MoraleHint {
  id: number
  timestamp: number
  source: string
  raw: string
  camp?: MoraleCamp
  loss?: number
  actorName?: string
}

export interface MoraleChangeEntry {
  id: number
  timestamp: number
  camp: MoraleCamp
  delta: number
  before: number
  after: number
  source: string
  raw?: string
}

export const useBattleReviewStore = defineStore('battleReview', () => {
  const actionSummaryLines = ref<string[]>([])
  const battleFeed = ref<BattleFeedEntry[]>([])
  const moraleHints = ref<MoraleHint[]>([])
  const moraleChanges = ref<MoraleChangeEntry[]>([])
  const logs = ref<string[]>([])

  let battleFeedId = 0
  let moraleHintId = 0
  let moraleChangeId = 0

  const moraleBurstRanking = computed(() => {
    return moraleChanges.value
      .filter(item => item.delta < 0)
      .slice()
      .sort((a, b) => {
        const byLoss = Math.abs(b.delta) - Math.abs(a.delta)
        if (byLoss !== 0) return byLoss
        return b.timestamp - a.timestamp
      })
  })

  function addLog(message: string) {
    logs.value.push(message)
    if (logs.value.length > 100) {
      logs.value = logs.value.slice(-100)
    }
  }

  function clearLogs() {
    logs.value = []
  }

  function addActionStep(line: string) {
    if (!line) return
    actionSummaryLines.value.push(line)
    if (actionSummaryLines.value.length > 12) {
      actionSummaryLines.value = actionSummaryLines.value.slice(-12)
    }
  }

  function clearActionSummary() {
    actionSummaryLines.value = []
  }

  function addBattleFeed(entry: Omit<BattleFeedEntry, 'id' | 'timestamp'>) {
    const now = Date.now()
    const last = battleFeed.value[battleFeed.value.length - 1]
    if (last && last.title === entry.title && last.detail === entry.detail && now - last.timestamp < 280) {
      return
    }

    battleFeedId++
    battleFeed.value.push({
      id: battleFeedId,
      timestamp: now,
      ...entry,
    })
    if (battleFeed.value.length > 80) {
      battleFeed.value = battleFeed.value.slice(-80)
    }
  }

  function clearBattleFeed() {
    battleFeed.value = []
  }

  function pushMoraleHint(hint: Omit<MoraleHint, 'id' | 'timestamp'>) {
    moraleHintId++
    moraleHints.value.push({
      id: moraleHintId,
      timestamp: Date.now(),
      ...hint,
    })
    if (moraleHints.value.length > 30) {
      moraleHints.value = moraleHints.value.slice(-30)
    }
  }

  function consumeMoraleHint(camp: MoraleCamp, expectedLoss?: number): MoraleHint | null {
    const now = Date.now()
    moraleHints.value = moraleHints.value.filter(hint => now - hint.timestamp <= 20000)

    for (let index = moraleHints.value.length - 1; index >= 0; index--) {
      const hint = moraleHints.value[index]
      if (!hint) continue
      const campMatch = !hint.camp || hint.camp === camp
      const lossMatch = expectedLoss === undefined || !hint.loss || hint.loss === expectedLoss
      if (campMatch && lossMatch) {
        moraleHints.value.splice(index, 1)
        return hint
      }
    }
    return null
  }

  function recordMoraleChange(
    camp: MoraleCamp,
    before: number,
    after: number,
    hint?: MoraleHint | null
  ) {
    if (before === after) return
    moraleChangeId++
    const delta = after - before
    moraleChanges.value.push({
      id: moraleChangeId,
      timestamp: Date.now(),
      camp,
      delta,
      before,
      after,
      source: hint?.source || (delta < 0 ? '未知来源（扣士气）' : '未知来源（恢复士气）'),
      raw: hint?.raw,
    })
    if (moraleChanges.value.length > 120) {
      moraleChanges.value = moraleChanges.value.slice(-120)
    }
  }

  function clearMoraleTracking() {
    moraleHints.value = []
    moraleChanges.value = []
    moraleHintId = 0
    moraleChangeId = 0
  }

  function reset() {
    clearLogs()
    clearActionSummary()
    clearBattleFeed()
    clearMoraleTracking()
    battleFeedId = 0
  }

  return {
    actionSummaryLines,
    battleFeed,
    moraleHints,
    moraleChanges,
    logs,
    moraleBurstRanking,
    addLog,
    clearLogs,
    addActionStep,
    clearActionSummary,
    addBattleFeed,
    clearBattleFeed,
    pushMoraleHint,
    consumeMoraleHint,
    recordMoraleChange,
    clearMoraleTracking,
    reset,
  }
})
