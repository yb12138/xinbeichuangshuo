import { ref } from 'vue'
import { useGameStore } from '../stores/gameStore'
import type { WSMessage, RoomEvent, GameEvent, PlayerAction } from '../types/game'

// 改造成一个函数，动态获取当前访问的 IP 和端口
const getWsUrl = () => {
  // 如果有配环境变量，优先用环境变量（方便以后线上部署）
  if (import.meta.env.VITE_WS_URL) {
    return import.meta.env.VITE_WS_URL
  }

  // 开发模式优先直连后端 8080，避免经过 Vite 的 /ws 代理产生 EPIPE 噪声。
  if (import.meta.env.DEV) {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const host = window.location.hostname
    return `${protocol}//${host}:8080/ws`
  }

  // 动态拼接地址
  // window.location.host 会自动变成比如 '192.168.1.100:5173' 或者 'localhost:5173'
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const host = window.location.host
  return `${protocol}//${host}/ws`
}

let ws: WebSocket | null = null
const reconnectAttempts = ref(0)
const maxReconnectAttempts = 5

type ReconnectInfo = {
  room_code: string
  player_id: string
  player_name: string
  token: string
}

const reconnectStorageKey = (roomCode: string, playerName: string) =>
  `xbs_reconnect_${roomCode}_${encodeURIComponent(playerName)}`

function loadReconnectInfo(roomCode: string, playerName: string): ReconnectInfo | null {
  if (!roomCode || !playerName || typeof window === 'undefined') return null
  try {
    const raw = localStorage.getItem(reconnectStorageKey(roomCode, playerName))
    if (!raw) return null
    const parsed = JSON.parse(raw) as ReconnectInfo
    if (parsed.room_code !== roomCode || parsed.player_name !== playerName) return null
    if (!parsed.player_id || !parsed.token) return null
    return parsed
  } catch {
    return null
  }
}

function saveReconnectInfo(roomCode: string, playerName: string, playerId: string, token: string) {
  if (!roomCode || !playerName || !playerId || !token || typeof window === 'undefined') return
  const payload: ReconnectInfo = {
    room_code: roomCode,
    player_id: playerId,
    player_name: playerName,
    token
  }
  try {
    localStorage.setItem(reconnectStorageKey(roomCode, playerName), JSON.stringify(payload))
  } catch {
    // ignore storage errors
  }
}

