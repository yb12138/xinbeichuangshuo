<script setup lang="ts">
/**
 * 浮动决策面板：所有"需要我做选择"的提示都在这里作答（行动选择枢纽除外，仍在右下角坞）。
 *
 * - 在面板内作答（分支 / 是否发动 / 数字 / 分配 / 方向）：放大面板，浮在棋盘中央偏下。
 * - 在棋盘上作答（选目标 / 选手牌 / 选扩展区盖牌）：窄条，贴在手牌上方，棋盘本身就是操作界面。
 * - 可拖拽（位置跨对局记住），可收起成胶囊；别人回合打断我时强制展开并显示红色打断色带。
 * 布局依据：outimage/决策面板布局规范图.png，示意页 public/mockups/decision-dock-layout.html。
 */
import { computed, onBeforeUnmount, provide, ref, watch } from 'vue'
import PromptDialog from './PromptDialog.vue'
import { DECISION_DOCKED } from './prompt/dockContext'
import { useInterruptStore } from '../stores/interrupt.store'
import { useSessionStore } from '../stores/session.store'
import { useSnapshotStore } from '../stores/snapshot.store'
import { useUiStore } from '../stores/ui.store'
import { characterArtSrc } from '../constants/characterArt'

provide(DECISION_DOCKED, true)

const interruptStore = useInterruptStore()
const sessionStore = useSessionStore()
const snapshotStore = useSnapshotStore()
const uiStore = useUiStore()

const prompt = computed(() => interruptStore.currentPrompt)
const myId = computed(() => sessionStore.myPlayerId)

/** 行动选择枢纽与爆牌弃牌仍由右下角 ActionPanel 处理。 */
const isOverflowDiscard = computed(() => {
  const p = prompt.value?.presentation
  return p?.discard_reason === 'hand_overflow' || (!p?.discard_reason && p?.card_filter === 'overflow_discard')
})
const active = computed(() => {
  const p = prompt.value
  if (!p || p.player_id !== myId.value) return false
  if (p.presentation?.kind === 'action_hub') return false
  return !isOverflowDiscard.value
})

/** 在棋盘上作答的提示：面板收成窄条，不挡座位与手牌。 */
const boardAnswered = computed(() => {
  const p = prompt.value?.presentation
  if (!p) return false
  // 应战 / 防御靠点手牌作答，面板只留命中与确认按钮
  if (p.kind === 'response' || p.kind === 'target_picker') return true
  if (p.kind === 'card_picker') return p.layout !== 'extract' && p.layout !== 'fraud_attack_element'
  return false
})

/** 别人回合里要我决策 = 打断。 */
const interrupted = computed(() => active.value && snapshotStore.currentPlayer !== myId.value)

// —— 来由条：谁、哪个技能、为什么 ——
function seatName(playerId: string): string {
  const player = snapshotStore.players[playerId]
  if (!player) return playerId
  return snapshotStore.characters[player.role || '']?.name || player.name || playerId
}

function skillTitle(skillId: string): string {
  if (!skillId) return ''
  for (const character of Object.values(snapshotStore.characters)) {
    const skill = character.skills?.find((s) => s.id === skillId)
    if (skill) return skill.title
  }
  return ''
}

/** 提示文案里的【技能名】：没有 skill_id 时从这里取标题。 */
const bracketTitle = computed(() => {
  const match = (prompt.value?.message || '').match(/【([^】]+)】/)
  return match?.[1] || ''
})

/** 没有技能名时按提示类型给一个能读懂的标题。 */
const KIND_TITLE: Record<string, string> = {
  response: '响应攻击',
  target_picker: '选择目标',
  card_picker: '选择手牌',
  numeric: '选择数值',
  skill_choice: '是否发动技能',
}

const header = computed(() => {
  const p = prompt.value
  if (!p) return null
  const attackerId = String(p.attacker_id || '').trim()
  const fromOther = attackerId && attackerId !== myId.value ? attackerId : ''
  const title = skillTitle(String(p.skill_id || '').trim()) || bracketTitle.value
  const message = String(p.message || '').replace(/^【[^】]+】\s*[:：]?\s*/, '').trim()
  // 头像：打断时是发起者，否则是我自己（我的技能在要我做选择）
  const avatarOwner = fromOther || myId.value
  return {
    title: title || KIND_TITLE[p.presentation?.kind || ''] || '请选择',
    message,
    avatar: characterArtSrc(snapshotStore.players[avatarOwner]?.role),
    who: fromOther ? `${seatName(fromOther)} → 你` : `${seatName(myId.value)}`,
    band: interrupted.value ? (fromOther ? `${seatName(fromOther)}${title ? `的【${title}】` : ''}要求你响应` : '对方回合中需要你做出选择') : '',
  }
})

