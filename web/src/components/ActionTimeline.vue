<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useGameStore } from '../stores/gameStore'
import type { BattleFeedType } from '../stores/gameStore'

const store = useGameStore()
const showHistory = ref(false)
const nowTs = ref(Date.now())
const waitingSinceTs = ref(0)
let timer: ReturnType<typeof setInterval> | null = null

const recentFeed = computed(() => {
  const rows = store.battleFeed.slice(-30)
  return [...rows].reverse()
})

const latestEntry = computed(() => store.battleFeed[store.battleFeed.length - 1] || null)

const waitingBotName = computed(() => {
  if (!store.waitingFor) return ''
  const p = store.roomPlayers.find(player => player.id === store.waitingFor)
  if (!p || !p.is_bot) return ''
  return p.name || p.id
})

watch(waitingBotName, (next, prev) => {
  if (next && next !== prev) {
    waitingSinceTs.value = Date.now()
  }
  if (!next) {
    waitingSinceTs.value = 0
  }
}, { immediate: true })

const waitingSeconds = computed(() => {
  if (!waitingBotName.value || waitingSinceTs.value <= 0) return 0
  return Math.max(0, Math.floor((nowTs.value - waitingSinceTs.value) / 1000))
})

const currentLine = computed(() => {
  if (waitingBotName.value) {
    const tail = waitingSeconds.value >= 2 ? ` ${waitingSeconds.value}s` : ''
    return `${waitingBotName.value} 正在思考下一步…${tail}`
  }
  if (!latestEntry.value) return '等待第一条战斗事件…'
  const { title, detail } = latestEntry.value
  return detail ? `${title} · ${detail}` : title
})

const currentType = computed<BattleFeedType>(() => {
  if (waitingBotName.value) return 'system'
  return latestEntry.value?.type || 'system'
})

function iconByType(type: BattleFeedType) {
  switch (type) {
    case 'turn':
      return '⏳'
    case 'skill':
      return '✨'
    case 'attack':
      return '⚔'
    case 'magic':
      return '✦'
    case 'respond':
      return '🛡'
    case 'damage':
      return '💥'
    case 'resource':
      return '♦'
    default:
      return '·'
  }
}

function classByType(type: BattleFeedType) {
  switch (type) {
    case 'turn':
      return 'row-turn'
    case 'skill':
      return 'row-skill'
    case 'attack':
      return 'row-attack'
    case 'magic':
      return 'row-magic'
    case 'respond':
      return 'row-respond'
    case 'damage':
      return 'row-damage'
    case 'resource':
      return 'row-resource'
    default:
      return 'row-system'
  }
}

function timeLabel(timestamp: number) {
  const diffSec = Math.max(0, Math.floor((nowTs.value - timestamp) / 1000))
  if (diffSec <= 0) return '刚刚'
  if (diffSec < 60) return `${diffSec}s`
  const m = Math.floor(diffSec / 60)
  const s = diffSec % 60
  return `${m}m${s}s`
}

onMounted(() => {
  timer = setInterval(() => {
    nowTs.value = Date.now()
  }, 1000)
})

onBeforeUnmount(() => {
  if (timer) clearInterval(timer)
})
</script>

<template>
  <div class="timeline-strip-wrap">
    <div class="timeline-strip" :class="classByType(currentType)">
      <div class="line-label">战斗播报</div>
      <div class="line-icon">{{ iconByType(currentType) }}</div>
      <div class="line-text">{{ currentLine }}</div>
      <button class="history-btn" type="button" @click="showHistory = !showHistory">
        {{ showHistory ? '收起历史' : '查看历史' }} ({{ store.battleFeed.length }})
      </button>
    </div>

    <Transition name="history">
      <div v-if="showHistory" class="history-panel">
        <div class="history-head">
          <div class="history-title">战斗历史</div>
          <label class="history-switch">
            <input
              :checked="store.cinematicMode"
              type="checkbox"
              @change="store.setCinematicMode(($event.target as HTMLInputElement).checked)"
            />
            <span>慢放演出</span>
          </label>
        </div>

        <div class="history-list">
          <div
            v-for="item in recentFeed"
            :key="item.id"
            class="history-row"
            :class="classByType(item.type)"
          >
            <span class="row-icon">{{ iconByType(item.type) }}</span>
            <span class="row-main">{{ item.title }}</span>
            <span class="row-time">{{ timeLabel(item.timestamp) }}</span>
          </div>
          <div v-if="recentFeed.length === 0" class="history-empty">暂无历史事件</div>
        </div>
      </div>
    </Transition>
  </div>
</template>

<style scoped>
.timeline-strip-wrap {
  display: inline-flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 6px;
  width: fit-content;
  max-width: min(92vw, 940px);
  border: none;
  background: none;
  padding: 0;
  box-shadow: none;
}

