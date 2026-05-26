import { defineStore, storeToRefs } from 'pinia'
import { ref } from 'vue'
import type { Card } from '../types/game'
import type { TimelineEvent, TimelineNotifyPayload } from '../network/protocol'
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

export type SkillAnnouncementPhase = 'featured' | 'settled'
export type NarrativeActorReason = 'turn' | 'attack' | 'response' | 'target' | 'skill' | 'damage'
export type NarrativeEventKind = 'turn' | 'attack' | 'magic' | 'respond' | 'take' | 'damage' | 'skill' | 'system'
export type NarrativeLinkKind = 'attack' | 'respond' | 'skill' | 'damage'

export interface SkillAnnouncementView {
  id: number
  actorId: string
  actorName: string
  skillName: string
  effectText: string
  phase: SkillAnnouncementPhase
  startedAt: number
}

export interface ActionNarrativeCardView {
  id: number
  timelineEventId?: number
  playerId: string
  targetId?: string
  card: Card
  actionType: string
  createdAt: number
}

export interface ActionNarrativeLinkView {
  id: number
  fromType: 'actor' | 'card'
  fromId: string | number
  toPlayerId: string
  kind: NarrativeLinkKind
  createdAt: number
}

export interface ActionNarrativeEventView {
  id: number
  timelineEventId?: number
  kind: NarrativeEventKind
  label: string
  actorId?: string
  targetId?: string
  damage?: number
  createdAt: number
}

export interface ActionNarrativeView {
  currentActionPlayerId: string
  featuredActorId: string
  featuredReason: NarrativeActorReason
  opposedActorIds: string[]
  settledActorIds: string[]
  playedCards: ActionNarrativeCardView[]
  links: ActionNarrativeLinkView[]
  events: ActionNarrativeEventView[]
  startedAt: number
}

interface CombatCueQueueItem {
  attackerId: string
  targetId: string
  phase: 'attack' | 'defend' | 'take' | 'counter' | 'shield'
}

interface SkillAnnouncementQueueItem {
  actorId: string
  actorName: string
  skillName: string
  effectText: string
}

