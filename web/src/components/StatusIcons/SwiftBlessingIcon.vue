<template>
  <svg viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg" class="status-icon swift-blessing-icon">
    <defs>
      <radialGradient id="swiftGradient">
        <stop offset="0%" style="stop-color:#67e8f9;stop-opacity:1" />
        <stop offset="100%" style="stop-color:#0891b2;stop-opacity:1" />
      </radialGradient>
      <filter id="swiftGlow">
        <feGaussianBlur stdDeviation="3" result="coloredBlur"/>
        <feMerge>
          <feMergeNode in="coloredBlur"/>
          <feMergeNode in="SourceGraphic"/>
        </feMerge>
      </filter>
    </defs>

    <circle cx="50" cy="50" r="40" fill="url(#swiftGradient)" filter="url(#swiftGlow)" class="swift-bg"/>
    <text x="50" y="62" text-anchor="middle" font-size="48" font-weight="900" fill="#fff" class="swift-char">迅</text>
  </svg>
</template>

<style scoped>
.status-icon { width: 100%; height: 100%; }
.swift-bg { animation: swiftPulse 2s ease-in-out infinite; }
.swift-char { animation: swiftCharShift 1.5s ease-in-out infinite; }

@keyframes swiftPulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.85; }
}
@keyframes swiftCharShift {
  0%, 100% { opacity: 1; transform: translateX(0); }
  25% { transform: translateX(-2px); }
  75% { transform: translateX(2px); }
}
</style>
