import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { AvailableSkill, Prompt } from '../types/game'

export const useInterruptStore = defineStore('interrupt', () => {
  const currentPrompt = ref<Prompt | null>(null)
  const waitingFor = ref('')
  const selectedCards = ref<number[]>([])
  const selectedTargets = ref<string[]>([])
  const promptCounterTarget = ref('')
  const errorMessage = ref('')
  const skillEffectToast = ref('')
  let errorTimer: ReturnType<typeof setTimeout> | null = null
  let skillEffectToastTimer: ReturnType<typeof setTimeout> | null = null
  const actionMode = ref<'none' | 'attack' | 'magic'>('none')
  const magicSubChoice = ref<'none' | 'card' | 'skill'>('none')
  const selectedCardForAction = ref<number | null>(null)
  const skillMode = ref<'none' | 'choosing_skill' | 'choosing_discard' | 'choosing_target'>('none')
  const selectedSkill = ref<AvailableSkill | null>(null)
  const skillTargetIds = ref<string[]>([])
  const skillDiscardIndices = ref<number[]>([])

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
    selectedCards.value = []
    selectedTargets.value = []
    promptCounterTarget.value = ''
  }

  function clearActionState() {
    actionMode.value = 'none'
    magicSubChoice.value = 'none'
    selectedCardForAction.value = null
    skillMode.value = 'none'
    selectedSkill.value = null
    skillTargetIds.value = []
    skillDiscardIndices.value = []
  }

  function setActionMode(mode: 'none' | 'attack' | 'magic') {
    actionMode.value = mode
    if (mode === 'none') {
      selectedCardForAction.value = null
    }
  }

  function setMagicSubChoice(choice: 'none' | 'card' | 'skill') {
    magicSubChoice.value = choice
  }

  function setSelectedCardForAction(cardIndex: number | null) {
    selectedCardForAction.value = cardIndex
  }

  function clearActionMode() {
    actionMode.value = 'none'
    magicSubChoice.value = 'none'
    selectedCardForAction.value = null
  }

  function setSkillMode(mode: 'none' | 'choosing_skill' | 'choosing_discard' | 'choosing_target') {
    skillMode.value = mode
    if (mode === 'none') {
      selectedSkill.value = null
      skillTargetIds.value = []
      skillDiscardIndices.value = []
    }
  }

  function setSelectedSkill(skill: AvailableSkill | null) {
    selectedSkill.value = skill
    skillTargetIds.value = []
    skillDiscardIndices.value = []
  }

  function setSkillTargetIds(targets: string[]) {
    skillTargetIds.value = [...targets]
  }

  function setSkillDiscardIndices(indices: number[]) {
    skillDiscardIndices.value = [...indices]
  }

  function toggleSkillTarget(playerId: string) {
    const idx = skillTargetIds.value.indexOf(playerId)
    if (idx === -1) {
      skillTargetIds.value = [...skillTargetIds.value, playerId]
      return
    }
    skillTargetIds.value = skillTargetIds.value.filter(id => id !== playerId)
  }

  function toggleSkillDiscard(cardIndex: number) {
    const idx = skillDiscardIndices.value.indexOf(cardIndex)
    if (idx === -1) {
      skillDiscardIndices.value = [...skillDiscardIndices.value, cardIndex]
      return
    }
    skillDiscardIndices.value = skillDiscardIndices.value.filter(i => i !== cardIndex)
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

  function setSelectedCards(cards: number[]) {
    selectedCards.value = [...cards]
  }

  function setSelectedTargets(targets: string[]) {
    selectedTargets.value = [...targets]
  }

  function syncAfterStateUpdate() {
    // 注意：不在此处清除 currentPrompt，因为 WebSocket state_update 到达时
    // 新的 prompt 可能尚未推送（或本轮无 prompt），清除会导致闪烁/丢失。
    // prompt 的生命周期由 setPrompt / reset 管理。
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
    selectedCards,
    selectedTargets,
    promptCounterTarget,
    errorMessage,
    skillEffectToast,
    actionMode,
    magicSubChoice,
    selectedCardForAction,
    skillMode,
    selectedSkill,
    skillTargetIds,
    skillDiscardIndices,
    clearSelections,
    clearActionState,
    setActionMode,
    setMagicSubChoice,
    setSelectedCardForAction,
    clearActionMode,
    setSkillMode,
    setSelectedSkill,
    setSkillTargetIds,
    setSkillDiscardIndices,
    toggleSkillTarget,
    toggleSkillDiscard,
    clearSkillMode,
    setPrompt,
    setWaiting,
    setPromptCounterTarget,
    setSelectedCards,
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
