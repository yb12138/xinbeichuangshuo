import { computed, onBeforeUnmount, onMounted, ref } from 'vue'

type MobileOrientationLockType =
  | 'any'
  | 'natural'
  | 'landscape'
  | 'portrait'
  | 'portrait-primary'
  | 'portrait-secondary'
  | 'landscape-primary'
  | 'landscape-secondary'

function readViewportSize(): { width: number; height: number } {
  const vv = window.visualViewport
  if (vv) {
    return {
      width: Math.round(vv.width),
      height: Math.round(vv.height)
    }
  }
  return {
    width: window.innerWidth,
    height: window.innerHeight
  }
}

function isMobileBrowserLike(): boolean {
  const ua = navigator.userAgent || ''
  const uaMobile = /Android|iPhone|iPad|iPod|Mobile|Windows Phone|HarmonyOS/i.test(ua)
  const coarsePointer = window.matchMedia('(pointer: coarse)').matches
  const hasTouch = navigator.maxTouchPoints > 0
  const minScreenEdge = Math.min(window.screen.width, window.screen.height)
  return (uaMobile || coarsePointer || hasTouch) && minScreenEdge <= 1200
}

export function useMobileLandscapeGuard() {
  const mobileLike = ref(false)
  const landscape = ref(true)

  const syncViewportVars = () => {
    const { width, height } = readViewportSize()
    document.documentElement.style.setProperty('--app-vw', `${width}px`)
    document.documentElement.style.setProperty('--app-vh', `${height}px`)
    landscape.value = width >= height
  }

  const syncState = () => {
    mobileLike.value = isMobileBrowserLike()
    syncViewportVars()
  }

  const requestLandscapeLock = async () => {
    if (!mobileLike.value || landscape.value) {
      return
    }
    const orientationApi = window.screen?.orientation as ScreenOrientation & {
      lock?: (orientation: MobileOrientationLockType) => Promise<void>
    }
    if (!orientationApi || typeof orientationApi.lock !== 'function') {
      return
    }
    try {
      await orientationApi.lock('landscape')
    } catch {
      // 浏览器权限或策略不允许时忽略，交给竖屏遮罩兜底。
    }
  }

  const maybeLockLandscape = () => {
    if (mobileLike.value && !landscape.value) {
      void requestLandscapeLock()
    }
  }

  const onResize = () => {
    syncState()
    maybeLockLandscape()
  }

  const onVisibilityChange = () => {
    if (document.visibilityState === 'visible') {
      syncState()
      maybeLockLandscape()
    }
  }

  onMounted(() => {
    syncState()
    maybeLockLandscape()
    window.addEventListener('resize', onResize)
    window.addEventListener('orientationchange', onResize)
    window.visualViewport?.addEventListener('resize', onResize)
    document.addEventListener('visibilitychange', onVisibilityChange)
  })

  onBeforeUnmount(() => {
    window.removeEventListener('resize', onResize)
    window.removeEventListener('orientationchange', onResize)
    window.visualViewport?.removeEventListener('resize', onResize)
    document.removeEventListener('visibilitychange', onVisibilityChange)
  })

  const needsLandscapeGuard = computed(() => mobileLike.value && !landscape.value)

  return {
    mobileLike,
    landscape,
    needsLandscapeGuard,
    requestLandscapeLock
  }
}
