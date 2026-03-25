import { describe, expect, it } from 'vitest'
import { buildWsConnectUrl, loadReconnectInfo, reconnectStorageKey, saveReconnectInfo } from '../wsReconnect'

function createStorage() {
  const data = new Map<string, string>()
  return {
    getItem(key: string) {
      return data.get(key) ?? null
    },
    setItem(key: string, value: string) {
      data.set(key, value)
    },
  }
}

describe('wsReconnect', () => {
  it('persists and reloads reconnect info with a stable storage key', () => {
    const storage = createStorage()

    saveReconnectInfo(storage, 'ROOM1', 'Alice Smith', 'p1', 'token-1')

    expect(reconnectStorageKey('ROOM1', 'Alice Smith')).toBe('xbs_reconnect_ROOM1_Alice%20Smith')
    expect(loadReconnectInfo(storage, 'ROOM1', 'Alice Smith')).toEqual({
      room_code: 'ROOM1',
      player_id: 'p1',
      player_name: 'Alice Smith',
      token: 'token-1',
    })
  })

  it('ignores malformed or mismatched reconnect payloads', () => {
    const storage = createStorage()
    storage.setItem(reconnectStorageKey('ROOM1', 'Alice'), '{"room_code":"ROOM2"}')

    expect(loadReconnectInfo(storage, 'ROOM1', 'Alice')).toBeNull()
  })

  it('builds create-room and join-room urls with reconnect parameters', () => {
    expect(buildWsConnectUrl({
      baseUrl: 'ws://localhost:8080/ws',
      roomCode: 'ROOM1',
      playerName: 'Alice',
      createRoom: true,
    })).toBe('ws://localhost:8080/ws?name=Alice&create=true')

    expect(buildWsConnectUrl({
      baseUrl: 'ws://localhost:8080/ws',
      roomCode: 'ROOM1',
      playerName: 'Alice',
      reconnectInfo: {
        player_id: 'p1',
        token: 'token-1',
      },
    })).toBe('ws://localhost:8080/ws?room=ROOM1&name=Alice&player_id=p1&reconnect_token=token-1')
  })
})
