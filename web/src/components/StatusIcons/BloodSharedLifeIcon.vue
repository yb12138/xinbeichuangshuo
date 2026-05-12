<template>
  <svg viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg" class="status-icon blood-shared-icon">
    <defs>
      <radialGradient id="bloodGradient">
        <stop offset="0%" style="stop-color:#f43f5e;stop-opacity:1" />
        <stop offset="100%" style="stop-color:#be123c;stop-opacity:1" />
      </radialGradient>
      <filter id="bloodGlow">
        <feGaussianBlur stdDeviation="2.5" result="coloredBlur"/>
        <feMerge>
          <feMergeNode in="coloredBlur"/>
          <feMergeNode in="SourceGraphic"/>
        </feMerge>
      </filter>
    </defs>

    <!-- 左心脏 -->
    <g class="heart heart-left">
      <path d="M 30 45
               C 30 40, 25 35, 20 35
               C 15 35, 12 40, 12 45
               C 12 55, 30 65, 30 65
               C 30 65, 12 55, 12 45
               M 30 45
               C 30 40, 35 35, 40 35
               C 45 35, 48 40, 48 45
               C 48 55, 30 65, 30 65
               C 30 65, 48 55, 48 45"
            fill="url(#bloodGradient)" filter="url(#bloodGlow)"/>
    </g>

    <!-- 右心脏 -->
    <g class="heart heart-right">
      <path d="M 70 45
               C 70 40, 65 35, 60 35
               C 55 35, 52 40, 52 45
               C 52 55, 70 65, 70 65
               C 70 65, 52 55, 52 45
               M 70 45
               C 70 40, 75 35, 80 35
               C 85 35, 88 40, 88 45
               C 88 55, 70 65, 70 65
               C 70 65, 88 55, 88 45"
            fill="url(#bloodGradient)" filter="url(#bloodGlow)"/>
    </g>

    <!-- 血管连接 -->
    <g class="blood-vessels" stroke="#f43f5e" stroke-width="3" fill="none">
      <path d="M 48 50 Q 50 45, 52 50" class="vessel vessel-1"/>
      <path d="M 48 55 Q 50 60, 52 55" class="vessel vessel-2"/>
    </g>

    <!-- 血液流动 -->
    <g class="blood-flow">
      <circle cx="48" cy="50" r="2" fill="#fb7185" class="blood-drop drop-1"/>
      <circle cx="50" cy="50" r="2" fill="#fb7185" class="blood-drop drop-2"/>
      <circle cx="52" cy="50" r="2" fill="#fb7185" class="blood-drop drop-3"/>
    </g>

    <!-- 生命脉搏波 -->
    <g class="pulse-waves" fill="none" stroke="#f43f5e" stroke-width="2" opacity="0.5">
      <circle cx="30" cy="50" r="20" class="pulse pulse-left-1"/>
      <circle cx="30" cy="50" r="25" class="pulse pulse-left-2"/>
      <circle cx="70" cy="50" r="20" class="pulse pulse-right-1"/>
      <circle cx="70" cy="50" r="25" class="pulse pulse-right-2"/>
    </g>

    <!-- 共生符文 -->
    <g class="runes" fill="#fb7185" opacity="0.7">
      <circle cx="50" cy="35" r="2" class="rune rune-1"/>
      <circle cx="50" cy="65" r="2" class="rune rune-2"/>
      <circle cx="40" cy="50" r="2" class="rune rune-3"/>
      <circle cx="60" cy="50" r="2" class="rune rune-4"/>
    </g>
  </svg>
</template>

<style scoped>
.status-icon {
  width: 100%;
  height: 100%;
}

.heart {
  animation: heartBeat 1.2s ease-in-out infinite;
}

.heart-left {
  animation-delay: 0s;
}

.heart-right {
  animation-delay: 0.6s;
}

.vessel {
  animation: vesselPulse 1.5s ease-in-out infinite;
}

.vessel-1 { animation-delay: 0s; }
.vessel-2 { animation-delay: 0.75s; }

.blood-drop {
  animation: bloodFlowAcross 2s ease-in-out infinite;
}

.drop-1 { animation-delay: 0s; }
.drop-2 { animation-delay: 0.6s; }
.drop-3 { animation-delay: 1.2s; }

.pulse {
  animation: pulseExpand 2s ease-out infinite;
}

.pulse-left-1 { animation-delay: 0s; }
.pulse-left-2 { animation-delay: 1s; }
.pulse-right-1 { animation-delay: 0.6s; }
.pulse-right-2 { animation-delay: 1.6s; }

.rune {
  animation: runeGlow 2s ease-in-out infinite;
}

.rune-1 { animation-delay: 0s; }
.rune-2 { animation-delay: 0.5s; }
.rune-3 { animation-delay: 1s; }
.rune-4 { animation-delay: 1.5s; }

@keyframes heartBeat {
  0%, 100% {
    transform: scale(1);
  }
  10% {
    transform: scale(1.1);
  }
  20% {
    transform: scale(1);
  }
  30% {
    transform: scale(1.15);
  }
  40% {
    transform: scale(1);
  }
}

@keyframes vesselPulse {
  0%, 100% {
    stroke-width: 3;
    opacity: 0.8;
  }
  50% {
    stroke-width: 4;
    opacity: 1;
  }
}

@keyframes bloodFlowAcross {
  0% {
    transform: translateX(-10px);
    opacity: 0;
  }
  50% {
    opacity: 1;
  }
  100% {
    transform: translateX(10px);
    opacity: 0;
  }
}

@keyframes pulseExpand {
  0% {
    transform: scale(0.9);
    opacity: 0.6;
  }
  100% {
    transform: scale(1.3);
    opacity: 0;
  }
}

@keyframes runeGlow {
  0%, 100% {
    opacity: 0.4;
    transform: scale(1);
  }
  50% {
    opacity: 1;
    transform: scale(1.3);
  }
}
</style>
