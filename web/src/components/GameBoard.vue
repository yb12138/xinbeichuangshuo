<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useGameStore } from '../stores/gameStore'
import type { FieldCard, PlayerView } from '../types/game'
import PlayerArea from './PlayerArea.vue'
import ActionPanel from './ActionPanel.vue'
import CardComponent from './CardComponent.vue'
import PromptDialog from './PromptDialog.vue'
import SkillDetailModal from './SkillDetailModal.vue'
import BattleZone from './BattleZone.vue'
import VfxLayer from './VfxLayer.vue'
import ActionTimeline from './ActionTimeline.vue'
import DamageNotification from './DamageNotification.vue'
import { useWebSocket } from '../composables/useWebSocket'
const store = useGameStore()
const ws = useWebSocket()

// 我的手牌
const myHand = computed(() => store.myHand)
const myExclusiveCards = computed(() => store.myExclusiveCards)
const myHandEntries = computed(() => store.myPlayableCards.filter(item => item.source === 'hand'))
const myAreaPlayer = computed(() => store.players[store.myPlayerId] || null)
const myCoverCards = computed(() =>
  (myAreaPlayer.value?.field || []).filter(
    (fc): fc is FieldCard => !!fc && fc.mode === 'Cover' && !!fc.card
  )
)
const expansionCardCount = computed(() => myExclusiveCards.value.length + myCoverCards.value.length)
const boardRootRef = ref<HTMLElement | null>(null)
const deckCounterRef = ref<HTMLElement | null>(null)
const showExpansionCards = ref(false)

const COVER_EFFECT_LABEL: Record<string, string> = {
  MagicBowCharge: '充能',
  SpiritCasterPower: '妖力',
  MoonDarkMoon: '暗月',
  ButterflyCocoon: '茧',
}

function coverEffectLabel(effect?: string): string {
  if (!effect) return '盖牌'
  return COVER_EFFECT_LABEL[effect] || '盖牌'
}

const orderedPlayerIds = computed(() => {
  const ids: string[] = []
  const seen = new Set<string>()
  for (const p of store.roomPlayers) {
    if (store.players[p.id] && !seen.has(p.id)) {
      ids.push(p.id)
      seen.add(p.id)
    }
  }
  for (const id of Object.keys(store.players).sort()) {
    if (!seen.has(id)) {
      ids.push(id)
      seen.add(id)
    }
  }
  return ids
})

const turnOrderMap = computed(() => {
  const map: Record<string, number> = {}
  orderedPlayerIds.value.forEach((id, idx) => {
    map[id] = idx + 1
  })
  return map
})

const orderedOtherPlayers = computed(() =>
  orderedPlayerIds.value
    .filter((id) => id !== store.myPlayerId)
    .map((id) => store.players[id])
    .filter((p): p is PlayerView => !!p)
)

const currentTurnCamp = computed(() => {
  const current = store.currentPlayer ? store.players[store.currentPlayer] : undefined
  if (current?.camp === 'Red' || current?.camp === 'Blue') return current.camp
  if (store.myCamp === 'Red' || store.myCamp === 'Blue') return store.myCamp
  return 'Red'
})

const leftCamp = computed(() => (currentTurnCamp.value === 'Red' ? 'Blue' : 'Red'))
const rightCamp = computed(() => currentTurnCamp.value)
const isHostInRoom = computed(() =>
  store.roomPlayers.some(p => p.id === store.myPlayerId && p.is_host)
)
const offlinePlayers = computed(() =>
  store.roomPlayers.filter(p => !p.is_bot && p.is_online === false)
)
const canHostTakeover = computed(() => isHostInRoom.value && offlinePlayers.value.length > 0)

const leftRailPlayers = computed(() =>
  orderedOtherPlayers.value
    .filter((p) => p.camp === leftCamp.value)
    .slice(0, 3)
)
const rightRailPlayers = computed(() =>
  orderedOtherPlayers.value
    .filter((p) => p.camp === rightCamp.value)
    .slice(0, 2)
)

// 行动选择 prompt 不触发 blur（已在 ActionPanel 内联展示）
const isActionPrompt = computed(() => {
  const prompt = store.currentPrompt
  if (!prompt) return false
  if (prompt.ui_mode === 'action_hub') return true
  const msg = prompt.message ?? ''
  return msg.includes('行动类型')
})
const gameEndTitle = computed(() => {
  const msg = store.gameEndMessage || ''
  if (msg.includes('红方胜利')) return '红方胜利'
  if (msg.includes('蓝方胜利')) return '蓝方胜利'
  return '对局结束'
})
const gameEndSnapshot = computed(() => store.gameEndSnapshot)
const moraleBurstRanking = computed(() => store.moraleBurstRanking.slice(0, 8))
const moraleChangesForReview = computed(() =>
  [...store.moraleChanges].sort((a, b) => b.timestamp - a.timestamp).slice(0, 12)
)
const gameEndTriggerText = computed(() => {
  const snap = gameEndSnapshot.value
  if (!snap) return '未记录触发点'
  if (snap.triggerType === 'cups') return '星杯达到 5（资源胜利）'
  if (snap.triggerType === 'morale') return '士气归零（战斗胜利）'
  return '服务器结束事件触发'
})

function campLabel(camp?: string): string {
  return camp === 'Red' ? '红方' : camp === 'Blue' ? '蓝方' : '未知'
}

function isMagicBulletCard(cardIdx: number): boolean {
  const card = store.myPlayableCards.find(item => item.index === cardIdx)?.card
  return !!card && card.type === 'Magic' && card.name === '魔弹'
}

function moraleDeltaLabel(delta: number): string {
  return delta > 0 ? `+${delta}` : `${delta}`
}

function isPlayerSelectable(playerId: string): boolean {
  if (store.isGameEnded) return false
  if (store.isPromptForMe) return false
  if (store.canTargetOpponent() && store.targetablePlayers.some(t => t.id === playerId)) return true
  if (store.skillMode === 'choosing_target' && store.targetablePlayersForSkill.some(t => t.id === playerId)) return true
  if (store.actionMode === 'magic' && store.selectedCardForAction !== null && store.targetablePlayers.some(t => t.id === playerId)) return true
  return false
}

function onTargetClick(playerId: string) {
  if (store.isGameEnded) return
  // 中断提示期间仅允许通过 PromptDialog/ActionPanel 操作，禁止点击角色区发普通行动
  if (store.currentPrompt && store.isPromptForMe) return

  // 技能选目标模式
  if (store.skillMode === 'choosing_target' && store.selectedSkill) {
    if (store.targetablePlayersForSkill.some(p => p.id === playerId)) {
      store.toggleSkillTarget(playerId)
      // 单目标技能选中后自动发动
      if (store.selectedSkill.max_targets === 1 && store.skillTargetIds.length === 1) {
        ws.useSkill(store.selectedSkill.id, store.skillTargetIds)
      }
    }
    return
  }
  // 攻击/法术模式
  if (!store.canTargetOpponent()) return
  const cardIdx = store.selectedCardForAction
  if (cardIdx === null) return
  const selectedItem = store.myPlayableCards.find(item => item.index === cardIdx)
  if (!selectedItem) {
    store.setSelectedCardForAction(null)
    store.setError('所选卡牌已变化，请重新选择')
    return
  }
  if (store.actionMode === 'attack') {
    if (selectedItem.card.type !== 'Attack') {
      store.setSelectedCardForAction(null)
      store.setError('所选卡牌不是攻击牌，请重新选择')
      return
    }
    ws.attack(playerId, cardIdx)
  } else if (store.actionMode === 'magic') {
    if (selectedItem.card.type !== 'Magic') {
      store.setSelectedCardForAction(null)
      store.setError('所选卡牌不是法术牌，请重新选择')
      return
    }
    if (isMagicBulletCard(cardIdx)) {
      ws.magic(undefined, cardIdx)
    } else {
      ws.magic(playerId, cardIdx)
    }
  }
}

