<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useInterruptStore } from '../stores/interrupt.store'
import { useSessionStore } from '../stores/session.store'
import { useSnapshotStore } from '../stores/snapshot.store'
import { useSubmitAction } from '../composables/useSubmitAction'
import { ROLE_NAME_MAP } from '../constants/roleNameMap'
import {
  promptImageButtonKindByOption,
  type PromptImageButtonKind,
} from '../constants/promptButtonRules'
import type { PlayerView, PromptOption } from '../types/game'

const interruptStore = useInterruptStore()
const sessionStore = useSessionStore()
const snapshotStore = useSnapshotStore()
const actions = useSubmitAction()

const prompt = computed(() => interruptStore.currentPrompt)
const myPlayerId = computed(() => sessionStore.myPlayerId)
const playerViews = computed(() => snapshotStore.players)
const myHand = computed(() => playerViews.value[myPlayerId.value]?.hand || [])

function getRoleDisplayName(roleId?: string): string {
  if (!roleId) return '未知角色'
  return snapshotStore.characters[roleId]?.name || ROLE_NAME_MAP[roleId] || '未知角色'
}

function showPromptError(message: string) {
  interruptStore.showError(message)
}

// 行动选择（攻击/法术/购买/提取/合成）不在这里显示，由 ActionPanel 承载
const isActionSelectionPrompt = computed(() => {
  if (!prompt.value) return false
  return prompt.value.presentation?.kind === 'action_hub'
})

const isVisible = computed(() =>
  prompt.value !== null && prompt.value.player_id === myPlayerId.value && !isActionSelectionPrompt.value
)

const selectedExtractIndices = ref<number[]>([])
const selectedInlineCardIDs = ref<string[]>([])
const autoResolvedPromptKey = ref('')

watch(() => prompt.value, () => {
  interruptStore.setPromptCounterTarget('')
  selectedExtractIndices.value = []
  selectedInlineCardIDs.value = []
  if (!prompt.value) {
    autoResolvedPromptKey.value = ''
  }
})

const hasCounterOption = computed(() => {
  if (!prompt.value?.options?.length) return false
  return prompt.value.options.some((option) => promptOptionResponseKind(option) === 'counter')
})

const hasDefendOption = computed(() => {
  if (!prompt.value?.options?.length) return false
  return prompt.value.options.some((option) => promptOptionResponseKind(option) === 'defend')
})

const hasCounterOrDefend = computed(() => {
  if (!prompt.value?.options?.length) return false
  return hasCounterOption.value || hasDefendOption.value
})

const responsePromptUnrespondable = computed(() => {
  if (!prompt.value || !hasCounterOrDefend.value) return false
  const effectHints = Array.isArray(prompt.value.effect_hints) ? prompt.value.effect_hints : []
  return effectHints.some((hint) => String(hint || '').includes('无法应战'))
})

type ResponseActionKind = 'take' | 'counter' | 'defend' | null

const isResponsePrompt = computed(() => prompt.value?.presentation?.kind === 'response')

function responseOptionKindByID(optionId: string): ResponseActionKind {
  const id = String(optionId || '').trim().toLowerCase()
  if (id === 'take' || id === 'take_damage') return 'take'
  if (id === 'counter') return 'counter'
  if (id === 'defend') return 'defend'
  return null
}

function promptOptionResponseKind(option: { id?: string }): ResponseActionKind {
  if (!isResponsePrompt.value) return null
  return responseOptionKindByID(option.id || '')
}

function promptAttackElementName(raw: string): string {
  const lower = String(raw || '').trim().toLowerCase()
  if (lower === 'water') return '水系'
  if (lower === 'fire') return '火系'
  if (lower === 'earth') return '地系'
  if (lower === 'wind') return '风系'
  if (lower === 'thunder') return '雷系'
  if (lower === 'light') return '光系'
  if (lower === 'dark') return '暗灭'
  return String(raw || '').trim()
}

const responseAttackElementHintText = computed(() => {
  if (!prompt.value || !hasCounterOrDefend.value) return ''
  const attackElement = String(prompt.value.attack_element || '').trim()
  if (!attackElement) return ''
  const displayName = promptAttackElementName(attackElement)
  if (!displayName) return ''
  if (responsePromptUnrespondable.value || !hasCounterOption.value) return `此次攻击系别：${displayName}（无法应战）`
  if (hasCounterOption.value) return `此次攻击系别：${displayName}（应战需同系或暗灭）`
  return `此次攻击系别：${displayName}`
})

const isDarkAttackResponsePrompt = computed(() => {
  if (!prompt.value || !hasCounterOrDefend.value) return false
  return String(prompt.value.attack_element || '').trim().toLowerCase() === 'dark'
})

const isMagicMissilePrompt = computed(() => {
  const p = prompt.value?.presentation
  return p?.kind === 'response' && p?.layout === 'magic_missile'
})

const isPlagueDeathTouchElementPrompt = computed(() =>
  prompt.value?.presentation?.kind === 'card_picker' && prompt.value?.presentation?.card_filter === 'plague_death_touch_element'
)

const isElfElementalShotPickPrompt = computed(() =>
  prompt.value?.presentation?.kind === 'card_picker' && prompt.value?.presentation?.card_filter === 'magic_or_elf_blessing'
)

const needsCardSelection = computed(() => {
  if (!prompt.value) return false
  if (isSystemBranchPromptChoice()) return false
  const p = prompt.value.presentation
  // card_picker prompts require card selection
  if (p?.kind === 'card_picker') return true
  if (promptHasHandCardOptions.value) return true
  if (hasCounterOrDefend.value) return true
  return false
})

const needsTargetSelection = computed(() => {
  if (!prompt.value) return false
  return prompt.value.presentation?.kind === 'target_picker'
})

const promptRequiresManualTargetConfirm = computed(() => {
  if (!prompt.value || prompt.value.presentation?.kind !== 'target_picker') return false
  // Multi-target pickers require a manual confirm step
  return !!prompt.value.presentation?.multi_target
})

const needsCounterTargetSelection = computed(() => {
  if (!prompt.value) return false
  const ids = prompt.value.counter_target_ids
  return hasCounterOrDefend.value && ids && ids.length > 0
})

const isConfirmType = computed(() => {
  if (!prompt.value) return false
  const kind = prompt.value.presentation?.kind
  return !!kind && kind !== 'card_picker' && kind !== 'target_picker' && kind !== 'action_hub'
})

const isExtractPrompt = computed(() => prompt.value?.presentation?.layout === 'extract')

// 圣疗 3 点治疗分配：每个目标独立 0..3 数字选择，要求总和=3。
const isSaintHealAllocatePrompt = computed(() => prompt.value?.presentation?.layout === 'heal_allocate')
const saintHealAllocations = ref<number[]>([])
const SAINT_HEAL_TOTAL = 3

// 符文改造分配：战纹/魔纹 0..3 数字选择，要求总和=3。
const isRuneReforgeAllocatePrompt = computed(() => prompt.value?.presentation?.layout === 'rune_allocate')
const runeReforgeAllocations = ref<number[]>([])
const RUNE_REFORGE_TOTAL = 3

watch(
  () => prompt.value,
  () => {
    if (isSaintHealAllocatePrompt.value && prompt.value) {
      saintHealAllocations.value = prompt.value.options.map(() => 0)
    } else {
      saintHealAllocations.value = []
    }
    if (isRuneReforgeAllocatePrompt.value && prompt.value) {
      runeReforgeAllocations.value = prompt.value.options.map(() => 0)
    } else {
      runeReforgeAllocations.value = []
    }
  },
  { immediate: true }
)

const saintHealRemaining = computed(() => {
  const used = saintHealAllocations.value.reduce((s, v) => s + (v || 0), 0)
  return SAINT_HEAL_TOTAL - used
})

const runeReforgeRemaining = computed(() => {
  const used = runeReforgeAllocations.value.reduce((s, v) => s + (v || 0), 0)
  return RUNE_REFORGE_TOTAL - used
})

function setSaintHealAllocation(index: number, value: number) {
  if (!isSaintHealAllocatePrompt.value) return
  const current = saintHealAllocations.value[index] || 0
  const otherSum = saintHealAllocations.value.reduce((s, v, i) => s + (i === index ? 0 : (v || 0)), 0)
  const maxAllowed = Math.max(0, SAINT_HEAL_TOTAL - otherSum)
  const clamped = Math.max(0, Math.min(value, Math.min(SAINT_HEAL_TOTAL, maxAllowed)))
  if (clamped === current) return
  const next = saintHealAllocations.value.slice()
  next[index] = clamped
  saintHealAllocations.value = next
}

function setRuneReforgeAllocation(index: number, value: number) {
  if (!isRuneReforgeAllocatePrompt.value) return
  const current = runeReforgeAllocations.value[index] || 0
  const otherSum = runeReforgeAllocations.value.reduce((s, v, i) => s + (i === index ? 0 : (v || 0)), 0)
  const maxAllowed = Math.max(0, RUNE_REFORGE_TOTAL - otherSum)
  const clamped = Math.max(0, Math.min(value, Math.min(RUNE_REFORGE_TOTAL, maxAllowed)))
  if (clamped === current) return
  const next = runeReforgeAllocations.value.slice()
  next[index] = clamped
  runeReforgeAllocations.value = next
}

const canSubmitSaintHeal = computed(() => {
  if (!isSaintHealAllocatePrompt.value) return false
  if (saintHealAllocations.value.length === 0) return false
  // 允许总和 <= 3（不强制等于 3）
  return saintHealAllocations.value.reduce((s, v) => s + (v || 0), 0) <= SAINT_HEAL_TOTAL
})

const canSubmitRuneReforge = computed(() => {
  if (!isRuneReforgeAllocatePrompt.value) return false
  if (runeReforgeAllocations.value.length === 0) return false
  // 符文改造强制总和 = 3
  return runeReforgeAllocations.value.reduce((s, v) => s + (v || 0), 0) === RUNE_REFORGE_TOTAL
})

function submitSaintHealAllocation() {
  if (!canSubmitSaintHeal.value) {
    showPromptError(`治疗分配无效（总和不能超过 ${SAINT_HEAL_TOTAL}）`)
    return
  }
  actions.submitSelect([...saintHealAllocations.value])
}

function submitRuneReforgeAllocation() {
  if (!canSubmitRuneReforge.value) {
    showPromptError(`分配无效（战纹+魔纹之和必须等于 ${RUNE_REFORGE_TOTAL}）`)
    return
  }
  actions.submitSelect([...runeReforgeAllocations.value])
}









function toggleExtractOption(index: number) {
  const idx = selectedExtractIndices.value.indexOf(index)
  if (idx >= 0) {
    selectedExtractIndices.value.splice(idx, 1)
  } else {
    const max = prompt.value?.max ?? 2
    if (selectedExtractIndices.value.length < max) {
      selectedExtractIndices.value.push(index)
      selectedExtractIndices.value.sort((a, b) => a - b)
    }
  }
}

function confirmExtractSelection() {
  const min = prompt.value?.min ?? 1
  const max = prompt.value?.max ?? 2
  const sel = selectedExtractIndices.value
  if (sel.length < min || sel.length > max) return
  actions.submitSelect(sel)
}

function resolveOptionPlayerId(option: { id: string; label: string }): string | null {
  if (isBranchPromptChoice()) return null
  if (playerViews.value[option.id]) return option.id
  const label = String(option.label || '')
  if (!label) return null
  const lowLabel = label.toLowerCase()

  const markersFor = (playerId: string): string[] => {
    const p = playerViews.value[playerId]
    if (!p) return []
    const markers = new Set<string>()
    if (p.id) markers.add(p.id)
    if (p.name) markers.add(p.name)
    if (p.role) {
      markers.add(p.role)
      const roleName = getRoleDisplayName(p.role)
      if (roleName && roleName !== '未知角色') markers.add(roleName)
    }
    return [...markers]
  }

  const matched = Object.values(playerViews.value).filter((p) => {
    const markers = markersFor(p.id)
    return markers.some((marker) => {
      const token = marker.trim().toLowerCase()
      return !!token && lowLabel.includes(token)
    })
  })

  if (matched.length !== 1) return null
  return matched[0]?.id || null
}

const playerOptionEntries = computed(() => {
  if (!prompt.value?.options?.length) return []
  // choose_skill 类型的选项是技能，不是玩家目标
  if (prompt.value.presentation?.kind === 'skill_choice') return []
  return prompt.value.options
    .map((option, index) => {
      const playerId = resolveOptionPlayerId(option)
      if (!playerId) return null
      const player = playerViews.value[playerId]
      if (!player) return null
      return { index, option, player }
    })
    .filter((entry): entry is { index: number; option: PromptOption; player: PlayerView } => entry != null)
})

const playerOptionIndexSet = computed(() => {
  const set = new Set<number>()
  for (const entry of playerOptionEntries.value) {
    set.add(entry.index)
  }
  return set
})

const selectedPromptTargetOptionIndexes = computed(() => {
  const indexByPlayerId = new Map<string, number>()
  for (const entry of playerOptionEntries.value) {
    indexByPlayerId.set(entry.player.id, entry.index)
  }
  const indexes: number[] = []
  for (const targetId of interruptStore.selectedTargets) {
    const index = indexByPlayerId.get(targetId)
    if (index === undefined) return []
    indexes.push(index)
  }
  return indexes
})

const nonPlayerOptions = computed(() => {
  const options = prompt.value?.options ?? []
  return options.filter((_, idx) => !playerOptionIndexSet.value.has(idx))
})

const isSpiritCasterPowerPickPrompt = computed(() => {
  const p = prompt.value?.presentation
  return p?.kind === 'card_picker' && p?.card_source === 'field'
})

const showConfirmButtonSection = computed(() => {
  return (
    isConfirmType.value &&
    !!prompt.value?.options?.length &&
    !needsCardSelection.value &&
    !needsTargetSelection.value &&
    !isSpiritCasterPowerPickPrompt.value
  )
})