// —— 收起：每次出现新的提示都重新展开；打断时强制展开 ——
const promptKey = computed(() => {
  const p = prompt.value
  return p ? `${p.choice_type || ''}|${p.skill_id || ''}|${p.message}|${(p.options || []).length}` : ''
})
watch(promptKey, () => uiStore.setDecisionDockCollapsed(false))
const collapsed = computed(() => uiStore.decisionDockCollapsed && !interrupted.value)

function collapse() {
  uiStore.setDecisionDockCollapsed(true)
}
function expand() {
  uiStore.setDecisionDockCollapsed(false)
}

// —— 拖拽：拖动顶部把手，位移限制在视口内，松手时记住 ——
const drag = ref<{ startX: number; startY: number; baseX: number; baseY: number } | null>(null)
const dockEl = ref<HTMLElement | null>(null)

function clampOffset(x: number, y: number): { x: number; y: number } {
  const el = dockEl.value
  if (!el || typeof window === 'undefined') return { x, y }
  const rect = el.getBoundingClientRect()
  const cur = uiStore.decisionDockOffset
  // 当前元素在 0 偏移时的位置
  const baseLeft = rect.left - cur.x
  const baseTop = rect.top - cur.y
  const margin = 8
  const minX = margin - baseLeft
  const maxX = window.innerWidth - margin - rect.width - baseLeft
  const minY = margin - baseTop
  const maxY = window.innerHeight - margin - rect.height - baseTop
  return { x: Math.min(maxX, Math.max(minX, x)), y: Math.min(maxY, Math.max(minY, y)) }
}

function onGripDown(event: PointerEvent) {
  if (event.button !== 0) return
  // 标题栏上的按钮（收起）自己处理点击，不触发拖动
  if ((event.target as HTMLElement | null)?.closest('button')) return
  const base = uiStore.decisionDockOffset
  drag.value = { startX: event.clientX, startY: event.clientY, baseX: base.x, baseY: base.y }
  ;(event.currentTarget as HTMLElement).setPointerCapture?.(event.pointerId)
  event.preventDefault()
}
function onGripMove(event: PointerEvent) {
  const d = drag.value
  if (!d) return
  const next = clampOffset(d.baseX + event.clientX - d.startX, d.baseY + event.clientY - d.startY)
  uiStore.decisionDockOffset = next
}
function onGripUp() {
  if (!drag.value) return
  drag.value = null
  uiStore.setDecisionDockOffset({ ...uiStore.decisionDockOffset })
}
function resetPosition() {
  uiStore.setDecisionDockOffset({ x: 0, y: 0 })
}
onBeforeUnmount(() => {
  drag.value = null
})

const dockStyle = computed(() => {
  const { x, y } = uiStore.decisionDockOffset
  return { transform: `translate(calc(-50% + ${x}px), ${y}px)` }
})
</script>

