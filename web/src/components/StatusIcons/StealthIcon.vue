<template>
  <svg viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg" class="status-icon stealth-icon">
    <defs>
      <linearGradient id="stealthGradient" x1="0%" y1="0%" x2="100%" y2="100%">
        <stop offset="0%" style="stop-color:#6b7280;stop-opacity:1" />
        <stop offset="100%" style="stop-color:#374151;stop-opacity:1" />
      </linearGradient>
    </defs>

    <!-- 眼睛符号 - 闭眼/隐身 -->
    <ellipse cx="50" cy="50" rx="35" ry="20" fill="none" stroke="url(#stealthGradient)" stroke-width="4" class="eye-outline"/>

    <!-- 闭眼的线条 -->
    <path d="M 15 50 Q 50 60, 85 50" stroke="#6b7280" stroke-width="5" fill="none" stroke-linecap="round" class="eye-closed"/>

    <!-- 睫毛效果 -->
    <g class="eyelashes" stroke="#6b7280" stroke-width="2" stroke-linecap="round">
      <line x1="20" y1="48" x2="18" y2="42" class="lash lash-1"/>
      <line x1="35" y1="52" x2="33" y2="46" class="lash lash-2"/>
      <line x1="50" y1="54" x2="50" y2="48" class="lash lash-3"/>
      <line x1="65" y1="52" x2="67" y2="46" class="lash lash-4"/>
      <line x1="80" y1="48" x2="82" y2="42" class="lash lash-5"/>
    </g>

    <!-- 隐身波纹 -->
    <g class="stealth-waves" fill="none" stroke="#6b7280" stroke-width="2" opacity="0.4">
      <ellipse cx="50" cy="50" rx="40" ry="25" class="wave wave-1"/>
      <ellipse cx="50" cy="50" rx="45" ry="30" class="wave wave-2"/>
    </g>

    <!-- 禁止符号斜线 -->
    <line x1="20" y1="25" x2="80" y2="75" stroke="#6b7280" stroke-width="4" stroke-linecap="round" opacity="0.6" class="slash"/>
  </svg>
</template>

<style scoped>
.status-icon {
  width: 100%;
  height: 100%;
}

.eye-outline {
  animation: stealthBlink 4s ease-in-out infinite;
}

.eye-closed {
  animation: stealthClosedPulse 3s ease-in-out infinite;
}

.lash {
  animation: stealthLashFade 2s ease-in-out infinite;
}

.lash-1 { animation-delay: 0s; }
.lash-2 { animation-delay: 0.2s; }
.lash-3 { animation-delay: 0.4s; }
.lash-4 { animation-delay: 0.6s; }
.lash-5 { animation-delay: 0.8s; }

.wave {
  animation: stealthWaveExpand 3s ease-out infinite;
}

.wave-1 { animation-delay: 0s; }
.wave-2 { animation-delay: 1.5s; }

.slash {
  animation: stealthSlashPulse 2s ease-in-out infinite;
}

@keyframes stealthBlink {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
}

@keyframes stealthClosedPulse {
  0%, 100% {
    stroke-width: 5;
    opacity: 1;
  }
  50% {
    stroke-width: 6;
    opacity: 0.8;
  }
}

@keyframes stealthLashFade {
  0%, 100% {
    opacity: 0.6;
  }
  50% {
    opacity: 0.3;
  }
}

@keyframes stealthWaveExpand {
  0% {
    transform: scale(0.9);
    opacity: 0.6;
  }
  100% {
    transform: scale(1.1);
    opacity: 0;
  }
}

@keyframes stealthSlashPulse {
  0%, 100% {
    opacity: 0.6;
  }
  50% {
    opacity: 0.3;
  }
}
</style>