const isResponseSkillConfirmPrompt = computed(() => {
  return prompt.value?.presentation?.kind === 'skill_choice'
})

function isPromptActivationCostCancelable(p: NonNullable<typeof prompt.value>): boolean {
  const cancelPolicy = p.presentation?.cancel_policy
  return cancelPolicy === 'abort' || cancelPolicy === 'decline' || cancelPolicy === 'back'
}

const canCancelPrompt = computed(() => {
  if (!prompt.value) return false
  return isPromptActivationCostCancelable(prompt.value)
})

function handleOptionClick(optionId: string) {
  if (optionId === 'counter_disabled') {
    showPromptError('此攻击无法应战')
    return
  }
  if (prompt.value?.presentation?.kind === 'skill_choice') {
    const idx = prompt.value.options.findIndex((o: { id: string }) => o.id === optionId)
    if (idx >= 0) {
      actions.submitSelect([idx])
    } else {
      if (!canCancelPrompt.value) {
        showPromptError('当前步骤不可取消，请先完成本次操作')
        return
      }
      actions.submitCancel()
    }
    return
  }
  const structuredOptionKind = prompt.value?.presentation?.kind
  if (structuredOptionKind === 'branch_select' || structuredOptionKind === 'numeric') {
    const optionIndex = prompt.value?.options?.findIndex((o: { id: string }) => o.id === optionId) ?? -1
    if (optionIndex >= 0) {
      if (prompt.value?.presentation?.has_decline && optionIndex === (prompt.value.presentation.decline_index ?? 0)) {
        actions.submitCancel()
        return
      }
      actions.submitSelect([optionIndex])
      return
    }
    if (optionId === 'skip' || optionId === 'cancel') {
      if (!canCancelPrompt.value) {
        showPromptError('当前步骤不可取消，请先完成本次操作')
        return
      }
      actions.submitCancel()
      return
    }
  }
  if (optionId === 'skip' || optionId === 'cancel') {
    if (!canCancelPrompt.value) {
      showPromptError('当前步骤不可取消，请先完成本次操作')
      return
    }
    actions.submitCancel()
    return
  }
  if (optionId === 'refuse') {
    // 魔爆冲击“不弃牌”是规则内显式选项，直接走 Cancel 语义。
    actions.submitCancel()
    return
  }
  if (optionId === 'confirm') {
    actions.submitConfirm()
    return
  }
  const optionIndex = prompt.value?.options?.findIndex((o: { id: string }) => o.id === optionId) ?? -1
  if (prompt.value?.presentation?.has_decline && optionIndex === (prompt.value.presentation.decline_index ?? 0)) {
    actions.submitCancel()
    return
  }
  // 魔弹融合等确认选项：yes=0, no=1
  if (optionId === 'yes' || optionId === 'no') {
    actions.submitSelect([optionId === 'yes' ? 0 : 1])
    return
  }
  // 魔弹掌控方向选择：normal=0, reverse=1
  if (optionId === 'normal' || optionId === 'reverse') {
    actions.submitSelect([optionId === 'normal' ? 0 : 1])
    return
  }
  const responseKind = promptOptionResponseKind({ id: optionId })
  if (responseKind === 'take') {
    actions.submitRespondTake()
    return
  }
  if (responseKind === 'counter') {
    if (interruptStore.selectedHandIndexes.length === 0) {
      showPromptError(isMagicMissilePrompt.value ? '请先选择一张【魔弹】进行传递' : '请先选择一张攻击牌进行应战')
      return
    }
    if (needsCounterTargetSelection.value && !interruptStore.promptCounterTarget) {
      showPromptError('请先选择反弹目标（攻击方的队友）')
      return
    }
    if (!actions.submitRespondCounter(isMagicMissilePrompt.value)) return
    return
  }
  if (responseKind === 'defend') {
    if (interruptStore.selectedHandIndexes.length === 0) {
      showPromptError('请先选择一张【圣光】进行防御（圣盾需提前放置）')
      return
    }
    if (!actions.submitRespondDefend()) return
    return
  }
  if (isNonHandChooseCardsMultiMode.value && isNonHandChooseCardOption(optionId)) {
    toggleInlineCardOption(optionId)
    return
  }
  {
    const optionIndex = prompt.value?.options?.findIndex((o: { id: string }) => o.id === optionId) ?? -1
    if (optionIndex >= 0) {
      actions.submitSelect([optionIndex])
    } else {
      const index = parseInt(optionId, 10)
      if (!Number.isNaN(index)) {
        actions.submitSelect([index])
      } else {
        actions.submitAction({
          player_id: myPlayerId.value,
          type: 'Select',
          skill_id: optionId
        })
      }
    }
  }
}

function normalizeElementToken(raw: string): string {
  const text = String(raw || '').trim().toLowerCase()
  if (!text) return ''
  if (text.includes('water') || text.includes('水')) return 'Water'
  if (text.includes('fire') || text.includes('火')) return 'Fire'
  if (text.includes('earth') || text.includes('地')) return 'Earth'
  if (text.includes('wind') || text.includes('风')) return 'Wind'
  if (text.includes('thunder') || text.includes('雷')) return 'Thunder'
  if (text.includes('light') || text.includes('光')) return 'Light'
  if (text.includes('dark') || text.includes('暗')) return 'Dark'
  return ''
}

function resolvePlagueDeathTouchElementOptionIndex(): number | null {
  if (!isPlagueDeathTouchElementPrompt.value || !prompt.value) return null
  if (interruptStore.selectedHandIndexes.length <= 0) return null
  const selectedIndex = interruptStore.selectedHandIndexes[0]
  if (selectedIndex == null || selectedIndex < 0 || selectedIndex >= myHand.value.length) return null
  const selectedCard = myHand.value[selectedIndex]
  if (!selectedCard?.element) return null

  const targetElement = selectedCard.element
  for (let i = 0; i < prompt.value.options.length; i++) {
    const option = prompt.value.options[i]
    if (!option) continue
    const optionElement = normalizeElementToken(`${option.label || ''} ${option.button_label || ''}`) || normalizeElementToken(option.id || '')
    if (optionElement && optionElement === targetElement) return i
  }
  return null
}

const canConfirmPrompt = computed(() => {
  if (!prompt.value) return false
  if (isPlagueDeathTouchElementPrompt.value) {
    return resolvePlagueDeathTouchElementOptionIndex() !== null
  }
  if (promptRequiresManualTargetConfirm.value) {
    const tCount = interruptStore.selectedTargets.length
    return (
      tCount >= prompt.value.min &&
      tCount <= prompt.value.max &&
      selectedPromptTargetOptionIndexes.value.length === tCount
    )
  }
  if (promptHasHandCardOptions.value) {
    const cCount = interruptStore.selectedHandIndexes.length
    return cCount >= prompt.value.min && cCount <= prompt.value.max
  }
  if (prompt.value.presentation?.kind === 'target_picker') {
    const tCount = interruptStore.selectedTargets.length
    return tCount >= prompt.value.min && tCount <= prompt.value.max
  }
  if (prompt.value.presentation?.kind === 'card_picker') {
    const cCount = isNonHandChooseCardsMultiMode.value
      ? selectedInlineCardIDs.value.length
      : interruptStore.selectedHandIndexes.length
    return cCount >= prompt.value.min && cCount <= prompt.value.max
  }
  return true
})

function confirmPromptAction() {
  if (!canConfirmPrompt.value) return

  if (isPlagueDeathTouchElementPrompt.value) {
    const optionIndex = resolvePlagueDeathTouchElementOptionIndex()
    if (optionIndex === null) {
      showPromptError('请先在手牌区选择可用于死亡之触的同系牌')
      return
    }
    actions.submitSelect([optionIndex])
    return
  }

  if (promptRequiresManualTargetConfirm.value) {
    const targetOptionIndexes = selectedPromptTargetOptionIndexes.value
    if (targetOptionIndexes.length !== interruptStore.selectedTargets.length) {
      showPromptError('请选择有效的治疗目标')
      return
    }
    actions.submitSelect(targetOptionIndexes)
    return
  }

  if (prompt.value?.presentation?.kind === 'target_picker' && interruptStore.selectedTargets.length > 0) {
    if (interruptStore.selectedTargets.length === 1) {
      const targetId = interruptStore.selectedTargets[0]
      if (!targetId) return
      actions.submitPromptTarget(targetId)
    } else {
      actions.submitAction({
        player_id: myPlayerId.value,
        type: 'Select',
        target_ids: interruptStore.selectedTargets
      })
    }
    return
  }

  if (isNonHandChooseCardsMultiMode.value) {
    if (selectedInlineCardIDs.value.length > 0 || prompt.value?.min === 0) {
      actions.submitSelectCardIDs(selectedInlineCardIDs.value)
    }
    return
  }

  const indices = interruptStore.selectedHandIndexes
  if (prompt.value?.presentation?.kind === 'card_picker' && indices.length === 0 && prompt.value.min === 0) {
    actions.submitSelectCardIDs([])
    return
  }
  if (indices.length > 0) {
    const cardIDs = selectedPromptHandCardIDs(indices)
    if (cardIDs.length === indices.length) {
      actions.submitSelectCardIDs(cardIDs)
      return
    }
    showPromptError('当前卡牌选择缺少 card_id，请刷新后重试')
  }
}

function parseCocoonFieldIndexFromOptionLabel(label: string): number | null {
  const matched = String(label || '').match(/茧\[(\d+)\]/)
  if (!matched) return null
  const parsed = Number.parseInt(matched[1] || '', 10)
  if (!Number.isFinite(parsed) || parsed < 0) return null
  return parsed
}

function parseCocoonFieldIndexFromOption(option: { label?: string; field_index?: number | string | null }): number | null {
  const rawFieldIndex = option.field_index
  if (rawFieldIndex !== undefined && rawFieldIndex !== null && rawFieldIndex !== '') {
    const parsed = Number.parseInt(String(rawFieldIndex), 10)
    if (Number.isFinite(parsed) && parsed >= 0) return parsed
  }
  return parseCocoonFieldIndexFromOptionLabel(String(option.label || ''))
}

function isIndexedCocoonOption(option: { label?: string; field_index?: number | string | null }): boolean {
  return parseCocoonFieldIndexFromOption(option) !== null
}

function optionCardID(option: { card_id?: string | null }): string {
  return String(option.card_id || '').trim()
}

function handIndexForPromptOption(option: { card_id?: string | null }): number | null {
  const cardID = optionCardID(option)
  if (!cardID) return null
  const idx = myHand.value.findIndex(card => String(card.id || '').trim() === cardID)
  return idx >= 0 ? idx : null
}

function promptOptionForHandIndex(handIndex: number): PromptOption | null {
  if (!prompt.value?.options?.length) return null
  const handCardID = String(myHand.value[handIndex]?.id || '').trim()
  if (!handCardID) return null
  for (const option of prompt.value.options) {
    const cardID = optionCardID(option)
    if (cardID && cardID === handCardID) return option
  }
  return null
}

function selectedPromptHandCardIDs(indices: number[]): string[] {
  if (!prompt.value || prompt.value.presentation?.kind !== 'card_picker' || prompt.value.presentation?.card_source !== 'hand') return []
  const ids: string[] = []
  for (const idx of indices) {
    const option = promptOptionForHandIndex(idx)
    const cardID = option ? optionCardID(option) : ''
    if (!cardID) return []
    ids.push(cardID)
  }
  return ids
}

function isPromptHandCardOption(option: { id: string; label: string; card_id?: string | null }): boolean {
  const p = prompt.value?.presentation
  if (!p) return false
  if (p.kind === 'numeric') return false
  if (p.kind === 'target_picker') return false
  if (p.kind === 'branch_select') return false
  if (p.kind === 'skill_choice') return false
  // card_picker with card_source=field or proxy: not hand cards
  if (p.kind === 'card_picker' && p.card_source !== 'hand') return false
  // card_picker with card_source=hand: these are hand cards
  if (p.kind === 'card_picker' && p.card_source === 'hand') {
    return handIndexForPromptOption(option) !== null
  }
  return false
}

const promptCardOptionIndexSet = computed(() => {
  const set = new Set<number>()
  if (!prompt.value?.options?.length) return set
  for (const option of prompt.value.options) {
    if (!isPromptHandCardOption(option)) continue
    const idx = handIndexForPromptOption(option)
    if (idx !== null) set.add(idx)
  }
  return set
})

const promptHasHandCardOptions = computed(() => promptCardOptionIndexSet.value.size > 0)

const hasIndexedCocoonOptions = computed(() => {
  if (!prompt.value?.options?.length) return false
  return prompt.value.options.some((option) => isIndexedCocoonOption(option))
})

const isNonHandChooseCardsMultiMode = computed(() => {
  if (!prompt.value) return false
  if (prompt.value.presentation?.kind !== 'card_picker') return false
  if (hasCounterOrDefend.value) return false
  if (promptCardOptionIndexSet.value.size > 0) return false
  if (!prompt.value.options?.length) return false
  if ((prompt.value.max ?? 1) <= 1) return false
  return prompt.value.options.every((option) => !!optionCardID(option))
})

function isNonHandChooseCardOption(optionId: string): boolean {
  if (!isNonHandChooseCardsMultiMode.value || !prompt.value?.options?.length) return false
  return prompt.value.options.some((option) => option.id === optionId && !!optionCardID(option))
}

function inlineCardIDForOption(optionId: string): string {
  if (!prompt.value?.options?.length) return ''
  const option = prompt.value.options.find((candidate) => candidate.id === optionId)
  return option ? optionCardID(option) : ''
}