function isCardSelectableForAction(idx: number): boolean {
  if (store.isGameEnded) return false
  if (store.skillMode === 'choosing_discard') return idx < store.myHand.length
  if (store.isPromptForMe) return true
  if (store.actionMode === 'attack') {
    const card = store.myPlayableCards.find(item => item.index === idx)?.card
    return !!card && card.type === 'Attack'
  }
  if (store.actionMode === 'magic' && store.magicSubChoice === 'card') {
    const card = store.myPlayableCards.find(item => item.index === idx)?.card
    return !!card && card.type === 'Magic'
  }
  return store.isMyTurn
}

function onCardClick(idx: number) {
  if (store.isGameEnded) return
  // 优先级：actionMode > skillMode(弃牌) > prompt 选牌 > 默认
  if (store.actionMode !== 'none') {
    const card = store.myPlayableCards.find(item => item.index === idx)?.card
    if (store.actionMode === 'magic' && card && card.type !== 'Magic') {
      store.setError('请选择法术牌')
      return
    }
    if (store.actionMode === 'attack' && card && card.type !== 'Attack') {
      store.setError('请选择攻击牌')
      return
    }
    if (store.actionMode === 'magic' && isMagicBulletCard(idx)) {
      // 魔弹按固定传递顺序自动结算，不需要手动点目标。
      ws.magic(undefined, idx)
      return
    }
    store.setSelectedCardForAction(store.selectedCardForAction === idx ? null : idx)
    return
  }
  // 技能弃牌模式：检查元素要求后切换选中
  if (store.skillMode === 'choosing_discard' && store.selectedSkill) {
    const card = store.myHand[idx]
    if (!card) return
    const skill = store.selectedSkill
    // 检查元素要求
    if (skill.discard_element && card.element !== skill.discard_element) {
      store.setError(`需要弃置${skill.discard_element}牌`)
      return
    }
    if (skill.discard_type && card.type !== skill.discard_type) {
      store.setError(`需要弃置${skill.discard_type === 'Magic' ? '法术' : '攻击'}牌`)
      return
    }
    if (skill.id === 'magic_bullet_fusion' && card.element !== 'Fire' && card.element !== 'Earth') {
      store.setError('魔弹融合需要弃置1张火系或地系牌')
      return
    }
    if (skill.id === 'onmyoji_shikigami_descend' && !store.skillDiscardIndices.includes(idx)) {
      if (!card.faction) {
        store.setError('式神降临需要弃置有命格的手牌')
        return
      }
      const selected = store.skillDiscardIndices
        .map((i) => store.myHand[i])
        .filter((c): c is NonNullable<typeof c> => !!c)
      if (selected.length > 0) {
        const reqFaction = selected[0]?.faction
        if (reqFaction && card.faction !== reqFaction) {
          store.setError('式神降临需要弃置2张命格相同的手牌')
          return
        }
      }
    }
    // 检查是否已选满
    if (!store.skillDiscardIndices.includes(idx) && store.skillDiscardIndices.length >= skill.cost_discards) {
      store.setError(`最多选择 ${skill.cost_discards} 张牌`)
      return
    }
    store.toggleSkillDiscard(idx)
    return
  }
  if (store.isPromptForMe) {
    store.toggleCardSelection(idx)
    return
  }
  if (store.isMyTurn) {
    store.setSelectedCardForAction(store.selectedCardForAction === idx ? null : idx)
  }
}

function turnOrderFor(playerId: string): number | undefined {
  return turnOrderMap.value[playerId]
}

type DrawFlightVisual = {
  id: string
  startX: number
  startY: number
  deltaX: number
  deltaY: number
  delayMs: number
}

const drawFlightCards = ref<DrawFlightVisual[]>([])

function rebuildDrawFlights() {
  const root = boardRootRef.value
  const deck = deckCounterRef.value
  if (!root || !deck || store.drawBursts.length === 0) {
    drawFlightCards.value = []
    return
  }

  const rootRect = root.getBoundingClientRect()
  const deckRect = deck.getBoundingClientRect()
  const startX = deckRect.left + deckRect.width / 2 - rootRect.left
  const startY = deckRect.top + deckRect.height / 2 - rootRect.top
  const visuals: DrawFlightVisual[] = []

  for (const burst of store.drawBursts) {
    const anchor = root.querySelector<HTMLElement>(`[data-player-anchor="${burst.playerId}"]`)
    if (!anchor) continue
    const targetRect = anchor.getBoundingClientRect()
    const targetX = targetRect.left + targetRect.width / 2 - rootRect.left
    const targetY = targetRect.top + targetRect.height / 2 - rootRect.top
    const visibleCount = Math.min(6, Math.max(1, burst.count))

    for (let i = 0; i < visibleCount; i++) {
      const jitterX = (i - Math.floor(visibleCount / 2)) * 10
      const jitterY = -Math.min(16, i * 3)
      visuals.push({
        id: `${burst.id}-${i}`,
        startX,
        startY,
        deltaX: targetX - startX + jitterX,
        deltaY: targetY - startY + jitterY,
        delayMs: i * 90
      })
    }
  }

  drawFlightCards.value = visuals
}

function drawFlightStyle(card: DrawFlightVisual): Record<string, string> {
  return {
    left: `${card.startX}px`,
    top: `${card.startY}px`,
    '--draw-dx': `${card.deltaX}px`,
    '--draw-dy': `${card.deltaY}px`,
    animationDelay: `${card.delayMs}ms`
  }
}

function refreshDrawFlightsSoon() {
  nextTick(() => {
    rebuildDrawFlights()
  })
}

watch(
  () => store.drawBursts.map((item) => `${item.id}-${item.playerId}-${item.count}`).join('|'),
  () => {
    refreshDrawFlightsSoon()
  },
  { immediate: true }
)

watch(
  () => [leftRailPlayers.value.length, rightRailPlayers.value.length, !!myAreaPlayer.value, store.myPlayerId, store.deckCount],
  () => {
    if (store.drawBursts.length > 0) {
      refreshDrawFlightsSoon()
    }
  }
)

function handleResize() {
  if (store.drawBursts.length > 0) {
    rebuildDrawFlights()
  }
}

function toggleExpansionCards() {
  showExpansionCards.value = !showExpansionCards.value
}

onMounted(() => {
  window.addEventListener('resize', handleResize)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleResize)
})

function leaveToLobby() {
  ws.disconnect()
}

function takeoverOfflinePlayer(playerId: string) {
  if (!playerId) return
  ws.sendRoomAction('takeover_player', { target_id: playerId })
}

function dissolveRoomByHost() {
  if (!isHostInRoom.value) return
  const confirmed = window.confirm('确认解散房间吗？所有玩家将被退出到大厅。')
  if (!confirmed) return
  ws.sendRoomAction('dissolve_room')
}
</script>

