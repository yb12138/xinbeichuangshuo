<script setup lang="ts">
import { computed, ref, useAttrs, watch } from 'vue'
import type { Card } from '../types/game'

defineOptions({
  inheritAttrs: false
})

const props = defineProps<{
  card: Card
  index?: number
  selected?: boolean
  selectable?: boolean
  faceDown?: boolean
  small?: boolean
  medium?: boolean
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

const CARD_IMAGE_PINYIN_MAP: Record<string, string> = {
  火焰斩: 'huoYanZhan',
  水涟斩: 'shuiLianZhan',
  地裂斩: 'diLieZhan',
  风神斩: 'fengShenZhan',
  雷光斩: 'leiGuangZhan',
  暗灭: 'anMie',
  魔弹: 'moDan',
  圣光: 'shengGuang',
  圣盾: 'shengDun',
  虚弱: 'xuRuo',
  中毒: 'zhongDu'
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

const detailRibbonText = computed(() => {
  const base = `${elementLabel.value}系${props.card.type === 'Attack' ? '攻击' : '法术'}`
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

function fallbackPinyinByCard(card: Card): string {
  if (card.type === 'Attack') {
    if (card.name.includes('暗')) return 'anMie'
    if (card.name.includes('火')) return 'huoYanZhan'
    if (card.name.includes('水')) return 'shuiLianZhan'
    if (card.name.includes('地')) return 'diLieZhan'
    if (card.name.includes('风')) return 'fengShenZhan'
    if (card.name.includes('雷')) return 'leiGuangZhan'
    if (card.element === 'Fire') return 'huoYanZhan'
    if (card.element === 'Water') return 'shuiLianZhan'
    if (card.element === 'Earth') return 'diLieZhan'
    if (card.element === 'Wind') return 'fengShenZhan'
    if (card.element === 'Thunder') return 'leiGuangZhan'
    if (card.element === 'Dark') return 'anMie'
    return 'moRen'
  }
  if (card.name.includes('圣盾')) return 'shengDun'
  if (card.name.includes('圣光')) return 'shengGuang'
  if (card.name.includes('魔弹')) return 'moDan'
  if (card.name.includes('中毒')) return 'zhongDu'
  if (card.name.includes('虚弱')) return 'xuRuo'
  return 'shiShuCard'
}

const cardArtFile = computed(() => {
  if (imageLoadFailed.value) return props.card.type === 'Attack' ? 'moRen' : 'shiShuCard'
  return CARD_IMAGE_PINYIN_MAP[props.card.name] ?? fallbackPinyinByCard(props.card)
})

const cardArtSrc = computed(() => `/image/${cardArtFile.value}.png`)
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
    :class="[
      elementClass,
      small
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
      <div class="layer card-base"></div>
      <div class="layer card-vignette"></div>

      <div class="layer card-title-banner">
        <span class="card-title-text">{{ card.name }}</span>
      </div>

      <div class="layer card-element-medal">
        <span>{{ elementLabel }}</span>
      </div>

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

      <div class="layer card-info-ribbon">
        <span class="card-ribbon-text">{{ detailRibbonText }}</span>
      </div>

      <div class="card-text-panel">
        <button class="card-desc-btn" type="button" @click.stop="openDescriptionModal">
          查看描述
        </button>
        <div class="card-text-empty"></div>
        <div v-if="exclusiveSkillText" class="card-exclusive-corner" :title="exclusiveSkillText">
          {{ exclusiveSkillText }}
        </div>
      </div>

      <div class="layer card-gloss"></div>
      <div class="layer card-noise"></div>
    </template>
  </div>

  <Teleport to="body">
    <div v-if="showDescModal" class="card-desc-modal-overlay" @click="closeDescriptionModal">
      <div class="card-desc-modal" @click.stop>
        <div class="card-desc-modal-title">{{ card.name }} · 卡面描述</div>
        <div class="card-desc-modal-content">{{ displayDescription }}</div>
        <button class="card-desc-modal-close" type="button" @click="closeDescriptionModal">关闭</button>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.card-shell {
  --edge-glow: rgba(250, 206, 120, 0.42);
  --edge-color: rgba(197, 165, 112, 0.72);
  --base-top: #2f2520;
  --base-bottom: #17120f;
  --ribbon-start: #a82d2d;
  --ribbon-end: #7f1717;
  border: 2px solid var(--edge-color);
  border-radius: 10px;
  background: linear-gradient(180deg, var(--base-top), var(--base-bottom));
  box-shadow:
    0 10px 18px rgba(0, 0, 0, 0.7),
    0 0 12px var(--edge-glow),
    inset 0 0 14px rgba(255, 255, 255, 0.08),
    inset 0 0 0 1px rgba(255, 245, 224, 0.2);
}

.card-shell.selectable {
  transition: transform 180ms ease, box-shadow 180ms ease, border-color 180ms ease, filter 180ms ease;
}

.card-shell.selectable:hover {
  transform: translateY(-4px);
  border-color: rgba(255, 226, 163, 0.95);
  box-shadow:
    0 16px 24px rgba(0, 0, 0, 0.75),
    0 0 18px rgba(255, 213, 138, 0.4),
    inset 0 0 0 1px rgba(255, 244, 214, 0.4);
}

.card-shell.selected {
  transform: translateY(-12px);
  box-shadow:
    0 20px 30px rgba(0, 0, 0, 0.78),
    inset 0 0 0 1px rgba(255, 245, 224, 0.22);
}

.card-shell.selected:hover {
  transform: translateY(-12px);
  border-color: var(--edge-color);
  box-shadow:
    0 20px 30px rgba(0, 0, 0, 0.78),
    inset 0 0 0 1px rgba(255, 245, 224, 0.22);
}

.card-shell:not(.selectable):not(.selected) {
  opacity: 0.62;
  filter: saturate(0.76);
}

.layer {
  position: absolute;
  pointer-events: none;
}

.card-base {
  inset: 0;
  border-radius: 8px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.08), transparent 24%),
    linear-gradient(180deg, rgba(19, 14, 12, 0), rgba(19, 14, 12, 0.65));
}

.card-vignette {
  inset: 0;
  border-radius: 8px;
  box-shadow: inset 0 0 20px rgba(0, 0, 0, 0.62);
}

.card-title-banner {
  left: 18%;
  right: 4%;
  top: 3%;
  height: 12%;
  border-radius: 4px 6px 6px 4px;
  border: 1px solid rgba(198, 181, 147, 0.9);
  background:
    linear-gradient(180deg, rgba(251, 249, 243, 0.96), rgba(221, 213, 194, 0.95)),
    url('/image/cardtempname_bg.png') center/cover no-repeat;
  display: flex;
  align-items: center;
  justify-content: flex-start;
  padding: 0 8px;
  z-index: 20;
}

.card-title-text {
  color: #2d231d;
  font-weight: 800;
  line-height: 1;
  letter-spacing: 0.2px;
  text-shadow: 0 1px 0 rgba(255, 255, 255, 0.55);
}

.card-element-medal {
  left: 2.8%;
  top: 2%;
  width: 17%;
  height: 17%;
  border-radius: 999px;
  border: 1px solid rgba(255, 255, 255, 0.7);
  background: radial-gradient(circle at 30% 30%, rgba(255, 255, 255, 0.95), rgba(200, 191, 175, 0.8) 42%, rgba(61, 53, 47, 0.88));
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 24;
  box-shadow:
    0 2px 6px rgba(0, 0, 0, 0.55),
    0 0 8px rgba(255, 244, 189, 0.44);
}

.card-element-medal > span {
  width: 70%;
  height: 70%;
  border-radius: 999px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: var(--medal-bg);
  color: var(--medal-fg);
  font-weight: 900;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.7);
}

.card-art-frame {
  left: 5%;
  right: 5%;
  top: 16%;
  height: 44%;
  border-radius: 4px;
  border: 1px solid rgba(174, 160, 130, 0.76);
  overflow: hidden;
  z-index: 12;
  box-shadow:
    inset 0 0 0 1px rgba(255, 236, 196, 0.3),
    inset 0 0 16px rgba(0, 0, 0, 0.3);
}

.card-art-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
  object-position: center;
  display: block;
  filter: saturate(1.06) contrast(1.04);
}