function toggleInlineCardOption(optionId: string) {
  if (!isNonHandChooseCardOption(optionId)) return
  const cardID = inlineCardIDForOption(optionId)
  if (!cardID) return
  const pos = selectedInlineCardIDs.value.indexOf(cardID)
  if (pos >= 0) {
    selectedInlineCardIDs.value.splice(pos, 1)
    return
  }
  const max = prompt.value?.max ?? 1
  if (selectedInlineCardIDs.value.length >= max) return
  selectedInlineCardIDs.value.push(cardID)
}

function isInlineCardOptionSelected(optionId: string): boolean {
  if (!isNonHandChooseCardOption(optionId)) return false
  const cardID = inlineCardIDForOption(optionId)
  return !!cardID && selectedInlineCardIDs.value.includes(cardID)
}

type RawDockOption = {
  id: string
  label: string
  button_label: string
  hint?: string
  field_index?: number
  disabled?: boolean
  optionIndex?: number
}

type DockButtonOption = {
  id: string
  label: string
  buttonLabel: string
  hint: string
  disabled?: boolean
  numeric: boolean
  optionIndex?: number
}

type SkillPromptEntry = {
  id: string
  promptText: string
  buttonLabel: string
  disabled: boolean
}

type SkillPromptButton = {
  id: string
  label: string
  disabled: boolean
  cancel: boolean
}

type FraudElementCardOption = {
  id: string
  title: string
  glyph: string
  tone: string
}

const PROMPT_IMAGE_BUTTON_CANDIDATES: Record<PromptImageButtonKind, string[]> = {
  take: ['/assets/ui/prompt_btn_take.png'],
  counter: ['/assets/ui/prompt_btn_counter.png'],
  defend: ['/assets/ui/prompt_btn_defend.png'],
  cancel: ['/assets/ui/action_cancel_btn.png'],
  confirm: ['/assets/ui/action_confirm.png'],
  card: ['/assets/ui/action_card.png'],
  action: ['/assets/ui/action_special_btn.png'],
}

const promptImageButtonIndex = ref<Record<PromptImageButtonKind, number>>({
  take: 0,
  counter: 0,
  defend: 0,
  cancel: 0,
  confirm: 0,
  card: 0,
  action: 0,
})

const promptImageButtonFailed = ref<Record<PromptImageButtonKind, boolean>>({
  take: false,
  counter: false,
  defend: false,
  cancel: false,
  confirm: false,
  card: false,
  action: false,
})

function promptImageButtonAsset(kind: PromptImageButtonKind): string {
  const candidates = PROMPT_IMAGE_BUTTON_CANDIDATES[kind]
  const index = promptImageButtonIndex.value[kind]
  return candidates[Math.min(index, candidates.length - 1)] || ''
}

function isPromptImageButtonReady(kind: PromptImageButtonKind): boolean {
  return !promptImageButtonFailed.value[kind]
}

function onPromptImageButtonError(kind: PromptImageButtonKind) {
  const candidates = PROMPT_IMAGE_BUTTON_CANDIDATES[kind]
  const nextIndex = promptImageButtonIndex.value[kind] + 1
  if (nextIndex < candidates.length) {
    promptImageButtonIndex.value[kind] = nextIndex
    return
  }
  promptImageButtonFailed.value[kind] = true
}

function promptImageButtonFallbackText(kind: PromptImageButtonKind | null, buttonLabel: string = ''): string {
  if (kind === 'take') return '命'
  if (kind === 'defend') return '防'
  if (kind === 'counter') return '应'
  if (kind === 'cancel') return '消'
  if (kind === 'confirm') return '确'
  if (kind === 'card') return '牌'
  if (kind === 'action') return buttonLabel ? buttonLabel.charAt(0) : '动'
  return ''
}

function dockButtonImageKind(option: DockButtonOption): PromptImageButtonKind | null {
  if (option.numeric) return null
  const responseKind = promptOptionResponseKind({ id: option.id })
  if (responseKind) return responseKind
  const presentation = prompt.value?.presentation
  return promptImageButtonKindByOption({
    id: option.id,
    presentationKind: presentation?.kind,
    cancelPolicy: presentation?.cancel_policy,
    hasDecline: presentation?.has_decline,
    declineIndex: presentation?.decline_index,
    optionIndex: option.optionIndex,
  })
}

function isDockButtonImageStyle(option: DockButtonOption): boolean {
  return dockButtonImageKind(option) !== null
}

function dockButtonImageSrc(option: DockButtonOption): string {
  const kind = dockButtonImageKind(option)
  return kind ? promptImageButtonAsset(kind) : ''
}

function isDockButtonImageReady(option: DockButtonOption): boolean {
  const kind = dockButtonImageKind(option)
  if (!kind) return false
  return isPromptImageButtonReady(kind)
}

function onDockButtonImageError(option: DockButtonOption) {
  const kind = dockButtonImageKind(option)
  if (!kind) return
  onPromptImageButtonError(kind)
}

function dockButtonFallbackText(option: DockButtonOption): string {
  return promptImageButtonFallbackText(dockButtonImageKind(option), option.buttonLabel)
}

function isPromptConfirmImageReady(): boolean {
  return isPromptImageButtonReady('confirm')
}

function promptConfirmImageSrc(): string {
  return promptImageButtonAsset('confirm')
}

function onPromptConfirmImageError() {
  onPromptImageButtonError('confirm')
}

function skillButtonImageSrc(option: SkillPromptButton): string {
  return option.cancel ? promptImageButtonAsset('cancel') : promptConfirmImageSrc()
}

function isSkillButtonImageReady(option: SkillPromptButton): boolean {
  return option.cancel ? isPromptImageButtonReady('cancel') : isPromptConfirmImageReady()
}

function onSkillButtonImageError(option: SkillPromptButton) {
  if (option.cancel) {
    onPromptImageButtonError('cancel')
    return
  }
  onPromptConfirmImageError()
}

function skillButtonFallbackText(option: SkillPromptButton): string {
  return option.cancel ? '消' : '确'
}

function isBranchPromptChoice(): boolean {
  return prompt.value?.presentation?.kind === 'branch_select'
}

function isSystemBranchPromptChoice(): boolean {
  return isBranchPromptChoice()
}

function isFieldCoverSelectionPrompt(): boolean {
  const p = prompt.value?.presentation
  return p?.kind === 'card_picker' && p?.card_source === 'field'
}

function isCocoonFieldSelectionPrompt(): boolean {
  return isFieldCoverSelectionPrompt()
}

function overlayDecisionOptionTitle(option: DockButtonOption): string {
  const label = String(option.label || '').trim()
  const buttonLabel = String(option.buttonLabel || '').trim()
  if (!label) return buttonLabel
  if (!buttonLabel || buttonLabel === label) return label
  // 分支选项 id 常为 0/1/2，buttonLabel 可能被落成 1/2/3，标题应展示完整说明。
  if (/^\d+$/.test(buttonLabel)) return label
  return buttonLabel
}

function normalizeDockOption(option: RawDockOption): DockButtonOption {
  const id = String(option.id || '').trim()
  const label = String(option.label || '').trim()
  const buttonLabel = String(option.button_label || '').trim()
  let hint = String(option.hint || '').trim()

  if (promptOptionResponseKind({ id }) !== null) {
    hint = ''
  }

  if (!hint && label && label !== buttonLabel) {
    hint = label
  }

  return {
    id,
    label,
    buttonLabel,
    hint,
    disabled: option.disabled,
    numeric: /^\d+$/.test(buttonLabel),
    optionIndex: option.optionIndex,
  }
}

function buildDockButtons(options: RawDockOption[]): DockButtonOption[] {
  if (options.length === 0) return []
  return options.map((option) => normalizeDockOption(option))
}

const fraudElementCardMetaById: Record<string, Omit<FraudElementCardOption, 'id' | 'title'>> = {
  water: { glyph: '水', tone: 'prompt-fraud-card--water' },
  fire: { glyph: '火', tone: 'prompt-fraud-card--fire' },
  earth: { glyph: '地', tone: 'prompt-fraud-card--earth' },
  wind: { glyph: '风', tone: 'prompt-fraud-card--wind' },
  thunder: { glyph: '雷', tone: 'prompt-fraud-card--thunder' },
}

function fraudAttackCardName(optionId: string, fallback: string): string {
  const lower = String(optionId || '').trim().toLowerCase()
  if (lower === 'water') return '水涟斩'
  if (lower === 'fire') return '火焰斩'
  if (lower === 'earth') return '地裂斩'
  if (lower === 'wind') return '风神斩'
  if (lower === 'thunder') return '雷光斩'
  const normalizedFallback = String(fallback || '').trim()
  if (normalizedFallback) return normalizedFallback
  return String(optionId || '').trim() || '系别'
}

const isFraudElementCardPickerPrompt = computed(() =>
  prompt.value?.presentation?.kind === 'branch_select' && prompt.value?.presentation?.layout === 'fraud_attack_element'
)

const fraudElementCardOptions = computed<FraudElementCardOption[]>(() => {
  if (!isFraudElementCardPickerPrompt.value || !prompt.value?.options?.length) return []
  return prompt.value.options.map((option) => {
    const lower = String(option.id || '').trim().toLowerCase()
    const meta = fraudElementCardMetaById[lower] ?? {
      glyph: '系',
      tone: 'prompt-fraud-card--generic',
    }
    return {
      id: option.id,
      title: fraudAttackCardName(option.id, option.label),
      glyph: meta.glyph,
      tone: meta.tone,
    }
  })
})

const cardFooterOptions = computed<RawDockOption[]>(() => {
  if (!prompt.value?.options || !needsCardSelection.value) return []
  if (hasCounterOrDefend.value) {
    const responseOrder: Record<Exclude<ResponseActionKind, null>, number> = {
      take: 0,
      defend: 1,
      counter: 2,
    }
    const responseRank = (kind: ResponseActionKind): number => {
      if (!kind) return 99
      return responseOrder[kind]
    }
    return prompt.value.options
      .filter((option) => promptOptionResponseKind(option) !== null)
      .sort((a, b) => {
        const rankA = responseRank(promptOptionResponseKind(a))
        const rankB = responseRank(promptOptionResponseKind(b))
        return rankA - rankB
      })
      .map((option) => ({
        id: option.id,
        label: option.label,
        button_label: option.button_label,
        hint: option.hint,
        optionIndex: prompt.value?.options.indexOf(option),
        disabled: false
      }))
  }
  if (prompt.value.presentation?.kind !== 'card_picker') return []
  // 选牌类提示统一在手牌区完成，行动区不再展示“选哪张牌”的按钮。
  return []
})

const promptNeedsHandCardConfirm = computed(() => {
  if (!prompt.value || !needsCardSelection.value || hasCounterOrDefend.value) return false
  if (isElfElementalShotPickPrompt.value) return true
  if (isPlagueDeathTouchElementPrompt.value) return true
  if (isNonHandChooseCardsMultiMode.value) return false
  if (promptCardOptionIndexSet.value.size > 0) return true
  return !prompt.value.options?.length
})

const promptNeedsInlineCardOptionConfirm = computed(() =>
  isNonHandChooseCardsMultiMode.value && !hasIndexedCocoonOptions.value
)

const promptNeedsCardConfirm = computed(() =>
  promptNeedsHandCardConfirm.value || promptNeedsInlineCardOptionConfirm.value
)

const cardConfirmHintText = computed(() => {
  if (isElfElementalShotPickPrompt.value) return '请从手牌区或扩展区选择法术牌/祝福牌并点击发动'
  if (isPlagueDeathTouchElementPrompt.value) return '请选择同系手牌并点击确认'
  if (prompt.value?.presentation?.card_filter === 'same_element_combo') return '请选择2~3张同系牌，3张将自动转为暗灭攻击'
  if (prompt.value?.presentation?.card_filter === 'same_element') return '请选择要弃置的同系手牌并点击确认'
  if (promptNeedsInlineCardOptionConfirm.value) return '完成选择后点击发动'
  return '完成选牌后点击发动'
})

const cardConfirmPromptMessage = computed(() => {
  const message = String(prompt.value?.message || '').trim()
  if (message) return message
  return cardConfirmHintText.value
})

const showCardConfirmCancelRow = computed(() =>
  promptNeedsCardConfirm.value && canCancelPrompt.value && !isSkillChoicePrompt.value
)

const targetSelectionPromptMessage = computed(() => {
  if (!prompt.value) return ''
  const message = String(prompt.value.message || '').trim()
  if (message) return message
  if (needsCounterTargetSelection.value) return '请选择反弹目标角色'
  if (needsTargetSelection.value) return '请选择目标角色'
  return ''
})

const showTargetSelectionHintRow = computed(() =>
  !isSkillChoicePrompt.value &&
  !promptNeedsCardConfirm.value &&
  inlinePrimaryButtons.value.length === 0 &&
  (needsTargetSelection.value || needsCounterTargetSelection.value)
)

const singleActivationCostConfirmOption = computed<DockButtonOption | null>(() => {
  if (!prompt.value) return null
  if (promptNeedsCardConfirm.value || isSkillChoicePrompt.value) return null
  if (!canCancelPrompt.value) return null
  if (inlinePrimaryButtons.value.length !== 1) return null
  const option = inlinePrimaryButtons.value[0]
  if (!option || option.numeric || option.disabled) return null
  if (promptOptionResponseKind({ id: option.id }) !== null) return null
  if (prompt.value.presentation?.layout !== 'activation_cost') return null
  return option
})

const singleActivationCostConfirmHintText = computed(() => {
  if (!singleActivationCostConfirmOption.value) return ''
  const hint = String(singleActivationCostConfirmOption.value.hint || '').trim()
  if (hint) return hint
  const label = String(singleActivationCostConfirmOption.value.label || '').trim()
  if (label) return label
  return String(prompt.value?.message || '').trim()
})

const inlinePrimaryPromptMessage = computed(() => {
  if (!showConfirmButtonSection.value) return ''
  return String(prompt.value?.message || '').trim()
})

