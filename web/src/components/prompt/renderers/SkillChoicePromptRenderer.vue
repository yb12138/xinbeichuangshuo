<script setup lang="ts">
import { computed, inject, ref, watch } from 'vue'
import { DECISION_DOCKED } from '../dockContext'
type SkillPromptButton = {
  id: string
  label: string
  disabled: boolean
  cancel: boolean
}

type SkillBranchOption = {
  id: string
  title: string
  description?: string
  cost?: string
  disabled: boolean
}

const props = defineProps<{
  inlineVisible: boolean
  overlayVisible: boolean
  title: string
  context?: string
  buttons: SkillPromptButton[]
  branches: SkillBranchOption[]
}>()

const emit = defineEmits<{
  select: [optionId: string]
}>()

function hasSkipBranch(branches: SkillBranchOption[]): boolean {
  return branches.some((entry) => entry.id === 'skip' || entry.id === 'cancel')
}

const docked = inject(DECISION_DOCKED, false)

// —— 停靠模式：响应 / 启动技能按"是否发动"模板呈现 ——
type DockedSkill = { id: string; title: string; description: string; cost: string; disabled: boolean }

const isCancelId = (id: string) => id === 'skip' || id === 'cancel'

/** 两条渲染路线（内联按钮 / 分支弹窗）统一成"可发动的技能 + 一个放弃选项"。 */
const dockedModel = computed(() => {
  if (!docked) return null
  if (props.overlayVisible && props.branches.length > 0) {
    const skills: DockedSkill[] = props.branches
      .filter((b) => !isCancelId(b.id))
      .map((b) => ({ id: b.id, title: b.title, description: b.description || '', cost: b.cost || '', disabled: b.disabled }))
    const cancel = props.branches.find((b) => isCancelId(b.id))
    return { skills, cancelId: cancel?.id || 'skip', cancelLabel: cancel?.description || '不发动' }
  }
  if (props.inlineVisible && props.buttons.length > 0) {
    const skills: DockedSkill[] = props.buttons
      .filter((b) => !b.cancel)
      .map((b) => ({ id: b.id, title: b.label, description: '', cost: '', disabled: b.disabled }))
    const cancel = props.buttons.find((b) => b.cancel)
    return { skills, cancelId: cancel?.id || 'skip', cancelLabel: cancel?.label || '不发动' }
  }
  return null
})

const picked = ref('')
watch(
  () => dockedModel.value?.skills.map((s) => s.id).join('|'),
  () => {
    const only = dockedModel.value?.skills.length === 1 ? dockedModel.value.skills[0] : undefined
    picked.value = only && !only.disabled ? only.id : ''
  },
  { immediate: true },
)

function pick(id: string) {
  if ((dockedModel.value?.skills.length ?? 0) <= 1) return
  picked.value = picked.value === id ? '' : id
}
</script>

<template>
  <!-- ══════ 停靠：是否发动响应 / 启动技能 ══════ -->
  <div v-if="dockedModel" class="skill-docked" data-testid="skill-choice-docked">
    <div class="skill-docked__list" :class="{ 'skill-docked__list--many': dockedModel.skills.length > 1 }">
      <button
        v-for="(skill, idx) in dockedModel.skills"
        :key="skill.id"
        type="button"
        class="skill-docked__card"
        :class="{ 'skill-docked__card--picked': picked === skill.id, 'skill-docked__card--single': dockedModel.skills.length === 1 }"
        :data-testid="`branch-option-${idx}`"
        :disabled="skill.disabled"
        :aria-pressed="picked === skill.id"
        @click="pick(skill.id)"
      >
        <span class="skill-docked__head">
          <span class="skill-docked__title" :data-testid="`prompt-option-${skill.id}`">{{ skill.title }}</span>
          <span v-if="skill.cost" class="skill-docked__cost">{{ skill.cost }}</span>
        </span>
        <span v-if="skill.description" class="skill-docked__desc">{{ skill.description }}</span>
      </button>
    </div>
    <div class="skill-docked__actions">
      <button type="button" class="dc-pill skill-docked__ghost" data-testid="prompt-option-skip" @click="emit('select', dockedModel.cancelId)">
        {{ dockedModel.cancelLabel }}
      </button>
      <button
        type="button"
        class="dc-pill dc-pill--gold skill-docked__gold"
        data-testid="decision-confirm-btn"
        :disabled="!picked"
        @click="picked && emit('select', picked)"
      >
        发动
      </button>
    </div>
  </div>

  <div v-else-if="inlineVisible && buttons.length > 0" class="prompt-skill-list" data-testid="skill-choice-prompt">
    <div class="prompt-skill-row">
      <div class="prompt-skill-text" :title="title">{{ title }}</div>
      <div class="prompt-skill-actions">
        <button
          v-for="option in buttons"
          :key="option.id"
          type="button"
          class="dc-pill prompt-skill-action"
          :class="[
            option.cancel ? 'prompt-skill-action--cancel' : 'dc-pill--gold prompt-skill-action--success',
            option.disabled ? 'prompt-skill-action--disabled' : ''
          ]"
          :disabled="option.disabled"
          :data-testid="`prompt-option-${option.id}`"
          :aria-label="option.label"
          @click="emit('select', option.id)"
        >
          {{ option.cancel ? '放弃' : '发动' }}
        </button>
      </div>
    </div>
  </div>

  <Teleport to="body" :disabled="docked">
    <Transition name="modal">
      <div
        v-if="overlayVisible && branches.length > 0 && !dockedModel"
        class="overlay-panel-root overlay-panel-root--skill" :class="{ 'overlay-panel-root--docked': docked }"
        data-testid="skill-branch-overlay"
      >
        <div class="overlay-panel" data-testid="decision-overlay" @click.stop>
          <div class="overlay-panel-header">
            <div v-if="context" class="overlay-panel-context" data-testid="prompt-context">{{ context }}</div>
            <h2>{{ title }}</h2>
          </div>
          <div class="overlay-panel-body">
            <button
              v-for="(entry, idx) in branches"
              :key="entry.id"
              class="overlay-panel-item"
              :data-testid="`branch-option-${idx}`"
              :disabled="entry.disabled"
              @click="emit('select', entry.id)"
            >
              <div
                v-if="entry.title"
                class="overlay-panel-item-title"
                :data-testid="`prompt-option-${entry.id}`"
              >{{ entry.title }}</div>
              <div
                v-if="entry.description"
                class="overlay-panel-item-desc"
                :data-testid="!entry.title ? `prompt-option-${entry.id}` : undefined"
              >{{ entry.description }}</div>
              <div v-if="entry.cost" class="overlay-panel-item-cost">{{ entry.cost }}</div>
            </button>
          </div>
          <div v-if="!hasSkipBranch(branches)" class="overlay-panel-footer">
            <button class="dc-pill overlay-panel-cancel" data-testid="prompt-option-skip" @click="emit('select', 'skip')">
              跳过
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.prompt-skill-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 2px;
}