<template>
  <div ref="boardRootRef" class="h-full w-full flex flex-col board-shell p-2 sm:p-3 md:p-4 min-h-0 relative">
    <div class="board-ambient board-ambient-left" />
    <div class="board-ambient board-ambient-right" />
    <button
      v-if="isHostInRoom"
      type="button"
      class="host-dissolve-btn"
      @click="dissolveRoomByHost"
    >
      解散房间
    </button>

    <div class="top-hud">
      <div class="camp-bar camp-blue-bar">
        <span class="camp-side-label camp-side-label-left">蓝阵营</span>
        <div class="camp-center-metrics">
          <span class="camp-score">{{ store.blueMorale }}</span>
          <span class="camp-cup">🏆 {{ store.blueCups }}</span>
          <span class="camp-gem">♦ {{ store.blueGems }}</span>
          <span class="camp-crystal">🔷 {{ store.blueCrystals }}</span>
        </div>
      </div>

      <div
        ref="deckCounterRef"
        class="top-deck-indicator"
        :class="{ 'top-deck-indicator--active': store.drawBursts.length > 0 }"
        title="当前公共牌堆剩余卡牌"
      >
        <span class="top-deck-label">公共牌堆</span>
        <span class="top-deck-count">{{ store.deckCount }}</span>
      </div>

      <div class="camp-bar camp-red-bar">
        <span class="camp-side-label camp-side-label-right">红阵营</span>
        <div class="camp-center-metrics">
          <span class="camp-score">{{ store.redMorale }}</span>
          <span class="camp-cup">🏆 {{ store.redCups }}</span>
          <span class="camp-gem">♦ {{ store.redGems }}</span>
          <span class="camp-crystal">🔷 {{ store.redCrystals }}</span>
        </div>
      </div>
    </div>

    <div v-if="offlinePlayers.length > 0" class="disconnect-panel" :class="{ 'disconnect-panel-host': canHostTakeover }">
      <div class="disconnect-title">
        {{ canHostTakeover ? '玩家离线（房主可选择托管）' : '玩家离线，等待房主处理' }}
      </div>
      <div class="disconnect-list">
        <div v-for="p in offlinePlayers" :key="`offline-${p.id}`" class="disconnect-item">
          <span class="disconnect-name">{{ p.name }} ({{ p.id }})</span>
          <button
            v-if="canHostTakeover"
            class="disconnect-takeover-btn"
            type="button"
            @click="takeoverOfflinePlayer(p.id)"
          >
            机器人接管
          </button>
        </div>
      </div>
    </div>


    <div
      class="main-grid flex-1 min-h-0 min-w-0 mt-2 arena-blur-focus"
      :class="{ 'blur-active': store.currentPrompt && store.isPromptForMe && !isActionPrompt }"
    >
      <aside class="side-rail side-rail-left">
        <div
          v-for="p in leftRailPlayers"
          :key="p.id"
          class="player-anchor-wrap"
          :data-player-anchor="p.id"
        >
          <PlayerArea
            :player="p"
            :isMe="p.id === store.myPlayerId"
            :isOpponent="p.camp !== store.myCamp"
            :selectable="isPlayerSelectable(p.id)"
            :selected="store.skillMode === 'choosing_target' && store.skillTargetIds.includes(p.id)"
            :turnOrder="turnOrderFor(p.id)"
            compact
            @select="onTargetClick"
          />
        </div>
      </aside>

      <section class="center-stage">
        <div class="stage-main">
          <div class="center-battle battle-field">
            <BattleZone class="battle-zone-fill" />
            <div class="battle-feed-float">
              <ActionTimeline />
            </div>
          </div>
        </div>

        <div class="bottom-hud flex-shrink-0 min-h-0 mt-2">
          <div class="bottom-hud-main">
            <div class="bottom-slot-me player-anchor-wrap" :data-player-anchor="store.myPlayerId">
              <PlayerArea
                v-if="myAreaPlayer"
                :player="myAreaPlayer"
                is-me
                :turnOrder="turnOrderFor(myAreaPlayer.id)"
                compact
              />
            </div>
            <div class="hand-rail bottom-slot-hand rounded-lg sm:rounded-xl p-2 sm:p-2 min-h-0">
              <div class="exclusive-toggle-row mb-2">
                <button
                  type="button"
                  class="exclusive-toggle-btn"
                  :disabled="expansionCardCount === 0"
                  @click="toggleExpansionCards"
                >
                  <span class="exclusive-toggle-title">扩展区</span>
                  <span class="exclusive-toggle-meta">
                    {{
                      expansionCardCount > 0
                        ? `专属 ${myExclusiveCards.length} ｜ 盖牌 ${myCoverCards.length}`
                        : '暂无扩展牌'
                    }}
                  </span>
                  <span v-if="expansionCardCount > 0" class="exclusive-toggle-arrow">
                    {{ showExpansionCards ? '收起 ▲' : '展开 ▼' }}
                  </span>
                </button>
              </div>
              <div v-if="showExpansionCards && expansionCardCount > 0" class="expansion-zone mb-2">
                <div class="expansion-zone-scroll">
                  <div class="expansion-zone-content">
                    <div v-if="myExclusiveCards.length > 0" class="expansion-group">
                      <div class="expansion-group-title">专属技能卡（{{ myExclusiveCards.length }}）</div>
                      <div class="expansion-card-row">
                        <CardComponent
                          v-for="(card, idx) in myExclusiveCards"
                          :key="`exclusive-${card.id || idx}`"
                          :card="card"
                          medium
                        />
                      </div>
                    </div>
                    <div v-if="myCoverCards.length > 0" class="expansion-group">
                      <div class="expansion-group-title">盖牌（{{ myCoverCards.length }}）</div>
                      <div class="expansion-card-row">
                        <div
                          v-for="(fc, idx) in myCoverCards"
                          :key="`cover-${fc.card.id || idx}`"
                          class="expansion-cover-item"
                        >
                          <CardComponent
                            :card="fc.card"
                            medium
                          />
                          <div class="expansion-cover-tag">{{ coverEffectLabel(fc.effect) }}</div>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
              <div class="overflow-x-auto hand-list pb-0.5">
                <div class="hand-card-row">
                  <CardComponent
                    v-for="entry in myHandEntries"
                    :key="entry.index"
                    :card="entry.card"
                    :index="entry.index"
                    medium
                    :selectable="isCardSelectableForAction(entry.index)"
                    :selected="store.selectedCards.includes(entry.index) || store.selectedCardForAction === entry.index || store.skillDiscardIndices.includes(entry.index)"
                    @click="onCardClick(entry.index)"
                  />
                </div>
                <div v-if="myHand.length === 0" class="text-gray-500 py-4 text-sm">没有手牌</div>
              </div>
            </div>
          </div>
        </div>
      </section>

      <aside class="side-rail side-rail-right">
        <div
          v-for="p in rightRailPlayers"
          :key="p.id"
          class="player-anchor-wrap"
          :data-player-anchor="p.id"
        >
          <PlayerArea
            :player="p"
            :isMe="p.id === store.myPlayerId"
            :isOpponent="p.camp !== store.myCamp"
            :selectable="isPlayerSelectable(p.id)"
            :selected="store.skillMode === 'choosing_target' && store.skillTargetIds.includes(p.id)"
            :turnOrder="turnOrderFor(p.id)"
            compact
            @select="onTargetClick"
          />
        </div>
      </aside>
    </div>

    <div v-if="drawFlightCards.length > 0" class="draw-flight-layer">
      <div
        v-for="card in drawFlightCards"
        :key="card.id"
        class="draw-flight-card"
        :style="drawFlightStyle(card)"
      >
        <div class="draw-flight-card-face" />
      </div>
    </div>

    <div class="right-action-dock" :class="{ 'right-action-dock--active': store.isMyTurn }">
      <ActionPanel />
    </div>

    <!-- Toast 通知（参考 noname） -->
    <Transition name="toast">
      <div 
        v-if="store.errorMessage" 
        class="toast error"
      >
        {{ store.errorMessage }}
      </div>
    </Transition>
    <Transition name="toast">
      <div 
        v-if="store.skillEffectToast" 
        class="toast skill"
      >
        {{ store.skillEffectToast }}
      </div>
    </Transition>

    <PromptDialog />

    <!-- 伤害结算通知弹框 -->
    <DamageNotification />

    <!-- 技能详情中央弹窗（所有人可查看任意角色） -->
    <SkillDetailModal
      :character="store.skillModalCharacter"
      :visible="!!store.skillModalCharacterId"
      :anchor="store.skillModalAnchor"
      @close="store.openSkillModal(null)"
    />

    <Transition name="game-end">
      <div v-if="store.isGameEnded" class="game-end-overlay">
        <div class="game-end-card">
          <div class="game-end-title">{{ gameEndTitle }}</div>
          <div class="game-end-message">{{ store.gameEndMessage || '游戏结束' }}</div>

          <div class="game-end-layout">
            <section class="game-end-summary">
              <div class="summary-title">胜利条件触发点</div>
              <div class="summary-trigger">{{ gameEndTriggerText }}</div>
              <div class="summary-source" v-if="gameEndSnapshot?.triggerSource">
                触发来源：{{ gameEndSnapshot.triggerSource }}
                <span v-if="gameEndSnapshot.triggerCamp">
                  （{{ campLabel(gameEndSnapshot.triggerCamp) }}{{ gameEndSnapshot.triggerDelta ? ` -${gameEndSnapshot.triggerDelta}` : '' }}）
                </span>
              </div>
              <div class="summary-source" v-else>
                触发来源：无明确来源记录（以服务端结算为准）
              </div>
              <div class="summary-metrics">
                <div class="metric-item">
                  <span>红方士气</span>
                  <strong>{{ gameEndSnapshot?.finalRedMorale ?? store.redMorale }}</strong>
                </div>
                <div class="metric-item">
                  <span>蓝方士气</span>
                  <strong>{{ gameEndSnapshot?.finalBlueMorale ?? store.blueMorale }}</strong>
                </div>
                <div class="metric-item">
                  <span>红方星杯</span>
                  <strong>{{ gameEndSnapshot?.finalRedCups ?? store.redCups }}</strong>
                </div>
                <div class="metric-item">
                  <span>蓝方星杯</span>
                  <strong>{{ gameEndSnapshot?.finalBlueCups ?? store.blueCups }}</strong>
                </div>
              </div>
            </section>

            <section class="game-end-review">
              <div class="review-block">
                <div class="review-title">爆士气排行（高到低）</div>
                <div v-if="moraleBurstRanking.length === 0" class="review-empty">
                  暂无可复盘的爆士气记录
                </div>
                <div v-else class="review-list">
                  <div v-for="(item, idx) in moraleBurstRanking" :key="item.id" class="review-row">
                    <span class="review-rank">#{{ idx + 1 }}</span>
                    <span class="review-camp">{{ campLabel(item.camp) }}</span>
                    <span class="review-delta">-{{ Math.abs(item.delta) }}</span>
                    <span class="review-source">{{ item.source }}</span>
                  </div>
                </div>
              </div>

              <div class="review-block">
                <div class="review-title">士气变化来源</div>
                <div v-if="moraleChangesForReview.length === 0" class="review-empty">
                  本局暂无士气变化记录
                </div>
                <div v-else class="review-list review-list-history">
                  <div v-for="item in moraleChangesForReview" :key="`history-${item.id}`" class="review-row review-row-history">
                    <span class="review-camp">{{ campLabel(item.camp) }}</span>
                    <span class="review-flow">{{ item.before }}→{{ item.after }}（{{ moraleDeltaLabel(item.delta) }}）</span>
                    <span class="review-source">{{ item.source }}</span>
                  </div>
                </div>
              </div>
            </section>
          </div>

          <button class="game-end-btn" @click="leaveToLobby">返回房间大厅</button>
        </div>
      </div>
    </Transition>
    <VfxLayer />
  </div>