.card-art-overlay {
  position: absolute;
  inset: 0;
  background:
    linear-gradient(180deg, rgba(255, 250, 235, 0.12), rgba(0, 0, 0, 0.2)),
    radial-gradient(120% 100% at 50% 0%, rgba(255, 255, 255, 0.13), transparent 54%);
}

.card-info-ribbon {
  left: 4%;
  right: 4%;
  top: 60%;
  height: 9%;
  border-radius: 4px;
  border: 1px solid rgba(122, 20, 20, 0.72);
  background: linear-gradient(90deg, var(--ribbon-start), var(--ribbon-end));
  display: flex;
  align-items: center;
  padding: 0 6px;
  z-index: 18;
}

.card-ribbon-text {
  color: #fdf2e6;
  font-weight: 700;
  letter-spacing: 0.2px;
  line-height: 1.08;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.6);
  white-space: normal;
  word-break: keep-all;
  overflow: hidden;
}

.card-text-panel {
  position: absolute;
  left: 4%;
  right: 4%;
  top: 70%;
  bottom: 4%;
  border-radius: 3px;
  border: 1px solid rgba(198, 193, 183, 0.98);
  background: #ffffff;
  padding: 4px;
  z-index: 44;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  pointer-events: auto;
}

.card-desc-btn {
  width: 100%;
  border: 1px solid rgba(116, 100, 80, 0.8);
  border-radius: 3px;
  background: linear-gradient(180deg, rgba(241, 237, 227, 0.98), rgba(220, 210, 191, 0.96));
  color: #342b23;
  font-weight: 700;
  line-height: 1.1;
  padding: 2px 4px;
  text-align: center;
  cursor: pointer;
  transition: filter 140ms ease, transform 140ms ease;
}

