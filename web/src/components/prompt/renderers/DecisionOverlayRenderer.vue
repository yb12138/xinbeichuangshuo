<script setup lang="ts">
import { computed, inject, ref, watch } from 'vue'
import { DECISION_DOCKED } from '../dockContext'
import { optionPreview, sharedConsequence } from '../optionPreview'
import { MARK_GLYPHS } from '../../marks/markGlyphs'
type DecisionOverlayMode = 'numeric' | 'text' | 'activation-cost' | 'yes-no'

type DecisionOverlayOption = {
  id: string
  label: string
  buttonLabel: string
  hint?: string
  disabled?: boolean
  disabledReason?: string
}

const props = defineProps<{
  visible: boolean
  title: string
  context?: string
  mode: DecisionOverlayMode
  options: DecisionOverlayOption[]
  activationHint?: string
  activationOptionId?: string
  activationDisabled?: boolean
  canCancel: boolean
  cancelLabel: string
  cancelOptionId: string
  /** 停靠时用于后果预览（"手牌 4→5"）。 */
  handCount?: number | null
  /** 技能完整描述，用于在选项卡下方显示共同后果（"然后…"）。 */
  skillDescription?: string
}>()

const emit = defineEmits<{
  select: [optionId: string]
  cancel: [optionId: string]
}>()

function overlayDecisionOptionTitle(option: DecisionOverlayOption): string {
  const label = String(option.label || '').trim()
  const buttonLabel = String(option.buttonLabel || '').trim()
  if (!label) return buttonLabel
  if (!buttonLabel || buttonLabel === label) return label
  if (/^\d+$/.test(buttonLabel)) return label
  return buttonLabel
}

const docked = inject(DECISION_DOCKED, false)

// —— 停靠模式（决策面板规范图模板 1）：选项卡先点选、再在底部操作条确认 ——
const picked = ref('')
watch(
  () => props.options.map((o) => o.id).join('|'),
  () => {
    picked.value = ''
  },
)

/** 文字分支在停靠时画成选项卡；其余模式沿用原有按钮。 */
const dockedCards = computed(() => docked && props.mode === 'text')

const cards = computed(() =>
  props.options.map((option) => {
    const preview = optionPreview(overlayDecisionOptionTitle(option), option.hint, props.handCount ?? null)
    return { option, ...preview, path: MARK_GLYPHS[preview.glyph] || '' }
  }),
)

const shared = computed(() => (dockedCards.value ? sharedConsequence(props.skillDescription) : ''))

function pick(optionId: string) {
  picked.value = picked.value === optionId ? '' : optionId
}

function confirmPicked() {
  if (picked.value) emit('select', picked.value)
}

const pickedNumericHint = computed(() => {
  if (!picked.value) return ''
  const opt = props.options.find((o) => o.id === picked.value)
  return opt?.hint || opt?.label || ''
})
</script>

