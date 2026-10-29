<script setup lang="ts">
import { computed, ref, useAttrs, watch } from 'vue'
import type { Card } from '../types/game'
import ElementGlyph from './card/ElementGlyph.vue'
import { cardArtSrc as resolveCardArtSrc } from '../constants/cardArt'


defineOptions({
  inheritAttrs: false
})

const props = defineProps<{
  card: Card
  index?: number
  testId?: string
  selected?: boolean
  selectable?: boolean
  faceDown?: boolean
  small?: boolean
  medium?: boolean
  battleMini?: boolean
  effectiveElement?: string
  effectiveElementHint?: string
}>()

const emit = defineEmits<{
  click: [index: number]
}>()

const ELEMENT_LABEL_MAP: Record<string, string> = {
  Water: '水',
  Fire: '火',
  Earth: '地',
  Wind: '风',
  Thunder: '雷',
  Light: '光',
  Dark: '暗'
}

const imageLoadFailed = ref(false)

watch(() => props.card.id, () => {
  imageLoadFailed.value = false
})

const elementClass = computed(() => {
  if (props.faceDown) return 'card-face-down'
  return `element-${props.card.element.toLowerCase()}`
})

const elementLabel = computed(() => {
  return ELEMENT_LABEL_MAP[props.card.element] ?? props.card.element
})

const effectiveElementLabel = computed(() => {
  if (!props.effectiveElement || props.effectiveElement === props.card.element) return ''
  return ELEMENT_LABEL_MAP[props.effectiveElement] ?? props.effectiveElement
})

const hasEffectiveElementHint = computed(() => !!effectiveElementLabel.value)

// 徽章显示"实际生效"的元素；与印刷元素不同时另加小标记
const badgeElement = computed(() => (hasEffectiveElementHint.value ? props.effectiveElement! : props.card.element))

const badgeElementClass = computed(() => `badge-${badgeElement.value.toLowerCase()}`)

const badgeTitle = computed(() => {
  const label = ELEMENT_LABEL_MAP[badgeElement.value] ?? badgeElement.value
  if (!hasEffectiveElementHint.value) return `${label}系`
  return props.effectiveElementHint || `印刷为${elementLabel.value}系，视为${label}系`
})

const cardTypeLabel = computed(() => (props.card.type === 'Attack' ? '攻击' : '法术'))

const detailRibbonText = computed(() => {
  const base = `${elementLabel.value}系${props.card.type === 'Attack' ? '攻击' : '法术'}`
  if (effectiveElementLabel.value) {
    const transformed = `${base} -> ${effectiveElementLabel.value}系`
    if (!fateText.value) return transformed
    return `${transformed}  ${fateText.value}`
  }
  if (!fateText.value) return base
  return `${base}  ${fateText.value}`
})

const fateText = computed(() => {
  if (props.card.faction && props.card.faction.trim()) return props.card.faction
  const m = props.card.description?.match(/([幻咏血技圣])命格/)
  if (m?.[1]) return m[1]
  return ''
})

const exclusiveSkillText = computed(() => {
  const names: string[] = []
  if (props.card.exclusive_skill1) names.push(props.card.exclusive_skill1)
  if (props.card.exclusive_skill2) names.push(props.card.exclusive_skill2)
  return names.join(' / ')
})

const exclusiveSkills = computed(() =>
  [props.card.exclusive_skill1, props.card.exclusive_skill2].filter((name): name is string => !!name)
)

const displayDescription = computed(() => {
  if (props.card.type === 'Attack') {
    return '主动攻击或应战其他攻击时打出，命中时造成两点攻击伤害。'
  }
  if (props.card.name === '圣盾') {
    return '（将此牌放置于目标角色前，他遭受攻击或【魔弹】时，移除此牌）视为未命中'
  }
  if (props.card.name === '魔弹') {
    return '（使用此牌）你右手边最近的一名对手选择以下一项发动：\n受到2点法术伤害③\n（打出1张【魔弹】【展示】）视为由你使用1张【魔弹】继续结算,且造成的法术伤害③额外+1\n（使用1张【圣光】【展示】或移除面前的【圣盾】）抵挡此次伤害'
  }
  if (props.card.name === '虚弱') {
    return '（将此牌放置于目标角色前，他的行动阶段开始前）他选择以下一项发动：\n1.跳过他的行动阶段\n2.（摸3张牌【强制】）继续他的行动阶段'
  }
  if (props.card.name === '中毒') {
    return '（将此牌放置于目标角色前，他的行动阶段开始前）对他造成1点法术伤害③'
  }
  if (props.card.name === '圣光') {
    return '抵挡一次攻击或【魔弹】。'
  }
  return props.card.description
})

