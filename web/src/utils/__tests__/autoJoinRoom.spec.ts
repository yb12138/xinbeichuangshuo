import { describe, expect, it, vi } from 'vitest'
import { autoJoinRoomFromUrl } from '../autoJoinRoom'

describe('autoJoinRoomFromUrl', () => {
  it('connects to the room and player name from the URL', () => {
    const connect = vi.fn()

    const didAutoJoin = autoJoinRoomFromUrl({
      search: '?room=abcd&name=%E6%B5%8B%E8%AF%95%E7%8E%A9%E5%AE%B61',
      isInRoom: () => false,
      connect,
    })

    expect(didAutoJoin).toBe(true)
    expect(connect).toHaveBeenCalledWith('ABCD', '测试玩家1')
  })

  it('does not connect when room or name is missing', () => {
    const connect = vi.fn()

    expect(autoJoinRoomFromUrl({
      search: '?room=ABCD',
      isInRoom: () => false,
      connect,
    })).toBe(false)
    expect(autoJoinRoomFromUrl({
      search: '?name=%E6%B5%8B%E8%AF%95%E7%8E%A9%E5%AE%B61',
      isInRoom: () => false,
      connect,
    })).toBe(false)

    expect(connect).not.toHaveBeenCalled()
  })

  it('does not connect when already in a room', () => {
    const connect = vi.fn()

    const didAutoJoin = autoJoinRoomFromUrl({
      search: '?room=ABCD&name=%E6%B5%8B%E8%AF%95%E7%8E%A9%E5%AE%B61',
      isInRoom: () => true,
      connect,
    })

    expect(didAutoJoin).toBe(false)
    expect(connect).not.toHaveBeenCalled()
  })
})
