<script setup lang="ts">
/**
 * 响应攻击（按钮组件规范图 ②）：命中 / 应战·传递 / 防御 三种竖向选项卡，
 * 色调区分后果（红 = 承受、紫 = 反击、金 = 抵挡），卡上直接写后果；不可选时灰化并写原因。
 */
type ResponseActionKind = 'take' | 'counter' | 'defend'

type ResponsePromptOption = {
  id: string
  buttonLabel: string
  disabled: boolean
  kind: ResponseActionKind
  /** 卡片标题；空 = 按 kind 取默认（命中 / 应战·传递 / 防御）。 */
  title?: string
  /** 后果一句话（"承受 2 点风系伤害"）。 */
  consequence?: string
  /** 不可选原因。 */
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

/** viewBox 0 0 24 24 单色剪影。 */
const ICON: Record<ResponseActionKind, string> = {
  // 碎裂的心
  take: 'M12 20.5l-1.4-1.3C5.4 14.6 2 11.5 2 7.7 2 4.6 4.4 2.3 7.4 2.3c1.3 0 2.6.5 3.6 1.3l-1.6 3.6 3 2.2-1.5 3.9 3.6-4.6-2.9-2.1 1.2-2.4c.8-.5 1.7-.7 2.7-.7 3 0 5.5 2.3 5.5 5.4 0 3.8-3.4 6.9-8.6 11.5L12 20.5z',
  // 交叉双剑 + 回旋
  counter: 'M4.2 2.8l6.2 6.2-1.4 1.4L2.8 4.2V2.8h1.4zm15.6 0h1.4v1.4L9.6 15.8l1.6 1.6-1.4 1.4-1.6-1.6-2.4 2.4H3.6v-2.2L6 15l-1.6-1.6 1.4-1.4 1.6 1.6L19.8 2.8zM14.4 13l1.4-1.4 5.4 5.4v2.2H19l-5.3-5.3.7-.9zM17.5 7.6a5 5 0 011.4 5.1l1.8.6a6.9 6.9 0 00-2-7l1-1H16v3.7l1.5-1.4z',
  // 盾徽
  defend: 'M12 2l8 3v6.1c0 5-3.4 9.4-8 10.9-4.6-1.5-8-5.9-8-10.9V5l8-3zm0 2.2L6 6.4v4.7c0 3.9 2.5 7.4 6 8.7 3.5-1.3 6-4.8 6-8.7V6.4l-6-2.2zm0 2.6l3.8 1.4v2.9c0 2.5-1.5 4.7-3.8 5.8-2.3-1.1-3.8-3.3-3.8-5.8V8.2L12 6.8z',
}

function titleOf(option: ResponsePromptOption): string {
  return option.title || DEFAULT_TITLE[option.kind]
}
</script>

<template>
  <div v-if="visible && options.length > 0" class="prompt-response-root" data-testid="response-prompt">
    <div v-if="hint" class="dc-banner" data-testid="response-prompt-hint">{{ hint }}</div>
    <div class="dc-tiles">
      <button
        v-for="option in options"
        :key="option.id"
        type="button"
        class="dc-tile dc-tile--tall"
        :class="`dc-tile--${option.kind}`"
        :data-testid="`prompt-option-${option.id}`"
        :data-kind="option.kind"
        :disabled="option.disabled"
        :aria-label="[titleOf(option), option.disabled ? option.disabledReason : option.consequence].filter(Boolean).join('：')"
        @click="emit('select', option.id)"
      >
        <span class="dc-tile__icon" aria-hidden="true">
          <svg viewBox="0 0 24 24"><path :d="ICON[option.kind]" /></svg>
        </span>
        <span class="dc-tile__title">{{ titleOf(option) }}</span>
        <span v-if="option.disabled && option.disabledReason" class="dc-tile__desc dc-tile__reason">{{ option.disabledReason }}</span>
        <span v-else-if="option.consequence" class="dc-tile__desc">{{ option.consequence }}</span>
      </button>
    </div>
  </div>
</template>

<style scoped>
.prompt-response-root {
  display: flex;
  flex-direction: column;
}
</style>
