import type { GameEvent, GameStateUpdate, Prompt } from '../types/game'
import { extractGameplayEventsFromTimeline } from './gameplayTimeline'
import type { RequireActionPayload, SyncStatePayload, TimelineNotifyPayload } from './protocol'
import { buildGameStateUpdateFromSyncState } from './syncState'
import { useBattleFxStore } from '../stores/battlefx.store'
import { useBattleReviewStore, type MoraleCamp } from '../stores/battleReview.store'
import { useInterruptStore } from '../stores/interrupt.store'
import { useMatchLifecycleStore } from '../stores/matchLifecycle.store'
import { useSessionStore } from '../stores/session.store'
import { useSnapshotStore } from '../stores/snapshot.store'
import { useTimelineStore } from '../stores/timeline.store'
import { useUiStore } from '../stores/ui.store'

export interface GameplayMessageHandlerDeps {
  interruptStore: ReturnType<typeof useInterruptStore>
  sessionStore: ReturnType<typeof useSessionStore>
  snapshotStore: ReturnType<typeof useSnapshotStore>
  timelineStore: ReturnType<typeof useTimelineStore>
  uiStore: ReturnType<typeof useUiStore>
  battleFxStore: ReturnType<typeof useBattleFxStore>
  battleReviewStore: ReturnType<typeof useBattleReviewStore>
  matchLifecycleStore: ReturnType<typeof useMatchLifecycleStore>
}