export const useBattleFxStore = defineStore('battlefx', () => {
  const sessionStore = useSessionStore()
  const snapshotStore = useSnapshotStore()
  const uiStore = useUiStore()
  const { roomPlayers, myPlayerId, myCamp } = storeToRefs(sessionStore)
  const { players } = storeToRefs(snapshotStore)
  const { cinematicMode } = storeToRefs(uiStore)

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

  const skillAnnouncements = ref<SkillAnnouncementView[]>([])
  let skillAnnouncementId = 0
  const skillAnnouncementQueue = ref<SkillAnnouncementQueueItem[]>([])
  let skillAnnouncementTimer: ReturnType<typeof setTimeout> | null = null
  const skillAnnouncementRemovalTimers = new Map<number, ReturnType<typeof setTimeout>>()

  const actionNarrative = ref<ActionNarrativeView | null>(null)
  let narrativeCardId = 0
  let narrativeLinkId = 0
  let narrativeEventId = 0
  const narrativeTargetByActor = new Map<string, string>()
  const structuredNarrativeEventIds = new Set<number>()

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

  function clearSkillAnnouncementTimer() {
    if (!skillAnnouncementTimer) return
    clearTimeout(skillAnnouncementTimer)
    skillAnnouncementTimer = null
  }

  function clearSkillAnnouncementRemovalTimers() {
    for (const timer of skillAnnouncementRemovalTimers.values()) {
      clearTimeout(timer)
    }
    skillAnnouncementRemovalTimers.clear()
  }

  function normalizeNarrativeActionType(actionType?: string) {
    return String(actionType || '').trim().toLowerCase()
  }

  function uniqueIds(ids: string[]) {
    return [...new Set(ids.filter(Boolean))]
  }

  function trimNarrative(next: ActionNarrativeView): ActionNarrativeView {
    return {
      ...next,
      settledActorIds: next.settledActorIds.slice(-5),
      opposedActorIds: next.opposedActorIds.slice(-2),
      playedCards: next.playedCards,
      links: next.links,
      events: next.events,
    }
  }

  function ensureActionNarrative(playerId: string, reason: NarrativeActorReason = 'turn') {
    if (!playerId) return null
    if (!actionNarrative.value) {
      actionNarrative.value = {
        currentActionPlayerId: playerId,
        featuredActorId: playerId,
        featuredReason: reason,
        opposedActorIds: [],
        settledActorIds: [],
        playedCards: [],
        links: [],
        events: [],
        startedAt: Date.now(),
      }
      return actionNarrative.value
    }
    return actionNarrative.value
  }

  function beginActionNarrative(playerId: string) {
    if (!playerId) return
    if (actionNarrative.value?.currentActionPlayerId === playerId) {
      actionNarrative.value = trimNarrative({
        ...actionNarrative.value,
        featuredActorId: actionNarrative.value.featuredActorId || playerId,
        featuredReason: actionNarrative.value.featuredReason || 'turn',
      })
      return
    }
    narrativeTargetByActor.clear()
    actionNarrative.value = {
      currentActionPlayerId: playerId,
      featuredActorId: playerId,
      featuredReason: 'turn',
      opposedActorIds: [],
      settledActorIds: [],
      playedCards: [],
      links: [],
      events: [],
      startedAt: Date.now(),
    }
    addNarrativeEvent({ kind: 'turn', label: '行动回合', actorId: playerId })
  }

  function settleNarrativeActor(playerId: string) {
    const current = actionNarrative.value
    if (!current || !playerId) return
    actionNarrative.value = trimNarrative({
      ...current,
      opposedActorIds: current.opposedActorIds.filter(id => id !== playerId),
      settledActorIds: uniqueIds([...current.settledActorIds, playerId]),
    })
  }

  function featureNarrativeActor(playerId: string, reason: NarrativeActorReason = 'turn') {
    const current = ensureActionNarrative(playerId, reason)
    if (!current || !playerId) return
    const prevFeatured = current.featuredActorId
    const settled = prevFeatured && prevFeatured !== playerId
      ? uniqueIds([...current.settledActorIds, prevFeatured])
      : current.settledActorIds
    actionNarrative.value = trimNarrative({
      ...current,
      featuredActorId: playerId,
      featuredReason: reason,
      opposedActorIds: current.opposedActorIds.filter(id => id !== playerId),
      settledActorIds: settled.filter(id => id !== playerId),
    })
  }

  function opposeNarrativeActor(playerId: string) {
    const current = actionNarrative.value
    if (!current || !playerId || playerId === current.featuredActorId) return
    actionNarrative.value = trimNarrative({
      ...current,
      opposedActorIds: uniqueIds([...current.opposedActorIds.filter(id => id !== playerId), playerId]),
      settledActorIds: current.settledActorIds.filter(id => id !== playerId),
    })
  }

  function addNarrativeLink(
    from: { type: 'actor' | 'card'; id: string | number },
    toPlayerId: string,
    kind: NarrativeLinkKind,
  ) {
    const current = actionNarrative.value
    if (!current || !toPlayerId) return
    narrativeLinkId++
    actionNarrative.value = trimNarrative({
      ...current,
      links: [
        ...current.links,
        {
          id: narrativeLinkId,
          fromType: from.type,
          fromId: from.id,
          toPlayerId,
          kind,
          createdAt: Date.now(),
        },
      ],
    })
  }

  function eventLabelForCombatPhase(phase: CombatCueView['phase']) {
    if (phase === 'attack') return '攻击'
    if (phase === 'counter') return '应战'
    if (phase === 'defend') return '防御'
    if (phase === 'shield') return '圣盾'
    return '命中'
  }

  function eventKindForCombatPhase(phase: CombatCueView['phase']): NarrativeEventKind {
    if (phase === 'attack') return 'attack'
    if (phase === 'counter' || phase === 'defend' || phase === 'shield') return 'respond'
    return 'take'
  }

  function linkKindForActionType(actionType: string): NarrativeLinkKind {
    const normalized = normalizeNarrativeActionType(actionType)
    if (normalized === 'counter' || normalized === 'defend' || normalized === 'shield') return 'respond'
    if (normalized === 'skill' || normalized === 'magic') return 'skill'
    return 'attack'
  }

  function addNarrativeEvent(event: Omit<ActionNarrativeEventView, 'id' | 'createdAt'> & { createdAt?: number }) {
    const actorId = event.actorId || actionNarrative.value?.featuredActorId || actionNarrative.value?.currentActionPlayerId
    const current = ensureActionNarrative(actorId || event.targetId || '', event.kind === 'skill' ? 'skill' : 'turn')
    if (!current) return
    narrativeEventId++
    const { createdAt, ...rest } = event
    actionNarrative.value = trimNarrative({
      ...current,
      events: [
        ...current.events,
        {
          ...rest,
          id: narrativeEventId,
          createdAt: createdAt ?? Date.now(),
        },
      ],
    })
  }

  function addNarrativeCard(playerId: string, card: Card, actionType: string, targetId?: string, options?: { createdAt?: number; timelineEventId?: number }) {
    const normalizedActionType = normalizeNarrativeActionType(actionType)
    if (!playerId || !card || !['attack', 'magic', 'counter', 'defend', 'shield'].includes(normalizedActionType)) return
    const current = ensureActionNarrative(playerId, normalizedActionType === 'magic' ? 'skill' : 'attack')
    if (!current) return
    const resolvedTargetId = targetId || narrativeTargetByActor.get(playerId)
    narrativeCardId++
    const cardView: ActionNarrativeCardView = {
      id: narrativeCardId,
      playerId,
      targetId: resolvedTargetId,
      card,
      actionType: normalizedActionType,
      createdAt: options?.createdAt ?? Date.now(),
      timelineEventId: options?.timelineEventId,
    }
    actionNarrative.value = trimNarrative({
      ...current,
      playedCards: [...current.playedCards, cardView],
    })
    if (resolvedTargetId) {
      addNarrativeLink({ type: 'card', id: cardView.id }, resolvedTargetId, linkKindForActionType(normalizedActionType))
    }
  }

  function addNarrativeCombatCue(attackerId: string, targetId: string, phase: CombatCueView['phase']) {
    if (!attackerId || !targetId) return
    if (!actionNarrative.value) {
      beginActionNarrative(attackerId)
    }
    narrativeTargetByActor.set(attackerId, targetId)
    if (phase === 'attack') {
      featureNarrativeActor(attackerId, 'attack')
      opposeNarrativeActor(targetId)
      addNarrativeLink({ type: 'actor', id: attackerId }, targetId, 'attack')
    } else {
      featureNarrativeActor(attackerId, 'response')
      opposeNarrativeActor(targetId)
      addNarrativeLink({ type: 'actor', id: attackerId }, targetId, phase === 'take' ? 'damage' : 'respond')
    }
    addNarrativeEvent({
      kind: eventKindForCombatPhase(phase),
      label: eventLabelForCombatPhase(phase),
      actorId: attackerId,
      targetId,
    })
  }

  function addNarrativeDamage(sourceId: string | undefined, targetId: string, damage: number, damageType?: string) {
    if (!targetId || damage <= 0) return
    const actorId = sourceId || actionNarrative.value?.featuredActorId || targetId
    ensureActionNarrative(actorId, 'damage')
    opposeNarrativeActor(targetId)
    addNarrativeEvent({
      kind: 'damage',
      label: `造成 ${damage} 点伤害`,
      actorId,
      targetId,
      damage,
    })
    if (actorId && actorId !== targetId) {
      addNarrativeLink({ type: 'actor', id: actorId }, targetId, 'damage')
    }
    void damageType
  }

  function addNarrativeSkill(actorId: string, skillName: string, effectText?: string, targetIds: string[] = []) {
    if (!actorId || !skillName) return
    featureNarrativeActor(actorId, 'skill')
    for (const targetId of targetIds) {
      opposeNarrativeActor(targetId)
      addNarrativeLink({ type: 'actor', id: actorId }, targetId, 'skill')
    }
    addNarrativeEvent({
      kind: 'skill',
      label: effectText ? `${skillName}：${effectText}` : `发动「${skillName}」`,
      actorId,
      targetId: targetIds[0],
    })
  }

  function clearActionNarrative() {
    actionNarrative.value = null
    narrativeTargetByActor.clear()
  }

  function hasStructuredNarrative(events: TimelineEvent[] = []) {
    return events.some(event => !!event.narrative_kind || !!event.visual_kind || !!event.narrative_window_id)
  }

  function applyStructuredTimelineNarrative(payload: TimelineNotifyPayload) {
    const events = [...(payload.events || [])]
      .filter(event => !!event.narrative_kind || !!event.visual_kind || !!event.narrative_window_id)
      .sort((a, b) => (a.event_id || 0) - (b.event_id || 0))
    if (!events.length) return

    if (payload.is_replay) {
      clearActionNarrative()
      structuredNarrativeEventIds.clear()
    }

    for (const event of events) {
      const eventId = Number(event.event_id || 0)
      if (eventId && structuredNarrativeEventIds.has(eventId)) continue
      if (eventId) structuredNarrativeEventIds.add(eventId)

      const kind = String(event.narrative_kind || '').trim()
      const visualKind = String(event.visual_kind || '').trim()
      const actorId = event.actor_user_id || ''
      const targetIds = uniqueIds(event.target_user_ids || [])
      const createdAt = eventId || Date.now()

      if (kind === 'action_started') {
        if (actorId) beginActionNarrative(actorId)
        continue
      }

      if (kind === 'action_closed') {
        clearActionNarrative()
        continue
      }

      if (actorId) {
        const reason: NarrativeActorReason =
          kind.startsWith('skill') ? 'skill' :
          kind === 'damage_dealt' ? 'damage' :
          kind === 'combat_response' ? 'response' :
          kind === 'combat_declared' || kind === 'card_played' ? 'attack' :
          'turn'
        featureNarrativeActor(actorId, reason)
      }
      for (const targetId of targetIds) {
        opposeNarrativeActor(targetId)
      }

      if (kind === 'combat_declared' || kind === 'combat_response') {
        for (const targetId of targetIds) {
          narrativeTargetByActor.set(actorId, targetId)
        }
      }

      if (visualKind === 'card' && event.cards?.length && actorId) {
        const cardRole = String(event.card_role || event.action_type || '').trim().toLowerCase()
        const targetId = targetIds[0] || narrativeTargetByActor.get(actorId)
        for (const card of event.cards) {
          addNarrativeCard(actorId, card as Card, cardRole || 'attack', targetId, { createdAt, timelineEventId: eventId })
        }
        continue
      }

      if (visualKind === 'skill_token' && actorId) {
        const skillName = event.skill_name || event.summary || '技能发动'
        for (const targetId of targetIds) {
          addNarrativeLink({ type: 'actor', id: actorId }, targetId, 'skill')
        }
        addNarrativeEvent({
          timelineEventId: eventId,
          kind: 'skill',
          label: event.effect_text ? `${skillName}：${event.effect_text}` : `发动「${skillName}」`,
          actorId,
          targetId: targetIds[0],
          createdAt,
        })
        continue
      }

      if ((visualKind === 'effect_token' || visualKind === 'action_marker') && actorId) {
        const title = event.field_card?.card?.name || event.summary || event.extra_action_type || event.effect_type || '效果'
        for (const targetId of targetIds) {
          addNarrativeLink({ type: 'actor', id: actorId }, targetId, 'skill')
        }
        addNarrativeEvent({
          timelineEventId: eventId,
          kind: 'skill',
          label: visualKind === 'action_marker' ? `获得额外行动：${title}` : `效果：${title}`,
          actorId,
          targetId: targetIds[0],
          createdAt,
        })
        continue
      }

      if (kind === 'damage_dealt' && targetIds[0] && event.damage) {
        addNarrativeDamage(actorId, targetIds[0], event.damage, event.damage_type)
      }
    }
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

  function addSkillAnnouncement(actorId: string, actorName: string, skillName: string, effectText: string) {
    const normalizedActorId = String(actorId || '').trim()
    const normalizedSkillName = String(skillName || '').trim()
    const normalizedEffectText = String(effectText || '').trim()
    if (!normalizedActorId || !normalizedSkillName) return
    const isGenericEffect = (text: string) => text === '技能发动' || text === '技能效果生效'
    const shouldReplaceEffect = (current: string, next: string) => !!next && (isGenericEffect(current) || next.length > current.length)

    const active = skillAnnouncements.value.find((item) =>
      item.actorId === normalizedActorId &&
      item.skillName === normalizedSkillName &&
      Date.now() - item.startedAt < 3200
    )
    if (active) {
      active.actorName = actorName || normalizedActorId
      if (shouldReplaceEffect(active.effectText, normalizedEffectText)) {
        active.effectText = normalizedEffectText
      }
      return
    }

    const queued = skillAnnouncementQueue.value.find((item) =>
      item.actorId === normalizedActorId &&
      item.skillName === normalizedSkillName
    )
    if (queued) {
      queued.actorName = actorName || normalizedActorId
      if (shouldReplaceEffect(queued.effectText, normalizedEffectText)) {
        queued.effectText = normalizedEffectText
      }
      return
    }

    skillAnnouncementQueue.value.push({
      actorId: normalizedActorId,
      actorName: actorName || normalizedActorId,
      skillName: normalizedSkillName,
      effectText: normalizedEffectText || '技能效果生效',
    })
    pumpSkillAnnouncements()
  }

  function removeSkillAnnouncement(id: number) {
    skillAnnouncements.value = skillAnnouncements.value.filter((item) => item.id !== id)
    const timer = skillAnnouncementRemovalTimers.get(id)
    if (timer) {
      clearTimeout(timer)
      skillAnnouncementRemovalTimers.delete(id)
    }
  }

  function pumpSkillAnnouncements() {
    if (skillAnnouncements.value.some((item) => item.phase === 'featured')) return
    if (skillAnnouncementQueue.value.length === 0) return
    const next = skillAnnouncementQueue.value.shift()
    if (!next) return

    skillAnnouncementId++
    const id = skillAnnouncementId
    skillAnnouncements.value = [
      ...skillAnnouncements.value,
      {
        id,
        actorId: next.actorId,
        actorName: next.actorName,
        skillName: next.skillName,
        effectText: next.effectText,
        phase: 'featured',
        startedAt: Date.now(),
      },
    ]

    const featuredMs = cinematicMode.value ? 1650 : 1100
    clearSkillAnnouncementTimer()
    skillAnnouncementTimer = setTimeout(() => {
      const item = skillAnnouncements.value.find((entry) => entry.id === id)
      if (!item) {
        skillAnnouncementTimer = null
        pumpSkillAnnouncements()
        return
      }
      item.phase = 'settled'
      skillAnnouncementTimer = null

      const removalMs = cinematicMode.value ? 7000 : 5200
      const removalTimer = setTimeout(() => {
        removeSkillAnnouncement(id)
      }, removalMs)
      skillAnnouncementRemovalTimers.set(id, removalTimer)

      pumpSkillAnnouncements()
    }, featuredMs)
  }

  function clearSkillAnnouncements() {
    skillAnnouncementQueue.value = []
    skillAnnouncements.value = []
    clearSkillAnnouncementTimer()
    clearSkillAnnouncementRemovalTimers()
  }

  function addDrawBurst(playerId: string, playerName: string, count: number) {
    if (!playerId || count <= 0) return
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
    addNarrativeCombatCue(attackerId, targetId, phase)
    if (phase === 'defend' || phase === 'take' || phase === 'counter' || phase === 'shield') {
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
    clearActionNarrative()
    clearSkillAnnouncements()
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
    clearActionNarrative()
    clearSkillAnnouncements()
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
    drawBursts,
    combatCue,
    initiatorFocus,
    damageEffects,
    skillAnnouncements,
    actionNarrative,
    beginActionNarrative,
    featureNarrativeActor,
    settleNarrativeActor,
    addNarrativeCard,
    addNarrativeLink,
    addNarrativeEvent,
    addNarrativeDamage,
    addNarrativeSkill,
    hasStructuredNarrative,
    applyStructuredTimelineNarrative,
    clearActionNarrative,
    startAttackInitiatorFocus,
    resolveAttackInitiatorFocus,
    startSkillInitiatorFocus,
    startActingPlayerFocus,
    touchSkillInitiatorFocus,
    settleSkillInitiatorFocus,
    prepareForFlowUpdate,
    syncInitiatorFocusWithState,
    addSkillAnnouncement,
    clearSkillAnnouncements,
    addDrawBurst,
    addDamageEffect,
    addCombatCue,
    clearInitiatorFocus,
    clearForGameEnd,
    reset,
  }
})