const inlinePrimaryButtons = computed<DockButtonOption[]>(() => {
  if (isExtractPrompt.value) return []
  if (isFraudElementCardPickerPrompt.value) return []
  if (isSaintHealAllocatePrompt.value) return []
  if (isRuneReforgeAllocatePrompt.value) return []
  if (needsCardSelection.value) return buildDockButtons(cardFooterOptions.value)
  // 响应技能选择（choose_skill）：直接从 prompt.options 构建按钮
  if (prompt.value?.presentation?.kind === 'skill_choice') {
    const options = (prompt.value.options || [])
      .map((option) => ({
        id: option.id,
        label: option.label,
        button_label: option.button_label,
        hint: option.hint,
        optionIndex: prompt.value?.options.indexOf(option),
        disabled: false
      }))
    return buildDockButtons(options)
  }
  if (showConfirmButtonSection.value) {
    const shouldExposeIndexedCocoonOptions =
      isSystemBranchPromptChoice()
    const shouldHideIndexedCocoonOptions = isCocoonFieldSelectionPrompt()
    const optionSource = shouldExposeIndexedCocoonOptions
      ? (prompt.value?.options || [])
      : nonPlayerOptions.value
    const declineIndex = prompt.value?.presentation?.has_decline ? (prompt.value.presentation.decline_index ?? 0) : -1
    const options = optionSource
      .filter((option, index) => {
        if (declineIndex >= 0 && index === declineIndex) return false
        if (shouldExposeIndexedCocoonOptions && (option.id === 'decline' || option.id === '-1')) return true
        if (prompt.value?.presentation?.kind === 'branch_select' || prompt.value?.presentation?.kind === 'numeric') return true
        return option.id !== 'cancel' && option.id !== 'skip'
      })
      .filter((option) => !shouldHideIndexedCocoonOptions || !isIndexedCocoonOption(option))
      .filter((option) => shouldExposeIndexedCocoonOptions || !isIndexedCocoonOption(option))
      .map((option) => ({
        id: option.id,
        label: option.label,
        button_label: option.button_label,
        hint: option.hint,
        field_index: option.field_index,
        optionIndex: prompt.value?.options.indexOf(option),
        disabled: false
      }))
    return buildDockButtons(options)
  }
  return []
})

const isSkillChoicePrompt = computed(() => {
  if (!prompt.value) return false
  return prompt.value.presentation?.kind === 'skill_choice' || isResponseSkillConfirmPrompt.value
})

function parseSkillTitle(option: DockButtonOption, index: number): string {
  const rawLabel = String(option.label || '').trim()
  let title = rawLabel || `技能 ${index + 1}`

  // 兼容旧服务端：若 label 仍为”标题：说明”，前端兜底只截标题。
  const separatorIndex = rawLabel.indexOf('：')
  if (separatorIndex > 0) {
    const parsedTitle = rawLabel.slice(0, separatorIndex).trim()
    if (parsedTitle) title = parsedTitle
  }

  // 去掉前缀序号与尾部消耗标记，按钮中尽量只保留技能名。
  title = title.replace(/^\d+\s*[.)、]\s*/, '').trim()
  title = title.replace(/\s*\[[^\]]+\]\s*$/, '').trim()

  // 若 title 仍为英文 ID 格式（如 wind_fury），尝试从角色数据中查找真实名称
  if (/^[a-z][a-z0-9_]*$/.test(title)) {
    const resolved = resolveSkillTitleById(title)
    if (resolved) title = resolved
  }

  return title
}

const skillTitleMap = computed(() => {
  const map = new Map<string, string>()
  const chars = snapshotStore.characters
  for (const charId in chars) {
    const skills = chars[charId]?.skills
    if (!skills) continue
    for (const sk of skills) {
      if (sk.id && sk.title) map.set(sk.id, sk.title)
    }
  }
  return map
})

function resolveSkillTitleById(skillId: string): string | null {
  return skillTitleMap.value.get(skillId) ?? null
}

const skillPromptEntries = computed<SkillPromptEntry[]>(() => {
  if (!isSkillChoicePrompt.value || inlinePrimaryButtons.value.length === 0) return []
  return inlinePrimaryButtons.value.map((option, index) => {
    const title = parseSkillTitle(option, index)
    return {
      id: option.id,
      promptText: `是否发动【${title}】`,
      buttonLabel: option.buttonLabel,
      disabled: !!option.disabled
    }
  })
})

const skillPromptTitle = computed(() => {
  if (!isSkillChoicePrompt.value || skillPromptEntries.value.length === 0) return ''
  if (skillPromptEntries.value.length === 1) return skillPromptEntries.value[0]?.promptText || ''
  const message = String(prompt.value?.message || '').trim()
  return message || '请选择要发动的技能'
})

const skillPromptButtons = computed<SkillPromptButton[]>(() => {
  if (!isSkillChoicePrompt.value || skillPromptEntries.value.length === 0) return []
  const skillCount = skillPromptEntries.value.length
  const buttons: SkillPromptButton[] = skillPromptEntries.value.map((entry, index) => {
    let label = entry.buttonLabel
    if (prompt.value?.presentation?.kind === 'skill_choice' && skillCount > 1) {
      const option = inlinePrimaryButtons.value[index]
      label = option ? parseSkillTitle(option, index) : `技能 ${index + 1}`
    } else if (skillCount > 1 && (label === '发动' || label === '确认')) {
      label = String(index + 1)
    }
    return {
      id: entry.id,
      label,
      disabled: !!entry.disabled,
      cancel: false
    }
  })

  const hasCancelLike = buttons.some((btn) => btn.id === 'cancel' || btn.id === 'skip' || btn.cancel)
  if (canCancelPrompt.value && !hasCancelLike) {
    buttons.push({
      id: 'cancel',
      label: '取消',
      disabled: false,
      cancel: true
    })
  }
  return buttons
})

const isMultiSkillNameChoiceMode = computed(() =>
  prompt.value?.presentation?.kind === 'skill_choice' &&
  skillBranchOptions.value.length > 0 &&
  skillPromptButtons.value.length > 1
)

interface SkillBranchOption {
  id: string
  title: string
  description?: string
  cost?: string
  disabled: boolean
}

const skillBranchOptions = computed<SkillBranchOption[]>(() => {
  if (prompt.value?.presentation?.kind !== 'skill_choice') return []
  return inlinePrimaryButtons.value
    .filter((opt) => opt.id !== 'skip' && opt.id !== 'cancel')
    .map((option, index) => {
      const rawLabel = String(option.label || '').trim()
      const title = parseSkillTitle(option, index)
      const costMatch = rawLabel.match(/\[[^\]]+\]/)
      return {
        id: option.id,
        title,
        description: option.hint || undefined,
        cost: costMatch ? costMatch[0] : undefined,
        disabled: !!option.disabled,
      }
    })
})

const inlinePrimaryGridClass = computed(() => {
  const count = inlinePrimaryButtons.value.length
  if (count <= 1) return 'prompt-inline-grid--1'
  if (count === 2) return 'prompt-inline-grid--2'
  if (count === 3) return 'prompt-inline-grid--3'
  return 'prompt-inline-grid--4'
})

function buildPromptAutoResolveKey(p: NonNullable<typeof prompt.value>): string {
  const options = (p.options || [])
    .map((option) => `${option.id}|${option.label}|${option.button_label || ''}|${option.hint || ''}`)
    .join('||')
  return `${p.type}::${p.player_id}::${p.message}::${options}`
}

const autoResolveOptionId = computed(() => {
  if (!isVisible.value || !prompt.value) return ''
  if (isExtractPrompt.value || isSkillChoicePrompt.value) return ''
  if (hasIndexedCocoonOptions.value) return ''
  // 存在角色目标选项时禁止自动确认，避免“完成选择”类按钮被误触发。
  if (playerOptionEntries.value.length > 0) return ''
  if (needsCardSelection.value || needsTargetSelection.value || needsCounterTargetSelection.value) return ''
  // 有取消/跳过时表示存在真实分支，不做自动确认。
  if (canCancelPrompt.value) return ''
  const candidates = inlinePrimaryButtons.value.filter((option) => !option.disabled)
  if (candidates.length !== 1) return ''
  const onlyOption = candidates[0]
  if (!onlyOption) return ''
  return onlyOption.id
})

// 是否显示通用决策弹窗（X值选择、分支选择、模式选择、发动确认）
// 排除条件：每种有专属 UI 的 prompt 类型都不走通用弹窗。
const showDecisionOverlay = computed(() => {
  if (!isVisible.value || !prompt.value) return false
  // 专属 UI 模式排除
  if (prompt.value.presentation?.kind === 'action_hub') return false
  if (isCocoonFieldSelectionPrompt()) return false      // → 扩展区点击茧，弹框只保留内联取消
  if (isSkillChoicePrompt.value) return false        // → skill-branch-overlay
  if (isFraudElementCardPickerPrompt.value) return false  // → 欺诈选牌弹窗
  if (isExtractPrompt.value) return false              // → 提取选择网格
  if (isSaintHealAllocatePrompt.value) return false    // → 圣疗分配专属弹窗
  if (isRuneReforgeAllocatePrompt.value) return false  // → 符文改造分配专属弹窗
  if (needsCardSelection.value) return false           // → 卡牌选择流程（命中/防御/应战按钮留内联）
  // 简单确认弹框（是/否两个按钮）→ 也走 overlay 决策弹框，便于统一展示
  // 注意：已删除旧的 return false 逻辑，让 是/否 弹框也进入 overlay
  // 符合通用弹窗条件
  if (singleActivationCostConfirmOption.value) return true
  if (inlinePrimaryButtons.value.length > 0 && showConfirmButtonSection.value) return true
  return false
})

const isYesNoDecision = computed(() => {
  if (!prompt.value) return false
  if (singleActivationCostConfirmOption.value) return false
  if (inlinePrimaryButtons.value.some(opt => opt.numeric)) return false
  const options = prompt.value.options || []
  if (options.length !== 2) return false
  // 检测原始 label 是否为简短的是/否式文本
  const labels = options.map(o => String(o.label || '').trim())
  return labels.every(l => l.length > 0 && l.length <= 3)
})

const decisionOverlayMode = computed<'numeric' | 'text' | 'activation-cost' | 'yes-no'>(() => {
  if (singleActivationCostConfirmOption.value) return 'activation-cost'
  if (isSystemBranchPromptChoice()) return 'text'
  if (inlinePrimaryButtons.value.some(opt => opt.numeric)) return 'numeric'
  if (isYesNoDecision.value) return 'yes-no'
  return 'text'
})

const decisionOverlayTitle = computed(() => {
  if (singleActivationCostConfirmOption.value) {
    return String(prompt.value?.message || '确认发动').trim()
  }
  return inlinePrimaryPromptMessage.value || '请选择'
})

const hasAnyInlineButton = computed(() => {
  if (!isVisible.value) return false
  if (prompt.value?.presentation?.kind === 'action_hub') return false
  if (isFraudElementCardPickerPrompt.value) return false
  if (prompt.value?.presentation?.kind === 'skill_choice' && isMultiSkillNameChoiceMode.value) return false
  if (showDecisionOverlay.value) return false
  if (isExtractPrompt.value && !!prompt.value?.options?.length) return true
  if (showTargetSelectionHintRow.value) return true
  if (inlinePrimaryButtons.value.length > 0) return true
  if (promptNeedsCardConfirm.value) return true
  if (canCancelPrompt.value) return true
  return false
})

const cancelDockButton = computed<DockButtonOption>(() => {
  const promptOptions = prompt.value?.options ?? []
  const declineIndex = prompt.value?.presentation?.has_decline ? (prompt.value.presentation.decline_index ?? 0) : -1
  const declineOption = declineIndex >= 0 ? promptOptions[declineIndex] : undefined
  const cancelLabel = prompt.value?.presentation?.cancel_label || ''
  const option = declineOption ?? {
    id: 'cancel',
    label: canCancelPrompt.value ? cancelLabel : '',
    button_label: canCancelPrompt.value ? cancelLabel : ''
  }
  return normalizeDockOption(
    {
      id: option.id,
      label: option.label,
      button_label: option.button_label,
      hint: option.hint,
      optionIndex: declineOption ? declineIndex : undefined,
    }
  )
})

function getDockButtonClass(optionId: string): string {
  const lowerOptionId = String(optionId || '').trim().toLowerCase()
  const kind = promptOptionResponseKind({ id: lowerOptionId })
  if (kind === 'take') return 'prompt-inline-btn--take'
  if (kind === 'counter') return 'prompt-inline-btn--counter'
  if (kind === 'defend') return 'prompt-inline-btn--defend'
  if (lowerOptionId === 'confirm' || lowerOptionId === 'yes') return 'prompt-inline-btn--success'
  if (lowerOptionId === 'skip' || lowerOptionId === 'cancel' || lowerOptionId === 'refuse' || lowerOptionId === 'no' || lowerOptionId === 'pass' || lowerOptionId === 'cannot_act') {
    return 'prompt-inline-btn--cancel'
  }
  return 'prompt-inline-btn--normal'
}

function shouldHideOptionHint(option: DockButtonOption): boolean {
  return promptOptionResponseKind({ id: option.id }) !== null
}

function shouldEnlargeResponseActionButton(option: DockButtonOption): boolean {
  if (!isDarkAttackResponsePrompt.value) return false
  const kind = promptOptionResponseKind({ id: option.id })
  return kind === 'take' || kind === 'defend'
}

watch(autoResolveOptionId, (optionId) => {
  if (!optionId || !prompt.value) return
  const key = buildPromptAutoResolveKey(prompt.value)
  if (autoResolvedPromptKey.value === key) return
  autoResolvedPromptKey.value = key
  handleOptionClick(optionId)
})
</script>

