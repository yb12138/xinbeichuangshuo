<script setup lang="ts">
import { inject } from 'vue'
import { DECISION_DOCKED } from '../dockContext'
import ElementGlyph from '../../card/ElementGlyph.vue'
type FraudElementOption = {
  id: string
  title: string
  glyph: string
  tone: string
}

defineProps<{
  visible: boolean
  title: string
  options: FraudElementOption[]
}>()

const emit = defineEmits<{
  select: [optionId: string]
}>()

const docked = inject(DECISION_DOCKED, false)

/** tone（"prompt-fraud-card--fire"）→ 元素名，用于元素徽章卡配色与字形（按钮组件规范图 ⑥）。 */
function elementOf(option: FraudElementOption): string {
  const m = String(option.tone || '').match(/--([a-z]+)$/)
  return m && m[1] !== 'generic' ? m[1]! : ''
}

function glyphElement(element: string): string {
  return element ? element.charAt(0).toUpperCase() + element.slice(1) : ''
}
</script>

<template>
  <Teleport to="body" :disabled="docked">
    <Transition name="prompt-fraud-side-pop">
      <div
        v-if="visible && options.length > 0"
        class="prompt-fraud-global-layer" :class="{ 'prompt-fraud-global-layer--docked': docked }"
        data-testid="decision-overlay"
      >
        <div class="prompt-fraud-global-panel">
          <div class="prompt-fraud-dialog prompt-fraud-dialog--global">
            <div class="prompt-fraud-title">{{ title || '请选择本次攻击系别' }}</div>
            <div class="dc-els prompt-fraud-grid">
              <button
                v-for="option in options"
                :key="option.id"
                type="button"
                class="dc-el prompt-fraud-card"
                :class="[option.tone, elementOf(option) ? `dc-el--${elementOf(option)}` : '']"
                :aria-label="option.title"
                :data-testid="`prompt-option-${option.id}`"
                @click="emit('select', option.id)"
              >
                <span class="dc-el__glyph" aria-hidden="true">
                  <ElementGlyph v-if="elementOf(option)" :element="glyphElement(elementOf(option))" />
                  <template v-else>{{ option.glyph }}</template>
                </span>
                <span class="dc-el__label">{{ option.title }}</span>
              </button>
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.prompt-fraud-global-layer {
  position: fixed;
  inset: 0;
  z-index: 60;
  pointer-events: auto;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(5, 8, 15, 0.85);
  padding:
    max(16px, calc(var(--safe-top, 0px) + 8px))
    max(16px, calc(var(--safe-right, 0px) + 8px))
    max(16px, calc(var(--safe-bottom, 0px) + 8px))
    max(16px, calc(var(--safe-left, 0px) + 8px));
}

.prompt-fraud-global-panel {
  pointer-events: auto;
  width: min(860px, calc(100vw - 40px));
  max-height: calc(100vh - 40px);
  overflow: auto;
  border-radius: 14px;
  border: 1px solid rgba(146, 183, 207, 0.42);
  background:
    linear-gradient(180deg, rgba(9, 20, 34, 0.96), rgba(6, 14, 24, 0.97));
  box-shadow:
    0 18px 34px rgba(2, 8, 18, 0.52),
    inset 0 1px 0 rgba(236, 246, 254, 0.12);
  padding: 10px;
}

.prompt-fraud-dialog {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 4px 2px 2px;
}

.prompt-fraud-dialog--global {
  padding: 4px;
}

.prompt-fraud-title {
  font-size: 13px;
  line-height: 1.4;
  color: rgba(225, 238, 249, 0.96);
  text-align: center;
  letter-spacing: 0.01em;
}

.prompt-fraud-side-pop-enter-active,
.prompt-fraud-side-pop-leave-active {
  transition: opacity 0.2s ease, transform 0.2s ease;
}

.prompt-fraud-side-pop-enter-from,
.prompt-fraud-side-pop-leave-to {
  opacity: 0;
  transform: translateX(18px) scale(0.98);
}

@media (max-width: 900px) {
  .prompt-fraud-global-panel {
    width: min(760px, calc(100vw - 28px));
    max-height: calc(100vh - 28px);
  }

}

@media (max-width: 560px) {
  .prompt-fraud-global-layer {
    justify-content: center;
    align-items: center;
    padding:
      max(10px, calc(var(--safe-top, 0px) + 4px))
      max(8px, calc(var(--safe-right, 0px) + 4px))
      max(10px, calc(var(--safe-bottom, 0px) + 4px))
      max(8px, calc(var(--safe-left, 0px) + 4px));
  }

  .prompt-fraud-global-panel {
    width: min(100%, 620px);
    border-radius: 12px;
    padding: 8px;
  }

}

/* 停靠在决策面板内：原地渲染，去掉全屏遮罩与自身外框，由 DecisionDock 提供外壳 */
.prompt-fraud-global-layer.prompt-fraud-global-layer--docked {
  position: static;
  inset: auto;
  z-index: auto;
  display: block;
  padding: 0;
  background: none;
  pointer-events: auto;
}
.prompt-fraud-global-layer.prompt-fraud-global-layer--docked .prompt-fraud-title {
  display: none;
}
.prompt-fraud-global-layer.prompt-fraud-global-layer--docked .prompt-fraud-global-panel {
  width: 100%;
  max-width: none;
}
</style>
