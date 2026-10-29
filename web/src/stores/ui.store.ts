import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { GameEndSnapshot, SkillModalAnchor } from './storeTypes'

const DOCK_OFFSET_KEY = 'starcup.decisionDock.offset'

/** 浏览器可能禁用 localStorage（隐私模式、测试环境），统一在这里探测。 */
function dockStorage(): Storage | null {
  try {
    const storage = typeof window !== 'undefined' ? window.localStorage : undefined
    return storage && typeof storage.getItem === 'function' ? storage : null
  } catch {
    return null
  }
}

function loadDockOffset(): { x: number; y: number } {
  const storage = dockStorage()
  if (!storage) return { x: 0, y: 0 }
  try {
    const raw = storage.getItem(DOCK_OFFSET_KEY)
    const parsed = raw ? JSON.parse(raw) : null
    if (parsed && Number.isFinite(parsed.x) && Number.isFinite(parsed.y)) return { x: parsed.x, y: parsed.y }
  } catch {
    // 数据损坏时回到默认位置
  }
  return { x: 0, y: 0 }
}

export const useUiStore = defineStore('ui', () => {
  const skillModalCharacterId = ref<string | null>(null)
  const skillModalAnchor = ref<SkillModalAnchor | null>(null)
  const isGameEnded = ref(false)
  const gameEndMessage = ref('')
  const gameEndSnapshot = ref<GameEndSnapshot | null>(null)
  const cinematicMode = ref(true)
  // 新版座位连线舞台；URL 带 ?stage=legacy 可切回旧的中央叙事
  const stageOverlayEnabled = ref(
    typeof window === 'undefined' || new URLSearchParams(window.location.search).get('stage') !== 'legacy'
  )

  // 决策面板：收起状态与拖拽位置（位置跨对局记住，收起状态按每次提示重置）
  const decisionDockCollapsed = ref(false)
  const decisionDockOffset = ref<{ x: number; y: number }>(loadDockOffset())

  function setDecisionDockCollapsed(collapsed: boolean) {
    decisionDockCollapsed.value = collapsed
  }

  function setDecisionDockOffset(offset: { x: number; y: number }) {
    decisionDockOffset.value = offset
    try {
      dockStorage()?.setItem(DOCK_OFFSET_KEY, JSON.stringify(offset))
    } catch {
      // 写入失败时仅在本次会话内记住位置
    }
  }

  function openSkillModal(characterId: string | null, anchor?: SkillModalAnchor | null) {
    skillModalCharacterId.value = characterId
    skillModalAnchor.value = characterId ? (anchor ?? null) : null
  }

  function setGameEnded(ended: boolean, message = '', snapshot: GameEndSnapshot | null = null) {
    isGameEnded.value = ended
    gameEndMessage.value = message
    gameEndSnapshot.value = snapshot
  }

  function setCinematicMode(enabled: boolean) {
    cinematicMode.value = enabled
  }

  function reset() {
    skillModalCharacterId.value = null
    skillModalAnchor.value = null
    isGameEnded.value = false
    gameEndMessage.value = ''
    gameEndSnapshot.value = null
    cinematicMode.value = true
    decisionDockCollapsed.value = false
  }

  return {
    skillModalCharacterId,
    skillModalAnchor,
    isGameEnded,
    gameEndMessage,
    gameEndSnapshot,
    cinematicMode,
    stageOverlayEnabled,
    decisionDockCollapsed,
    decisionDockOffset,
    setDecisionDockCollapsed,
    setDecisionDockOffset,
    openSkillModal,
    setGameEnded,
    setCinematicMode,
    reset,
  }
})
