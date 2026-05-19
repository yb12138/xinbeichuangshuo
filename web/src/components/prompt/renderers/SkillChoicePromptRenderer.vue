<script setup lang="ts">
type SkillPromptButton = {
  id: string
  label: string
  disabled: boolean
  cancel: boolean
  imageSrc: string
  imageReady: boolean
  fallbackText: string
}

type SkillBranchOption = {
  id: string
  title: string
  description?: string
  cost?: string
  disabled: boolean
}

defineProps<{
  inlineVisible: boolean
  overlayVisible: boolean
  title: string
  buttons: SkillPromptButton[]
  branches: SkillBranchOption[]
}>()

const emit = defineEmits<{
  select: [optionId: string]
  imageError: [optionId: string]
}>()
</script>

<template>
  <div v-if="inlineVisible && buttons.length > 0" class="prompt-skill-list" data-testid="skill-choice-prompt">
    <div class="prompt-skill-row">
      <div class="prompt-skill-text" :title="title">{{ title }}</div>
      <div class="prompt-skill-actions">
        <button
          v-for="option in buttons"
          :key="option.id"
          class="prompt-skill-action action-image-btn"
          :class="[
            option.cancel ? 'prompt-skill-action--cancel' : 'prompt-skill-action--success',
            option.disabled ? 'prompt-skill-action--disabled' : ''
          ]"
          :disabled="option.disabled"
          :data-testid="`prompt-option-${option.id}`"
          :title="option.label"
          :aria-label="option.label"
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

  <Teleport to="body">
    <Transition name="modal">
      <div
        v-if="overlayVisible && branches.length > 0"
        class="overlay-panel-root overlay-panel-root--skill"
        data-testid="skill-branch-overlay"
      >
        <div class="overlay-panel" data-testid="decision-overlay" @click.stop>
          <div class="overlay-panel-header">
            <h2>{{ title }}</h2>
          </div>
          <div class="overlay-panel-body">
            <button
              v-for="(entry, idx) in branches"
              :key="entry.id"
              class="overlay-panel-item"
              :data-testid="`branch-option-${idx}`"
              :disabled="entry.disabled"
              @click="emit('select', entry.id)"
            >
              <div class="overlay-panel-item-title" :data-testid="`prompt-option-${entry.id}`">{{ entry.title }}</div>
              <div v-if="entry.description" class="overlay-panel-item-desc">{{ entry.description }}</div>
              <div v-if="entry.cost" class="overlay-panel-item-cost">{{ entry.cost }}</div>
            </button>
          </div>
          <div class="overlay-panel-footer">
            <button class="overlay-panel-cancel" data-testid="prompt-option-skip" @click="emit('select', 'skip')">
              跳过
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.prompt-skill-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 2px;
}

.prompt-skill-row {
  display: flex;
  flex-direction: column;
  align-items: stretch;
  gap: 8px;
  padding: 2px;
}

.prompt-skill-actions {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 8px;
  align-items: center;
}

.prompt-skill-text {
  font-size: 13px;
  line-height: 1.3;
  color: rgba(221, 237, 248, 0.94);
  letter-spacing: 0.01em;
  text-align: center;
  white-space: normal;
  word-break: break-word;
}

.prompt-skill-action {
  -webkit-appearance: none;
  appearance: none;
  justify-self: center;
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
  flex-shrink: 0;
  color: #f6ecda;
  text-shadow: 0 1px 2px rgba(8, 8, 12, 0.65);
  background-repeat: no-repeat;
  background-size: cover;
  background-position: center;
  cursor: pointer;
}

.prompt-skill-action:hover:not(:disabled) {
  transform: translateY(-1px);
  filter: brightness(1.08);
}

.prompt-skill-action:focus,
.prompt-skill-action:focus-visible {
  outline: none;
  box-shadow: none;
}

.prompt-skill-action--success {
  border-color: rgba(111, 185, 141, 0.52);
}

.prompt-skill-action--cancel {
  border-color: rgba(196, 152, 102, 0.56);
}

.prompt-skill-action--disabled {
  opacity: 0.45;
  cursor: not-allowed;
  transform: none !important;
  filter: none !important;
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

.overlay-panel-root--skill {
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
  padding: 18px 24px 14px;
  border-bottom: 1px solid rgba(180, 150, 90, 0.24);
  text-align: center;
  background: linear-gradient(110deg, rgba(40, 56, 80, 0.9), rgba(80, 60, 30, 0.85));
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
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 14px;
  background: rgba(6, 14, 24, 0.4);
}

.overlay-panel-item {
  display: flex;
  flex-direction: column;
  gap: 6px;
  width: 100%;
  padding: 12px 14px;
  border-radius: 10px;
  border: 1px solid rgba(149, 186, 204, 0.26);
  background: rgba(15, 28, 46, 0.68);
  color: #e3eef8;
  text-align: left;
  cursor: pointer;
  transition: transform 0.16s ease, border-color 0.16s ease, background 0.16s ease;
  font-family: inherit;
}

.overlay-panel-item:hover:not(:disabled) {
  transform: translateY(-1px);
  border-color: rgba(255, 210, 120, 0.62);
  background: rgba(26, 42, 62, 0.84);
}

.overlay-panel-item:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

.overlay-panel-item-title {
  font-size: 0.95rem;
  font-weight: 700;
  color: #ffe2ad;
}

.overlay-panel-item-desc,
.overlay-panel-item-cost {
  font-size: 0.82rem;
  line-height: 1.35;
  color: rgba(226, 238, 248, 0.78);
}

.overlay-panel-item-cost {
  color: rgba(255, 217, 138, 0.86);
}

.overlay-panel-footer {
  flex-shrink: 0;
  padding: 12px 16px 16px;
  display: flex;
  justify-content: center;
  border-top: 1px solid rgba(149, 186, 204, 0.18);
}

.overlay-panel-cancel {
  min-width: 96px;
  padding: 8px 18px;
  border-radius: 10px;
  border: 1px solid rgba(196, 152, 102, 0.5);
  background: rgba(54, 38, 28, 0.76);
  color: #f4dfc1;
  font-size: 0.9rem;
  font-weight: 650;
  cursor: pointer;
  font-family: inherit;
}

.overlay-panel-cancel:hover {
  border-color: rgba(255, 210, 120, 0.68);
  background: rgba(76, 52, 34, 0.86);
}

@media (max-width: 900px) {
  .prompt-skill-action {
    width: 72px;
    height: 72px;
    max-width: 72px;
  }

  .overlay-panel {
    width: min(94vw, 480px);
  }
}
</style>
