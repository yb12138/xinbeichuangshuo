import { storeToRefs } from 'pinia'
import { computed } from 'vue'
import { ROLE_NAME_MAP } from '../constants/roleNameMap'
import { useInterruptStore } from '../stores/interrupt.store'
import { useSessionStore } from '../stores/session.store'
import { useSnapshotStore } from '../stores/snapshot.store'
import type { AvailableSkill, CharacterView, PlayerView, SkillView } from '../types/game'

const BASIC_EFFECT_BY_MAGIC_CARD_NAME: Record<string, string> = {
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

export function useBattleInteractionState() {
  const sessionStore = useSessionStore()
  const snapshotStore = useSnapshotStore()
  const interruptStore = useInterruptStore()

  const { myPlayerId, myCamp, myCharRole } = storeToRefs(sessionStore)
  const { currentPlayer, turnStage, players, characters, availableSkills } = storeToRefs(snapshotStore)
  const {
    currentPrompt,
    actionMode,
    selectedCardForAction,
    selectedSkill,
    skillTargetIds,
  } = storeToRefs(interruptStore)

  const myPlayer = computed(() => players.value[myPlayerId.value] || null)
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

  function getRoleDisplayName(roleId?: string): string {
    if (!roleId) return '未知角色'
    return characters.value[roleId]?.name || ROLE_NAME_MAP[roleId] || '未知角色'
  }

  function getCharacter(roleId: string): CharacterView | null {
    return characters.value[roleId] ?? null
  }

  function cardMatchesExclusive(
    card: {
      exclusive_char1?: string
      exclusive_char2?: string
      exclusive_skill1?: string
      exclusive_skill2?: string
    },
    roleId: string,
    skillTitle: string,
    roleName?: string
  ): boolean {
    if (!card || !roleId || !skillTitle) return false
    const normalize = (text?: string): string => String(text || '').trim()
    const roleCandidates = new Set<string>()
    const pushCandidate = (value?: string) => {
      const raw = normalize(value)
      if (!raw) return
      roleCandidates.add(raw)
      roleCandidates.add(raw.toLowerCase())
      // 兼容：如果传入的是中文名，补充对应角色 ID；如果传入的是角色 ID，补充中文名。
      for (const [id, name] of Object.entries(ROLE_NAME_MAP)) {
        if (raw === id || raw === id.toLowerCase()) {
          roleCandidates.add(name)
          break
        }
        if (raw === name) {
          roleCandidates.add(id)
          roleCandidates.add(id.toLowerCase())
          break
        }
      }
    }
    pushCandidate(roleId)
    pushCandidate(roleName)

    const cardChar1 = normalize(card.exclusive_char1)
    const cardChar2 = normalize(card.exclusive_char2)
    const skill1 = normalize(card.exclusive_skill1)
    const skill2 = normalize(card.exclusive_skill2)
    const expectedSkill = normalize(skillTitle)
    return (
      (roleCandidates.has(cardChar1) && skill1 === expectedSkill) ||
      (roleCandidates.has(cardChar1.toLowerCase()) && skill1 === expectedSkill) ||
      (roleCandidates.has(cardChar2) && skill2 === expectedSkill) ||
      (roleCandidates.has(cardChar2.toLowerCase()) && skill2 === expectedSkill)
    )
  }

  const selectedActionCard = computed(() => {
    if (selectedCardForAction.value === null) return null
    return myPlayableCards.value.find((item) => item.index === selectedCardForAction.value)?.card || null
  })

  const selectedActionIsMagicBullet = computed(() =>
    actionMode.value === 'magic' &&
    selectedActionCard.value?.type === 'Magic' &&
    selectedActionCard.value?.name === '魔弹'
  )

  function selectedMagicBasicEffect(): string {
    if (actionMode.value !== 'magic' || selectedCardForAction.value === null) return ''
    const item = myPlayableCards.value.find((it) => it.index === selectedCardForAction.value)
    if (!item || item.card.type !== 'Magic') return ''
    return BASIC_EFFECT_BY_MAGIC_CARD_NAME[item.card.name] || ''
  }

  const redPlayers = computed(() =>
    Object.values(players.value).filter((player) => player.camp === 'Red')
  )
  const bluePlayers = computed(() =>
    Object.values(players.value).filter((player) => player.camp === 'Blue')
  )
  const opponentPlayers = computed(() =>
    myCamp.value === 'Red' ? bluePlayers.value : redPlayers.value
  )

  const targetablePlayers = computed(() => {
    if (actionMode.value === 'attack') {
      return opponentPlayers.value.filter((player) =>
        !player.field?.some((fieldCard) => fieldCard.mode === 'Effect' && fieldCard.effect === 'Stealth')
      )
    }
    if (actionMode.value === 'magic') {
      if (selectedActionIsMagicBullet.value) return []
      const allPlayers = Object.values(players.value)
      const basicEffect = selectedMagicBasicEffect()
      if (!basicEffect) return allPlayers
      return allPlayers.filter((player) => !hasEffect(player, basicEffect))
    }
    return []
  })

  const targetablePlayersForSkill = computed(() => {
    const skill = selectedSkill.value
    if (!skill) return []
    const allPlayers = Object.values(players.value)
    const me = myPlayerId.value
    const camp = myCamp.value
    const baseTargets = (() => {
      switch (skill.target_type) {
        case 0:
          return []
        case 1:
          return allPlayers.filter((player) => player.id === me)
        case 2:
          return allPlayers.filter((player) => player.camp !== camp)
        case 3:
          return allPlayers.filter((player) => player.camp === camp && player.id !== me)
        case 4:
          return allPlayers.filter((player) => player.camp === camp)
        case 5:
        case 6:
          return allPlayers
        default:
          return allPlayers
      }
    })()

    if (skill.id === 'seal_break') {
      return baseTargets.filter((player) =>
        Array.isArray(player.field) && player.field.some((fieldCard) =>
          fieldCard.mode === 'Effect' && isBasicEffect(fieldCard.effect)
        )
      )
    }

    if (skill.place_card && isBasicEffect(skill.place_effect)) {
      return baseTargets.filter((player) => !hasEffect(player, skill.place_effect || ''))
    }

    return baseTargets
  })

  const canTargetOpponent = computed(() =>
    actionMode.value !== 'none' &&
    selectedCardForAction.value !== null &&
    !(actionMode.value === 'magic' && selectedActionIsMagicBullet.value)
  )

  const effectiveAvailableSkills = computed((): AvailableSkill[] => {
    // 行动阶段且轮到自己时，主动技可用性以后端 available_skills 为准（包含空列表）。
    // 避免前端 fallback 重新“猜”规则导致与后端可用态漂移。
    if (isMyTurn.value && turnStage.value === 'ActionExecution') {
      return availableSkills.value
    }
    if (availableSkills.value.length > 0) return availableSkills.value
    const roleLookup = myCharRole.value || myPlayer.value?.role || ''
    const char = getCharacter(roleLookup)
    if (!char?.skills?.length) return []
    const roleId = myCharRole.value || char.id
    const roleName = char.name
    const actionSkills = char.skills.filter((skill: { type?: number }) => (skill.type ?? 2) === 2)
    return actionSkills
      .filter((skill: SkillView) => {
        if (!skill.require_exclusive) return true
        return myHand.value.some((card) => cardMatchesExclusive(card, roleId, skill.title, roleName)) ||
          myExclusiveCards.value.some((card) => cardMatchesExclusive(card, roleId, skill.title, roleName))
      })
      .map((skill: SkillView) => {
        const targetType = skill.target_type ?? 0
        const minTargets = skill.min_targets ?? 0
        const maxTargets = skill.max_targets ?? 0
        return {
          id: skill.id,
          title: skill.title,
          description: skill.description,
          min_targets: minTargets || (targetType >= 2 ? 1 : 0),
          max_targets: maxTargets || (targetType >= 2 ? 1 : 1),
          target_type: targetType,
          cost_gem: skill.cost_gem ?? 0,
          cost_crystal: skill.cost_crystal ?? 0,
          cost_discards: skill.cost_discards ?? 0,
          discard_element: skill.discard_element,
          require_exclusive: skill.require_exclusive,
        }
      })
  })

  const canConfirmSkill = computed(() => {
    const skill = selectedSkill.value
    if (!skill) return false
    if (skill.target_type === 0) return true
    const targetCount = skillTargetIds.value.length
    const minTargets = (skill.min_targets ?? 0) || (skill.target_type >= 2 ? 1 : 0)
    const maxTargets = (skill.max_targets ?? 99) || (skill.target_type >= 2 ? 1 : 99)
    return targetCount >= minTargets && targetCount <= maxTargets
  })

  return {
    myPlayer,
    myHand,
    myBlessings,
    myExclusiveCards,
    myPlayableCards,
    isMyTurn,
    isPromptForMe,
    selectedActionCard,
    selectedActionIsMagicBullet,
    targetablePlayers,
    targetablePlayersForSkill,
    canTargetOpponent,
    effectiveAvailableSkills,
    canConfirmSkill,
    getCharacter,
    getRoleDisplayName,
    cardMatchesExclusive,
  }
}
