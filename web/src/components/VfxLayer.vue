<template>
  <div class="vfx-layer pointer-events-none overflow-visible" style="position: absolute !important; inset: 0 !important; width: 100% !important; height: 100% !important; z-index: 9999;">

    <!-- Explosions and Damage Numbers -->
    <div v-for="exp in explosions" :key="'exp'+exp.id" class="absolute explosion-container" :style="{ left: exp.x + 'px', top: exp.y + 'px' }">
      <div class="explosion-effect text-6xl">💥</div>
      <div v-if="exp.damage" class="damage-number font-black text-red-500 drop-shadow-[0_0_8px_rgba(255,0,0,0.8)]">
        -{{ exp.damage }}
      </div>
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

watch(() => store.flyingCards, (newVals) => {
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

        const entity: FlyingCardEntity = {
          id: batch.id,
          cards: batch.cards,
          hidden: batch.hidden,
          targetOffsetX: offsetX,
          targetOffsetY: offsetY,
          actionType: batch.actionType,
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
            let dmgValue = 0
            if (store.damageEffects.length > 0) {
              const lastDmg = store.damageEffects[store.damageEffects.length - 1]
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
