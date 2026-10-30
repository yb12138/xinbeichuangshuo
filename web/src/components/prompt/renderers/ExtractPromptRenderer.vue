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
}>()

const emit = defineEmits<{
  toggle: [index: number]
  confirm: []
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
        type="button"
        class="dc-tile extract-option-btn"
        :class="[isSelected(idx) ? 'dc-tile--picked extract-option-btn--selected' : '', /宝石/.test(option.label) ? 'dc-tile--take' : 'dc-tile--gain']"
        :aria-pressed="isSelected(idx)"
        :data-testid="`extract-option-${idx}`"
        @click="emit('toggle', idx)"
      >
        {{ formatOptionLabel(option.label) }}
      </button>
    </div>

    <div class="dc-actions dc-actions--end extract-confirm-row">
      <button
        type="button"
        class="dc-pill dc-pill--gold extract-confirm-btn"
        :disabled="confirmDisabled"
        data-testid="prompt-confirm-btn"
        :aria-label="confirmAriaTitle"
        @click="onConfirmClick"
      >
        确认 {{ selectedCount }}/{{ max }}
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
  max-width: none;
  min-width: 0;
  min-height: 44px;
  padding: 10px;
  font-size: 14px;
  font-weight: 700;
}
</style>