<template>
  <div v-if="active" class="decision-dock-layer" data-testid="decision-dock-layer">
    <Transition name="decision-dock-pop">
      <div
        v-show="!collapsed"
        ref="dockEl"
        class="decision-dock"
        :class="{
          'decision-dock--strip': boardAnswered,
          'decision-dock--interrupt': interrupted,
          'decision-dock--dragging': !!drag,
        }"
        :style="dockStyle"
        role="dialog"
        :aria-label="header?.title || '决策'"
        data-testid="decision-dock"
      >

        <div v-if="header?.band" class="decision-dock__band" data-testid="decision-dock-band">
          <span class="decision-dock__band-tag">打断</span>{{ header.band }}
        </div>

        <header
          v-if="header"
          class="decision-dock__head"
          title="按住标题栏拖动面板（双击回到默认位置）"
          @pointerdown="onGripDown"
          @pointermove="onGripMove"
          @pointerup="onGripUp"
          @pointercancel="onGripUp"
          @dblclick="resetPosition"
        >
          <img v-if="header.avatar" class="decision-dock__avatar" :src="header.avatar" alt="" />
          <div class="decision-dock__titles">
            <div class="decision-dock__who">{{ header.who }}</div>
            <div class="decision-dock__title" data-testid="decision-dock-title">{{ header.title }}</div>
            <div v-if="header.message && !boardAnswered" class="decision-dock__message">{{ header.message }}</div>
          </div>
          <button
            v-if="!interrupted"
            type="button"
            class="decision-dock__tool"
            title="收起面板，查看场上局势"
            aria-label="收起面板"
            data-testid="decision-dock-collapse"
            @pointerdown.stop
            @click="collapse"
          >
            —
          </button>
        </header>

        <div class="decision-dock__body">
          <PromptDialog />
        </div>
      </div>
    </Transition>

    <Transition name="decision-dock-pop">
      <button
        v-if="collapsed"
        type="button"
        class="decision-dock-pill"
        data-testid="decision-dock-pill"
        :aria-label="`展开决策面板：${header?.title || ''}`"
        @click="expand"
      >
        <img v-if="header?.avatar" class="decision-dock-pill__avatar" :src="header.avatar" alt="" />
        <span class="decision-dock-pill__text">
          <span class="decision-dock-pill__title">待决策：{{ header?.title }}</span>
          <span class="decision-dock-pill__sub">点击展开</span>
        </span>
        <span class="decision-dock-pill__open" aria-hidden="true">▲</span>
      </button>
    </Transition>
  </div>
</template>

<style scoped>
.decision-dock-layer {
  position: absolute;
  inset: 0;
  z-index: 40;
  pointer-events: none;
}

