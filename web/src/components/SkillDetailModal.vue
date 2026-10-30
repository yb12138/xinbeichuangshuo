<script setup lang="ts">
/**
 * 角色技能详情面板（所有角色通用）：头部（头像 / 名字 / 命格）、资源行（在座时）、技能列表。
 * 技能条目：左侧类型色带 + 类型图标；名字 + 类型 / 标签胶囊；消耗靠右；描述中数字高亮、①② 分支分行；
 * 当前回合我自己不可发动的主动技降亮并显示服务端给出的原因。
 */
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import type { CharacterView, PlayerView, SkillView } from '../types/game'
import { characterArtSrc } from '../constants/characterArt'
import { seatTokenMarks } from '../constants/seatMarks'
import { MARK_GLYPHS } from './marks/markGlyphs'
import BattleMarkIcon from './marks/BattleMarkIcon.vue'
import { useSnapshotStore } from '../stores/snapshot.store'
import { useSessionStore } from '../stores/session.store'

interface SkillModalAnchor {
  x: number
  y: number
  width: number
  height: number
}

const props = defineProps<{
  character: CharacterView | null
  visible: boolean
  anchor?: SkillModalAnchor | null
}>()

const emit = defineEmits<{
  close: []
}>()

const snapshotStore = useSnapshotStore()
const sessionStore = useSessionStore()

const viewport = ref({
  width: typeof window !== 'undefined' ? window.innerWidth : 1280,
  height: typeof window !== 'undefined' ? window.innerHeight : 720,
})

function syncViewport() {
  if (typeof window === 'undefined') return
  viewport.value = {
    width: window.innerWidth,
    height: window.innerHeight,
  }
}

onMounted(() => {
  if (typeof window === 'undefined') return
  window.addEventListener('resize', syncViewport)
})

onBeforeUnmount(() => {
  if (typeof window === 'undefined') return
  window.removeEventListener('resize', syncViewport)
})

const isAnchored = computed(() => !!props.anchor)

/** 对局中扮演该角色的座位（选角大厅里没有）。 */
const seat = computed<PlayerView | null>(() => {
  const id = props.character?.id
  if (!id) return null
  return Object.values(snapshotStore.players).find((p) => p.role === id) ?? null
})

const campClass = computed(() => (seat.value?.camp === 'Red' ? 'skill-card--red' : seat.value?.camp === 'Blue' ? 'skill-card--blue' : ''))
const portrait = computed(() => characterArtSrc(props.character?.id))
const portraitFailed = ref(false)

/** 资源行：治疗、专属指示物、能量。只在对局中显示。 */
const resources = computed(() => {
  const p = seat.value
  if (!p) return null
  return {
    heal: p.heal ?? 0,
    maxHeal: p.max_heal ?? 0,
    tokens: seatTokenMarks(p),
    gem: p.gem ?? 0,
    crystal: p.crystal ?? 0,
  }
})

/** 我自己的角色才有可发动判定（服务端 available_skills）。 */
const isMine = computed(() => !!seat.value && seat.value.id === sessionStore.myPlayerId)

function disabledReason(skill: SkillView): string {
  if (!isMine.value) return ''
  const server = snapshotStore.availableSkills.find((s) => s.id === skill.id)
  return server?.disabled ? server.disabled_reason || '当前不可发动' : ''
}

const TYPE_META: Record<number, { label: string; cls: string; glyph: string }> = {
  0: { label: '被动', cls: 'passive', glyph: 'laurel' },
  1: { label: '启动', cls: 'startup', glyph: 'sun' },
  2: { label: '法术', cls: 'action', glyph: 'prism' },
  3: { label: '响应', cls: 'response', glyph: 'eye' },
}

function typeMeta(type?: number) {
  return TYPE_META[type ?? 2] ?? TYPE_META[2]!
}

const TAG_LABELS: Record<string, string> = {
  Ultimate: '大招',
  TurnLimit: '回合限定',
  Force: '强制',
  Unique: '独有',
  Exclusive: '专属',
}

