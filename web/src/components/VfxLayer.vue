<template>
  <div class="vfx-layer pointer-events-none overflow-visible" style="position: absolute !important; inset: 0 !important; width: 100% !important; height: 100% !important; z-index: 9999; border: 2px solid red;">
    <!-- SVG for Laser Beams -->
    <svg class="pointer-events-none" style="position: absolute !important; inset: 0 !important; width: 100% !important; height: 100% !important; z-index: 9999;">
      <defs>
        <filter id="glow-red" x="-50%" y="-50%" width="200%" height="200%">
          <feGaussianBlur stdDeviation="3" result="coloredBlur"/>
          <feMerge>
            <feMergeNode in="coloredBlur"/>
            <feMergeNode in="SourceGraphic"/>
          </feMerge>
        </filter>
        <filter id="glow-purple" x="-50%" y="-50%" width="200%" height="200%">
          <feGaussianBlur stdDeviation="3" result="coloredBlur"/>
          <feMerge>
            <feMergeNode in="coloredBlur"/>
            <feMergeNode in="SourceGraphic"/>
          </feMerge>
        </filter>
        <filter id="glow-yellow" x="-50%" y="-50%" width="200%" height="200%">
          <feGaussianBlur stdDeviation="3" result="coloredBlur"/>
          <feMerge>
            <feMergeNode in="coloredBlur"/>
            <feMergeNode in="SourceGraphic"/>
          </feMerge>
        </filter>

        <marker id="arrow-red" viewBox="0 0 10 10" refX="8" refY="5" markerWidth="6" markerHeight="6" orient="auto">
          <path d="M 0 0 L 10 5 L 0 10 z" fill="#ff4444" />
        </marker>
        <marker id="arrow-purple" viewBox="0 0 10 10" refX="8" refY="5" markerWidth="6" markerHeight="6" orient="auto">
          <path d="M 0 0 L 10 5 L 0 10 z" fill="#b044ff" />
        </marker>
      </defs>

      <line
        v-for="laser in lasers"
        :key="laser.id"
        :x1="laser.x1"
        :y1="laser.y1"
        :x2="laser.x2"
        :y2="laser.y2"
        :stroke="laser.color"
        :stroke-width="laser.width"
        :filter="laser.filter"
        :marker-end="laser.marker"
        pathLength="1"
        class="laser-beam"
      />
    </svg>

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
        transition: `all ${fc.duration}ms cubic-bezier(0.2, 0.8, 0.2, 1)`,
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

interface Laser {
  id: number
  x1: number
  y1: number
  x2: number
  y2: number
  color: string
  width: number
  filter: string
  marker: string
}

interface FlyingCardEntity {
  id: number
  cards: Card[]
  hidden?: boolean
  label?: string
  x: number
  y: number
  transform: string
  opacity: number
  duration: number
}

const lasers = ref<Laser[]>([])
const displayCards = ref<FlyingCardEntity[]>([])
let laserIdCounter = 0

// 计算元素中心点在 boardRoot 中的相对坐标
function getElementCenter(selector: string) {
  const el = document.querySelector(selector)
  if (!el) { console.warn('VFX: Could not find element', selector); return null }
  const rect = el.getBoundingClientRect()
  
  const boardEl = document.querySelector('.board-shell')
  if (!boardEl) return null
  const boardRect = boardEl.getBoundingClientRect()
  
  return {
    x: rect.left + rect.width / 2 - boardRect.left,
    y: rect.top + rect.height / 2 - boardRect.top
  }
}

