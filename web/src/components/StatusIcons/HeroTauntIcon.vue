<template>
  <svg viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg" class="status-icon taunt-icon">
    <defs>
      <radialGradient id="tauntGradient">
        <stop offset="0%" style="stop-color:#dc2626;stop-opacity:1" />
        <stop offset="100%" style="stop-color:#991b1b;stop-opacity:1" />
      </radialGradient>
      <filter id="tauntGlow">
        <feGaussianBlur stdDeviation="2.5" result="coloredBlur"/>
        <feMerge>
          <feMergeNode in="coloredBlur"/>
          <feMergeNode in="SourceGraphic"/>
        </feMerge>
      </filter>
    </defs>

    <!-- 交叉剑 -->
    <g class="swords">
      <!-- 左剑 -->
      <g class="sword sword-left">
        <rect x="25" y="20" width="6" height="50" rx="1" fill="url(#tauntGradient)" filter="url(#tauntGlow)" transform="rotate(-45 28 45)"/>
        <polygon points="28,15 25,20 31,20" fill="#dc2626"/>
        <rect x="24" y="68" width="8" height="8" rx="1" fill="#7f1d1d" transform="rotate(-45 28 72)"/>
      </g>

      <!-- 右剑 -->
      <g class="sword sword-right">
        <rect x="69" y="20" width="6" height="50" rx="1" fill="url(#tauntGradient)" filter="url(#tauntGlow)" transform="rotate(45 72 45)"/>
        <polygon points="72,15 69,20 75,20" fill="#dc2626"/>
        <rect x="68" y="68" width="8" height="8" rx="1" fill="#7f1d1d" transform="rotate(45 72 72)"/>
      </g>
    </g>

    <!-- 中心盾牌 -->
    <ellipse cx="50" cy="50" rx="15" ry="18" fill="url(#tauntGradient)" filter="url(#tauntGlow)" class="shield"/>
    <path d="M 50 35 L 45 45 L 50 50 L 55 45 Z" fill="#fff" opacity="0.8" class="shield-emblem"/>

    <!-- 挑衅波纹 -->
    <g class="taunt-waves" fill="none" stroke="#dc2626" stroke-width="2">
      <circle cx="50" cy="50" r="30" opacity="0.6" class="wave wave-1"/>
      <circle cx="50" cy="50" r="35" opacity="0.4" class="wave wave-2"/>
      <circle cx="50" cy="50" r="40" opacity="0.3" class="wave wave-3"/>
    </g>

    <!-- 愤怒符号 -->
    <g class="anger-marks" fill="#dc2626">
      <path d="M 20 25 L 18 20 L 22 22 Z" class="mark mark-1"/>
      <path d="M 80 25 L 82 20 L 78 22 Z" class="mark mark-2"/>
      <path d="M 15 50 L 10 50 L 13 53 Z" class="mark mark-3"/>
      <path d="M 85 50 L 90 50 L 87 53 Z" class="mark mark-4"/>
    </g>
  </svg>
</template>

<style scoped>
.status-icon {
  width: 100%;
  height: 100%;
}

.sword {
  animation: tauntSwordClash 1.5s ease-in-out infinite;
}

.sword-left {
  transform-origin: 28px 45px;
  animation-delay: 0s;
}

.sword-right {
  transform-origin: 72px 45px;
  animation-delay: 0.75s;
}

.shield {
  animation: tauntShieldPulse 1s ease-in-out infinite;
}

.shield-emblem {
  animation: tauntEmblemFlash 1.5s ease-in-out infinite;
}

.wave {
  animation: tauntWaveExpand 2s ease-out infinite;
}

.wave-1 { animation-delay: 0s; }
.wave-2 { animation-delay: 0.6s; }
.wave-3 { animation-delay: 1.2s; }

.mark {
  animation: tauntMarkPop 1s ease-in-out infinite;
}

.mark-1 { animation-delay: 0s; }
.mark-2 { animation-delay: 0.25s; }
.mark-3 { animation-delay: 0.5s; }
.mark-4 { animation-delay: 0.75s; }

@keyframes tauntSwordClash {
  0%, 100% {
    transform: scale(1);
  }
  50% {
    transform: scale(1.05);
  }
}

@keyframes tauntShieldPulse {
  0%, 100% {
    transform: scale(1);
    opacity: 1;
  }
  50% {
    transform: scale(1.1);
    opacity: 0.9;
  }
}

@keyframes tauntEmblemFlash {
  0%, 100% {
    opacity: 0.8;
  }
  50% {
    opacity: 0.4;
  }
}

@keyframes tauntWaveExpand {
  0% {
    transform: scale(0.8);
    opacity: 0.6;
  }
  100% {
    transform: scale(1.3);
    opacity: 0;
  }
}

@keyframes tauntMarkPop {
  0%, 100% {
    transform: scale(1);
    opacity: 0.8;
  }
  50% {
    transform: scale(1.3);
    opacity: 1;
  }
}
</style>