</template>

<style scoped>
.game-end-enter-active,
.game-end-leave-active {
  transition: opacity 0.24s ease;
}

.game-end-enter-from,
.game-end-leave-to {
  opacity: 0;
}

.game-end-overlay {
  position: absolute;
  inset: 0;
  z-index: 80;
  display: flex;
  align-items: center;
  justify-content: center;
  background:
    radial-gradient(460px 210px at 50% 45%, rgba(209, 165, 98, 0.22), transparent 72%),
    rgba(2, 8, 18, 0.72);
  backdrop-filter: blur(4px);
}

.game-end-card {
  width: min(96vw, 920px);
  max-height: 84vh;
  overflow: hidden;
  border-radius: 16px;
  border: 1px solid rgba(181, 145, 90, 0.56);
  background:
    linear-gradient(180deg, rgba(19, 26, 41, 0.96), rgba(12, 18, 30, 0.98)),
    url('/assets/ui/modal-aura.svg') center/cover no-repeat;
  box-shadow:
    inset 0 1px 0 rgba(255, 242, 205, 0.12),
    0 22px 40px rgba(0, 0, 0, 0.52);
  padding: 22px 20px 18px;
  text-align: center;
  display: flex;
  flex-direction: column;
}

.game-end-title {
  font-family: var(--font-ui-title);
  font-size: 28px;
  line-height: 1.1;
  font-weight: 700;
  color: #ffe2ad;
  letter-spacing: 0.06em;
  text-shadow: 0 2px 10px rgba(12, 6, 2, 0.58);
}

.game-end-message {
  margin-top: 10px;
  font-size: 14px;
  color: rgba(225, 238, 251, 0.9);
  line-height: 1.5;
}

.game-end-layout {
  margin-top: 14px;
  display: grid;
  grid-template-columns: minmax(0, 0.95fr) minmax(0, 1.05fr);
  gap: 12px;
  min-height: 0;
  overflow: hidden;
}

.game-end-summary,
.game-end-review {
  border-radius: 12px;
  border: 1px solid rgba(141, 172, 192, 0.3);
  background: rgba(6, 17, 30, 0.62);
  text-align: left;
  padding: 12px;
}

.summary-title,
.review-title {
  font-size: 12px;
  letter-spacing: 0.08em;
  font-weight: 700;
  color: #bfd7e8;
  margin-bottom: 6px;
}

.summary-trigger {
  color: #f6dfb1;
  font-size: 15px;
  font-weight: 700;
  line-height: 1.35;
}

.summary-source {
  margin-top: 6px;
  font-size: 12px;
  color: #afc7d8;
  line-height: 1.4;
}

.summary-metrics {
  margin-top: 10px;
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 8px;
}

.metric-item {
  border-radius: 10px;
  border: 1px solid rgba(126, 161, 183, 0.3);
  background: rgba(8, 20, 34, 0.72);
  padding: 8px 10px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  color: #c5d8e6;
  font-size: 12px;
}

.metric-item strong {
  color: #f5d7a0;
  font-size: 16px;
  font-weight: 800;
}

