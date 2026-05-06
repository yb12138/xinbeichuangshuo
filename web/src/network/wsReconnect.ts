export interface ReconnectInfo {
  room_code: string
  player_id: string
  player_name: string
  token: string
}

interface StorageReader {
  getItem(key: string): string | null
}

interface StorageWriter {
  setItem(key: string, value: string): void
}

export const reconnectStorageKey = (roomCode: string, playerName: string) =>
  `xbs_reconnect_${roomCode}_${encodeURIComponent(playerName)}`

export function loadReconnectInfo(
  storage: StorageReader | undefined,
  roomCode: string,
  playerName: string,
): ReconnectInfo | null {
  if (!storage || !roomCode || !playerName) return null
  try {
    const raw = storage.getItem(reconnectStorageKey(roomCode, playerName))
    if (!raw) return null
    const parsed = JSON.parse(raw) as ReconnectInfo
    if (parsed.room_code !== roomCode || parsed.player_name !== playerName) return null
    if (!parsed.player_id || !parsed.token) return null
    return parsed
  } catch {
    return null
  }
}

export function saveReconnectInfo(
  storage: StorageWriter | undefined,
  roomCode: string,
  playerName: string,
  playerId: string,
  token: string,
) {
  if (!storage || !roomCode || !playerName || !playerId || !token) return
  const payload: ReconnectInfo = {
    room_code: roomCode,
    player_id: playerId,
    player_name: playerName,
    token,
  }
  try {
    storage.setItem(reconnectStorageKey(roomCode, playerName), JSON.stringify(payload))
  } catch {
    // ignore storage errors
  }
}

export function buildWsConnectUrl(options: {
  baseUrl: string
  roomCode: string
  playerName: string
  createRoom?: boolean
  reconnectInfo?: Pick<ReconnectInfo, 'player_id' | 'token'> | null
}) {
  const { baseUrl, roomCode, playerName, createRoom = false, reconnectInfo } = options
  if (createRoom) {
    return `${baseUrl}?name=${encodeURIComponent(playerName)}&create=true`
  }

  const reconnectParams = reconnectInfo
    ? `&player_id=${encodeURIComponent(reconnectInfo.player_id)}&reconnect_token=${encodeURIComponent(reconnectInfo.token)}`
    : ''

  return `${baseUrl}?room=${roomCode}&name=${encodeURIComponent(playerName)}${reconnectParams}`
}
