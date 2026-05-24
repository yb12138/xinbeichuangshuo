import { defineStore, storeToRefs } from 'pinia'
import { ref } from 'vue'
import type { Card } from '../types/game'
import { useSessionStore } from './session.store'
import { useSnapshotStore } from './snapshot.store'
import { useUiStore } from './ui.store'

export type InitiatorFocusMode = 'attack' | 'magic' | 'skill' | 'turn' | 'response'
export type InitiatorFocusSide = 'left' | 'right'

export interface InitiatorFocusState {
  playerId: string
  side: InitiatorFocusSide
  mode: InitiatorFocusMode
  startedAt: number
}

export interface FlyingCardView {
  id: number
  cards: Card[]
  playerId: string
  playerName: string
  actionType: string
  holdMode: 'timed' | 'until_response' | 'until_next_card_or_draw'
  hidden?: boolean
}

interface FlyingCardQueueItem {
  cards: Card[]
  playerId: string
  playerName: string
  actionType: string
  holdMode: 'timed' | 'until_response' | 'until_next_card_or_draw'
  hidden?: boolean
}

export interface DrawBurstView {
  id: number
  playerId: string
  playerName: string
  count: number
}

export interface DamageEffectView {
  id: number
  targetId: string
  targetName: string
  damage: number
  damageType: string
}

export interface CombatCueView {
  id: number
  attackerId: string
  targetId: string
  phase: 'attack' | 'defend' | 'take' | 'counter' | 'shield'
}

interface CombatCueQueueItem {
  attackerId: string
  targetId: string
  phase: 'attack' | 'defend' | 'take' | 'counter' | 'shield'
}

