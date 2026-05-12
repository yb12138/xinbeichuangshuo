<template>
  <svg viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg" class="status-icon rose-courtyard-icon">
    <defs>
      <radialGradient id="roseGradient">
        <stop offset="0%" style="stop-color:#fb7185;stop-opacity:1" />
        <stop offset="100%" style="stop-color:#be123c;stop-opacity:1" />
      </radialGradient>
      <filter id="roseGlow">
        <feGaussianBlur stdDeviation="2" result="coloredBlur"/>
        <feMerge>
          <feMergeNode in="coloredBlur"/>
          <feMergeNode in="SourceGraphic"/>
        </feMerge>
      </filter>
    </defs>

    <!-- 玫瑰花瓣 -->
    <g class="rose-petals">
      <!-- 中心花瓣 -->
      <ellipse cx="50" cy="50" rx="8" ry="10" fill="url(#roseGradient)" filter="url(#roseGlow)" class="petal petal-center"/>

      <!-- 外层花瓣 -->
      <ellipse cx="50" cy="38" rx="10" ry="12" fill="url(#roseGradient)" opacity="0.9" class="petal petal-1"/>
      <ellipse cx="62" cy="44" rx="10" ry="12" fill="url(#roseGradient)" opacity="0.9" transform="rotate(60 62 44)" class="petal petal-2"/>
      <ellipse cx="62" cy="56" rx="10" ry="12" fill="url(#roseGradient)" opacity="0.9" transform="rotate(120 62 56)" class="petal petal-3"/>
      <ellipse cx="50" cy="62" rx="10" ry="12" fill="url(#roseGradient)" opacity="0.9" transform="rotate(180 50 62)" class="petal petal-4"/>
      <ellipse cx="38" cy="56" rx="10" ry="12" fill="url(#roseGradient)" opacity="0.9" transform="rotate(240 38 56)" class="petal petal-5"/>
      <ellipse cx="38" cy="44" rx="10" ry="12" fill="url(#roseGradient)" opacity="0.9" transform="rotate(300 38 44)" class="petal petal-6"/>
    </g>

    <!-- 荆棘藤蔓 -->
    <g class="thorns" stroke="#be123c" stroke-width="2" fill="none">
      <path d="M 20 80 Q 30 70, 40 75 Q 50 80, 60 75 Q 70 70, 80 80" class="vine vine-1"/>
      <path d="M 15 70 Q 25 60, 35 65 Q 45 70, 55 65 Q 65 60, 75 70" class="vine vine-2"/>

      <!-- 荆棘刺 -->
      <path d="M 25 75 L 23 70" stroke-width="2" stroke-linecap="round" class="thorn thorn-1"/>
      <path d="M 35 72 L 33 67" stroke-width="2" stroke-linecap="round" class="thorn thorn-2"/>
      <path d="M 50 77 L 48 72" stroke-width="2" stroke-linecap="round" class="thorn thorn-3"/>
      <path d="M 65 72 L 63 67" stroke-width="2" stroke-linecap="round" class="thorn thorn-4"/>
      <path d="M 75 75 L 73 70" stroke-width="2" stroke-linecap="round" class="thorn thorn-5"/>
    </g>

    <!-- 血滴效果 -->
    <g class="blood-drops">
      <ellipse cx="30" cy="65" rx="2" ry="3" fill="#be123c" class="drop drop-1"/>
      <ellipse cx="50" cy="70" rx="2" ry="3" fill="#be123c" class="drop drop-2"/>
      <ellipse cx="70" cy="65" rx="2" ry="3" fill="#be123c" class="drop drop-3"/>
    </g>

    <!-- 庭院光环 -->
    <g class="courtyard-aura" fill="none" stroke="#fb7185" stroke-width="2" opacity="0.4">
      <circle cx="50" cy="50" r="40" class="aura aura-1"/>
      <circle cx="50" cy="50" r="45" class="aura aura-2"/>
    </g>

    <!-- 玫瑰叶子 -->
    <g class="leaves" fill="#881337" opacity="0.7">
      <ellipse cx="35" cy="35" rx="5" ry="8" transform="rotate(-30 35 35)" class="leaf leaf-1"/>
      <ellipse cx="65" cy="35" rx="5" ry="8" transform="rotate(30 65 35)" class="leaf leaf-2"/>
    </g>
  </svg>
</template>

<style scoped>
.status-icon {
  width: 100%;
  height: 100%;
}

.petal {
  animation: petalBloom 3s ease-in-out infinite;
}

.petal-center { animation-delay: 0s; }
.petal-1 { animation-delay: 0.2s; }
.petal-2 { animation-delay: 0.4s; }
.petal-3 { animation-delay: 0.6s; }
.petal-4 { animation-delay: 0.8s; }
.petal-5 { animation-delay: 1s; }
.petal-6 { animation-delay: 1.2s; }

.vine {
  animation: vineGrow 4s ease-in-out infinite;
}

.vine-1 { animation-delay: 0s; }
.vine-2 { animation-delay: 2s; }

.thorn {
  animation: thornPrick 2s ease-in-out infinite;
}

.thorn-1 { animation-delay: 0s; }
.thorn-2 { animation-delay: 0.4s; }
.thorn-3 { animation-delay: 0.8s; }
.thorn-4 { animation-delay: 1.2s; }
.thorn-5 { animation-delay: 1.6s; }

.drop {
  animation: bloodDrip 3s ease-in-out infinite;
}

.drop-1 { animation-delay: 0s; }
.drop-2 { animation-delay: 1s; }
.drop-3 { animation-delay: 2s; }

.aura {
  animation: courtyardAuraExpand 3s ease-out infinite;
}

.aura-1 { animation-delay: 0s; }
.aura-2 { animation-delay: 1.5s; }

.leaf {
  animation: leafSway 2s ease-in-out infinite;
}

.leaf-1 { animation-delay: 0s; }
.leaf-2 { animation-delay: 1s; }

@keyframes petalBloom {
  0%, 100% {
    transform: scale(1);
    opacity: 0.9;
  }
  50% {
    transform: scale(1.05);
    opacity: 1;
  }
}

@keyframes vineGrow {
  0%, 100% {
    stroke-dasharray: 0 200;
  }
  50% {
    stroke-dasharray: 100 100;
  }
}

@keyframes thornPrick {
  0%, 100% {
    opacity: 0.6;
    transform: translateY(0);
  }
  50% {
    opacity: 1;
    transform: translateY(-2px);
  }
}

@keyframes bloodDrip {
  0% {
    transform: translateY(0);
    opacity: 0;
  }
  50% {
    opacity: 0.8;
  }
  100% {
    transform: translateY(10px);
    opacity: 0;
  }
}

@keyframes courtyardAuraExpand {
  0% {
    transform: scale(0.9);
    opacity: 0.6;
  }
  100% {
    transform: scale(1.1);
    opacity: 0;
  }
}

@keyframes leafSway {
  0%, 100% {
    transform: rotate(0deg);
  }
  50% {
    transform: rotate(5deg);
  }
}
</style>