.game-end-review {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.review-block {
  min-height: 0;
}

.review-empty {
  border-radius: 10px;
  border: 1px dashed rgba(130, 162, 182, 0.32);
  color: #9fb8cb;
  background: rgba(7, 18, 32, 0.56);
  padding: 10px;
  font-size: 12px;
}

.review-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
  max-height: 132px;
  overflow: auto;
  padding-right: 4px;
}

.review-list-history {
  max-height: 164px;
}

.review-row {
  border-radius: 9px;
  border: 1px solid rgba(120, 155, 176, 0.28);
  background: rgba(9, 21, 35, 0.68);
  padding: 7px 8px;
  display: grid;
  grid-template-columns: auto auto auto minmax(0, 1fr);
  gap: 8px;
  align-items: center;
  font-size: 12px;
  color: #d7e6f1;
}

.review-row-history {
  grid-template-columns: auto auto minmax(0, 1fr);
}

.review-rank {
  color: #ffdfab;
  font-weight: 700;
}

.review-camp {
  color: #b9d1e3;
  font-weight: 600;
}

.review-delta {
  color: #ffb4a5;
  font-weight: 700;
}

.review-flow {
  color: #d5e6f3;
  font-weight: 600;
}

.review-source {
  color: #a9c1d2;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.game-end-btn {
  margin-top: 18px;
  width: 100%;
  height: 40px;
  border-radius: 10px;
  border: 1px solid rgba(212, 163, 90, 0.52);
  background: linear-gradient(140deg, rgba(157, 106, 44, 0.92), rgba(116, 73, 28, 0.96));
  color: #fff4de;
  font-size: 14px;
  font-weight: 600;
  letter-spacing: 0.02em;
  transition: transform 0.14s ease, box-shadow 0.14s ease, filter 0.14s ease;
}

@media (max-width: 900px) {
  .game-end-card {
    width: min(96vw, 640px);
    max-height: 88vh;
  }

  .game-end-layout {
    grid-template-columns: 1fr;
    overflow: auto;
    padding-right: 2px;
  }

  .review-list,
  .review-list-history {
    max-height: none;
  }
}

.game-end-btn:hover {
  transform: translateY(-1px);
  box-shadow: 0 10px 18px rgba(102, 61, 23, 0.36);
  filter: brightness(1.03);
}

.game-end-btn:active {
  transform: translateY(0);
}

.toast-enter-active,
.toast-leave-active {
  transition: transform 0.28s ease, opacity 0.28s ease;
}
.toast-enter-from,
.toast-leave-to {
  transform: translate(-50%, 38px);
  opacity: 0;
}

.board-shell {
  width: 100%;
  max-width: 1760px;
  margin: 0 auto;
  overflow: hidden;
  position: relative;
  padding-top: max(8px, var(--safe-top));
  padding-bottom: calc(8px + var(--safe-bottom));
  background: transparent;
  border: none;
  box-shadow: none;
}

.board-shell::before {
  content: '';
  position: absolute;
  inset: 0;
  background:
    radial-gradient(960px 460px at 50% 40%, rgba(42, 86, 132, 0.18), rgba(16, 27, 42, 0.34) 58%, rgba(8, 14, 24, 0.58) 100%),
    linear-gradient(180deg, rgba(8, 16, 28, 0.42), rgba(5, 10, 18, 0.56));
  z-index: 0;
}

.board-shell::after {
  content: '';
  position: absolute;
  inset: 0;
  background:
    radial-gradient(980px 420px at 50% 42%, rgba(120, 195, 219, 0.16), transparent 72%),
    linear-gradient(180deg, rgba(2, 10, 20, 0.2), rgba(2, 7, 16, 0.46));
  pointer-events: none;
  z-index: 0;
}

.board-shell > * {
  position: relative;
  z-index: 2;
}

.board-ambient {
  position: absolute;
  pointer-events: none;
  border-radius: 9999px;
  filter: blur(34px);
  opacity: 0.36;
  z-index: 1;
}

.board-ambient-left {
  width: 210px;
  height: 210px;
  left: -84px;
  top: 24%;
  background: rgba(106, 182, 188, 0.18);
}

.board-ambient-right {
  width: 230px;
  height: 230px;
  right: -104px;
  top: 10%;
  background: rgba(213, 168, 104, 0.16);
}

.host-dissolve-btn {
  position: absolute;
  top: max(8px, var(--safe-top));
  right: 10px;
  z-index: 9;
  height: 30px;
  padding: 0 10px;
  border-radius: 999px;
  border: 1px solid rgba(226, 136, 136, 0.52);
  background: linear-gradient(135deg, rgba(138, 51, 51, 0.92), rgba(92, 29, 29, 0.95));
  color: #ffe7e7;
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.02em;
  box-shadow: 0 8px 18px rgba(12, 3, 3, 0.45);
}

.host-dissolve-btn:hover {
  filter: brightness(1.08);
}

.top-hud {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto minmax(0, 1fr);
  align-items: center;
  gap: 10px;
  margin-bottom: 8px;
}

.disconnect-panel {
  border-radius: 10px;
  border: 1px solid rgba(186, 132, 121, 0.38);
  background: rgba(42, 18, 18, 0.54);
  padding: 6px 10px;
  margin: 0 0 8px;
}

.disconnect-panel-host {
  border-color: rgba(196, 158, 108, 0.46);
  background: rgba(44, 30, 16, 0.56);
}

.disconnect-title {
  font-size: 12px;
  color: #f4d7ac;
  font-weight: 700;
}

.disconnect-list {
  margin-top: 4px;
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.disconnect-item {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  border-radius: 999px;
  border: 1px solid rgba(194, 145, 132, 0.34);
  background: rgba(58, 25, 24, 0.46);
  padding: 2px 7px;
}

.disconnect-name {
  font-size: 11px;
  color: #f6dbd3;
}

.disconnect-takeover-btn {
  border-radius: 999px;
  border: 1px solid rgba(127, 177, 208, 0.5);
  background: rgba(17, 52, 76, 0.8);
  color: #d7ecfa;
  font-size: 10px;
  font-weight: 700;
  padding: 1px 7px;
}

.top-deck-indicator {
  min-width: 92px;
  height: 44px;
  border-radius: 999px;
  border: 1px solid rgba(152, 183, 201, 0.52);
  background: linear-gradient(138deg, rgba(14, 34, 53, 0.9), rgba(8, 20, 33, 0.92));
  box-shadow:
    inset 0 1px 0 rgba(242, 250, 255, 0.1),
    0 8px 20px rgba(3, 10, 20, 0.34);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 0 10px;
  white-space: nowrap;
}

.top-deck-indicator--active {
  box-shadow:
    inset 0 1px 0 rgba(242, 250, 255, 0.12),
    0 10px 24px rgba(3, 10, 20, 0.42),
    0 0 0 1px rgba(236, 203, 140, 0.34);
  animation: deckPulse 0.52s ease-out;
}

.top-deck-label {
  font-size: 11px;
  letter-spacing: 0.05em;
  color: rgba(181, 209, 226, 0.9);
}

.top-deck-count {
  font-family: var(--font-ui-title);
  font-size: 24px;
  font-weight: 800;
  line-height: 1;
  color: #f4e4c5;
  text-shadow: 0 1px 3px rgba(2, 7, 17, 0.62);
}

.draw-flight-layer {
  position: absolute;
  inset: 0;
  z-index: 28;
  pointer-events: none;
  overflow: hidden;
}

.draw-flight-card {
  position: absolute;
  width: 30px;
  height: 42px;
  margin-left: -15px;
  margin-top: -21px;
  animation: drawCardFlight 0.95s cubic-bezier(0.22, 0.61, 0.36, 1) forwards;
}

.draw-flight-card-face {
  width: 100%;
  height: 100%;
  border-radius: 7px;
  border: 1px solid rgba(229, 197, 137, 0.72);
  background:
    linear-gradient(145deg, rgba(97, 64, 32, 0.95), rgba(66, 43, 22, 0.96)),
    repeating-linear-gradient(40deg, rgba(229, 197, 137, 0.24) 0 3px, rgba(229, 197, 137, 0) 3px 7px);
  box-shadow:
    0 8px 20px rgba(5, 10, 20, 0.5),
    inset 0 1px 0 rgba(255, 243, 219, 0.24);
}

@keyframes drawCardFlight {
  0% {
    transform: translate(0, 0) scale(0.78) rotate(-6deg);
    opacity: 0;
  }
  18% {
    opacity: 1;
  }
  100% {
    transform: translate(var(--draw-dx), var(--draw-dy)) scale(0.96) rotate(0deg);
    opacity: 0;
  }
}

@keyframes deckPulse {
  0% {
    transform: scale(0.92);
    filter: brightness(0.88);
  }
  55% {
    transform: scale(1.04);
    filter: brightness(1.08);
  }
  100% {
    transform: scale(1);
    filter: brightness(1);
  }
}

.camp-bar {
  height: 46px;
  border-radius: 999px;
  border: 1px solid rgba(143, 176, 195, 0.45);
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0 12px;
  box-shadow:
    inset 0 1px 0 rgba(247, 252, 255, 0.14),
    0 8px 20px rgba(2, 8, 20, 0.32);
  overflow: hidden;
}

.camp-bar::before {
  content: '';
  position: absolute;
  inset: 0;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.14), rgba(255, 255, 255, 0) 45%);
  pointer-events: none;
}