const cardArtSrc = computed(() => resolveCardArtSrc(props.card, imageLoadFailed.value))
const showDescModal = ref(false)
const attrs = useAttrs()

function handleImageError() {
  imageLoadFailed.value = true
}

function openDescriptionModal() {
  showDescModal.value = true
}

function closeDescriptionModal() {
  showDescModal.value = false
}

function handleClick() {
  if (props.selectable && props.index !== undefined) {
    emit('click', props.index)
  }
}
</script>

<template>
  <div
    v-bind="attrs"
    class="game-card card-shell relative overflow-hidden flex-shrink-0"
    :data-testid="testId || (index !== undefined ? `hand-card-${index}` : undefined)"
    :class="[
      elementClass,
      battleMini
        ? 'w-[74px] h-[111px] text-xs card-size-battle-mini card-battle-mini'
        : small
        ? 'w-[92px] h-[138px] text-sm card-size-small'
        : medium
          ? 'w-[112px] h-[168px] text-sm card-size-medium'
          : 'w-[132px] h-[198px] text-base card-size-default',
      selected ? 'selected' : '',
      selectable ? 'cursor-pointer selectable' : 'cursor-default',
      faceDown ? 'items-center justify-center' : ''
    ]"
    @click="handleClick"
  >
    <template v-if="faceDown">
      <div class="layer card-back-art"></div>
      <div class="layer card-back-fx"></div>
      <div class="layer card-back-symbol-wrap">
        <div class="card-back-symbol">✦</div>
        <div class="card-back-text">星杯</div>
      </div>
    </template>

    <template v-else>
      <!-- 满版插画：统一 cover 裁切 + 暗角，让横版/方图观感一致 -->
      <div class="layer card-art-frame">
        <img
          class="card-art-image"
          :src="cardArtSrc"
          :alt="card.name"
          loading="lazy"
          decoding="async"
          @error="handleImageError"
        >
        <div class="card-art-overlay"></div>
      </div>
      <div class="layer card-vignette"></div>

      <div class="layer card-title-banner" :class="{ 'has-fate': fateText }">
        <span class="card-title-text">{{ card.name }}</span>
      </div>

      <!-- 命格：右上角印章，与名称条同高，一眼可辨 -->
      <div v-if="fateText" class="layer card-fate-seal" :title="`${fateText}命格`" :aria-label="`${fateText}命格`" data-testid="card-fate-seal">
        {{ fateText }}
      </div>

      <div
        class="layer card-element-medal"
        :class="[badgeElementClass, hasEffectiveElementHint ? 'is-transformed' : '']"
        :title="badgeTitle"
      >
        <span>
          <ElementGlyph :element="badgeElement" />
        </span>
      </div>

      <div
        v-if="hasEffectiveElementHint"
        class="layer card-effective-element-badge"
        :title="effectiveElementHint || `视为${effectiveElementLabel}系`"
        :aria-label="`视为${effectiveElementLabel}系`"
      >
        {{ effectiveElementLabel }}
      </div>

      <div class="card-text-panel">
        <div class="card-footer-row">
          <div class="card-info-ribbon" :title="detailRibbonText">
            <span class="card-ribbon-text">
              {{ cardTypeLabel }}<span class="card-ribbon-sep">·</span>{{ elementLabel }}<template v-if="hasEffectiveElementHint"><span class="card-ribbon-arrow">→</span>{{ effectiveElementLabel }}</template>
            </span>
          </div>
          <button class="card-desc-btn" type="button" @click.stop="openDescriptionModal">
            查看描述
          </button>
        </div>
        <div v-if="exclusiveSkills.length" class="card-exclusive-row" :title="exclusiveSkillText">
          <span v-for="skill in exclusiveSkills" :key="skill" class="card-exclusive-corner">
            {{ skill }}
          </span>
        </div>
      </div>

      <div class="layer card-gloss"></div>
    </template>
  </div>

  <Teleport to="body">
    <div v-if="showDescModal" class="card-desc-modal-overlay" @click="closeDescriptionModal">
      <div class="card-desc-modal" role="dialog" aria-modal="true" :aria-label="`${card.name} 卡面描述`" @click.stop>
        <div class="card-desc-modal-title">{{ card.name }} · 卡面描述</div>
        <div class="card-desc-modal-content">{{ displayDescription }}</div>
        <button class="card-desc-modal-close" type="button" @click="closeDescriptionModal">关闭</button>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
