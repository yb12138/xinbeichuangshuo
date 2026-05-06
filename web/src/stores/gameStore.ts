import { defineStore, storeToRefs } from 'pinia'
import { computed } from 'vue'
import { useBattleInteractionState } from '../composables/useBattleInteractionState'
import type { TimelineNotifyPayload } from '../network/protocol'
import type { GameStateUpdate, PlayerInfo, Prompt } from '../types/game'
import { useBattleFxStore } from './battlefx.store'
import { useBattleReviewStore } from './battleReview.store'
import { useInterruptStore } from './interrupt.store'
import { useMatchLifecycleStore } from './matchLifecycle.store'
import { useSessionStore } from './session.store'
import { useSnapshotStore } from './snapshot.store'
import { useTimelineStore } from './timeline.store'
import { useUiStore } from './ui.store'

export type {
  BattleFeedEntry,
  BattleFeedType,
  MoraleCamp,
  MoraleChangeEntry,
  MoraleHint,
} from './battleReview.store'
export type { GameEndSnapshot, SkillModalAnchor } from './storeTypes'
export type { InitiatorFocusMode, InitiatorFocusSide, InitiatorFocusState } from './battlefx.store'

export const useGameStore = defineStore('game', () => {
  const sessionStore = useSessionStore()
  const snapshotStore = useSnapshotStore()
  const interruptStore = useInterruptStore()
  const timelineStore = useTimelineStore()
  const uiStore = useUiStore()
  const battleFxStore = useBattleFxStore()
  const battleReviewStore = useBattleReviewStore()
  const matchLifecycleStore = useMatchLifecycleStore()

  const sessionRefs = storeToRefs(sessionStore)
  const snapshotRefs = storeToRefs(snapshotStore)
  const interruptRefs = storeToRefs(interruptStore)
  const timelineRefs = storeToRefs(timelineStore)
  const uiRefs = storeToRefs(uiStore)
  const battleFxRefs = storeToRefs(battleFxStore)
  const battleReviewRefs = storeToRefs(battleReviewStore)

  const interaction = useBattleInteractionState()
  const {
    myPlayer,
    myHand,
    myBlessings,
    myExclusiveCards,
    myPlayableCards,
    isMyTurn,
    isPromptForMe,
    targetablePlayers,
    targetablePlayersForSkill,
    canTargetOpponent: canTargetOpponentRef,
    effectiveAvailableSkills,
    canConfirmSkill,
    getCharacter,
    getRoleDisplayName,
    cardMatchesExclusive,
  } = interaction

  const redPlayers = computed(() =>
    Object.values(snapshotRefs.players.value).filter((player) => player.camp === 'Red')
  )
  const bluePlayers = computed(() =>
    Object.values(snapshotRefs.players.value).filter((player) => player.camp === 'Blue')
  )
  const opponentPlayers = computed(() =>
    sessionRefs.myCamp.value === 'Red' ? bluePlayers.value : redPlayers.value
  )
  const allyPlayers = computed(() =>
    sessionRefs.myCamp.value === 'Red' ? redPlayers.value : bluePlayers.value
  )
  const skillModalCharacter = computed(() => {
    const id = uiRefs.skillModalCharacterId.value
    return id ? (snapshotRefs.characters.value[id] ?? null) : null
  })
  const timelinePayloads = timelineRefs.payloads

  function updateRoomPlayers(playerList: PlayerInfo[]) {
    sessionStore.updateRoomPlayers(playerList, sessionRefs.myPlayerId.value || undefined)
  }

  function setGameStarted() {
    timelineStore.clear()
    matchLifecycleStore.setGameStarted()
  }

  function updateGameState(state: GameStateUpdate) {
    battleFxStore.prepareForFlowUpdate(state.combat_stage, state.subflow)
    snapshotStore.updateGameState(state)
    const me = sessionRefs.myPlayerId.value ? state.players[sessionRefs.myPlayerId.value] : undefined
    if (me?.camp || me?.role) {
      sessionStore.setSeat(me.camp || sessionRefs.myCamp.value, me.role || sessionRefs.myCharRole.value)
    }
    interruptStore.syncAfterStateUpdate()
    battleFxStore.syncInitiatorFocusWithState(state.combat_stage, state.subflow)
    if (uiStore.isGameEnded) {
      matchLifecycleStore.refreshGameEndSnapshot(uiStore.gameEndMessage || '游戏结束')
    }
  }

  function setPrompt(prompt: Prompt | null) {
    if (prompt?.player_id) {
      battleFxStore.touchSkillInitiatorFocus(prompt.player_id)
    }
    interruptStore.setPrompt(prompt)
  }

  function toggleCardSelection(index: number) {
    const next = [...interruptRefs.selectedCards.value]
    const idx = next.indexOf(index)
    if (idx === -1) {
      if (interruptRefs.currentPrompt.value?.max === 1) {
        interruptStore.setSelectedCards([index])
        return
      }
      next.push(index)
    } else {
      next.splice(idx, 1)
    }
    interruptStore.setSelectedCards(next)
  }

  function selectTarget(playerId: string) {
    const next = [...interruptRefs.selectedTargets.value]
    const idx = next.indexOf(playerId)
    if (idx >= 0) {
      next.splice(idx, 1)
    } else if (interruptRefs.currentPrompt.value?.max === 1) {
      interruptStore.setSelectedTargets([playerId])
      return
    } else {
      next.push(playerId)
    }
    interruptStore.setSelectedTargets(next)
  }

  function setActionModeForAttack(mode: 'none' | 'attack' | 'magic') {
    interruptStore.setActionMode(mode)
    if (mode === 'none') {
      interruptStore.setSelectedCardForAction(null)
    }
  }

  function canTargetOpponent() {
    return canTargetOpponentRef.value
  }

  function pushTimelinePayload(payload: TimelineNotifyPayload) {
    timelineStore.push(payload)
  }

  return {
    ...sessionRefs,
    ...snapshotRefs,
    ...interruptRefs,
    ...uiRefs,
    ...battleFxRefs,
    actionSummaryLines: battleReviewRefs.actionSummaryLines,
    battleFeed: battleReviewRefs.battleFeed,
    moraleChanges: battleReviewRefs.moraleChanges,
    moraleBurstRanking: battleReviewRefs.moraleBurstRanking,
    logs: battleReviewRefs.logs,
    timelinePayloads,
    effectiveAvailableSkills,
    canConfirmSkill,
    skillModalCharacter,
    myPlayer,
    myHand,
    myBlessings,
    myExclusiveCards,
    myPlayableCards,
    isMyTurn,
    isPromptForMe,
    redPlayers,
    bluePlayers,
    opponentPlayers,
    allyPlayers,
    targetablePlayers,
    targetablePlayersForSkill,
    setRoomInfo: sessionStore.setRoomInfo,
    setReconnectToken: sessionStore.setReconnectToken,
    updateRoomPlayers,
    setCharacters: snapshotStore.setCharacters,
    setGameStarted,
    updateGameState,
    setPrompt,
    setWaiting: interruptStore.setWaiting,
    addLog: battleReviewStore.addLog,
    clearLogs: battleReviewStore.clearLogs,
    toggleCardSelection,
    selectTarget,
    setPromptCounterTarget: interruptStore.setPromptCounterTarget,
    setActionModeForAttack,
    setMagicSubChoice: interruptStore.setMagicSubChoice,
    setSelectedCardForAction: interruptStore.setSelectedCardForAction,
    canTargetOpponent,
    clearActionMode: interruptStore.clearActionMode,
    setSkillMode: interruptStore.setSkillMode,
    setSelectedSkill: interruptStore.setSelectedSkill,
    toggleSkillTarget: interruptStore.toggleSkillTarget,
    toggleSkillDiscard: interruptStore.toggleSkillDiscard,
    clearSkillMode: interruptStore.clearSkillMode,
    setError: interruptStore.showError,
    setSkillEffectToast: interruptStore.showSkillEffectToast,
    openSkillModal: uiStore.openSkillModal,
    addFlyingCards: battleFxStore.addFlyingCards,
    addDrawBurst: battleFxStore.addDrawBurst,
    addDamageEffect: battleFxStore.addDamageEffect,
    addActionStep: battleReviewStore.addActionStep,
    clearActionSummary: battleReviewStore.clearActionSummary,
    addCombatCue: battleFxStore.addCombatCue,
    clearInitiatorFocus: battleFxStore.clearInitiatorFocus,
    startAttackInitiatorFocus: battleFxStore.startAttackInitiatorFocus,
    resolveAttackInitiatorFocus: battleFxStore.resolveAttackInitiatorFocus,
    startSkillInitiatorFocus: battleFxStore.startSkillInitiatorFocus,
    touchSkillInitiatorFocus: battleFxStore.touchSkillInitiatorFocus,
    settleSkillInitiatorFocus: battleFxStore.settleSkillInitiatorFocus,
    syncInitiatorFocusWithState: battleFxStore.syncInitiatorFocusWithState,
    addBattleFeed: battleReviewStore.addBattleFeed,
    clearBattleFeed: battleReviewStore.clearBattleFeed,
    setCinematicMode: uiStore.setCinematicMode,
    getCharacter,
    getRoleDisplayName,
    setConnected: sessionStore.setConnected,
    pushMoraleHint: battleReviewStore.pushMoraleHint,
    consumeMoraleHint: battleReviewStore.consumeMoraleHint,
    recordMoraleChange: battleReviewStore.recordMoraleChange,
    buildGameEndSnapshot: matchLifecycleStore.buildGameEndSnapshot,
    setGameEnded: matchLifecycleStore.setGameEnded,
    refreshGameEndSnapshot: matchLifecycleStore.refreshGameEndSnapshot,
    pushTimelinePayload,
    clearTimelinePayloads: timelineStore.clear,
    clearGameEnded: matchLifecycleStore.clearGameEnded,
    reset: matchLifecycleStore.resetAll,
    cardMatchesExclusive,
  }
})
