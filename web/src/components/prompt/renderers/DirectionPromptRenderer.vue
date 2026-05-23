<script setup lang="ts">
type DirectionPromptOption = {
  id: string
  label: string
  hint?: string
  description?: string
  disabled?: boolean
  tone?: string
  icon?: string
}

defineProps<{
  visible: boolean
  title: string
  options: DirectionPromptOption[]
}>()

const emit = defineEmits<{
  select: [optionId: string]
}>()
</script>

<template>
  <Teleport to="body">
    <Transition name="modal">
      <div
        v-if="visible"
        class="overlay-panel-root overlay-panel-root--decision"
        data-testid="direction-prompt"
      >
        <div class="overlay-panel" data-testid="decision-overlay" @click.stop>
          <div class="overlay-panel-header overlay-panel-header--decision">
            <h2>{{ title }}</h2>
          </div>
          <div class="overlay-panel-body overlay-panel-body--text">
            <button
              v-for="(option, idx) in options"
              :key="option.id"
              class="overlay-panel-item overlay-panel-item--text"
              :class="[option.tone || '', option.icon ? 'overlay-panel-item--icon' : '']"
              :data-testid="`branch-option-${idx}`"
              :disabled="!!option.disabled"
              @click="emit('select', option.id)"
            >
              <div
                class="overlay-panel-item-title"
                :data-testid="`prompt-option-${option.id}`"
              >
                <span :data-testid="`direction-option-${option.id}`">{{ option.label }}</span>
              </div>
              <div
                v-if="option.hint || option.description"
                class="overlay-panel-item-desc"
                :data-testid="`direction-option-desc-${idx}`"
              >
                {{ option.hint || option.description }}
              </div>
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
  padding: 16px 20px;
  background: rgba(6, 14, 24, 0.4);
}

.overlay-panel-item {
  position: relative;
  width: 100%;
  min-height: 68px;
  padding: 14px 18px 14px 56px;
  border-radius: 10px;
  border: 1px solid rgba(150, 130, 80, 0.3);
  background: rgba(14, 28, 44, 0.56);
  box-shadow: inset 0 1px 0 rgba(255, 240, 200, 0.05);
  color: #e3eef8;
  text-align: left;
  cursor: pointer;
  transition: transform 0.18s ease, border-color 0.18s ease, background 0.18s ease, box-shadow 0.18s ease;
  font-family: inherit;
  font-size: inherit;
  line-height: inherit;
}

.overlay-panel-item:hover:not(:disabled) {
  transform: translateY(-1px);
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

.overlay-panel-item--icon::before {
  content: "→";
  position: absolute;
  left: 17px;
  top: 50%;
  width: 26px;
  height: 26px;
  transform: translateY(-50%);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 999px;
  border: 1px solid rgba(255, 210, 120, 0.38);
  background: rgba(10, 22, 38, 0.62);
  color: #ffd98a;
  font-size: 16px;
  font-weight: 800;
  line-height: 1;
}

.prompt-direction--reverse::before {
  content: "←";
}

.prompt-direction--reverse {
  border-color: rgba(132, 165, 210, 0.34);
}

.prompt-direction--reverse:hover:not(:disabled) {
  border-color: rgba(150, 190, 232, 0.62);
  box-shadow:
    0 0 12px rgba(150, 190, 232, 0.12),
    inset 0 1px 0 rgba(225, 240, 255, 0.08);
}

.overlay-panel-item-title {
  margin-bottom: 4px;
  color: #ffd98a;
  font-size: 0.95rem;
  font-weight: 700;
}

.overlay-panel-item-desc {
  color: rgba(199, 219, 237, 0.78);
  font-size: 0.8rem;
  line-height: 1.5;
  white-space: pre-wrap;
  word-break: break-word;
}

@media (max-width: 560px) {
  .overlay-panel {
    width: min(94vw, 480px);
  }

  .overlay-panel-body {
    padding: 14px;
  }
}
</style>