/* —— 面板外壳：与行动面板同一材质 —— */
.decision-dock {
  position: absolute;
  left: 50%;
  /* 中央偏下：底边贴住手牌上沿，向上展开；不压手牌，也尽量不压战斗链 */
  bottom: var(--decision-dock-hand-clearance, 196px);
  width: min(520px, calc(100% - 32px));
  max-height: calc(100% - var(--decision-dock-hand-clearance, 196px) - 80px);
  display: flex;
  flex-direction: column;
  border-radius: 10px;
  background:
    radial-gradient(120% 80% at 50% 0%, var(--bt-panel-raised), transparent 70%),
    linear-gradient(180deg, var(--bt-bg-1, #0d1420), var(--bt-bg-0, #070b14));
  box-shadow:
    0 0 0 1px var(--bt-chrome-mid),
    0 0 0 3px rgba(0, 0, 0, 0.55),
    0 18px 40px rgba(0, 0, 0, 0.6);
  pointer-events: auto;
  transition: box-shadow 0.2s ease;
}
.decision-dock::before {
  content: '';
  position: absolute;
  inset: 4px;
  border-radius: 7px;
  border: 1px solid var(--bt-hairline-soft);
  pointer-events: none;
}
.decision-dock--dragging {
  box-shadow:
    0 0 0 1px var(--bt-select),
    0 0 0 3px rgba(0, 0, 0, 0.55),
    0 0 24px var(--bt-select-glow),
    0 24px 50px rgba(0, 0, 0, 0.65);
}
.decision-dock--interrupt {
  box-shadow:
    0 0 0 1px var(--bt-danger),
    0 0 0 3px rgba(0, 0, 0, 0.55),
    0 0 30px rgba(226, 82, 94, 0.35),
    0 18px 40px rgba(0, 0, 0, 0.6);
}

/* 窄条：贴近手牌上方，宽一些、矮一些 */
.decision-dock--strip {
  width: min(680px, calc(100% - 32px));
  bottom: var(--decision-dock-hand-clearance, 196px);
}

.decision-dock__band {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 16px;
  border-radius: 10px 10px 0 0;
  background: linear-gradient(90deg, rgba(192, 48, 60, 0.85), rgba(192, 48, 60, 0));
  font-size: 12px;
  font-weight: 700;
  color: var(--bt-text);
}
.decision-dock__band-tag {
  padding: 0 8px;
  border-radius: 3px;
  background: #fff;
  color: var(--bt-red);
}

.decision-dock__head {
  cursor: grab;
  user-select: none;
  touch-action: none;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 14px 8px;
  border-bottom: 1px solid var(--bt-hairline-soft);
}
.decision-dock--dragging .decision-dock__head {
  cursor: grabbing;
}
.decision-dock--strip .decision-dock__head {
  padding: 10px 14px 8px;
}
.decision-dock__avatar {
  width: 38px;
  height: 38px;
  flex: none;
  border-radius: 50%;
  object-fit: cover;
  object-position: center 20%;
  box-shadow: 0 0 0 2px var(--bt-active), 0 0 10px var(--bt-active-glow);
}
.decision-dock--interrupt .decision-dock__avatar {
  box-shadow: 0 0 0 2px var(--bt-danger), 0 0 10px rgba(226, 82, 94, 0.45);
}
.decision-dock--strip .decision-dock__avatar {
  width: 34px;
  height: 34px;
}
.decision-dock__titles {
  min-width: 0;
  flex: 1;
}
.decision-dock__who {
  font-size: 11px;
  color: var(--bt-text-sub);
}
.decision-dock__title {
  font-size: 18px;
  font-weight: 800;
  letter-spacing: 3px;
  line-height: 1.2;
  color: var(--bt-text);
}
.decision-dock--strip .decision-dock__title {
  font-size: 16px;
  letter-spacing: 2px;
}
.decision-dock__message {
  margin-top: 2px;
  font-size: 12px;
  line-height: 1.4;
  color: var(--bt-text-sub);
}
.decision-dock__tool {
  width: 28px;
  height: 28px;
  flex: none;
  align-self: flex-start;
  border: none;
  border-radius: 6px;
  color: var(--bt-chrome);
  background: var(--bt-bg-0);
  box-shadow: 0 0 0 1px var(--bt-chrome-mid);
  font-size: 15px;
  line-height: 1;
  cursor: pointer;
}
.decision-dock__tool:hover {
  color: var(--bt-chrome-hi);
  box-shadow: 0 0 0 1px var(--bt-chrome);
}

.decision-dock__body {
  min-height: 0;
  overflow-y: auto;
  padding: 10px 14px 12px;
}
.decision-dock--strip .decision-dock__body {
  padding: 8px 12px 10px;
}

/* 面板内的 PromptDialog：去掉它自己的外框，标题由来由条提供 */
.decision-dock__body :deep(.prompt-inline-root) {
  width: 100%;
}
.decision-dock__body :deep(.prompt-inline-surface) {
  width: 100%;
  border: none;
  border-radius: 0;
  background: transparent;
  box-shadow: none;
  padding: 0;
}
.decision-dock__body :deep(.prompt-inline-context),
.decision-dock__body :deep(.overlay-panel-context) {
  display: none;
}

/* —— 收起胶囊 —— */
.decision-dock-pill {
  position: absolute;
  right: 24px;
  bottom: calc(var(--decision-dock-hand-clearance, 196px) - 20px);
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 7px 16px 7px 7px;
  border: none;
  border-radius: 999px;
  color: var(--bt-text);
  background: var(--bt-panel-raised);
  box-shadow: 0 0 0 1px var(--bt-active), 0 0 0 3px rgba(0, 0, 0, 0.55), 0 0 18px var(--bt-active-glow);
  pointer-events: auto;
  cursor: pointer;
  animation: decision-pill-breathe 1.4s ease-in-out infinite;
}
.decision-dock-pill__avatar {
  width: 30px;
  height: 30px;
  border-radius: 50%;
  object-fit: cover;
  object-position: center 20%;
  box-shadow: 0 0 0 2px var(--bt-active);
}
.decision-dock-pill__text {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  line-height: 1.25;
}
.decision-dock-pill__title {
  font-size: 14px;
  font-weight: 800;
}
.decision-dock-pill__sub {
  font-size: 11px;
  color: var(--bt-text-sub);
}
.decision-dock-pill__open {
  margin-left: 4px;
  color: var(--bt-active);
  font-size: 14px;
}
@keyframes decision-pill-breathe {
  50% {
    box-shadow: 0 0 0 1px var(--bt-active), 0 0 0 3px rgba(0, 0, 0, 0.55), 0 0 6px transparent;
  }
}

.decision-dock-pop-enter-active,
.decision-dock-pop-leave-active {
  transition: opacity 0.18s ease;
}
.decision-dock-pop-enter-from,
.decision-dock-pop-leave-to {
  opacity: 0;
}

@media (prefers-reduced-motion: reduce) {
  .decision-dock-pill {
    animation: none;
  }
}
</style>
