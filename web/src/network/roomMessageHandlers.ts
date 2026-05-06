import type { RoomEvent } from '../types/game'
import { useBattleReviewStore } from '../stores/battleReview.store'
import { useInterruptStore } from '../stores/interrupt.store'
import { useMatchLifecycleStore } from '../stores/matchLifecycle.store'
import { useSessionStore } from '../stores/session.store'
import { useSnapshotStore } from '../stores/snapshot.store'

export interface RoomMessageHandlerDeps {
  sessionStore: ReturnType<typeof useSessionStore>
  snapshotStore: ReturnType<typeof useSnapshotStore>
  interruptStore: ReturnType<typeof useInterruptStore>
  battleReviewStore: ReturnType<typeof useBattleReviewStore>
  matchLifecycleStore: ReturnType<typeof useMatchLifecycleStore>
  closeTransport: () => void
  persistReconnectInfo?: (roomCode: string, playerId: string, token: string) => void
}

export function createRoomMessageHandlers(deps: RoomMessageHandlerDeps) {
  const {
    sessionStore,
    snapshotStore,
    interruptStore,
    battleReviewStore,
    matchLifecycleStore,
    closeTransport,
    persistReconnectInfo,
  } = deps

  function handleRoomEvent(event: RoomEvent) {
    console.log('Room event:', event)

    switch (event.action) {
      case 'assigned':
        sessionStore.setRoomInfo(
          event.room_code,
          event.player_id!,
          event.camp || '',
          event.char_role || '',
        )
        if (event.reconnect_token && event.player_id) {
          sessionStore.setReconnectToken(event.reconnect_token)
          persistReconnectInfo?.(event.room_code, event.player_id, event.reconnect_token)
        }
        if (event.characters?.length) {
          snapshotStore.setCharacters(event.characters)
        }
        battleReviewStore.addLog(`已加入房间 ${event.room_code}，你是 ${event.player_id}`)
        break

      case 'player_list':
        sessionStore.updateRoomPlayers(event.players || [], sessionStore.myPlayerId || undefined)
        if (event.characters?.length) {
          snapshotStore.setCharacters(event.characters)
        }
        break

      case 'started':
        matchLifecycleStore.setGameStarted()
        if (event.characters?.length) {
          snapshotStore.setCharacters(event.characters)
        }
        battleReviewStore.addLog('游戏开始！')
        break

      case 'joined':
        battleReviewStore.addLog(event.message || `${event.player_name} 加入了房间`)
        break

      case 'left':
        battleReviewStore.addLog(event.message || `${event.player_name} 离开了房间`)
        break

      case 'error':
        interruptStore.showError(event.message || '房间错误')
        break

      case 'dissolved':
        {
          const msg = event.message || '房间已解散'
          closeTransport()
          matchLifecycleStore.resetAll()
          interruptStore.showError(msg)
        }
        break
    }
  }

  return {
    handleRoomEvent,
  }
}