.camp-center-metrics {
  display: flex;
  align-items: center;
  gap: 10px;
  justify-content: center;
  min-width: 0;
  z-index: 1;
}

.camp-side-label {
  position: absolute;
  top: 50%;
  transform: translateY(-50%);
  font-family: var(--font-ui-title);
  font-size: 14px;
  font-weight: 700;
  letter-spacing: 0.1em;
  opacity: 0.95;
  white-space: nowrap;
  pointer-events: none;
  text-shadow: 0 1px 3px rgba(4, 9, 17, 0.8);
}

.camp-side-label-left {
  left: 14px;
}

.camp-side-label-right {
  right: 14px;
}

.camp-score {
  font-family: var(--font-ui-title);
  font-size: 30px;
  font-weight: 800;
  line-height: 1;
  color: #f8fbff;
  min-width: 28px;
  text-align: center;
  text-shadow: 0 1px 5px rgba(2, 8, 18, 0.64);
}

.camp-gem,
.camp-crystal,
.camp-cup {
  border-radius: 999px;
  border: 1px solid rgba(151, 181, 200, 0.42);
  font-size: 11px;
  font-weight: 700;
  padding: 3px 8px;
  line-height: 1;
  white-space: nowrap;
  background: rgba(5, 13, 23, 0.42);
  box-shadow: inset 0 1px 0 rgba(240, 247, 252, 0.08);
}

.camp-gem {
  color: #f4b3ab;
}

.camp-crystal {
  color: #acd7ef;
}

.camp-cup {
  color: #f8dd96;
}

.camp-red-bar {
  background: linear-gradient(132deg, rgba(112, 35, 31, 0.82), rgba(79, 27, 24, 0.86));
  color: #f8d4ce;
  border-color: rgba(198, 103, 93, 0.54);
  box-shadow:
    inset 0 1px 0 rgba(255, 200, 190, 0.12),
    0 8px 20px rgba(80, 15, 10, 0.25);
}

.camp-blue-bar {
  background: linear-gradient(132deg, rgba(17, 60, 96, 0.84), rgba(13, 42, 68, 0.88));
  color: #d9edfa;
  border-color: rgba(106, 168, 205, 0.54);
  box-shadow:
    inset 0 1px 0 rgba(180, 220, 255, 0.12),
    0 8px 20px rgba(5, 25, 50, 0.25);
}

.main-grid {
  display: grid;
  grid-template-columns: 144px minmax(0, 1fr) 144px;
  gap: 12px;
}

@media (min-width: 1600px) {
  .main-grid {
    grid-template-columns: 168px minmax(0, 1fr) 168px;
    gap: 16px;
  }

  .bottom-hud {
    --me-slot-width: 158px;
    --hand-max-width: 920px;
  }
}

@media (min-width: 2000px) {
  .board-shell {
    max-width: 2080px;
  }

  .main-grid {
    grid-template-columns: 196px minmax(0, 1fr) 196px;
    gap: 18px;
  }

  .bottom-hud {
    --me-slot-width: 186px;
    --hand-max-width: 1020px;
  }

  .hand-rail {
    max-width: 1020px;
  }
}

.side-rail {
  min-height: 0;
  min-width: 0;
  display: flex;
  flex-direction: column;
  justify-content: flex-start;
  gap: 8px;
}

@media (min-width: 901px) {
  /* 侧边栏中的角色卡占满整列宽度，避免比 rail 更窄。 */
  .side-rail :deep(.player-area) {
    width: 100%;
    min-width: 100%;
    max-width: 100%;
  }
}

.player-anchor-wrap {
  width: 100%;
}

.side-rail-left {
  align-items: flex-start;
}

.side-rail-right {
  align-items: flex-end;
}

.center-stage {
  min-height: 0;
  min-width: 0;
  position: relative;
  border-radius: 14px;
  border: none;
  background: transparent;
  box-shadow: none;
  padding: 2px 0 0;
  display: flex;
  flex-direction: column;
  overflow: visible;
}

.table-decor {
  position: absolute;
  left: 50%;
  pointer-events: none;
  z-index: 0;
}

.table-decor-base {
  width: min(98%, 1220px);
  height: clamp(310px, 58vh, 610px);
  top: clamp(34px, 6.2vh, 82px);
  transform: translateX(-50%);
  background:
    radial-gradient(110% 76% at 50% 54%, rgba(76, 120, 168, 0.34), rgba(38, 59, 92, 0.22) 58%, rgba(14, 21, 32, 0.15) 100%);
  filter:
    drop-shadow(0 18px 34px rgba(2, 10, 16, 0.62))
    drop-shadow(0 0 16px rgba(133, 178, 214, 0.14));
  opacity: 0.98;
}

.table-decor-edge {
  width: min(90%, 1020px);
  height: clamp(54px, 8vh, 86px);
  bottom: clamp(150px, 20.5vh, 254px);
  transform: translateX(-50%);
  background:
    linear-gradient(180deg, rgba(111, 141, 168, 0.64), rgba(51, 67, 88, 0.84));
  filter: drop-shadow(0 10px 20px rgba(1, 8, 14, 0.6));
  opacity: 0.96;
}

.stage-main {
  flex: 1 1 auto;
  min-height: 0;
  min-width: 0;
  display: flex;
  align-items: stretch;
  position: relative;
  overflow: hidden;
  border-radius: 14px;
  border: 1px solid rgba(120, 165, 210, 0.08);
  background:
    linear-gradient(180deg, rgba(12, 25, 42, 0.2), rgba(8, 18, 32, 0.35));
  z-index: 2;
  padding: 0;
  box-shadow: inset 0 1px 0 rgba(200, 230, 255, 0.04);
}

.stage-main::before {
  content: none;
}

