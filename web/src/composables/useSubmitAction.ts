import { useWebSocket } from './useWebSocket'
import { useInterruptStore } from '../stores/interrupt.store'
import { useSessionStore } from '../stores/session.store'
import { useBattleInteractionState } from './useBattleInteractionState'
import type { PlayerAction } from '../types/game'

export function useSubmitAction() {
  const ws = useWebSocket()
  const interruptStore = useInterruptStore()
  const sessionStore = useSessionStore()
  const { canTargetOpponent, myPlayableCards } = useBattleInteractionState()

  function myPlayerID() {
    return sessionStore.myPlayerId
  }

  /**
   * 发出一次对局操作：先记下当前选择并进入"提交中"。
   * 选择不在发送前清空——成功时由随后的 state_update / prompt 清理，被拒绝时由 error 处理恢复。
   */
  function send(fn: () => void) {
    interruptStore.beginSubmit()
    fn()
  }

  function submitAction(action: PlayerAction) {
    send(() => ws.sendAction(action))
  }

  function submitCannotAct() {
    submitAction({
      player_id: myPlayerID(),
      type: 'CannotAct',
    })
  }

  function submitSynthesize() {
    submitAction({
      player_id: myPlayerID(),
      type: 'Synthesize',
    })
  }

  function submitBuy() {
    send(() => ws.buy())
  }

  function submitExtract() {
    send(() => ws.extract())
  }

  function submitPass() {
    send(() => ws.pass())
  }

  function submitConfirm() {
    send(() => ws.confirm())
  }

  function submitCancel() {
    send(() => ws.cancel())
  }

  function submitSelect(selections: number[]) {
    send(() => ws.select(selections))
  }

  function submitSelectCardIDs(cardIds: string[]) {
    submitAction({
      player_id: myPlayerID(),
      type: 'Select',
      card_ids: cardIds,
    })
  }

  function submitPromptTarget(playerId: string) {
    submitAction({
      player_id: myPlayerID(),
      type: 'Select',
      target_id: playerId,
    })
  }

  function submitRespondTake() {
    send(() => ws.respond('take'))
  }

  function selectedCardIDByPlayableIndex(playableIndex: number | undefined): string {
    if (playableIndex === undefined) return ''
    return String(myPlayableCards.value.find(item => item.index === playableIndex)?.card?.id || '').trim()
  }

  function submitRespondCounter(isMagicMissilePrompt = false) {
    if (interruptStore.selectedHandIndexes.length === 0) {
      interruptStore.showError(isMagicMissilePrompt ? '请先选择一张【魔弹】再传递' : '请先选择一张应战牌')
      return false
    }
    const cardID = selectedCardIDByPlayableIndex(interruptStore.selectedHandIndexes[0])
    if (!cardID) {
      interruptStore.showError('所选卡牌已变化，请重新选择')
      return false
    }
    const target = interruptStore.promptCounterTarget || undefined
    send(() => ws.respond('counter', cardID, target))
    return true
  }

  function submitRespondDefend() {
    if (interruptStore.selectedHandIndexes.length === 0) {
      interruptStore.showError('请先选择一张【圣光】进行防御（圣盾需提前放置）')
      return false
    }
    const cardID = selectedCardIDByPlayableIndex(interruptStore.selectedHandIndexes[0])
    if (!cardID) {
      interruptStore.showError('所选卡牌已变化，请重新选择')
      return false
    }
    send(() => ws.respond('defend', cardID))
    return true
  }

  /** 发动技能。已在等待上一次提交的回包时忽略，防止连点重复发动。 */
  function submitUseSkill(skillId: string, targetIds: string[] = [], selections?: number[]) {
    if (interruptStore.isSubmitting) return false
    send(() => ws.useSkill(skillId, targetIds, selections))
    return true
  }

  function submitAttack(targetId: string, cardID: string) {
    send(() => ws.attack(targetId, cardID))
  }

  function submitMagic(targetId: string | undefined, cardID: string) {
    send(() => ws.magic(targetId, cardID))
  }

  function submitSelectedBoardTarget(playerId: string) {
    if (!canTargetOpponent.value) {
      return false
    }

    const cardIdx = interruptStore.selectedHandIndexForAction
    if (cardIdx === null) {
      return false
    }

    const selectedItem = myPlayableCards.value.find(item => item.index === cardIdx)
    if (!selectedItem) {
      interruptStore.setSelectedHandIndexForAction(null)
      interruptStore.showError('所选卡牌已变化，请重新选择')
      return false
    }

    if (interruptStore.actionMode === 'attack') {
      if (selectedItem.card.type !== 'Attack') {
        interruptStore.setSelectedHandIndexForAction(null)
        interruptStore.showError('所选卡牌不是攻击牌，请重新选择')
        return false
      }
      submitAttack(playerId, selectedItem.card.id)
      return true
    }

    if (interruptStore.actionMode === 'magic') {
      if (selectedItem.card.type !== 'Magic') {
        interruptStore.setSelectedHandIndexForAction(null)
        interruptStore.showError('所选卡牌不是法术牌，请重新选择')
        return false
      }
      if (selectedItem.card.name === '魔弹') {
        submitMagic(undefined, selectedItem.card.id)
        return true
      }
      submitMagic(playerId, selectedItem.card.id)
      return true
    }

    return false
  }

  return {
    disconnect: ws.disconnect,
    sendChat: ws.sendChat,
    changeCamp: ws.changeCamp,
    changeRole: ws.changeRole,
    addBot: ws.addBot,
    removeBot: ws.removeBot,
    takeoverPlayer: ws.takeoverPlayer,
    startRoom: ws.startRoom,
    dissolveRoom: ws.dissolveRoom,
    submitAction,
    submitCannotAct,
    submitSynthesize,
    submitBuy,
    submitExtract,
    submitPass,
    submitConfirm,
    submitCancel,
    submitSelect,
    submitSelectCardIDs,
    submitPromptTarget,
    submitRespondTake,
    submitRespondCounter,
    submitRespondDefend,
    submitUseSkill,
    submitAttack,
    submitMagic,
    submitSelectedBoardTarget,
    cheatSkill: ws.cheatSkill,
    cheatToken: ws.cheatToken,
    cheatSet: ws.cheatSet,
    cheatEffect: ws.cheatEffect,
    cheatGiveExclusive: ws.cheatGiveExclusive,
    cheatGiveByElement: ws.cheatGiveByElement,
    cheatGiveByFaction: ws.cheatGiveByFaction,
    cheatGiveMagicByName: ws.cheatGiveMagicByName,
    cheatDiscard: ws.cheatDiscard,
  }
}
