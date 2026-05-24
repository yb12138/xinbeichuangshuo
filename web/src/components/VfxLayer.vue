<template>
  <div class="vfx-layer pointer-events-none overflow-visible" style="position: absolute !important; inset: 0 !important; width: 100% !important; height: 100% !important; z-index: 9999;">

    <!-- Explosions and Damage Numbers -->
    <div v-for="exp in explosions" :key="'exp'+exp.id" class="absolute explosion-container" :style="{ left: exp.x + 'px', top: exp.y + 'px' }">
      <div class="explosion-effect text-6xl">💥</div>
      <div v-if="exp.damage" class="damage-number font-black text-red-500 drop-shadow-[0_0_8px_rgba(255,0,0,0.8)]">
        -{{ exp.damage }}
      </div>
    </div>

    <!-- Flying Cards / battlefield reveal cards -->
    <div
      v-for="fc in displayCards"
      :key="fc.id"
      class="absolute flex flex-col items-center"
      :class="[
        fc.persistentBattleReveal ? 'pointer-events-auto cursor-zoom-in battlefield-reveal-card' : 'pointer-events-none',
        fc.revealState === 'featured' ? 'battlefield-reveal-card--featured' : 'battlefield-reveal-card--settled',
      ]"
      :style="{
        left: fc.x + 'px',
        top: fc.y + 'px',
        transform: fc.transform,
        opacity: fc.opacity,
        transition: `transform ${fc.duration}ms cubic-bezier(0.2, 0.8, 0.2, 1), left ${fc.duration}ms cubic-bezier(0.2, 0.8, 0.2, 1), top ${fc.duration}ms cubic-bezier(0.2, 0.8, 0.2, 1), opacity ${fc.duration}ms linear`,
        zIndex: fc.revealState === 'featured' ? 10025 : 10000
      }"
      @click.stop="openPreview(fc)"
    >
      <div
        v-if="fc.persistentBattleReveal"
        class="battlefield-reveal-owner"
        :class="`battlefield-reveal-owner--${fc.revealSide || 'bottom'}`"
      >
        {{ fc.playerName }}
      </div>
      <div class="relative flex battlefield-reveal-stack">
        <div v-for="(c, cidx) in fc.cards" :key="cidx" class="relative" :style="{ marginLeft: cidx > 0 ? '-24px' : '0' }">
          <CardComponent
            :card="c"
            :face-down="fc.hidden"
            :small="fc.revealState !== 'featured'"
            class="shadow-2xl"
          />
        </div>
      </div>
    </div>

    <Teleport to="body">
      <div v-if="previewCard" class="battlefield-preview-overlay" @click="closePreview">
        <div class="battlefield-preview-panel" @click.stop>
          <div class="battlefield-preview-meta">
            <div class="battlefield-preview-title">{{ previewCard.playerName }}</div>
            <div class="battlefield-preview-subtitle">
              {{ previewCard.hidden ? '暗牌' : '公开牌' }} · {{ actionLabel(previewCard.actionType) }}
            </div>
          </div>
          <div class="battlefield-preview-card-wrap">
            <CardComponent
              v-if="previewCard.cards[0]"
              :card="previewCard.cards[0]"
              :face-down="previewCard.hidden"
              class="battlefield-preview-card"
            />
          </div>
          <button type="button" class="battlefield-preview-close" @click="closePreview">关闭</button>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { computed, onBeforeUnmount, ref, watch, nextTick } from 'vue'
import { useBattleFxStore } from '../stores/battlefx.store'
import CardComponent from './CardComponent.vue'
import type { Card } from '../types/game'

const battleFxStore = useBattleFxStore()
const { flyingCards, combatCue, damageEffects, battlefieldRevealClearToken } = storeToRefs(battleFxStore)

interface Explosion {
  id: number
  x: number
  y: number
  damage?: number
}
const explosions = ref<Explosion[]>([])
let expIdCounter = 0

function spawnExplosion(x: number, y: number, damage?: number) {
  const id = ++expIdCounter
  explosions.value.push({ id, x, y, damage })
  setTimeout(() => {
    explosions.value = explosions.value.filter(e => e.id !== id)
  }, 800)
}

interface FlyingCardEntity {
  id: number
  cards: Card[]
  hidden?: boolean
  actionType: string
  playerId: string
  playerName: string
  x: number
  y: number
  transform: string
  opacity: number
  duration: number
  isRemoving?: boolean
  persistentBattleReveal?: boolean
  revealState?: 'featured' | 'settled'
  revealSide?: RevealSide
  targetOffsetX?: number
  targetOffsetY?: number
}