// 监听对战提示，绘制连线
watch(() => store.combatCue, (cue, oldCue) => {
  console.log("VFX: combatCue changed", cue);
  if (!cue) return
  if (oldCue && oldCue.attackerId === cue.attackerId && oldCue.targetId === cue.targetId && oldCue.phase === cue.phase) return
  
  // 只在攻击/应战时画线
  if (cue.phase === 'defend' || cue.phase === 'take') return

  nextTick(() => {
    const p1 = getElementCenter(`[data-player-anchor="${cue.attackerId}"]`)
    const p2 = getElementCenter(`[data-player-anchor="${cue.targetId}"]`)
    
    if (p1 && p2) {
    let color = '#ff4444' // 攻击红线
    let filter = 'url(#glow-red)'
    let marker = 'url(#arrow-red)'
    
    if (cue.phase === 'counter') {
      color = '#eab308' // 应战黄线
      filter = 'url(#glow-yellow)'
      marker = ''
    }

    const id = ++laserIdCounter
    lasers.value.push({
      id,
      x1: p1.x,
      y1: p1.y,
      x2: p2.x,
      y2: p2.y,
      color,
      width: cue.phase === 'attack' ? 4 : 3,
      filter,
      marker
    })
    
    setTimeout(() => {
      lasers.value = lasers.value.filter(l => l.id !== id)
    }, 1200)
  }
  })
}, { deep: true })

// 监听飞牌事件
watch(() => store.flyingCards, (newVals) => {
  console.log("VFX: flyingCards changed", newVals);
  if (!newVals || newVals.length === 0) return
  const batch = newVals[0]
  if (!batch) return
  
  // 避免重复动画同一个 id
  if (displayCards.value.some(f => f.id === batch.id)) return

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

  // 动画起点：如果有玩家头像则从头像中心出发，否则直接在中央生成
  const startX = pCenter ? pCenter.x - 40 : destX - 40
  const startY = pCenter ? pCenter.y - 60 : destY - 60
  
  // 判断是否有目标
  let finalDestX = destX - 40
  let finalDestY = destY - 60
  
  // 如果当前刚好有 target，且是主动操作（不是承受伤害的明牌等），可以让牌最终飞向目标
  if (store.combatCue && (batch.actionType === 'attack' || batch.actionType === 'counter')) {
     const tCenter = getElementCenter(`[data-player-anchor="${store.combatCue.targetId}"]`)
     if (tCenter) {
       finalDestX = tCenter.x - 40
       finalDestY = tCenter.y - 60
     }
  }

  const fcId = batch.id
  const actionLabels: Record<string, string> = {
    attack: '攻击',
    magic: '法术',
    counter: '应战',
    defend: '抵挡',
    skill: '发动技能'
  }

  const entity: FlyingCardEntity = {
    id: fcId,
    cards: batch.cards,
    hidden: batch.hidden,
    label: actionLabels[batch.actionType] || batch.actionType,
    x: startX,
    y: startY,
    transform: 'scale(0.3) rotate(-15deg)',
    opacity: 0,
    duration: 0
  }
  
  displayCards.value.push(entity)
  
  // 使用 setTimeout 等待 Vue 渲染了刚刚 push 的初始状态元素
  setTimeout(() => {
    void document.body.offsetHeight; // 强制 reflow
    const el = displayCards.value.find(f => f.id === fcId)
    if (el) {
      el.duration = 400
      el.x = destX - 40
      el.y = destY - 60
      el.transform = 'scale(1.2) rotate(0deg)'
      el.opacity = 1
    }
  }, 50)
  
  // 如果是攻击，在中间停留一会后再冲向目标
  if (store.combatCue && (batch.actionType === 'attack' || batch.actionType === 'counter')) {
    setTimeout(() => {
      const el = displayCards.value.find(f => f.id === fcId)
      if (el) {
        el.duration = 300
        el.x = finalDestX
        el.y = finalDestY
        el.transform = 'scale(0.5) rotate(15deg)'
        el.opacity = 0
      }
    }, 900)
  }

  // 清理
  setTimeout(() => {
    displayCards.value = displayCards.value.filter(f => f.id !== fcId)
  }, 1400)
  })
}, { deep: true })
</script>

<style scoped>
.laser-beam {
  stroke-dasharray: 1;
  stroke-dashoffset: 1;
  animation: dash 0.8s cubic-bezier(0.2, 0.8, 0.2, 1) forwards;
}

@keyframes dash {
  0% {
    stroke-dashoffset: 1;
    opacity: 1;
  }
  40% {
    stroke-dashoffset: 0;
    opacity: 1;
  }
  100% {
    stroke-dashoffset: 0;
    opacity: 0;
  }
}
</style>
