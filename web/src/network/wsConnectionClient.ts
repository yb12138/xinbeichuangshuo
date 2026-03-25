import { ref } from 'vue'
import { useBattleReviewStore } from '../stores/battleReview.store'
import { useInterruptStore } from '../stores/interrupt.store'
import { useMatchLifecycleStore } from '../stores/matchLifecycle.store'
import { useSessionStore } from '../stores/session.store'
import type { WsMessage } from './protocol'
import { buildWsConnectUrl, type ReconnectInfo } from './wsReconnect'

const reconnectAttempts = ref(0)
const maxReconnectAttempts = 5

let ws: WebSocket | null = null
let reconnectTimer: ReturnType<typeof setTimeout> | null = null

export interface WsConnectionClientDeps {
  sessionStore: ReturnType<typeof useSessionStore>
  interruptStore: ReturnType<typeof useInterruptStore>
  battleReviewStore: ReturnType<typeof useBattleReviewStore>
  matchLifecycleStore: ReturnType<typeof useMatchLifecycleStore>
  getWsUrl: () => string
  loadReconnectInfo?: (roomCode: string, playerName: string) => ReconnectInfo | null
  createSocket?: (url: string) => WebSocket
  safeStringify: (data: unknown) => string
  onMessage: (msg: WsMessage) => void
}

function clearReconnectTimer() {
  if (!reconnectTimer) return
  clearTimeout(reconnectTimer)
  reconnectTimer = null
}

export function createWsConnectionClient(deps: WsConnectionClientDeps) {
  const {
    sessionStore,
    interruptStore,
    battleReviewStore,
    matchLifecycleStore,
    getWsUrl,
    loadReconnectInfo,
    safeStringify,
    onMessage,
  } = deps

  const createSocket = deps.createSocket ?? ((url: string) => new WebSocket(url))

  function closeSocketForManualStop() {
    clearReconnectTimer()
    reconnectAttempts.value = 0
    if (!ws) {
      sessionStore.setConnected(false)
      return
    }

    const socket = ws
    ws = null
    socket.onopen = null
    socket.onmessage = null
    socket.onerror = null
    socket.onclose = null
    sessionStore.setConnected(false)
    battleReviewStore.addLog('[WS] 连接关闭')
    socket.close()
  }

  function resolveReconnectInfo(roomCode: string, playerName: string, createRoom: boolean) {
    if (createRoom) return null

    const sessionReconnect =
      sessionStore.reconnectToken &&
      sessionStore.myPlayerId &&
      sessionStore.roomCode === roomCode &&
      sessionStore.myName === playerName
        ? {
            player_id: sessionStore.myPlayerId,
            token: sessionStore.reconnectToken,
          }
        : null

    return sessionReconnect || loadReconnectInfo?.(roomCode, playerName) || null
  }

  function connect(roomCode: string, playerName: string, createRoom = false) {
    sessionStore.setMyName(playerName)

    closeSocketForManualStop()

    const url = buildWsConnectUrl({
      baseUrl: getWsUrl(),
      roomCode,
      playerName,
      createRoom,
      reconnectInfo: resolveReconnectInfo(roomCode, playerName, createRoom),
    })

    console.log('Connecting to:', url)
    battleReviewStore.addLog(`[WS] 连接中: ${url}`)

    ws = createSocket(url)

    ws.onopen = () => {
      console.log('WebSocket connected')
      sessionStore.setConnected(true)
      reconnectAttempts.value = 0
      clearReconnectTimer()
      battleReviewStore.addLog('[WS] 连接成功')
    }

    ws.onmessage = (event) => {
      try {
        battleReviewStore.addLog(`[WS][RX] raw: ${String(event.data)}`)
        const msg: WsMessage = JSON.parse(String(event.data))
        battleReviewStore.addLog(`[WS][RX] ${msg.Cmd}: ${safeStringify(msg.Data)}`)
        onMessage(msg)
      } catch (error) {
        console.error('Failed to parse message:', error)
      }
    }

    ws.onerror = (error) => {
      console.error('WebSocket error:', error)
      interruptStore.showError('连接错误')
      battleReviewStore.addLog('[WS] 连接错误')
    }

    ws.onclose = () => {
      console.log('WebSocket closed')
      sessionStore.setConnected(false)
      battleReviewStore.addLog('[WS] 连接关闭')

      if (reconnectAttempts.value >= maxReconnectAttempts || !sessionStore.isInRoom) {
        return
      }

      reconnectAttempts.value++
      console.log(`Reconnecting... attempt ${reconnectAttempts.value}`)
      clearReconnectTimer()
      reconnectTimer = setTimeout(() => {
        reconnectTimer = null
        if (sessionStore.roomCode) {
          connect(sessionStore.roomCode, sessionStore.myName)
        }
      }, 2000 * reconnectAttempts.value)
    }
  }

  function disconnect() {
    closeSocketForManualStop()
    matchLifecycleStore.resetAll()
  }

  function closeTransport() {
    closeSocketForManualStop()
  }

  function isTransportOpen() {
    return !!ws && ws.readyState === WebSocket.OPEN
  }

  function sendEnvelope(msg: WsMessage) {
    ws?.send(JSON.stringify(msg))
  }

  return {
    connect,
    disconnect,
    closeTransport,
    isTransportOpen,
    sendEnvelope,
  }
}