function skillTags(skill: SkillView): string[] {
  const tags: string[] = []
  for (const tag of skill.tags ?? []) {
    const label = TAG_LABELS[tag]
    if (label && !tags.includes(label)) tags.push(label)
  }
  if (skill.require_exclusive && !tags.includes('独有')) tags.push('独有')
  if (!tags.includes('回合限定') && /［回合限定］|\[回合限定\]/.test(skill.description)) tags.push('回合限定')
  // "仅圣煌形态可发动" → 形态限制胶囊
  const form = skill.description.match(/仅([^，,：:；;。\s]{1,8}?形态)/)
  if (form) tags.push(`仅${form[1]}`)
  return tags
}

const ELEMENT_LABELS: Record<string, string> = {
  Earth: '地', Water: '水', Fire: '火', Wind: '风', Thunder: '雷', Light: '光', Dark: '暗',
}

function discardText(skill: SkillView): string {
  if (!skill.cost_discards) return ''
  const ele = skill.discard_element ? ELEMENT_LABELS[skill.discard_element] || skill.discard_element : ''
  return `弃${skill.cost_discards}${ele ? ` ${ele}系` : ''}`
}

type Segment = { text: string; hi: boolean }
type DescLine = { marker: string; segments: Segment[] }

/** 数字（含 +1 / -1）高亮；①② 等分支各占一行。 */
function highlight(text: string): Segment[] {
  return text
    .split(/([+\-＋－]?\d+)/)
    .filter(Boolean)
    .map((part) => ({ text: part, hi: /^[+\-＋－]?\d+$/.test(part) }))
}

function descLines(description: string): DescLine[] {
  const parts = String(description || '').split(/(?=[①②③④⑤])/)
  return parts
    .map((part) => {
      const m = part.match(/^([①②③④⑤])\s*/)
      const marker = m ? m[1]! : ''
      const body = (m ? part.slice(m[0].length) : part).trim()
      return { marker, segments: highlight(marker ? body.replace(/[；;]$/, '') : body) }
    })
    .filter((line) => line.segments.length)
}

const skillRows = computed(() =>
  (props.character?.skills ?? []).map((skill) => {
    const meta = typeMeta(skill.type)
    return {
      skill,
      meta,
      glyph: MARK_GLYPHS[meta.glyph] || '',
      tags: skillTags(skill),
      discard: discardText(skill),
      lines: descLines(skill.description),
      reason: disabledReason(skill),
    }
  }),
)

const panelStyle = computed(() => {
  if (!props.anchor) return {}
  const margin = 12
  const gap = 8
  const preferredWidth = 440
  const minWidth = 300
  const estimatedHeight = 560
  const width = Math.min(preferredWidth, Math.max(minWidth, viewport.value.width - margin * 2))

  let left = props.anchor.x - width - gap
  if (left < margin) {
    left = props.anchor.x + props.anchor.width + gap
  }
  if (left + width > viewport.value.width - margin) {
    left = viewport.value.width - width - margin
  }
  if (left < margin) left = margin

  let top = props.anchor.y - 6
  const maxTop = viewport.value.height - estimatedHeight - margin
  if (top > maxTop) {
    top = Math.max(margin, maxTop)
  }
  if (top < margin) top = margin

  return {
    width: `${Math.round(width)}px`,
    left: `${Math.round(left)}px`,
    top: `${Math.round(top)}px`,
    // 面板比预估高时不越出视口底部，列表区自己滚动
    maxHeight: `${Math.round(Math.min(720, viewport.value.height - top - margin))}px`,
  }
})
</script>

