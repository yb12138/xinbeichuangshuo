<script setup lang="ts">
type ResponseActionKind = 'take' | 'counter' | 'defend'

type ResponsePromptOption = {
  id: string
  buttonLabel: string
  disabled: boolean
  kind: ResponseActionKind
  imageSrc: string
  imageReady: boolean
  fallbackText: string
  enlarged: boolean
}

defineProps<{
  visible: boolean
  hint: string
  options: ResponsePromptOption[]
}>()

const emit = defineEmits<{
  select: [optionId: string]
  imageError: [optionId: string]
}>()

function responseButtonClass(option: ResponsePromptOption): string {
  if (option.kind === 'take') return 'prompt-response-btn--take'
  if (option.kind === 'counter') return 'prompt-response-btn--counter'
  return 'prompt-response-btn--defend'
}
</script>

<template>
  <div v-if="visible && options.length > 0" class="prompt-response-root" data-testid="response-prompt">
    <div v-if="hint" class="prompt-response-hint prompt-response-hint--attack-element">
      {{ hint }}
    </div>
    <div class="prompt-response-grid">
      <div
        v-for="option in options"
        :key="option.id"
        class="prompt-response-entry"
      >
        <button
          class="prompt-response-btn"
          :class="[
            responseButtonClass(option),
            option.enlarged ? 'prompt-response-btn--large' : '',
            option.disabled ? 'prompt-response-btn--disabled' : ''
          ]"
          :data-testid="`prompt-option-${option.id}`"
          :disabled="option.disabled"
          :title="option.buttonLabel"
          :aria-label="option.buttonLabel"
          @click="emit('select', option.id)"
        >
          <img
            v-if="option.imageReady"
            class="action-image-btn-fill"
            :src="option.imageSrc"
            alt=""
            @error="emit('imageError', option.id)"
          />
          <span v-else class="action-image-fallback-text">{{ option.fallbackText }}</span>
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.prompt-response-root {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.prompt-response-hint {
  min-height: 18px;
  padding: 0 4px;
  text-align: center;
  color: rgba(199, 219, 237, 0.88);
  font-size: 11px;
  line-height: 1.35;
  letter-spacing: 0.01em;
}

.prompt-response-hint--attack-element {
  min-height: 0;
  margin: 0 0 6px;
  padding: 4px 8px;
  border-radius: 9px;
  border: 1px solid rgba(132, 165, 187, 0.36);
  background: linear-gradient(180deg, rgba(28, 43, 61, 0.72), rgba(17, 30, 45, 0.75));
  color: rgba(223, 239, 250, 0.96);
  font-size: 12px;
  font-weight: 640;
  letter-spacing: 0.01em;
}

.prompt-response-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 8px;
  align-items: center;
  justify-items: center;
}

.prompt-response-entry {
  display: flex;
  flex-direction: column;
  gap: 4px;
  align-items: center;
}

.prompt-response-btn {
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
  justify-self: center;
  flex-shrink: 0;
  color: #f6ecda;
  text-shadow: 0 1px 2px rgba(8, 8, 12, 0.65);
  background-repeat: no-repeat;
  background-size: cover;
  background-position: center;
  cursor: pointer;
}

.prompt-response-btn:focus,
.prompt-response-btn:focus-visible {
  outline: none;
  box-shadow: none;
}

.prompt-response-btn--large {
  width: 96px;
  height: 96px;
  max-width: 96px;
  border-radius: 14px;
}

.prompt-response-btn--take {
  border-color: rgba(205, 171, 113, 0.68);
  background-image: url('/assets/ui/prompt_btn_take.png');
}

.prompt-response-btn--counter {
  border-color: rgba(157, 141, 228, 0.56);
  background-image: url('/assets/ui/prompt_btn_counter.png');
}

.prompt-response-btn--defend {
  border-color: rgba(111, 170, 225, 0.6);
  background-image: url('/assets/ui/prompt_btn_defend.png');
}

.prompt-response-btn--disabled {
  opacity: 0.45;
  cursor: not-allowed;
  transform: none;
  filter: none;
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

@media (max-width: 900px) {
  .prompt-response-btn--large {
    width: 88px;
    height: 88px;
    max-width: 88px;
  }
}

@media (max-width: 560px) {
  .prompt-response-grid {
    grid-template-columns: 1fr;
  }

  .prompt-response-btn,
  .prompt-response-btn--large {
    width: 84px;
    height: 84px;
    max-width: 84px;
  }
}
</style>
