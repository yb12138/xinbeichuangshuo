<script setup lang="ts">
import { DECISION_DOCKED } from '../dockContext'
import { computed, inject } from 'vue'

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
  totalLabel?: string
}>(), {
  submitLabel: '确认分配',
  totalLabel: '剩余可分配',
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

const docked = inject(DECISION_DOCKED, false)
</script>

<template>
  <Teleport to="body" :disabled="docked">
    <Transition name="modal">
      <div
        v-if="visible"
        class="overlay-panel-root overlay-panel-root--decision" :class="{ 'overlay-panel-root--docked': docked }"
        data-testid="allocation-overlay"
      >
        <div class="overlay-panel" @click.stop>
          <div class="overlay-panel-header overlay-panel-header--decision">
            <h2>{{ title }}</h2>
          </div>
          <!-- 停靠模板 7：总量条 + 每行 −/＋ 步进器 + 底部确认 -->
          <div v-if="docked" class="overlay-panel-body alloc-docked">
            <div class="dc-alloc-total alloc-docked__total" data-testid="allocation-summary">
              <span>已分配 {{ total - remaining }}/{{ total }}</span>
              <span class="dc-alloc-bar alloc-docked__bar" aria-hidden="true">
                <i :style="{ width: total > 0 ? `${((total - remaining) / total) * 100}%` : '0%' }" />
              </span>
            </div>
            <div class="alloc-docked__rows">
              <div v-for="(row, index) in rows" :key="row.id" class="alloc-docked__row">
                <span class="alloc-docked__label">{{ row.label }}</span>
                <span class="dc-stepper alloc-docked__stepper" :class="{ 'dc-stepper--active': currentValue(index) > 0 }">
                  <button
                    type="button"
                    class="dc-step alloc-docked__step"
                    :data-testid="`allocation-dec-${index}`"
                    :disabled="currentValue(index) <= 0"
                    :aria-label="`${row.label} 减少`"
                    @click="emit('change', index, currentValue(index) - 1)"
                  >−</button>
                  <span class="dc-step-value alloc-docked__value" :data-testid="`allocation-value-${index}`">{{ currentValue(index) }}</span>
                  <button
                    type="button"
                    class="dc-step alloc-docked__step"
                    :data-testid="`allocation-inc-${index}`"
                    :disabled="remaining <= 0"
                    :aria-label="`${row.label} 增加`"
                    @click="emit('change', index, currentValue(index) + 1)"
                  >+</button>
                </span>
              </div>
            </div>
            <div class="dc-actions dc-actions--end alloc-docked__actions">
              <button type="button" class="dc-pill dc-pill--gold alloc-docked__gold" data-testid="allocation-submit" :disabled="!canSubmit" @click="emit('submit')">
                确认
              </button>
            </div>
          </div>

          <div v-else class="overlay-panel-body overlay-saint-heal">
            <div class="overlay-saint-heal-summary" data-testid="allocation-summary">
              {{ totalLabel }}：{{ remaining }} / {{ total }}
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
                  type="button"
                  class="dc-num overlay-saint-heal-tile"
                  :class="{ 'dc-num--picked overlay-saint-heal-tile--active': currentValue(index) === value }"
                  :data-testid="`allocation-option-${index}-${value}`"
                  :disabled="isValueDisabled(index, value)"
                  @click="emit('change', index, value)"
                >
                  <span>{{ value }}</span>
                </button>
              </div>
            </div>
            <button
              type="button"
              class="dc-pill dc-pill--gold overlay-confirm-btn"
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
  background: rgba(5, 8, 15, 0.85);
}

.overlay-panel-root--decision {
  background:
    radial-gradient(380px 200px at 50% 46%, rgba(180, 140, 60, 0.12), transparent 70%),
    rgba(5, 8, 15, 0.85);
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

/* 停靠在决策面板内：原地渲染，去掉全屏遮罩与自身外框，由 DecisionDock 提供外壳 */
.overlay-panel-root.overlay-panel-root--docked {
  position: static;
  inset: auto;
  z-index: auto;
  display: block;
  padding: 0;
  background: none;
  pointer-events: auto;
}
.overlay-panel-root.overlay-panel-root--docked .overlay-panel {
  width: 100%;
  max-height: none;
  border: none;
  border-radius: 0;
  box-shadow: none;
  background: transparent;
}
.overlay-panel-root.overlay-panel-root--docked .overlay-panel-header {
  /* 决策面板的来由条已经写明技能与问题，这里不再重复 */
  display: none;
}
.overlay-panel-root.overlay-panel-root--docked .overlay-panel-body {
  padding: 0 4px;
  background: none;
}

/* —— 停靠模板 7：分配数值 —— */
.alloc-docked {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 2px 0 0;
}
.alloc-docked__rows {
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.alloc-docked__row {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 6px 10px;
  border-radius: 8px;
  background: rgba(9, 14, 26, 0.75);
  box-shadow: 0 0 0 1px var(--bt-hairline-soft, rgba(185, 193, 205, 0.22));
}
.alloc-docked__label {
  flex: 1;
  min-width: 0;
  font-size: 13px;
  color: var(--bt-text, #e8ecf4);
}
</style>
