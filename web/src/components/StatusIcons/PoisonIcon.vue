<template>
  <svg viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg" class="status-icon poison-icon">
    <defs>
      <radialGradient id="poisonGradient">
        <stop offset="0%" style="stop-color:#22c55e;stop-opacity:1" />
        <stop offset="100%" style="stop-color:#15803d;stop-opacity:1" />
      </radialGradient>
      <filter id="poisonGlow">
        <feGaussianBlur stdDeviation="3" result="coloredBlur"/>
        <feMerge>
          <feMergeNode in="coloredBlur"/>
          <feMergeNode in="SourceGraphic"/>
        </feMerge>
      </filter>
    </defs>

    <!-- 骷髅头 -->
    <ellipse cx="50" cy="45" rx="25" ry="28" fill="url(#poisonGradient)" filter="url(#poisonGlow)" class="poison-skull"/>

    <!-- 眼睛 -->
    <ellipse cx="40" cy="40" rx="6" ry="8" fill="#000" class="poison-eye-left"/>
    <ellipse cx="60" cy="40" rx="6" ry="8" fill="#000" class="poison-eye-right"/>

    <!-- 鼻子 -->
    <path d="M 50 48 L 47 55 L 53 55 Z" fill="#000" class="poison-nose"/>

    <!-- 嘴巴/牙齿 -->
    <rect x="38" y="58" width="4" height="8" fill="#000" class="poison-tooth"/>
    <rect x="45" y="58" width="4" height="8" fill="#000" class="poison-tooth"/>
    <rect x="52" y="58" width="4" height="8" fill="#000" class="poison-tooth"/>
    <rect x="59" y="58" width="4" height="8" fill="#000" class="poison-tooth"/>

    <!-- 毒液滴落效果 -->
    <g class="poison-drops">
      <ellipse cx="30" cy="70" rx="3" ry="4" fill="#22c55e" opacity="0.8" class="drop drop-1"/>
      <ellipse cx="50" cy="75" rx="3" ry="4" fill="#22c55e" opacity="0.8" class="drop drop-2"/>
      <ellipse cx="70" cy="72" rx="3" ry="4" fill="#22c55e" opacity="0.8" class="drop drop-3"/>
    </g>

    <!-- 毒气效果 -->
    <g class="poison-smoke" opacity="0.4">
      <path d="M 25 30 Q 20 25, 22 20" stroke="#22c55e" stroke-width="2" fill="none" class="smoke smoke-1"/>
      <path d="M 75 30 Q 80 25, 78 20" stroke="#22c55e" stroke-width="2" fill="none" class="smoke smoke-2"/>
    </g>
  </svg>
</template>

<style scoped>
.status-icon {
  width: 100%;
  height: 100%;
}

.poison-skull {
  animation: poisonFloat 3s ease-in-out infinite;
}

.poison-eye-left, .poison-eye-right {
  animation: poisonEyeGlow 2s ease-in-out infinite;
}

.drop {
  animation: poisonDrop 2s ease-in-out infinite;
}

.drop-1 {
  animation-delay: 0s;
}

.drop-2 {
  animation-delay: 0.6s;
}

.drop-3 {
  animation-delay: 1.2s;
}

.smoke {
  animation: poisonSmoke 3s ease-in-out infinite;
}

.smoke-1 {
  animation-delay: 0s;
}

.smoke-2 {
  animation-delay: 1.5s;
}

@keyframes poisonFloat {
  0%, 100% {
    transform: translateY(0);
  }
  50% {
    transform: translateY(-5px);
  }
}

@keyframes poisonEyeGlow {
  0%, 100% {
    fill: #000;
  }
  50% {
    fill: #22c55e;
  }
}

@keyframes poisonDrop {
  0% {
    transform: translateY(0);
    opacity: 0;
  }
  50% {
    opacity: 0.8;
  }
  100% {
    transform: translateY(15px);
    opacity: 0;
  }
}

@keyframes poisonSmoke {
  0% {
    opacity: 0;
    transform: translateY(0);
  }
  50% {
    opacity: 0.6;
  }
  100% {
    opacity: 0;
    transform: translateY(-10px);
  }
}
</style>
