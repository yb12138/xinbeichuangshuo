<script setup lang="ts">
/**
 * 扩展区里的盖牌。
 * - 自己可见（faceUp）：直接用正常卡面，左上角挂种类角标（茧/暗月/剑魂…）。
 * - 他人视角（牌面已被服务端遮蔽）：整张画种类牌背，颜色/剪影来自 constants/expansionZone。
 */
import { computed } from 'vue'
import type { Card } from '../../types/game'
import { expansionGroupView } from '../../constants/expansionZone'
import CardComponent from '../CardComponent.vue'

const props = defineProps<{
  card: Card
  /** FieldCard.effect，决定牌背纹样。 */
  effect?: string
  index: number
  testId?: string
  selectable?: boolean
  selected?: boolean
  /** 当前处于扩展区选牌中：不可选的牌压暗。 */
  choosing?: boolean
  /** 牌面可见（盖牌主人自己）；为 false 时画牌背。 */
  faceUp?: boolean
}>()

const emit = defineEmits<{
  click: [index: number]
}>()

const view = computed(() => {
  const v = expansionGroupView(props.effect || '')
  return { label: v.label, accent: v.tint, glyph: v.glyph }
})

function handleClick() {
  if (props.selectable) emit('click', props.index)
}
</script>

<template>
  <!-- 自己的盖牌：正常卡面 + 种类角标 -->
  <div
    v-if="faceUp"
    class="cover-face"
    :class="{
      'cover-face--choosing': choosing,
      'cover-face--selected': selected
    }"
    :style="{ '--cover-accent': view.accent }"
  >
    <CardComponent
      :card="card"
      :index="index"
      :test-id="testId"
      battle-mini
      :selectable="selectable"
      :selected="selected"
      @click="handleClick"
    />
    <span class="cover-face__badge" :title="`${view.label}（盖牌，仅自己可见）`">
      <svg viewBox="0 0 24 24" aria-hidden="true"><path :d="view.glyph" /></svg>
      {{ view.label }}
    </span>
    <span v-if="selected" class="cover-card__check" aria-hidden="true">✓</span>
  </div>

  <!-- 他人视角：种类牌背 -->
  <button
    v-else
    type="button"
    class="cover-card"
    :class="{
      'cover-card--selectable': selectable,
      'cover-card--selected': selected,
      'cover-card--dimmed': choosing && !selectable && !selected
    }"
    :style="{ '--cover-accent': view.accent }"
    :data-testid="testId"
    :aria-pressed="selectable ? !!selected : undefined"
    :aria-label="`${view.label}（盖牌，牌面不可见）`"
    :tabindex="selectable ? 0 : -1"
    @click="handleClick"
  >
    <span class="cover-card__back" aria-hidden="true">
      <svg class="cover-card__pattern" viewBox="0 0 60 84" preserveAspectRatio="none">
        <!-- 外框 + 内框双线 -->
        <rect class="cover-card__frame" x="2.5" y="2.5" width="55" height="79" rx="4" />
        <rect class="cover-card__inner" x="5.5" y="5.5" width="49" height="73" rx="2.5" />
        <!-- 四角卷草 -->
        <path
          class="cover-card__corner"
          d="M6 14c0-5 3-8 8-8M9 11c1.5 0 2.5-1 2.5-2.5M54 14c0-5-3-8-8-8M51 11c-1.5 0-2.5-1-2.5-2.5M6 70c0 5 3 8 8 8M9 73c1.5 0 2.5 1 2.5 2.5M54 70c0 5-3 8-8 8M51 73c-1.5 0-2.5 1-2.5 2.5"
        />
        <!-- 上下中线饰 -->
        <path class="cover-card__corner" d="M24 6.5l6 3.5 6-3.5M24 77.5l6-3.5 6 3.5M30 10v6M30 68v6" />
        <!-- 中央椭圆徽章 -->
        <ellipse class="cover-card__halo" cx="30" cy="40" rx="20" ry="25" />
        <ellipse class="cover-card__ring" cx="30" cy="40" rx="16" ry="20.5" />
        <ellipse class="cover-card__ring cover-card__ring--soft" cx="30" cy="40" rx="18.5" ry="23" />
        <path class="cover-card__corner" d="M14 40h-4M46 40h4M30 19.5v-2M30 60.5v2" />
        <g transform="translate(19.5 29.5) scale(0.875)">
          <path class="cover-card__glyph" :d="view.glyph" />
        </g>
      </svg>
      <span class="cover-card__type">{{ view.label }}</span>
    </span>
    <span v-if="selected" class="cover-card__check" aria-hidden="true">✓</span>
  </button>
</template>

<style scoped>
.cover-card {
  --cover-accent: #b9c1cd;
  position: relative;
  display: flex;
  flex-direction: column;
  width: 74px;
  height: 111px;
  flex: none;
  padding: 0;
  border: 0;
  border-radius: 6px;
  background: #050a14;
  box-shadow:
    0 0 0 1px color-mix(in srgb, var(--cover-accent) 65%, transparent),
    0 0 0 2px rgba(0, 0, 0, 0.75),
    0 0 10px color-mix(in srgb, var(--cover-accent) 18%, transparent),
    0 6px 14px rgba(0, 0, 0, 0.6);
  cursor: default;
  overflow: visible;
  transition: transform 0.18s ease, box-shadow 0.18s ease, opacity 0.18s ease, filter 0.18s ease;
}