<template>
  <Transition name="prompt-inline-pop">
    <div v-if="hasAnyInlineButton" class="prompt-inline-root" data-testid="prompt-dialog">
      <div class="prompt-inline-surface">
        <template v-if="isExtractPrompt && prompt?.options?.length">
          <div class="prompt-inline-grid prompt-inline-grid--2">
            <button
              v-for="(option, idx) in prompt.options"
              :key="option.id"
              class="prompt-inline-btn prompt-inline-btn--extract"
              :class="{ 'prompt-inline-btn--selected': selectedExtractIndices.includes(idx) }"
              @click="toggleExtractOption(idx)"
            >
              {{ option.label === '红宝石' ? '♦ 红宝石' : '🔷 蓝水晶' }}
            </button>
          </div>
          <div class="flex justify-center mt-2">
            <button
              class="prompt-inline-btn prompt-inline-btn--success action-image-btn"
              :class="{ 'prompt-inline-btn--disabled': selectedExtractIndices.length < (prompt?.min ?? 1) || selectedExtractIndices.length > (prompt?.max ?? 2) }"
              :disabled="selectedExtractIndices.length < (prompt?.min ?? 1) || selectedExtractIndices.length > (prompt?.max ?? 2)"
              @click="confirmExtractSelection"
              :title="`确认提炼（${selectedExtractIndices.length}/${prompt?.max ?? 2}）`"
              :aria-label="`确认提炼（${selectedExtractIndices.length}/${prompt?.max ?? 2}）`"
            >
              <img
                v-if="isPromptConfirmImageReady()"
                class="action-image-btn-fill"
                :src="promptConfirmImageSrc()"
                alt=""
                @error="onPromptConfirmImageError"
              />
              <span v-else class="action-image-fallback-text">确</span>
            </button>
          </div>
        </template>

        <template v-else>
          <div v-if="isSkillChoicePrompt && skillPromptButtons.length > 0 && !isMultiSkillNameChoiceMode" class="prompt-skill-list">
            <div class="prompt-skill-row">
              <div class="prompt-skill-text" :title="skillPromptTitle">{{ skillPromptTitle }}</div>
              <div class="prompt-skill-actions">
                <button
                  v-for="option in skillPromptButtons"
                  :key="option.id"
                  class="prompt-inline-btn prompt-skill-action"
                  :class="[
                    isMultiSkillNameChoiceMode ? 'prompt-inline-btn--normal prompt-skill-action--plain' : '',
                    !isMultiSkillNameChoiceMode ? 'action-image-btn' : '',
                    !isMultiSkillNameChoiceMode ? (option.cancel ? 'prompt-inline-btn--cancel' : 'prompt-inline-btn--success') : '',
                    option.disabled ? 'prompt-inline-btn--disabled' : ''
                  ]"
                  :disabled="option.disabled"
                  :data-testid="`prompt-option-${option.id}`"
                  :title="!isMultiSkillNameChoiceMode ? option.label : undefined"
                  :aria-label="!isMultiSkillNameChoiceMode ? option.label : undefined"
                  @click="handleOptionClick(option.id)"
                >
                  <template v-if="!isMultiSkillNameChoiceMode">
                    <img
                      v-if="isSkillButtonImageReady(option)"
                      class="action-image-btn-fill"
                      :src="skillButtonImageSrc(option)"
                      alt=""
                      @error="onSkillButtonImageError(option)"
                    />
                    <span v-else class="action-image-fallback-text">{{ skillButtonFallbackText(option) }}</span>
                  </template>
                  <template v-else>
                    {{ option.label }}
                  </template>
                </button>
              </div>
            </div>
          </div>

          <div v-else-if="showTargetSelectionHintRow" class="prompt-inline-entry">
            <div class="prompt-inline-hint">{{ targetSelectionPromptMessage }}</div>
            <button
              v-if="promptRequiresManualTargetConfirm"
              class="prompt-inline-btn prompt-inline-btn--success action-image-btn"
              :class="{ 'prompt-inline-btn--disabled': !canConfirmPrompt }"
              :disabled="!canConfirmPrompt"
              data-testid="prompt-confirm-btn"
              title="确认"
              aria-label="确认"
              @click="confirmPromptAction"
            >
              <img
                v-if="isPromptConfirmImageReady()"
                class="action-image-btn-fill"
                :src="promptConfirmImageSrc()"
                alt=""
                @error="onPromptConfirmImageError"
              />
              <span v-else class="action-image-fallback-text">确</span>
            </button>
          </div>

          <div v-else-if="inlinePrimaryButtons.length > 0 && !singleActivationCostConfirmOption && !showDecisionOverlay && prompt?.presentation?.kind !== 'action_hub'">
            <div v-if="inlinePrimaryPromptMessage" class="prompt-inline-hint">
              {{ inlinePrimaryPromptMessage }}
            </div>
            <div v-if="responseAttackElementHintText" class="prompt-inline-hint prompt-inline-hint--attack-element">
              {{ responseAttackElementHintText }}
            </div>
            <div class="prompt-inline-grid" :class="inlinePrimaryGridClass">
            <div
              v-for="option in inlinePrimaryButtons"
              :key="option.id"
              class="prompt-inline-entry"
              :class="{ 'prompt-inline-entry--hinted-numeric': option.numeric && !!option.hint }"
            >
              <div
                v-if="option.hint && !shouldHideOptionHint(option)"
                class="prompt-inline-hint"
                :class="{ 'prompt-inline-hint--hinted-numeric': option.numeric }"
              >
                {{ option.hint }}
              </div>
              <button
                class="prompt-inline-btn"
                :data-testid="`prompt-option-${option.id}`"
                :class="[
                  isDockButtonImageStyle(option) ? 'action-image-btn' : '',
                  getDockButtonClass(option.id),
                  shouldEnlargeResponseActionButton(option) ? 'prompt-inline-btn--response-large' : '',
                  option.numeric ? 'prompt-inline-btn--numeric' : '',
                  option.numeric && !!option.hint ? 'prompt-inline-btn--hinted-numeric' : '',
                  isInlineCardOptionSelected(option.id) ? 'prompt-inline-btn--selected' : '',
                  option.disabled ? 'prompt-inline-btn--disabled' : ''
                ]"
                :disabled="!!option.disabled"
                :title="isDockButtonImageStyle(option) ? option.buttonLabel : undefined"
                :aria-label="isDockButtonImageStyle(option) ? option.buttonLabel : undefined"
                @click="handleOptionClick(option.id)"
              >
                <template v-if="isDockButtonImageStyle(option)">
                  <img
                    v-if="isDockButtonImageReady(option)"
                    class="action-image-btn-fill"
                    :src="dockButtonImageSrc(option)"
                    alt=""
                    @error="onDockButtonImageError(option)"
                  />
                  <span v-else class="action-image-fallback-text">{{ dockButtonFallbackText(option) }}</span>
                </template>
                <template v-else>
                  <template v-if="option.numeric && !!option.hint">
                    <span class="prompt-inline-btn-number">{{ option.buttonLabel }}</span>
                    <span class="prompt-inline-btn-caption">选择</span>
                  </template>
                  <template v-else>
                    {{ option.buttonLabel }}
                  </template>
                </template>
              </button>
            </div>
          </div>
          </div>

          <div v-if="singleActivationCostConfirmOption && !showDecisionOverlay" class="prompt-inline-entry">
            <div class="prompt-inline-hint">{{ singleActivationCostConfirmHintText }}</div>
            <div class="prompt-inline-actions-row">
              <button
                class="prompt-inline-btn prompt-inline-btn--success action-image-btn"
                :class="{ 'prompt-inline-btn--disabled': !!singleActivationCostConfirmOption.disabled }"
                :disabled="!!singleActivationCostConfirmOption.disabled"
                data-testid="prompt-confirm-btn"
                title="确认"
                aria-label="确认"
                @click="handleOptionClick(singleActivationCostConfirmOption.id)"
              >
                <img
                  v-if="isPromptConfirmImageReady()"
                  class="action-image-btn-fill"
                  :src="promptConfirmImageSrc()"
                  alt=""
                  @error="onPromptConfirmImageError"
                />
                <span v-else class="action-image-fallback-text">确</span>
              </button>
              <button
                class="prompt-inline-btn prompt-inline-btn--cancel action-image-btn"
                data-testid="prompt-cancel-btn"
                :title="isDockButtonImageStyle(cancelDockButton) ? cancelDockButton.buttonLabel : undefined"
                :aria-label="isDockButtonImageStyle(cancelDockButton) ? cancelDockButton.buttonLabel : undefined"
                @click="handleOptionClick(cancelDockButton.id)"
              >
                <img
                  v-if="isDockButtonImageReady(cancelDockButton)"
                  class="action-image-btn-fill"
                  :src="dockButtonImageSrc(cancelDockButton)"
                  alt=""
                  @error="onDockButtonImageError(cancelDockButton)"
                />
                <span v-else class="action-image-fallback-text">{{ dockButtonFallbackText(cancelDockButton) }}</span>
              </button>
            </div>
          </div>

          <div v-else-if="showCardConfirmCancelRow" class="prompt-inline-entry">
            <div class="prompt-inline-hint">{{ cardConfirmPromptMessage }}</div>
            <div class="prompt-inline-actions-row">
              <button
                class="prompt-inline-btn prompt-inline-btn--success action-image-btn"
                :class="{ 'prompt-inline-btn--disabled': !canConfirmPrompt }"
                :disabled="!canConfirmPrompt"
                data-testid="prompt-confirm-btn"
                title="发动"
                aria-label="发动"
                @click="confirmPromptAction"
              >
                <img
                  v-if="isPromptConfirmImageReady()"
                  class="action-image-btn-fill"
                  :src="promptConfirmImageSrc()"
                  alt=""
                  @error="onPromptConfirmImageError"
                />
                <span v-else class="action-image-fallback-text">确</span>
              </button>
              <button
                class="prompt-inline-btn prompt-inline-btn--cancel action-image-btn"
                data-testid="prompt-cancel-btn"
                :title="isDockButtonImageStyle(cancelDockButton) ? cancelDockButton.buttonLabel : undefined"
                :aria-label="isDockButtonImageStyle(cancelDockButton) ? cancelDockButton.buttonLabel : undefined"
                @click="handleOptionClick(cancelDockButton.id)"
              >
                <img
                  v-if="isDockButtonImageReady(cancelDockButton)"
                  class="action-image-btn-fill"
                  :src="dockButtonImageSrc(cancelDockButton)"
                  alt=""
                  @error="onDockButtonImageError(cancelDockButton)"
                />
                <span v-else class="action-image-fallback-text">{{ dockButtonFallbackText(cancelDockButton) }}</span>
              </button>
            </div>
          </div>

          <div v-else-if="promptNeedsCardConfirm" class="prompt-inline-entry">
            <div class="prompt-inline-hint">{{ cardConfirmHintText }}</div>
            <button
              class="prompt-inline-btn prompt-inline-btn--success action-image-btn"
              :class="{ 'prompt-inline-btn--disabled': !canConfirmPrompt }"
              :disabled="!canConfirmPrompt"
              data-testid="prompt-confirm-btn"
              title="发动"
              aria-label="发动"
              @click="confirmPromptAction"
            >
              <img
                v-if="isPromptConfirmImageReady()"
                class="action-image-btn-fill"
                :src="promptConfirmImageSrc()"
                alt=""
                @error="onPromptConfirmImageError"
              />
              <span v-else class="action-image-fallback-text">确</span>
            </button>
          </div>
        </template>

        <div
          v-if="canCancelPrompt && !isSkillChoicePrompt && !showCardConfirmCancelRow && !singleActivationCostConfirmOption && !showDecisionOverlay"
          class="prompt-inline-entry"
        >
          <div v-if="cancelDockButton.hint" class="prompt-inline-hint">{{ cancelDockButton.hint }}</div>
          <button
            class="prompt-inline-btn prompt-inline-btn--cancel"
            :class="isDockButtonImageStyle(cancelDockButton) ? 'action-image-btn' : ''"
            :title="isDockButtonImageStyle(cancelDockButton) ? cancelDockButton.buttonLabel : undefined"
            :aria-label="isDockButtonImageStyle(cancelDockButton) ? cancelDockButton.buttonLabel : undefined"
            @click="handleOptionClick(cancelDockButton.id)"
          >
            <template v-if="isDockButtonImageStyle(cancelDockButton)">
              <img
                v-if="isDockButtonImageReady(cancelDockButton)"
                class="action-image-btn-fill"
                :src="dockButtonImageSrc(cancelDockButton)"
                alt=""
                @error="onDockButtonImageError(cancelDockButton)"
              />
              <span v-else class="action-image-fallback-text">{{ dockButtonFallbackText(cancelDockButton) }}</span>
            </template>
            <template v-else>
              {{ cancelDockButton.buttonLabel }}
            </template>
          </button>
        </div>
      </div>
    </div>
  </Transition>

  <Teleport to="body">
    <Transition name="prompt-fraud-side-pop">
      <div
        v-if="isFraudElementCardPickerPrompt && fraudElementCardOptions.length > 0"
        class="prompt-fraud-global-layer"
        data-testid="decision-overlay"
      >
        <div class="prompt-fraud-global-panel">
          <div class="prompt-fraud-dialog prompt-fraud-dialog--global">
            <div class="prompt-fraud-title">{{ prompt?.message || '请选择本次攻击系别' }}</div>
            <div class="prompt-fraud-grid">
              <button
                v-for="option in fraudElementCardOptions"
                :key="option.id"
                class="prompt-fraud-card"
                :class="option.tone"
                :title="option.title"
                :aria-label="option.title"
                :data-testid="`prompt-option-${option.id}`"
                @click="handleOptionClick(option.id)"
              >
                <span class="prompt-fraud-card-title-banner">
                  <span class="prompt-fraud-card-title">{{ option.title }}</span>
                </span>
                <span class="prompt-fraud-card-medal">
                  <span>{{ option.glyph }}</span>
                </span>
                <span class="prompt-fraud-card-art">
                  <span class="prompt-fraud-card-glyph">{{ option.glyph }}</span>
                </span>
              </button>
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>

  <Teleport to="body">
    <Transition name="modal">
      <div
        v-if="isMultiSkillNameChoiceMode && skillBranchOptions.length > 0"
        class="overlay-panel-root overlay-panel-root--skill"
        data-testid="skill-branch-overlay"
      >
        <div class="overlay-panel" data-testid="decision-overlay" @click.stop>
          <div class="overlay-panel-header">
            <h2>{{ skillPromptTitle }}</h2>
          </div>
          <div class="overlay-panel-body">
            <button
              v-for="(entry, idx) in skillBranchOptions"
              :key="entry.id"
              class="overlay-panel-item"
              :data-testid="`branch-option-${idx}`"
              :disabled="entry.disabled"
              @click="handleOptionClick(entry.id)"
            >
              <div class="overlay-panel-item-title" :data-testid="`prompt-option-${entry.id}`">{{ entry.title }}</div>
              <div v-if="entry.description" class="overlay-panel-item-desc">{{ entry.description }}</div>
              <div v-if="entry.cost" class="overlay-panel-item-cost">{{ entry.cost }}</div>
            </button>
          </div>
          <div class="overlay-panel-footer">
            <button class="overlay-panel-cancel" data-testid="prompt-option-skip" @click="handleOptionClick('skip')">跳过</button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>

  <Teleport to="body">
    <Transition name="modal">
      <div v-if="showDecisionOverlay" class="overlay-panel-root overlay-panel-root--decision" data-testid="decision-overlay">
        <div class="overlay-panel" @click.stop>
          <div class="overlay-panel-header overlay-panel-header--decision">
            <h2>{{ decisionOverlayTitle }}</h2>
          </div>

          <div v-if="decisionOverlayMode === 'activation-cost'" class="overlay-panel-body overlay-panel-body--cost">
            <div class="overlay-cost-card">
              <span class="overlay-cost-text">{{ singleActivationCostConfirmHintText }}</span>
            </div>
            <button
              class="overlay-confirm-btn"
              :disabled="!!singleActivationCostConfirmOption?.disabled"
              @click="handleOptionClick(singleActivationCostConfirmOption!.id)"
            >
              确认发动
            </button>
          </div>

          <div v-else-if="decisionOverlayMode === 'numeric'" class="overlay-panel-body overlay-panel-body--numeric">
            <div class="overlay-numeric-grid">
              <button
                v-for="option in inlinePrimaryButtons"
                :key="option.id"
                class="overlay-numeric-tile"
                :data-testid="`numeric-option-${option.buttonLabel}`"
                :disabled="!!option.disabled"
                @click="handleOptionClick(option.id)"
              >
                <span class="overlay-numeric-value">{{ option.buttonLabel }}</span>
              </button>
            </div>
          </div>

          <div v-else-if="decisionOverlayMode === 'yes-no'" class="overlay-panel-body overlay-panel-body--yesno">
            <div class="overlay-yesno-row">
              <button
                v-for="option in prompt?.options || []"
                :key="option.id"
                class="overlay-yesno-btn"
                :class="option.id === '0' || option.id === 'yes' ? 'overlay-yesno-btn--yes' : 'overlay-yesno-btn--no'"
                :data-testid="`prompt-option-${option.id}`"
                :disabled="!!inlinePrimaryButtons.find(b => b.id === option.id)?.disabled"
                @click="handleOptionClick(option.id)"
              >
                {{ String(option.label || '').trim() }}
              </button>
            </div>
          </div>

          <div v-else class="overlay-panel-body overlay-panel-body--text">
            <button
              v-for="(option, idx) in inlinePrimaryButtons"
              :key="option.id"
              class="overlay-panel-item overlay-panel-item--text"
              :data-testid="`branch-option-${idx}`"
              :disabled="!!option.disabled"
              @click="handleOptionClick(option.id)"
            >
              <div class="overlay-panel-item-title" :data-testid="`prompt-option-${option.id}`">{{ overlayDecisionOptionTitle(option) }}</div>
              <div
                v-if="option.hint && option.hint !== overlayDecisionOptionTitle(option)"
                class="overlay-panel-item-desc"
              >{{ option.hint }}</div>
            </button>
          </div>

          <div v-if="canCancelPrompt && decisionOverlayMode !== 'yes-no'" class="overlay-panel-footer">
            <button class="overlay-panel-cancel" data-testid="prompt-cancel-btn" @click="handleOptionClick(cancelDockButton.id)">
              {{ cancelDockButton.buttonLabel || '取消' }}
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>

  <Teleport to="body">
    <Transition name="modal">
      <div v-if="isVisible && isSaintHealAllocatePrompt" class="overlay-panel-root overlay-panel-root--decision">
        <div class="overlay-panel" @click.stop>
          <div class="overlay-panel-header overlay-panel-header--decision">
            <h2>{{ prompt?.message || '请分配治疗' }}</h2>
          </div>
          <div class="overlay-panel-body overlay-saint-heal">
            <div class="overlay-saint-heal-summary">
              剩余可分配：{{ saintHealRemaining }} / {{ SAINT_HEAL_TOTAL }}
            </div>
            <div
              v-for="(option, index) in prompt?.options || []"
              :key="option.id"
              class="overlay-saint-heal-row"
            >
              <div class="overlay-saint-heal-row-label">{{ option.label }}</div>
              <div class="overlay-saint-heal-row-grid">
                <button
                  v-for="n in [0, 1, 2, 3]"
                  :key="n"
                  class="overlay-numeric-tile overlay-saint-heal-tile"
                  :class="{ 'overlay-saint-heal-tile--active': (saintHealAllocations[index] || 0) === n }"
                  :disabled="n > (saintHealAllocations[index] || 0) + saintHealRemaining"
                  @click="setSaintHealAllocation(index, n)"
                >
                  <span class="overlay-numeric-value">{{ n }}</span>
                </button>
              </div>
            </div>
            <button
              class="overlay-confirm-btn"
              :disabled="!canSubmitSaintHeal"
              @click="submitSaintHealAllocation"
            >
              确认分配
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>

  <Teleport to="body">
    <Transition name="modal">
      <div v-if="isVisible && isRuneReforgeAllocatePrompt" class="overlay-panel-root overlay-panel-root--decision">
        <div class="overlay-panel" @click.stop>
          <div class="overlay-panel-header overlay-panel-header--decision">
            <h2>{{ prompt?.message || '请分配战纹/魔纹' }}</h2>
          </div>
          <div class="overlay-panel-body overlay-saint-heal">
            <div class="overlay-saint-heal-summary">
              剩余可分配：{{ runeReforgeRemaining }} / {{ RUNE_REFORGE_TOTAL }}
            </div>
            <div
              v-for="(option, index) in prompt?.options || []"
              :key="option.id"
              class="overlay-saint-heal-row"
            >
              <div class="overlay-saint-heal-row-label">{{ option.label }}</div>
              <div class="overlay-saint-heal-row-grid">
                <button
                  v-for="n in [0, 1, 2, 3]"
                  :key="n"
                  class="overlay-numeric-tile overlay-saint-heal-tile"
                  :class="{ 'overlay-saint-heal-tile--active': (runeReforgeAllocations[index] || 0) === n }"
                  :disabled="n > (runeReforgeAllocations[index] || 0) + runeReforgeRemaining"
                  @click="setRuneReforgeAllocation(index, n)"
                >
                  <span class="overlay-numeric-value">{{ n }}</span>
                </button>
              </div>
            </div>
            <button
              class="overlay-confirm-btn"
              :disabled="!canSubmitRuneReforge"
              @click="submitRuneReforgeAllocation"
            >
              确认分配
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.prompt-inline-root {
  width: 100%;
  display: flex;
  justify-content: center;
  pointer-events: auto;
}

