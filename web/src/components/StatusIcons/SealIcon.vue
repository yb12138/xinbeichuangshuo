<template>
  <svg viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg" class="status-icon seal-icon">
    <defs>
      <radialGradient :id="`sealGradient-${element}`">
        <stop offset="0%" :style="`stop-color:${colors.light};stop-opacity:1`" />
        <stop offset="100%" :style="`stop-color:${colors.dark};stop-opacity:1`" />
      </radialGradient>
      <filter :id="`sealGlow-${element}`">
        <feGaussianBlur stdDeviation="3" result="coloredBlur"/>
        <feMerge>
          <feMergeNode in="coloredBlur"/>
          <feMergeNode in="SourceGraphic"/>
        </feMerge>
      </filter>
    </defs>

    <!-- 外圈 -->
    <circle
      cx="50" cy="50" r="35"
      fill="none"
      :stroke="colors.light"
      stroke-width="3"
      class="seal-outer-ring"
      :filter="`url(#sealGlow-${element})`"
    />

    <!-- 中圈 -->
    <circle
      cx="50" cy="50" r="28"
      fill="none"
      :stroke="colors.medium"
      stroke-width="2"
      class="seal-middle-ring"
    />

    <!-- 内圈 -->
    <circle
      cx="50" cy="50" r="20"
      :fill="`url(#sealGradient-${element})`"
      class="seal-inner-circle"
    />

    <!-- 元素符号 -->
    <text
      x="50" y="58"
      text-anchor="middle"
      font-size="24"
      font-weight="bold"
      fill="#fff"
      class="seal-symbol"
    >
      {{ symbol }}
    </text>

    <!-- 锁链效果 -->
    <g class="seal-chains" :stroke="colors.dark" stroke-width="2" fill="none">
      <path d="M 50 15 L 50 35" class="chain chain-top"/>
      <path d="M 50 65 L 50 85" class="chain chain-bottom"/>
      <path d="M 15 50 L 35 50" class="chain chain-left"/>
      <path d="M 65 50 L 85 50" class="chain chain-right"/>
    </g>

    <!-- 旋转的符文 -->
    <g class="seal-runes" :fill="colors.light" opacity="0.6">
      <circle cx="50" cy="20" r="3" class="rune rune-1"/>
      <circle cx="70" cy="30" r="3" class="rune rune-2"/>
      <circle cx="80" cy="50" r="3" class="rune rune-3"/>
      <circle cx="70" cy="70" r="3" class="rune rune-4"/>
      <circle cx="50" cy="80" r="3" class="rune rune-5"/>
      <circle cx="30" cy="70" r="3" class="rune rune-6"/>
      <circle cx="20" cy="50" r="3" class="rune rune-7"/>
      <circle cx="30" cy="30" r="3" class="rune rune-8"/>
    </g>
  </svg>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  element: 'fire' | 'water' | 'earth' | 'wind' | 'thunder'
}>()

const elementConfig = {
  fire: { symbol: '火', colors: { light: '#ef4444', medium: '#dc2626', dark: '#991b1b' } },
  water: { symbol: '水', colors: { light: '#3b82f6', medium: '#2563eb', dark: '#1e40af' } },
  earth: { symbol: '地', colors: { light: '#d97706', medium: '#b45309', dark: '#78350f' } },
  wind: { symbol: '风', colors: { light: '#14b8a6', medium: '#0d9488', dark: '#115e59' } },
  thunder: { symbol: '雷', colors: { light: '#6366f1', medium: '#4f46e5', dark: '#3730a3' } },
}

const symbol = computed(() => elementConfig[props.element].symbol)
const colors = computed(() => elementConfig[props.element].colors)
</script>

<style scoped>
.status-icon {
  width: 100%;
  height: 100%;
}

.seal-outer-ring {
  animation: sealRotate 8s linear infinite;
  transform-origin: center;
}

.seal-middle-ring {
  animation: sealRotate 6s linear infinite reverse;
  transform-origin: center;
}

.seal-inner-circle {
  animation: sealPulse 2s ease-in-out infinite;
}

.seal-symbol {
  animation: sealSymbolGlow 2s ease-in-out infinite;
}

.chain {
  animation: sealChainPulse 1.5s ease-in-out infinite;
}

.chain-top { animation-delay: 0s; }
.chain-right { animation-delay: 0.375s; }
.chain-bottom { animation-delay: 0.75s; }
.chain-left { animation-delay: 1.125s; }

.seal-runes {
  animation: sealRotate 12s linear infinite;
  transform-origin: center;
}

.rune {
  animation: sealRuneBlink 2s ease-in-out infinite;
}

.rune-1 { animation-delay: 0s; }
.rune-2 { animation-delay: 0.25s; }
.rune-3 { animation-delay: 0.5s; }
.rune-4 { animation-delay: 0.75s; }
.rune-5 { animation-delay: 1s; }
.rune-6 { animation-delay: 1.25s; }
.rune-7 { animation-delay: 1.5s; }
.rune-8 { animation-delay: 1.75s; }

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
    opacity: 1;
  }
  50% {
    transform: scale(1.05);
    opacity: 0.9;
  }
}

@keyframes sealSymbolGlow {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.7;
  }
}

@keyframes sealChainPulse {
  0%, 100% {
    opacity: 0.6;
  }
  50% {
    opacity: 1;
  }
}

@keyframes sealRuneBlink {
  0%, 100% {
    opacity: 0.3;
  }
  50% {
    opacity: 0.9;
  }
}
</style>