/* 黑漆卡面 + 银色镀铬细边框；只引用 --bt-* 变量，元素色仅用于徽章外环 */
.card-shell {
  --el-ring: var(--bt-chrome-mid);
  border-radius: var(--bt-radius);
  background: var(--bt-bg-0);
  box-shadow:
    0 0 0 1px var(--bt-chrome-mid),
    0 0 0 2px rgba(0, 0, 0, 0.7),
    0 8px 16px rgba(0, 0, 0, 0.6);
  isolation: isolate;
}

/* 镀铬内框：细亮线 + 内侧暗线 */
.card-shell::after {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: inherit;
  pointer-events: none;
  z-index: 40;
  box-shadow:
    inset 0 0 0 1px var(--bt-chrome),
    inset 0 0 0 2px rgba(0, 0, 0, 0.55),
    inset 0 1px 0 2px rgba(233, 237, 243, 0.18);
}

.card-shell.selectable {
  transition: transform 180ms ease, box-shadow 180ms ease, filter 180ms ease, opacity 180ms ease;
}

.card-shell.selectable:hover {
  transform: translateY(-4px);
  box-shadow:
    0 0 0 1px var(--bt-chrome-hi),
    0 0 0 2px rgba(0, 0, 0, 0.7),
    0 12px 20px rgba(0, 0, 0, 0.65);
}

.card-shell.selected,
.card-shell.selected:hover {
  transform: translateY(-12px);
  box-shadow:
    0 0 0 1px var(--bt-select),
    0 0 0 2px rgba(0, 0, 0, 0.6),
    0 0 16px 2px var(--bt-select-glow),
    0 16px 26px rgba(0, 0, 0, 0.7);
}

.card-shell.selected::after {
  box-shadow:
    inset 0 0 0 1px var(--bt-select),
    inset 0 0 10px var(--bt-select-glow);
}

.card-shell:not(.selectable):not(.selected) {
  opacity: 0.62;
  filter: saturate(0.7) brightness(0.9);
}

.card-shell.card-battle-mini:not(.selectable):not(.selected) {
  opacity: 1;
  filter: none;
}

.layer {
  position: absolute;
  pointer-events: none;
}

/* 插画 */
.card-art-frame {
  inset: 0;
  border-radius: inherit;
  overflow: hidden;
  z-index: 1;
  background: var(--bt-bg-1);
}

.card-art-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
  object-position: center 30%;
  display: block;
  filter: saturate(0.95) contrast(1.05);
}

.card-art-overlay {
  position: absolute;
  inset: 0;
  background: radial-gradient(130% 95% at 50% 42%, transparent 52%, rgba(11, 15, 22, 0.55) 100%);
}

.card-vignette {
  inset: 0;
  z-index: 2;
  border-radius: inherit;
  background:
    linear-gradient(180deg, rgba(11, 15, 22, 0.72) 0%, rgba(11, 15, 22, 0) 22%),
    linear-gradient(0deg, rgba(11, 15, 22, 0.94) 0%, rgba(11, 15, 22, 0.6) 22%, rgba(11, 15, 22, 0) 40%);
}

.card-gloss {
  inset: 0;
  z-index: 30;
  border-radius: inherit;
  background: linear-gradient(125deg, rgba(233, 237, 243, 0.12) 0%, rgba(233, 237, 243, 0.03) 22%, transparent 45%);
  mix-blend-mode: screen;
}

/* 顶部：元素徽章 + 名称条 */
.card-element-medal {
  left: 5%;
  top: 3.5%;
  width: 21%;
  aspect-ratio: 1;
  z-index: 24;
  border-radius: 999px;
  background: var(--bt-bg-0);
  box-shadow:
    0 0 0 1px var(--bt-chrome-mid),
    inset 0 0 0 1.5px var(--el-ring),
    inset 0 0 6px rgba(0, 0, 0, 0.8),
    0 2px 5px rgba(0, 0, 0, 0.6);
  display: flex;
  align-items: center;
  justify-content: center;
}

.card-element-medal > span {
  width: 54%;
  height: 54%;
  display: block;
  color: var(--el-ring);
  filter: drop-shadow(0 0 2px rgba(0, 0, 0, 0.8));
}

.card-element-medal.is-transformed {
  box-shadow:
    0 0 0 1px var(--bt-chrome-hi),
    inset 0 0 0 1.5px var(--el-ring),
    inset 0 0 6px rgba(0, 0, 0, 0.8),
    0 0 6px var(--el-ring);
}