<template>
  <Teleport to="body">
    <Transition name="modal">
      <div
        v-if="visible && character"
        class="skill-overlay fixed inset-0 z-[13050]"
        :class="isAnchored ? 'skill-overlay--anchored' : 'skill-overlay--centered'"
        @click.self="!isAnchored && emit('close')"
      >
        <div
          class="skill-card"
          :class="[isAnchored ? 'skill-card--anchored' : 'skill-card--centered', campClass]"
          :style="isAnchored ? panelStyle : undefined"
          role="dialog"
          :aria-label="`${character.name} 技能详情`"
          data-testid="skill-detail"
          @click.stop
        >
          <i class="skill-corner skill-corner--tl" aria-hidden="true" />
          <i class="skill-corner skill-corner--tr" aria-hidden="true" />
          <i class="skill-corner skill-corner--bl" aria-hidden="true" />
          <i class="skill-corner skill-corner--br" aria-hidden="true" />

          <!-- 头部 -->
          <header class="skill-head">
            <span class="skill-head__avatar">
              <img v-if="portrait && !portraitFailed" :src="portrait" alt="" @error="portraitFailed = true" />
              <span v-else class="skill-head__avatar-fallback">{{ character.name.slice(0, 1) }}</span>
            </span>
            <div class="skill-head__titles">
              <h2 class="skill-head__name">{{ character.name }}</h2>
              <div v-if="character.faction || character.title" class="skill-head__sub">命格：{{ character.faction || character.title }}</div>
            </div>
            <button type="button" class="skill-head__close" aria-label="关闭" @click="emit('close')">
              <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M6 6l12 12M18 6L6 18" /></svg>
            </button>
          </header>

          <!-- 资源行（对局中） -->
          <div v-if="resources" class="skill-res" data-testid="skill-detail-resources">
            <span class="skill-res__item" title="治疗">
              <svg class="skill-res__ico skill-res__ico--heal" viewBox="0 0 16 16" aria-hidden="true"><circle cx="8" cy="8" r="7" /><path d="M8 4.2v7.6M4.2 8h7.6" /></svg>
              治疗<b>{{ resources.heal }}/{{ resources.maxHeal }}</b>
            </span>
            <span v-for="token in resources.tokens" :key="token.key" class="skill-res__item" :title="token.hint || token.label">
              <span v-if="token.mark" class="skill-res__mark"><BattleMarkIcon kind="token" :mark-key="token.key" /></span>
              {{ token.label }}<b>{{ token.cap ? `${token.value}/${token.cap}` : token.value }}</b>
            </span>
            <span class="skill-res__item skill-res__item--stone" title="红宝石">
              <svg class="skill-res__ico" viewBox="0 0 16 16" aria-hidden="true"><polygon class="stone-gem" points="8,1 13.4,5.6 8,15 2.6,5.6" /></svg>
              <b>{{ resources.gem }}</b>
            </span>
            <span class="skill-res__item skill-res__item--stone" title="蓝水晶">
              <svg class="skill-res__ico" viewBox="0 0 16 16" aria-hidden="true"><polygon class="stone-crystal" points="8,0.8 12.6,4.6 12.6,10.8 8,15.2 3.4,10.8 3.4,4.6" /></svg>
              <b>{{ resources.crystal }}</b>
            </span>
          </div>

          <!-- 技能列表 -->
          <div class="skill-list">
            <article
              v-for="row in skillRows"
              :key="row.skill.id"
              class="skill-row"
              :class="[`skill-row--${row.meta.cls}`, { 'skill-row--disabled': !!row.reason }]"
              :data-testid="`skill-detail-${row.skill.id}`"
            >
              <span class="skill-row__icon" aria-hidden="true">
                <svg viewBox="0 0 24 24"><path :d="row.glyph" /></svg>
              </span>
              <div class="skill-row__main">
                <div class="skill-row__head">
                  <span class="skill-row__name">{{ row.skill.title }}</span>
                  <span class="skill-pill skill-pill--type">{{ row.meta.label }}</span>
                  <span v-for="tag in row.tags" :key="tag" class="skill-pill skill-pill--tag">{{ tag }}</span>
                  <span v-if="row.skill.cost_gem || row.skill.cost_crystal || row.discard" class="skill-row__cost">
                    <span v-if="row.skill.cost_gem" class="skill-cost" :title="`消耗${row.skill.cost_gem}红宝石`">
                      <svg viewBox="0 0 16 16" aria-hidden="true"><polygon class="stone-gem" points="8,1 13.4,5.6 8,15 2.6,5.6" /></svg>{{ row.skill.cost_gem }}
                    </span>
                    <span v-if="row.skill.cost_crystal" class="skill-cost" :title="`消耗${row.skill.cost_crystal}蓝水晶`">
                      <svg viewBox="0 0 16 16" aria-hidden="true"><polygon class="stone-crystal" points="8,0.8 12.6,4.6 12.6,10.8 8,15.2 3.4,10.8 3.4,4.6" /></svg>{{ row.skill.cost_crystal }}
                    </span>
                    <span v-if="row.discard" class="skill-cost skill-cost--discard">{{ row.discard }}</span>
                  </span>
                </div>
                <p
                  v-for="(line, li) in row.lines"
                  :key="li"
                  class="skill-row__line"
                  :class="{ 'skill-row__line--branch': !!line.marker }"
                >
                  <span v-if="line.marker" class="skill-row__marker">{{ line.marker }}</span>
                  <span><template v-for="(seg, si) in line.segments" :key="si"><em v-if="seg.hi">{{ seg.text }}</em><template v-else>{{ seg.text }}</template></template></span>
                </p>
                <p v-if="row.reason" class="skill-row__reason" :data-testid="`skill-detail-reason-${row.skill.id}`">不可发动：{{ row.reason }}</p>
              </div>
            </article>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.skill-overlay {
  z-index: 13050 !important;
}