.prompt-skill-row {
  display: flex;
  flex-direction: column;
  align-items: stretch;
  gap: 8px;
  padding: 2px;
}

.prompt-skill-actions {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 8px;
  align-items: center;
}

.prompt-skill-text {
  font-size: 13px;
  line-height: 1.3;
  color: rgba(221, 237, 248, 0.94);
  letter-spacing: 0.01em;
  text-align: center;
  white-space: normal;
  word-break: break-word;
}

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
  padding: 14px;
  background: rgba(6, 14, 24, 0.4);
}

.overlay-panel-item {
  display: flex;
  flex-direction: column;
  gap: 6px;
  width: 100%;
  padding: 12px 14px;
  border-radius: 10px;
  border: 1px solid rgba(149, 186, 204, 0.26);
  background: rgba(15, 28, 46, 0.68);
  color: #e3eef8;
  text-align: left;
  cursor: pointer;
  transition: transform 0.16s ease, border-color 0.16s ease, background 0.16s ease;
  font-family: inherit;
}

.overlay-panel-item:hover:not(:disabled) {
  transform: translateY(-1px);
  border-color: rgba(255, 210, 120, 0.62);
  background: rgba(26, 42, 62, 0.84);
}

.overlay-panel-item:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

.overlay-panel-item-title {
  font-size: 0.95rem;
  font-weight: 700;
  color: #ffe2ad;
}

.overlay-panel-item-desc,
.overlay-panel-item-cost {
  font-size: 0.82rem;
  line-height: 1.35;
  color: rgba(226, 238, 248, 0.78);
}

.overlay-panel-item-cost {
  color: rgba(255, 217, 138, 0.86);
}

.overlay-panel-footer {
  flex-shrink: 0;
  padding: 12px 16px 16px;
  display: flex;
  justify-content: center;
  border-top: 1px solid rgba(149, 186, 204, 0.18);
}

@media (max-width: 900px) {

  .overlay-panel {
    width: min(94vw, 480px);
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

/* —— 停靠：是否发动技能 —— */
.skill-docked {
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.skill-docked__list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.skill-docked__card {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 12px 14px;
  border: none;
  border-radius: 10px;
  text-align: left;
  color: var(--bt-text, #e8ecf4);
  background: rgba(9, 14, 26, 0.85);
  box-shadow: 0 0 0 1px var(--bt-chrome-mid, #7d8796);
  cursor: pointer;
  transition: box-shadow 0.15s ease;
}
/* 只有一个技能时它就是"效果说明"，不需要点选 */
.skill-docked__card--single {
  cursor: default;
}
.skill-docked__list--many .skill-docked__card:hover:not(:disabled) {
  box-shadow: 0 0 0 1px var(--bt-select, #9fd4ff);
}
.skill-docked__card--picked:not(.skill-docked__card--single) {
  background: color-mix(in srgb, var(--bt-select, #9fd4ff) 12%, rgba(9, 14, 26, 0.85));
  box-shadow: 0 0 0 2px var(--bt-select, #9fd4ff), 0 0 16px var(--bt-select-glow, rgba(159, 212, 255, 0.45));
}
.skill-docked__card:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}
.skill-docked__head {
  display: flex;
  align-items: baseline;
  gap: 8px;
}
.skill-docked__title {
  font-size: 15px;
  font-weight: 800;
  color: var(--bt-active, #f0c35a);
}
.skill-docked__cost {
  font-size: 11px;
  color: var(--bt-select, #9fd4ff);
}
.skill-docked__desc {
  font-size: 12px;
  line-height: 1.5;
  color: var(--bt-text-sub, #9aa4b5);
}
.skill-docked__actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
