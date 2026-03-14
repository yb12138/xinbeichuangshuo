<template>
  <div class="vfx-layer pointer-events-none overflow-visible" style="position: absolute !important; inset: 0 !important; width: 100% !important; height: 100% !important; z-index: 9999;">

    <!-- Explosions -->
    <div v-for="exp in explosions" :key="'exp'+exp.id" class="absolute explosion-effect text-6xl" :style="{ left: exp.x + 'px', top: exp.y + 'px' }">
      💥
    </div>

    <!-- Flying Cards -->
    <div
      v-for="fc in displayCards"
      :key="fc.id"
      class="absolute flex flex-col items-center pointer-events-none"
      :style="{
        left: fc.x + 'px',
        top: fc.y + 'px',
        transform: fc.transform,
        opacity: fc.opacity,
        transition: `transform ${fc.duration}ms cubic-bezier(0.2, 0.8, 0.2, 1), left ${fc.duration}ms cubic-bezier(0.2, 0.8, 0.2, 1), top ${fc.duration}ms cubic-bezier(0.2, 0.8, 0.2, 1), opacity ${fc.duration}ms linear`,
        zIndex: 10000
      }"
    >
      <div class="relative flex">
        <div v-for="(c, cidx) in fc.cards" :key="cidx" class="relative" :style="{ marginLeft: cidx > 0 ? '-30px' : '0' }">
          <CardComponent :card="c" :face-down="fc.hidden" small class="shadow-2xl" />
        </div>
      </div>
      <div v-if="fc.label" class="mt-2 px-2 py-0.5 bg-black/60 text-white text-xs rounded-full shadow-lg border border-white/20">
        {{ fc.label }}
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, nextTick } from 'vue'
import { useGameStore } from '../stores/gameStore'
import CardComponent from './CardComponent.vue'
import type { Card } from '../types/game'

const store = useGameStore()

interface Explosion {
  id: number
  x: number
  y: number
}
const explosions = ref<Explosion[]>([])
let expIdCounter = 0

function spawnExplosion(x: number, y: number) {
  const id = ++expIdCounter
  explosions.value.push({ id, x, y })
  setTimeout(() => {
    explosions.value = explosions.value.filter(e => e.id !== id)
  }, 600)
}

interface FlyingCardEntity {
  id: number
  cards: Card[]
  hidden?: boolean
  actionType: string
  label?: string
  x: number
  y: number
  transform: string
  opacity: number
  duration: number
  isRemoving?: boolean
  targetOffsetX?: number
  targetOffsetY?: number
}

const displayCards = ref<FlyingCardEntity[]>([])

const actionLabels: Record<string, string> = {
  attack: '攻击',
  magic: '法术',
  counter: '应战',
  defend: '抵挡',
  skill: '发动技能',
  discard: '弃牌'
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

watch(() => store.flyingCards, (newVals) => {
  // 1. 处理新增的卡牌
  newVals.forEach(batch => {
    if (!displayCards.value.some(f => f.id === batch.id)) {
      nextTick(() => {
        const pCenter = getElementCenter(`[data-player-anchor="${batch.playerId}"]`)
        const boardEl = document.querySelector('.board-shell')
        let destX = window.innerWidth / 2
        let destY = window.innerHeight / 2
        if (boardEl) {
          const r = boardEl.getBoundingClientRect()
          destX = r.width / 2
          destY = r.height / 2
        }

        const startX = pCenter ? pCenter.x - 40 : destX - 40
        const startY = pCenter ? pCenter.y - 60 : destY - 60

        // If there's already a card waiting in the center, shift the new one slightly
        const offsetIdx = displayCards.value.length
        const offsetX = offsetIdx * 20
        const offsetY = offsetIdx * 20

        const entity: FlyingCardEntity = {
          id: batch.id,
          cards: batch.cards,
          hidden: batch.hidden,
          targetOffsetX: offsetX,
          targetOffsetY: offsetY,
          actionType: batch.actionType,
          label: actionLabels[batch.actionType] || batch.actionType,
          x: startX,
          y: startY,
          transform: 'scale(0.3) rotate(-15deg)',
          opacity: 0,
          duration: 0
        }
        
        displayCards.value.push(entity)
        
        setTimeout(() => {
          void document.body.offsetHeight;
          const el = displayCards.value.find(f => f.id === batch.id)
          if (el) {
            el.duration = 800 // 速度放慢一倍 (原来是400)
            el.x = destX - 40 + (el.targetOffsetX || 0)
            el.y = destY - 60 + (el.targetOffsetY || 0)
            el.transform = 'scale(1.2) rotate(0deg)'
            el.opacity = 1
          }
        }, 50)
      })
    }
  })

  // 2. 处理被移除的卡牌
  const currentIds = newVals.map(b => b.id)
  displayCards.value.forEach(fc => {
    if (!currentIds.includes(fc.id) && !fc.isRemoving) {
      fc.isRemoving = true
      
      const cue = store.combatCue
      const isAttackOrMagic = fc.actionType === 'attack' || fc.actionType === 'counter' || fc.actionType === 'magic'
      
      let targetId = cue?.targetId
      // 如果没有对战提示，但有刚产生的伤害记录，也认为命中了
      if (!targetId && store.damageEffects.length > 0) {
        targetId = store.damageEffects[store.damageEffects.length - 1]?.targetId
      }
      
      // 如果是攻击且目标承受了伤害/产生了伤害特效
      const isHit = isAttackOrMagic && targetId && (cue?.phase === 'take' || store.damageEffects.length > 0)

      if (isHit && targetId) {
        const tCenter = getElementCenter(`[data-player-anchor="${targetId}"]`)
        if (tCenter) {
          fc.duration = 500 // 飞向脸部
          fc.x = tCenter.x - 40
          fc.y = tCenter.y - 60
          fc.transform = 'scale(0.5) rotate(20deg)'
          
          setTimeout(() => {
            fc.opacity = 0
            spawnExplosion(tCenter.x, tCenter.y)
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
</script>

<style scoped>
.explosion-effect {
  transform: translate(-50%, -50%);
  animation: explodeAnim 0.6s ease-out forwards;
  pointer-events: none;
  z-index: 10001;
  text-shadow: 0 0 20px rgba(255, 100, 0, 0.8);
}

@keyframes explodeAnim {
  0% { transform: translate(-50%, -50%) scale(0.3); opacity: 1; }
  20% { transform: translate(-50%, -50%) scale(1.4); opacity: 1; }
  100% { transform: translate(-50%, -50%) scale(2.2); opacity: 0; }
}
</style>