<template>
  <Teleport to="body" :disabled="docked">
    <Transition name="modal">
      <div
        v-if="visible"
        class="overlay-panel-root overlay-panel-root--decision" :class="{ 'overlay-panel-root--docked': docked }"
        data-testid="decision-overlay"
      >
        <div class="overlay-panel" @click.stop>
          <div class="overlay-panel-header overlay-panel-header--decision">
            <div v-if="context" class="overlay-panel-context" data-testid="prompt-context">{{ context }}</div>
            <h2>{{ title }}</h2>
          </div>

          <!-- ══════ DOCKED: activation-cost → effect preview + 放弃/发动 footer ══════ -->
          <template v-if="docked && mode === 'activation-cost'">
            <div class="overlay-panel-body decision-effect-preview">
              <div class="decision-effect-card">{{ activationHint }}</div>
            </div>
            <div class="decision-actions">
              <button v-if="canCancel" type="button" class="dc-pill decision-actions__ghost" data-testid="prompt-cancel-btn" @click="emit('cancel', cancelOptionId)">{{ cancelLabel || '放弃' }}</button>
              <span v-else />
              <button type="button" class="dc-pill dc-pill--gold decision-actions__gold" data-testid="prompt-option-confirm" :disabled="!!activationDisabled" @click="emit('select', activationOptionId || '')">发动</button>
            </div>
          </template>

          <!-- ══════ DOCKED: yes-no → 放弃/发动 footer (no cards needed) ══════ -->
          <template v-else-if="docked && mode === 'yes-no'">
            <div class="decision-actions">
              <button
                v-for="option in options"
                :key="option.id"
                type="button"
                :class="option.id === '0' || option.id === 'yes' || /发动/.test(option.label) ? 'dc-pill dc-pill--gold decision-actions__gold' : 'dc-pill decision-actions__ghost'"
                :data-testid="`prompt-option-${option.id}`"
                :disabled="!!option.disabled"
                @click="emit('select', option.id)"
              >
                {{ String(option.label || '').trim() }}
              </button>
            </div>
          </template>

          <!-- ══════ DOCKED: numeric → tiles + consequence line ══════ -->
          <template v-else-if="docked && mode === 'numeric'">
            <div class="overlay-panel-body decision-numeric">
              <div class="dc-nums decision-numeric__grid">
                <button
                  v-for="option in options"
                  :key="option.id"
                  type="button"
                  class="dc-num decision-numeric__tile"
                  :class="{ 'dc-num--picked decision-numeric__tile--picked': picked === option.id }"
                  :data-testid="`numeric-option-${option.buttonLabel}`"
                  :aria-pressed="picked === option.id"
                  :disabled="!!option.disabled"
                  @click="pick(option.id)"
                >
                  <span>{{ option.buttonLabel }}</span>
                </button>
              </div>
              <div v-if="pickedNumericHint" class="dc-note decision-numeric__hint">
                <span class="dc-note__label">代价</span>{{ pickedNumericHint }}
              </div>
            </div>
            <div class="decision-actions">
              <button v-if="canCancel" type="button" class="dc-pill decision-actions__ghost" data-testid="prompt-cancel-btn" @click="emit('cancel', cancelOptionId)">{{ cancelLabel || '放弃' }}</button>
              <span v-else />
              <button type="button" class="dc-pill dc-pill--gold decision-actions__gold" data-testid="decision-confirm-btn" :disabled="!picked" @click="confirmPicked">确认</button>
            </div>
          </template>

          <div v-else-if="mode === 'activation-cost'" class="overlay-panel-body overlay-panel-body--cost">
            <div class="overlay-cost-card">
              <span class="overlay-cost-text">{{ activationHint }}</span>
            </div>
            <button
              class="dc-pill dc-pill--gold overlay-confirm-btn"
              :disabled="!!activationDisabled"
              @click="emit('select', activationOptionId || '')"
            >
              确认发动
            </button>
          </div>

          <div v-else-if="mode === 'numeric'" class="overlay-panel-body overlay-panel-body--numeric">
            <div class="dc-nums overlay-numeric-grid">
              <button
                v-for="option in options"
                :key="option.id"
                class="dc-num overlay-numeric-tile"
                :data-testid="`numeric-option-${option.buttonLabel}`"
                :disabled="!!option.disabled"
                @click="emit('select', option.id)"
              >
                <span class="overlay-numeric-value">{{ option.buttonLabel }}</span>
              </button>
            </div>
          </div>

          <div v-else-if="mode === 'yes-no'" class="overlay-panel-body overlay-panel-body--yesno">
            <div class="overlay-yesno-row">
              <button
                v-for="option in options"
                :key="option.id"
                class="dc-pill overlay-yesno-btn"
                :class="option.id === '0' || option.id === 'yes' ? 'dc-pill--gold overlay-yesno-btn--yes' : 'overlay-yesno-btn--no'"
                :data-testid="`prompt-option-${option.id}`"
                :disabled="!!option.disabled"
                @click="emit('select', option.id)"
              >
                {{ String(option.label || '').trim() }}
              </button>
            </div>
          </div>

          <div v-else-if="dockedCards" class="overlay-panel-body">
            <div class="dc-tiles decision-cards" :class="{ 'dc-tiles--list decision-cards--many': cards.length > 2 }">
              <button
                v-for="(card, idx) in cards"
                :key="card.option.id"
                type="button"
                class="dc-tile decision-card"
                :class="[`dc-tile--${card.tone} decision-card--${card.tone}`, { 'dc-tile--wide': cards.length <= 2, 'dc-tile--picked decision-card--picked': picked === card.option.id }]"
                :data-testid="`branch-option-${idx}`"
                :disabled="!!card.option.disabled"
                :aria-pressed="picked === card.option.id"
                @click="pick(card.option.id)"
              >
                <span class="dc-tile__icon decision-card__icon" aria-hidden="true">
                  <svg viewBox="0 0 24 24"><path :d="card.path" /></svg>
                </span>
                <span class="dc-tile__title decision-card__title" :data-testid="`prompt-option-${card.option.id}`">{{ card.title }}</span>
                <span
                  v-if="card.option.disabled && card.option.disabledReason"
                  class="dc-tile__desc dc-tile__reason decision-card__reason"
                  :data-testid="`decision-card-reason-${idx}`"
                >{{ card.option.disabledReason }}</span>
                <span v-else-if="card.desc" class="dc-tile__desc decision-card__desc">{{ card.desc }}</span>
              </button>
            </div>
          </div>

          <div v-else class="overlay-panel-body overlay-panel-body--text">
            <button
              v-for="(option, idx) in options"
              :key="option.id"
              class="overlay-panel-item overlay-panel-item--text"
              :data-testid="`branch-option-${idx}`"
              :disabled="!!option.disabled"
              @click="emit('select', option.id)"
            >
              <div class="overlay-panel-item-title" :data-testid="`prompt-option-${option.id}`">{{ overlayDecisionOptionTitle(option) }}</div>
              <div
                v-if="option.hint && option.hint !== overlayDecisionOptionTitle(option)"
                class="overlay-panel-item-desc"
              >{{ option.hint }}</div>
            </button>
          </div>

          <div v-if="dockedCards && shared" class="decision-shared" data-testid="decision-shared">{{ shared }}</div>
          <div v-if="dockedCards" class="decision-actions">
            <button
              v-if="canCancel"
              type="button"
              class="dc-pill decision-actions__ghost"
              data-testid="prompt-cancel-btn"
              @click="emit('cancel', cancelOptionId)"
            >
              {{ cancelLabel || '放弃' }}
            </button>
            <span v-else />
            <button
              type="button"
              class="dc-pill dc-pill--gold decision-actions__gold"
              data-testid="decision-confirm-btn"
              :disabled="!picked"
              @click="confirmPicked"
            >
              确认
            </button>
          </div>
          <div v-else-if="canCancel && mode !== 'yes-no' && !docked" class="overlay-panel-footer">
            <button class="dc-pill overlay-panel-cancel" data-testid="prompt-cancel-btn" @click="emit('cancel', cancelOptionId)">
              {{ cancelLabel || '取消' }}
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.overlay-panel-root {
  /* 停靠在右下角行动区，不遮挡、不模糊棋盘：能看清"谁 → 我，用什么" */
  position: fixed;
  right: max(12px, calc(var(--safe-right, 0px) + 8px));
  bottom: max(12px, calc(var(--safe-bottom, 0px) + 8px));
  z-index: 13050;
  display: flex;
  align-items: flex-end;
  justify-content: flex-end;
  pointer-events: none;
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
  background: linear-gradient(110deg, rgba(40, 56, 80, 0.9), rgba(80, 60, 30, 0.85));
  border-bottom: 1px solid rgba(180, 150, 90, 0.24);
  text-align: center;
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
  padding: 16px 20px;
  display: flex;
  flex-direction: column;
  gap: 10px;
  background: rgba(6, 14, 24, 0.4);
}

