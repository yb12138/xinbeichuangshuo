<script setup lang="ts">
/**
 * 应战两步流程：① 选应战牌 → ② 选应战目标。
 * 与右下角攻击行动坞的「选牌 › 选目标」步骤条、牌槽 → 目标槽流程图保持同一套视觉。
 * 牌和目标都在棋盘上点选：选中可应战的手牌后自动进入第二步，点选目标角色即提交。
 */
type CounterFlowStage = 'card' | 'target'

const props = defineProps<{
  stage: CounterFlowStage
  /** 魔弹传递：文案改成「传递」 */
  isMagicMissile: boolean
  /** 已选应战牌的牌面图（第二步显示在牌槽里） */
  cardArt: string
  cardName: string
  /** 可应战牌的规则说明，如「打出风系攻击或暗灭」 */
  rule: string
}>()

const emit = defineEmits<{
  back: []
}>()

function verb(): string {
  return props.isMagicMissile ? '传递' : '应战'
}
</script>

<template>
  <div class="counter-flow" data-testid="counter-flow">
    <header class="counter-flow-hd">
      <span class="counter-flow-title">{{ verb() }}</span>
      <nav class="counter-flow-steps" aria-label="应战步骤">
        <span :class="stage === 'card' ? 'on' : 'done'" data-testid="counter-step-card">
          ① 选牌{{ stage === 'card' ? '' : ' ✓' }}
        </span>
        <i class="counter-flow-sep" aria-hidden="true">›</i>
        <span :class="stage === 'target' ? 'on' : ''" data-testid="counter-step-target">② 选目标</span>
      </nav>
    </header>

    <div class="counter-flow-bd" aria-hidden="true">
      <span class="counter-slot counter-slot--card" :class="stage === 'card' ? 'counter-slot--pulse' : 'counter-slot--filled'">
        <svg v-if="stage === 'card' || !cardArt" class="counter-slot-ico" viewBox="0 0 24 24">
          <rect x="6.5" y="4" width="11" height="16" rx="1.6" fill="none" stroke="currentColor" stroke-width="1.3" />
          <path d="M9.5 9h5M9.5 12h5" stroke="currentColor" stroke-width="1.1" stroke-linecap="round" fill="none" />
        </svg>
        <img v-else class="counter-slot-art" :src="cardArt" :alt="cardName" />
      </span>
      <span class="counter-flow-arrow">➤</span>
      <span class="counter-slot counter-slot--target" :class="stage === 'target' ? 'counter-slot--pulse' : ''">
        <svg class="counter-slot-ico" viewBox="0 0 24 24">
          <circle cx="12" cy="12" r="6.4" fill="none" stroke="currentColor" stroke-width="1.3" />
          <circle cx="12" cy="12" r="2.2" fill="none" stroke="currentColor" stroke-width="1.3" />
          <path d="M12 2.6v3.2M12 18.2v3.2M2.6 12h3.2M18.2 12h3.2" stroke="currentColor" stroke-width="1.3" stroke-linecap="round" fill="none" />
        </svg>
      </span>
    </div>

    <div class="counter-flow-hint" data-testid="counter-flow-hint">
      <template v-if="stage === 'card'">先在下方手牌选一张{{ isMagicMissile ? '【魔弹】' : '应战牌' }}<span v-if="rule" class="counter-flow-rule">（{{ rule }}）</span></template>
      <template v-else>已选{{ cardName ? `【${cardName}】` : '牌' }}，请直接点击战场角色立绘完成{{ verb() }}。</template>
    </div>

    <footer class="counter-flow-ft">
      <button type="button" class="dc-pill" data-testid="counter-flow-back" @click="emit('back')">
        {{ stage === 'target' ? '重新选牌' : '返回' }}
      </button>
    </footer>
  </div>
</template>

<style scoped>
.counter-flow {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  min-width: 260px;
}

.counter-flow-hd {
  align-self: stretch;
  display: flex;
  align-items: center;
  gap: 10px;
}

.counter-flow-title {
  font-size: 14px;
  font-weight: 700;
  color: var(--bt-text, #e8edf5);
  letter-spacing: 0.06em;
}

/* 与 ActionPanel .dock-steps 一致 */
.counter-flow-steps {
  margin-left: auto;
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 11px;
  color: var(--bt-text-dim);
  white-space: nowrap;
}
.counter-flow-steps .on {
  color: var(--bt-active);
  font-weight: 700;
}
.counter-flow-steps .done {
  color: var(--bt-select);
}
.counter-flow-sep {
  font-style: normal;
  color: var(--bt-chrome-lo);
}

/* 与 ActionPanel .dock-flow / .dock-slot 一致 */
.counter-flow-bd {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
}

.counter-slot {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: var(--bt-bg-0);
  color: var(--bt-text-dim);
  box-shadow: inset 0 0 0 1px var(--bt-hairline-soft);
}

.counter-slot--card {
  width: 38px;
  height: 52px;
  border-radius: var(--bt-radius-sm);
}

.counter-slot--target {
  width: 46px;
  height: 46px;
  border-radius: 999px;
}

.counter-slot-ico {
  width: 24px;
  height: 24px;
}

.counter-slot-art {
  width: 100%;
  height: 100%;
  object-fit: cover;
  border-radius: inherit;
}

.counter-slot--filled {
  color: var(--bt-select);
  background: linear-gradient(180deg, var(--bt-blue-soft), var(--bt-bg-0));
  box-shadow: inset 0 0 0 1px var(--bt-select), 0 0 8px var(--bt-select-glow);
}

.counter-slot--pulse {
  color: var(--bt-active);
  box-shadow: inset 0 0 0 1px var(--bt-active), 0 0 0 0 var(--bt-active-glow);
  animation: counterSlotPulse 1.4s ease-in-out infinite;
}

.counter-flow-arrow {
  font-size: 12px;
  color: var(--bt-chrome-mid);
}

.counter-flow-hint {
  font-size: 13px;
  color: var(--bt-text-sub, #9aa4b5);
  text-align: center;
}

.counter-flow-rule {
  color: var(--bt-text-dim);
}

.counter-flow-ft {
  display: flex;
  justify-content: center;
}

@keyframes counterSlotPulse {
  0%, 100% { box-shadow: inset 0 0 0 1px var(--bt-active), 0 0 0 0 var(--bt-active-glow); }
  50% { box-shadow: inset 0 0 0 1px var(--bt-active), 0 0 12px 2px var(--bt-active-glow); }
}

@media (prefers-reduced-motion: reduce) {
  .counter-slot--pulse {
    animation: none;
  }
}
</style>