.stage-main::after {
  content: none;
}

.stage-main > * {
  position: relative;
  z-index: 2;
}

.battle-zone-fill {
  flex: 1 1 auto;
  width: 100%;
  height: 100%;
  min-height: 0;
  min-width: 0;
}

.center-battle {
  flex: 1 1 auto;
  width: 100%;
  height: 100%;
  min-height: 0;
  min-width: 0;
  display: flex;
  align-items: stretch;
  justify-content: center;
  position: relative;
  border-radius: 14px;
  border: 1px solid rgba(100, 145, 195, 0.1);
  background:
    radial-gradient(ellipse 90% 80% at 50% 50%, rgba(25, 55, 90, 0.15), transparent 65%);
  box-shadow:
    inset 0 0 40px rgba(40, 80, 130, 0.06),
    0 0 20px rgba(20, 50, 90, 0.08);
}

.battle-field {
  position: relative;
  overflow: hidden;
}

.battle-feed-float {
  position: absolute;
  top: 8px;
  left: 10px;
  right: auto;
  width: fit-content;
  max-width: calc(100% - 20px);
  z-index: 8;
  overflow: visible;
}

.battle-feed-float :deep(.timeline-strip-wrap) {
  width: fit-content;
  max-width: 100%;
  min-height: 0;
}

.hand-rail {
  flex: 1 1 560px;
  min-width: 280px;
  max-width: var(--hand-max-width);
  position: relative;
  background:
    linear-gradient(180deg, rgba(12, 26, 42, 0.92), rgba(8, 18, 31, 0.95));
  border: 1px solid rgba(130, 170, 210, 0.15);
  border-radius: 12px;
  box-shadow:
    inset 0 1px 0 rgba(235, 245, 252, 0.1),
    inset 0 0 20px rgba(20, 50, 90, 0.08),
    0 8px 24px rgba(1, 8, 16, 0.4);
}

.hand-list {
  width: 100%;
  min-width: 0;
  scrollbar-width: thin;
  scrollbar-color: rgba(94, 138, 165, 0.74) rgba(7, 14, 22, 0.45);
}

.hand-card-row {
  display: flex;
  align-items: flex-end;
  width: max-content;
  min-width: 100%;
  gap: 6px;
  padding-right: 2px;
}

.exclusive-toggle-btn {
  width: 100%;
  border: 1px solid rgba(130, 170, 210, 0.2);
  border-radius: 10px;
  background: linear-gradient(180deg, rgba(16, 32, 48, 0.72), rgba(11, 22, 34, 0.86));
  color: #dce7f5;
  padding: 8px 10px;
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 12px;
  transition: border-color 0.16s ease, background 0.16s ease;
}

.exclusive-toggle-btn:not(:disabled):hover {
  border-color: rgba(180, 210, 239, 0.52);
  background: linear-gradient(180deg, rgba(22, 44, 66, 0.76), rgba(12, 25, 39, 0.9));
}

.exclusive-toggle-btn:disabled {
  opacity: 0.6;
  cursor: default;
}

.exclusive-toggle-title {
  color: rgba(255, 226, 156, 0.94);
  font-weight: 600;
}

.exclusive-toggle-meta {
  color: rgba(191, 214, 236, 0.8);
  margin-left: auto;
}

.exclusive-toggle-arrow {
  color: rgba(244, 226, 175, 0.95);
  min-width: 56px;
  text-align: right;
}

.expansion-zone {
  max-width: 100%;
  min-width: 0;
}

.expansion-zone-scroll {
  width: fit-content;
  max-width: 100%;
  min-width: 0;
  overflow-x: auto;
  overflow-y: hidden;
  padding-bottom: 2px;
  scrollbar-width: thin;
  scrollbar-color: rgba(94, 138, 165, 0.74) rgba(7, 14, 22, 0.45);
}

.expansion-zone-content {
  display: inline-flex;
  flex-direction: column;
  gap: 8px;
  width: max-content;
  min-width: max-content;
}

.expansion-group {
  width: max-content;
  min-width: 0;
}

.expansion-group-title {
  font-size: 12px;
  line-height: 1.2;
  color: rgba(255, 236, 189, 0.9);
  letter-spacing: 0.4px;
  margin-bottom: 4px;
}

.expansion-card-row {
  display: flex;
  align-items: flex-end;
  gap: 6px;
  width: max-content;
  min-width: 0;
  padding-right: 2px;
}

.expansion-cover-item {
  position: relative;
}

.expansion-cover-tag {
  position: absolute;
  right: 6px;
  bottom: 6px;
  font-size: 10px;
  line-height: 1;
  padding: 4px 6px;
  border-radius: 999px;
  background: rgba(4, 12, 22, 0.82);
  border: 1px solid rgba(174, 213, 252, 0.42);
  color: rgba(214, 232, 255, 0.95);
  pointer-events: none;
}

.bottom-hud {
  padding-top: 4px;
  display: flex;
  flex-direction: column;
  align-items: stretch;
  width: 100%;
  gap: 8px;
  position: relative;
  z-index: 2;
  --me-slot-width: 142px;
  --hand-max-width: 840px;
  --hud-main-gap: 8px;
}

.bottom-hud-main {
  width: min(100%, calc(var(--me-slot-width) + var(--hand-max-width) + var(--hud-main-gap)));
  min-width: 0;
  display: grid;
  grid-template-columns: var(--me-slot-width) minmax(0, 1fr);
  align-items: end;
  column-gap: var(--hud-main-gap);
  margin: 0;
}

.bottom-slot-me {
  flex-shrink: 0;
}

.bottom-slot-me {
  width: var(--me-slot-width);
  justify-self: start;
}

.bottom-slot-me :deep(.player-area) {
  width: 100%;
  min-width: 100% !important;
  max-width: 100% !important;
}

.bottom-slot-hand {
  width: 100%;
  max-width: min(100%, var(--hand-max-width));
  min-width: 0;
  justify-self: stretch;
}

.right-action-dock {
  position: absolute;
  right: max(10px, var(--safe-right));
  bottom: calc(12px + var(--safe-bottom));
  width: clamp(250px, 18vw, 320px);
  z-index: 24;
  pointer-events: auto;
  transition: filter 0.22s ease, transform 0.22s ease;
}

.right-action-dock--active {
  filter: drop-shadow(0 8px 22px rgba(6, 30, 43, 0.42));
  transform: translateY(-2px);
}

@media (max-width: 1200px) {
  .right-action-dock {
    width: clamp(198px, 19vw, 248px);
  }
}

@media (max-width: 900px) {
  .right-action-dock {
    position: fixed;
    right: max(8px, var(--safe-right));
    bottom: calc(10px + var(--safe-bottom));
    width: min(198px, 46vw);
    z-index: 36;
  }
}

@media (max-width: 640px) {
  .right-action-dock {
    width: min(176px, 48vw);
  }
}

@media (min-width: 640px) {
  .bottom-hud {
    --me-slot-width: 142px;
    --hand-max-width: 860px;
  }
}

/* 针对 1440x678 这类“宽屏但高度较矮”的桌面，放大角色位宽度，避免立绘被 rail 压缩。 */
@media (min-width: 1360px) and (max-width: 1599px) and (max-height: 760px) {
  .main-grid {
    grid-template-columns: 168px minmax(0, 1fr) 168px;
    gap: 14px;
  }

  .side-rail {
    gap: 10px;
  }

  .bottom-hud {
    --me-slot-width: 162px;
    --hand-max-width: 760px;
  }

  .right-action-dock {
    width: clamp(270px, 20vw, 330px);
  }
}

