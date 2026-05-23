<script setup lang="ts">
import { computed } from 'vue'

type ExtractOption = {
  id: string
  label: string
}

const props = defineProps<{
  visible: boolean
  options: ExtractOption[]
  selectedIndexes: number[]
  min: number
  max: number
  confirmImageSrc: string
  confirmImageReady: boolean
  confirmFallbackText?: string
}>()

const emit = defineEmits<{
  toggle: [index: number]
  confirm: []
  confirmImageError: []
}>()

const selectedCount = computed(() => props.selectedIndexes.length)
const confirmDisabled = computed(() => selectedCount.value < props.min || selectedCount.value > props.max)
const confirmAriaTitle = computed(() => `确认提炼（${selectedCount.value}/${props.max}）`)

function isSelected(index: number): boolean {
  return props.selectedIndexes.includes(index)
}

function formatOptionLabel(label: string): string {
  if (label === '红宝石') return '♦ 红宝石'
  if (label === '蓝水晶') return '🔷 蓝水晶'
  return label
}

function onConfirmClick() {
  if (confirmDisabled.value) return
  emit('confirm')
}
</script>

<template>
  <div v-if="visible" class="extract-prompt-root" data-testid="extract-prompt">
    <div class="extract-option-grid">
      <button
        v-for="(option, idx) in options"
        :key="option.id"
        class="extract-option-btn"
        :class="{ 'extract-option-btn--selected': isSelected(idx) }"
        :data-testid="`extract-option-${idx}`"
        @click="emit('toggle', idx)"
      >
        {{ formatOptionLabel(option.label) }}
      </button>
    </div>

    <div class="extract-confirm-row">
      <button
        class="extract-confirm-btn action-image-btn"
        :class="{ 'extract-confirm-btn--disabled': confirmDisabled }"
        :disabled="confirmDisabled"
        data-testid="prompt-confirm-btn"
        :title="confirmAriaTitle"
        :aria-label="confirmAriaTitle"
        @click="onConfirmClick"
      >
        <img
          v-if="confirmImageReady"
          class="action-image-btn-fill"
          :src="confirmImageSrc"
          alt=""
          @error="emit('confirmImageError')"
        />
        <span v-else class="action-image-fallback-text">{{ confirmFallbackText || '确' }}</span>
      </button>
    </div>
  </div>
</template>

<style scoped>
.extract-prompt-root {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.extract-option-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 8px;
}

.extract-option-btn {
  min-height: 40px;
  width: 100%;
  max-width: 100%;
  border-radius: 10px;
  border: 1px solid rgba(183, 154, 105, 0.56);
  background: linear-gradient(180deg, rgba(91, 69, 38, 0.9), rgba(68, 50, 28, 0.92));
  color: #e3eef8;
  font-size: 13px;
  font-weight: 700;
  letter-spacing: 0.01em;
  transition: transform 0.16s ease, border-color 0.16s ease, filter 0.16s ease;
}

.extract-option-btn:hover:not(:disabled) {
  transform: translateY(-1px);
  border-color: rgba(180, 210, 227, 0.66);
  filter: brightness(1.05);
}

.extract-option-btn--selected {
  box-shadow:
    0 0 0 2px rgba(241, 211, 150, 0.74),
    0 0 18px rgba(189, 152, 90, 0.38);
}

.extract-confirm-row {
  display: flex;
  justify-content: center;
  margin-top: 2px;
}

.extract-confirm-btn {
  -webkit-appearance: none;
  appearance: none;
  width: 72px;
  height: 72px;
  min-height: 0;
  max-width: 72px;
  aspect-ratio: 1 / 1;
  border: none;
  border-radius: 12px;
  background: transparent;
  box-shadow: none;
  padding: 0;
  overflow: hidden;
  position: relative;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  flex-shrink: 0;
}

.extract-confirm-btn:focus,
.extract-confirm-btn:focus-visible {
  outline: none;
  box-shadow: none;
}

.extract-confirm-btn--disabled {
  opacity: 0.45;
  cursor: not-allowed;
  transform: none;
  filter: none;
}

.action-image-btn {
  -webkit-appearance: none;
  appearance: none;
  border: none;
  background: transparent;
  box-shadow: none;
}

.action-image-btn-fill {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
  pointer-events: none;
  user-select: none;
}

.action-image-fallback-text {
  position: relative;
  z-index: 1;
  font-size: 14px;
  font-weight: 700;
  color: #f2f8ff;
  text-shadow: 0 1px 3px rgba(0, 0, 0, 0.45);
}

@media (max-width: 560px) {
  .extract-option-grid {
    grid-template-columns: 1fr;
  }
}
</style>
