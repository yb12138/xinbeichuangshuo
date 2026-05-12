<template>
  <svg viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg" class="status-icon soul-link-icon">
    <defs>
      <linearGradient id="soulGradient" x1="0%" y1="0%" x2="100%" y2="100%">
        <stop offset="0%" style="stop-color:#6366f1;stop-opacity:1" />
        <stop offset="100%" style="stop-color:#4f46e5;stop-opacity:1" />
      </linearGradient>
      <filter id="soulGlow">
        <feGaussianBlur stdDeviation="2" result="coloredBlur"/>
        <feMerge>
          <feMergeNode in="coloredBlur"/>
          <feMergeNode in="SourceGraphic"/>
        </feMerge>
      </filter>
    </defs>

    <!-- 左侧灵魂 -->
    <g class="soul soul-left">
      <circle cx="30" cy="50" r="12" fill="url(#soulGradient)" filter="url(#soulGlow)" opacity="0.8"/>
      <circle cx="30" cy="50" r="8" fill="#818cf8" opacity="0.6"/>
      <circle cx="28" cy="48" r="2" fill="#fff"/>
    </g>

    <!-- 右侧灵魂 -->
    <g class="soul soul-right">
      <circle cx="70" cy="50" r="12" fill="url(#soulGradient)" filter="url(#soulGlow)" opacity="0.8"/>
      <circle cx="70" cy="50" r="8" fill="#818cf8" opacity="0.6"/>
      <circle cx="68" cy="48" r="2" fill="#fff"/>
    </g>

    <!-- 连接链条 -->
    <g class="link-chain" stroke="#6366f1" stroke-width="3" fill="none">
      <path d="M 42 50 L 58 50" class="chain-line"/>
      <ellipse cx="45" cy="50" rx="4" ry="6" stroke-width="2" class="chain-link link-1"/>
      <ellipse cx="50" cy="50" rx="4" ry="6" stroke-width="2" class="chain-link link-2"/>
      <ellipse cx="55" cy="50" rx="4" ry="6" stroke-width="2" class="chain-link link-3"/>
    </g>

    <!-- 能量流动 -->
    <g class="energy-flow">
      <circle cx="42" cy="50" r="2" fill="#818cf8" class="energy-particle particle-1"/>
      <circle cx="50" cy="50" r="2" fill="#818cf8" class="energy-particle particle-2"/>
      <circle cx="58" cy="50" r="2" fill="#818cf8" class="energy-particle particle-3"/>
    </g>

    <!-- 灵魂光环 -->
    <g class="soul-aura" fill="none" stroke="#6366f1" stroke-width="1.5" opacity="0.4">
      <circle cx="30" cy="50" r="18" class="aura aura-left-1"/>
      <circle cx="30" cy="50" r="22" class="aura aura-left-2"/>
      <circle cx="70" cy="50" r="18" class="aura aura-right-1"/>
      <circle cx="70" cy="50" r="22" class="aura aura-right-2"/>
    </g>

    <!-- 连接符文 -->
    <g class="runes" fill="#818cf8" opacity="0.6">
      <circle cx="50" cy="40" r="2" class="rune rune-1"/>
      <circle cx="50" cy="60" r="2" class="rune rune-2"/>
    </g>
  </svg>
</template>

<style scoped>
.status-icon {
  width: 100%;
  height: 100%;
}

.soul {
  animation: soulFloat 3s ease-in-out infinite;
}

.soul-left {
  animation-delay: 0s;
}

.soul-right {
  animation-delay: 1.5s;
}

.chain-line {
  animation: chainPulse 2s ease-in-out infinite;
}

.chain-link {
  animation: chainLinkRotate 3s linear infinite;
}

.link-1 { animation-delay: 0s; }
.link-2 { animation-delay: 1s; }
.link-3 { animation-delay: 2s; }

.energy-particle {
  animation: energyFlow 2s ease-in-out infinite;
}

.particle-1 { animation-delay: 0s; }
.particle-2 { animation-delay: 0.6s; }
.particle-3 { animation-delay: 1.2s; }

.aura {
  animation: auraExpand 2s ease-out infinite;
}

.aura-left-1 { animation-delay: 0s; }
.aura-left-2 { animation-delay: 1s; }
.aura-right-1 { animation-delay: 0.5s; }
.aura-right-2 { animation-delay: 1.5s; }

.rune {
  animation: runeOrbit 4s linear infinite;
}

.rune-1 { animation-delay: 0s; }
.rune-2 { animation-delay: 2s; }

@keyframes soulFloat {
  0%, 100% {
    transform: translateY(0);
  }
  50% {
    transform: translateY(-5px);
  }
}

@keyframes chainPulse {
  0%, 100% {
    stroke-width: 3;
    opacity: 0.8;
  }
  50% {
    stroke-width: 4;
    opacity: 1;
  }
}

@keyframes chainLinkRotate {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

@keyframes energyFlow {
  0% {
    transform: translateX(-15px);
    opacity: 0;
  }
  50% {
    opacity: 1;
  }
  100% {
    transform: translateX(15px);
    opacity: 0;
  }
}

@keyframes auraExpand {
  0% {
    transform: scale(0.9);
    opacity: 0.6;
  }
  100% {
    transform: scale(1.2);
    opacity: 0;
  }
}

@keyframes runeOrbit {
  0% {
    transform: rotate(0deg) translateX(10px) rotate(0deg);
  }
  100% {
    transform: rotate(360deg) translateX(10px) rotate(-360deg);
  }
}
</style>
