<script setup lang="ts">
type DecisionOverlayMode = 'numeric' | 'text' | 'activation-cost' | 'yes-no'

type DecisionOverlayOption = {
  id: string
  label: string
  buttonLabel: string
  hint?: string
  disabled?: boolean
}

const props = defineProps<{
  visible: boolean
  title: string
  mode: DecisionOverlayMode
  options: DecisionOverlayOption[]
  activationHint?: string
  activationOptionId?: string
  activationDisabled?: boolean
  canCancel: boolean
  cancelLabel: string
  cancelOptionId: string
}>()

const emit = defineEmits<{
  select: [optionId: string]
  cancel: [optionId: string]
}>()

function overlayDecisionOptionTitle(option: DecisionOverlayOption): string {
  const label = String(option.label || '').trim()
  const buttonLabel = String(option.buttonLabel || '').trim()
  if (!label) return buttonLabel
  if (!buttonLabel || buttonLabel === label) return label
  if (/^\d+$/.test(buttonLabel)) return label
  return buttonLabel
}
</script>

<template>
  <Teleport to="body">
    <Transition name="modal">
      <div
        v-if="visible"
        class="overlay-panel-root overlay-panel-root--decision"
        data-testid="decision-overlay"
      >
        <div class="overlay-panel" @click.stop>
          <div class="overlay-panel-header overlay-panel-header--decision">
            <h2>{{ title }}</h2>
          </div>

          <div v-if="mode === 'activation-cost'" class="overlay-panel-body overlay-panel-body--cost">
            <div class="overlay-cost-card">
              <span class="overlay-cost-text">{{ activationHint }}</span>
            </div>
            <button
              class="overlay-confirm-btn"
              :disabled="!!activationDisabled"
              @click="emit('select', activationOptionId || '')"
            >
              确认发动
            </button>
          </div>

          <div v-else-if="mode === 'numeric'" class="overlay-panel-body overlay-panel-body--numeric">
            <div class="overlay-numeric-grid">
              <button
                v-for="option in options"
                :key="option.id"
                class="overlay-numeric-tile"
                :data-testid="`numeric-option-${option.buttonLabel}`"
                :disabled="!!option.disabled"
                @click="emit('select', option.id)"
              >
                <span class="overlay-numeric-value">{{ option.buttonLabel }}</span>
              </button>
            </div>
          </div>

          <div v-else-if="mode === 'yes-no'" class="overlay-panel-body overlay-panel-body--yesno">
            <div class="overlay-yesno-row">
              <button
                v-for="option in options"
                :key="option.id"
                class="overlay-yesno-btn"
                :class="option.id === '0' || option.id === 'yes' ? 'overlay-yesno-btn--yes' : 'overlay-yesno-btn--no'"
                :data-testid="`prompt-option-${option.id}`"
                :disabled="!!option.disabled"
                @click="emit('select', option.id)"
              >
                {{ String(option.label || '').trim() }}
              </button>
            </div>
          </div>

          <div v-else class="overlay-panel-body overlay-panel-body--text">
            <button
              v-for="(option, idx) in options"
              :key="option.id"
              class="overlay-panel-item overlay-panel-item--text"
              :data-testid="`branch-option-${idx}`"
              :disabled="!!option.disabled"
              @click="emit('select', option.id)"
            >
              <div class="overlay-panel-item-title" :data-testid="`prompt-option-${option.id}`">{{ overlayDecisionOptionTitle(option) }}</div>
              <div
                v-if="option.hint && option.hint !== overlayDecisionOptionTitle(option)"
                class="overlay-panel-item-desc"
              >{{ option.hint }}</div>
            </button>
          </div>

          <div v-if="canCancel && mode !== 'yes-no'" class="overlay-panel-footer">
            <button class="overlay-panel-cancel" data-testid="prompt-cancel-btn" @click="emit('cancel', cancelOptionId)">
              {{ cancelLabel || '取消' }}
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
  background:
    radial-gradient(380px 200px at 50% 46%, rgba(180, 140, 60, 0.12), transparent 70%),
    rgba(0, 0, 0, 0.74);
  backdrop-filter: blur(3px);
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
  padding: 18px 24px 14px;
  background: linear-gradient(110deg, rgba(40, 56, 80, 0.9), rgba(80, 60, 30, 0.85));
  border-bottom: 1px solid rgba(180, 150, 90, 0.24);
  text-align: center;
}

.overlay-panel-header h2 {
  margin: 0;
  color: #ffd98a;
  font-size: 1rem;
  font-weight: 600;
  line-height: 1.5;
}

.overlay-panel-body {
  flex: 1;
  overflow-y: auto;
  padding: 16px 20px;
  display: flex;
  flex-direction: column;
  gap: 10px;
  background: rgba(6, 14, 24, 0.4);
}

.overlay-panel-item {
  position: relative;
  width: 100%;
  text-align: left;
  padding: 14px 18px;
  border-radius: 10px;
  border: 1px solid rgba(150, 130, 80, 0.3);
  background: rgba(14, 28, 44, 0.56);
  box-shadow: inset 0 1px 0 rgba(255, 240, 200, 0.05);
  cursor: pointer;
  transition: all 0.18s ease;
  font-family: inherit;
  font-size: inherit;
  line-height: inherit;
}