export function createGameplayMessageHandlers(deps: GameplayMessageHandlerDeps) {
  const {
    interruptStore,
    sessionStore,
    snapshotStore,
    timelineStore,
    uiStore,
    battleFxStore,
    battleReviewStore,
    matchLifecycleStore,
  } = deps

  const playerByName = (name?: string) => {
    if (!name) return undefined
    return Object.values(snapshotStore.players).find(p => p.name === name)
  }

  const playerNameById = (id?: string) => {
    if (!id) return ''
    return snapshotStore.players[id]?.name || id
  }

  const resolvePlayerIdByToken = (token?: string) => {
    if (!token) return ''
    const normalized = token.trim()
    if (!normalized) return ''
    if (snapshotStore.players[normalized]) return normalized
    const byName = playerByName(normalized)
    return byName?.id || ''
  }

  const normalizeCamp = (camp?: string): MoraleCamp | undefined => {
    if (camp === 'Red' || camp === 'Blue') return camp
    return undefined
  }

  const cleanLogMessage = (message: string) => {
    return String(message || '').replace(/^\[[^\]]+\]\s*/, '').trim()
  }

  const parseSkillAnnouncementFromLog = (message: string) => {
    const normalized = cleanLogMessage(message)
    if (!normalized) return null

    const directUse = message.match(/^\[Skill\]\s*(.+?)\s*使用了技能[:：]\s*(.+?)\s*$/)
    if (directUse) {
      return {
        actorToken: directUse[1]?.trim() || '',
        skillName: directUse[2]?.trim() || '',
        effectText: '技能发动',
      }
    }

    const bracketed = normalized.match(/^(.+?)\s*发动\s*\[([^\]]+)\]\s*[，,:：]?\s*(.*)$/)
    if (bracketed) {
      const effectText = bracketed[3]?.trim()
      return {
        actorToken: bracketed[1]?.trim() || '',
        skillName: bracketed[2]?.trim() || '',
        effectText: effectText || '技能效果生效',
      }
    }

    return null
  }

  const parseMoraleHintFromLog = (line: string) => {
    if (!line) return
    const normalized = line.replace(/^\[[^\]]+\]\s*/, '').trim()

    const discardLoss = line.match(/^\[System\]\s*(.+?)\s*丢弃了\s*(\d+)\s*张牌！士气\s*-(\d+)/)
    if (discardLoss) {
      const actorName = discardLoss[1]?.trim()
      const loss = Number(discardLoss[3] || 0)
      const actor = playerByName(actorName)
      battleReviewStore.pushMoraleHint({
        source: `${actorName} 爆牌弃牌`,
        raw: normalized,
        camp: normalizeCamp(actor?.camp),
        loss,
        actorName,
      })
      return
    }

    const cupLoss = line.match(/^\[Action\]\s*(.+?)\s*合成星杯！.*?(红方|蓝方)士气-(\d+)/)
    if (cupLoss) {
      const actorName = cupLoss[1]?.trim()
      const targetCamp = cupLoss[2] === '红方' ? 'Red' : 'Blue'
      const loss = Number(cupLoss[3] || 0)
      battleReviewStore.pushMoraleHint({
        source: `${actorName} 合成星杯`,
        raw: normalized,
        camp: targetCamp,
        loss,
        actorName,
      })
      return
    }

    const campDelta = normalized.match(/(红方|蓝方)士气\s*([+-]\d+)/)
    if (campDelta) {
      const camp = campDelta[1] === '红方' ? 'Red' : 'Blue'
      const delta = Number(campDelta[2] || 0)
      battleReviewStore.pushMoraleHint({
        source: '士气变化',
        raw: normalized,
        camp,
        loss: delta < 0 ? Math.abs(delta) : undefined,
      })
      return
    }

    const genericDelta = normalized.match(/士气\s*([+-]\d+)/)
    if (genericDelta) {
      const delta = Number(genericDelta[1] || 0)
      battleReviewStore.pushMoraleHint({
        source: '士气变化',
        raw: normalized,
        loss: delta < 0 ? Math.abs(delta) : undefined,
      })
    }
  }

  const deriveEndMessageFromState = (state?: GameStateUpdate) => {
    if (!state) return ''
    if (state.red_cups >= 5) return '红方胜利！星杯达到 5'
    if (state.blue_cups >= 5) return '蓝方胜利！星杯达到 5'
    if (state.red_morale <= 0) return '蓝方胜利！红方士气归零'
    if (state.blue_morale <= 0) return '红方胜利！蓝方士气归零'
    return ''
  }

  const isResponsePrompt = (prompt?: Prompt | null) => {
    if (!prompt) return false
    if (prompt.attacker_id || prompt.attack_element || (prompt.counter_target_ids?.length ?? 0) > 0) return true
    const responseOptionIds = new Set(['take', 'hit', 'defend', 'counter', 'shield'])
    return (prompt.options || []).some(option => responseOptionIds.has(String(option?.id || '').trim()))
  }

  const focusPromptActor = (prompt?: Prompt | null, fallbackPlayerId?: string) => {
    const playerId = prompt?.player_id || fallbackPlayerId || ''
    if (!playerId) return
    if (isResponsePrompt(prompt)) {
      battleFxStore.startActingPlayerFocus(playerId, 'response')
      return
    }
    if (prompt?.presentation?.kind === 'action_hub') {
      battleFxStore.startActingPlayerFocus(playerId, 'turn')
      return
    }
    battleFxStore.startActingPlayerFocus(playerId, 'skill')
  }

  function handleSyncState(payload: SyncStatePayload) {
    const state = buildGameStateUpdateFromSyncState(payload)

    if (!sessionStore.gameStarted && payload.room_state === 'Playing') {
      matchLifecycleStore.setGameStarted()
    }

    handleGameplayEvent({
      event_type: 'state_update',
      state,
    })
  }

  function handleRequireAction(payload: RequireActionPayload) {
    const prompt = payload.prompt
    if (prompt) {
      focusPromptActor(prompt, payload.target_user_id)
      if (payload.target_user_id === sessionStore.myPlayerId) {
        interruptStore.setPrompt(prompt)
        interruptStore.setWaiting('')
      } else {
        interruptStore.setPrompt(null)
        interruptStore.setWaiting(payload.target_user_id || '')
      }
      return
    }

    if (payload.target_user_id === sessionStore.myPlayerId) {
      battleFxStore.startActingPlayerFocus(payload.target_user_id, 'skill')
      interruptStore.setWaiting('')
    } else {
      interruptStore.setPrompt(null)
      interruptStore.setWaiting(payload.target_user_id || '')
      battleFxStore.startActingPlayerFocus(payload.target_user_id || '', 'skill')
    }
  }

  function handleNotifyTimeline(payload: TimelineNotifyPayload) {
    timelineStore.push(payload)
    for (const event of extractGameplayEventsFromTimeline(payload.events || [])) {
      handleGameplayEvent(event)
    }
  }

  function handleGameplayEvent(event: GameEvent) {
    console.log('Game event:', event)

    switch (event.event_type) {
      case 'log':
        if (event.message) {
          battleReviewStore.addLog(event.message)
          parseMoraleHintFromLog(event.message)
          const skillAnnouncement = parseSkillAnnouncementFromLog(event.message)
          if (skillAnnouncement?.skillName) {
            const skillActorId = resolvePlayerIdByToken(skillAnnouncement.actorToken)
            if (skillActorId) {
              const actorName = playerNameById(skillActorId) || skillAnnouncement.actorToken
              battleFxStore.startSkillInitiatorFocus(skillActorId, 'skill')
              battleFxStore.addSkillAnnouncement(
                skillActorId,
                actorName,
                skillAnnouncement.skillName,
                skillAnnouncement.effectText,
              )
              battleReviewStore.addBattleFeed({
                type: 'skill',
                title: `${actorName} 发动「${skillAnnouncement.skillName}」`,
                detail: skillAnnouncement.effectText,
                actorId: skillActorId,
                actorName,
              })
            }
          }
        }
        break

      case 'state_update':
        if (event.state) {
          if (!sessionStore.gameStarted) {
            matchLifecycleStore.setGameStarted()
          }
          const prevCurrent = snapshotStore.currentPlayer
          const prevRedMorale = snapshotStore.redMorale
          const prevBlueMorale = snapshotStore.blueMorale
          battleFxStore.prepareForFlowUpdate(event.state.combat_stage, event.state.subflow)
          snapshotStore.updateGameState(event.state)
          interruptStore.syncAfterStateUpdate()
          const me = sessionStore.myPlayerId ? event.state.players[sessionStore.myPlayerId] : undefined
          if (me?.camp || me?.role) {
            sessionStore.setSeat(me.camp || sessionStore.myCamp, me.role || sessionStore.myCharRole)
          }
          battleFxStore.syncInitiatorFocusWithState(event.state.combat_stage, event.state.subflow)
          if (event.state.red_morale !== prevRedMorale) {
            const delta = event.state.red_morale - prevRedMorale
            const hint = battleReviewStore.consumeMoraleHint('Red', delta < 0 ? Math.abs(delta) : undefined)
            battleReviewStore.recordMoraleChange('Red', prevRedMorale, event.state.red_morale, hint)
          }
          if (event.state.blue_morale !== prevBlueMorale) {
            const delta = event.state.blue_morale - prevBlueMorale
            const hint = battleReviewStore.consumeMoraleHint('Blue', delta < 0 ? Math.abs(delta) : undefined)
            battleReviewStore.recordMoraleChange('Blue', prevBlueMorale, event.state.blue_morale, hint)
          }
          if (uiStore.isGameEnded) {
            matchLifecycleStore.refreshGameEndSnapshot(uiStore.gameEndMessage || '游戏结束')
          }
          const fallbackEndMsg = deriveEndMessageFromState(event.state)
          if (fallbackEndMsg && !uiStore.isGameEnded) {
            matchLifecycleStore.setGameEnded(fallbackEndMsg)
            battleReviewStore.addLog(`游戏结束: ${fallbackEndMsg}`)
          }
          const nextCurrent = event.state.current_player
          if (nextCurrent && nextCurrent !== prevCurrent) {
            battleFxStore.clearBattlefieldReveals()
          }
          if (nextCurrent && event.state.turn_stage === 'ActionExecution') {
            battleFxStore.startActingPlayerFocus(nextCurrent, 'turn')
          }
          if (nextCurrent && nextCurrent !== prevCurrent) {
            if (prevCurrent) {
              const prevName = playerNameById(prevCurrent)
              battleReviewStore.addBattleFeed({
                type: 'turn',
                title: `回合结束：${prevName}`,
                actorId: prevCurrent,
                actorName: prevName,
              })
            }
            const currentName = playerNameById(nextCurrent)
            battleReviewStore.addBattleFeed({
              type: 'turn',
              title: `回合开始：${currentName}`,
              actorId: nextCurrent,
              actorName: currentName,
            })
          }
        }
        break

      case 'prompt':
        if (event.prompt) {
          interruptStore.setPrompt(event.prompt)
          focusPromptActor(event.prompt)
          interruptStore.setWaiting('')
        }
        break

      case 'waiting':
        interruptStore.setPrompt(null)
        interruptStore.setWaiting(event.player_id || '')
        battleFxStore.startActingPlayerFocus(event.player_id || '', 'skill')
        break

      case 'error':
        {
          const msg = event.message || '游戏错误'
          interruptStore.showError(msg)
          if (msg.includes('技能发动失败')) {
            interruptStore.clearSkillMode()
            interruptStore.clearActionMode()
          }
          if (msg.includes('无效的卡牌索引')) {
            interruptStore.clearActionMode()
          }
        }
        break

      case 'game_end':
        {
          const message = event.message || '游戏结束'
          const isFirstEndSignal = !uiStore.isGameEnded
          matchLifecycleStore.setGameEnded(message)
          if (isFirstEndSignal) {
            battleReviewStore.addBattleFeed({
              type: 'system',
              title: message,
            })
          }
          battleReviewStore.addLog(`游戏结束: ${message}`)
        }
        break

      case 'chat':
        battleReviewStore.addLog(`[${event.player_name}] ${event.message}`)
        break

      case 'card_revealed':
        if (event.cards?.length && event.player_id) {
          battleFxStore.addFlyingCards(
            event.cards,
            event.player_id,
            event.player_name || event.player_id,
            event.action_type || 'discard',
            event.hidden,
          )
          if (event.action_type === 'magic') {
            battleFxStore.startSkillInitiatorFocus(event.player_id, 'magic')
          }
        }
        break

      case 'damage_dealt':
        if (event.target_id && event.damage) {
          battleFxStore.addDamageEffect(
            event.target_id,
            event.target_name || event.target_id,
            event.damage,
            event.damage_type || 'Attack',
          )
          battleFxStore.settleSkillInitiatorFocus(event.source_id)
        }
        break

      case 'action_step':
        if (event.line) {
          if (event.kind === 'summary') {
            battleReviewStore.addActionStep(event.line)
            battleReviewStore.addBattleFeed({
              type: 'system',
              title: event.line,
            })
          }
        }
        break

      case 'combat_cue':
        if (
          event.attacker_id &&
          event.target_id &&
          (event.phase === 'attack' ||
            event.phase === 'defend' ||
            event.phase === 'take' ||
            event.phase === 'counter' ||
            event.phase === 'shield')
        ) {
          battleFxStore.addCombatCue(event.attacker_id, event.target_id, event.phase)
        }
        break

      case 'draw_cards':
        if (event.player_id && event.draw_count && event.draw_count > 0) {
          const name = event.player_name || playerNameById(event.player_id) || event.player_id
          battleFxStore.addDrawBurst(event.player_id, name, event.draw_count)
        }
        break
    }
  }

  return {
    handleSyncState,
    handleRequireAction,
    handleNotifyTimeline,
    handleGameplayEvent,
  }
}