.cover-card__back {
  position: relative;
  flex: 1 1 auto;
  min-height: 0;
  border-radius: 6px;
  overflow: hidden;
  /* 深蓝漆面 + 色调光晕 + 细网纹 + 顶部高光 */
  background:
    linear-gradient(165deg, rgba(255, 255, 255, 0.1), transparent 32%),
    radial-gradient(70% 55% at 50% 46%, color-mix(in srgb, var(--cover-accent) 34%, transparent), transparent 72%),
    repeating-linear-gradient(60deg, color-mix(in srgb, var(--cover-accent) 6%, transparent) 0 1px, transparent 1px 6px),
    repeating-linear-gradient(-60deg, color-mix(in srgb, var(--cover-accent) 6%, transparent) 0 1px, transparent 1px 6px),
    linear-gradient(180deg, color-mix(in srgb, var(--cover-accent) 14%, #0b1630), #050a16);
}

.cover-card__pattern {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
}

.cover-card__frame {
  fill: none;
  stroke: color-mix(in srgb, var(--cover-accent) 75%, transparent);
  stroke-width: 1;
}

.cover-card__inner {
  fill: none;
  stroke: color-mix(in srgb, var(--cover-accent) 30%, transparent);
  stroke-width: 0.6;
}

.cover-card__halo {
  fill: color-mix(in srgb, var(--cover-accent) 8%, transparent);
  stroke: none;
}

.cover-card__ring {
  fill: color-mix(in srgb, var(--cover-accent) 16%, rgba(4, 8, 20, 0.6));
  stroke: color-mix(in srgb, var(--cover-accent) 80%, #ffffff);
  stroke-width: 0.9;
}

.cover-card__ring--soft {
  fill: none;
  stroke: color-mix(in srgb, var(--cover-accent) 25%, transparent);
  stroke-dasharray: 2 2.5;
}

.cover-card__corner {
  fill: none;
  stroke: color-mix(in srgb, var(--cover-accent) 70%, #ffffff);
  stroke-width: 0.8;
  stroke-linecap: round;
  opacity: 0.85;
}

.cover-card__glyph {
  fill: color-mix(in srgb, var(--cover-accent) 45%, #ffffff);
  filter: drop-shadow(0 0 2.5px var(--cover-accent));
}

.cover-card__type {
  position: absolute;
  left: 0;
  right: 0;
  bottom: 6px;
  text-align: center;
  color: var(--cover-accent);
  font-size: 10px;
  font-weight: 700;
  letter-spacing: 0.2em;
  padding-left: 0.2em;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.9);
}

/* 可选：抬起 + 选择色发光 */
.cover-card--selectable {
  cursor: pointer;
  transform: translateY(-6px);
  box-shadow:
    0 0 0 1px var(--bt-select),
    0 0 0 2px rgba(0, 0, 0, 0.7),
    0 0 14px var(--bt-select-glow),
    0 10px 18px rgba(0, 0, 0, 0.6);
}

.cover-card--selectable:hover {
  transform: translateY(-9px);
}

.cover-card--selectable:focus-visible {
  outline: 2px solid var(--bt-select);
  outline-offset: 2px;
}

.cover-card--selected,
.cover-card--selected:hover {
  transform: translateY(-10px);
  box-shadow:
    0 0 0 2px var(--bt-select),
    0 0 0 3px rgba(0, 0, 0, 0.7),
    0 0 20px 2px var(--bt-select-glow),
    0 14px 22px rgba(0, 0, 0, 0.65);
}

.cover-card--dimmed {
  opacity: 0.45;
  filter: saturate(0.6) brightness(0.8);
}

.cover-card__check {
  position: absolute;
  top: -8px;
  right: -6px;
  z-index: 2;
  display: grid;
  place-items: center;
  width: 20px;
  height: 20px;
  border-radius: 999px;
  background: var(--bt-select);
  color: var(--bt-bg-0);
  font-size: 12px;
  font-weight: 800;
  box-shadow: 0 0 0 2px var(--bt-bg-0), 0 0 8px var(--bt-select-glow);
  pointer-events: none;
}

/* —— 正面盖牌 —— */
.cover-face {
  --cover-accent: #b9c1cd;
  position: relative;
  flex: none;
}

/* 非选牌时卡面保持原亮度（CardComponent 默认把不可选的牌压暗） */
.cover-face:not(.cover-face--choosing) :deep(.card-shell:not(.selectable):not(.selected)) {
  opacity: 1;
  filter: none;
}

/* 选牌中：不可选的压暗更明显 */
.cover-face--choosing :deep(.card-shell:not(.selectable):not(.selected)) {
  opacity: 0.4;
  filter: saturate(0.5) brightness(0.8);
}

.cover-face__badge {
  position: absolute;
  left: 50%;
  bottom: -7px;
  z-index: 3;
  display: inline-flex;
  align-items: center;
  gap: 3px;
  transform: translateX(-50%);
  padding: 1px 7px 1px 5px;
  border-radius: 999px;
  background: linear-gradient(180deg, color-mix(in srgb, var(--cover-accent) 32%, #0a1324), #050a14);
  box-shadow:
    0 0 0 1px color-mix(in srgb, var(--cover-accent) 75%, transparent),
    0 0 8px color-mix(in srgb, var(--cover-accent) 35%, transparent);
  color: color-mix(in srgb, var(--cover-accent) 35%, #ffffff);
  font-size: 9px;
  font-weight: 700;
  letter-spacing: 0.08em;
  white-space: nowrap;
  pointer-events: none;
  transition: transform 0.18s ease;
}

.cover-face__badge svg {
  width: 10px;
  height: 10px;
  fill: currentColor;
}

/* 跟随卡面抬起 */
.cover-face--selected .cover-face__badge {
  transform: translate(-50%, -12px);
}

.cover-face .cover-card__check {
  top: -18px;
  right: -6px;
}

@media (prefers-reduced-motion: reduce) {
  .cover-card,
  .cover-face__badge {
    transition: none;
  }
}
</style>
