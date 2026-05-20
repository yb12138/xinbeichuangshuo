import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { AvailableSkill, Prompt } from '../types/game'

export const useInterruptStore = defineStore('interrupt', () => {
  const currentPrompt = ref<Prompt | null>(null)
  const waitingFor = ref('')
  const selectedHandIndexes = ref<number[]>([])
  const selectedFieldOptionIndexes = ref<number[]>([])
  const selectedTargets = ref<string[]>([])
  const promptCounterTarget = ref('')
  const errorMessage = ref('')
  const skillEffectToast = ref('')
  let errorTimer: ReturnType<typeof setTimeout> | null = null
  let skillEffectToastTimer: ReturnType<typeof setTimeout> | null = null
  const actionMode = ref<'none' | 'attack' | 'magic'>('none')
  const magicSubChoice = ref<'none' | 'card' | 'skill'>('none')
  const selectedHandIndexForAction = ref<number | null>(null)
  const skillMode = ref<'none' | 'choosing_skill' | 'choosing_discard' | 'choosing_target'>('none')
  const selectedSkill = ref<AvailableSkill | null>(null)
  const skillTargetIds = ref<string[]>([])
  const skillDiscardHandIndexes = ref<number[]>([])

  function clearErrorTimer() {
    if (!errorTimer) return
    clearTimeout(errorTimer)
    errorTimer = null
  }

  function clearSkillEffectToastTimer() {
    if (!skillEffectToastTimer) return
    clearTimeout(skillEffectToastTimer)
    skillEffectToastTimer = null
  }

  function clearSelections() {
    selectedHandIndexes.value = []
    selectedFieldOptionIndexes.value = []
    selectedTargets.value = []
    promptCounterTarget.value = ''
  }

  function clearActionState() {
    actionMode.value = 'none'
    magicSubChoice.value = 'none'
    selectedHandIndexForAction.value = null
    skillMode.value = 'none'
    selectedSkill.value = null
    skillTargetIds.value = []
    skillDiscardHandIndexes.value = []
  }

  function setActionMode(mode: 'none' | 'attack' | 'magic') {
    actionMode.value = mode
    if (mode === 'none') {
      selectedHandIndexForAction.value = null
    }
  }

  function setMagicSubChoice(choice: 'none' | 'card' | 'skill') {
    magicSubChoice.value = choice
  }

  function setSelectedHandIndexForAction(cardIndex: number | null) {
    selectedHandIndexForAction.value = cardIndex
  }

  function clearActionMode() {
    actionMode.value = 'none'
    magicSubChoice.value = 'none'
    selectedHandIndexForAction.value = null
  }

  function setSkillMode(mode: 'none' | 'choosing_skill' | 'choosing_discard' | 'choosing_target') {
    skillMode.value = mode
    if (mode === 'none') {
      selectedSkill.value = null
      skillTargetIds.value = []
      skillDiscardHandIndexes.value = []
    }
  }

  function setSelectedSkill(skill: AvailableSkill | null) {
    selectedSkill.value = skill
    skillTargetIds.value = []
    skillDiscardHandIndexes.value = []
  }

  function setSkillTargetIds(targets: string[]) {
    skillTargetIds.value = [...targets]
  }

  function setSkillDiscardHandIndexes(indices: number[]) {
    skillDiscardHandIndexes.value = [...indices]
  }

  function toggleSkillTarget(playerId: string) {
    const idx = skillTargetIds.value.indexOf(playerId)
    if (idx === -1) {
      skillTargetIds.value = [...skillTargetIds.value, playerId]
      return
    }
    skillTargetIds.value = skillTargetIds.value.filter(id => id !== playerId)
  }

  function toggleSkillDiscardHandIndex(cardIndex: number) {
    const idx = skillDiscardHandIndexes.value.indexOf(cardIndex)
    if (idx === -1) {
      skillDiscardHandIndexes.value = [...skillDiscardHandIndexes.value, cardIndex]
      return
    }
    skillDiscardHandIndexes.value = skillDiscardHandIndexes.value.filter(i => i !== cardIndex)
  }

  function clearSkillMode() {
    setSkillMode('none')
  }

  function setPrompt(prompt: Prompt | null) {
    currentPrompt.value = prompt
    clearSelections()
    if (prompt) {
      clearActionState()
    }
  }

  function setWaiting(playerId: string) {
    waitingFor.value = playerId
  }

  function setPromptCounterTarget(playerId: string) {
    promptCounterTarget.value = playerId
  }

  function setSelectedHandIndexes(cards: number[]) {
    selectedHandIndexes.value = [...cards]
  }

  function setSelectedFieldOptionIndexes(options: number[]) {
    selectedFieldOptionIndexes.value = [...options]
  }

  function setSelectedTargets(targets: string[]) {
    selectedTargets.value = [...targets]
  }

  function syncAfterStateUpdate() {
    // state_update means the previous prompt was accepted or superseded.
    // If a new prompt is needed, the server sends RequireAction right after
    // the fresh SyncState, so clearing here prevents stale controls sticking.
    currentPrompt.value = null
    waitingFor.value = ''
    clearSelections()
    clearActionState()
  }

  function setError(message: string) {
    clearErrorTimer()
    errorMessage.value = message
  }

  function showError(message: string, durationMs = 3000) {
    setError(message)
    if (durationMs <= 0) return
    errorTimer = setTimeout(() => {
      clearError()
    }, durationMs)
  }

  function clearError() {
    clearErrorTimer()
    errorMessage.value = ''
  }

  function setSkillEffectToast(message: string) {
    clearSkillEffectToastTimer()
    skillEffectToast.value = message
  }

  function showSkillEffectToast(message: string, durationMs = 2500) {
    setSkillEffectToast(message)
    if (durationMs <= 0) return
    skillEffectToastTimer = setTimeout(() => {
      clearSkillEffectToast()
    }, durationMs)
  }

  function clearSkillEffectToast() {
    clearSkillEffectToastTimer()
    skillEffectToast.value = ''
  }

  function reset() {
    currentPrompt.value = null
    waitingFor.value = ''
    clearSelections()
    clearActionState()
    clearError()
    clearSkillEffectToast()
  }

  return {
    currentPrompt,
    waitingFor,
    selectedHandIndexes,
    selectedFieldOptionIndexes,
    selectedTargets,
    promptCounterTarget,
    errorMessage,
    skillEffectToast,
    actionMode,
    magicSubChoice,
    selectedHandIndexForAction,
    skillMode,
    selectedSkill,
    skillTargetIds,
    skillDiscardHandIndexes,
    clearSelections,
    clearActionState,
    setActionMode,
    setMagicSubChoice,
    setSelectedHandIndexForAction,
    clearActionMode,
    setSkillMode,
    setSelectedSkill,
    setSkillTargetIds,
    setSkillDiscardHandIndexes,
    toggleSkillTarget,
    toggleSkillDiscardHandIndex,
    clearSkillMode,
    setPrompt,
    setWaiting,
    setPromptCounterTarget,
    setSelectedHandIndexes,
    setSelectedFieldOptionIndexes,
    setSelectedTargets,
    syncAfterStateUpdate,
    setError,
    showError,
    clearError,
    setSkillEffectToast,
    showSkillEffectToast,
    clearSkillEffectToast,
    reset,
  }
})
