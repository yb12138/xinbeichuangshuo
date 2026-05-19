<script setup lang="ts">
withDefaults(defineProps<{
  visible: boolean
  message: string
  canConfirm: boolean
  showCancel: boolean
  confirmTitle?: string
  confirmAriaLabel?: string
  confirmImageSrc: string
  confirmImageReady: boolean
  confirmFallbackText?: string
  cancelImageSrc: string
  cancelImageReady: boolean
  cancelFallbackText?: string
  cancelTitle?: string
  cancelAriaLabel?: string
}>(), {
  confirmTitle: '发动',
  confirmAriaLabel: '发动',
  confirmFallbackText: '确',
  cancelTitle: '',
  cancelAriaLabel: '',
  cancelFallbackText: '消',
})

const emit = defineEmits<{
  confirm: []
  cancel: []
  confirmImageError: []
  cancelImageError: []
}>()
</script>

<template>
  <div v-if="visible" class="prompt-card-picker-entry" data-testid="card-picker-prompt">
    <div class="prompt-card-picker-hint">{{ message }}</div>

    <div v-if="showCancel" class="prompt-card-picker-actions-row">
      <button
        class="prompt-card-picker-btn prompt-card-picker-btn--success action-image-btn"
        :class="{ 'prompt-card-picker-btn--disabled': !canConfirm }"
        :disabled="!canConfirm"
        data-testid="prompt-confirm-btn"
        :title="confirmTitle"
        :aria-label="confirmAriaLabel"
        @click="emit('confirm')"
      >
        <img
          v-if="confirmImageReady"
          class="action-image-btn-fill"
          :src="confirmImageSrc"
          alt=""
          @error="emit('confirmImageError')"
        />
        <span v-else class="action-image-fallback-text">{{ confirmFallbackText }}</span>
      </button>

      <button
        class="prompt-card-picker-btn prompt-card-picker-btn--cancel action-image-btn"
        data-testid="prompt-cancel-btn"
        :title="cancelTitle || undefined"
        :aria-label="cancelAriaLabel || undefined"
        @click="emit('cancel')"
      >
        <img
          v-if="cancelImageReady"
          class="action-image-btn-fill"
          :src="cancelImageSrc"
          alt=""
          @error="emit('cancelImageError')"
        />
        <span v-else class="action-image-fallback-text">{{ cancelFallbackText }}</span>
      </button>
    </div>

    <button
      v-else
      class="prompt-card-picker-btn prompt-card-picker-btn--success action-image-btn"
      :class="{ 'prompt-card-picker-btn--disabled': !canConfirm }"
      :disabled="!canConfirm"
      data-testid="prompt-confirm-btn"
      :title="confirmTitle"
      :aria-label="confirmAriaLabel"
      @click="emit('confirm')"
    >
      <img
        v-if="confirmImageReady"
        class="action-image-btn-fill"
        :src="confirmImageSrc"
        alt=""
        @error="emit('confirmImageError')"
      />
      <span v-else class="action-image-fallback-text">{{ confirmFallbackText }}</span>
    </button>
  </div>
</template>

<style scoped>
.prompt-card-picker-entry {
  display: flex;
  flex-direction: column;
  gap: 4px;
  align-items: center;
}

.prompt-card-picker-hint {
  min-height: 18px;
  padding: 0 4px;
  text-align: center;
  color: rgba(199, 219, 237, 0.88);
  font-size: 11px;
  line-height: 1.35;
  letter-spacing: 0.01em;
}

.prompt-card-picker-actions-row {
  width: 100%;
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 8px;
  justify-items: center;
  align-items: center;
}

.prompt-card-picker-btn {
  -webkit-appearance: none;
  appearance: none;
  width: 72px;
  height: 72px;
  min-height: 0;
  max-width: 72px;
  aspect-ratio: 1 / 1;
  border-radius: 12px;
  align-self: center;
  justify-self: center;
  flex-shrink: 0;
}

.action-image-btn {
  border: none;
  background: transparent;
  box-shadow: none;
  padding: 0;
  overflow: hidden;
  position: relative;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
}

.action-image-btn:focus,
.action-image-btn:focus-visible {
  outline: none;
  box-shadow: none;
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

.prompt-card-picker-btn--disabled {
  opacity: 0.45;
  cursor: not-allowed;
  transform: none;
  filter: none;
}
</style>
