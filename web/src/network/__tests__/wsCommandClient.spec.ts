import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { createWsCommandClient } from '../wsCommandClient'
import { useBattleFxStore } from '../../stores/battlefx.store'
import { useBattleReviewStore } from '../../stores/battleReview.store'
import { useInterruptStore } from '../../stores/interrupt.store'
import { useSessionStore } from '../../stores/session.store'
import type { WsOutboundMessage } from '../protocol'

function buildClient(options?: {
  connected?: boolean
}) {
  const interruptStore = useInterruptStore()
  const sessionStore = useSessionStore()
  const battleFxStore = useBattleFxStore()
  const battleReviewStore = useBattleReviewStore()
  const sendEnvelope = vi.fn<(msg: WsOutboundMessage) => void>()
  let connected = options?.connected ?? true
  sessionStore.setRoomInfo('ROOM1', 'p1', 'Red', 'hero')

  const client = createWsCommandClient({
    interruptStore,
    sessionStore,
    battleFxStore,
    battleReviewStore,
    isTransportOpen: () => connected,
    sendEnvelope,
    safeStringify: (data) => JSON.stringify(data),
  })

  return {
    client,
    interruptStore,
    battleFxStore,
    battleReviewStore,
    sendEnvelope,
    setConnected(value: boolean) {
      connected = value
    },
  }
}

describe('createWsCommandClient', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('blocks submit and room actions when transport is offline', () => {
    const { client, interruptStore, sendEnvelope } = buildClient({ connected: false })

    client.sendAction({
      player_id: 'p1',
      type: 'Pass',
    })
    expect(interruptStore.errorMessage).toBe('未连接到服务器')
    expect(sendEnvelope).not.toHaveBeenCalled()

    interruptStore.clearError()
    client.startRoom()
    expect(interruptStore.errorMessage).toBe('未连接到服务器')
    expect(sendEnvelope).not.toHaveBeenCalled()
  })

  it('silently skips chat sends while disconnected and emits ChatMessage when reconnected', () => {
    const { client, battleReviewStore, sendEnvelope, setConnected } = buildClient({ connected: false })

    client.sendChat('hello')
    expect(sendEnvelope).not.toHaveBeenCalled()
    expect(battleReviewStore.logs).toEqual([])

    setConnected(true)
    client.sendChat('hello')

    expect(sendEnvelope).toHaveBeenCalledWith({
      Cmd: 'ChatMessage',
      Data: { message: 'hello' },
    })
    expect(battleReviewStore.logs[battleReviewStore.logs.length - 1]).toBe(
      '[WS][TX] ChatMessage: {"message":"hello"}'
    )
  })

  it('sends submit envelopes and starts focus for magic and skill actions', () => {
    const { client, battleFxStore, sendEnvelope } = buildClient()
    const focusSpy = vi.spyOn(battleFxStore, 'startSkillInitiatorFocus')

    client.magic('p2', 'magic-1')
    client.useSkill('skill-1', ['p2', 'p3'], [1])

    expect(focusSpy).toHaveBeenNthCalledWith(1, 'p1', 'magic')
    expect(focusSpy).toHaveBeenNthCalledWith(2, 'p1', 'skill')
    expect(sendEnvelope).toHaveBeenNthCalledWith(1, {
      Cmd: 'SubmitAction',
      Data: {
        action_type: 'Magic',
        card_id: 'magic-1',
        targets: [{ target_user_id: 'p2' }],
      },
    })
    expect(sendEnvelope).toHaveBeenNthCalledWith(2, {
      Cmd: 'SubmitAction',
      Data: {
        action_type: 'Skill',
        skill_id: 'skill-1',
        targets: [{ target_user_id: 'p2' }, { target_user_id: 'p3' }],
        option_indexes: [1],
      },
    })
  })

  it('wraps lobby intents into RoomAction envelopes', () => {
    const { client, battleReviewStore, sendEnvelope } = buildClient()

    client.changeCamp('Red')
    client.changeRole('mage', 'bot-1')
    client.addBot('机器人1')
    client.removeBot('bot-1')
    client.takeoverPlayer('p2')
    client.startRoom()
    client.dissolveRoom()

    expect(sendEnvelope).toHaveBeenNthCalledWith(1, {
      Cmd: 'RoomAction',
      Data: {
        action: 'change_camp',
        camp: 'Red',
      },
    })
    expect(sendEnvelope).toHaveBeenNthCalledWith(2, {
      Cmd: 'RoomAction',
      Data: {
        action: 'change_role',
        target_id: 'bot-1',
        char_role: 'mage',
      },
    })
    expect(sendEnvelope).toHaveBeenNthCalledWith(3, {
      Cmd: 'RoomAction',
      Data: {
        action: 'add_bot',
        bot_name: '机器人1',
      },
    })
    expect(sendEnvelope).toHaveBeenNthCalledWith(4, {
      Cmd: 'RoomAction',
      Data: {
        action: 'remove_bot',
        target_id: 'bot-1',
      },
    })
    expect(sendEnvelope).toHaveBeenNthCalledWith(5, {
      Cmd: 'RoomAction',
      Data: {
        action: 'takeover_player',
        target_id: 'p2',
      },
    })
    expect(sendEnvelope).toHaveBeenNthCalledWith(6, {
      Cmd: 'RoomAction',
      Data: {
        action: 'start',
      },
    })
    expect(sendEnvelope).toHaveBeenNthCalledWith(7, {
      Cmd: 'RoomAction',
      Data: {
        action: 'dissolve_room',
      },
    })
    expect(battleReviewStore.logs[battleReviewStore.logs.length - 1]).toBe(
      '[WS][TX] RoomAction: {"action":"dissolve_room"}'
    )
  })

  it('sends cheat discard as Cheat action payload', () => {
    const { client, sendEnvelope } = buildClient()

    client.cheatDiscard('p2', 3)

    expect(sendEnvelope).toHaveBeenCalledWith({
      Cmd: 'SubmitAction',
      Data: {
        action_type: 'Cheat',
        targets: [{ target_user_id: 'discard' }],
        extra_args: ['p2', '3'],
      },
    })
  })
})