.skill-overlay--centered {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1rem;
  background: rgba(0, 0, 0, 0.72);
  backdrop-filter: blur(3px);
}

.skill-overlay--anchored {
  pointer-events: none;
  padding: 0;
}

/* —— 面板外框：黑漆底 + 金色细边 + 四角装饰 —— */
.skill-card {
  --skill-gold: #c9a45c;
  --skill-gold-hi: #ffd98a;
  --skill-camp: rgba(59, 111, 216, 0.38);
  position: relative;
  z-index: 13051 !important;
  display: flex;
  flex-direction: column;
  max-height: 85vh;
  overflow: hidden;
  border-radius: 6px;
  background:
    radial-gradient(120% 60% at 50% 0%, rgba(255, 220, 150, 0.05), transparent 60%),
    linear-gradient(180deg, rgba(13, 18, 27, 0.96), rgba(9, 13, 20, 0.98));
  box-shadow:
    0 0 0 1px rgba(201, 164, 92, 0.55),
    0 0 0 3px rgba(0, 0, 0, 0.6),
    inset 0 1px 0 rgba(255, 240, 200, 0.08),
    0 18px 40px rgba(0, 0, 0, 0.6);
  color: var(--bt-text, #e8ecf4);
}
.skill-card--red {
  --skill-camp: rgba(192, 48, 60, 0.42);
}
.skill-card--centered {
  width: min(480px, calc(100vw - 2rem));
}
.skill-card--anchored {
  position: fixed;
  pointer-events: auto;
  max-height: min(85vh, 720px);
}

.skill-corner {
  position: absolute;
  z-index: 2;
  width: 14px;
  height: 14px;
  pointer-events: none;
  border-color: var(--skill-gold);
  border-style: solid;
  border-width: 0;
}
.skill-corner::after {
  content: '';
  position: absolute;
  width: 4px;
  height: 4px;
  background: var(--skill-gold-hi);
  transform: rotate(45deg);
}
.skill-corner--tl { top: 4px; left: 4px; border-top-width: 2px; border-left-width: 2px; }
.skill-corner--tr { top: 4px; right: 4px; border-top-width: 2px; border-right-width: 2px; }
.skill-corner--bl { bottom: 4px; left: 4px; border-bottom-width: 2px; border-left-width: 2px; }
.skill-corner--br { bottom: 4px; right: 4px; border-bottom-width: 2px; border-right-width: 2px; }
.skill-corner--tl::after { top: -3px; left: -3px; }
.skill-corner--tr::after { top: -3px; right: -3px; }
.skill-corner--bl::after { bottom: -3px; left: -3px; }
.skill-corner--br::after { bottom: -3px; right: -3px; }

/* —— 头部 —— */
.skill-head {
  flex: none;
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 16px 16px 14px 18px;
  background:
    linear-gradient(110deg, var(--skill-camp), rgba(80, 60, 30, 0.22) 70%, transparent),
    rgba(10, 14, 22, 0.6);
  border-bottom: 1px solid rgba(201, 164, 92, 0.28);
}
.skill-head__avatar {
  flex: none;
  width: 64px;
  height: 64px;
  border-radius: 50%;
  overflow: hidden;
  box-shadow:
    0 0 0 2px var(--skill-gold),
    0 0 0 4px rgba(0, 0, 0, 0.55),
    0 0 14px rgba(255, 217, 138, 0.35);
  background: var(--bt-bg-1, #111722);
}
.skill-head__avatar img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  object-position: center 18%;
}
.skill-head__avatar-fallback {
  display: flex;
  width: 100%;
  height: 100%;
  align-items: center;
  justify-content: center;
  font-size: 26px;
  font-weight: 700;
  color: var(--skill-gold-hi);
}
.skill-head__titles {
  flex: 1;
  min-width: 0;
}
.skill-head__name {
  margin: 0;
  font-size: 24px;
  font-weight: 800;
  letter-spacing: 4px;
  line-height: 1.2;
  color: #f4ecd8;
  text-shadow: 0 1px 0 rgba(0, 0, 0, 0.6);
}
.skill-head__sub {
  margin-top: 4px;
  font-size: 13px;
  color: var(--bt-text-sub, #9aa4b5);
}
.skill-head__close {
  flex: none;
  align-self: flex-start;
  width: 30px;
  height: 30px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: none;
  border-radius: 4px;
  background: transparent;
  color: var(--bt-chrome, #b9c1cd);
  cursor: pointer;
}
.skill-head__close:hover,
.skill-head__close:focus-visible {
  color: var(--skill-gold-hi);
  background: rgba(255, 255, 255, 0.06);
}
.skill-head__close svg {
  width: 18px;
  height: 18px;
  fill: none;
  stroke: currentColor;
  stroke-width: 2;
  stroke-linecap: round;
}

/* —— 资源行 —— */
.skill-res {
  flex: none;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 6px 16px;
  margin: 10px 14px 4px;
  padding: 7px 12px;
  border-radius: 4px;
  background: rgba(0, 0, 0, 0.32);
  box-shadow: inset 0 0 0 1px rgba(201, 164, 92, 0.3);
  font-size: 13px;
  color: var(--bt-text-sub, #9aa4b5);
}
.skill-res__item {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  white-space: nowrap;
}
.skill-res__item b {
  margin-left: 2px;
  font-weight: 600;
  color: var(--bt-text, #e8ecf4);
}
.skill-res__ico,
.skill-res__mark {
  width: 16px;
  height: 16px;
  flex: none;
}
.skill-res__ico--heal circle {
  fill: #2f9e5b;
}
.skill-res__ico--heal path {
  stroke: #eafff0;
  stroke-width: 2;
  stroke-linecap: round;
}

.stone-gem {
  fill: var(--bt-red, #c0303c);
  stroke: var(--bt-chrome-hi, #e9edf3);
  stroke-opacity: 0.7;
  stroke-width: 0.8;
  stroke-linejoin: round;
}
.stone-crystal {
  fill: var(--bt-blue, #3b6fd8);
  stroke: var(--bt-chrome-hi, #e9edf3);
  stroke-opacity: 0.7;
  stroke-width: 0.8;
  stroke-linejoin: round;
}

/* —— 技能列表 —— */
.skill-list {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  padding: 6px 10px 14px 14px;
  scrollbar-width: thin;
  scrollbar-color: rgba(201, 164, 92, 0.55) transparent;
}
.skill-list::-webkit-scrollbar {
  width: 4px;
}
.skill-list::-webkit-scrollbar-thumb {
  border-radius: 2px;
  background: rgba(201, 164, 92, 0.55);
}

.skill-row {
  --row-accent: #c9a45c;
  position: relative;
  display: flex;
  gap: 12px;
  padding: 10px 6px 10px 14px;
  border-bottom: 1px solid var(--bt-hairline-soft, rgba(185, 193, 205, 0.16));
}
.skill-row::before {
  content: '';
  position: absolute;
  left: 0;
  top: 8px;
  bottom: 8px;
  width: 3px;
  border-radius: 2px;
  background: var(--row-accent);
  box-shadow: 0 0 6px var(--row-accent);
  opacity: 0.85;
}
.skill-row:last-child {
  border-bottom: none;
}
.skill-row--passive { --row-accent: #9aa4b5; }
.skill-row--startup { --row-accent: #3fb6c8; }
.skill-row--action { --row-accent: #e0a64a; }
.skill-row--response { --row-accent: #9b7be0; }

.skill-row__icon {
  flex: none;
  width: 44px;
  height: 44px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 6px;
  background:
    radial-gradient(circle at 50% 35%, color-mix(in srgb, var(--row-accent) 35%, transparent), transparent 70%),
    linear-gradient(180deg, #1a2231, #0d121b);
  box-shadow:
    0 0 0 1px color-mix(in srgb, var(--row-accent) 70%, transparent),
    inset 0 1px 0 rgba(255, 255, 255, 0.08);
}
.skill-row__icon svg {
  width: 26px;
  height: 26px;
  fill: color-mix(in srgb, var(--row-accent) 65%, #fff);
  filter: drop-shadow(0 0 4px color-mix(in srgb, var(--row-accent) 60%, transparent));
}

.skill-row__main {
  flex: 1;
  min-width: 0;
}
.skill-row__head {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 6px;
  margin-bottom: 4px;
}
.skill-row__name {
  margin-right: 2px;
  font-size: 16px;
  font-weight: 700;
  letter-spacing: 1px;
  color: var(--skill-gold-hi, #ffd98a);
}
.skill-pill {
  display: inline-flex;
  align-items: center;
  padding: 0 8px;
  height: 18px;
  border-radius: 999px;
  font-size: 11px;
  line-height: 1;
  white-space: nowrap;
}
.skill-pill--type {
  color: color-mix(in srgb, var(--row-accent) 55%, #fff);
  background: color-mix(in srgb, var(--row-accent) 16%, transparent);
  box-shadow: inset 0 0 0 1px color-mix(in srgb, var(--row-accent) 60%, transparent);
}
.skill-pill--tag {
  color: #f0d9a8;
  background: rgba(120, 84, 30, 0.35);
  box-shadow: inset 0 0 0 1px rgba(201, 164, 92, 0.6);
}
.skill-row__cost {
  margin-left: auto;
  display: inline-flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: var(--bt-text, #e8ecf4);
}
.skill-cost {
  display: inline-flex;
  align-items: center;
  gap: 3px;
}
.skill-cost svg {
  width: 14px;
  height: 14px;
}
.skill-cost--discard {
  padding: 0 6px;
  border-radius: 3px;
  font-size: 12px;
  box-shadow: inset 0 0 0 1px var(--bt-hairline, rgba(185, 193, 205, 0.4));
}

.skill-row__line {
  margin: 0;
  font-size: 13px;
  line-height: 1.6;
  color: #cfd5df;
  word-break: break-word;
}
.skill-row__line--branch {
  display: flex;
  gap: 6px;
}
.skill-row__marker {
  flex: none;
  color: var(--skill-gold, #c9a45c);
}
.skill-row__line em {
  font-style: normal;
  font-weight: 600;
  color: var(--skill-gold-hi, #ffd98a);
}
.skill-row__reason {
  margin: 2px 0 0;
  font-size: 12px;
  line-height: 1.5;
  color: #f07c72;
}
.skill-row--disabled .skill-row__icon,
.skill-row--disabled .skill-row__name,
.skill-row--disabled .skill-row__line {
  opacity: 0.55;
}
.skill-row--disabled .skill-row__icon {
  filter: grayscale(0.6);
}

.modal-enter-active,
.modal-leave-active {
  transition: opacity 0.24s ease;
}
.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}
.modal-enter-active .skill-card,
.modal-leave-active .skill-card {
  transition: transform 0.24s ease;
}
.modal-enter-from .skill-card,
.modal-leave-to .skill-card {
  transform: scale(0.95) translateY(8px);
}

@media (prefers-reduced-motion: reduce) {
  .modal-enter-active,
  .modal-leave-active,
  .modal-enter-active .skill-card,
  .modal-leave-active .skill-card {
    transition: none;
  }
}
</style>
