import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { createWsCommandClient } from '../wsCommandClient'
import { useBattleFxStore } from '../../stores/battlefx.store'
import { useBattleReviewStore } from '../../stores/battleReview.store'
import { useInterruptStore } from '../../stores/interrupt.store'
import { useSessionStore } from '../../stores/session.store'
import type { Card } from '../../types/game'
import type { WsMessage } from '../protocol'

function buildCard(overrides: Partial<Card> = {}): Card {
  return {
    id: 'card-1',
    name: '烈焰斩',
    type: 'Attack',
    element: 'Fire',
    damage: 2,
    description: 'test',
    ...overrides,
  }
}

function buildClient(options?: {
  connected?: boolean
  playableCards?: Card[]
}) {
  const interruptStore = useInterruptStore()
  const sessionStore = useSessionStore()
  const battleFxStore = useBattleFxStore()
  const battleReviewStore = useBattleReviewStore()
  const sendEnvelope = vi.fn<(msg: WsMessage) => void>()
  let connected = options?.connected ?? true
  const playableCards = options?.playableCards ?? []

  sessionStore.setRoomInfo('ROOM1', 'p1', 'Red', 'hero')

  const client = createWsCommandClient({
    interruptStore,
    sessionStore,
    battleFxStore,
    battleReviewStore,
    getPlayableCards: () => playableCards.map((card, index) => ({ card, index })),
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
    client.sendRoomAction('start_game')
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
    const { client, battleFxStore, sendEnvelope } = buildClient({
      playableCards: [
        buildCard({ id: 'magic-1', type: 'Magic', element: 'Dark', name: '魔弹' }),
      ],
    })
    const focusSpy = vi.spyOn(battleFxStore, 'startSkillInitiatorFocus')

    client.magic('p2', 0)
    client.useSkill('skill-1', ['p2', 'p3'], [1])

    expect(focusSpy).toHaveBeenNthCalledWith(1, 'p1', 'magic')
    expect(focusSpy).toHaveBeenNthCalledWith(2, 'p1', 'skill')
    expect(sendEnvelope).toHaveBeenNthCalledWith(1, {
      Cmd: 'SubmitAction',
      Data: {
        action_type: 'Magic',
        used_card_uuids: ['magic-1'],
        targets: [{ target_user_id: 'p2' }],
        target_ref: 'p2',
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

  it('wraps lobby commands into RoomAction envelopes', () => {
    const { client, battleReviewStore, sendEnvelope } = buildClient()

    client.sendRoomAction('select_character', {
      camp: 'Red',
      char_role: 'mage',
    })

    expect(sendEnvelope).toHaveBeenCalledWith({
      Cmd: 'RoomAction',
      Data: {
        action: 'select_character',
        camp: 'Red',
        char_role: 'mage',
      },
    })
    expect(battleReviewStore.logs[battleReviewStore.logs.length - 1]).toBe(
      '[WS][TX] RoomAction: {"action":"select_character","camp":"Red","char_role":"mage"}'
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
        target_ref: 'discard',
        extra_args: ['p2', '3'],
      },
    })
  })
})