.card-title-banner {
  left: 21%;
  right: 5%;
  top: 6.5%;
  height: 10.5%;
  z-index: 20;
  display: flex;
  align-items: center;
  padding: 0 6px 0 calc(9% + 4px);
  border-radius: 0 var(--bt-radius-sm) var(--bt-radius-sm) 0;
  background: linear-gradient(90deg, var(--bt-panel-raised), var(--bt-panel) 70%, rgba(13, 18, 27, 0.6));
  box-shadow:
    inset 0 1px 0 var(--bt-hairline-soft),
    inset 0 -1px 0 var(--bt-hairline-soft);
  overflow: hidden;
}

.card-title-text {
  color: var(--bt-text);
  font-weight: 700;
  line-height: 1;
  letter-spacing: 0.04em;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.8);
}

/* 命格印章占住名称条右端 */
.card-title-banner.has-fate {
  right: 20%;
}

.card-fate-seal {
  right: 5%;
  top: 5%;
  width: 14%;
  aspect-ratio: 1;
  z-index: 24;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 3px;
  background: linear-gradient(180deg, #7a1f1f, #4a1010);
  box-shadow:
    0 0 0 1px var(--bt-chrome-mid),
    inset 0 0 0 1px rgba(255, 210, 180, 0.35),
    0 2px 4px rgba(0, 0, 0, 0.6);
  color: #ffe9d6;
  font-weight: 800;
  line-height: 1;
  text-shadow: 0 1px 1px rgba(0, 0, 0, 0.8);
}

/* 生效元素与印刷不同：徽章右下角小标记（显示生效元素文字） */
.card-effective-element-badge {
  left: 18%;
  top: 13%;
  z-index: 26;
  min-width: 1.35em;
  height: 1.35em;
  padding: 0 2px;
  border-radius: 999px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bt-panel-raised);
  color: var(--bt-text);
  font-size: 7px;
  font-weight: 800;
  line-height: 1;
  box-shadow:
    0 0 0 1px var(--bt-select),
    0 0 5px var(--bt-select-glow);
}

/* 底部信息条 */
.card-text-panel {
  position: absolute;
  left: 4%;
  right: 4%;
  bottom: 3.5%;
  z-index: 34;
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding: 3px 4px;
  border-radius: var(--bt-radius-sm);
  background: var(--bt-panel);
  box-shadow:
    0 0 0 1px var(--bt-hairline-soft),
    inset 0 1px 0 rgba(233, 237, 243, 0.06);
  pointer-events: auto;
}

.card-footer-row {
  display: flex;
  align-items: center;
  gap: 3px;
  min-width: 0;
}

.card-info-ribbon {
  flex: 1;
  min-width: 0;
  overflow: hidden;
}