const displayCards = ref<FlyingCardEntity[]>([])
const previewCardId = ref<number | null>(null)
const previewCard = computed(() => displayCards.value.find((item) => item.id === previewCardId.value) ?? null)
const featuredSettleTimers = new Map<number, ReturnType<typeof setTimeout>>()

type RevealSide = 'top' | 'right' | 'bottom' | 'left'

const FEATURED_CARD_WIDTH = 132
const FEATURED_CARD_HEIGHT = 198
const SETTLED_CARD_WIDTH = 92
const SETTLED_CARD_HEIGHT = 138

function clamp(value: number, min: number, max: number) {
  return Math.min(Math.max(value, min), max)
}

function getBoardRect() {
  return document.querySelector('.board-shell')?.getBoundingClientRect() ?? null
}

function getBattleRectRelativeToBoard() {
  const boardRect = getBoardRect()
  const battleEl = document.querySelector('.center-battle') || document.querySelector('.battle-zone-fill')
  if (!boardRect || !battleEl) return null
  const battleRect = battleEl.getBoundingClientRect()
  return {
    left: battleRect.left - boardRect.left,
    top: battleRect.top - boardRect.top,
    width: battleRect.width,
    height: battleRect.height,
    centerX: battleRect.left + battleRect.width / 2 - boardRect.left,
    centerY: battleRect.top + battleRect.height / 2 - boardRect.top,
  }
}

function getElementCenter(selector: string) {
  const el = document.querySelector(selector)
  if (!el) return null
  const rect = el.getBoundingClientRect()
  
  const boardEl = document.querySelector('.board-shell')
  if (!boardEl) return null
  const boardRect = boardEl.getBoundingClientRect()
  
  return {
    x: rect.left + rect.width / 2 - boardRect.left,
    y: rect.top + rect.height / 2 - boardRect.top
  }
}

function getBattleCenter() {
  const centerBattle = getElementCenter('.center-battle')
  if (centerBattle) return centerBattle
  const battleZone = getElementCenter('.battle-zone-fill')
  if (battleZone) return battleZone

  const boardEl = document.querySelector('.board-shell')
  if (!boardEl) return { x: window.innerWidth / 2, y: window.innerHeight / 2 }
  const r = boardEl.getBoundingClientRect()
  return {
    x: r.width / 2,
    y: r.height / 2
  }
}

function actionLabel(actionType: string) {
  const normalized = String(actionType || '').trim().toLowerCase()
  if (normalized === 'attack') return '攻击'
  if (normalized === 'magic') return '法术'
  if (normalized === 'skill') return '技能'
  if (normalized === 'respond' || normalized === 'counter' || normalized === 'defend') return '响应'
  if (normalized === 'discard') return '弃置'
  return normalized || '出牌'
}

function resolveRevealSide(playerCenter: { x: number; y: number } | null, battle: NonNullable<ReturnType<typeof getBattleRectRelativeToBoard>>): RevealSide {
  if (!playerCenter) return 'bottom'
  const dx = playerCenter.x - battle.centerX
  const dy = playerCenter.y - battle.centerY
  return Math.abs(dx) > Math.abs(dy)
    ? (dx < 0 ? 'left' : 'right')
    : (dy < 0 ? 'top' : 'bottom')
}

function revealEntitiesForPlayer(playerId: string) {
  return displayCards.value
    .filter((item) => item.persistentBattleReveal && item.playerId === playerId)
    .sort((a, b) => a.id - b.id)
}

function persistentRevealPosition(playerId: string, playerCenter: { x: number; y: number } | null, revealIndex?: number) {
  const battle = getBattleRectRelativeToBoard()
  if (!battle) {
    const fallback = getBattleCenter()
    return { x: fallback.x - SETTLED_CARD_WIDTH / 2, y: fallback.y - SETTLED_CARD_HEIGHT / 2, side: 'bottom' as RevealSide }
  }

  const side = resolveRevealSide(playerCenter, battle)
  const siblingIndex = revealIndex ?? revealEntitiesForPlayer(playerId).length
  const spread = 24
  const gap = 12

  let x = battle.centerX - SETTLED_CARD_WIDTH / 2
  let y = battle.centerY - SETTLED_CARD_HEIGHT / 2

  if (side === 'left' || side === 'right') {
    y = (playerCenter?.y ?? battle.centerY) - SETTLED_CARD_HEIGHT / 2 + siblingIndex * spread
    y = clamp(y, battle.top + gap, battle.top + battle.height - SETTLED_CARD_HEIGHT - gap)
    x = side === 'left'
      ? battle.left + gap + siblingIndex * 5
      : battle.left + battle.width - SETTLED_CARD_WIDTH - gap - siblingIndex * 5
  } else {
    x = (playerCenter?.x ?? battle.centerX) - SETTLED_CARD_WIDTH / 2 + siblingIndex * spread
    x = clamp(x, battle.left + gap, battle.left + battle.width - SETTLED_CARD_WIDTH - gap)
    y = side === 'top'
      ? battle.top + gap + siblingIndex * 4
      : battle.top + battle.height - SETTLED_CARD_HEIGHT - gap - siblingIndex * 4
  }

  return { x, y, side }
}

