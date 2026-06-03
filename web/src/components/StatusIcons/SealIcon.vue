<template>
  <img
    v-if="imageSrc"
    class="status-icon seal-image-icon"
    :src="imageSrc"
    :alt="label"
    :style="sealStyle"
  >
  <svg
    v-else
    viewBox="0 0 100 128"
    xmlns="http://www.w3.org/2000/svg"
    class="status-icon seal-icon"
    :style="sealStyle"
  >
    <defs>
      <linearGradient :id="`sealFrame-${element}`" x1="0%" y1="0%" x2="100%" y2="100%">
        <stop offset="0%" :style="`stop-color:${colors.light};stop-opacity:0.94`" />
        <stop offset="48%" :style="`stop-color:${colors.medium};stop-opacity:0.5`" />
        <stop offset="100%" :style="`stop-color:${colors.dark};stop-opacity:0.98`" />
      </linearGradient>
      <radialGradient :id="`sealCore-${element}`" cx="50%" cy="38%" r="62%">
        <stop offset="0%" :style="`stop-color:${colors.light};stop-opacity:0.96`" />
        <stop offset="55%" :style="`stop-color:${colors.medium};stop-opacity:0.72`" />
        <stop offset="100%" :style="`stop-color:${colors.dark};stop-opacity:0.28`" />
      </radialGradient>
      <filter :id="`sealGlow-${element}`" x="-42%" y="-32%" width="184%" height="164%">
        <feGaussianBlur stdDeviation="4.4" result="coloredBlur"/>
        <feMerge>
          <feMergeNode in="coloredBlur"/>
          <feMergeNode in="SourceGraphic"/>
        </feMerge>
      </filter>
      <filter :id="`sealSoftShadow-${element}`" x="-35%" y="-25%" width="170%" height="150%">
        <feDropShadow dx="0" dy="6" stdDeviation="5" flood-color="#020617" flood-opacity="0.72"/>
        <feDropShadow dx="0" dy="0" stdDeviation="4" :flood-color="colors.light" flood-opacity="0.58"/>
      </filter>
      <clipPath :id="`sealCardClip-${element}`">
        <rect x="13" y="8" width="74" height="112" rx="9" />
      </clipPath>
    </defs>

    <rect
      x="13"
      y="8"
      width="74"
      height="112"
      rx="9"
      fill="rgba(5, 13, 24, 0.88)"
      :stroke="colors.light"
      stroke-width="2"
      :filter="`url(#sealSoftShadow-${element})`"
      class="seal-card-base"
    />
    <rect
      x="17"
      y="12"
      width="66"
      height="104"
      rx="7"
      :fill="`url(#sealFrame-${element})`"
      opacity="0.24"
      class="seal-card-aura"
    />
    <g :clip-path="`url(#sealCardClip-${element})`" opacity="0.7">
      <path d="M 12 25 C 34 12, 51 18, 88 6" :stroke="colors.light" stroke-width="4" fill="none" opacity="0.28"/>
      <path d="M 12 110 C 39 94, 58 105, 88 86" :stroke="colors.light" stroke-width="5" fill="none" opacity="0.22"/>
      <path d="M 18 16 L 82 16 L 70 27 L 28 27 Z" fill="#ffffff" opacity="0.06"/>
    </g>

    <circle
      cx="50"
      cy="50"
      r="30"
      :fill="`url(#sealCore-${element})`"
      :stroke="colors.light"
      stroke-width="2.8"
      class="seal-core"
      :filter="`url(#sealGlow-${element})`"
    />
    <circle
      cx="50"
      cy="50"
      r="22"
      fill="none"
      :stroke="colors.light"
      stroke-width="1.7"
      opacity="0.56"
      class="seal-orbit"
    />
    <circle
      cx="50"
      cy="50"
      r="36"
      fill="none"
      :stroke="colors.light"
      stroke-width="1.2"
      stroke-dasharray="7 8"
      opacity="0.45"
      class="seal-rune-ring"
    />

    <g class="seal-symbol" :fill="colors.light" :stroke="colors.light" stroke-linecap="round" stroke-linejoin="round">
      <template v-if="isWater">
        <path d="M50 24 C42 36 35 44 35 55 C35 66 42 74 50 74 C58 74 65 66 65 55 C65 44 58 36 50 24 Z" fill="currentColor" opacity="0.95"/>
        <path d="M26 68 C35 77 65 77 74 68 M32 78 C41 85 59 85 68 78" fill="none" stroke-width="4" opacity="0.76"/>
      </template>
      <template v-else-if="isThunder">
        <path d="M57 22 L33 57 H48 L42 79 L69 43 H53 Z" fill="currentColor" opacity="0.96"/>
      </template>
      <template v-else-if="isFire">
        <path d="M51 25 C58 36 70 44 68 58 C66 71 57 78 49 78 C37 78 29 69 31 57 C32 49 38 44 39 36 C43 40 45 44 44 50 C51 44 54 36 51 25 Z" fill="currentColor" opacity="0.95"/>
        <path d="M50 55 C55 61 56 68 50 73 C43 70 42 62 50 55 Z" fill="#fff" opacity="0.2"/>
      </template>
      <template v-else-if="isEarth">
        <path d="M25 70 L41 42 L52 61 L61 50 L76 70 Z" fill="currentColor" opacity="0.9"/>
        <path d="M47 68 L54 53 L62 70 M37 69 L42 59" fill="none" stroke="#1b1209" stroke-width="3" opacity="0.35"/>
      </template>
      <template v-else>
        <path d="M29 50 C38 36 62 38 69 51 C59 45 44 47 38 57 C49 50 66 54 72 65 C58 59 41 61 31 73 C38 61 49 56 62 57 C51 50 39 51 29 50 Z" fill="currentColor" opacity="0.92"/>
        <path d="M32 77 C45 84 60 82 70 73" fill="none" stroke-width="4" opacity="0.7"/>
      </template>
    </g>

    <text
      x="50"
      y="102"
      text-anchor="middle"
      font-size="14"
      font-weight="800"
      fill="#f8fafc"
      class="seal-label"
    >
      {{ label }}
    </text>
  </svg>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  element: 'fire' | 'water' | 'earth' | 'wind' | 'thunder'
}>()