.card-ribbon-text {
  display: block;
  color: var(--bt-text-sub);
  font-weight: 600;
  line-height: 1.2;
  letter-spacing: 0.02em;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.card-ribbon-sep {
  margin: 0 0.2em;
  color: var(--bt-text-dim);
}

.card-ribbon-arrow {
  margin: 0 0.15em;
  color: var(--bt-select);
}

.card-desc-btn {
  flex-shrink: 0;
  border: 0;
  border-radius: 3px;
  background: var(--bt-panel-raised);
  box-shadow: inset 0 0 0 1px var(--bt-hairline);
  color: var(--bt-chrome-hi);
  font-weight: 600;
  line-height: 1.15;
  padding: 1px 4px;
  cursor: pointer;
  transition: box-shadow 140ms ease, color 140ms ease;
}

.card-desc-btn:hover {
  color: var(--bt-select);
  box-shadow: inset 0 0 0 1px var(--bt-select);
}

.card-desc-btn:focus-visible {
  outline: 1px solid var(--bt-select);
  outline-offset: 1px;
}

.card-exclusive-row {
  display: flex;
  flex-wrap: wrap;
  gap: 2px;
  min-width: 0;
}

.card-exclusive-corner {
  max-width: 100%;
  padding: 0 3px;
  border-radius: 999px;
  background: var(--bt-bg-1);
  box-shadow: inset 0 0 0 1px var(--bt-hairline);
  color: var(--bt-chrome);
  font-weight: 600;
  line-height: 1.35;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* 卡背：黑漆 + 镀铬徽记 */
.card-back-art {
  inset: 0;
  border-radius: inherit;
  background: radial-gradient(120% 80% at 50% 0%, var(--bt-bg-1), var(--bt-bg-0) 70%);
}

.card-back-fx {
  inset: 7%;
  border-radius: var(--bt-radius-sm);
  box-shadow:
    inset 0 0 0 1px var(--bt-hairline),
    inset 0 0 0 3px rgba(0, 0, 0, 0.6),
    inset 0 0 0 4px var(--bt-hairline-soft);
  background: linear-gradient(125deg, rgba(233, 237, 243, 0.08), transparent 40%);
}

.card-back-symbol-wrap {
  inset: 0;
  z-index: 8;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 4px;
}

.card-back-symbol {
  width: 2.2em;
  height: 2.2em;
  border-radius: 999px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bt-bg-0);
  box-shadow:
    0 0 0 1px var(--bt-chrome),
    0 0 0 3px var(--bt-bg-0),
    0 0 0 4px var(--bt-chrome-lo);
  color: var(--bt-chrome-hi);
  font-size: 1.1em;
  line-height: 1;
}

.card-back-text {
  color: var(--bt-chrome);
  font-weight: 600;
  letter-spacing: 0.3em;
  padding-left: 0.3em;
  font-size: 0.8em;
}

/* 尺寸 */
.card-size-battle-mini .card-title-text { font-size: 7.5px; }
.card-size-battle-mini .card-fate-seal { font-size: 7px; }
.card-size-small .card-fate-seal { font-size: 9px; }
.card-size-medium .card-fate-seal { font-size: 10px; }
.card-size-default .card-fate-seal { font-size: 12px; }
.card-size-small .card-title-text { font-size: 9px; }
.card-size-medium .card-title-text { font-size: 10px; }
.card-size-default .card-title-text { font-size: 11px; }

.card-size-battle-mini .card-ribbon-text { font-size: 5.5px; }
.card-size-small .card-ribbon-text { font-size: 6.5px; }
.card-size-medium .card-ribbon-text { font-size: 7px; }
.card-size-default .card-ribbon-text { font-size: 8px; }

.card-size-battle-mini .card-desc-btn { font-size: 5.5px; padding: 0 2px; }
.card-size-small .card-desc-btn { font-size: 6.5px; }
.card-size-medium .card-desc-btn { font-size: 7px; }
.card-size-default .card-desc-btn { font-size: 8px; }

.card-size-battle-mini .card-exclusive-corner { font-size: 5.5px; }
.card-size-small .card-exclusive-corner { font-size: 6.5px; }
.card-size-medium .card-exclusive-corner { font-size: 7px; }
.card-size-default .card-exclusive-corner { font-size: 8px; }

.card-size-battle-mini .card-effective-element-badge { font-size: 5.5px; }
.card-size-default .card-effective-element-badge { font-size: 8px; }

.card-size-battle-mini .card-text-panel { padding: 2px 3px; gap: 1px; }

/* 描述弹窗 */
.card-desc-modal-overlay {
  position: fixed;
  inset: 0;
  z-index: 2000;
  background: rgba(11, 15, 22, 0.72);
  backdrop-filter: blur(2px);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px;
}

.card-desc-modal {
  width: min(560px, calc(100vw - 32px));
  max-height: min(82vh, 760px);
  overflow: hidden;
  border-radius: var(--bt-radius);
  background: var(--bt-panel-raised);
  box-shadow: var(--bt-frame-shadow);
  display: flex;
  flex-direction: column;
}

.card-desc-modal-title {
  padding: 10px 12px;
  border-bottom: 1px solid var(--bt-hairline-soft);
  color: var(--bt-text);
  font-size: 15px;
  font-weight: 700;
}

.card-desc-modal-content {
  padding: 12px;
  color: var(--bt-text-sub);
  font-size: 14px;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-word;
  overflow: auto;
}

.card-desc-modal-close {
  align-self: flex-end;
  margin: 0 12px 12px;
  min-width: 84px;
  border: 0;
  border-radius: var(--bt-radius-sm);
  background: var(--bt-panel);
  box-shadow: inset 0 0 0 1px var(--bt-hairline);
  color: var(--bt-chrome-hi);
  font-weight: 600;
  font-size: 13px;
  padding: 6px 10px;
  cursor: pointer;
}

.card-desc-modal-close:hover,
.card-desc-modal-close:focus-visible {
  box-shadow: inset 0 0 0 1px var(--bt-select);
  color: var(--bt-select);
  outline: none;
}

/* 徽章外环：每个元素一个克制的色相（唯一允许的非 --bt-* 色值） */
.badge-fire { --el-ring: #d9694a; }
.badge-water { --el-ring: #4f9fd8; }
.badge-wind { --el-ring: #4fb58a; }
.badge-earth { --el-ring: #c49a5a; }
.badge-thunder { --el-ring: #9a7fe0; }
.badge-light { --el-ring: #e3c46a; }
.badge-dark { --el-ring: #8c79c2; }

@media (prefers-reduced-motion: reduce) {
  .card-shell.selectable,
  .card-desc-btn {
    transition: none;
  }

  .card-shell.selectable:hover:not(.selected) {
    transform: none;
  }
}
</style>