function featuredRevealPosition() {
  const battle = getBattleRectRelativeToBoard()
  if (!battle) {
    const fallback = getBattleCenter()
    return {
      x: fallback.x - FEATURED_CARD_WIDTH / 2,
      y: fallback.y - FEATURED_CARD_HEIGHT / 2,
      side: undefined,
    }
  }
  return {
    x: battle.centerX - FEATURED_CARD_WIDTH / 2,
    y: battle.centerY - FEATURED_CARD_HEIGHT / 2,
    side: undefined,
  }
}

function refreshPersistentRevealLayout() {
  const persistent = displayCards.value
    .filter((item) => item.persistentBattleReveal)
    .sort((a, b) => a.id - b.id)
  const playerCounts = new Map<string, number>()
  for (const item of persistent) {
    if (item.revealState === 'featured') {
      const position = featuredRevealPosition()
      item.duration = 260
      item.x = position.x
      item.y = position.y
      item.revealSide = undefined
      item.transform = 'scale(1) rotate(0deg)'
      item.opacity = 1
      continue
    }

    const revealIndex = playerCounts.get(item.playerId) ?? 0
    playerCounts.set(item.playerId, revealIndex + 1)
    const playerCenter = getElementCenter(`[data-player-anchor="${item.playerId}"]`)
    const position = persistentRevealPosition(item.playerId, playerCenter, revealIndex)
    item.duration = 260
    item.x = position.x
    item.y = position.y
    item.revealSide = position.side
    item.transform = 'scale(1) rotate(0deg)'
    item.opacity = 1
  }
}

function settleFeaturedReveals(exceptId?: number) {
  for (const item of displayCards.value) {
    if (!item.persistentBattleReveal) continue
    if (item.id === exceptId) continue
    if (item.revealState !== 'featured') continue
    item.revealState = 'settled'
    cancelFeaturedSettleTimer(item.id)
  }
  nextTick(() => refreshPersistentRevealLayout())
}

function cancelFeaturedSettleTimer(id: number) {
  const timer = featuredSettleTimers.get(id)
  if (!timer) return
  clearTimeout(timer)
  featuredSettleTimers.delete(id)
}

function armFeaturedSettleTimer(id: number) {
  cancelFeaturedSettleTimer(id)
  const timer = setTimeout(() => {
    const item = displayCards.value.find((entry) => entry.id === id)
    if (item?.persistentBattleReveal && item.revealState === 'featured') {
      item.revealState = 'settled'
      nextTick(() => refreshPersistentRevealLayout())
    }
    featuredSettleTimers.delete(id)
  }, 1350)
  featuredSettleTimers.set(id, timer)
}

function openPreview(item: FlyingCardEntity) {
  if (!item.persistentBattleReveal) return
  previewCardId.value = item.id
}

function closePreview() {
  previewCardId.value = null
}

function shouldPersistBattleReveal(hidden?: boolean) {
  return hidden !== true
}

