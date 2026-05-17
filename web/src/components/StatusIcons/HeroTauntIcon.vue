<template>
  <svg viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg" class="status-icon taunt-icon">
    <defs>
      <linearGradient id="tauntBlade" x1="0%" y1="0%" x2="100%" y2="100%">
        <stop offset="0%" stop-color="#fff8e7" />
        <stop offset="28%" stop-color="#f5d060" />
        <stop offset="62%" stop-color="#d4a428" />
        <stop offset="100%" stop-color="#9a7420" />
      </linearGradient>
      <linearGradient id="tauntGuard" x1="0%" y1="0%" x2="0%" y2="100%">
        <stop offset="0%" stop-color="#ffe9a8" />
        <stop offset="100%" stop-color="#b8860b" />
      </linearGradient>
      <linearGradient id="tauntGrip" x1="0%" y1="0%" x2="100%" y2="0%">
        <stop offset="0%" stop-color="#6b4423" />
        <stop offset="50%" stop-color="#8f5a2c" />
        <stop offset="100%" stop-color="#5a3818" />
      </linearGradient>
      <radialGradient id="tauntGlow" cx="50%" cy="50%" r="50%">
        <stop offset="0%" stop-color="#ef4444" stop-opacity="0.35" />
        <stop offset="70%" stop-color="#991b1b" stop-opacity="0.12" />
        <stop offset="100%" stop-color="#450a0a" stop-opacity="0" />
      </radialGradient>
      <filter id="tauntSwordShadow">
        <feDropShadow dx="0" dy="1.5" stdDeviation="1.2" flood-color="#1a0505" flood-opacity="0.55" />
      </filter>
    </defs>

    <circle cx="50" cy="50" r="42" fill="url(#tauntGlow)" />

    <g class="swords" filter="url(#tauntSwordShadow)">
      <g class="sword sword-left" transform="translate(50 50) rotate(-42)">
        <path
          d="M0 -36 L5 -10 L3 8 L-3 8 L-5 -10 Z"
          fill="url(#tauntBlade)"
          stroke="#7a5a12"
          stroke-width="0.6"
          stroke-linejoin="round"
        />
        <path d="M0 -34 L0 -6" stroke="#fff8dc" stroke-width="1" stroke-linecap="round" opacity="0.45" />
        <path
          d="M-11 8 H11 L9 14 H-9 Z"
          fill="url(#tauntGuard)"
          stroke="#8b6914"
          stroke-width="0.5"
        />
        <rect x="-3.5" y="14" width="7" height="13" rx="1.2" fill="url(#tauntGrip)" />
        <rect x="-2" y="16" width="4" height="2" rx="0.5" fill="#c49a5c" opacity="0.55" />
        <circle cx="0" cy="30" r="4.2" fill="url(#tauntBlade)" stroke="#8b6914" stroke-width="0.5" />
        <circle cx="-1.2" cy="28.8" r="1.2" fill="#fff" opacity="0.35" />
      </g>

      <g class="sword sword-right" transform="translate(50 50) rotate(42)">
        <path
          d="M0 -36 L5 -10 L3 8 L-3 8 L-5 -10 Z"
          fill="url(#tauntBlade)"
          stroke="#7a5a12"
          stroke-width="0.6"
          stroke-linejoin="round"
        />
        <path d="M0 -34 L0 -6" stroke="#fff8dc" stroke-width="1" stroke-linecap="round" opacity="0.45" />
        <path
          d="M-11 8 H11 L9 14 H-9 Z"
          fill="url(#tauntGuard)"
          stroke="#8b6914"
          stroke-width="0.5"
        />
        <rect x="-3.5" y="14" width="7" height="13" rx="1.2" fill="url(#tauntGrip)" />
        <rect x="-2" y="16" width="4" height="2" rx="0.5" fill="#c49a5c" opacity="0.55" />
        <circle cx="0" cy="30" r="4.2" fill="url(#tauntBlade)" stroke="#8b6914" stroke-width="0.5" />
        <circle cx="-1.2" cy="28.8" r="1.2" fill="#fff" opacity="0.35" />
      </g>
    </g>

  </svg>
</template>

<style scoped>
.status-icon {
  width: 100%;
  height: 100%;
}

.swords {
  transform-origin: 50px 50px;
  animation: tauntSwordsPulse 2.2s ease-in-out infinite;
}

.sword-left {
  transform-origin: 50px 50px;
}

.sword-right {
  transform-origin: 50px 50px;
}

@keyframes tauntSwordsPulse {
  0%,
  100% {
    transform: scale(1);
  }
  50% {
    transform: scale(1.04);
  }
}
</style>
