<script setup lang="ts">
import { useAudioStore } from '../stores/audio.store'

const audioStore = useAudioStore()
</script>

<template>
  <button
    type="button"
    class="sound-toggle"
    :class="{ 'sound-toggle--muted': audioStore.muted }"
    :aria-pressed="!audioStore.muted"
    :aria-label="audioStore.muted ? '开启音效' : '关闭音效'"
    :title="audioStore.muted ? '开启音效' : '关闭音效'"
    @click="audioStore.toggleMuted()"
  >
    <svg viewBox="0 0 16 16" aria-hidden="true" focusable="false">
      <path d="M2.5 6h2.2L8 3.2v9.6L4.7 10H2.5z" fill="currentColor" />
      <template v-if="!audioStore.muted">
        <path d="M10.2 5.6a3.2 3.2 0 0 1 0 4.8" />
        <path d="M11.9 3.9a5.6 5.6 0 0 1 0 8.2" />
      </template>
      <path v-else d="M10.4 5.8l4 4.4M14.4 5.8l-4 4.4" />
    </svg>
  </button>
</template>

<style scoped>
.sound-toggle {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 30px;
  height: 30px;
  border-radius: 999px;
  border: 1px solid rgba(200, 210, 230, 0.4);
  background: rgba(16, 20, 32, 0.78);
  color: #e6ecf8;
  box-shadow: 0 6px 14px rgba(0, 0, 0, 0.35);
  cursor: pointer;
}

.sound-toggle:hover {
  filter: brightness(1.15);
}

.sound-toggle:focus-visible {
  outline: 2px solid #9cc4ff;
  outline-offset: 2px;
}

.sound-toggle--muted {
  color: #8d95a8;
}

.sound-toggle svg {
  width: 16px;
  height: 16px;
  fill: none;
  stroke: currentColor;
  stroke-width: 1.3;
  stroke-linecap: round;
}
</style>
