<script setup lang="ts">
/**
 * 响应攻击：命中 / 应战·传递 / 防御 → 一行小按钮（和「是 / 否」同尺寸），
 * 色调区分后果：命中 = 红、应战 = 默认银、防御 = 金。不可选时灰化。
 */
type ResponseActionKind = 'take' | 'counter' | 'defend'

type ResponsePromptOption = {
  id: string
  buttonLabel: string
  disabled: boolean
  kind: ResponseActionKind
  title?: string
  consequence?: string
  disabledReason?: string
}

defineProps<{
  visible: boolean
  hint: string
  options: ResponsePromptOption[]
}>()

const emit = defineEmits<{
  select: [optionId: string]
}>()

const DEFAULT_TITLE: Record<ResponseActionKind, string> = {
  take: '命中',
  counter: '应战 / 传递',
  defend: '防御',
}

const PILL_CLASS: Record<ResponseActionKind, string> = {
  take: 'dc-pill dc-pill--danger',
  counter: 'dc-pill',
  defend: 'dc-pill dc-pill--gold',
}

function titleOf(option: ResponsePromptOption): string {
  return option.title || DEFAULT_TITLE[option.kind]
}
function pillClass(option: ResponsePromptOption): string {
  return PILL_CLASS[option.kind] || 'dc-pill'
}
</script>

<template>
  <div v-if="visible && options.length > 0" class="prompt-response-root" data-testid="response-prompt">
    <div v-if="hint" class="response-hint" data-testid="response-hint">{{ hint }}</div>
    <div class="response-actions" data-testid="response-actions">
      <button
        v-for="option in options"
        :key="option.id"
        type="button"
        :class="pillClass(option)"
        :data-testid="`response-${option.kind}`"
        :disabled="option.disabled"
        :title="option.consequence || (option.disabled && option.disabledReason ? option.disabledReason : undefined)"
        @click="emit('select', option.id)"
      >
        {{ titleOf(option) }}
      </button>
    </div>
  </div>
</template>

<style scoped>
.prompt-response-root {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
}
.response-hint {
  font-size: 13px;
  color: var(--bt-text-sub, #9aa4b5);
  text-align: center;
}
.response-actions {
  display: flex;
  justify-content: center;
  gap: 12px;
  flex-wrap: wrap;
}
</style>
