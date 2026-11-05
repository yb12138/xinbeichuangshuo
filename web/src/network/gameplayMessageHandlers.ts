import { watch } from 'vue'
import type { GameEvent, GameStateUpdate, Prompt } from '../types/game'
import { extractGameplayEventsFromTimeline } from './gameplayTimeline'
import type { RequireActionPayload, SyncStatePayload, TimelineDelta, TimelineNotifyPayload } from './protocol'
import { buildGameStateUpdateFromSyncState } from './syncState'
import { useBattleFxStore } from '../stores/battlefx.store'
import { useBattleReviewStore, type MoraleCamp } from '../stores/battleReview.store'
import { useInterruptStore } from '../stores/interrupt.store'
import { useMatchLifecycleStore } from '../stores/matchLifecycle.store'
import { useSessionStore } from '../stores/session.store'
import { useSnapshotStore } from '../stores/snapshot.store'
import { useTimelineStore } from '../stores/timeline.store'
import { useStageQueueStore } from '../stores/stageQueue.store'
import { createStageQueueBridge } from './stageQueueBridge'
import { cardArtSrc } from '../constants/cardArt'
import { useUiStore } from '../stores/ui.store'
import { playSfx } from '../audio/sfx'
import { endCueFor } from '../audio/endCue'

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

  const playerNameById = (id?: string) => {
    if (!id) return ''
    return snapshotStore.players[id]?.name || id
  }

  const stageQueue = useStageQueueStore()
  const stageBridge = createStageQueueBridge({
    stageQueue,
    playerName: (id: string) => playerNameById(id),
    roleName: (id: string) => {
      const role = snapshotStore.players[id]?.role || ''
      return snapshotStore.characters[role]?.name || ''
    },
    cardImage: (card) => (card?.name ? cardArtSrc(card) : undefined),
    playerCamp: (id: string) => snapshotStore.players[id]?.camp || '',
  })

  const normalizeCamp = (camp?: string): MoraleCamp | undefined => {
    if (camp === 'Red' || camp === 'Blue') return camp
    return undefined
  }

  const campLabel = (camp?: string) => {
    if (camp === 'Red') return '红方'
    if (camp === 'Blue') return '蓝方'
    return camp || ''
  }

  /**
   * 回合徽章内容：主标题"狂战士 的回合"（轮到我时为"你的回合"），
   * 副标题"红阵营 · 第 12 回合"；角色未知时退回玩家名。
   */
  const turnBannerContent = (playerId: string) => {
    const player = snapshotStore.players[playerId]
    if (!player) return null
    const roleId = String(player.role || '').trim()
    const roleName = (roleId && snapshotStore.characters[roleId]?.name) || roleId || player.name || playerId
    const isMine = !!sessionStore.myPlayerId && playerId === sessionStore.myPlayerId
    const campName = player.camp === 'Red' ? '红阵营' : player.camp === 'Blue' ? '蓝阵营' : ''
    const turn = snapshotStore.turnNumber
    const subtitle = [campName, turn > 0 ? `第 ${turn} 回合` : ''].filter(Boolean).join(' · ')
    return { text: isMine ? '你的回合' : `${roleName} 的回合`, subtitle, roleId, isMine }
  }

  const signed = (value?: number) => {
    const numeric = Number(value || 0)
    if (numeric > 0) return `+${numeric}`
    return String(numeric)
  }

  const playerLabel = (playerId?: string) => {
    if (!playerId) return ''
    return playerNameById(playerId) || playerId
  }

  const fieldCardLabel = (delta: TimelineDelta) => {
    return delta.field_card?.card?.name || delta.after_text || delta.before_text || '场上牌'
  }

  const stateDeltaTitle = (delta: TimelineDelta) => {
    const label = playerLabel(delta.target_user_id)
    switch (delta.type) {
      case 'morale':
        return `${campLabel(delta.camp)}士气 ${signed(delta.value)}`
      case 'team_gem':
        return `${campLabel(delta.camp)}阵营宝石 ${signed(delta.value)}`
      case 'team_crystal':
        return `${campLabel(delta.camp)}阵营水晶 ${signed(delta.value)}`
      case 'team_cup':
        return `${campLabel(delta.camp)}星杯 ${signed(delta.value)}`
      case 'player_gem':
        return `${label} 宝石 ${signed(delta.value)}`
      case 'player_crystal':
        return `${label} 水晶 ${signed(delta.value)}`
      case 'heal':
        return `${label} 治疗 ${signed(delta.value)}`
      case 'hand_count':
        return `${label} 手牌数 ${signed(delta.value)}`
      case 'discard_count':
        return `弃牌堆 ${signed(delta.value)}`
      case 'deck_count':
        return `牌库 ${signed(delta.value)}`
      case 'field_card_added':
        return `${label} 放置 ${fieldCardLabel(delta)}`
      case 'field_card_removed':
        return `${label} 移除 ${fieldCardLabel(delta)}`
      case 'field_card_changed':
        return `${label} 场上牌变化`
      case 'form':
        return `${label} 形态：${delta.before_text || '默认'} -> ${delta.after_text || '默认'}`
      case 'orientation':
        return `${label} 朝向：${delta.before_text || 'Normal'} -> ${delta.after_text || 'Normal'}`
      case 'token':
      case 'status':
        return `${label} ${delta.field || delta.type} ${signed(delta.value)}`
      default:
        return `${label || campLabel(delta.camp) || '状态'} ${delta.type} ${signed(delta.value)}`
    }
  }

  const battleFeedTypeForDelta = (delta: TimelineDelta) => {
    if (delta.type === 'morale') return 'damage'
    if (delta.type.includes('gem') || delta.type.includes('crystal') || delta.type.includes('cup') || delta.type.endsWith('_count')) {
      return 'resource'
    }
    return 'system'
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

  /**
   * 战斗链还在播放时先压住我的操作提示，播完再交出操作权，避免上一手还没看清就被催着决策。
   * 新的 state_update 会作废压着的提示（与 syncAfterStateUpdate 语义一致）；
   * 最多等 PROMPT_GATE_MAX_MS，防止后台标签页 rAF 暂停时永远拿不到操作权。
   */
  const PROMPT_GATE_MAX_MS = 8000
  let gatedPrompt: Prompt | null = null
  let gateTimer: ReturnType<typeof setTimeout> | null = null
  let stopGateWatch: (() => void) | null = null

  function clearPromptGate() {
    gatedPrompt = null
    if (gateTimer) clearTimeout(gateTimer)
    gateTimer = null
    stopGateWatch?.()
    stopGateWatch = null
  }

  function releaseGatedPrompt() {
    const prompt = gatedPrompt
    clearPromptGate()
    if (!prompt) return
    showMyPrompt(prompt)
  }

  /** 把操作权交给我；从"没有提示"变成"有提示"时响一声，重复同步同一提示不重复提醒。 */
  function showMyPrompt(prompt: Prompt) {
    const wasIdle = !interruptStore.currentPrompt
    interruptStore.setPrompt(prompt)
    interruptStore.setWaiting('')
    if (wasIdle) playSfx('prompt')
  }

  function presentMyPrompt(prompt: Prompt) {
    clearPromptGate()
    if (!uiStore.stageOverlayEnabled || !stageQueue.isPlaying) {
      showMyPrompt(prompt)
      return
    }
    gatedPrompt = prompt
    interruptStore.setPrompt(null)
    interruptStore.setWaiting('')
    stopGateWatch = watch(
      () => stageQueue.isPlaying,
      (playing) => {
        if (!playing) releaseGatedPrompt()
      },
      { flush: 'sync' },
    )
    gateTimer = setTimeout(releaseGatedPrompt, PROMPT_GATE_MAX_MS)
  }

  /**
   * 游戏结束有两条入口：game_end 事件，或 state_update 里星杯 / 士气已到终局（可能先到）。
   * 哪条先把游戏标记为结束，就由哪条放结尾音；另一条看到已结束就不会重复。
   */
  function playEndCueFor(message: string) {
    const myCamp = snapshotStore.players[sessionStore.myPlayerId]?.camp || sessionStore.myCamp
    playEndCue(endCueFor(message, myCamp))
  }

  /** 结尾音乐等最后一手演出播完再响，免得盖住决胜的那一下；最多等 END_CUE_MAX_WAIT_MS。 */
  const END_CUE_MAX_WAIT_MS = 4000
  function playEndCue(cue: Parameters<typeof playSfx>[0]) {
    if (!stageQueue.isPlaying) {
      playSfx(cue)
      return
    }
    let done = false
    const fire = () => {
      if (done) return
      done = true
      stop()
      clearTimeout(timer)
      playSfx(cue)
    }
    const stop = watch(
      () => stageQueue.isPlaying,
      (playing) => {
        if (!playing) fire()
      },
    )
    const timer = setTimeout(fire, END_CUE_MAX_WAIT_MS)
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
        presentMyPrompt(prompt)
      } else {
        clearPromptGate()
        interruptStore.setPrompt(null)
        interruptStore.setWaiting(payload.target_user_id || '')
      }
      return
    }

    if (payload.target_user_id === sessionStore.myPlayerId) {
      battleFxStore.startActingPlayerFocus(payload.target_user_id, 'skill')
      interruptStore.setWaiting('')
    } else {
      clearPromptGate()
      interruptStore.setPrompt(null)
      interruptStore.setWaiting(payload.target_user_id || '')
      battleFxStore.startActingPlayerFocus(payload.target_user_id || '', 'skill')
    }
  }

  function handleNotifyTimeline(payload: TimelineNotifyPayload) {
    timelineStore.push(payload)
    if (uiStore.stageOverlayEnabled) {
      stageBridge.onTimeline(payload)
    }
    const hasStructuredNarrative = battleFxStore.hasStructuredNarrative(payload.events || [])
    if (hasStructuredNarrative) {
      battleFxStore.applyStructuredTimelineNarrative(payload)
    }
    if (payload.is_replay) {
      return
    }
    for (const event of extractGameplayEventsFromTimeline(payload.events || [])) {
      handleGameplayEvent(event, { suppressNarrative: hasStructuredNarrative })
    }
  }

  function handleGameplayEvent(event: GameEvent, options: { suppressNarrative?: boolean } = {}) {
    console.log('Game event:', event)

    switch (event.event_type) {
      case 'log':
        if (event.message) {
          battleReviewStore.addLog(event.message)
        }
        break

      case 'state_update':
        if (event.state) {
          if (!sessionStore.gameStarted) {
            matchLifecycleStore.setGameStarted()
          }
          const prevCurrent = snapshotStore.currentPlayer
          battleFxStore.prepareForFlowUpdate(event.state.combat_stage, event.state.subflow)
          snapshotStore.updateGameState(event.state)
          clearPromptGate()
          interruptStore.syncAfterStateUpdate()
          const me = sessionStore.myPlayerId ? event.state.players[sessionStore.myPlayerId] : undefined
          if (me?.camp || me?.role) {
            sessionStore.setSeat(me.camp || sessionStore.myCamp, me.role || sessionStore.myCharRole)
          }
          battleFxStore.syncInitiatorFocusWithState(event.state.combat_stage, event.state.subflow)
          if (uiStore.isGameEnded) {
            matchLifecycleStore.refreshGameEndSnapshot(uiStore.gameEndMessage || '游戏结束')
          }
          const fallbackEndMsg = deriveEndMessageFromState(event.state)
          if (fallbackEndMsg && !uiStore.isGameEnded) {
            matchLifecycleStore.setGameEnded(fallbackEndMsg)
            playEndCueFor(fallbackEndMsg)
            battleReviewStore.addLog(`游戏结束: ${fallbackEndMsg}`)
          }
          const nextCurrent = event.state.current_player
          const nextActionPlayerStarted =
            !!nextCurrent &&
            nextCurrent !== prevCurrent &&
            event.state.turn_stage === 'ActionExecution'
          if (nextActionPlayerStarted) {
            battleFxStore.clearActionNarrative()
          }
          if (nextCurrent && event.state.turn_stage === 'ActionExecution') {
            battleFxStore.beginActionNarrative(nextCurrent)
            battleFxStore.startActingPlayerFocus(nextCurrent, 'turn')
          }
          if (nextCurrent && nextCurrent !== prevCurrent) {
            const banner = turnBannerContent(nextCurrent)
            if (banner) {
              battleFxStore.showTurnBanner(banner.text, normalizeCamp(snapshotStore.players[nextCurrent]?.camp) || '', {
                subtitle: banner.subtitle,
                roleId: banner.roleId,
                isMine: banner.isMine,
              })
            }
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
          focusPromptActor(event.prompt)
          presentMyPrompt(event.prompt)
        }
        break

      case 'waiting':
        clearPromptGate()
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
            playEndCueFor(message)
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
          if (event.action_type === 'magic') {
            battleFxStore.startSkillInitiatorFocus(event.player_id, 'magic')
          }
          if (!event.hidden && !options.suppressNarrative) {
            for (const card of event.cards) {
              battleFxStore.addNarrativeCard(event.player_id, card, event.action_type || 'discard')
            }
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
          if (!options.suppressNarrative) {
            battleFxStore.addNarrativeDamage(
              event.source_id,
              event.target_id,
              event.damage,
              event.damage_type || 'Attack',
            )
          }
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
          if (!options.suppressNarrative) {
            battleFxStore.addCombatCue(event.attacker_id, event.target_id, event.phase)
          }
        }
        break

      case 'draw_cards':
        if (event.player_id && event.draw_count && event.draw_count > 0) {
          const name = event.player_name || playerNameById(event.player_id) || event.player_id
          battleFxStore.addDrawBurst(event.player_id, name, event.draw_count)
        }
        break

      case 'skill_activated':
        if (event.player_id && event.skill_name) {
          const actorName = event.player_name || playerNameById(event.player_id) || event.player_id
          battleFxStore.startSkillInitiatorFocus(event.player_id, 'skill')
          battleFxStore.addSkillAnnouncement(
            event.player_id,
            actorName,
            event.skill_name,
            event.effect_text || '技能效果生效',
          )
          if (!options.suppressNarrative) {
            battleFxStore.addNarrativeSkill(
              event.player_id,
              event.skill_name,
              event.effect_text || '技能效果生效',
              event.target_ids || [],
            )
          }
          battleReviewStore.addBattleFeed({
            type: 'skill',
            title: `${actorName} 发动「${event.skill_name}」`,
            detail: event.effect_text || '技能效果生效',
            actorId: event.player_id,
            actorName,
          })
        }
        break

      case 'special_action':
        if (event.player_id) {
          const actorName = event.player_name || playerNameById(event.player_id) || event.player_id
          battleReviewStore.addBattleFeed({
            type: 'resource',
            title: event.summary || `${actorName} 执行特殊行动`,
            actorId: event.player_id,
            actorName,
          })
        }
        break

      case 'state_delta':
        for (const delta of event.deltas || []) {
          if (delta.type === 'morale') {
            const camp = normalizeCamp(delta.camp)
            if (camp && delta.before !== undefined && delta.after !== undefined) {
              battleReviewStore.recordMoraleChange(camp, delta.before, delta.after, {
                id: 0,
                timestamp: Date.now(),
                source: delta.reason || '状态变化',
                raw: stateDeltaTitle(delta),
                camp,
                loss: (delta.value || 0) < 0 ? Math.abs(delta.value || 0) : undefined,
              })
            }
          }
          battleReviewStore.addBattleFeed({
            type: battleFeedTypeForDelta(delta),
            title: stateDeltaTitle(delta),
            targetId: delta.target_user_id,
            targetName: playerLabel(delta.target_user_id),
          })
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
