<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import type { CharacterView } from '../types/game'

interface SkillModalAnchor {
  x: number
  y: number
  width: number
  height: number
}

const props = defineProps<{
  character: CharacterView | null
  visible: boolean
  anchor?: SkillModalAnchor | null
}>()

const emit = defineEmits<{
  close: []
}>()

const charDisplayName = computed(() => {
  if (!props.character) return ''
  return `${props.character.name} - ${props.character.title}`
})

const viewport = ref({
  width: typeof window !== 'undefined' ? window.innerWidth : 1280,
  height: typeof window !== 'undefined' ? window.innerHeight : 720,
})

function syncViewport() {
  if (typeof window === 'undefined') return
  viewport.value = {
    width: window.innerWidth,
    height: window.innerHeight,
  }
}

onMounted(() => {
  if (typeof window === 'undefined') return
  window.addEventListener('resize', syncViewport)
})

onBeforeUnmount(() => {
  if (typeof window === 'undefined') return
  window.removeEventListener('resize', syncViewport)
})

const isAnchored = computed(() => !!props.anchor)

const panelStyle = computed(() => {
  if (!props.anchor) return {}
  const margin = 12
  const gap = 8
  const preferredWidth = 440
  const minWidth = 300
  const estimatedHeight = 560
  const width = Math.min(preferredWidth, Math.max(minWidth, viewport.value.width - margin * 2))

  let left = props.anchor.x - width - gap
  if (left < margin) {
    left = props.anchor.x + props.anchor.width + gap
  }
  if (left + width > viewport.value.width - margin) {
    left = viewport.value.width - width - margin
  }
  if (left < margin) left = margin

  let top = props.anchor.y - 6
  const maxTop = viewport.value.height - estimatedHeight - margin
  if (top > maxTop) {
    top = Math.max(margin, maxTop)
  }
  if (top < margin) top = margin

  return {
    width: `${Math.round(width)}px`,
    left: `${Math.round(left)}px`,
    top: `${Math.round(top)}px`,
  }
})
</script>

<template>
  <Transition name="modal">
    <div
      v-if="visible && character"
      class="skill-overlay fixed inset-0 z-[100]"
      :class="isAnchored ? 'skill-overlay--anchored' : 'skill-overlay--centered'"
      @click.self="!isAnchored && emit('close')"
    >
      <div
        class="skill-card rounded-2xl shadow-2xl max-h-[85vh] overflow-hidden flex flex-col"
        :class="isAnchored ? 'skill-card--anchored' : 'skill-card--centered'"
        :style="isAnchored ? panelStyle : undefined"
        @click.stop
      >
        <!-- 标题栏 -->
        <div class="skill-header flex-shrink-0 px-6 py-4 flex justify-between items-center">
          <h2 class="text-xl font-bold text-white">{{ charDisplayName }}</h2>
          <button
            type="button"
            class="p-2 rounded-lg hover:bg-white/20 text-white transition-colors"
            aria-label="关闭"
            @click="emit('close')"
          >
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <!-- 技能列表 - 可滚动，不截断 -->
        <div class="skill-body flex-1 overflow-y-auto p-6 space-y-4">
          <div
            v-for="skill in character.skills"
            :key="skill.id"
            class="skill-item rounded-xl p-4"
          >
            <div class="font-semibold text-amber-400 text-base mb-2">{{ skill.title }}</div>
            <div class="text-gray-300 text-sm leading-relaxed whitespace-pre-wrap break-words">
              {{ skill.description }}
            </div>
            <div v-if="skill.cost_gem || skill.cost_crystal || skill.cost_discards" class="mt-2 text-xs text-gray-400">
              <span v-if="skill.cost_gem">消耗♦{{ skill.cost_gem }} </span>
              <span v-if="skill.cost_crystal">消耗🔷{{ skill.cost_crystal }} </span>
              <span v-if="skill.cost_discards">弃{{ skill.cost_discards }}张牌</span>
              <span v-if="skill.discard_element">({{ skill.discard_element }})</span>
            </div>
          </div>
        </div>

        <!-- 底部 -->
        <div class="skill-footer flex-shrink-0 px-6 py-4">
          <button
            type="button"
            class="w-full py-2.5 rounded-lg font-medium btn-skill text-white"
            @click="emit('close')"
          >
            关闭
          </button>
        </div>
      </div>
    </div>
  </Transition>
</template>

<style scoped>
.skill-overlay--centered {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1rem;
  background:
    radial-gradient(420px 220px at 50% 44%, rgba(136, 188, 195, 0.18), transparent 70%),
    rgba(0, 0, 0, 0.72);
  backdrop-filter: blur(3px);
}

.skill-overlay--anchored {
  pointer-events: none;
  padding: 0;
}

.skill-card {
  background:
    linear-gradient(180deg, rgba(8, 20, 34, 0.92), rgba(6, 15, 28, 0.95)),
    url('/assets/ui/modal-aura.svg') center/cover no-repeat;
  border: 1px solid rgba(132, 167, 186, 0.36);
}

.skill-card--centered {
  width: min(680px, calc(100vw - 2rem));
}

.skill-card--anchored {
  position: fixed;
  pointer-events: auto;
  max-height: min(78vh, 640px);
}

.skill-header {
  background: linear-gradient(110deg, rgba(34, 74, 97, 0.88), rgba(94, 72, 43, 0.88));
  border-bottom: 1px solid rgba(149, 186, 204, 0.26);
}

.skill-body {
  background: rgba(6, 17, 29, 0.42);
}

.skill-item {
  border: 1px solid rgba(118, 152, 173, 0.34);
  background: rgba(14, 32, 48, 0.56);
  box-shadow: inset 0 1px 0 rgba(237, 247, 254, 0.06);
}

.skill-footer {
  background: rgba(6, 16, 28, 0.66);
  border-top: 1px solid rgba(118, 153, 173, 0.24);
}

.modal-enter-active,
.modal-leave-active {
  transition: opacity 0.24s ease;
}
.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}
.modal-enter-active .skill-card,
.modal-leave-active .skill-card {
  transition: transform 0.24s ease;
}
.modal-enter-from .skill-card,
.modal-leave-to .skill-card {
  transform: scale(0.95) translateY(8px);
}
</style>
