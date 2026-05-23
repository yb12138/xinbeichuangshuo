<script setup lang="ts">
defineProps<{
  visible: boolean
  message: string
  showConfirm: boolean
  canConfirm: boolean
  confirmImageSrc: string
  confirmImageReady: boolean
  confirmFallbackText?: string
}>()

const emit = defineEmits<{
  confirm: []
  confirmImageError: []
}>()
</script>

<template>
  <div v-if="visible" class="prompt-target-entry" data-testid="target-picker-prompt">
    <div class="prompt-target-hint">{{ message }}</div>
    <button
      v-if="showConfirm"
      class="prompt-target-confirm-btn"
      :class="{ 'prompt-target-confirm-btn--disabled': !canConfirm }"
      :disabled="!canConfirm"
      data-testid="prompt-confirm-btn"
      title="确认"
      aria-label="确认"
      @click="emit('confirm')"
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
</template>

<style scoped>
.prompt-target-entry {
  display: flex;
  flex-direction: column;
  gap: 4px;
  align-items: center;
}

.prompt-target-hint {
  min-height: 18px;
  padding: 0 4px;
  text-align: center;
  color: rgba(199, 219, 237, 0.88);
  font-size: 11px;
  line-height: 1.35;
  letter-spacing: 0.01em;
}

.prompt-target-confirm-btn {
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

.prompt-target-confirm-btn:focus,
.prompt-target-confirm-btn:focus-visible {
  outline: none;
  box-shadow: none;
}

.prompt-target-confirm-btn--disabled {
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
</style>