watch(flyingCards, (newVals) => {
  // 1. 处理新增的卡牌
  newVals.forEach(batch => {
    if (!displayCards.value.some(f => f.id === batch.id)) {
      nextTick(() => {
        const pCenter = getElementCenter(`[data-player-anchor="${batch.playerId}"]`)
        const battleCenter = getBattleCenter()
        const destX = battleCenter.x
        const destY = battleCenter.y

        const startX = pCenter ? pCenter.x - 40 : destX - 40
        const startY = pCenter ? pCenter.y - 60 : destY - 60

        // If there's already a card waiting in the center, shift the new one slightly
        const offsetIdx = displayCards.value.length
        const offsetX = offsetIdx * 20
        const offsetY = offsetIdx * 20
        const isPersistentBattleReveal = shouldPersistBattleReveal(batch.hidden)

        const entity: FlyingCardEntity = {
          id: batch.id,
          cards: batch.cards,
          hidden: batch.hidden,
          playerId: batch.playerId,
          playerName: batch.playerName,
          targetOffsetX: offsetX,
          targetOffsetY: offsetY,
          actionType: batch.actionType,
          x: startX,
          y: startY,
          transform: 'scale(0.3) rotate(-15deg)',
          opacity: 0,
          duration: 0,
          persistentBattleReveal: isPersistentBattleReveal,
          revealState: isPersistentBattleReveal ? 'featured' : undefined,
        }
        
        displayCards.value.push(entity)
        
        setTimeout(() => {
          void document.body.offsetHeight;
          const el = displayCards.value.find(f => f.id === batch.id)
          if (el) {
            const revealPosition = isPersistentBattleReveal
              ? featuredRevealPosition()
              : null
            el.duration = 800 // 速度放慢一倍 (原来是400)
            el.x = revealPosition ? revealPosition.x : destX - 40 + (el.targetOffsetX || 0)
            el.y = revealPosition ? revealPosition.y : destY - 60 + (el.targetOffsetY || 0)
            el.revealSide = revealPosition?.side
            el.transform = revealPosition ? 'scale(1) rotate(0deg)' : 'scale(1.2) rotate(0deg)'
            el.opacity = 1
            el.persistentBattleReveal = isPersistentBattleReveal
            if (isPersistentBattleReveal) {
              armFeaturedSettleTimer(el.id)
              battleFxStore.settleFlyingCardToBattlefield(el.id)
            }
          }
        }, 50)

        if (!batch.hidden) {
          settleFeaturedReveals(batch.id)
        }
      })
    }
  })

  // 2. 处理被移除的卡牌
  const currentIds = newVals.map(b => b.id)
  displayCards.value.forEach(fc => {
    if (!currentIds.includes(fc.id) && !fc.isRemoving) {
      if (fc.persistentBattleReveal) {
        if (fc.revealState === 'featured') {
          fc.revealState = 'settled'
          cancelFeaturedSettleTimer(fc.id)
          nextTick(() => refreshPersistentRevealLayout())
        }
        return
      }
      fc.isRemoving = true
      
      const cue = combatCue.value
      const isAttackOrMagic = fc.actionType === 'attack' || fc.actionType === 'counter' || fc.actionType === 'magic'
      
      let targetId = cue?.targetId
      // 如果没有对战提示，但有刚产生的伤害记录，也认为命中了
      if (!targetId && damageEffects.value.length > 0) {
        targetId = damageEffects.value[damageEffects.value.length - 1]?.targetId
      }
      
      // 如果是攻击且目标承受了伤害/产生了伤害特效
      const isHit = isAttackOrMagic && targetId && (cue?.phase === 'take' || damageEffects.value.length > 0)

      if (isHit && targetId) {
        const tCenter = getElementCenter(`[data-player-anchor="${targetId}"]`)
        if (tCenter) {
          fc.duration = 500 // 飞向脸部
          fc.x = tCenter.x - 40
          fc.y = tCenter.y - 60
          fc.transform = 'scale(0.5) rotate(20deg)'
          
          setTimeout(() => {
            fc.opacity = 0
            let dmgValue = 0
            if (damageEffects.value.length > 0) {
              const lastDmg = damageEffects.value[damageEffects.value.length - 1]
              if (lastDmg && lastDmg.targetId === targetId) {
                dmgValue = lastDmg.damage
              }
            }
            spawnExplosion(tCenter.x, tCenter.y, dmgValue)
            setTimeout(() => {
              displayCards.value = displayCards.value.filter(f => f.id !== fc.id)
            }, 400)
          }, 500)
          return
        }
      }

      // 其他情况（防御、应战、没命中的攻击、弃牌）：原地淡出，就像是在中间碰碎了
      fc.duration = 500
      fc.opacity = 0
      fc.transform = 'scale(1) rotate(0deg)'
      setTimeout(() => {
        displayCards.value = displayCards.value.filter(f => f.id !== fc.id)
      }, 500)
    }
  })

}, { deep: true })

watch(battlefieldRevealClearToken, () => {
  closePreview()
  for (const item of displayCards.value) {
    if (item.persistentBattleReveal) {
      cancelFeaturedSettleTimer(item.id)
    }
  }
  displayCards.value = displayCards.value.filter((item) => !item.persistentBattleReveal)
})

watch(
  () => displayCards.value.filter((item) => item.persistentBattleReveal).map((item) => `${item.id}:${item.playerId}:${item.revealState || ''}`).join('|'),
  () => nextTick(() => refreshPersistentRevealLayout())
)

watch(
  () => displayCards.value.some((item) => item.persistentBattleReveal && item.revealState === 'featured'),
  () => nextTick(() => refreshPersistentRevealLayout())
)

