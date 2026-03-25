import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { AvailableSkill, CharacterView, GameStateUpdate, PlayerView } from '../types/game'

export const useSnapshotStore = defineStore('snapshot', () => {
  const turnStage = ref('')
  const combatStage = ref('')
  const subflow = ref('')
  const currentPlayer = ref('')
  const hasPerformedStartup = ref(false)
  const players = ref<Record<string, PlayerView>>({})
  const redMorale = ref(15)
  const blueMorale = ref(15)
  const redCups = ref(0)
  const blueCups = ref(0)
  const redGems = ref(0)
  const blueGems = ref(0)
  const redCrystals = ref(0)
  const blueCrystals = ref(0)
  const deckCount = ref(0)
  const discardCount = ref(0)
  const availableSkills = ref<AvailableSkill[]>([])
  const characters = ref<Record<string, CharacterView>>({})

  function setCharacters(list: CharacterView[]) {
    const map: Record<string, CharacterView> = {}
    for (const character of list) {
      map[character.id] = character
    }
    characters.value = map
  }

  function updateGameState(state: GameStateUpdate) {
    turnStage.value = state.turn_stage ?? ''
    combatStage.value = state.combat_stage ?? ''
    subflow.value = state.subflow ?? ''
    currentPlayer.value = state.current_player
    hasPerformedStartup.value = state.has_performed_startup ?? false
    players.value = state.players
    redMorale.value = state.red_morale
    blueMorale.value = state.blue_morale
    redCups.value = state.red_cups
    blueCups.value = state.blue_cups
    redGems.value = state.red_gems
    blueGems.value = state.blue_gems
    redCrystals.value = state.red_crystals
    blueCrystals.value = state.blue_crystals
    deckCount.value = state.deck_count
    discardCount.value = state.discard_count ?? 0
    availableSkills.value = state.available_skills ?? []
    if (state.characters?.length) {
      setCharacters(state.characters)
    }
  }

  function reset() {
    turnStage.value = ''
    combatStage.value = ''
    subflow.value = ''
    currentPlayer.value = ''
    hasPerformedStartup.value = false
    players.value = {}
    redMorale.value = 15
    blueMorale.value = 15
    redCups.value = 0
    blueCups.value = 0
    redGems.value = 0
    blueGems.value = 0
    redCrystals.value = 0
    blueCrystals.value = 0
    deckCount.value = 0
    discardCount.value = 0
    availableSkills.value = []
    characters.value = {}
  }

  return {
    turnStage,
    combatStage,
    subflow,
    currentPlayer,
    hasPerformedStartup,
    players,
    redMorale,
    blueMorale,
    redCups,
    blueCups,
    redGems,
    blueGems,
    redCrystals,
    blueCrystals,
    deckCount,
    discardCount,
    availableSkills,
    characters,
    setCharacters,
    updateGameState,
    reset,
  }
})