export const useBattleFxStore = defineStore('battlefx', () => {
  const sessionStore = useSessionStore()
  const snapshotStore = useSnapshotStore()
  const uiStore = useUiStore()
  const { roomPlayers, myPlayerId, myCamp } = storeToRefs(sessionStore)
  const { players } = storeToRefs(snapshotStore)
  const { cinematicMode } = storeToRefs(uiStore)

  const flyingCards = ref<FlyingCardView[]>([])
  let flyingCardsId = 0
  const flyingCardsQueue = ref<FlyingCardQueueItem[]>([])
  let flyingCardsTimer: ReturnType<typeof setTimeout> | null = null
  const battlefieldRevealClearToken = ref(0)

  const drawBursts = ref<DrawBurstView[]>([])
  let drawBurstId = 0
  const drawBurstTimers = new Map<number, ReturnType<typeof setTimeout>>()

  const combatCue = ref<CombatCueView | null>(null)
  let combatCueId = 0
  const combatCueQueue = ref<CombatCueQueueItem[]>([])
  let combatCueTimer: ReturnType<typeof setTimeout> | null = null

  const initiatorFocus = ref<InitiatorFocusState | null>(null)
  let initiatorFocusIdleTimer: ReturnType<typeof setTimeout> | null = null
  let initiatorFocusResolveTimer: ReturnType<typeof setTimeout> | null = null

  const damageEffects = ref<DamageEffectView[]>([])
  let damageEffectsId = 0

  function resolveInitiatorFocusSide(playerId: string): InitiatorFocusSide {
    const rosterIndex = roomPlayers.value.findIndex((player) => player.id === playerId)
    if (rosterIndex >= 0) {
      return rosterIndex < 3 ? 'left' : 'right'
    }

    const actorCamp = players.value[playerId]?.camp
    if ((myCamp.value === 'Red' || myCamp.value === 'Blue') && (actorCamp === 'Red' || actorCamp === 'Blue')) {
      return actorCamp === myCamp.value ? 'left' : 'right'
    }
    if (playerId === myPlayerId.value) return 'left'
    return 'right'
  }

  function cancelInitiatorFocusIdleTimer() {
    if (!initiatorFocusIdleTimer) return
    clearTimeout(initiatorFocusIdleTimer)
    initiatorFocusIdleTimer = null
  }

  function cancelInitiatorFocusResolveTimer() {
    if (!initiatorFocusResolveTimer) return
    clearTimeout(initiatorFocusResolveTimer)
    initiatorFocusResolveTimer = null
  }

  function clearInitiatorFocus() {
    cancelInitiatorFocusIdleTimer()
    cancelInitiatorFocusResolveTimer()
    initiatorFocus.value = null
  }

  function setInitiatorFocus(playerId: string, mode: InitiatorFocusMode) {
    if (!playerId) return
    cancelInitiatorFocusResolveTimer()
    initiatorFocus.value = {
      playerId,
      side: resolveInitiatorFocusSide(playerId),
      mode,
      startedAt: Date.now(),
    }
  }

  function armSkillFocusIdleTimer() {
    cancelInitiatorFocusIdleTimer()
    const idleMs = cinematicMode.value ? 8200 : 6200
    initiatorFocusIdleTimer = setTimeout(() => {
      if (initiatorFocus.value && initiatorFocus.value.mode !== 'attack') {
        initiatorFocus.value = null
      }
      initiatorFocusIdleTimer = null
    }, idleMs)
  }

  function startAttackInitiatorFocus(attackerId: string) {
    if (!attackerId) return
    setInitiatorFocus(attackerId, 'attack')
    cancelInitiatorFocusIdleTimer()
  }

  function resolveAttackInitiatorFocus(attackerId: string, delayMs?: number) {
    const focus = initiatorFocus.value
    if (!focus || focus.mode !== 'attack' || focus.playerId !== attackerId) return
    cancelInitiatorFocusResolveTimer()
    const holdMs = delayMs ?? (cinematicMode.value ? 820 : 460)
    initiatorFocusResolveTimer = setTimeout(() => {
      if (initiatorFocus.value?.mode === 'attack' && initiatorFocus.value.playerId === attackerId) {
        initiatorFocus.value = null
      }
      initiatorFocusResolveTimer = null
    }, holdMs)
  }

  function startSkillInitiatorFocus(playerId: string, mode: 'magic' | 'skill' = 'skill') {
    if (!playerId) return
    setInitiatorFocus(playerId, mode)
    armSkillFocusIdleTimer()
  }

  function startActingPlayerFocus(playerId: string, mode: 'turn' | 'response' | 'magic' | 'skill' = 'turn') {
    if (!playerId) return
    setInitiatorFocus(playerId, mode)
    cancelInitiatorFocusIdleTimer()
  }

  function touchSkillInitiatorFocus(playerId?: string) {
    const focus = initiatorFocus.value
    if (!focus || focus.mode === 'attack') return
    if (playerId && focus.playerId !== playerId) return
    armSkillFocusIdleTimer()
  }

  function settleSkillInitiatorFocus(playerId?: string, delayMs?: number) {
    const focus = initiatorFocus.value
    if (!focus || focus.mode === 'attack') return
    if (playerId && focus.playerId !== playerId) return
    cancelInitiatorFocusIdleTimer()
    cancelInitiatorFocusResolveTimer()
    const holdMs = delayMs ?? (cinematicMode.value ? 1080 : 700)
    const expectedPlayerId = focus.playerId
    const expectedMode = focus.mode
    initiatorFocusResolveTimer = setTimeout(() => {
      if (
        initiatorFocus.value?.playerId === expectedPlayerId &&
        initiatorFocus.value?.mode === expectedMode
      ) {
        initiatorFocus.value = null
      }
      initiatorFocusResolveTimer = null
    }, holdMs)
  }

  function prepareForFlowUpdate(nextCombatStage?: string, nextSubflow?: string) {
    const inCombat = !!nextCombatStage || nextSubflow === 'Response'
    if (!inCombat && combatCueQueue.value.length === 0) {
      combatCue.value = null
      if (initiatorFocus.value?.mode === 'attack') {
        resolveAttackInitiatorFocus(initiatorFocus.value.playerId, cinematicMode.value ? 260 : 160)
      }
    }
  }

  function syncInitiatorFocusWithState(nextCombatStage?: string, nextSubflow?: string) {
    const focus = initiatorFocus.value
    if (!focus) return
    const nextSide = resolveInitiatorFocusSide(focus.playerId)
    if (focus.side !== nextSide) {
      initiatorFocus.value = { ...focus, side: nextSide }
    }

    const inCombat = !!nextCombatStage || nextSubflow === 'Response'

    if (focus.mode === 'attack') {
      if (!inCombat) {
        resolveAttackInitiatorFocus(focus.playerId, cinematicMode.value ? 260 : 160)
      }
      return
    }

    if (nextSubflow === 'Response' || inCombat) {
      touchSkillInitiatorFocus(focus.playerId)
      return
    }

    if (damageEffects.value.length > 0) {
      touchSkillInitiatorFocus(focus.playerId)
      return
    }

    settleSkillInitiatorFocus(focus.playerId, cinematicMode.value ? 420 : 240)
  }

  function resolveFlyingHoldMode(actionType: string): 'timed' | 'until_response' | 'until_next_card_or_draw' {
    if (normalizeFlyingActionType(actionType) !== 'discard') return 'until_response'
    return 'until_next_card_or_draw'
  }

  function normalizeFlyingActionType(actionType?: string) {
    return String(actionType || '').trim().toLowerCase()
  }

  function dropActiveFlyingCards() {
    if (flyingCards.value.length === 0) return
    flyingCards.value = []
    if (flyingCardsTimer) {
      clearTimeout(flyingCardsTimer)
      flyingCardsTimer = null
    }
  }

  function notifyFlyingCardsEvent(kind: 'card_revealed' | 'draw' | 'combat_response' | 'damage', actionType?: string) {
    const normalizedActionType = normalizeFlyingActionType(actionType)
    if (kind === 'card_revealed' && (normalizedActionType === 'defend' || normalizedActionType === 'counter')) {
      return
    }

    if (kind === 'damage' || kind === 'combat_response') {
      dropActiveFlyingCards()
      pumpFlyingCards()
      return
    }

    if (kind === 'draw' || kind === 'card_revealed') {
      dropActiveFlyingCards()
      pumpFlyingCards()
    }
  }

  function addFlyingCards(cards: Card[], playerId: string, playerName: string, actionType: string, hidden?: boolean) {
    if (!cards?.length) return
    const normalizedActionType = normalizeFlyingActionType(actionType)
    if (normalizedActionType === 'discard' && hidden) return
    notifyFlyingCardsEvent('card_revealed', normalizedActionType)
    flyingCardsQueue.value.push({
      cards,
      playerId,
      playerName,
      actionType: normalizedActionType,
      holdMode: resolveFlyingHoldMode(normalizedActionType),
      hidden,
    })
    pumpFlyingCards()
  }

  function clearBattlefieldReveals() {
    battlefieldRevealClearToken.value += 1
  }

  function pumpFlyingCards() {
    if (flyingCards.value.length > 0 || flyingCardsQueue.value.length === 0) return
    const next = flyingCardsQueue.value.shift()
    if (!next) return

    flyingCardsId++
    const id = flyingCardsId
    if (next.holdMode === 'until_response') {
      flyingCards.value = [...flyingCards.value, { id, ...next }]
    } else {
      flyingCards.value = [{ id, ...next }]
    }

    if (next.holdMode === 'timed') {
      const displayMs = cinematicMode.value ? 2400 : 1600
      if (flyingCardsTimer) clearTimeout(flyingCardsTimer)
      flyingCardsTimer = setTimeout(() => {
        flyingCards.value = flyingCards.value.filter((item) => item.id !== id)
        flyingCardsTimer = null
        pumpFlyingCards()
      }, displayMs)
    }
  }

  function addDrawBurst(playerId: string, playerName: string, count: number) {
    if (!playerId || count <= 0) return
    notifyFlyingCardsEvent('draw')
    drawBurstId++
    const id = drawBurstId
    drawBursts.value.push({
      id,
      playerId,
      playerName,
      count,
    })
    const durationMs = cinematicMode.value ? 1850 : 1050
    const timer = setTimeout(() => {
      drawBursts.value = drawBursts.value.filter((item) => item.id !== id)
      drawBurstTimers.delete(id)
    }, durationMs)
    drawBurstTimers.set(id, timer)
  }

  function addDamageEffect(targetId: string, targetName: string, damage: number, damageType: string) {
    if (damage <= 0) return
    notifyFlyingCardsEvent('damage')
    touchSkillInitiatorFocus()
    damageEffectsId++
    const id = damageEffectsId
    damageEffects.value.push({
      id,
      targetId,
      targetName,
      damage,
      damageType,
    })
    setTimeout(() => {
      damageEffects.value = damageEffects.value.filter((item) => item.id !== id)
    }, 1500)
  }

  function addCombatCue(attackerId: string, targetId: string, phase: 'attack' | 'defend' | 'take' | 'counter' | 'shield') {
    if (!attackerId || !targetId) return
    if (phase === 'defend' || phase === 'take' || phase === 'counter' || phase === 'shield') {
      notifyFlyingCardsEvent('combat_response')
      resolveAttackInitiatorFocus(attackerId)
      startActingPlayerFocus(targetId, 'response')
    } else if (phase === 'attack') {
      startAttackInitiatorFocus(attackerId)
    }

    if (phase === 'attack') {
      if (combatCueTimer) {
        clearTimeout(combatCueTimer)
        combatCueTimer = null
      }
      combatCueQueue.value = []
      combatCueId++
      combatCue.value = {
        id: combatCueId,
        attackerId,
        targetId,
        phase,
      }
      return
    }

    if (
      combatCue.value &&
      combatCue.value.attackerId === attackerId &&
      combatCue.value.targetId === targetId &&
      combatCue.value.phase === 'attack'
    ) {
      if (combatCueTimer) clearTimeout(combatCueTimer)
      combatCueId++
      const id = combatCueId
      combatCue.value = {
        id,
        attackerId,
        targetId,
        phase,
      }
      const displayMs = cinematicMode.value ? 2600 : 1500
      combatCueTimer = setTimeout(() => {
        if (combatCue.value?.id === id) {
          combatCue.value = null
        }
        combatCueTimer = null
        pumpCombatCue()
      }, displayMs)
      return
    }

    combatCueQueue.value.push({
      attackerId,
      targetId,
      phase,
    })
    pumpCombatCue()
  }

  function pumpCombatCue() {
    if (combatCue.value || combatCueQueue.value.length === 0) return
    const next = combatCueQueue.value.shift()
    if (!next) return
    combatCueId++
    const id = combatCueId
    combatCue.value = { id, ...next }

    const displayMs = cinematicMode.value ? 1900 : 1000
    if (combatCueTimer) clearTimeout(combatCueTimer)
    combatCueTimer = setTimeout(() => {
      if (combatCue.value?.id === id) {
        combatCue.value = null
      }
      combatCueTimer = null
      pumpCombatCue()
    }, displayMs)
  }

  function clearForGameEnd() {
    clearBattlefieldReveals()
    drawBursts.value = []
    for (const timer of drawBurstTimers.values()) clearTimeout(timer)
    drawBurstTimers.clear()
    combatCue.value = null
    combatCueQueue.value = []
    if (combatCueTimer) {
      clearTimeout(combatCueTimer)
      combatCueTimer = null
    }
    clearInitiatorFocus()
  }

  function reset() {
    clearBattlefieldReveals()
    flyingCards.value = []
    flyingCardsQueue.value = []
    if (flyingCardsTimer) {
      clearTimeout(flyingCardsTimer)
      flyingCardsTimer = null
    }
    drawBursts.value = []
    for (const timer of drawBurstTimers.values()) clearTimeout(timer)
    drawBurstTimers.clear()
    combatCue.value = null
    combatCueQueue.value = []
    if (combatCueTimer) {
      clearTimeout(combatCueTimer)
      combatCueTimer = null
    }
    damageEffects.value = []
    clearInitiatorFocus()
  }

  return {
    flyingCards,
    battlefieldRevealClearToken,
    drawBursts,
    combatCue,
    initiatorFocus,
    damageEffects,
    startAttackInitiatorFocus,
    resolveAttackInitiatorFocus,
    startSkillInitiatorFocus,
    startActingPlayerFocus,
    touchSkillInitiatorFocus,
    settleSkillInitiatorFocus,
    prepareForFlowUpdate,
    syncInitiatorFocusWithState,
    addFlyingCards,
    clearBattlefieldReveals,
    addDrawBurst,
    addDamageEffect,
    addCombatCue,
    clearInitiatorFocus,
    clearForGameEnd,
    reset,
  }
})