.timeline-strip {
  border-radius: 9px;
  border: 1px solid rgba(117, 152, 173, 0.38);
  background: rgba(10, 22, 36, 0.78);
  min-height: 32px;
  display: grid;
  grid-template-columns: auto 18px minmax(128px, 1fr) auto;
  align-items: center;
  gap: 8px;
  width: max-content;
  max-width: 100%;
  padding: 4px 8px;
  backdrop-filter: blur(3px);
}

.line-label {
  font-size: 10px;
  line-height: 1;
  color: #f0cf98;
  letter-spacing: 0.04em;
  white-space: nowrap;
  padding: 3px 6px;
  border-radius: 6px;
  border: 1px solid rgba(206, 170, 118, 0.34);
  background: rgba(51, 36, 19, 0.44);
}

.line-icon {
  font-size: 12px;
  text-align: center;
  color: #dcebf6;
}

.line-text {
  min-width: 0;
  font-size: 12px;
  color: #e3eef8;
  line-height: 1.25;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: min(56vw, 640px);
}

.history-btn {
  border: 1px solid rgba(136, 170, 190, 0.36);
  background: rgba(24, 47, 67, 0.56);
  color: #d8ebf7;
  border-radius: 7px;
  font-size: 10px;
  line-height: 1;
  white-space: nowrap;
  padding: 4px 7px;
  cursor: pointer;
  transition: background-color 0.2s ease, transform 0.2s ease;
}

.history-btn:hover {
  transform: translateY(-1px);
  background: rgba(34, 58, 80, 0.7);
}

.history-panel {
  margin-top: 2px;
  border-radius: 10px;
  border: 1px solid rgba(120, 155, 176, 0.3);
  background: rgba(6, 15, 26, 0.9);
  padding: 8px;
  width: min(86vw, 620px);
  box-shadow: 0 12px 28px rgba(2, 8, 20, 0.42);
}

.history-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 6px;
  gap: 10px;
}

.history-title {
  font-size: 12px;
  font-weight: 700;
  color: #e6f0f8;
}

.history-switch {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 11px;
  color: #a4b9cb;
}

.history-switch input {
  accent-color: #89bdc5;
}

.history-list {
  max-height: 186px;
  overflow: auto;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.history-row {
  border: 1px solid rgba(123, 154, 176, 0.25);
  border-radius: 8px;
  padding: 4px 7px;
  display: grid;
  grid-template-columns: 16px 1fr auto;
  align-items: center;
  gap: 7px;
}

.row-icon {
  font-size: 11px;
  text-align: center;
}

.row-main {
  min-width: 0;
  font-size: 11px;
  color: #dceaf6;
  white-space: pre-line;
  overflow: visible;
  line-height: 1.35;
}

.row-time {
  font-size: 10px;
  color: #95acc2;
}

.history-empty {
  text-align: center;
  color: #86a0b8;
  font-size: 11px;
  padding: 8px 0;
}

.row-turn {
  border-color: rgba(219, 178, 112, 0.38);
  background: rgba(95, 67, 35, 0.18);
}

.row-skill {
  border-color: rgba(133, 200, 177, 0.36);
  background: rgba(18, 74, 66, 0.17);
}

.row-attack {
  border-color: rgba(220, 123, 112, 0.36);
  background: rgba(88, 30, 28, 0.18);
}

.row-magic {
  border-color: rgba(127, 182, 215, 0.36);
  background: rgba(23, 54, 80, 0.16);
}

.row-respond {
  border-color: rgba(132, 188, 205, 0.34);
  background: rgba(17, 49, 67, 0.15);
}

.row-damage {
  border-color: rgba(212, 143, 100, 0.38);
  background: rgba(92, 44, 25, 0.17);
}

.row-resource {
  border-color: rgba(130, 198, 175, 0.35);
  background: rgba(19, 72, 56, 0.15);
}

.row-system {
  border-color: rgba(129, 156, 178, 0.34);
  background: rgba(24, 41, 58, 0.2);
}

.history-enter-active,
.history-leave-active {
  transition: opacity 0.2s ease, transform 0.2s ease;
}

.history-enter-from,
.history-leave-to {
  opacity: 0;
  transform: translateY(-5px);
}

@media (max-width: 900px) {
  .history-list {
    max-height: 128px;
  }
}

@media (max-width: 640px) {
  .timeline-strip {
    min-height: 30px;
    grid-template-columns: auto 16px minmax(72px, 1fr) auto;
    gap: 6px;
    padding: 4px 6px;
  }

  .history-btn {
    padding: 3px 6px;
  }

  .line-text {
    max-width: min(42vw, 280px);
  }

  .history-panel {
    width: min(94vw, 560px);
  }
}
</style>