.prompt-inline-surface {
  width: min(760px, 100%);
  border-radius: 14px;
  border: 1px solid rgba(146, 183, 207, 0.36);
  background:
    linear-gradient(180deg, rgba(12, 24, 40, 0.94), rgba(7, 16, 27, 0.96));
  box-shadow:
    0 16px 28px rgba(2, 8, 18, 0.44),
    inset 0 1px 0 rgba(236, 246, 254, 0.1);
  padding: 10px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.prompt-inline-grid {
  display: grid;
  gap: 8px;
}

.prompt-skill-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 2px;
}

.prompt-skill-row {
  display: flex;
  flex-direction: column;
  align-items: stretch;
  gap: 8px;
  padding: 2px;
}

.prompt-skill-row + .prompt-skill-row {
  border-top: 1px dashed rgba(138, 171, 192, 0.28);
  padding-top: 9px;
}

.prompt-skill-actions {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 8px;
  align-items: center;
}

.prompt-skill-text {
  font-size: 13px;
  line-height: 1.3;
  color: rgba(221, 237, 248, 0.94);
  letter-spacing: 0.01em;
  text-align: center;
  white-space: normal;
  word-break: break-word;
}

.prompt-skill-action {
  justify-self: center;
}

.prompt-skill-action--plain {
  justify-self: stretch;
  width: 100%;
  min-height: 42px;
}

.prompt-skill-action:hover:not(:disabled) {
  filter: brightness(1.08);
}

.prompt-fraud-global-layer {
  position: fixed;
  inset: 0;
  z-index: 60;
  pointer-events: auto;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(7, 14, 24, 0.38);
  backdrop-filter: blur(7px) saturate(0.98);
  -webkit-backdrop-filter: blur(7px) saturate(0.98);
  padding:
    max(16px, calc(var(--safe-top, 0px) + 8px))
    max(16px, calc(var(--safe-right, 0px) + 8px))
    max(16px, calc(var(--safe-bottom, 0px) + 8px))
    max(16px, calc(var(--safe-left, 0px) + 8px));
}

.prompt-fraud-global-panel {
  pointer-events: auto;
  width: min(860px, calc(100vw - 40px));
  max-height: calc(100vh - 40px);
  overflow: auto;
  border-radius: 14px;
  border: 1px solid rgba(146, 183, 207, 0.42);
  background:
    linear-gradient(180deg, rgba(9, 20, 34, 0.96), rgba(6, 14, 24, 0.97));
  box-shadow:
    0 18px 34px rgba(2, 8, 18, 0.52),
    inset 0 1px 0 rgba(236, 246, 254, 0.12);
  padding: 10px;
}

.prompt-fraud-dialog {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 4px 2px 2px;
}

.prompt-fraud-dialog--global {
  padding: 4px;
}

.prompt-fraud-title {
  font-size: 13px;
  line-height: 1.4;
  color: rgba(225, 238, 249, 0.96);
  text-align: center;
  letter-spacing: 0.01em;
}

.prompt-fraud-grid {
  display: grid;
  grid-template-columns: repeat(5, minmax(0, 1fr));
  gap: 8px;
}

.prompt-fraud-card {
  --fraud-edge-color: rgba(185, 152, 102, 0.78);
  --fraud-edge-glow: rgba(232, 191, 121, 0.38);
  --fraud-base-top: #2f2520;
  --fraud-base-bottom: #17120f;
  --fraud-ribbon-start: #8c5a2f;
  --fraud-ribbon-end: #60401f;
  --fraud-medal-bg: radial-gradient(circle at 32% 28%, #f1d79b, #c6924f 58%, #784d1d);
  --fraud-medal-fg: #fff7ea;
  --fraud-art-top: rgba(214, 174, 116, 0.3);
  --fraud-art-bottom: rgba(71, 45, 24, 0.82);
  display: flex;
  flex-direction: column;
  align-items: stretch;
  justify-content: flex-start;
  position: relative;
  min-height: 136px;
  border-radius: 10px;
  border: 2px solid var(--fraud-edge-color);
  background: linear-gradient(180deg, var(--fraud-base-top), var(--fraud-base-bottom));
  color: rgba(245, 250, 255, 0.97);
  box-shadow:
    0 8px 16px rgba(0, 0, 0, 0.55),
    0 0 12px var(--fraud-edge-glow),
    inset 0 0 0 1px rgba(255, 244, 214, 0.24),
    inset 0 0 10px rgba(255, 255, 255, 0.08);
  transition: transform 0.16s ease, filter 0.16s ease, border-color 0.16s ease, box-shadow 0.16s ease;
  overflow: hidden;
  text-align: center;
}

.prompt-fraud-card:hover:not(:disabled) {
  transform: translateY(-2px);
  border-color: rgba(255, 228, 170, 0.96);
  box-shadow:
    0 12px 22px rgba(0, 0, 0, 0.62),
    0 0 16px rgba(255, 214, 138, 0.44),
    inset 0 0 0 1px rgba(255, 247, 229, 0.38);
}

.prompt-fraud-card-title-banner {
  position: absolute;
  top: 6px;
  left: 16px;
  right: 16px;
  height: 20px;
  border-radius: 4px;
  border: 1px solid rgba(202, 184, 148, 0.88);
  background: linear-gradient(180deg, rgba(251, 249, 243, 0.96), rgba(222, 214, 197, 0.94));
  display: inline-flex;
  align-items: center;
  justify-content: center;
  z-index: 4;
}

.prompt-fraud-card-title {
  color: rgba(46, 34, 22, 0.94);
  font-size: 11px;
  font-weight: 820;
  letter-spacing: 0.04em;
  line-height: 1;
}

.prompt-fraud-card-medal {
  position: absolute;
  top: 2px;
  left: 3px;
  width: 28px;
  height: 28px;
  border-radius: 999px;
  border: 1px solid rgba(255, 255, 255, 0.7);
  background: radial-gradient(circle at 30% 30%, rgba(255, 255, 255, 0.95), rgba(201, 191, 176, 0.8) 42%, rgba(61, 53, 47, 0.88));
  display: inline-flex;
  align-items: center;
  justify-content: center;
  z-index: 8;
  box-shadow:
    0 2px 6px rgba(0, 0, 0, 0.55),
    0 0 8px rgba(255, 244, 189, 0.44);
}

.prompt-fraud-card-medal > span {
  width: 72%;
  height: 72%;
  border-radius: 999px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: var(--fraud-medal-bg);
  color: var(--fraud-medal-fg);
  font-size: 13px;
  font-weight: 900;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.7);
}

