<template>
  <svg viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg" class="status-icon power-blessing-icon">
    <defs>
      <radialGradient id="powerGradient">
        <stop offset="0%" style="stop-color:#fdba74;stop-opacity:1" />
        <stop offset="100%" style="stop-color:#ea580c;stop-opacity:1" />
      </radialGradient>
      <filter id="powerGlow">
        <feGaussianBlur stdDeviation="3" result="coloredBlur"/>
        <feMerge>
          <feMergeNode in="coloredBlur"/>
          <feMergeNode in="SourceGraphic"/>
        </feMerge>
      </filter>
    </defs>

    <circle cx="50" cy="50" r="40" fill="url(#powerGradient)" filter="url(#powerGlow)" class="power-bg"/>
    <text x="50" y="62" text-anchor="middle" font-size="48" font-weight="900" fill="#fff" class="power-char">威</text>
  </svg>
</template>

<style scoped>
.status-icon { width: 100%; height: 100%; }
.power-bg { animation: powerPulse 2s ease-in-out infinite; }
.power-char { animation: powerCharGlow 1.5s ease-in-out infinite; }

@keyframes powerPulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.85; }
}
@keyframes powerCharGlow {
  0%, 100% { opacity: 1; filter: brightness(1); }
  50% { opacity: 0.85; filter: brightness(1.3); }
}
</style>