onBeforeUnmount(() => {
  for (const timer of featuredSettleTimers.values()) {
    clearTimeout(timer)
  }
  featuredSettleTimers.clear()
})
</script>

<style scoped>
.explosion-container {
  transform: translate(-50%, -50%);
  pointer-events: none;
  z-index: 10001;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}

.battlefield-reveal-card {
  filter: drop-shadow(0 10px 18px rgba(0, 0, 0, 0.42));
}

.battlefield-reveal-card--featured {
  transform-origin: center center;
}

.battlefield-reveal-card--featured .battlefield-reveal-stack {
  padding: 6px;
  border-radius: 14px;
  background: linear-gradient(180deg, rgba(14, 28, 44, 0.5), rgba(7, 14, 24, 0.28));
  box-shadow:
    inset 0 0 0 1px rgba(240, 219, 160, 0.3),
    0 10px 28px rgba(0, 0, 0, 0.34);
}

.battlefield-reveal-card--featured .battlefield-reveal-owner {
  margin-bottom: 5px;
}

.battlefield-reveal-card--settled .battlefield-reveal-stack {
  padding: 3px;
  border-radius: 10px;
  background: linear-gradient(180deg, rgba(9, 19, 32, 0.42), rgba(5, 10, 18, 0.28));
  box-shadow:
    inset 0 0 0 1px rgba(210, 185, 126, 0.22),
    0 0 0 1px rgba(5, 10, 18, 0.26);
}

.battlefield-reveal-stack {
  transition: padding 0.24s ease, border-radius 0.24s ease, box-shadow 0.24s ease, transform 0.24s ease;
}

.battlefield-reveal-owner {
  max-width: 92px;
  margin-bottom: 3px;
  padding: 2px 6px;
  border-radius: 999px;
  border: 1px solid rgba(224, 198, 128, 0.4);
  background: rgba(8, 16, 28, 0.72);
  color: #f8e7b8;
  font-size: 10px;
  font-weight: 800;
  line-height: 1.15;
  text-align: center;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.7);
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.battlefield-reveal-owner--bottom {
  order: 2;
  margin-top: 3px;
  margin-bottom: 0;
}

.battlefield-preview-overlay {
  position: fixed;
  inset: 0;
  z-index: 13000;
  background: rgba(4, 10, 18, 0.78);
  backdrop-filter: blur(3px);
  display: flex;
  align-items: center;
  justify-content: center;
  pointer-events: auto;
}

.battlefield-preview-panel {
  width: min(420px, calc(100vw - 32px));
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  padding: 16px 16px 14px;
  border-radius: 16px;
  background: linear-gradient(180deg, rgba(10, 18, 30, 0.96), rgba(7, 12, 22, 0.98));
  box-shadow:
    0 20px 54px rgba(0, 0, 0, 0.54),
    inset 0 0 0 1px rgba(226, 197, 128, 0.18);
}

.battlefield-preview-meta {
  width: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
}

.battlefield-preview-title {
  color: #f7e6ba;
  font-size: 14px;
  font-weight: 800;
  line-height: 1.2;
}

.battlefield-preview-subtitle {
  color: rgba(219, 230, 243, 0.72);
  font-size: 12px;
  line-height: 1.2;
}

.battlefield-preview-card-wrap {
  transform: scale(1.48);
  transform-origin: center center;
  margin: 10px 0 8px;
}

.battlefield-preview-card {
  pointer-events: none;
}

.battlefield-preview-close {
  border: 1px solid rgba(231, 205, 146, 0.28);
  border-radius: 999px;
  background: rgba(15, 26, 42, 0.88);
  color: #f2e2b6;
  font-size: 12px;
  font-weight: 700;
  padding: 6px 14px;
  cursor: pointer;
}

.explosion-effect {
  animation: explodeAnim 0.6s ease-out forwards;
  text-shadow: 0 0 20px rgba(255, 100, 0, 0.8);
}

.damage-number {
  position: absolute;
  font-size: 3rem;
  -webkit-text-stroke: 2px #4a0000;
  animation: damagePop 0.8s cubic-bezier(0.2, 0.8, 0.2, 1) forwards;
}

@keyframes damagePop {
  0% { transform: scale(0.5) translateY(20px); opacity: 0; }
  20% { transform: scale(1.2) translateY(-10px); opacity: 1; }
  100% { transform: scale(1) translateY(-40px); opacity: 0; }
}

@keyframes explodeAnim {
  0% { transform: scale(0.3); opacity: 1; }
  20% { transform: scale(1.4); opacity: 1; }
  100% { transform: scale(2.2); opacity: 0; }
}
</style>
