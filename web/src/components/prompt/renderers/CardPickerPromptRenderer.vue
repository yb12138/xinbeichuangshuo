<script setup lang="ts">
withDefaults(defineProps<{
  visible: boolean
  message: string
  effectHints?: string[]
  canConfirm: boolean
  showCancel: boolean
  confirmLabel?: string
  cancelLabel?: string
}>(), {
  confirmLabel: '确认',
  cancelLabel: '放弃',
  effectHints: () => [],
})

const emit = defineEmits<{
  confirm: []
  cancel: []
}>()
</script>

<template>
  <div v-if="visible" class="prompt-card-picker-entry" data-testid="card-picker-prompt">
    <div class="prompt-card-picker-hint">{{ message }}</div>
    <div v-if="effectHints.length > 0" class="prompt-card-picker-effects">
      <div v-for="hint in effectHints" :key="hint" class="prompt-card-picker-effect">
        {{ hint }}
      </div>
    </div>

    <div class="dc-actions prompt-card-picker-actions-row" :class="{ 'dc-actions--end': !showCancel }">
      <button
        v-if="showCancel"
        type="button"
        class="dc-pill prompt-card-picker-btn prompt-card-picker-btn--cancel"
        data-testid="prompt-cancel-btn"
        @click="emit('cancel')"
      >
        {{ cancelLabel }}
      </button>
      <button
        type="button"
        class="dc-pill dc-pill--gold prompt-card-picker-btn prompt-card-picker-btn--success"
        :disabled="!canConfirm"
        data-testid="prompt-confirm-btn"
        @click="emit('confirm')"
      >
        {{ confirmLabel }}
      </button>
    </div>
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

.prompt-card-picker-effects {
  display: flex;
  flex-direction: column;
  gap: 3px;
  width: 100%;
  padding: 2px 6px 0;
}

.prompt-card-picker-effect {
  text-align: center;
  color: rgba(228, 241, 252, 0.95);
  font-size: 11px;
  line-height: 1.35;
  white-space: pre-wrap;
  word-break: break-word;
}

.prompt-card-picker-actions-row {
  width: 100%;
}
</style>
