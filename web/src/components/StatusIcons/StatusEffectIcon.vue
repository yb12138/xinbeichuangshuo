<script setup lang="ts">
// 场上状态图标：统一由 BattleMarkIcon 绘制（见 constants/battleMarks 的 STATUS_MARKS）。
import BattleMarkIcon from '../marks/BattleMarkIcon.vue'

defineProps<{
  effect: string
  count?: number
}>()
</script>

<template>
  <div class="status-effect-icon">
    <BattleMarkIcon kind="status" :mark-key="effect" :count="count" />
  </div>
</template>

<style scoped>
.status-effect-icon {
  position: relative;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
}
</style>
