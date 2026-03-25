import { defineStore } from 'pinia'
import { useBattleFxStore } from './battlefx.store'
import { useBattleReviewStore, type MoraleCamp } from './battleReview.store'
import { useInterruptStore } from './interrupt.store'
import { useSessionStore } from './session.store'
import { useSnapshotStore } from './snapshot.store'
import { useTimelineStore } from './timeline.store'
import { useUiStore } from './ui.store'
import type { GameEndSnapshot } from './storeTypes'

export const useMatchLifecycleStore = defineStore('matchLifecycle', () => {
  const sessionStore = useSessionStore()
  const snapshotStore = useSnapshotStore()
  const interruptStore = useInterruptStore()
  const timelineStore = useTimelineStore()
  const uiStore = useUiStore()
  const battleFxStore = useBattleFxStore()
  const battleReviewStore = useBattleReviewStore()

  function buildGameEndSnapshot(message: string): GameEndSnapshot {
    const triggerType: GameEndSnapshot['triggerType'] =
      snapshotStore.redCups >= 5 || snapshotStore.blueCups >= 5
        ? 'cups'
        : snapshotStore.redMorale <= 0 || snapshotStore.blueMorale <= 0
          ? 'morale'
          : 'unknown'

    const triggerCamp: MoraleCamp | undefined =
      snapshotStore.redMorale <= 0
        ? 'Red'
        : snapshotStore.blueMorale <= 0
          ? 'Blue'
          : undefined

    const triggerEntry = [...battleReviewStore.moraleChanges]
      .reverse()
      .find(item => (triggerCamp ? item.camp === triggerCamp : true) && item.delta < 0)

    return {
      message: message || '游戏结束',
      triggerType,
      finalRedMorale: snapshotStore.redMorale,
      finalBlueMorale: snapshotStore.blueMorale,
      finalRedCups: snapshotStore.redCups,
      finalBlueCups: snapshotStore.blueCups,
      triggerCamp: triggerEntry?.camp,
      triggerDelta: triggerEntry ? Math.abs(triggerEntry.delta) : undefined,
      triggerSource: triggerEntry?.source,
    }
  }

  function setGameStarted() {
    battleReviewStore.clearMoraleTracking()
    sessionStore.setGameStarted(true)
    uiStore.setGameEnded(false, '', null)
  }

  function refreshGameEndSnapshot(message?: string) {
    const nextMessage = message || uiStore.gameEndMessage || '游戏结束'
    const snapshot = buildGameEndSnapshot(nextMessage)
    if (uiStore.isGameEnded) {
      uiStore.setGameEnded(true, nextMessage, snapshot)
    }
    return snapshot
  }

  function setGameEnded(message: string) {
    const nextMessage = message || '游戏结束'
    const snapshot = buildGameEndSnapshot(nextMessage)
    battleFxStore.clearForGameEnd()
    uiStore.setGameEnded(true, nextMessage, snapshot)
    interruptStore.reset()
  }

  function clearGameEnded() {
    uiStore.setGameEnded(false, '', null)
    timelineStore.clear()
  }

  function resetAll() {
    sessionStore.reset()
    snapshotStore.reset()
    interruptStore.reset()
    timelineStore.clear()
    uiStore.reset()
    battleFxStore.reset()
    battleReviewStore.reset()
  }

  return {
    buildGameEndSnapshot,
    setGameStarted,
    refreshGameEndSnapshot,
    setGameEnded,
    clearGameEnded,
    resetAll,
  }
})
