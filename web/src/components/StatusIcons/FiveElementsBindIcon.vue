<template>
  <svg viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg" class="status-icon bind-icon">
    <defs>
      <linearGradient id="bindGradient" x1="0%" y1="0%" x2="100%" y2="100%">
        <stop offset="0%" style="stop-color:#6b7280;stop-opacity:1" />
        <stop offset="50%" style="stop-color:#4b5563;stop-opacity:1" />
        <stop offset="100%" style="stop-color:#374151;stop-opacity:1" />
      </linearGradient>
      <filter id="bindGlow">
        <feGaussianBlur stdDeviation="2" result="coloredBlur"/>
        <feMerge>
          <feMergeNode in="coloredBlur"/>
          <feMergeNode in="SourceGraphic"/>
        </feMerge>
      </filter>
    </defs>

    <!-- 中心锁 -->
    <g class="bind-lock">
      <rect x="40" y="45" width="20" height="25" rx="2" fill="url(#bindGradient)" filter="url(#bindGlow)"/>
      <circle cx="50" cy="57" r="4" fill="#9ca3af"/>
      <rect x="48" y="57" width="4" height="8" fill="#9ca3af"/>
      <path d="M 35 45 L 35 35 Q 35 25, 50 25 Q 65 25, 65 35 L 65 45"
            fill="none" stroke="url(#bindGradient)" stroke-width="4" filter="url(#bindGlow)"/>
    </g>

    <!-- 锁链 -->
    <g class="bind-chains" stroke="#6b7280" stroke-width="3" fill="none">
      <!-- 左上链 -->
      <path d="M 30 30 Q 20 25, 15 20" class="chain chain-1"/>
      <ellipse cx="18" cy="23" rx="4" ry="6" fill="none" stroke="#6b7280" stroke-width="2" class="chain-link"/>

      <!-- 右上链 -->
      <path d="M 70 30 Q 80 25, 85 20" class="chain chain-2"/>
      <ellipse cx="82" cy="23" rx="4" ry="6" fill="none" stroke="#6b7280" stroke-width="2" class="chain-link"/>

      <!-- 左下链 -->
      <path d="M 30 70 Q 20 75, 15 80" class="chain chain-3"/>
      <ellipse cx="18" cy="77" rx="4" ry="6" fill="none" stroke="#6b7280" stroke-width="2" class="chain-link"/>

      <!-- 右下链 -->
      <path d="M 70 70 Q 80 75, 85 80" class="chain chain-4"/>
      <ellipse cx="82" cy="77" rx="4" ry="6" fill="none" stroke="#6b7280" stroke-width="2" class="chain-link"/>
    </g>

    <!-- 五行元素符号环绕 -->
    <g class="bind-elements" font-size="12" font-weight="bold" fill="#9ca3af">
      <text x="50" y="15" text-anchor="middle" class="element element-1">火</text>
      <text x="80" y="35" text-anchor="middle" class="element element-2">水</text>
      <text x="85" y="65" text-anchor="middle" class="element element-3">地</text>
      <text x="50" y="92" text-anchor="middle" class="element element-4">风</text>
      <text x="15" y="65" text-anchor="middle" class="element element-5">雷</text>
    </g>

    <!-- 束缚光环 -->
    <circle cx="50" cy="50" r="40" fill="none" stroke="#6b7280" stroke-width="2" opacity="0.3" class="bind-aura aura-1"/>
    <circle cx="50" cy="50" r="45" fill="none" stroke="#6b7280" stroke-width="1" opacity="0.2" class="bind-aura aura-2"/>
  </svg>
</template>

<style scoped>
.status-icon {
  width: 100%;
  height: 100%;
}

.bind-lock {
  animation: bindShake 3s ease-in-out infinite;
}

.chain {
  animation: bindChainTighten 2s ease-in-out infinite;
}

.chain-1 { animation-delay: 0s; }
.chain-2 { animation-delay: 0.5s; }
.chain-3 { animation-delay: 1s; }
.chain-4 { animation-delay: 1.5s; }

.element {
  animation: bindElementPulse 2s ease-in-out infinite;
}

.element-1 { animation-delay: 0s; }
.element-2 { animation-delay: 0.4s; }
.element-3 { animation-delay: 0.8s; }
.element-4 { animation-delay: 1.2s; }
.element-5 { animation-delay: 1.6s; }

.bind-aura {
  animation: bindAuraPulse 3s ease-in-out infinite;
}

.aura-1 {
  animation-delay: 0s;
}

.aura-2 {
  animation-delay: 1.5s;
}

@keyframes bindShake {
  0%, 100% {
    transform: translateX(0);
  }
  25% {
    transform: translateX(-2px);
  }
  75% {
    transform: translateX(2px);
  }
}

@keyframes bindChainTighten {
  0%, 100% {
    stroke-width: 3;
    opacity: 0.8;
  }
  50% {
    stroke-width: 4;
    opacity: 1;
  }
}

@keyframes bindElementPulse {
  0%, 100% {
    opacity: 0.5;
    transform: scale(1);
  }
  50% {
    opacity: 1;
    transform: scale(1.1);
  }
}

@keyframes bindAuraPulse {
  0% {
    transform: scale(0.95);
    opacity: 0.4;
  }
  50% {
    transform: scale(1.05);
    opacity: 0.2;
  }
  100% {
    transform: scale(0.95);
    opacity: 0.4;
  }
}
</style>
