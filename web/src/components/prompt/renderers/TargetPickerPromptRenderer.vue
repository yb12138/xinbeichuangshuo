<script setup lang="ts">
defineProps<{
  visible: boolean
  message: string
  showConfirm: boolean
  canConfirm: boolean
}>()

const emit = defineEmits<{
  confirm: []
}>()
</script>

<template>
  <div v-if="visible" class="prompt-target-entry" data-testid="target-picker-prompt">
    <div class="prompt-target-hint">{{ message }}</div>
    <div v-if="showConfirm" class="dc-actions dc-actions--end prompt-target-actions">
      <button
        type="button"
        class="dc-pill dc-pill--gold prompt-target-confirm-btn"
        :disabled="!canConfirm"
        data-testid="prompt-confirm-btn"
        @click="emit('confirm')"
      >
        确认
      </button>
    </div>
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

.prompt-target-actions {
  width: 100%;
}
</style>