const elementConfig = {
  fire: { label: '火之封印', image: '/assets/status/seal-fire.png', colors: { light: '#fb6b5e', medium: '#ef4444', dark: '#7f1d1d' } },
  water: { label: '水之封印', image: '/assets/status/seal-water.png', colors: { light: '#7dd3fc', medium: '#38bdf8', dark: '#0f3c77' } },
  earth: { label: '地之封印', image: '/assets/status/seal-earth.png', colors: { light: '#f0b06a', medium: '#b87333', dark: '#5c3217' } },
  wind: { label: '风之封印', image: '/assets/status/seal-wind.png', colors: { light: '#86efac', medium: '#34d399', dark: '#14532d' } },
  thunder: { label: '雷之封印', image: '/assets/status/seal-thunder.png', colors: { light: '#fde047', medium: '#eab308', dark: '#713f12' } },
}

const label = computed(() => elementConfig[props.element].label)
const imageSrc = computed(() => elementConfig[props.element].image)
const colors = computed(() => elementConfig[props.element].colors)
const sealStyle = computed(() => ({
  color: colors.value.light,
  '--seal-glow-color': colors.value.light,
}))
const isWater = computed(() => props.element === 'water')
const isThunder = computed(() => props.element === 'thunder')
const isFire = computed(() => props.element === 'fire')
const isEarth = computed(() => props.element === 'earth')
</script>

<style scoped>
.status-icon {
  width: 100%;
  height: 100%;
}

.seal-icon {
  overflow: visible;
  filter: drop-shadow(0 0 10px color-mix(in srgb, var(--seal-glow-color), transparent 45%));
}

.seal-image-icon {
  object-fit: contain;
  object-position: center;
  background: transparent;
  filter:
    drop-shadow(0 0 8px color-mix(in srgb, var(--seal-glow-color), transparent 36%))
    drop-shadow(0 5px 10px rgba(0, 0, 0, 0.42));
}

.seal-card-aura {
  animation: sealCardShimmer 3.6s ease-in-out infinite;
}

.seal-core {
  animation: sealPulse 2.2s ease-in-out infinite;
}

.seal-orbit {
  transform-origin: 50px 50px;
  animation: sealRotate 7s linear infinite reverse;
}

.seal-rune-ring {
  transform-origin: 50px 50px;
  animation: sealRotate 11s linear infinite;
}

.seal-symbol {
  filter: drop-shadow(0 0 7px color-mix(in srgb, currentColor, transparent 20%));
  animation: sealSymbolGlow 2.1s ease-in-out infinite;
}

.seal-label {
  letter-spacing: 0;
  paint-order: stroke;
  stroke: rgba(3, 7, 18, 0.72);
  stroke-width: 3px;
  stroke-linejoin: round;
}

@keyframes sealRotate {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

@keyframes sealPulse {
  0%, 100% {
    transform: scale(1);
    opacity: 0.96;
  }
  50% {
    transform: scale(1.04);
    opacity: 1;
  }
}

@keyframes sealSymbolGlow {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.78;
  }
}

@keyframes sealCardShimmer {
  0%, 100% {
    opacity: 0.2;
  }
  50% {
    opacity: 0.42;
  }
}
</style>
