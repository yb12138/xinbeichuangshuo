import { render } from '@testing-library/vue'
import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import RoomLobby from '../RoomLobby.vue'

const mocks = vi.hoisted(() => ({
  connect: vi.fn(),
  changeCamp: vi.fn(),
  changeRole: vi.fn(),
  addBot: vi.fn(),
  removeBot: vi.fn(),
  startRoom: vi.fn(),
  dissolveRoom: vi.fn(),
}))

vi.mock('../../composables/useWebSocket', () => ({
  useWebSocket: () => ({
    connect: mocks.connect,
  }),
}))

vi.mock('../../composables/useSubmitAction', () => ({
  useSubmitAction: () => ({
    changeCamp: mocks.changeCamp,
    changeRole: mocks.changeRole,
    addBot: mocks.addBot,
    removeBot: mocks.removeBot,
    startRoom: mocks.startRoom,
    dissolveRoom: mocks.dissolveRoom,
  }),
}))

describe('RoomLobby URL auto join', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    for (const mock of Object.values(mocks)) {
      mock.mockReset()
    }
    window.history.pushState({}, '', '/')
  })

  it('joins the room from room and name query params', () => {
    window.history.pushState({}, '', '/?room=abcd&name=%E6%B5%8B%E8%AF%95%E7%8E%A9%E5%AE%B61')

    render(RoomLobby, {
      global: {
        plugins: [createPinia()],
      },
    })

    expect(mocks.connect).toHaveBeenCalledWith('ABCD', '测试玩家1')
  })

  it('does not auto join when required query params are missing', () => {
    window.history.pushState({}, '', '/?room=ABCD')

    render(RoomLobby, {
      global: {
        plugins: [createPinia()],
      },
    })

    expect(mocks.connect).not.toHaveBeenCalled()
  })
})