.overlay-panel-item:hover:not(:disabled) {
  border-color: rgba(255, 210, 120, 0.55);
  background: rgba(22, 38, 56, 0.72);
  box-shadow:
    0 0 12px rgba(255, 210, 120, 0.12),
    inset 0 1px 0 rgba(255, 240, 200, 0.08);
}

.overlay-panel-item:active:not(:disabled) {
  transform: scale(0.98);
}

.overlay-panel-item:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.overlay-panel-item-title {
  margin-bottom: 4px;
  color: #ffd98a;
  font-size: 0.95rem;
  font-weight: 600;
}

.overlay-panel-item-desc {
  color: rgba(199, 219, 237, 0.78);
  font-size: 0.8rem;
  line-height: 1.5;
  white-space: pre-wrap;
  word-break: break-word;
}

.overlay-panel-footer {
  flex-shrink: 0;
  padding: 12px 20px 16px;
  text-align: center;
  background: rgba(6, 14, 24, 0.66);
  border-top: 1px solid rgba(150, 130, 80, 0.2);
}

.overlay-panel-cancel {
  padding: 8px 32px;
  border-radius: 8px;
  border: 1px solid rgba(156, 166, 184, 0.3);
  background: rgba(59, 67, 84, 0.6);
  color: rgba(199, 219, 237, 0.7);
  font-size: 0.85rem;
  cursor: pointer;
  transition: all 0.15s ease;
  font-family: inherit;
}

.overlay-panel-cancel:hover {
  background: rgba(59, 67, 84, 0.9);
  color: rgba(199, 219, 237, 0.95);
  border-color: rgba(156, 166, 184, 0.5);
}

.overlay-panel-body--numeric {
  padding: 20px 16px;
}

.overlay-numeric-grid {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 10px;
}

.overlay-numeric-tile {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  width: 56px;
  min-height: 56px;
  padding: 8px 4px;
  border-radius: 12px;
  border: 1px solid rgba(180, 150, 90, 0.36);
  background: rgba(20, 32, 48, 0.7);
  cursor: pointer;
  transition: all 0.18s ease;
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
  color: #ffd98a;
  font-size: 1.4rem;
  font-weight: 700;
  line-height: 1;
}

.overlay-panel-body--yesno {
  align-items: center;
  padding: 24px 20px;
}

.overlay-yesno-row {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
  width: 100%;
  max-width: 320px;
}

.overlay-yesno-btn {
  padding: 14px 24px;
  border-radius: 10px;
  border: 1px solid rgba(150, 130, 80, 0.36);
  background: rgba(14, 28, 44, 0.56);
  color: #ffd98a;
  font-size: 1.1rem;
  font-weight: 700;
  cursor: pointer;
  transition: all 0.18s ease;
  font-family: inherit;
  text-align: center;
}

.overlay-yesno-btn:hover:not(:disabled) {
  border-color: rgba(255, 210, 120, 0.7);
  background: rgba(22, 38, 56, 0.72);
  box-shadow: 0 0 12px rgba(255, 210, 120, 0.12);
}

.overlay-yesno-btn:active:not(:disabled) {
  transform: scale(0.96);
}

.overlay-yesno-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.overlay-yesno-btn--yes {
  border-color: rgba(180, 150, 90, 0.5);
  background: linear-gradient(180deg, rgba(100, 75, 30, 0.7), rgba(70, 52, 20, 0.8));
  color: #ffe2ad;
}

.overlay-yesno-btn--yes:hover:not(:disabled) {
  border-color: rgba(255, 210, 120, 0.7);
  background: linear-gradient(180deg, rgba(120, 90, 35, 0.85), rgba(85, 62, 25, 0.9));
  box-shadow: 0 0 12px rgba(255, 210, 120, 0.15);
}

.overlay-yesno-btn--no {
  border-color: rgba(180, 130, 90, 0.4);
  background: linear-gradient(180deg, rgba(80, 55, 25, 0.6), rgba(55, 38, 18, 0.7));
  color: #ffd98a;
}

.overlay-yesno-btn--no:hover:not(:disabled) {
  border-color: rgba(255, 200, 120, 0.6);
  background: linear-gradient(180deg, rgba(95, 65, 30, 0.75), rgba(65, 45, 22, 0.85));
  box-shadow: 0 0 12px rgba(255, 200, 120, 0.12);
}

.overlay-panel-body--cost {
  align-items: center;
  gap: 16px;
  padding: 24px 20px;
}

.overlay-cost-card {
  padding: 16px 24px;
  border-radius: 10px;
  border: 1px solid rgba(180, 150, 90, 0.3);
  background: rgba(14, 28, 44, 0.6);
  text-align: center;
}

.overlay-cost-text {
  color: rgba(225, 238, 249, 0.92);
  font-size: 0.9rem;
  line-height: 1.5;
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

.modal-enter-active,
.modal-leave-active {
  transition: opacity 0.24s ease;
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}

.modal-enter-active .overlay-panel,
.modal-leave-active .overlay-panel {
  transition: transform 0.24s ease;
}

.modal-enter-from .overlay-panel,
.modal-leave-to .overlay-panel {
  transform: scale(0.95) translateY(8px);
}

@media (max-width: 640px) {
  .overlay-panel {
    width: calc(100vw - 1.5rem);
    max-height: 88vh;
  }

  .overlay-panel-body {
    padding: 14px 16px;
  }

  .overlay-numeric-tile {
    width: 52px;
    min-height: 52px;
  }
}
</style>
