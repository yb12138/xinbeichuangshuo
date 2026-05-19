import type { PlayerAction } from '../types/game'
import { useBattleFxStore } from '../stores/battlefx.store'
import { useBattleReviewStore } from '../stores/battleReview.store'
import { useInterruptStore } from '../stores/interrupt.store'
import { useSessionStore } from '../stores/session.store'
import { buildClientActionRequest } from './actionRequestAdapter'
import type { ClientActionRequest, RoomActionRequest, WsMessage } from './protocol'

export interface WsCommandClientDeps {
  interruptStore: ReturnType<typeof useInterruptStore>
  sessionStore: ReturnType<typeof useSessionStore>
  battleFxStore: ReturnType<typeof useBattleFxStore>
  battleReviewStore: ReturnType<typeof useBattleReviewStore>
  isTransportOpen: () => boolean
  sendEnvelope: (msg: WsMessage) => void
  safeStringify: (data: unknown) => string
}

export function createWsCommandClient(deps: WsCommandClientDeps) {
  const {
    interruptStore,
    sessionStore,
    battleFxStore,
    battleReviewStore,
    isTransportOpen,
    sendEnvelope,
    safeStringify,
  } = deps

  function sendAction(action: PlayerAction) {
    if (!isTransportOpen()) {
      interruptStore.showError('未连接到服务器')
      return
    }

    const payload = buildClientActionRequest(action)
    const msg: WsMessage<ClientActionRequest> = {
      Cmd: 'SubmitAction',
      Data: payload,
    }

    if (action.type === 'Skill' && action.player_id) {
      battleFxStore.startSkillInitiatorFocus(action.player_id, 'skill')
    }
    if (action.type === 'Magic' && action.player_id) {
      battleFxStore.startSkillInitiatorFocus(action.player_id, 'magic')
    }

    battleReviewStore.addLog(`[WS][TX] SubmitAction: ${safeStringify(payload)}`)
    sendEnvelope(msg)
  }

  function sendRoomAction(action: string, data?: Record<string, unknown>) {
    if (!isTransportOpen()) {
      interruptStore.showError('未连接到服务器')
      return
    }

    const payload: RoomActionRequest = {
      action,
      ...(data || {}),
    }
    const msg: WsMessage<RoomActionRequest> = {
      Cmd: 'RoomAction',
      Data: payload,
    }

    battleReviewStore.addLog(`[WS][TX] RoomAction: ${safeStringify(payload)}`)
    sendEnvelope(msg)
  }

  function sendChat(message: string) {
    if (!isTransportOpen()) {
      return
    }

    const msg: WsMessage<Record<string, string>> = {
      Cmd: 'ChatMessage',
      Data: { message },
    }

    battleReviewStore.addLog(`[WS][TX] ChatMessage: ${safeStringify(msg.Data)}`)
    sendEnvelope(msg)
  }

  function attack(targetId: string, cardID: string) {
    const action: PlayerAction = {
      player_id: sessionStore.myPlayerId,
      type: 'Attack',
      target_id: targetId,
      card_id: cardID,
    }
    sendAction(action)
  }

  function magic(targetId: string | undefined, cardID: string) {
    const action: PlayerAction = {
      player_id: sessionStore.myPlayerId,
      type: 'Magic',
      card_id: cardID,
    }
    if (targetId) action.target_id = targetId
    sendAction(action)
  }

  function useSkill(skillId: string, targetIds: string[] = [], selections?: number[]) {
    const action: PlayerAction = {
      player_id: sessionStore.myPlayerId,
      type: 'Skill',
      skill_id: skillId,
    }
    if (targetIds.length > 0) action.target_ids = targetIds
    if (selections && selections.length > 0) action.selections = selections
    sendAction(action)
  }

  function pass() {
    sendAction({
      player_id: sessionStore.myPlayerId,
      type: 'Pass',
    })
  }

  function confirm() {
    sendAction({
      player_id: sessionStore.myPlayerId,
      type: 'Confirm',
    })
  }

  function cancel() {
    sendAction({
      player_id: sessionStore.myPlayerId,
      type: 'Cancel',
    })
  }

  function select(selections: number[]) {
    sendAction({
      player_id: sessionStore.myPlayerId,
      type: 'Select',
      selections,
    })
  }

  function respond(action: string, cardID?: string, targetId?: string) {
    const payload: PlayerAction = {
      player_id: sessionStore.myPlayerId,
      type: 'Respond',
      extra_args: [action],
    }
    if (cardID) payload.card_id = cardID
    if (targetId) payload.target_id = targetId
    sendAction(payload)
  }

  function buy() {
    sendAction({
      player_id: sessionStore.myPlayerId,
      type: 'Buy',
    })
  }

  function extract() {
    sendAction({
      player_id: sessionStore.myPlayerId,
      type: 'Extract',
    })
  }

  function cheatSkill(playerId: string, roleId: string, skillId: string) {
    const pid = playerId || sessionStore.myPlayerId
    const args = roleId ? [pid, roleId, skillId] : [pid, skillId]
    sendAction({
      player_id: sessionStore.myPlayerId,
      type: 'Cheat',
      target_id: 'skill',
      extra_args: args,
    })
  }

  function cheatToken(playerId: string, tokenKey: string, value: number) {
    const pid = playerId || sessionStore.myPlayerId
    sendAction({
      player_id: sessionStore.myPlayerId,
      type: 'Cheat',
      target_id: 'token',
      extra_args: [pid, tokenKey, String(value)],
    })
  }

  function cheatSet(playerId: string, field: string, value: number) {
    const pid = playerId || sessionStore.myPlayerId
    sendAction({
      player_id: sessionStore.myPlayerId,
      type: 'Cheat',
      target_id: 'set',
      extra_args: [pid, field, String(value)],
    })
  }

  function cheatEffect(playerId: string, effect: string, count: number) {
    const pid = playerId || sessionStore.myPlayerId
    sendAction({
      player_id: sessionStore.myPlayerId,
      type: 'Cheat',
      target_id: 'effect',
      extra_args: [pid, effect, String(count)],
    })
  }

  function cheatGiveExclusive(playerId: string, roleId: string, skillId: string, count: number) {
    const pid = playerId || sessionStore.myPlayerId
    sendAction({
      player_id: sessionStore.myPlayerId,
      type: 'Cheat',
      target_id: 'card_exclusive',
      extra_args: [pid, roleId, skillId, String(count)],
    })
  }

  function cheatGiveByElement(playerId: string, element: string, count: number) {
    const pid = playerId || sessionStore.myPlayerId
    sendAction({
      player_id: sessionStore.myPlayerId,
      type: 'Cheat',
      target_id: 'card_element',
      extra_args: [pid, element, String(count)],
    })
  }

  function cheatGiveByFaction(playerId: string, faction: string, count: number) {
    const pid = playerId || sessionStore.myPlayerId
    sendAction({
      player_id: sessionStore.myPlayerId,
      type: 'Cheat',
      target_id: 'card_faction',
      extra_args: [pid, faction, String(count)],
    })
  }

  function cheatGiveMagicByName(playerId: string, cardName: string, count: number) {
    const pid = playerId || sessionStore.myPlayerId
    sendAction({
      player_id: sessionStore.myPlayerId,
      type: 'Cheat',
      target_id: 'card_magic',
      extra_args: [pid, cardName, String(count)],
    })
  }

  function cheatDiscard(playerId: string, count: number) {
    const pid = playerId || sessionStore.myPlayerId
    sendAction({
      player_id: sessionStore.myPlayerId,
      type: 'Cheat',
      target_id: 'discard',
      extra_args: [pid, String(count)],
    })
  }

  return {
    sendAction,
    sendRoomAction,
    sendChat,
    attack,
    magic,
    useSkill,
    pass,
    confirm,
    cancel,
    select,
    respond,
    buy,
    extract,
    cheatSkill,
    cheatToken,
    cheatSet,
    cheatEffect,
    cheatGiveExclusive,
    cheatGiveByElement,
    cheatGiveByFaction,
    cheatGiveMagicByName,
    cheatDiscard,
  }
}