.prompt-fraud-card-art {
  margin: 30px 7px 8px;
  height: 88px;
  border-radius: 4px;
  border: 1px solid rgba(175, 161, 132, 0.8);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background:
    radial-gradient(100% 120% at 50% 0%, var(--fraud-art-top), transparent 56%),
    linear-gradient(180deg, rgba(255, 250, 235, 0.12), var(--fraud-art-bottom));
  box-shadow:
    inset 0 0 0 1px rgba(255, 236, 196, 0.3),
    inset 0 0 16px rgba(0, 0, 0, 0.38);
}

.prompt-fraud-card-glyph {
  font-size: 38px;
  font-weight: 900;
  line-height: 1;
  color: rgba(247, 253, 255, 0.96);
  text-shadow:
    0 0 10px rgba(255, 255, 255, 0.24),
    0 2px 5px rgba(0, 0, 0, 0.6);
}

.prompt-fraud-card--water {
  --fraud-edge-color: rgba(102, 152, 196, 0.78);
  --fraud-edge-glow: rgba(124, 196, 255, 0.38);
  --fraud-base-top: #1a2a3e;
  --fraud-base-bottom: #0f1826;
  --fraud-ribbon-start: #25689f;
  --fraud-ribbon-end: #1b446c;
  --fraud-medal-bg: radial-gradient(circle at 32% 28%, #ccf3ff, #4ea1d6 58%, #195580);
  --fraud-medal-fg: #effbff;
  --fraud-art-top: rgba(138, 206, 255, 0.4);
  --fraud-art-bottom: rgba(18, 48, 78, 0.84);
}

.prompt-fraud-card--fire {
  --fraud-edge-color: rgba(205, 123, 82, 0.78);
  --fraud-edge-glow: rgba(255, 140, 98, 0.4);
  --fraud-base-top: #3c1f18;
  --fraud-base-bottom: #1c120f;
  --fraud-ribbon-start: #c6352f;
  --fraud-ribbon-end: #8e1b17;
  --fraud-medal-bg: radial-gradient(circle at 32% 28%, #ffca7f, #f36d33 58%, #9b2e1a);
  --fraud-medal-fg: #fff8eb;
  --fraud-art-top: rgba(255, 176, 126, 0.42);
  --fraud-art-bottom: rgba(88, 30, 20, 0.84);
}

.prompt-fraud-card--earth {
  --fraud-edge-color: rgba(174, 138, 93, 0.8);
  --fraud-edge-glow: rgba(225, 186, 113, 0.34);
  --fraud-base-top: #32261a;
  --fraud-base-bottom: #1a1410;
  --fraud-ribbon-start: #8a5d2f;
  --fraud-ribbon-end: #60401f;
  --fraud-medal-bg: radial-gradient(circle at 32% 28%, #f1d79b, #c6924f 58%, #784d1d);
  --fraud-medal-fg: #fff6e4;
  --fraud-art-top: rgba(236, 199, 128, 0.36);
  --fraud-art-bottom: rgba(85, 57, 22, 0.84);
}

.prompt-fraud-card--wind {
  --fraud-edge-color: rgba(96, 169, 145, 0.78);
  --fraud-edge-glow: rgba(116, 223, 181, 0.34);
  --fraud-base-top: #183329;
  --fraud-base-bottom: #101f1b;
  --fraud-ribbon-start: #237258;
  --fraud-ribbon-end: #194e3d;
  --fraud-medal-bg: radial-gradient(circle at 32% 28%, #c8f9e6, #55b68d 58%, #216f54);
  --fraud-medal-fg: #edfff6;
  --fraud-art-top: rgba(145, 241, 205, 0.36);
  --fraud-art-bottom: rgba(25, 73, 56, 0.84);
}

.prompt-fraud-card--thunder {
  --fraud-edge-color: rgba(140, 124, 200, 0.8);
  --fraud-edge-glow: rgba(183, 148, 255, 0.36);
  --fraud-base-top: #24213d;
  --fraud-base-bottom: #171427;
  --fraud-ribbon-start: #5f4a99;
  --fraud-ribbon-end: #40306f;
  --fraud-medal-bg: radial-gradient(circle at 32% 28%, #efe2ff, #9c79dc 58%, #4e3385);
  --fraud-medal-fg: #faf3ff;
  --fraud-art-top: rgba(208, 180, 255, 0.38);
  --fraud-art-bottom: rgba(54, 36, 89, 0.84);
}

.prompt-fraud-card--generic {
  --fraud-edge-color: rgba(170, 190, 216, 0.56);
  --fraud-edge-glow: rgba(166, 193, 227, 0.34);
  --fraud-base-top: #223044;
  --fraud-base-bottom: #121c29;
  --fraud-ribbon-start: #44648a;
  --fraud-ribbon-end: #2d4159;
  --fraud-medal-bg: radial-gradient(circle at 32% 28%, #dfe9f7, #8ba5cc 58%, #4d6283);
  --fraud-medal-fg: #f2f7ff;
  --fraud-art-top: rgba(170, 200, 235, 0.34);
  --fraud-art-bottom: rgba(35, 51, 72, 0.84);
}

.prompt-inline-entry {
  display: flex;
  flex-direction: column;
  gap: 4px;
  align-items: center;
}

.prompt-inline-actions-row {
  width: 100%;
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 8px;
  justify-items: center;
  align-items: center;
}

.prompt-inline-entry--hinted-numeric {
  padding: 6px;
  border-radius: 12px;
  border: 1px solid rgba(114, 150, 177, 0.3);
  background:
    linear-gradient(180deg, rgba(26, 41, 58, 0.78), rgba(17, 30, 45, 0.82));
  box-shadow:
    inset 0 1px 0 rgba(208, 226, 241, 0.08),
    0 6px 14px rgba(3, 9, 18, 0.26);
}

.prompt-inline-btn.action-image-btn {
  width: 72px;
  height: 72px;
  min-height: 0;
  max-width: 72px;
  aspect-ratio: 1 / 1;
  border-radius: 12px !important;
  align-self: center;
  justify-self: center;
  flex-shrink: 0;
}

.prompt-inline-btn.action-image-btn.prompt-inline-btn--response-large {
  width: 96px;
  height: 96px;
  max-width: 96px;
  border-radius: 14px !important;
}

.prompt-inline-hint {
  min-height: 18px;
  padding: 0 4px;
  text-align: center;
  color: rgba(199, 219, 237, 0.88);
  font-size: 11px;
  line-height: 1.35;
  letter-spacing: 0.01em;
}

.prompt-inline-hint--hinted-numeric {
  min-height: 0;
  padding: 2px 6px 0;
  font-size: 11.5px;
  font-weight: 600;
  color: rgba(214, 231, 246, 0.96);
}

.prompt-inline-hint--attack-element {
  min-height: 0;
  margin: 0 0 6px;
  padding: 4px 8px;
  border-radius: 9px;
  border: 1px solid rgba(132, 165, 187, 0.36);
  background: linear-gradient(180deg, rgba(28, 43, 61, 0.72), rgba(17, 30, 45, 0.75));
  color: rgba(223, 239, 250, 0.96);
  font-size: 12px;
  font-weight: 640;
  letter-spacing: 0.01em;
}

.prompt-inline-grid--1 {
  grid-template-columns: 1fr;
}

.prompt-inline-grid--2 {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.prompt-inline-grid--3 {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.prompt-inline-grid--4 {
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.prompt-inline-btn {
  min-height: 40px;
  width: 100%;
  max-width: 100%;
  border-radius: 10px;
  border: 1px solid rgba(137, 167, 186, 0.42);
  background: rgba(32, 48, 67, 0.68);
  color: #e3eef8;
  font-size: 13px;
  font-weight: 700;
  letter-spacing: 0.01em;
  transition: transform 0.16s ease, border-color 0.16s ease, filter 0.16s ease, background 0.16s ease;
}

.action-image-btn {
  -webkit-appearance: none !important;
  appearance: none !important;
  border: none !important;
  background: transparent !important;
  box-shadow: none !important;
  padding: 0 !important;
  overflow: hidden;
  position: relative;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.action-image-btn:focus,
.action-image-btn:focus-visible {
  outline: none !important;
  box-shadow: none !important;
}

.action-image-btn-fill {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
  pointer-events: none;
  user-select: none;
}

.prompt-inline-btn.action-image-btn .action-image-btn-fill {
  transform-origin: center;
}

.action-image-fallback-text {
  position: relative;
  z-index: 1;
  font-size: 14px;
  font-weight: 700;
  color: #f2f8ff;
  text-shadow: 0 1px 3px rgba(0, 0, 0, 0.45);
}

.prompt-inline-btn:hover:not(:disabled) {
  transform: translateY(-1px);
  border-color: rgba(180, 210, 227, 0.66);
  filter: brightness(1.05);
}

.prompt-inline-btn--normal {
  background: linear-gradient(180deg, rgba(47, 67, 88, 0.86), rgba(35, 53, 72, 0.88));
}

.prompt-inline-btn--numeric {
  font-size: 15px;
  font-weight: 800;
}

.prompt-inline-btn--hinted-numeric {
  min-height: 42px;
  border-radius: 10px;
  border-color: rgba(175, 197, 223, 0.52);
  background:
    linear-gradient(180deg, rgba(74, 101, 128, 0.92), rgba(54, 77, 99, 0.94));
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  box-shadow:
    inset 0 1px 0 rgba(228, 239, 252, 0.26),
    0 6px 12px rgba(5, 12, 21, 0.28);
}

.prompt-inline-btn-number {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 1.8em;
  min-height: 1.8em;
  padding: 0 0.45em;
  border-radius: 999px;
  font-size: 14px;
  font-weight: 900;
  line-height: 1;
  color: #f6fbff;
  background: rgba(16, 29, 45, 0.46);
  border: 1px solid rgba(197, 220, 245, 0.38);
}

.prompt-inline-btn-caption {
  font-size: 12px;
  font-weight: 650;
  letter-spacing: 0.03em;
  color: rgba(234, 244, 255, 0.94);
}

.prompt-inline-btn--hinted-numeric:hover:not(:disabled) {
  filter: brightness(1.06);
  border-color: rgba(197, 220, 246, 0.75);
}

.prompt-inline-btn--success {
  border-color: rgba(111, 185, 141, 0.52);
  background: linear-gradient(180deg, rgba(30, 109, 74, 0.9), rgba(22, 78, 55, 0.9));
}

.prompt-inline-btn--take,
.prompt-inline-btn--counter,
.prompt-inline-btn--defend,
.prompt-inline-btn--cancel {
  color: #f6ecda;
  text-shadow: 0 1px 2px rgba(8, 8, 12, 0.65);
  background-repeat: no-repeat;
  background-size: cover;
  background-position: center;
  box-shadow:
    inset 0 1px 0 rgba(255, 238, 206, 0.22),
    0 8px 18px rgba(6, 7, 14, 0.32);
}

.prompt-inline-btn--take {
  border-color: rgba(205, 171, 113, 0.68);
  background-image: url('/assets/ui/prompt_btn_take.png');
}

.prompt-inline-btn--counter {
  border-color: rgba(157, 141, 228, 0.56);
  background-image: url('/assets/ui/prompt_btn_counter.png');
}

.prompt-inline-btn--defend {
  border-color: rgba(111, 170, 225, 0.6);
  background-image: url('/assets/ui/prompt_btn_defend.png');
}

.prompt-inline-btn--cancel {
  border-color: rgba(196, 152, 102, 0.56);
  background-image: url('/assets/ui/action_cancel_btn.png');
}

.prompt-inline-btn--extract {
  border-color: rgba(183, 154, 105, 0.56);
  background: linear-gradient(180deg, rgba(91, 69, 38, 0.9), rgba(68, 50, 28, 0.92));
}

.prompt-inline-btn--selected {
  box-shadow:
    0 0 0 2px rgba(241, 211, 150, 0.74),
    0 0 18px rgba(189, 152, 90, 0.38);
}

.prompt-inline-btn--disabled {
  opacity: 0.45;
  cursor: not-allowed;
  transform: none !important;
  filter: none !important;
}

.prompt-inline-pop-enter-active,
.prompt-inline-pop-leave-active {
  transition: opacity 0.22s ease, transform 0.22s ease;
}

.prompt-inline-pop-enter-from,
.prompt-inline-pop-leave-to {
  opacity: 0;
  transform: translateY(10px) scale(0.98);
}

.prompt-fraud-side-pop-enter-active,
.prompt-fraud-side-pop-leave-active {
  transition: opacity 0.2s ease, transform 0.2s ease;
}

.prompt-fraud-side-pop-enter-from,
.prompt-fraud-side-pop-leave-to {
  opacity: 0;
  transform: translateX(18px) scale(0.98);
}

@media (max-width: 900px) {
  .prompt-inline-surface {
    width: min(92vw, 680px);
    padding: 8px;
    gap: 7px;
  }

  .prompt-inline-grid--3,
  .prompt-inline-grid--4 {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .prompt-inline-btn {
    min-height: 38px;
    font-size: 12px;
  }

  .prompt-inline-btn.action-image-btn {
    width: 72px;
    height: 72px;
    min-height: 0;
    max-width: 72px;
    justify-self: center;
  }

  .prompt-inline-btn.action-image-btn.prompt-inline-btn--response-large {
    width: 88px;
    height: 88px;
    max-width: 88px;
  }

  .prompt-inline-hint {
    min-height: 16px;
    font-size: 10px;
  }

  .prompt-inline-hint--hinted-numeric {
    font-size: 10.5px;
  }

  .prompt-inline-btn--hinted-numeric {
    min-height: 38px;
    gap: 6px;
  }

  .prompt-inline-btn-number {
    font-size: 12px;
  }

  .prompt-inline-btn-caption {
    font-size: 11px;
  }

  .prompt-skill-text {
    font-size: 12px;
  }

  .prompt-fraud-global-panel {
    width: min(760px, calc(100vw - 28px));
    max-height: calc(100vh - 28px);
  }

  .prompt-fraud-grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }

  .prompt-fraud-card {
    min-height: 126px;
  }
}

@media (max-width: 560px) {
  .prompt-inline-surface {
    width: 100%;
    padding: 7px;
    border-radius: 12px;
  }

  .prompt-inline-grid--2,
  .prompt-inline-grid--3,
  .prompt-inline-grid--4 {
    grid-template-columns: 1fr;
  }

  .prompt-skill-row {
    flex-direction: column;
    align-items: stretch;
    gap: 6px;
  }

  .prompt-skill-text {
    text-align: center;
  }

  .prompt-fraud-global-layer {
    justify-content: center;
    align-items: center;
    padding:
      max(10px, calc(var(--safe-top, 0px) + 4px))
      max(8px, calc(var(--safe-right, 0px) + 4px))
      max(10px, calc(var(--safe-bottom, 0px) + 4px))
      max(8px, calc(var(--safe-left, 0px) + 4px));
  }

  .prompt-fraud-global-panel {
    width: min(100%, 620px);
    border-radius: 12px;
    padding: 8px;
  }

  .prompt-inline-btn.action-image-btn {
    width: 72px;
    height: 72px;
    min-height: 0;
    max-width: 72px;
    justify-self: center;
  }

  .prompt-inline-btn.action-image-btn.prompt-inline-btn--response-large {
    width: 84px;
    height: 84px;
    max-width: 84px;
  }

  .prompt-fraud-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .prompt-fraud-card {
    min-height: 120px;
  }

  .prompt-fraud-card-glyph {
    font-size: 32px;
  }
}

/* ── Overlay Panel (shared: skill-branch & decision) ── */
.overlay-panel-root {
  position: fixed;
  inset: 0;
  z-index: 13050;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1rem;
  backdrop-filter: blur(3px);
}
.overlay-panel-root--skill {
  background:
    radial-gradient(420px 220px at 50% 44%, rgba(136, 188, 195, 0.18), transparent 70%),
    rgba(0, 0, 0, 0.72);
}
.overlay-panel-root--decision {
  background:
    radial-gradient(380px 200px at 50% 46%, rgba(180, 140, 60, 0.12), transparent 70%),
    rgba(0, 0, 0, 0.74);
}

.overlay-panel {
  position: relative;
  width: min(480px, calc(100vw - 2rem));
  max-height: 85vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  border-radius: 14px;
  border: 1px solid rgba(132, 167, 186, 0.36);
  box-shadow: 0 18px 34px rgba(2, 8, 18, 0.52),
              inset 0 1px 0 rgba(236, 246, 254, 0.12);
  background: linear-gradient(180deg, rgba(8, 20, 34, 0.92), rgba(6, 15, 28, 0.95)),
              url('/assets/ui/modal-aura.svg') center/cover no-repeat;
}
.overlay-panel-root--decision .overlay-panel {
  border-color: rgba(180, 150, 90, 0.32);
  box-shadow: 0 18px 34px rgba(2, 8, 18, 0.52),
              inset 0 1px 0 rgba(255, 240, 200, 0.1);
  background: linear-gradient(180deg, rgba(10, 18, 30, 0.94), rgba(6, 12, 22, 0.96));
}

.overlay-panel-header {
  flex-shrink: 0;
  padding: 20px 24px 16px;
  background: linear-gradient(110deg, rgba(34, 74, 97, 0.88), rgba(94, 72, 43, 0.88));
  border-bottom: 1px solid rgba(149, 186, 204, 0.26);
  text-align: center;
}
.overlay-panel-header h2 {
  font-size: 1.05rem;
  font-weight: 600;
  color: #ffe2ad;
  margin: 0;
  line-height: 1.5;
}
.overlay-panel-header--decision {
  padding: 18px 24px 14px;
  background: linear-gradient(110deg, rgba(40, 56, 80, 0.9), rgba(80, 60, 30, 0.85));
  border-bottom-color: rgba(180, 150, 90, 0.24);
}
.overlay-panel-header--decision h2 {
  font-size: 1rem;
  color: #ffd98a;
}

.overlay-panel-body {
  flex: 1;
  overflow-y: auto;
  padding: 16px 20px;
  display: flex;
  flex-direction: column;
  gap: 10px;
  background: rgba(6, 17, 29, 0.42);
}
.overlay-panel-root--decision .overlay-panel-body {
  background: rgba(6, 14, 24, 0.4);
}

.overlay-panel-item {
  position: relative;
  width: 100%;
  text-align: left;
  padding: 14px 18px;
  border-radius: 10px;
  border: 1px solid rgba(118, 152, 173, 0.34);
  background: rgba(14, 32, 48, 0.56);
  box-shadow: inset 0 1px 0 rgba(237, 247, 254, 0.06);
  cursor: pointer;
  transition: all 0.18s ease;
  font-family: inherit;
  font-size: inherit;
  line-height: inherit;
}
.overlay-panel-item:hover:not(:disabled) {
  border-color: rgba(211, 188, 142, 0.6);
  background: rgba(20, 42, 60, 0.72);
  box-shadow: 0 0 12px rgba(211, 188, 142, 0.15),
              inset 0 1px 0 rgba(237, 247, 254, 0.1);
}
.overlay-panel-item:active:not(:disabled) {
  transform: scale(0.98);
}
.overlay-panel-item:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}
.overlay-panel-item--text {
  border-color: rgba(150, 130, 80, 0.3);
  background: rgba(14, 28, 44, 0.56);
  box-shadow: inset 0 1px 0 rgba(255, 240, 200, 0.05);
}
.overlay-panel-item--text:hover:not(:disabled) {
  border-color: rgba(255, 210, 120, 0.55);
  background: rgba(22, 38, 56, 0.72);
  box-shadow: 0 0 12px rgba(255, 210, 120, 0.12),
              inset 0 1px 0 rgba(255, 240, 200, 0.08);
}

.overlay-panel-item-title {
  font-size: 0.95rem;
  font-weight: 600;
  color: #ffe2ad;
  margin-bottom: 4px;
}
.overlay-panel-root--decision .overlay-panel-item-title {
  color: #ffd98a;
}

.overlay-panel-item-desc {
  font-size: 0.8rem;
  line-height: 1.5;
  color: rgba(199, 219, 237, 0.78);
  white-space: pre-wrap;
  word-break: break-word;
}

.overlay-panel-item-cost {
  margin-top: 6px;
  font-size: 0.75rem;
  color: rgba(156, 166, 184, 0.8);
}

.overlay-panel-footer {
  flex-shrink: 0;
  padding: 12px 20px 16px;
  text-align: center;
  background: rgba(6, 16, 28, 0.66);
  border-top: 1px solid rgba(118, 153, 173, 0.24);
}
.overlay-panel-root--decision .overlay-panel-footer {
  background: rgba(6, 14, 24, 0.66);
  border-top-color: rgba(150, 130, 80, 0.2);
}

.overlay-panel-cancel {
  padding: 8px 32px;
  border-radius: 8px;
  border: 1px solid rgba(156, 166, 184, 0.3);
  background: rgba(59, 67, 84, 0.6);
  color: rgba(199, 219, 237, 0.7);
  font-size: 0.85rem;
  cursor: pointer;
  transition: all 0.15s ease;
  font-family: inherit;
}
.overlay-panel-cancel:hover {
  background: rgba(59, 67, 84, 0.9);
  color: rgba(199, 219, 237, 0.95);
  border-color: rgba(156, 166, 184, 0.5);
}

/* ── Decision-specific body variants ── */
.overlay-panel-body--numeric {
  padding: 20px 16px;
}
.overlay-numeric-grid {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 10px;
}
.overlay-numeric-tile {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  width: 56px;
  min-height: 56px;
  border-radius: 12px;
  border: 1px solid rgba(180, 150, 90, 0.36);
  background: rgba(20, 32, 48, 0.7);
  cursor: pointer;
  transition: all 0.18s ease;
  padding: 8px 4px;
  font-family: inherit;
}
.overlay-numeric-tile:hover:not(:disabled) {
  border-color: rgba(255, 210, 120, 0.7);
  background: rgba(30, 44, 60, 0.85);
  box-shadow: 0 0 14px rgba(255, 210, 120, 0.18);
}
.overlay-numeric-tile:active:not(:disabled) {
  transform: scale(0.95);
}
.overlay-numeric-tile:disabled {
  opacity: 0.35;
  cursor: not-allowed;
}
.overlay-numeric-value {
  font-size: 1.4rem;
  font-weight: 700;
  color: #ffd98a;
  line-height: 1;
}

.overlay-panel-body--text {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.overlay-panel-body--yesno {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 24px 20px;
}

.overlay-yesno-row {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
  width: 100%;
  max-width: 320px;
}

.overlay-yesno-btn {
  padding: 14px 24px;
  border-radius: 10px;
  border: 1px solid rgba(150, 130, 80, 0.36);
  background: rgba(14, 28, 44, 0.56);
  color: #ffd98a;
  font-size: 1.1rem;
  font-weight: 700;
  cursor: pointer;
  transition: all 0.18s ease;
  font-family: inherit;
  text-align: center;
}

.overlay-yesno-btn:hover:not(:disabled) {
  border-color: rgba(255, 210, 120, 0.7);
  background: rgba(22, 38, 56, 0.72);
  box-shadow: 0 0 12px rgba(255, 210, 120, 0.12);
}

.overlay-yesno-btn:active:not(:disabled) {
  transform: scale(0.96);
}

.overlay-yesno-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.overlay-yesno-btn--yes {
  border-color: rgba(180, 150, 90, 0.5);
  background: linear-gradient(180deg, rgba(100, 75, 30, 0.7), rgba(70, 52, 20, 0.8));
  color: #ffe2ad;
}

.overlay-yesno-btn--yes:hover:not(:disabled) {
  border-color: rgba(255, 210, 120, 0.7);
  background: linear-gradient(180deg, rgba(120, 90, 35, 0.85), rgba(85, 62, 25, 0.9));
  box-shadow: 0 0 12px rgba(255, 210, 120, 0.15);
}

.overlay-yesno-btn--no {
  border-color: rgba(180, 130, 90, 0.4);
  background: linear-gradient(180deg, rgba(80, 55, 25, 0.6), rgba(55, 38, 18, 0.7));
  color: #ffd98a;
}

.overlay-yesno-btn--no:hover:not(:disabled) {
  border-color: rgba(255, 200, 120, 0.6);
  background: linear-gradient(180deg, rgba(95, 65, 30, 0.75), rgba(65, 45, 22, 0.85));
  box-shadow: 0 0 12px rgba(255, 200, 120, 0.12);
}

.overlay-panel-body--cost {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
  padding: 24px 20px;
}
.overlay-cost-card {
  padding: 16px 24px;
  border-radius: 10px;
  border: 1px solid rgba(180, 150, 90, 0.3);
  background: rgba(14, 28, 44, 0.6);
  text-align: center;
}
.overlay-cost-text {
  font-size: 0.9rem;
  color: rgba(225, 238, 249, 0.92);
  line-height: 1.5;
}
.overlay-confirm-btn {
  padding: 10px 36px;
  border-radius: 10px;
  border: 1px solid rgba(180, 150, 90, 0.5);
  background: linear-gradient(180deg, rgba(120, 90, 30, 0.7), rgba(80, 60, 20, 0.8));
  color: #ffe2ad;
  font-size: 0.9rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.15s ease;
  font-family: inherit;
}
.overlay-confirm-btn:hover:not(:disabled) {
  background: linear-gradient(180deg, rgba(140, 105, 35, 0.85), rgba(100, 75, 25, 0.9));
  border-color: rgba(255, 210, 120, 0.7);
}
.overlay-confirm-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.overlay-saint-heal {
  display: flex;
  flex-direction: column;
  gap: 14px;
  padding: 18px 24px 20px;
}
.overlay-saint-heal-summary {
  text-align: center;
  font-size: 0.9rem;
  color: rgba(255, 217, 138, 0.92);
}
.overlay-saint-heal-row {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 10px 12px;
  border-radius: 10px;
  background: rgba(10, 22, 38, 0.55);
  border: 1px solid rgba(118, 153, 173, 0.22);
}
.overlay-saint-heal-row-label {
  font-size: 0.95rem;
  font-weight: 600;
  color: #ffd98a;
}
.overlay-saint-heal-row-grid {
  display: flex;
  gap: 8px;
}
.overlay-saint-heal-tile {
  flex: 1;
}
.overlay-saint-heal-tile--active {
  border-color: rgba(255, 210, 120, 0.85);
  background: linear-gradient(180deg, rgba(140, 100, 35, 0.7), rgba(90, 65, 20, 0.85));
  color: #ffe9b8;
}
</style>

<style>
/* Overlay panel transitions (unscoped for Vue Transition compatibility) */
.overlay-panel-root.modal-enter-active,
.overlay-panel-root.modal-leave-active {
  transition: opacity 0.24s ease;
}
.overlay-panel-root.modal-enter-from,
.overlay-panel-root.modal-leave-to {
  opacity: 0;
}
.overlay-panel-root.modal-enter-active .overlay-panel,
.overlay-panel-root.modal-leave-active .overlay-panel {
  transition: transform 0.24s ease;
}
.overlay-panel-root.modal-enter-from .overlay-panel,
.overlay-panel-root.modal-leave-to .overlay-panel {
  transform: scale(0.95) translateY(8px);
}
</style>