export function useWebSocket() {
  const store = useGameStore()

  function safeStringify(data: unknown) {
    try {
      return JSON.stringify(data)
    } catch {
      return '[unserializable]'
    }
  }

  function connect(roomCode: string, playerName: string, createRoom: boolean = false) {
    store.myName = playerName

    const baseUrl = getWsUrl()
    const sessionReconnect = !createRoom &&
      store.reconnectToken &&
      store.myPlayerId &&
      store.roomCode === roomCode &&
      store.myName === playerName
        ? { player_id: store.myPlayerId, token: store.reconnectToken }
        : null
    const reconnectInfo = sessionReconnect || (!createRoom ? loadReconnectInfo(roomCode, playerName) : null)
    const reconnectParams = reconnectInfo
      ? `&player_id=${encodeURIComponent(reconnectInfo.player_id)}&reconnect_token=${encodeURIComponent(reconnectInfo.token)}`
      : ''

    const url = createRoom
        ? `${baseUrl}?name=${encodeURIComponent(playerName)}&create=true`
        : `${baseUrl}?room=${roomCode}&name=${encodeURIComponent(playerName)}${reconnectParams}`
    
    console.log('Connecting to:', url)
    store.addLog(`[WS] 连接中: ${url}`)
    
    ws = new WebSocket(url)

    ws.onopen = () => {
      console.log('WebSocket connected')
      store.setConnected(true)
      reconnectAttempts.value = 0
      store.addLog('[WS] 连接成功')
    }

    ws.onmessage = (event) => {
      try {
        store.addLog(`[WS][RX] raw: ${String(event.data)}`)
        const msg: WSMessage = JSON.parse(event.data)
        store.addLog(`[WS][RX] ${msg.type}: ${safeStringify(msg.payload)}`)
        handleMessage(msg)
      } catch (e) {
        console.error('Failed to parse message:', e)
      }
    }

    ws.onerror = (error) => {
      console.error('WebSocket error:', error)
      store.setError('连接错误')
      store.addLog('[WS] 连接错误')
    }

    ws.onclose = () => {
      console.log('WebSocket closed')
      store.setConnected(false)
      store.addLog('[WS] 连接关闭')
      
      // 尝试重连
      if (reconnectAttempts.value < maxReconnectAttempts && store.isInRoom) {
        reconnectAttempts.value++
        console.log(`Reconnecting... attempt ${reconnectAttempts.value}`)
        setTimeout(() => {
          if (store.roomCode) {
            connect(store.roomCode, store.myName)
          }
        }, 2000 * reconnectAttempts.value)
      }
    }
  }

  function disconnect() {
    if (ws) {
      ws.close()
      ws = null
    }
    store.reset()
  }

  function handleMessage(msg: WSMessage) {
    switch (msg.type) {
      case 'room':
        handleRoomEvent(msg.payload as RoomEvent)
        break
      case 'event':
        handleGameEvent(msg.payload as GameEvent)
        break
      default:
        console.log('Unknown message type:', msg.type)
    }
  }

  function handleRoomEvent(event: RoomEvent) {
    console.log('Room event:', event)
    
    switch (event.action) {
      case 'assigned':
        store.setRoomInfo(
          event.room_code,
          event.player_id!,
          event.camp || '',
          event.char_role || ''
        )
        if (event.reconnect_token && event.player_id) {
          store.setReconnectToken(event.reconnect_token)
          saveReconnectInfo(event.room_code, store.myName, event.player_id, event.reconnect_token)
        }
        if (event.characters?.length) {
          store.setCharacters(event.characters)
        }
        store.addLog(`已加入房间 ${event.room_code}，你是 ${event.player_id}`)
        break
        
      case 'player_list':
        store.updateRoomPlayers(event.players || [])
        if (event.characters?.length) {
          store.setCharacters(event.characters)
        }
        break
        
      case 'started':
        store.setGameStarted()
        if (event.characters?.length) {
          store.setCharacters(event.characters)
        }
        store.addLog('游戏开始！')
        break
        
      case 'joined':
        store.addLog(event.message || `${event.player_name} 加入了房间`)
        break
        
      case 'left':
        store.addLog(event.message || `${event.player_name} 离开了房间`)
        break
        
      case 'error':
        store.setError(event.message || '房间错误')
        break

      case 'dissolved':
        {
          const msg = event.message || '房间已解散'
          // 先断开并重置，再展示提示，避免重置把提示清掉。
          if (ws) {
            ws.close()
            ws = null
          }
          store.reset()
          store.setError(msg)
        }
        break
    }
  }

  function handleGameEvent(event: GameEvent) {
    console.log('Game event:', event)

    const playerByName = (name?: string) => {
      if (!name) return undefined
      return Object.values(store.players).find(p => p.name === name)
    }
    const playerNameById = (id?: string) => {
      if (!id) return ''
      return store.players[id]?.name || id
    }
    const normalizeCamp = (camp?: string): 'Red' | 'Blue' | undefined => {
      if (camp === 'Red' || camp === 'Blue') return camp
      return undefined
    }
    const parseMoraleHintFromLog = (line: string) => {
      if (!line) return
      const normalized = line.replace(/^\[[^\]]+\]\s*/, '').trim()

      // 1) 爆牌弃牌（典型）
      const discardLoss = line.match(/^\[System\]\s*(.+?)\s*丢弃了\s*(\d+)\s*张牌！士气\s*-(\d+)/)
      if (discardLoss) {
        const actorName = discardLoss[1]?.trim()
        const loss = Number(discardLoss[3] || 0)
        const actor = playerByName(actorName)
        store.pushMoraleHint({
          source: `${actorName} 爆牌弃牌`,
          raw: normalized,
          camp: normalizeCamp(actor?.camp),
          loss,
          actorName
        })
        return
      }

      // 2) 合成星杯导致对方士气下降
      const cupLoss = line.match(/^\[Action\]\s*(.+?)\s*合成星杯！.*?(红方|蓝方)士气-(\d+)/)
      if (cupLoss) {
        const actorName = cupLoss[1]?.trim()
        const targetCamp = cupLoss[2] === '红方' ? 'Red' : 'Blue'
        const loss = Number(cupLoss[3] || 0)
        store.pushMoraleHint({
          source: `${actorName} 合成星杯`,
          raw: normalized,
          camp: targetCamp,
          loss,
          actorName
        })
        return
      }

      // 3) 泛化匹配：红/蓝方士气±X
      const campDelta = normalized.match(/(红方|蓝方)士气\s*([+-]\d+)/)
      if (campDelta) {
        const camp = campDelta[1] === '红方' ? 'Red' : 'Blue'
        const delta = Number(campDelta[2] || 0)
        store.pushMoraleHint({
          source: '士气变化',
          raw: normalized,
          camp,
          loss: delta < 0 ? Math.abs(delta) : undefined
        })
        return
      }

      // 4) 兜底：只要日志出现“士气±X”，记录提示，等 state_update 对齐到具体阵营
      const genericDelta = normalized.match(/士气\s*([+-]\d+)/)
      if (genericDelta) {
        const delta = Number(genericDelta[1] || 0)
        store.pushMoraleHint({
          source: '士气变化',
          raw: normalized,
          loss: delta < 0 ? Math.abs(delta) : undefined
        })
      }
    }

    const deriveEndMessageFromState = (state?: GameEvent['state']) => {
      if (!state) return ''
      if (state.red_cups >= 5) return '红方胜利！星杯达到 5'
      if (state.blue_cups >= 5) return '蓝方胜利！星杯达到 5'
      if (state.red_morale <= 0) return '蓝方胜利！红方士气归零'
      if (state.blue_morale <= 0) return '红方胜利！蓝方士气归零'
      return ''
    }
    
    switch (event.event_type) {
      case 'log':
        if (event.message) {
          store.addLog(event.message)
          parseMoraleHintFromLog(event.message)
          // 技能发动日志：弹出明显提示
          const m = event.message.match(/发动\s*\[([^\]]+)\]/)
          if (m) {
            store.setSkillEffectToast(`${m[1]} 技能生效！`)
          }
        }
        break
        
      case 'state_update':
        if (event.state) {
          const prevCurrent = store.currentPlayer
          const prevRedMorale = store.redMorale
          const prevBlueMorale = store.blueMorale
          store.updateGameState(event.state)
          if (event.state.red_morale !== prevRedMorale) {
            const delta = event.state.red_morale - prevRedMorale
            const hint = store.consumeMoraleHint('Red', delta < 0 ? Math.abs(delta) : undefined)
            store.recordMoraleChange('Red', prevRedMorale, event.state.red_morale, hint)
          }
          if (event.state.blue_morale !== prevBlueMorale) {
            const delta = event.state.blue_morale - prevBlueMorale
            const hint = store.consumeMoraleHint('Blue', delta < 0 ? Math.abs(delta) : undefined)
            store.recordMoraleChange('Blue', prevBlueMorale, event.state.blue_morale, hint)
          }
          // 兜底：若后端漏发 game_end，客户端根据终局状态也能收口到“结束界面”
          const fallbackEndMsg = deriveEndMessageFromState(event.state)
          if (fallbackEndMsg && !store.isGameEnded) {
            store.setGameEnded(fallbackEndMsg)
            store.addLog(`游戏结束: ${fallbackEndMsg}`)
          }
          const nextCurrent = event.state.current_player
          if (nextCurrent && nextCurrent !== prevCurrent) {
            if (prevCurrent) {
              const prevName = playerNameById(prevCurrent)
              store.addBattleFeed({
                type: 'turn',
                title: `回合结束：${prevName}`,
                actorId: prevCurrent,
                actorName: prevName
              })
            }
            const currentName = playerNameById(nextCurrent)
            store.addBattleFeed({
              type: 'turn',
              title: `回合开始：${currentName}`,
              actorId: nextCurrent,
              actorName: currentName
            })
          }
        }
        break
        
      case 'prompt':
        if (event.prompt) {
          store.setPrompt(event.prompt)
          store.setWaiting('')
        }
        break
        
      case 'waiting':
        store.setPrompt(null)
        store.setWaiting(event.player_id || '')
        break
        
      case 'error':
        {
          const msg = event.message || '游戏错误'
          store.setError(msg)
          // 主动技执行失败后，回到行动主面板，允许重新选择攻击/法术/特殊行动。
          if (msg.includes('技能发动失败')) {
            store.clearSkillMode()
            store.clearActionMode()
          }
          // 卡牌索引过期时，清理前端选牌状态，避免继续发送同一无效索引。
          if (msg.includes('无效的卡牌索引')) {
            store.clearActionMode()
          }
        }
        break
        
      case 'game_end':
        {
          const message = event.message || '游戏结束'
          const isFirstEndSignal = !store.isGameEnded
          store.setGameEnded(message)
          if (isFirstEndSignal) {
            store.addBattleFeed({
              type: 'system',
              title: message
            })
          }
          store.addLog(`游戏结束: ${message}`)
        }
        break
        
      case 'chat':
        store.addLog(`[${event.player_name}] ${event.message}`)
        break

      case 'card_revealed':
        // 明牌/暗牌展示：出牌/弃牌动画
        if (event.cards?.length && event.player_id) {
          store.addFlyingCards(
            event.cards,
            event.player_id,
            event.player_name || event.player_id,
            event.action_type || 'discard',
            event.hidden
          )
        }
        break

      case 'damage_dealt':
        // 伤害结算：弹出通知弹框，同时显示暴血特效
        if (event.target_id && event.damage) {
          store.addDamageEffect(
            event.target_id,
            event.target_name || event.target_id,
            event.damage,
            event.damage_type || 'Attack'
          )

        }
        break

      case 'action_step':
        // 行动步骤：桌面展示行动流程
        if (event.line && event.kind === 'summary') {
          store.addActionStep(event.line)
          store.addBattleFeed({
            type: 'system',
            title: event.line
          })
        }
        break

      case 'combat_cue':
        if (event.attacker_id && event.target_id && event.phase) {
          store.addCombatCue(event.attacker_id, event.target_id, event.phase)
        }
        break

      case 'draw_cards':
        if (event.player_id && event.draw_count && event.draw_count > 0) {
          const name = event.player_name || playerNameById(event.player_id) || event.player_id
          store.addDrawBurst(event.player_id, name, event.draw_count)
        }
        break
    }
  }

  function sendAction(action: PlayerAction) {
    if (!ws || ws.readyState !== WebSocket.OPEN) {
      store.setError('未连接到服务器')
      return
    }

    const msg: WSMessage = {
      type: 'action',
      payload: action
    }
    
    store.addLog(`[WS][TX] action: ${safeStringify(action)}`)
    ws.send(JSON.stringify(msg))
  }

  function sendRoomAction(action: string, data?: Record<string, unknown>) {
    if (!ws || ws.readyState !== WebSocket.OPEN) {
      store.setError('未连接到服务器')
      return
    }

    const msg: WSMessage = {
      type: 'room',
      payload: { action, ...data }
    }
    
    store.addLog(`[WS][TX] room: ${safeStringify(msg.payload)}`)
    ws.send(JSON.stringify(msg))
  }

  function sendChat(message: string) {
    if (!ws || ws.readyState !== WebSocket.OPEN) {
      return
    }

    const msg: WSMessage = {
      type: 'chat',
      payload: { message }
    }
    
    store.addLog(`[WS][TX] chat: ${safeStringify(msg.payload)}`)
    ws.send(JSON.stringify(msg))
  }

  // 便捷方法
  function attack(targetId: string, cardIndex: number) {
    sendAction({
      player_id: store.myPlayerId,
      type: 'Attack',
      target_id: targetId,
      card_index: cardIndex
    })
  }

  function magic(targetId: string | undefined, cardIndex: number) {
    const action: PlayerAction = {
      player_id: store.myPlayerId,
      type: 'Magic',
      card_index: cardIndex
    }
    if (targetId) action.target_id = targetId
    sendAction(action)
  }

  function useSkill(skillId: string, targetIds: string[] = [], selections?: number[]) {
    const action: PlayerAction = {
      player_id: store.myPlayerId,
      type: 'Skill',
      skill_id: skillId
    }
    if (targetIds.length > 0) action.target_ids = targetIds
    if (selections && selections.length > 0) action.selections = selections
    sendAction(action)
  }

  function pass() {
    sendAction({
      player_id: store.myPlayerId,
      type: 'Pass'
    })
  }

  function confirm() {
    sendAction({
      player_id: store.myPlayerId,
      type: 'Confirm'
    })
  }

  function cancel() {
    sendAction({
      player_id: store.myPlayerId,
      type: 'Cancel'
    })
  }

  function select(selections: number[]) {
    sendAction({
      player_id: store.myPlayerId,
      type: 'Select',
      selections
    })
  }

  function respond(action: string, cardIndex?: number, targetId?: string) {
    const payload: PlayerAction = {
      player_id: store.myPlayerId,
      type: 'Respond',
      extra_args: [action]
    }
    if (cardIndex !== undefined) payload.card_index = cardIndex
    if (targetId) payload.target_id = targetId
    sendAction(payload)
  }

  function buy() {
    sendAction({
      player_id: store.myPlayerId,
      type: 'Buy'
    })
  }

  function extract() {
    sendAction({
      player_id: store.myPlayerId,
      type: 'Extract'
    })
  }

  function cheatSkill(playerId: string, roleId: string, skillId: string) {
    const pid = playerId || store.myPlayerId
    const args = roleId ? [pid, roleId, skillId] : [pid, skillId]
    sendAction({
      player_id: store.myPlayerId,
      type: 'Cheat',
      target_id: 'skill',
      extra_args: args
    })
  }

  function cheatToken(playerId: string, tokenKey: string, value: number) {
    const pid = playerId || store.myPlayerId
    sendAction({
      player_id: store.myPlayerId,
      type: 'Cheat',
      target_id: 'token',
      extra_args: [pid, tokenKey, String(value)]
    })
  }

  function cheatSet(playerId: string, field: string, value: number) {
    const pid = playerId || store.myPlayerId
    sendAction({
      player_id: store.myPlayerId,
      type: 'Cheat',
      target_id: 'set',
      extra_args: [pid, field, String(value)]
    })
  }

  function cheatEffect(playerId: string, effect: string, count: number) {
    const pid = playerId || store.myPlayerId
    sendAction({
      player_id: store.myPlayerId,
      type: 'Cheat',
      target_id: 'effect',
      extra_args: [pid, effect, String(count)]
    })
  }

  function cheatGiveExclusive(playerId: string, roleId: string, skillId: string, count: number) {
    const pid = playerId || store.myPlayerId
    sendAction({
      player_id: store.myPlayerId,
      type: 'Cheat',
      target_id: 'card_exclusive',
      extra_args: [pid, roleId, skillId, String(count)]
    })
  }

  function cheatGiveByElement(playerId: string, element: string, count: number) {
    const pid = playerId || store.myPlayerId
    sendAction({
      player_id: store.myPlayerId,
      type: 'Cheat',
      target_id: 'card_element',
      extra_args: [pid, element, String(count)]
    })
  }

  function cheatGiveByFaction(playerId: string, faction: string, count: number) {
    const pid = playerId || store.myPlayerId
    sendAction({
      player_id: store.myPlayerId,
      type: 'Cheat',
      target_id: 'card_faction',
      extra_args: [pid, faction, String(count)]
    })
  }

  function cheatGiveMagicByName(playerId: string, cardName: string, count: number) {
    const pid = playerId || store.myPlayerId
    sendAction({
      player_id: store.myPlayerId,
      type: 'Cheat',
      target_id: 'card_magic',
      extra_args: [pid, cardName, String(count)]
    })
  }

  return {
    connect,
    disconnect,
    sendAction,
    sendRoomAction,
    sendChat,
    // 便捷方法
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
    cheatGiveMagicByName
  }
}
