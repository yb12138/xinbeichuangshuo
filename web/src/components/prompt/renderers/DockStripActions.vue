<script setup lang="ts">
/**
 * 窄条操作行（决策面板规范图模板 3/4/5：选目标 / 选手牌 / 选扩展区盖牌）。
 * 棋盘本身是操作界面，这里只留规则胶囊、实时计数与放弃/确认。
 */
defineProps<{
  /** 一行提示，如"请选择要弃置的1张手牌"。 */
  message: string
  /** 规则胶囊（筛选条件、效果提示），可为空。 */
  rules?: string[]
  selected: number
  /** 需要的数量上限；0 = 不显示分母。 */
  max: number
  canConfirm: boolean
  showConfirm: boolean
  showCancel: boolean
  cancelLabel?: string
}>()

const emit = defineEmits<{
  confirm: []
  cancel: []
}>()
</script>

<template>
  <div class="dock-strip" data-testid="dock-strip-actions">
    <div class="dock-strip__info">
      <div class="dock-strip__message">{{ message }}</div>
      <div v-if="rules && rules.length" class="dock-strip__rules">
        <span v-for="rule in rules" :key="rule" class="dock-strip__rule">{{ rule }}</span>
      </div>
    </div>
    <div class="dock-strip__counter" aria-live="polite">
      已选 <span class="dock-strip__num">{{ selected }}</span><template v-if="max > 0">/{{ max }}</template>
    </div>
    <div class="dock-strip__btns">
      <button
        v-if="showCancel"
        type="button"
        class="dc-pill dock-strip__ghost"
        data-testid="prompt-cancel-btn"
        @click="emit('cancel')"
      >
        {{ cancelLabel || '放弃' }}
      </button>
      <button
        v-if="showConfirm"
        type="button"
        class="dc-pill dc-pill--gold dock-strip__gold"
        data-testid="prompt-confirm-btn"
        :disabled="!canConfirm"
        @click="emit('confirm')"
      >
        确认
      </button>
    </div>
  </div>
</template>

<style scoped>
.dock-strip {
  display: flex;
  align-items: center;
  gap: 14px;
}
.dock-strip__info {
  min-width: 0;
  flex: 1;
}
.dock-strip__message {
  font-size: 13px;
  line-height: 1.4;
  color: var(--bt-text, #e8ecf4);
}
.dock-strip__rules {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  margin-top: 4px;
}
.dock-strip__rule {
  padding: 1px 9px;
  border-radius: 999px;
  font-size: 11px;
  color: var(--bt-select, #9fd4ff);
  box-shadow: inset 0 0 0 1px color-mix(in srgb, var(--bt-select, #9fd4ff) 60%, transparent);
}
.dock-strip__counter {
  flex: none;
  font-size: 12px;
  color: var(--bt-text-sub, #9aa4b5);
  white-space: nowrap;
}
.dock-strip__num {
  font-family: var(--bt-font-num, Georgia, serif);
  font-size: 18px;
  font-weight: 800;
  color: var(--bt-select, #9fd4ff);
}
.dock-strip__btns {
  display: flex;
  gap: 8px;
  flex: none;
}
</style>
