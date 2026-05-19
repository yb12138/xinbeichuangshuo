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
