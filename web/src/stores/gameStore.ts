import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { 
  PlayerView, 
  GameStateUpdate, 
  Prompt, 
  PlayerInfo,
  AvailableSkill,
  CharacterView,
  SkillView,
  Card
} from '../types/game'
import { ROLE_NAME_MAP } from '../constants/roleNameMap'

export type BattleFeedType =
  | 'turn'
  | 'skill'
  | 'attack'
  | 'magic'
  | 'respond'
  | 'damage'
  | 'resource'
  | 'system'

export interface BattleFeedEntry {
  id: number
  type: BattleFeedType
  title: string
  detail?: string
  actorId?: string
  actorName?: string
  targetId?: string
  targetName?: string
  timestamp: number
}

export type MoraleCamp = 'Red' | 'Blue'

export interface MoraleHint {
  id: number
  timestamp: number
  source: string
  raw: string
  camp?: MoraleCamp
  loss?: number
  actorName?: string
}

export interface MoraleChangeEntry {
  id: number
  timestamp: number
  camp: MoraleCamp
  delta: number
  before: number
  after: number
  source: string
  raw?: string
}

export interface GameEndSnapshot {
  message: string
  triggerType: 'cups' | 'morale' | 'unknown'
  finalRedMorale: number
  finalBlueMorale: number
  finalRedCups: number
  finalBlueCups: number
  triggerCamp?: MoraleCamp
  triggerDelta?: number
  triggerSource?: string
}

export interface SkillModalAnchor {
  x: number
  y: number
  width: number
  height: number
}

export type InitiatorFocusMode = 'attack' | 'magic' | 'skill'
export type InitiatorFocusSide = 'left' | 'right'

export interface InitiatorFocusState {
  playerId: string
  side: InitiatorFocusSide
  mode: InitiatorFocusMode
  startedAt: number
}

