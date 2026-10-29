<script setup lang="ts">
// 元素徽章的小图标：每个元素一个独立的极简字形（24x24，使用 currentColor）。
defineProps<{
  element: string
}>()
</script>

<template>
  <svg class="element-glyph" viewBox="0 0 24 24" aria-hidden="true" focusable="false">
    <!-- 火：火焰 -->
    <path
      v-if="element === 'Fire'"
      fill="currentColor"
      d="M12 2c1 3.5 4.6 5.6 4.6 10.2a4.6 4.6 0 0 1-9.2 0c0-2.1 1-3.4 2-4.4.2 1.6 1 2.7 2 3.1 0-3.1-.5-5.7.6-8.9z"
    />
    <!-- 水：水滴 -->
    <path
      v-else-if="element === 'Water'"
      fill="currentColor"
      d="M12 3c3 4 6 7.2 6 10.6a6 6 0 0 1-12 0C6 10.2 9 7 12 3z"
    />
    <!-- 风：气流 -->
    <g
      v-else-if="element === 'Wind'"
      fill="none"
      stroke="currentColor"
      stroke-width="2.2"
      stroke-linecap="round"
    >
      <path d="M3.5 9H14a3 3 0 1 0-3-3" />
      <path d="M3 13.5h15a3 3 0 1 1-3 3" />
      <path d="M5 18h5" />
    </g>
    <!-- 地：山岩 -->
    <path
      v-else-if="element === 'Earth'"
      fill="currentColor"
      d="M2 19.5 9 7l3.6 5.4L15.2 9l6.8 10.5z"
    />
    <!-- 雷：闪电 -->
    <path
      v-else-if="element === 'Thunder'"
      fill="currentColor"
      d="M13.5 2 5 13.6h5.6L9 22l9.4-12.3h-5.7z"
    />
    <!-- 光：日轮 -->
    <g v-else-if="element === 'Light'" stroke="currentColor" stroke-width="2" stroke-linecap="round">
      <circle cx="12" cy="12" r="4" fill="currentColor" />
      <path d="M12 2.5v2.6M12 18.9v2.6M2.5 12h2.6M18.9 12h2.6M5.3 5.3l1.8 1.8M16.9 16.9l1.8 1.8M5.3 18.7l1.8-1.8M16.9 7.1l1.8-1.8" />
    </g>
    <!-- 暗：新月 -->
    <path
      v-else-if="element === 'Dark'"
      fill="currentColor"
      d="M15 3a8.6 8.6 0 1 0 6 13.6A7 7 0 0 1 15 3z"
    />
    <!-- 未知元素：菱形 -->
    <path v-else fill="currentColor" d="M12 4 19 12l-7 8-7-8z" />
  </svg>
</template>

<style scoped>
.element-glyph {
  width: 100%;
  height: 100%;
  display: block;
}
</style>
