import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { AvailableSkill, Prompt } from '../types/game'

type SkillMode = 'none' | 'choosing_skill' | 'choosing_exclusive' | 'choosing_discard' | 'choosing_target'

/** 提交前的选择快照：后端报错时恢复，玩家不用从头再来。 */
type SubmissionSnapshot = {
  actionMode: 'none' | 'attack' | 'magic'
  magicSubChoice: 'none' | 'card' | 'skill'
  selectedHandIndexForAction: number | null
  skillMode: SkillMode
  selectedSkill: AvailableSkill | null
  skillTargetIds: string[]
  skillDiscardHandIndexes: number[]
  selectedHandIndexes: number[]
  selectedTargets: string[]
  promptCounterTarget: string
  counterFlowStage: 'none' | 'card' | 'target'
}

/** 发出后迟迟没有回包（断线等）时，自动解除"提交中"，避免界面一直卡住。 */
export const SUBMIT_RESPONSE_TIMEOUT_MS = 12000

export const useInterruptStore = defineStore('interrupt', () => {
  const currentPrompt = ref<Prompt | null>(null)
  const waitingFor = ref('')
  const selectedHandIndexes = ref<number[]>([])
  const selectedFieldOptionIndexes = ref<number[]>([])
  const selectedTargets = ref<string[]>([])
  const promptCounterTarget = ref('')
  // 应战两步流程：none = 还在选响应方式；card = ① 选应战牌；target = ② 选应战目标
  const counterFlowStage = ref<'none' | 'card' | 'target'>('none')
  const errorMessage = ref('')
  const skillEffectToast = ref('')
  let errorTimer: ReturnType<typeof setTimeout> | null = null
  let skillEffectToastTimer: ReturnType<typeof setTimeout> | null = null
  const actionMode = ref<'none' | 'attack' | 'magic'>('none')
  const magicSubChoice = ref<'none' | 'card' | 'skill'>('none')
  const selectedHandIndexForAction = ref<number | null>(null)
  const skillMode = ref<SkillMode>('none')
  const selectedSkill = ref<AvailableSkill | null>(null)
  const skillTargetIds = ref<string[]>([])
  const skillDiscardHandIndexes = ref<number[]>([])
  // 已发出操作、还没等到 state_update / prompt / error 回包；期间不重新显示行动枢纽
  const isSubmitting = ref(false)
  // 我的提示被战斗链播放压住（gameplayMessageHandlers 的 prompt gate）；期间同样压住行动枢纽
  const promptGated = ref(false)
  let pendingSubmission: SubmissionSnapshot | null = null
  let submitTimer: ReturnType<typeof setTimeout> | null = null

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
    counterFlowStage.value = 'none'
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

  function setSkillMode(mode: SkillMode) {
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
    // 提交中收到新提示 / 提示被收回：说明上一步已被服务器受理，残留的本地选择一并作废
    const acknowledged = isSubmitting.value
    currentPrompt.value = prompt
    clearSelections()
    if (prompt || acknowledged) {
      clearActionState()
    }
    endSubmit()
  }

  function setWaiting(playerId: string) {
    waitingFor.value = playerId
  }

  function setCounterFlowStage(stage: 'none' | 'card' | 'target') {
    counterFlowStage.value = stage
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

  function clearSubmitTimer() {
    if (!submitTimer) return
    clearTimeout(submitTimer)
    submitTimer = null
  }

  /** 发出操作前调用：记下当前选择并进入"提交中"。 */
  function beginSubmit() {
    pendingSubmission = {
      actionMode: actionMode.value,
      magicSubChoice: magicSubChoice.value,
      selectedHandIndexForAction: selectedHandIndexForAction.value,
      skillMode: skillMode.value,
      selectedSkill: selectedSkill.value,
      skillTargetIds: [...skillTargetIds.value],
      skillDiscardHandIndexes: [...skillDiscardHandIndexes.value],
      selectedHandIndexes: [...selectedHandIndexes.value],
      selectedTargets: [...selectedTargets.value],
      promptCounterTarget: promptCounterTarget.value,
      counterFlowStage: counterFlowStage.value,
    }
    isSubmitting.value = true
    clearSubmitTimer()
    // 兜底：迟迟没有回包（断线、服务器静默受理）时只解除"提交中"，不改动当前选择
    submitTimer = setTimeout(() => {
      submitTimer = null
      endSubmit()
    }, SUBMIT_RESPONSE_TIMEOUT_MS)
  }

  /** 收到回包（state_update / prompt / waiting）：提交已被受理，丢弃快照。 */
  function endSubmit() {
    clearSubmitTimer()
    isSubmitting.value = false
    pendingSubmission = null
  }

  /** 提交被拒绝：恢复提交前的选择。返回是否有可恢复的快照。 */
  function restoreSubmission(): boolean {
    const snap = pendingSubmission
    endSubmit()
    if (!snap) return false
    actionMode.value = snap.actionMode
    magicSubChoice.value = snap.magicSubChoice
    selectedHandIndexForAction.value = snap.selectedHandIndexForAction
    skillMode.value = snap.skillMode
    selectedSkill.value = snap.selectedSkill
    skillTargetIds.value = snap.skillTargetIds
    skillDiscardHandIndexes.value = snap.skillDiscardHandIndexes
    selectedHandIndexes.value = snap.selectedHandIndexes
    selectedTargets.value = snap.selectedTargets
    promptCounterTarget.value = snap.promptCounterTarget
    counterFlowStage.value = snap.counterFlowStage
    return true
  }

  function hasPendingSubmission() {
    return pendingSubmission !== null
  }

  function setPromptGated(gated: boolean) {
    promptGated.value = gated
  }

  function syncAfterStateUpdate() {
    // state_update means the previous prompt was accepted or superseded.
    // If a new prompt is needed, the server sends RequireAction right after
    // the fresh SyncState, so clearing here prevents stale controls sticking.
    currentPrompt.value = null
    waitingFor.value = ''
    clearSelections()
    clearActionState()
    endSubmit()
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
    endSubmit()
    promptGated.value = false
  }

  return {
    currentPrompt,
    waitingFor,
    selectedHandIndexes,
    selectedFieldOptionIndexes,
    selectedTargets,
    promptCounterTarget,
    counterFlowStage,
    errorMessage,
    skillEffectToast,
    actionMode,
    magicSubChoice,
    selectedHandIndexForAction,
    skillMode,
    selectedSkill,
    skillTargetIds,
    skillDiscardHandIndexes,
    isSubmitting,
    promptGated,
    beginSubmit,
    endSubmit,
    restoreSubmission,
    hasPendingSubmission,
    setPromptGated,
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
    setCounterFlowStage,
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