.overlay-panel-item {
  position: relative;
  width: 100%;
  text-align: left;
  padding: 14px 18px;
  border-radius: 10px;
  border: 1px solid rgba(150, 130, 80, 0.3);
  background: rgba(14, 28, 44, 0.56);
  box-shadow: inset 0 1px 0 rgba(255, 240, 200, 0.05);
  cursor: pointer;
  transition: all 0.18s ease;
  font-family: inherit;
  font-size: inherit;
  line-height: inherit;
}

.overlay-panel-item:hover:not(:disabled) {
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

.overlay-panel-item-title {
  margin-bottom: 4px;
  color: #ffd98a;
  font-size: 0.95rem;
  font-weight: 600;
}

.overlay-panel-item-desc {
  color: rgba(199, 219, 237, 0.78);
  font-size: 0.8rem;
  line-height: 1.5;
  white-space: pre-wrap;
  word-break: break-word;
}

.overlay-panel-footer {
  flex-shrink: 0;
  padding: 12px 20px 16px;
  text-align: center;
  background: rgba(6, 14, 24, 0.66);
  border-top: 1px solid rgba(150, 130, 80, 0.2);
}

.overlay-panel-body--numeric {
  padding: 20px 16px;
}

.overlay-numeric-grid {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 10px;
}

.overlay-panel-body--yesno {
  align-items: center;
  padding: 24px 20px;
}

.overlay-yesno-row {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
  width: 100%;
  max-width: 320px;
}

.overlay-panel-body--cost {
  align-items: center;
  gap: 16px;
  padding: 24px 20px;
}

.overlay-cost-card {
  padding: 16px 24px;
  border-radius: 10px;
  border: 1px solid rgba(180, 150, 90, 0.3);
  background: rgba(14, 28, 44, 0.6);
  text-align: center;
}

.overlay-cost-text {
  color: rgba(225, 238, 249, 0.92);
  font-size: 0.9rem;
  line-height: 1.5;
}

.modal-enter-active,
.modal-leave-active {
  transition: opacity 0.24s ease;
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}

.modal-enter-active .overlay-panel,
.modal-leave-active .overlay-panel {
  transition: transform 0.24s ease;
}

.modal-enter-from .overlay-panel,
.modal-leave-to .overlay-panel {
  transform: scale(0.95) translateY(8px);
}

@media (max-width: 640px) {
  .overlay-panel {
    width: calc(100vw - 1.5rem);
    max-height: 88vh;
  }

  .overlay-panel-body {
    padding: 14px 16px;
  }

}

/* 停靠面板：紧凑尺寸，金色强调当前待决策 */
.overlay-panel-root .overlay-panel {
  width: min(300px, calc(100vw - 24px));
  max-height: min(62vh, 520px);
  pointer-events: auto;
  border-color: #26354f;
  box-shadow:
    0 0 0 1px rgba(240, 195, 90, 0.35),
    0 14px 30px rgba(2, 8, 18, 0.6),
    inset 0 1px 0 rgba(255, 240, 200, 0.08);
}

.overlay-panel-root .overlay-panel-header {
  padding: 10px 14px 9px;
}

.overlay-panel-root .overlay-panel-body {
  padding: 10px 12px;
  gap: 8px;
}

.overlay-panel-root .overlay-panel-item {
  padding: 10px 12px;
}

.overlay-panel-context {
  margin-bottom: 3px;
  color: #f0c35a;
  font-size: 0.75rem;
  font-weight: 600;
  line-height: 1.4;
  letter-spacing: 0.02em;
  word-break: break-word;
}

@media (max-width: 640px) {
  .overlay-panel-root {
    left: max(8px, calc(var(--safe-left, 0px) + 6px));
    right: max(8px, calc(var(--safe-right, 0px) + 6px));
    bottom: max(8px, calc(var(--safe-bottom, 0px) + 6px));
    justify-content: center;
  }

  .overlay-panel-root .overlay-panel {
    width: 100%;
    max-height: 55vh;
  }
}

/* 停靠在决策面板内：原地渲染，去掉全屏遮罩与自身外框，由 DecisionDock 提供外壳 */
.overlay-panel-root.overlay-panel-root--docked {
  position: static;
  inset: auto;
  z-index: auto;
  display: block;
  padding: 0;
  background: none;
  pointer-events: auto;
}
.overlay-panel-root.overlay-panel-root--docked .overlay-panel {
  width: 100%;
  max-height: none;
  border: none;
  border-radius: 0;
  box-shadow: none;
  background: transparent;
}
.overlay-panel-root.overlay-panel-root--docked .overlay-panel-header {
  /* 决策面板的来由条已经写明技能与问题，这里不再重复 */
  display: none;
}
.overlay-panel-root.overlay-panel-root--docked .overlay-panel-body {
  padding: 0 4px;
  background: none;
}

/* —— 停靠模板 1：文字分支 → 横排选项卡 + 底部操作条 —— */
.decision-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  padding: 12px 0 0;
}
.decision-shared {
  margin-top: 10px;
  padding: 8px 12px;
  border-radius: 8px;
  font-size: 12px;
  line-height: 1.5;
  color: var(--bt-text-sub, #9aa4b5);
  background: rgba(9, 14, 26, 0.7);
  box-shadow: 0 0 0 1px var(--bt-hairline-soft, rgba(185, 193, 205, 0.22));
}

/* —— 停靠模板 2：是否发动 / 效果预览 —— */
.decision-effect-preview {
  padding: 4px 0 0;
}
.decision-effect-card {
  padding: 12px 16px;
  border-radius: 10px;
  background: rgba(9, 14, 26, 0.85);
  box-shadow: 0 0 0 1px var(--bt-chrome-mid, #7d8796);
  font-size: 14px;
  line-height: 1.5;
  color: var(--bt-text, #e8ecf4);
  text-align: center;
}

/* —— 停靠模板 6：选数字 —— */
.decision-numeric {
  padding: 4px 0 0;
}
</style>
