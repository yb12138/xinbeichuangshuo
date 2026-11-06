<script setup lang="ts">
import { computed } from 'vue'
import { useSessionStore } from './stores/session.store'
import RoomLobby from './components/RoomLobby.vue'
import GameBoard from './components/GameBoard.vue'
import { useMobileLandscapeGuard } from './composables/useMobileLandscapeGuard'
import { useStageScale } from './composables/useStageScale'

const sessionStore = useSessionStore()

const showGame = computed(() => sessionStore.gameStarted)
const { needsLandscapeGuard, requestLandscapeLock } = useMobileLandscapeGuard()
// 对局按逻辑舞台整体缩放；大厅自带响应式布局，只在大屏放大
const { stageStyle } = useStageScale(computed(() => (showGame.value ? 'full' : 'upscale-only')))
</script>

<template>
  <div class="app-root w-full h-full">
    <div class="app-stage" :style="stageStyle">
      <!-- 游戏界面 -->
      <GameBoard v-if="showGame" />

      <!-- 房间大厅 -->
      <RoomLobby v-else />

      <!-- 弹层传送目标：放在舞台内才能跟着一起缩放 -->
      <div id="stage-portal"></div>
    </div>

    <div v-if="needsLandscapeGuard" class="mobile-orientation-guard">
      <div class="mobile-orientation-card">
        <div class="mobile-orientation-title">请横屏体验</div>
        <div class="mobile-orientation-text">
          检测到当前为手机竖屏。为保证战场布局完整，请将手机旋转为横屏后继续。
        </div>
        <button class="mobile-orientation-btn" @click="requestLandscapeLock">
          尝试切换横屏
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.app-root {
  min-height: var(--app-vh);
  height: var(--app-vh);
  width: var(--app-vw);
  position: relative;
  overflow: hidden;
}

.app-stage {
  position: absolute;
  left: 0;
  top: 0;
  transform-origin: 0 0;
  container: stage / size;
  overflow: hidden;
}

/* 舞台带 transform，fixed 弹层以舞台为包含块。
   这一层本身不占位、不拦截指针，弹层各自的 pointer-events 照旧生效。 */
#stage-portal {
  position: absolute;
  left: 0;
  top: 0;
  width: 0;
  height: 0;
  z-index: 10000;
}

.mobile-orientation-guard {
  position: fixed;
  inset: 0;
  z-index: 10000;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: calc(16px + var(--safe-top)) calc(16px + var(--safe-right)) calc(16px + var(--safe-bottom)) calc(16px + var(--safe-left));
  background:
    radial-gradient(420px 260px at 50% 40%, rgba(78, 133, 183, 0.28), transparent 72%),
    rgba(6, 12, 21, 0.92);
  backdrop-filter: blur(8px);
}

.mobile-orientation-card {
  width: min(92vw, 420px);
  border-radius: 16px;
  border: 1px solid rgba(148, 183, 208, 0.45);
  background: linear-gradient(180deg, rgba(15, 28, 43, 0.96), rgba(9, 19, 31, 0.96));
  box-shadow:
    inset 0 1px 0 rgba(233, 244, 255, 0.15),
    0 18px 32px rgba(2, 8, 18, 0.56);
  padding: 18px 16px;
  text-align: center;
}

.mobile-orientation-title {
  font-size: 22px;
  font-weight: 700;
  color: #ebf7ff;
  letter-spacing: 0.04em;
}

.mobile-orientation-text {
  margin-top: 8px;
  font-size: 14px;
  line-height: 1.55;
  color: rgba(216, 230, 243, 0.9);
}

.mobile-orientation-btn {
  margin-top: 14px;
  width: 100%;
  height: 42px;
  border-radius: 10px;
  border: 1px solid rgba(125, 169, 201, 0.62);
  background: linear-gradient(130deg, rgba(66, 116, 162, 0.94), rgba(47, 89, 131, 0.96));
  color: #e8f5ff;
  font-size: 14px;
  font-weight: 600;
}
</style>
