import { useInterruptStore } from '../stores/interrupt.store'
import { useSessionStore } from '../stores/session.store'
import { useSnapshotStore } from '../stores/snapshot.store'
import { useTimelineStore } from '../stores/timeline.store'
import { useUiStore } from '../stores/ui.store'
import { useBattleFxStore } from '../stores/battlefx.store'
import { useBattleReviewStore } from '../stores/battleReview.store'
import { useMatchLifecycleStore } from '../stores/matchLifecycle.store'
import { createWsConnectionClient } from '../network/wsConnectionClient'
import { createGameplayMessageHandlers } from '../network/gameplayMessageHandlers'
import { createRoomMessageHandlers } from '../network/roomMessageHandlers'
import { createWsCommandClient } from '../network/wsCommandClient'
import { routeWsMessage } from '../network/messageRouter'
import type { WsMessage } from '../network/protocol'
import { loadReconnectInfo, saveReconnectInfo } from '../network/wsReconnect'

// 改造成一个函数，动态获取当前访问的 IP 和端口
const getWsUrl = () => {
  // 如果有配环境变量，优先用环境变量（方便以后线上部署）
  if (import.meta.env.VITE_WS_URL) {
    return import.meta.env.VITE_WS_URL
  }

  // 开发模式优先直连后端 8080，避免经过 Vite 的 /ws 代理产生 EPIPE 噪声。
  if (import.meta.env.DEV) {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const host = window.location.hostname
    return `${protocol}//${host}:8080/ws`
  }

  // 动态拼接地址
  // window.location.host 会自动变成比如 '192.168.1.100:5173' 或者 'localhost:5173'
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const host = window.location.host
  return `${protocol}//${host}/ws`
}

export function useWebSocket() {
  const interruptStore = useInterruptStore()
  const sessionStore = useSessionStore()
  const snapshotStore = useSnapshotStore()
  const timelineStore = useTimelineStore()
  const uiStore = useUiStore()
  const battleFxStore = useBattleFxStore()
  const battleReviewStore = useBattleReviewStore()
  const matchLifecycleStore = useMatchLifecycleStore()
  let routeMessage = (_msg: WsMessage) => {}
  const gameplayHandlers = createGameplayMessageHandlers({
    interruptStore,
    sessionStore,
    snapshotStore,
    timelineStore,
    uiStore,
    battleFxStore,
    battleReviewStore,
    matchLifecycleStore,
  })
  const connectionClient = createWsConnectionClient({
    sessionStore,
    interruptStore,
    battleReviewStore,
    matchLifecycleStore,
    getWsUrl,
    loadReconnectInfo: (roomCode, playerName) => (
      typeof window === 'undefined'
        ? null
        : loadReconnectInfo(window.localStorage, roomCode, playerName)
    ),
    safeStringify,
    onMessage: (msg) => {
      routeMessage(msg)
    },
  })
  const roomHandlers = createRoomMessageHandlers({
    sessionStore,
    snapshotStore,
    interruptStore,
    battleReviewStore,
    matchLifecycleStore,
    closeTransport: connectionClient.closeTransport,
    persistReconnectInfo: (roomCode, playerId, token) => {
      if (typeof window === 'undefined') return
      saveReconnectInfo(window.localStorage, roomCode, sessionStore.myName, playerId, token)
    },
  })
  routeMessage = (msg: WsMessage) => {
    routeWsMessage(msg, {
      onRoomEvent: roomHandlers.handleRoomEvent,
      onSyncState: gameplayHandlers.handleSyncState,
      onRequireAction: gameplayHandlers.handleRequireAction,
      onNotifyTimeline: gameplayHandlers.handleNotifyTimeline,
      onUnknown: (cmd, payload) => {
        console.log('Unknown message cmd:', cmd, payload)
      }
    })
  }
  const commandClient = createWsCommandClient({
    interruptStore,
    sessionStore,
    battleFxStore,
    battleReviewStore,
    isTransportOpen: connectionClient.isTransportOpen,
    sendEnvelope: connectionClient.sendEnvelope,
    safeStringify,
  })

  function safeStringify(data: unknown) {
    try {
      return JSON.stringify(data)
    } catch {
      return '[unserializable]'
    }
  }
  return {
    connect: connectionClient.connect,
    disconnect: connectionClient.disconnect,
    ...commandClient,
  }
}
