<template>
  <svg viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg" class="status-icon eternal-movement-icon">
    <defs>
      <linearGradient id="musicGradient" x1="0%" y1="0%" x2="100%" y2="100%">
        <stop offset="0%" style="stop-color:#a78bfa;stop-opacity:1" />
        <stop offset="100%" style="stop-color:#7c3aed;stop-opacity:1" />
      </linearGradient>
      <filter id="musicGlow">
        <feGaussianBlur stdDeviation="2" result="coloredBlur"/>
        <feMerge>
          <feMergeNode in="coloredBlur"/>
          <feMergeNode in="SourceGraphic"/>
        </feMerge>
      </filter>
    </defs>

    <!-- 五线谱 -->
    <g class="staff-lines" stroke="#a78bfa" stroke-width="1.5" opacity="0.6">
      <line x1="10" y1="35" x2="90" y2="35" class="staff-line line-1"/>
      <line x1="10" y1="42" x2="90" y2="42" class="staff-line line-2"/>
      <line x1="10" y1="49" x2="90" y2="49" class="staff-line line-3"/>
      <line x1="10" y1="56" x2="90" y2="56" class="staff-line line-4"/>
      <line x1="10" y1="63" x2="90" y2="63" class="staff-line line-5"/>
    </g>

    <!-- 音符 -->
    <g class="notes" fill="url(#musicGradient)" filter="url(#musicGlow)">
      <!-- 第一个音符 -->
      <g class="note note-1">
        <ellipse cx="25" cy="52" rx="4" ry="3" transform="rotate(-20 25 52)"/>
        <rect x="28" y="35" width="2" height="17"/>
        <path d="M 30 35 Q 35 32, 35 38" fill="none" stroke="url(#musicGradient)" stroke-width="2"/>
      </g>

      <!-- 第二个音符 -->
      <g class="note note-2">
        <ellipse cx="45" cy="45" rx="4" ry="3" transform="rotate(-20 45 45)"/>
        <rect x="48" y="28" width="2" height="17"/>
        <path d="M 50 28 Q 55 25, 55 31" fill="none" stroke="url(#musicGradient)" stroke-width="2"/>
      </g>

      <!-- 第三个音符 -->
      <g class="note note-3">
        <ellipse cx="65" cy="56" rx="4" ry="3" transform="rotate(-20 65 56)"/>
        <rect x="68" y="39" width="2" height="17"/>
        <path d="M 70 39 Q 75 36, 75 42" fill="none" stroke="url(#musicGradient)" stroke-width="2"/>
      </g>
    </g>

    <!-- 高音谱号 -->
    <g class="treble-clef" fill="url(#musicGradient)" opacity="0.8">
      <path d="M 15 45 Q 12 40, 15 35 Q 18 30, 20 35 Q 22 40, 20 45 Q 18 50, 15 50 Q 12 50, 12 47 Q 12 44, 15 44"
            stroke="url(#musicGradient)" stroke-width="1.5" fill="none" class="clef-curve"/>
      <circle cx="15" cy="56" r="2" class="clef-dot"/>
    </g>

    <!-- 音波 -->
    <g class="sound-waves" fill="none" stroke="#a78bfa" stroke-width="2" opacity="0.4">
      <path d="M 80 30 Q 85 35, 80 40" class="wave wave-1"/>
      <path d="M 85 28 Q 92 35, 85 42" class="wave wave-2"/>
      <path d="M 80 60 Q 85 65, 80 70" class="wave wave-3"/>
      <path d="M 85 58 Q 92 65, 85 72" class="wave wave-4"/>
    </g>

    <!-- 永恒循环符号 -->
    <g class="infinity" stroke="#a78bfa" stroke-width="2.5" fill="none" opacity="0.5">
      <path d="M 30 75 Q 40 70, 50 75 Q 60 80, 70 75 Q 60 70, 50 75 Q 40 80, 30 75 Z" class="infinity-loop"/>
    </g>

    <!-- 音乐粒子 -->
    <g class="music-particles">
      <circle cx="35" cy="25" r="1.5" fill="#c4b5fd" class="particle particle-1"/>
      <circle cx="55" cy="20" r="1.5" fill="#c4b5fd" class="particle particle-2"/>
      <circle cx="75" cy="25" r="1.5" fill="#c4b5fd" class="particle particle-3"/>
      <circle cx="40" cy="70" r="1.5" fill="#c4b5fd" class="particle particle-4"/>
      <circle cx="60" cy="75" r="1.5" fill="#c4b5fd" class="particle particle-5"/>
    </g>
  </svg>
</template>

<style scoped>
.status-icon {
  width: 100%;
  height: 100%;
}

.staff-line {
  animation: staffShimmer 3s ease-in-out infinite;
}

.line-1 { animation-delay: 0s; }
.line-2 { animation-delay: 0.2s; }
.line-3 { animation-delay: 0.4s; }
.line-4 { animation-delay: 0.6s; }
.line-5 { animation-delay: 0.8s; }

.note {
  animation: noteFloat 2s ease-in-out infinite;
}

.note-1 { animation-delay: 0s; }
.note-2 { animation-delay: 0.6s; }
.note-3 { animation-delay: 1.2s; }

.treble-clef {
  animation: clefPulse 3s ease-in-out infinite;
}

.wave {
  animation: soundWaveExpand 2s ease-in-out infinite;
}

.wave-1 { animation-delay: 0s; }
.wave-2 { animation-delay: 0.5s; }
.wave-3 { animation-delay: 1s; }
.wave-4 { animation-delay: 1.5s; }

.infinity-loop {
  animation: infinityFlow 4s linear infinite;
}

.particle {
  animation: particleSparkle 2s ease-in-out infinite;
}

.particle-1 { animation-delay: 0s; }
.particle-2 { animation-delay: 0.4s; }
.particle-3 { animation-delay: 0.8s; }
.particle-4 { animation-delay: 1.2s; }
.particle-5 { animation-delay: 1.6s; }

@keyframes staffShimmer {
  0%, 100% {
    opacity: 0.4;
  }
  50% {
    opacity: 0.8;
  }
}

@keyframes noteFloat {
  0%, 100% {
    transform: translateY(0);
  }
  50% {
    transform: translateY(-5px);
  }
}

@keyframes clefPulse {
  0%, 100% {
    opacity: 0.8;
    transform: scale(1);
  }
  50% {
    opacity: 1;
    transform: scale(1.05);
  }
}

@keyframes soundWaveExpand {
  0% {
    opacity: 0;
    transform: scaleX(0.8);
  }
  50% {
    opacity: 0.6;
  }
  100% {
    opacity: 0;
    transform: scaleX(1.2);
  }
}

@keyframes infinityFlow {
  0% {
    stroke-dasharray: 0 200;
    stroke-dashoffset: 0;
  }
  100% {
    stroke-dasharray: 100 100;
    stroke-dashoffset: -200;
  }
}

@keyframes particleSparkle {
  0%, 100% {
    opacity: 0.3;
    transform: scale(1);
  }
  50% {
    opacity: 1;
    transform: scale(1.5);
  }
}
</style>
