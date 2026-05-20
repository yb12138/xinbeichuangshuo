import { computed, ref, watch } from 'vue'
import { useInterruptStore } from '../stores/interrupt.store'
import { useSessionStore } from '../stores/session.store'
import { useSnapshotStore } from '../stores/snapshot.store'
import { useSubmitAction } from './useSubmitAction'
import type { PlayerView, Prompt, PromptOption } from '../types/game'

export type PromptTargetOptionEntry = {
  index: number
  option: PromptOption
  player: PlayerView
}

export function useInteractionController() {
  const interruptStore = useInterruptStore()
  const sessionStore = useSessionStore()
  const snapshotStore = useSnapshotStore()
  const actions = useSubmitAction()

  const prompt = computed(() => interruptStore.currentPrompt)
  const myPlayerId = computed(() => sessionStore.myPlayerId)
  const playerViews = computed(() => snapshotStore.players)
  const selectedExtractIndices = ref<number[]>([])
  const selectedInlineCardIDs = ref<string[]>([])

  watch(() => prompt.value, () => {
    interruptStore.setPromptCounterTarget('')
    selectedExtractIndices.value = []
    selectedInlineCardIDs.value = []
  })

  function showPromptError(message: string) {
    interruptStore.showError(message)
  }

  function resolveOptionPlayerId(option: { target_id?: string | null }): string | null {
    const targetId = String(option.target_id || '').trim()
    if (!targetId) return null
    return playerViews.value[targetId] ? targetId : null
  }

  const playerOptionEntries = computed<PromptTargetOptionEntry[]>(() => {
    if (!prompt.value?.options?.length) return []
    if (prompt.value.presentation?.kind === 'skill_choice') return []
    return prompt.value.options
      .map((option, index) => {
        const playerId = resolveOptionPlayerId(option)
        if (!playerId) return null
        const player = playerViews.value[playerId]
        if (!player) return null
        return { index, option, player }
      })
      .filter((entry): entry is PromptTargetOptionEntry => entry != null)
  })

  const playerOptionIndexSet = computed(() => {
    const set = new Set<number>()
    for (const entry of playerOptionEntries.value) {
      set.add(entry.index)
    }
    return set
  })

  const selectedPromptTargetOptionIndexes = computed(() => {
    const indexByPlayerId = new Map<string, number>()
    for (const entry of playerOptionEntries.value) {
      indexByPlayerId.set(entry.player.id, entry.index)
    }
    const indexes: number[] = []
    for (const targetId of interruptStore.selectedTargets) {
      const index = indexByPlayerId.get(targetId)
      if (index === undefined) return []
      indexes.push(index)
    }
    return indexes
  })

  const nonPlayerOptions = computed(() => {
    const options = prompt.value?.options ?? []
    return options.filter((_, idx) => !playerOptionIndexSet.value.has(idx))
  })

  function isPromptCancellationAllowedByPolicy(p: Prompt): boolean {
    const cancelPolicy = p.presentation?.cancel_policy
    return cancelPolicy === 'abort' || cancelPolicy === 'decline' || cancelPolicy === 'back'
  }

  const canCancelPrompt = computed(() => {
    if (!prompt.value) return false
    return isPromptCancellationAllowedByPolicy(prompt.value)
  })

  function cancelPrompt() {
    if (!prompt.value || !canCancelPrompt.value) {
      showPromptError('当前步骤不可取消，请先完成本次操作')
      return false
    }
    actions.submitCancel()
    return true
  }

  function submitOptionIndexes(indexes: number[]) {
    actions.submitSelect(indexes)
  }

  function submitOptionIndex(index: number) {
    actions.submitSelect([index])
  }

  function submitConfirm() {
    actions.submitConfirm()
  }

  function submitTargetSelection() {
    if (!prompt.value) return false
    if (interruptStore.selectedTargets.length === 1) {
      const targetId = interruptStore.selectedTargets[0]
      if (!targetId) return false
      actions.submitPromptTarget(targetId)
      return true
    }
    actions.submitAction({
      player_id: myPlayerId.value,
      type: 'Select',
      target_ids: interruptStore.selectedTargets,
    })
    return true
  }

  function submitSelectedCardIDs(cardIds: string[]) {
    actions.submitSelectCardIDs(cardIds)
  }

  function submitRespondTake() {
    actions.submitRespondTake()
  }

  function submitRespondCounter(isMagicMissilePrompt = false) {
    return actions.submitRespondCounter(isMagicMissilePrompt)
  }

  function submitRespondDefend() {
    return actions.submitRespondDefend()
  }

  function toggleExtractOption(index: number, maxCount: number) {
    const idx = selectedExtractIndices.value.indexOf(index)
    if (idx >= 0) {
      selectedExtractIndices.value.splice(idx, 1)
      return
    }
    if (selectedExtractIndices.value.length < maxCount) {
      selectedExtractIndices.value.push(index)
      selectedExtractIndices.value.sort((a, b) => a - b)
    }
  }

  function toggleInlineCardID(cardID: string, maxCount: number) {
    if (!cardID) return
    const pos = selectedInlineCardIDs.value.indexOf(cardID)
    if (pos >= 0) {
      selectedInlineCardIDs.value.splice(pos, 1)
      return
    }
    if (selectedInlineCardIDs.value.length >= maxCount) return
    selectedInlineCardIDs.value.push(cardID)
  }

  return {
    prompt,
    myPlayerId,
    playerViews,
    selectedExtractIndices,
    selectedInlineCardIDs,
    playerOptionEntries,
    playerOptionIndexSet,
    selectedPromptTargetOptionIndexes,
    nonPlayerOptions,
    canCancelPrompt,
    showPromptError,
    cancelPrompt,
    submitConfirm,
    submitOptionIndex,
    submitOptionIndexes,
    submitTargetSelection,
    submitSelectedCardIDs,
    submitRespondTake,
    submitRespondCounter,
    submitRespondDefend,
    toggleExtractOption,
    toggleInlineCardID,
  }
}