@media (max-width: 1024px) {
  .main-grid {
    grid-template-columns: 132px minmax(0, 1fr) 132px;
  }

  .side-rail {
    gap: 6px;
  }

  .bottom-hud {
    --me-slot-width: 132px;
    --hand-max-width: 700px;
  }

  .hand-rail {
    max-width: 700px;
  }

  .table-decor-base {
    width: min(104%, 1040px);
    height: clamp(280px, 54vh, 520px);
    top: clamp(24px, 5vh, 60px);
  }

  .table-decor-edge {
    width: min(94%, 900px);
    bottom: clamp(138px, 20vh, 214px);
  }
}

@media (max-width: 1024px) and (orientation: landscape) and (pointer: coarse) {
  .board-shell {
    width: 100%;
    max-width: none;
    min-height: var(--app-vh);
    height: var(--app-vh);
    overflow: hidden;
    padding-top: max(4px, var(--safe-top));
    padding-right: max(6px, var(--safe-right));
    padding-bottom: calc(4px + var(--safe-bottom));
    padding-left: max(6px, var(--safe-left));
    border-left: none;
    border-right: none;
    border-radius: 0;
  }

  .top-hud {
    margin-bottom: 4px;
    gap: 6px;
  }

  .top-deck-indicator {
    min-width: 84px;
    height: 36px;
    gap: 5px;
    padding: 0 8px;
  }

  .top-deck-label {
    font-size: 10px;
  }

  .top-deck-count {
    font-size: 20px;
  }

  .draw-flight-card {
    width: 26px;
    height: 36px;
    margin-left: -13px;
    margin-top: -18px;
  }

  .camp-bar {
    height: 38px;
    padding: 0 8px;
  }

  .camp-side-label {
    font-size: 11px;
  }

  .camp-score {
    font-size: 23px;
  }

  .camp-center-metrics {
    gap: 4px;
  }

  .camp-gem,
  .camp-crystal,
  .camp-cup {
    font-size: 10px;
    padding: 2px 5px;
  }

  .main-grid {
    grid-template-columns: 124px minmax(0, 1fr) 124px;
    gap: 6px;
  }

  .side-rail {
    gap: 4px;
  }

  .center-stage {
    padding: 0;
    border-radius: 10px;
  }

  .stage-main {
    min-height: 0;
  }

  .table-decor-base {
    width: min(104%, 920px);
    height: clamp(238px, 58vh, 420px);
    top: clamp(18px, 4.2vh, 42px);
  }

  .table-decor-edge {
    width: min(92%, 760px);
    bottom: clamp(118px, 18vh, 176px);
  }

  .bottom-hud {
    --me-slot-width: 124px;
    --hand-max-width: 100%;
    --hud-main-gap: 6px;
  }

  .bottom-slot-hand {
    width: 100%;
  }

  .hand-rail {
    min-width: 0;
    max-width: none;
  }
}

@media (max-width: 760px) and (orientation: landscape) and (pointer: coarse) {
  .main-grid {
    grid-template-columns: 108px minmax(0, 1fr) 108px;
  }

  .camp-side-label {
    display: none;
  }

  .table-decor-base {
    width: min(106%, 780px);
    height: clamp(210px, 62vh, 360px);
    top: clamp(16px, 4vh, 34px);
  }

  .table-decor-edge {
    width: min(94%, 640px);
    bottom: clamp(100px, 18vh, 154px);
  }

  .bottom-hud {
    --me-slot-width: 108px;
  }
}

@media (max-width: 900px) and (orientation: portrait) {
  .board-shell {
    overflow-y: auto;
    overflow-x: hidden;
    -webkit-overflow-scrolling: touch;
    overscroll-behavior: contain;
    padding-bottom: calc(10px + var(--safe-bottom));
  }

  .top-hud {
    position: sticky;
    top: 0;
    z-index: 10;
    margin-bottom: 6px;
    padding: 2px 0;
    backdrop-filter: blur(4px);
  }

  .main-grid {
    grid-template-columns: 1fr;
    gap: 8px;
  }

  .side-rail {
    flex-direction: row;
    overflow-x: auto;
    gap: 6px;
    padding-bottom: 2px;
    scrollbar-width: thin;
    scroll-snap-type: x proximity;
  }

  .side-rail :deep(.player-area) {
    scroll-snap-align: start;
  }

  .side-rail-left,
  .side-rail-right {
    align-items: stretch;
    justify-content: flex-start;
  }

  .center-stage {
    width: 100%;
  }

  .stage-main {
    min-height: clamp(300px, 44vh, 520px);
  }

  .table-decor-base {
    width: min(108%, 920px);
    height: clamp(310px, 46vh, 520px);
    top: clamp(44px, 9vh, 96px);
  }

  .table-decor-edge {
    width: min(98%, 720px);
    bottom: clamp(188px, 24vh, 310px);
  }

  .bottom-hud {
    width: 100%;
    gap: 6px;
    --me-slot-width: 128px;
    --hand-max-width: 100%;
    --hud-main-gap: 6px;
  }

  .bottom-hud-main {
    grid-template-columns: var(--me-slot-width) minmax(0, 1fr);
  }

  .bottom-slot-me {
    width: var(--me-slot-width);
  }

  .hand-rail {
    flex: 1 1 auto;
    max-width: none;
  }
}

@media (max-width: 640px) {
  .board-shell {
    border-left: none;
    border-right: none;
    border-radius: 0;
    box-shadow: none;
  }

  .top-hud {
    gap: 5px;
  }

  .top-deck-indicator {
    min-width: 74px;
    height: 34px;
    gap: 4px;
    padding: 0 7px;
  }

  .top-deck-label {
    display: none;
  }

  .top-deck-count {
    font-size: 20px;
  }

  .bottom-hud {
    --hand-max-width: 100%;
  }

  .table-decor-base {
    width: min(112%, 720px);
    height: clamp(280px, 42vh, 420px);
    top: clamp(38px, 8.8vh, 72px);
  }

  .table-decor-edge {
    width: min(100%, 580px);
    bottom: clamp(178px, 23vh, 254px);
    opacity: 0.9;
  }

  .camp-bar {
    height: 38px;
    padding: 0 8px;
  }

  .camp-side-label {
    font-size: 10px;
  }

  .camp-side-label-left {
    left: 9px;
  }

  .camp-side-label-right {
    right: 9px;
  }


  .camp-score {
    font-size: 22px;
  }

  .camp-center-metrics {
    gap: 4px;
  }

  .camp-gem,
  .camp-crystal,
  .camp-cup {
    font-size: 10px;
    padding: 2px 4px;
  }
}

@media (max-width: 480px) {
  .camp-side-label {
    display: none;
  }

  .camp-center-metrics {
    width: 100%;
    justify-content: space-between;
    gap: 4px;
  }

  .camp-score {
    min-width: 24px;
    font-size: 20px;
  }

  .camp-gem,
  .camp-crystal,
  .camp-cup {
    font-size: 9px;
    padding: 2px 3px;
  }
}

.arena-blur-focus {
  transition: filter 0.3s ease;
}

.arena-blur-focus.blur-active {
  filter: blur(2px) brightness(0.85);
  pointer-events: none;
}

.side-rail .player-anchor-wrap {
  transition: transform 0.2s ease;
}

.side-rail .player-anchor-wrap:hover {
  transform: translateY(-1px);
}
</style>