.card-desc-btn:hover {
  filter: brightness(1.04);
  transform: translateY(-1px);
}

.card-text-empty {
  flex: 1;
}

.card-exclusive-corner {
  align-self: flex-end;
  max-width: 100%;
  color: #4b3d31;
  background: linear-gradient(180deg, rgba(244, 240, 231, 0.98), rgba(226, 218, 202, 0.96));
  border: 1px solid rgba(167, 150, 125, 0.88);
  border-radius: 3px;
  padding: 1px 3px;
  line-height: 1.15;
  font-weight: 700;
  white-space: normal;
  word-break: break-word;
  overflow: hidden;
}

.card-gloss {
  inset: 0;
  border-radius: 8px;
  z-index: 30;
  background: linear-gradient(120deg, rgba(255, 255, 255, 0.22) 0%, rgba(255, 255, 255, 0.07) 24%, transparent 52%);
  mix-blend-mode: screen;
}

.card-noise {
  inset: 0;
  border-radius: 8px;
  z-index: 32;
  opacity: 0.15;
  background:
    radial-gradient(circle at 12% 18%, rgba(255, 255, 255, 0.4) 0 0.6px, transparent 0.7px),
    radial-gradient(circle at 77% 64%, rgba(255, 255, 255, 0.32) 0 0.5px, transparent 0.6px),
    radial-gradient(circle at 44% 91%, rgba(255, 255, 255, 0.26) 0 0.7px, transparent 0.8px);
}

.card-back-art {
  inset: 0;
  border-radius: 8px;
  background:
    linear-gradient(145deg, rgba(8, 22, 40, 0.96), rgba(15, 44, 71, 0.9)),
    url('/image/moRen.png') center/cover no-repeat;
}

.card-back-fx {
  inset: 0;
  border-radius: 8px;
  background:
    radial-gradient(60% 55% at 50% 44%, rgba(124, 190, 255, 0.34), transparent),
    linear-gradient(180deg, rgba(255, 255, 255, 0.15), transparent 48%);
}

.card-back-symbol-wrap {
  inset: 0;
  z-index: 8;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 2px;
}

.card-back-symbol {
  color: #e6f3ff;
  text-shadow: 0 0 12px rgba(130, 196, 255, 0.62), 0 2px 2px rgba(0, 0, 0, 0.6);
  font-size: 1.35em;
  font-weight: 800;
}

.card-back-text {
  color: #d8e7f3;
  font-weight: 700;
  letter-spacing: 0.16em;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.7);
}

.card-size-small .card-title-text {
  font-size: 9px;
}

.card-size-medium .card-title-text {
  font-size: 10px;
}

.card-size-default .card-title-text {
  font-size: 11px;
}

.card-size-small .card-element-medal > span {
  font-size: 8px;
}

.card-size-medium .card-element-medal > span {
  font-size: 9px;
}

.card-size-default .card-element-medal > span {
  font-size: 10px;
}

.card-size-small .card-ribbon-text {
  font-size: 6.5px;
}

.card-size-medium .card-ribbon-text {
  font-size: 7px;
}

.card-size-default .card-ribbon-text {
  font-size: 8px;
}

.card-size-small .card-desc-btn {
  font-size: 7px;
}

.card-size-medium .card-desc-btn {
  font-size: 7.5px;
}

.card-size-default .card-desc-btn {
  font-size: 8.5px;
}

.card-size-small .card-exclusive-corner {
  font-size: 6.5px;
}

.card-size-medium .card-exclusive-corner {
  font-size: 9px;
}

.card-size-default .card-exclusive-corner {
  font-size: 14px;
}

.card-desc-modal-overlay {
  position: fixed;
  inset: 0;
  z-index: 2000;
  background: rgba(6, 8, 14, 0.72);
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
  border-radius: 10px;
  border: 1px solid rgba(176, 157, 124, 0.72);
  background: linear-gradient(180deg, rgba(249, 246, 239, 0.98), rgba(235, 228, 216, 0.97));
  box-shadow: 0 18px 48px rgba(0, 0, 0, 0.52);
  display: flex;
  flex-direction: column;
}

