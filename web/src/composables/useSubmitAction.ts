import { useWebSocket } from './useWebSocket'
import { useInterruptStore } from '../stores/interrupt.store'
import { useSessionStore } from '../stores/session.store'
import { useBattleInteractionState } from './useBattleInteractionState'
import type { PlayerAction } from '../types/game'

type SkillSubmitOptions = {
  clearSkillMode?: boolean
}

export function useSubmitAction() {
  const ws = useWebSocket()
  const interruptStore = useInterruptStore()
  const sessionStore = useSessionStore()
  const { canTargetOpponent, myPlayableCards } = useBattleInteractionState()

  function myPlayerID() {
    return sessionStore.myPlayerId
  }

  function submitAction(action: PlayerAction) {
    ws.sendAction(action)
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
    ws.buy()
  }

  function submitExtract() {
    ws.extract()
  }

  function submitPass() {
    ws.pass()
  }

  function submitConfirm() {
    ws.confirm()
  }

  function submitCancel() {
    ws.cancel()
  }

  function submitSelect(selections: number[]) {
    ws.select(selections)
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
    ws.respond('take')
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
    ws.respond('counter', cardID, interruptStore.promptCounterTarget || undefined)
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
    ws.respond('defend', cardID)
    return true
  }

  function submitUseSkill(skillId: string, targetIds: string[] = [], selections?: number[], options?: SkillSubmitOptions) {
    if (options?.clearSkillMode) {
      interruptStore.clearSkillMode()
    }
    ws.useSkill(skillId, targetIds, selections)
    return true
  }

  function submitAttack(targetId: string, cardID: string) {
    ws.attack(targetId, cardID)
  }

  function submitMagic(targetId: string | undefined, cardID: string) {
    ws.magic(targetId, cardID)
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
      ws.attack(playerId, selectedItem.card.id)
      return true
    }

    if (interruptStore.actionMode === 'magic') {
      if (selectedItem.card.type !== 'Magic') {
        interruptStore.setSelectedHandIndexForAction(null)
        interruptStore.showError('所选卡牌不是法术牌，请重新选择')
        return false
      }
      if (selectedItem.card.name === '魔弹') {
        ws.magic(undefined, selectedItem.card.id)
        return true
      }
      ws.magic(playerId, selectedItem.card.id)
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
