<template>
  <svg viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg" class="status-icon shield-icon">
    <defs>
      <!-- 盾牌金属主体渐变 -->
      <linearGradient id="shieldBody" x1="0%" y1="0%" x2="100%" y2="100%">
        <stop offset="0%" style="stop-color:#f0d060;stop-opacity:1" />
        <stop offset="35%" style="stop-color:#daa520;stop-opacity:1" />
        <stop offset="65%" style="stop-color:#b8860b;stop-opacity:1" />
        <stop offset="100%" style="stop-color:#8b6914;stop-opacity:1" />
      </linearGradient>
      <!-- 盾牌内框渐变 -->
      <linearGradient id="shieldInner" x1="0%" y1="0%" x2="100%" y2="100%">
        <stop offset="0%" style="stop-color:#e8c84a;stop-opacity:1" />
        <stop offset="40%" style="stop-color:#c9a030;stop-opacity:1" />
        <stop offset="100%" style="stop-color:#9a7520;stop-opacity:1" />
      </linearGradient>
      <!-- 盾牌中心宝石渐变 -->
      <radialGradient id="shieldGem" cx="40%" cy="35%">
        <stop offset="0%" style="stop-color:#fffacd;stop-opacity:1" />
        <stop offset="50%" style="stop-color:#ffd700;stop-opacity:1" />
        <stop offset="100%" style="stop-color:#b8860b;stop-opacity:1" />
      </radialGradient>
      <!-- 高光 -->
      <linearGradient id="shieldHighlight" x1="0%" y1="0%" x2="100%" y2="100%">
        <stop offset="0%" style="stop-color:#fff;stop-opacity:0.5" />
        <stop offset="100%" style="stop-color:#fff;stop-opacity:0" />
      </linearGradient>
      <!-- 边框金属渐变 -->
      <linearGradient id="shieldRim" x1="0%" y1="0%" x2="0%" y2="100%">
        <stop offset="0%" style="stop-color:#ffe680;stop-opacity:1" />
        <stop offset="30%" style="stop-color:#daa520;stop-opacity:1" />
        <stop offset="70%" style="stop-color:#8b6914;stop-opacity:1" />
        <stop offset="100%" style="stop-color:#6b4f10;stop-opacity:1" />
      </linearGradient>
      <filter id="shieldGlow">
        <feGaussianBlur stdDeviation="2.5" result="coloredBlur"/>
        <feMerge>
          <feMergeNode in="coloredBlur"/>
          <feMergeNode in="SourceGraphic"/>
        </feMerge>
      </filter>
      <filter id="innerShadow">
        <feGaussianBlur in="SourceAlpha" stdDeviation="2" result="blur"/>
        <feOffset dx="1" dy="2" result="offsetBlur"/>
        <feFlood flood-color="#5a3e0a" flood-opacity="0.4" result="color"/>
        <feComposite in="color" in2="offsetBlur" operator="in" result="shadow"/>
        <feMerge>
          <feMergeNode in="shadow"/>
          <feMergeNode in="SourceGraphic"/>
        </feMerge>
      </filter>
    </defs>

    <!-- 盾牌外轮廓（厚重边框） -->
    <path
      d="M 50 8
         C 28 8, 12 20, 12 36
         L 12 56
         C 12 72, 28 88, 50 93
         C 72 88, 88 72, 88 56
         L 88 36
         C 88 20, 72 8, 50 8 Z"
      fill="url(#shieldRim)"
      stroke="#6b4f10"
      stroke-width="1.5"
      filter="url(#shieldGlow)"
      class="shield-outer"
    />

    <!-- 盾牌金属主体 -->
    <path
      d="M 50 14
         C 32 14, 18 24, 18 38
         L 18 55
         C 18 69, 32 83, 50 87
         C 68 83, 82 69, 82 55
         L 82 38
         C 82 24, 68 14, 50 14 Z"
      fill="url(#shieldBody)"
      filter="url(#innerShadow)"
    />

    <!-- 金属高光条（左侧） -->
    <path
      d="M 50 16
         C 34 16, 21 25, 20 38
         L 20 42
         C 20 28, 34 18, 50 18 Z"
      fill="url(#shieldHighlight)"
      opacity="0.5"
    />

    <!-- 内框装饰线 -->
    <path
      d="M 50 22
         C 37 22, 27 30, 27 42
         L 27 53
         C 27 64, 37 76, 50 79
         C 63 76, 73 64, 73 53
         L 73 42
         C 73 30, 63 22, 50 22 Z"
      fill="url(#shieldInner)"
      stroke="#8b6914"
      stroke-width="1"
    />

    <!-- 内框高光 -->
    <path
      d="M 50 24
         C 38 24, 29 31, 28 42
         L 28 44
         C 29 33, 38 26, 50 26 Z"
      fill="#fff"
      opacity="0.2"
    />

    <!-- 横向装饰带 -->
    <path
      d="M 28 48 L 72 48 L 72 54 L 28 54 Z"
      fill="#8b6914"
      opacity="0.6"
    />
    <path
      d="M 28 48 L 72 48 L 72 50 L 28 50 Z"
      fill="#fff"
      opacity="0.15"
    />

    <!-- 中心菱形宝石底座 -->
    <path
      d="M 50 34 L 62 48 L 50 62 L 38 48 Z"
      fill="#6b4f10"
      stroke="#8b6914"
      stroke-width="1"
    />

    <!-- 中心宝石 -->
    <path
      d="M 50 37 L 59 48 L 50 59 L 41 48 Z"
      fill="url(#shieldGem)"
      class="shield-gem"
    />

    <!-- 宝石高光 -->
    <path
      d="M 50 39 L 46 44 L 50 42 L 54 44 Z"
      fill="#fff"
      opacity="0.7"
      class="shield-gem-highlight"
    />

    <!-- 底部尖角装饰 -->
    <path
      d="M 46 79 L 50 87 L 54 79"
      fill="#8b6914"
      stroke="#6b4f10"
      stroke-width="0.5"
    />

    <!-- 铆钉 - 左上 -->
    <circle cx="25" cy="30" r="2.5" fill="#daa520" stroke="#8b6914" stroke-width="0.5"/>
    <circle cx="25" cy="30" r="1" fill="#fff" opacity="0.4"/>

    <!-- 铆钉 - 右上 -->
    <circle cx="75" cy="30" r="2.5" fill="#daa520" stroke="#8b6914" stroke-width="0.5"/>
    <circle cx="75" cy="30" r="1" fill="#fff" opacity="0.4"/>

    <!-- 铆钉 - 左下 -->
    <circle cx="25" cy="64" r="2.5" fill="#daa520" stroke="#8b6914" stroke-width="0.5"/>
    <circle cx="25" cy="64" r="1" fill="#fff" opacity="0.4"/>

    <!-- 铆钉 - 右下 -->
    <circle cx="75" cy="64" r="2.5" fill="#daa520" stroke="#8b6914" stroke-width="0.5"/>
    <circle cx="75" cy="64" r="1" fill="#fff" opacity="0.4"/>

    <!-- 神圣光芒 -->
    <g class="shield-glow-rays" opacity="0.3">
      <line x1="50" y1="5" x2="50" y2="0" stroke="#ffd700" stroke-width="2" stroke-linecap="round"/>
      <line x1="30" y1="10" x2="26" y2="5" stroke="#ffd700" stroke-width="1.5" stroke-linecap="round"/>
      <line x1="70" y1="10" x2="74" y2="5" stroke="#ffd700" stroke-width="1.5" stroke-linecap="round"/>
      <line x1="14" y1="30" x2="8" y2="27" stroke="#ffd700" stroke-width="1.5" stroke-linecap="round"/>
      <line x1="86" y1="30" x2="92" y2="27" stroke="#ffd700" stroke-width="1.5" stroke-linecap="round"/>
    </g>
  </svg>
</template>

<style scoped>
.status-icon { width: 100%; height: 100%; }

.shield-outer {
  animation: shieldBreathe 2.5s ease-in-out infinite;
}

.shield-gem {
  animation: gemGlow 2s ease-in-out infinite;
}

.shield-gem-highlight {
  animation: gemSparkle 2s ease-in-out infinite;
}

.shield-glow-rays {
  animation: raysFlicker 2.5s ease-in-out infinite;
}

@keyframes shieldBreathe {
  0%, 100% { transform: scale(1); }
  50% { transform: scale(1.03); }
}

@keyframes gemGlow {
  0%, 100% { opacity: 1; filter: brightness(1); }
  50% { opacity: 0.85; filter: brightness(1.4); }
}

@keyframes gemSparkle {
  0%, 100% { opacity: 0.7; }
  50% { opacity: 0.3; }
}

@keyframes raysFlicker {
  0%, 100% { opacity: 0.3; }
  50% { opacity: 0.6; }
}
</style>
