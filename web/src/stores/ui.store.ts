import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { GameEndSnapshot, SkillModalAnchor } from './storeTypes'

export const useUiStore = defineStore('ui', () => {
  const skillModalCharacterId = ref<string | null>(null)
  const skillModalAnchor = ref<SkillModalAnchor | null>(null)
  const isGameEnded = ref(false)
  const gameEndMessage = ref('')
  const gameEndSnapshot = ref<GameEndSnapshot | null>(null)
  const cinematicMode = ref(true)

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
  }

  return {
    skillModalCharacterId,
    skillModalAnchor,
    isGameEnded,
    gameEndMessage,
    gameEndSnapshot,
    cinematicMode,
    openSkillModal,
    setGameEnded,
    setCinematicMode,
    reset,
  }
})