.card-desc-modal-title {
  padding: 10px 12px;
  border-bottom: 1px solid rgba(170, 151, 120, 0.55);
  color: #2f261e;
  font-size: 15px;
  font-weight: 800;
}

.card-desc-modal-content {
  padding: 12px;
  color: #2c241c;
  font-size: 14px;
  line-height: 1.55;
  white-space: pre-wrap;
  word-break: break-word;
  overflow: auto;
}

.card-desc-modal-close {
  align-self: flex-end;
  margin: 0 12px 12px;
  min-width: 84px;
  border-radius: 6px;
  border: 1px solid rgba(122, 101, 76, 0.8);
  background: linear-gradient(180deg, rgba(241, 232, 215, 0.98), rgba(219, 203, 176, 0.96));
  color: #32271d;
  font-weight: 700;
  font-size: 13px;
  padding: 6px 10px;
  cursor: pointer;
}

.element-fire {
  --edge-color: rgba(205, 123, 82, 0.78);
  --edge-glow: rgba(255, 140, 98, 0.4);
  --base-top: #3c1f18;
  --base-bottom: #1c120f;
  --ribbon-start: #c6352f;
  --ribbon-end: #8e1b17;
  --medal-bg: radial-gradient(circle at 32% 28%, #ffca7f, #f36d33 58%, #9b2e1a);
  --medal-fg: #fff8eb;
}

.element-water {
  --edge-color: rgba(102, 152, 196, 0.78);
  --edge-glow: rgba(124, 196, 255, 0.38);
  --base-top: #1a2a3e;
  --base-bottom: #0f1826;
  --ribbon-start: #25689f;
  --ribbon-end: #1b446c;
  --medal-bg: radial-gradient(circle at 32% 28%, #ccf3ff, #4ea1d6 58%, #195580);
  --medal-fg: #effbff;
}

.element-earth {
  --edge-color: rgba(174, 138, 93, 0.8);
  --edge-glow: rgba(225, 186, 113, 0.34);
  --base-top: #32261a;
  --base-bottom: #1a1410;
  --ribbon-start: #8a5d2f;
  --ribbon-end: #60401f;
  --medal-bg: radial-gradient(circle at 32% 28%, #f1d79b, #c6924f 58%, #784d1d);
  --medal-fg: #fff6e4;
}

.element-wind {
  --edge-color: rgba(96, 169, 145, 0.78);
  --edge-glow: rgba(116, 223, 181, 0.34);
  --base-top: #183329;
  --base-bottom: #101f1b;
  --ribbon-start: #237258;
  --ribbon-end: #194e3d;
  --medal-bg: radial-gradient(circle at 32% 28%, #c8f9e6, #55b68d 58%, #216f54);
  --medal-fg: #edfff6;
}

.element-thunder {
  --edge-color: rgba(140, 124, 200, 0.8);
  --edge-glow: rgba(183, 148, 255, 0.36);
  --base-top: #24213d;
  --base-bottom: #171427;
  --ribbon-start: #5f4a99;
  --ribbon-end: #40306f;
  --medal-bg: radial-gradient(circle at 32% 28%, #efe2ff, #9c79dc 58%, #4e3385);
  --medal-fg: #faf3ff;
}

.element-light {
  --edge-color: rgba(205, 176, 103, 0.8);
  --edge-glow: rgba(255, 217, 123, 0.4);
  --base-top: #4a3a1a;
  --base-bottom: #231a0c;
  --ribbon-start: #b9892e;
  --ribbon-end: #7f5f21;
  --medal-bg: radial-gradient(circle at 32% 28%, #fff7cc, #e7c05c 58%, #9f7826);
  --medal-fg: #fff9ec;
}

.element-dark {
  --edge-color: rgba(128, 132, 160, 0.82);
  --edge-glow: rgba(149, 155, 214, 0.4);
  --base-top: #26293a;
  --base-bottom: #151925;
  --ribbon-start: #444a79;
  --ribbon-end: #2b3156;
  --medal-bg: radial-gradient(circle at 32% 28%, #dfe6ff, #7a86c7 58%, #3e497c);
  --medal-fg: #f1f4ff;
}

.card-face-down {
  --edge-color: rgba(118, 159, 191, 0.82);
  --edge-glow: rgba(147, 206, 255, 0.35);
  --base-top: #11213a;
  --base-bottom: #0b1627;
  --ribbon-start: #245a80;
  --ribbon-end: #193e59;
  --medal-bg: radial-gradient(circle at 32% 28%, #d6f5ff, #67b0d7 58%, #2b678e);
  --medal-fg: #f2fbff;
}
</style>