export const useGameStore = defineStore('game', () => {
  // 房间状态
  const roomCode = ref<string>('')
  const myPlayerId = ref<string>('')
  const myName = ref<string>('')
  const myCamp = ref<string>('')
  const myCharRole = ref<string>('')
  const reconnectToken = ref<string>('')
  const roomPlayers = ref<PlayerInfo[]>([])
  const isInRoom = ref(false)
  const gameStarted = ref(false)

  // 游戏状态
  const phase = ref<string>('')
  const currentPlayer = ref<string>('')
  const hasPerformedStartup = ref(false)
  const players = ref<Record<string, PlayerView>>({})
  const redMorale = ref(15)
  const blueMorale = ref(15)
  const redCups = ref(0)
  const blueCups = ref(0)
  const redGems = ref(0)
  const blueGems = ref(0)
  const redCrystals = ref(0)
  const blueCrystals = ref(0)
  const deckCount = ref(0)
  const discardCount = ref(0)

  // 明牌展示动画：{ cards, playerId, playerName, actionType, id, hidden }
  const flyingCards = ref<Array<{
    id: number
    cards: Card[]
    playerId: string
    playerName: string
    actionType: string
    holdMode: 'timed' | 'until_response' | 'until_next_card_or_draw'
    hidden?: boolean
  }>>([])
  let flyingCardsId = 0
  const flyingCardsQueue = ref<Array<{
    cards: Card[]
    playerId: string
    playerName: string
    actionType: string
    holdMode: 'timed' | 'until_response' | 'until_next_card_or_draw'
    hidden?: boolean
  }>>([])
  let flyingCardsTimer: ReturnType<typeof setTimeout> | null = null

  // 摸牌演出：从顶部公共牌堆飞向角色区域
  const drawBursts = ref<Array<{
    id: number
    playerId: string
    playerName: string
    count: number
  }>>([])
  let drawBurstId = 0
  const drawBurstTimers = new Map<number, ReturnType<typeof setTimeout>>()

  // 行动步骤摘要（桌面展示）
  const actionSummaryLines = ref<string[]>([])
  const combatCue = ref<{
    id: number
    attackerId: string
    targetId: string
    phase: 'attack' | 'defend' | 'take' | 'counter' | 'shield'
  } | null>(null)
  let combatCueId = 0
  const combatCueQueue = ref<Array<{
    attackerId: string
    targetId: string
    phase: 'attack' | 'defend' | 'take' | 'counter' | 'shield'
  }>>([])
  let combatCueTimer: ReturnType<typeof setTimeout> | null = null
  const initiatorFocus = ref<InitiatorFocusState | null>(null)
  let initiatorFocusIdleTimer: ReturnType<typeof setTimeout> | null = null
  let initiatorFocusResolveTimer: ReturnType<typeof setTimeout> | null = null

  // 战斗播报流：用于“动作可读性”增强
  const battleFeed = ref<BattleFeedEntry[]>([])
  let battleFeedId = 0

  // 士气变化复盘：日志提示 + 结算条目
  const moraleHints = ref<MoraleHint[]>([])
  let moraleHintId = 0
  const moraleChanges = ref<MoraleChangeEntry[]>([])
  let moraleChangeId = 0
  const gameEndSnapshot = ref<GameEndSnapshot | null>(null)

  // 演出模式：慢速更易读，快速更流畅
  const cinematicMode = ref(true)

  // 伤害结算暴血特效：{ targetId, damage, damageType, id }
  const damageEffects = ref<Array<{
    id: number
    targetId: string
    targetName: string
    damage: number
    damageType: string
  }>>([])
  let damageEffectsId = 0

  // 伤害通知队列（弹框显示，需要用户确认）
  const damageNotifications = ref<Array<{
    id: number
    targetId: string
    targetName: string
    damage: number
    damageType: string
  }>>([])
  let damageNotificationId = 0

  // 角色/技能数据（从后端获取，替代静态 CHARACTER_INFO）
  const characters = ref<Record<string, CharacterView>>({})

  // UI 状态
  const currentPrompt = ref<Prompt | null>(null)
  const waitingFor = ref<string>('')
  const logs = ref<string[]>([])
  const selectedCards = ref<number[]>([])
  const selectedTargets = ref<string[]>([])
  const promptCounterTarget = ref<string>('')
  const errorMessage = ref<string>('')
  const skillEffectToast = ref<string>('')
  const isConnected = ref(false)
  const isGameEnded = ref(false)
  const gameEndMessage = ref('')

  // 攻击/法术操作状态（选牌后点击对手即发送）
  const actionMode = ref<'none' | 'attack' | 'magic'>('none')
  const magicSubChoice = ref<'none' | 'card' | 'skill'>('none')  // 法术行动子选择：出牌 或 发动技能
  const selectedCardForAction = ref<number | null>(null)

  // 技能发动流程：选技能 -> (选弃牌) -> 选目标 -> 确认
  const availableSkills = ref<AvailableSkill[]>([])
  const skillMode = ref<'none' | 'choosing_skill' | 'choosing_discard' | 'choosing_target'>('none')
  const selectedSkill = ref<AvailableSkill | null>(null)
  const skillTargetIds = ref<string[]>([])
  const skillDiscardIndices = ref<number[]>([])  // 技能弃牌索引

  // 技能详情弹窗：要查看的角色ID（空则关闭）
  const skillModalCharacterId = ref<string | null>(null)
  const skillModalAnchor = ref<SkillModalAnchor | null>(null)

  // 计算属性
  const myPlayer = computed(() => players.value[myPlayerId.value])
  const myHand = computed(() => myPlayer.value?.hand || [])
  const myBlessings = computed(() => myPlayer.value?.blessings || [])
  const myExclusiveCards = computed(() => myPlayer.value?.exclusive_cards || [])
  const myPlayableCards = computed(() =>
    [
      ...myHand.value.map((card, index) => ({
        card,
        index,
        source: 'hand' as const
      })),
      ...myBlessings.value.map((card, index) => ({
        card,
        index: myHand.value.length + index,
        source: 'blessing' as const
      }))
    ]
  )
  const isMyTurn = computed(() => currentPlayer.value === myPlayerId.value)
  const isPromptForMe = computed(() => currentPrompt.value?.player_id === myPlayerId.value)

  // 根据 roleId 获取角色信息
  function getCharacter(roleId: string): CharacterView | null {
    return characters.value[roleId] ?? null
  }

  // 角色显示名（优先后端下发，其次本地回退映射，最后兜底中文）
  function getRoleDisplayName(roleId?: string): string {
    if (!roleId) return '未知角色'
    return characters.value[roleId]?.name || ROLE_NAME_MAP[roleId] || '未知角色'
  }

  function resolveInitiatorFocusSide(playerId: string): InitiatorFocusSide {
    const actorCamp = players.value[playerId]?.camp
    if ((myCamp.value === 'Red' || myCamp.value === 'Blue') && (actorCamp === 'Red' || actorCamp === 'Blue')) {
      return actorCamp === myCamp.value ? 'right' : 'left'
    }
    if (playerId === myPlayerId.value) return 'right'
    return 'left'
  }

  function cancelInitiatorFocusIdleTimer() {
    if (initiatorFocusIdleTimer) {
      clearTimeout(initiatorFocusIdleTimer)
      initiatorFocusIdleTimer = null
    }
  }

  function cancelInitiatorFocusResolveTimer() {
    if (initiatorFocusResolveTimer) {
      clearTimeout(initiatorFocusResolveTimer)
      initiatorFocusResolveTimer = null
    }
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
      startedAt: Date.now()
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

  function syncInitiatorFocusWithState(nextPhase: string) {
    const focus = initiatorFocus.value
    if (!focus) return
    // 玩家阵营可能在首包 state_update 才可用，动态修正左右落位。
    const nextSide = resolveInitiatorFocusSide(focus.playerId)
    if (focus.side !== nextSide) {
      initiatorFocus.value = { ...focus, side: nextSide }
    }

    if (focus.mode === 'attack') {
      if (nextPhase !== 'CombatInteraction') {
        resolveAttackInitiatorFocus(focus.playerId, cinematicMode.value ? 260 : 160)
      }
      return
    }

    if (nextPhase === 'Response' || nextPhase === 'CombatInteraction') {
      touchSkillInitiatorFocus(focus.playerId)
      return
    }

    if (damageEffects.value.length > 0) {
      touchSkillInitiatorFocus(focus.playerId)
      return
    }

    settleSkillInitiatorFocus(focus.playerId, cinematicMode.value ? 420 : 240)
  }

  function setCharacters(list: CharacterView[]) {
    const map: Record<string, CharacterView> = {}
    for (const c of list) {
      map[c.id] = c
    }
    characters.value = map
  }

  const redPlayers = computed(() => 
    Object.values(players.value).filter(p => p.camp === 'Red')
  )
  const bluePlayers = computed(() => 
    Object.values(players.value).filter(p => p.camp === 'Blue')
  )
  const opponentPlayers = computed(() => 
    myCamp.value === 'Red' ? bluePlayers.value : redPlayers.value
  )
  const allyPlayers = computed(() => 
    myCamp.value === 'Red' ? redPlayers.value : bluePlayers.value
  )

  const basicEffectByMagicCardName: Record<string, string> = {
    '中毒': 'Poison',
    '虚弱': 'Weak',
    '圣盾': 'Shield'
  }

  function isBasicEffect(effect?: string | null): boolean {
    return effect === 'Shield' || effect === 'Weak' || effect === 'Poison' ||
      effect === 'SealFire' || effect === 'SealWater' || effect === 'SealEarth' ||
      effect === 'SealWind' || effect === 'SealThunder' ||
      effect === 'PowerBlessing' || effect === 'SwiftBlessing'
  }

  function hasEffect(player: PlayerView | undefined, effect: string): boolean {
    if (!player || !Array.isArray(player.field)) return false
    return player.field.some((fc) => fc?.mode === 'Effect' && fc.effect === effect)
  }

  function selectedMagicBasicEffect(): string {
    if (actionMode.value !== 'magic' || selectedCardForAction.value === null) return ''
    const item = myPlayableCards.value.find((it) => it.index === selectedCardForAction.value)
    if (!item || item.card.type !== 'Magic') return ''
    return basicEffectByMagicCardName[item.card.name] || ''
  }

  const selectedActionCard = computed(() => {
    if (selectedCardForAction.value === null) return null
    return myPlayableCards.value.find((it) => it.index === selectedCardForAction.value)?.card || null
  })

  const selectedActionIsMagicBullet = computed(() =>
    actionMode.value === 'magic' &&
    selectedActionCard.value?.type === 'Magic' &&
    selectedActionCard.value?.name === '魔弹'
  )

  // 法术模式可对己方/自己施放（如圣盾），攻击模式仅对手
  const targetablePlayers = computed(() => {
    if (actionMode.value === 'attack') {
      return opponentPlayers.value.filter((p) =>
        !p.field?.some((fc) => fc.mode === 'Effect' && fc.effect === 'Stealth')
      )
    }
    if (actionMode.value === 'magic') {
      // 魔弹按固定顺序自动结算，不需要手动选择目标
      if (selectedActionIsMagicBullet.value) return []
      const all = Object.values(players.value)
      const basicEffect = selectedMagicBasicEffect()
      if (!basicEffect) return all
      return all.filter((p) => !hasEffect(p, basicEffect))
    }
    return []
  })

  // 根据技能 target_type 返回可选的目标玩家列表 (用于技能发动)
  const targetablePlayersForSkill = computed(() => {
    const skill = selectedSkill.value
    if (!skill) return []
    const all = Object.values(players.value)
    const me = myPlayerId.value
    const camp = myCamp.value
    // target_type: 0=None, 1=Self, 2=Enemy, 3=Ally, 4=AllySelf, 5=Any, 6=Specific
    const baseTargets = (() => {
      switch (skill.target_type) {
        case 0: return []
        case 1: return all.filter(p => p.id === me)
        case 2: return all.filter(p => p.camp !== camp)
        case 3: return all.filter(p => p.camp === camp && p.id !== me)
        case 4: return all.filter(p => p.camp === camp)
        case 5:
        case 6: return all
        default: return all
      }
    })()

    if (skill.id === 'angel_cleanse') {
      return baseTargets.filter((p) =>
        Array.isArray(p.field) && p.field.some((fc) =>
          fc.mode === 'Effect' && isBasicEffect(fc.effect)
        )
      )
    }

    if (skill.id === 'seal_break') {
      return baseTargets.filter((p) =>
        Array.isArray(p.field) && p.field.some((fc) => fc.mode === 'Effect' && isBasicEffect(fc.effect))
      )
    }

    if (skill.place_card && isBasicEffect(skill.place_effect)) {
      return baseTargets.filter((p) => !hasEffect(p, skill.place_effect || ''))
    }

    return baseTargets
  })

  const skillModalCharacter = computed(() => {
    const id = skillModalCharacterId.value
    return id ? (characters.value[id] ?? null) : null
  })

  const moraleBurstRanking = computed(() => {
    return moraleChanges.value
      .filter(item => item.delta < 0)
      .slice()
      .sort((a, b) => {
        const byLoss = Math.abs(b.delta) - Math.abs(a.delta)
        if (byLoss !== 0) return byLoss
        return b.timestamp - a.timestamp
      })
  })

  const canConfirmSkill = computed(() => {
    const skill = selectedSkill.value
    if (!skill) return false
    if (skill.target_type === 0) return true // 无需目标
    const n = skillTargetIds.value.length
    // 当 min/max 为 0 但 target_type 需要选人时，按 1 处理（如封印技）
    const min = (skill.min_targets ?? 0) || (skill.target_type >= 2 ? 1 : 0)
    const max = (skill.max_targets ?? 99) || (skill.target_type >= 2 ? 1 : 99)
    return n >= min && n <= max
  })

  // 卡牌是否匹配独有技（卡牌下标了该技能名）
  function cardMatchesExclusive(
    card: { exclusive_char1?: string; exclusive_char2?: string; exclusive_skill1?: string; exclusive_skill2?: string },
    charName: string,
    skillTitle: string
  ): boolean {
    if (!card || !charName || !skillTitle) return false
    return (
      (card.exclusive_char1 === charName && card.exclusive_skill1 === skillTitle) ||
      (card.exclusive_char2 === charName && card.exclusive_skill2 === skillTitle)
    )
  }

  // 可用技能列表：优先用后端返回的 available_skills，为空时从角色数据构建（fallback）
  const effectiveAvailableSkills = computed((): AvailableSkill[] => {
    if (availableSkills.value.length > 0) return availableSkills.value
    const char = getCharacter(myCharRole.value)
    if (!char?.skills?.length) return []
    const hand = myHand.value
    const exclusiveCards = myExclusiveCards.value
    const charName = char.name
    const actionSkills = char.skills.filter((s: { type?: number }) => (s.type ?? 2) === 2)
    return actionSkills
      .filter((s: SkillView) => {
        if (!s.require_exclusive) return true
        return hand.some((c) => cardMatchesExclusive(c, charName, s.title)) ||
          exclusiveCards.some((c) => cardMatchesExclusive(c, charName, s.title))
      })
      .map((s: SkillView) => {
        const targetType = s.target_type ?? 0
        const minT = s.min_targets ?? 0
        const maxT = s.max_targets ?? 0
        return {
        id: s.id,
        title: s.title,
        description: s.description,
        min_targets: minT || (targetType >= 2 ? 1 : 0),
        max_targets: maxT || (targetType >= 2 ? 1 : 1),
        target_type: targetType,
        cost_gem: s.cost_gem ?? 0,
        cost_crystal: s.cost_crystal ?? 0,
        cost_discards: s.cost_discards ?? 0,
        discard_element: s.discard_element,
        require_exclusive: s.require_exclusive,
      }
      })
  })

  // 方法
  function setRoomInfo(code: string, playerId: string, camp: string, charRole: string) {
    roomCode.value = code
    myPlayerId.value = playerId
    myCamp.value = camp
    myCharRole.value = charRole
    isInRoom.value = true
  }

  function setReconnectToken(token: string) {
    reconnectToken.value = token || ''
  }

  function updateRoomPlayers(playerList: PlayerInfo[]) {
    roomPlayers.value = playerList
    const me = playerList.find(p => p.id === myPlayerId.value)
    if (me) {
      if (me.camp) myCamp.value = me.camp
      if (me.char_role) myCharRole.value = me.char_role
    }
  }

  function setGameStarted() {
    gameStarted.value = true
    isGameEnded.value = false
    gameEndMessage.value = ''
    moraleHints.value = []
    moraleChanges.value = []
    gameEndSnapshot.value = null
  }

  function updateGameState(state: GameStateUpdate) {
    phase.value = state.phase
    // 任意状态刷新都说明流程有推进，清理“等待某玩家”提示，避免UI残留“机器人思考中”。
    waitingFor.value = ''
    // 对战提示在战斗交互阶段常驻显示，离开该阶段后清除
    if (state.phase !== 'CombatInteraction' && combatCueQueue.value.length === 0) {
      combatCue.value = null
      if (initiatorFocus.value?.mode === 'attack') {
        resolveAttackInitiatorFocus(initiatorFocus.value.playerId, cinematicMode.value ? 260 : 160)
      }
    }
    currentPlayer.value = state.current_player
    hasPerformedStartup.value = state.has_performed_startup ?? false
    players.value = state.players
    // 收到 state_update 表示上一动作已被服务器成功处理，清除所有操作中的 UI 状态
    currentPrompt.value = null
    selectedCards.value = []
    selectedTargets.value = []
    promptCounterTarget.value = ''
    actionMode.value = 'none'
    magicSubChoice.value = 'none'
    selectedCardForAction.value = null
    skillMode.value = 'none'
    selectedSkill.value = null
    skillTargetIds.value = []
    skillDiscardIndices.value = []
    // 从 state 同步 myCharRole（游戏中的权威数据）
    const me = state.players[myPlayerId.value]
    if (me?.role) myCharRole.value = me.role
    redMorale.value = state.red_morale
    blueMorale.value = state.blue_morale
    redCups.value = state.red_cups
    blueCups.value = state.blue_cups
    redGems.value = state.red_gems
    blueGems.value = state.blue_gems
    redCrystals.value = state.red_crystals
    blueCrystals.value = state.blue_crystals
    deckCount.value = state.deck_count
    discardCount.value = state.discard_count ?? 0
    availableSkills.value = state.available_skills ?? []
    if (state.characters?.length) {
      setCharacters(state.characters)
    }
    syncInitiatorFocusWithState(state.phase)
    // 终局后仍可能收到一次最终状态推送，刷新复盘快照中的最终面板数据
    if (isGameEnded.value) {
      gameEndSnapshot.value = buildGameEndSnapshot(gameEndMessage.value || '游戏结束')
    }
  }

  function setPrompt(prompt: Prompt | null) {
    currentPrompt.value = prompt
    selectedCards.value = []
    selectedTargets.value = []
    promptCounterTarget.value = ''
    if (prompt?.player_id) {
      touchSkillInitiatorFocus(prompt.player_id)
    }
    // 进入中断/交互提示时，清理本地行动态，避免继续发送 Skill/Attack 指令
    if (prompt) {
      actionMode.value = 'none'
      magicSubChoice.value = 'none'
      selectedCardForAction.value = null
      skillMode.value = 'none'
      selectedSkill.value = null
      skillTargetIds.value = []
      skillDiscardIndices.value = []
    }
  }

  function setPromptCounterTarget(playerId: string) {
    promptCounterTarget.value = playerId
  }

  function setWaiting(playerId: string) {
    waitingFor.value = playerId
  }

  function addLog(message: string) {
    logs.value.push(message)
    // 只保留最近100条日志
    if (logs.value.length > 100) {
      logs.value = logs.value.slice(-100)
    }
  }

  function clearLogs() {
    logs.value = []
  }

  function toggleCardSelection(index: number) {
    const idx = selectedCards.value.indexOf(index)
    if (idx === -1) {
      // 如果是单选，清空之前的选择
      if (currentPrompt.value && currentPrompt.value.max === 1) {
        selectedCards.value = [index]
      } else {
        selectedCards.value.push(index)
      }
    } else {
      selectedCards.value.splice(idx, 1)
    }
  }

  function selectTarget(playerId: string) {
    const idx = selectedTargets.value.indexOf(playerId)
    if (idx >= 0) {
      selectedTargets.value.splice(idx, 1)
    } else {
      if (currentPrompt.value && currentPrompt.value.max === 1) {
        selectedTargets.value = [playerId]
      } else {
        selectedTargets.value.push(playerId)
      }
    }
  }

  function setActionModeForAttack(mode: 'none' | 'attack' | 'magic') {
    actionMode.value = mode
    // 不清空已选牌，支持「先选牌再点攻击」
    if (mode === 'none') {
      selectedCardForAction.value = null
    }
  }

  function setSelectedCardForAction(idx: number | null) {
    selectedCardForAction.value = idx
  }

  function canTargetOpponent() {
    if (actionMode.value === 'magic' && selectedActionIsMagicBullet.value) return false
    return actionMode.value !== 'none' && selectedCardForAction.value !== null
  }

  function setMagicSubChoice(choice: 'none' | 'card' | 'skill') {
    magicSubChoice.value = choice
  }

  function clearActionMode() {
    actionMode.value = 'none'
    magicSubChoice.value = 'none'
    selectedCardForAction.value = null
  }

  function setSkillMode(mode: 'none' | 'choosing_skill' | 'choosing_discard' | 'choosing_target') {
    skillMode.value = mode
    if (mode === 'none') {
      selectedSkill.value = null
      skillTargetIds.value = []
      skillDiscardIndices.value = []
    }
  }

  function setSelectedSkill(skill: AvailableSkill | null) {
    selectedSkill.value = skill
    skillTargetIds.value = []
    skillDiscardIndices.value = []
  }

  function toggleSkillTarget(playerId: string) {
    const idx = skillTargetIds.value.indexOf(playerId)
    if (idx === -1) {
      skillTargetIds.value = [...skillTargetIds.value, playerId]
    } else {
      skillTargetIds.value = skillTargetIds.value.filter(id => id !== playerId)
    }
  }

  function toggleSkillDiscard(cardIndex: number) {
    const idx = skillDiscardIndices.value.indexOf(cardIndex)
    if (idx === -1) {
      skillDiscardIndices.value = [...skillDiscardIndices.value, cardIndex]
    } else {
      skillDiscardIndices.value = skillDiscardIndices.value.filter(i => i !== cardIndex)
    }
  }

  function clearSkillMode() {
    setSkillMode('none')
  }

  function setError(message: string) {
    errorMessage.value = message
    setTimeout(() => {
      errorMessage.value = ''
    }, 3000)
  }

  function setSkillEffectToast(message: string) {
    skillEffectToast.value = message
    setTimeout(() => {
      skillEffectToast.value = ''
    }, 2500)
  }

  function openSkillModal(characterId: string | null, anchor?: SkillModalAnchor | null) {
    skillModalCharacterId.value = characterId
    skillModalAnchor.value = characterId ? (anchor ?? null) : null
  }

  function resolveFlyingHoldMode(actionType: string): 'timed' | 'until_response' | 'until_next_card_or_draw' {
    // 任何与对战相关的牌（攻击/法术/抵挡/应战/技能），都在屏幕中心悬浮等待，直到响应结束
    if (actionType !== 'discard') return 'until_response'
    return 'until_next_card_or_draw'
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
    // 收到别人打出的响应牌时，不要清空屏幕上的攻击牌，让它们同框出现
    if (kind === 'card_revealed' && (actionType === 'defend' || actionType === 'counter')) {
      return // 不在此处清空，等真正的 combat 结束再清空
    }

    // 真正的对局结束（承受伤害或下一回合开始）才清空
    if (kind === 'damage' || kind === 'combat_response') {
      dropActiveFlyingCards()
      pumpFlyingCards()
      return
    }
    
    // 如果之前有悬挂的卡牌，并且不是战斗相关，清空
    if (kind === 'draw' || kind === 'card_revealed') {
      dropActiveFlyingCards()
      pumpFlyingCards()
    }
  }

  function addFlyingCards(cards: Card[], playerId: string, playerName: string, actionType: string, hidden?: boolean) {
    if (!cards?.length) return
    notifyFlyingCardsEvent('card_revealed', actionType)
    flyingCardsQueue.value.push({
      cards,
      playerId,
      playerName,
      actionType,
      holdMode: resolveFlyingHoldMode(actionType),
      hidden
    })
    pumpFlyingCards()
  }

  function pumpFlyingCards() {
    if (flyingCards.value.length > 0 || flyingCardsQueue.value.length === 0) return
    const next = flyingCardsQueue.value.shift()
    if (!next) return

    flyingCardsId++
    const id = flyingCardsId
    // 如果是对方响应（抵挡/应战）或任何响应牌，都堆叠在屏幕中间
    if (next.holdMode === 'until_response') {
      flyingCards.value = [...flyingCards.value, { id, ...next }]
    } else {
      flyingCards.value = [{ id, ...next }]
    }

    if (next.holdMode === 'timed') {
      const displayMs = cinematicMode.value ? 2400 : 1600
      if (flyingCardsTimer) clearTimeout(flyingCardsTimer)
      flyingCardsTimer = setTimeout(() => {
        flyingCards.value = flyingCards.value.filter(f => f.id !== id)
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
      count
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
      damageType
    })
    // 1.5秒后移除（暴血动画完成后）
    setTimeout(() => {
      damageEffects.value = damageEffects.value.filter(d => d.id !== id)
    }, 1500)
  }

  function addDamageNotification(targetId: string, targetName: string, damage: number, damageType: string) {
    if (damage <= 0) return
    damageNotificationId++
    damageNotifications.value.push({
      id: damageNotificationId,
      targetId,
      targetName,
      damage,
      damageType
    })
  }

  function confirmDamageNotification() {
    if (damageNotifications.value.length > 0) {
      damageNotifications.value.shift()
    }
  }

  function addActionStep(line: string) {
    if (!line) return
    actionSummaryLines.value.push(line)
    if (actionSummaryLines.value.length > 12) {
      actionSummaryLines.value = actionSummaryLines.value.slice(-12)
    }
  }

  function clearActionSummary() {
    actionSummaryLines.value = []
  }

  function addCombatCue(attackerId: string, targetId: string, phase: 'attack' | 'defend' | 'take' | 'counter' | 'shield') {
    if (!attackerId || !targetId) return
    if (phase === 'defend' || phase === 'take' || phase === 'counter' || phase === 'shield') {
      notifyFlyingCardsEvent('combat_response')
      resolveAttackInitiatorFocus(attackerId)
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
        phase
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
        phase
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
      phase
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

  function addBattleFeed(entry: Omit<BattleFeedEntry, 'id' | 'timestamp'>) {
    const now = Date.now()
    const last = battleFeed.value[battleFeed.value.length - 1]
    if (last && last.title === entry.title && last.detail === entry.detail && now - last.timestamp < 280) {
      return
    }
    battleFeedId++
    battleFeed.value.push({
      id: battleFeedId,
      timestamp: now,
      ...entry
    })
    if (battleFeed.value.length > 80) {
      battleFeed.value = battleFeed.value.slice(-80)
    }
  }

  function clearBattleFeed() {
    battleFeed.value = []
  }

  function setCinematicMode(enabled: boolean) {
    cinematicMode.value = enabled
  }

  function setConnected(connected: boolean) {
    isConnected.value = connected
  }

  function pushMoraleHint(hint: Omit<MoraleHint, 'id' | 'timestamp'>) {
    moraleHintId++
    moraleHints.value.push({
      id: moraleHintId,
      timestamp: Date.now(),
      ...hint
    })
    if (moraleHints.value.length > 30) {
      moraleHints.value = moraleHints.value.slice(-30)
    }
  }

  function consumeMoraleHint(camp: MoraleCamp, expectedLoss?: number): MoraleHint | null {
    const now = Date.now()
    moraleHints.value = moraleHints.value.filter(h => now - h.timestamp <= 20000)

    for (let i = moraleHints.value.length - 1; i >= 0; i--) {
      const hint = moraleHints.value[i]
      if (!hint) continue
      const campMatch = !hint.camp || hint.camp === camp
      const lossMatch = expectedLoss === undefined || !hint.loss || hint.loss === expectedLoss
      if (campMatch && lossMatch) {
        moraleHints.value.splice(i, 1)
        return hint
      }
    }
    return null
  }

  function recordMoraleChange(
    camp: MoraleCamp,
    before: number,
    after: number,
    hint?: MoraleHint | null
  ) {
    if (before === after) return
    moraleChangeId++
    const delta = after - before
    moraleChanges.value.push({
      id: moraleChangeId,
      timestamp: Date.now(),
      camp,
      delta,
      before,
      after,
      source: hint?.source || (delta < 0 ? '未知来源（扣士气）' : '未知来源（恢复士气）'),
      raw: hint?.raw
    })
    if (moraleChanges.value.length > 120) {
      moraleChanges.value = moraleChanges.value.slice(-120)
    }
  }

  function buildGameEndSnapshot(message: string): GameEndSnapshot {
    const triggerType: GameEndSnapshot['triggerType'] =
      redCups.value >= 5 || blueCups.value >= 5
        ? 'cups'
        : redMorale.value <= 0 || blueMorale.value <= 0
          ? 'morale'
          : 'unknown'
    const triggerCamp: MoraleCamp | undefined =
      redMorale.value <= 0
        ? 'Red'
        : blueMorale.value <= 0
          ? 'Blue'
          : undefined

    const triggerEntry = [...moraleChanges.value]
      .reverse()
      .find(item => (triggerCamp ? item.camp === triggerCamp : true) && item.delta < 0)

    return {
      message: message || '游戏结束',
      triggerType,
      finalRedMorale: redMorale.value,
      finalBlueMorale: blueMorale.value,
      finalRedCups: redCups.value,
      finalBlueCups: blueCups.value,
      triggerCamp: triggerEntry?.camp,
      triggerDelta: triggerEntry ? Math.abs(triggerEntry.delta) : undefined,
      triggerSource: triggerEntry?.source
    }
  }

  function setGameEnded(message: string) {
    isGameEnded.value = true
    gameEndMessage.value = message || '游戏结束'
    gameEndSnapshot.value = buildGameEndSnapshot(gameEndMessage.value)
    // 游戏结束后立即清理交互态，避免界面残留在“等待/响应”状态
    currentPrompt.value = null
    waitingFor.value = ''
    selectedCards.value = []
    selectedTargets.value = []
    promptCounterTarget.value = ''
    actionMode.value = 'none'
    magicSubChoice.value = 'none'
    selectedCardForAction.value = null
    skillMode.value = 'none'
    selectedSkill.value = null
    skillTargetIds.value = []
    skillDiscardIndices.value = []
    damageNotifications.value = []
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

  function clearGameEnded() {
    isGameEnded.value = false
    gameEndMessage.value = ''
    gameEndSnapshot.value = null
  }

  function reset() {
    roomCode.value = ''
    myPlayerId.value = ''
    myCamp.value = ''
    myCharRole.value = ''
    reconnectToken.value = ''
    characters.value = {}
    roomPlayers.value = []
    isInRoom.value = false
    gameStarted.value = false
    phase.value = ''
    currentPlayer.value = ''
    hasPerformedStartup.value = false
    players.value = {}
    redMorale.value = 15
    blueMorale.value = 15
    redCups.value = 0
    blueCups.value = 0
    redGems.value = 0
    blueGems.value = 0
    redCrystals.value = 0
    blueCrystals.value = 0
    deckCount.value = 0
    discardCount.value = 0
    flyingCards.value = []
    flyingCardsQueue.value = []
    if (flyingCardsTimer) {
      clearTimeout(flyingCardsTimer)
      flyingCardsTimer = null
    }
    damageEffects.value = []
    damageNotifications.value = []
    drawBursts.value = []
    for (const timer of drawBurstTimers.values()) clearTimeout(timer)
    drawBurstTimers.clear()
    actionSummaryLines.value = []
    combatCue.value = null
    combatCueQueue.value = []
    if (combatCueTimer) {
      clearTimeout(combatCueTimer)
      combatCueTimer = null
    }
    clearInitiatorFocus()
    battleFeed.value = []
    moraleHints.value = []
    moraleChanges.value = []
    gameEndSnapshot.value = null
    currentPrompt.value = null
    waitingFor.value = ''
    logs.value = []
    selectedCards.value = []
    selectedTargets.value = []
    promptCounterTarget.value = ''
    errorMessage.value = ''
    skillEffectToast.value = ''
    isConnected.value = false
    isGameEnded.value = false
    gameEndMessage.value = ''
    actionMode.value = 'none'
    magicSubChoice.value = 'none'
    selectedCardForAction.value = null
    availableSkills.value = []
    skillMode.value = 'none'
    selectedSkill.value = null
    skillTargetIds.value = []
    skillDiscardIndices.value = []
    skillModalCharacterId.value = null
    skillModalAnchor.value = null
    cinematicMode.value = true
  }

  return {
    // 房间状态
    roomCode,
    myPlayerId,
    myName,
    myCamp,
    myCharRole,
    reconnectToken,
    roomPlayers,
    isInRoom,
    gameStarted,
    
    // 游戏状态
    phase,
    currentPlayer,
    hasPerformedStartup,
    players,
    redMorale,
    blueMorale,
    redCups,
    blueCups,
    redGems,
    blueGems,
    redCrystals,
    blueCrystals,
    deckCount,
    discardCount,
    flyingCards,
    drawBursts,
    damageEffects,
    damageNotifications,
    battleFeed,
    moraleChanges,
    moraleBurstRanking,
    gameEndSnapshot,
    cinematicMode,

    // UI 状态
    currentPrompt,
    waitingFor,
    logs,
    selectedCards,
    selectedTargets,
    promptCounterTarget,
    errorMessage,
    skillEffectToast,
    isConnected,
    isGameEnded,
    gameEndMessage,
    actionMode,
    magicSubChoice,
    selectedCardForAction,
    availableSkills,
    effectiveAvailableSkills,
    skillMode,
    selectedSkill,
    skillTargetIds,
    skillDiscardIndices,
    targetablePlayersForSkill,
    canConfirmSkill,

    // 技能弹窗
    skillModalCharacterId,
    skillModalAnchor,
    skillModalCharacter,
    openSkillModal,
    addFlyingCards,
    addDrawBurst,
    addDamageEffect,
    addDamageNotification,
    confirmDamageNotification,
    actionSummaryLines,
    addActionStep,
    clearActionSummary,
    combatCue,
    addCombatCue,
    initiatorFocus,
    startAttackInitiatorFocus,
    resolveAttackInitiatorFocus,
    startSkillInitiatorFocus,
    touchSkillInitiatorFocus,
    settleSkillInitiatorFocus,
    clearInitiatorFocus,
    syncInitiatorFocusWithState,
    addBattleFeed,
    clearBattleFeed,
    setCinematicMode,

    // 角色数据
    characters,
    getCharacter,
    getRoleDisplayName,

    // 计算属性
    myPlayer,
    myHand,
    myBlessings,
    myExclusiveCards,
    myPlayableCards,
    isMyTurn,
    isPromptForMe,
    redPlayers,
    bluePlayers,
    opponentPlayers,
    allyPlayers,
    targetablePlayers,
    
    // 方法
    setRoomInfo,
    setReconnectToken,
    updateRoomPlayers,
    setCharacters,
    setGameStarted,
    updateGameState,
    setPrompt,
    setWaiting,
    addLog,
    clearLogs,
    toggleCardSelection,
    selectTarget,
    setPromptCounterTarget,
    setActionModeForAttack,
    setMagicSubChoice,
    setSelectedCardForAction,
    canTargetOpponent,
    clearActionMode,
    setSkillMode,
    setSelectedSkill,
    toggleSkillTarget,
    toggleSkillDiscard,
    clearSkillMode,
    setError,
    setSkillEffectToast,
    setConnected,
    pushMoraleHint,
    consumeMoraleHint,
    recordMoraleChange,
    setGameEnded,
    clearGameEnded,
    reset,
    cardMatchesExclusive
  }
})
