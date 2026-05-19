<script setup lang="ts">
import { computed } from 'vue'

type AllocationRow = {
  id: string
  label: string
}

const props = withDefaults(defineProps<{
  visible: boolean
  title: string
  rows: AllocationRow[]
  values: number[]
  remaining: number
  total: number
  canSubmit: boolean
  submitLabel?: string
}>(), {
  submitLabel: '确认分配',
})

const emit = defineEmits<{
  change: [index: number, value: number]
  submit: []
}>()

const selectableValues = computed(() => {
  const total = Math.max(0, Math.floor(props.total || 0))
  return Array.from({ length: total + 1 }, (_, index) => index)
})

function currentValue(index: number): number {
  return props.values[index] || 0
}

function isValueDisabled(index: number, value: number): boolean {
  return value > currentValue(index) + props.remaining
}
</script>

<template>
  <Teleport to="body">
    <Transition name="modal">
      <div
        v-if="visible"
        class="overlay-panel-root overlay-panel-root--decision"
        data-testid="allocation-overlay"
      >
        <div class="overlay-panel" @click.stop>
          <div class="overlay-panel-header overlay-panel-header--decision">
            <h2>{{ title }}</h2>
          </div>
          <div class="overlay-panel-body overlay-saint-heal">
            <div class="overlay-saint-heal-summary" data-testid="allocation-summary">
              剩余可分配：{{ remaining }} / {{ total }}
            </div>
            <div
              v-for="(row, index) in rows"
              :key="row.id"
              class="overlay-saint-heal-row"
            >
              <div class="overlay-saint-heal-row-label">{{ row.label }}</div>
              <div class="overlay-saint-heal-row-grid">
                <button
                  v-for="value in selectableValues"
                  :key="value"
                  class="overlay-numeric-tile overlay-saint-heal-tile"
                  :class="{ 'overlay-saint-heal-tile--active': currentValue(index) === value }"
                  :data-testid="`allocation-option-${index}-${value}`"
                  :disabled="isValueDisabled(index, value)"
                  @click="emit('change', index, value)"
                >
                  <span class="overlay-numeric-value">{{ value }}</span>
                </button>
              </div>
            </div>
            <button
              class="overlay-confirm-btn"
              data-testid="allocation-submit"
              :disabled="!canSubmit"
              @click="emit('submit')"
            >
              {{ submitLabel }}
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.overlay-panel-root {
  position: fixed;
  inset: 0;
  z-index: 13050;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1rem;
  backdrop-filter: blur(3px);
}

.overlay-panel-root--decision {
  background:
    radial-gradient(380px 200px at 50% 46%, rgba(180, 140, 60, 0.12), transparent 70%),
    rgba(0, 0, 0, 0.74);
}

.overlay-panel {
  position: relative;
  width: min(480px, calc(100vw - 2rem));
  max-height: 85vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  border-radius: 14px;
  border: 1px solid rgba(180, 150, 90, 0.32);
  box-shadow:
    0 18px 34px rgba(2, 8, 18, 0.52),
    inset 0 1px 0 rgba(255, 240, 200, 0.1);
  background: linear-gradient(180deg, rgba(10, 18, 30, 0.94), rgba(6, 12, 22, 0.96));
}

.overlay-panel-header {
  flex-shrink: 0;
  padding: 20px 24px 16px;
  border-bottom: 1px solid rgba(149, 186, 204, 0.26);
  text-align: center;
}

.overlay-panel-header h2 {
  margin: 0;
  font-size: 1.05rem;
  font-weight: 600;
  line-height: 1.5;
  color: #ffe2ad;
}

.overlay-panel-header--decision {
  padding: 18px 24px 14px;
  background: linear-gradient(110deg, rgba(40, 56, 80, 0.9), rgba(80, 60, 30, 0.85));
  border-bottom-color: rgba(180, 150, 90, 0.24);
}

.overlay-panel-header--decision h2 {
  font-size: 1rem;
  color: #ffd98a;
}

.overlay-panel-body {
  flex: 1;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 10px;
  background: rgba(6, 14, 24, 0.4);
}

.overlay-numeric-tile {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  width: 56px;
  min-height: 56px;
  border-radius: 12px;
  border: 1px solid rgba(180, 150, 90, 0.36);
  background: rgba(20, 32, 48, 0.7);
  cursor: pointer;
  transition: all 0.18s ease;
  padding: 8px 4px;
  font-family: inherit;
}

.overlay-numeric-tile:hover:not(:disabled) {
  border-color: rgba(255, 210, 120, 0.7);
  background: rgba(30, 44, 60, 0.85);
  box-shadow: 0 0 14px rgba(255, 210, 120, 0.18);
}

.overlay-numeric-tile:active:not(:disabled) {
  transform: scale(0.95);
}

.overlay-numeric-tile:disabled {
  opacity: 0.35;
  cursor: not-allowed;
}

.overlay-numeric-value {
  font-size: 1.4rem;
  font-weight: 700;
  color: #ffd98a;
  line-height: 1;
}

.overlay-confirm-btn {
  padding: 10px 36px;
  border-radius: 10px;
  border: 1px solid rgba(180, 150, 90, 0.5);
  background: linear-gradient(180deg, rgba(120, 90, 30, 0.7), rgba(80, 60, 20, 0.8));
  color: #ffe2ad;
  font-size: 0.9rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.15s ease;
  font-family: inherit;
}

.overlay-confirm-btn:hover:not(:disabled) {
  background: linear-gradient(180deg, rgba(140, 105, 35, 0.85), rgba(100, 75, 25, 0.9));
  border-color: rgba(255, 210, 120, 0.7);
}

.overlay-confirm-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.overlay-saint-heal {
  display: flex;
  flex-direction: column;
  gap: 14px;
  padding: 18px 24px 20px;
}

.overlay-saint-heal-summary {
  text-align: center;
  font-size: 0.9rem;
  color: rgba(255, 217, 138, 0.92);
}

.overlay-saint-heal-row {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 10px 12px;
  border-radius: 10px;
  background: rgba(10, 22, 38, 0.55);
  border: 1px solid rgba(118, 153, 173, 0.22);
}

.overlay-saint-heal-row-label {
  font-size: 0.95rem;
  font-weight: 600;
  color: #ffd98a;
}

.overlay-saint-heal-row-grid {
  display: flex;
  gap: 8px;
}

.overlay-saint-heal-tile {
  flex: 1;
}

.overlay-saint-heal-tile--active {
  border-color: rgba(255, 210, 120, 0.85);
  background: linear-gradient(180deg, rgba(140, 100, 35, 0.7), rgba(90, 65, 20, 0.85));
  color: #ffe9b8;
}
</style>
