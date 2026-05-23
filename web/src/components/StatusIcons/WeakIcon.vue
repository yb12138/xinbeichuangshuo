<template>
  <svg viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg" class="status-icon weak-icon">
    <defs>
      <radialGradient id="weakGradient">
        <stop offset="0%" style="stop-color:#c084fc;stop-opacity:1" />
        <stop offset="100%" style="stop-color:#7c3aed;stop-opacity:1" />
      </radialGradient>
      <filter id="weakGlow">
        <feGaussianBlur stdDeviation="3" result="coloredBlur"/>
        <feMerge>
          <feMergeNode in="coloredBlur"/>
          <feMergeNode in="SourceGraphic"/>
        </feMerge>
      </filter>
    </defs>

    <circle cx="50" cy="50" r="40" fill="url(#weakGradient)" filter="url(#weakGlow)" class="weak-bg"/>
    <text x="50" y="62" text-anchor="middle" font-size="48" font-weight="900" fill="#fff" class="weak-char">虚</text>
  </svg>
</template>

<style scoped>
.status-icon { width: 100%; height: 100%; }
.weak-bg { animation: weakPulse 2s ease-in-out infinite; }
.weak-char { animation: weakCharPulse 2s ease-in-out infinite; }

@keyframes weakPulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.75; }
}
@keyframes weakCharPulse {
  0%, 100% { opacity: 1; transform: scale(1); }
  50% { opacity: 0.7; transform: scale(0.92); }
}
</style>
