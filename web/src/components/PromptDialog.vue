<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useInterruptStore } from '../stores/interrupt.store'
import { useInteractionController } from '../composables/useInteractionController'
import { useSnapshotStore } from '../stores/snapshot.store'
import {
  promptImageButtonKindByOption,
  type PromptImageButtonKind,
} from '../constants/promptButtonRules'
import AllocationOverlayRenderer from './prompt/renderers/AllocationOverlayRenderer.vue'
import CardPickerPromptRenderer from './prompt/renderers/CardPickerPromptRenderer.vue'
import DecisionOverlayRenderer from './prompt/renderers/DecisionOverlayRenderer.vue'
import DirectionPromptRenderer from './prompt/renderers/DirectionPromptRenderer.vue'
import ExtractPromptRenderer from './prompt/renderers/ExtractPromptRenderer.vue'
import FraudElementRenderer from './prompt/renderers/FraudElementRenderer.vue'
import ResponsePromptRenderer from './prompt/renderers/ResponsePromptRenderer.vue'
import SkillChoicePromptRenderer from './prompt/renderers/SkillChoicePromptRenderer.vue'
import TargetPickerPromptRenderer from './prompt/renderers/TargetPickerPromptRenderer.vue'
import { promptRendererUsesInlineSurface, selectPromptRenderer } from './prompt/rendererRegistry'
import type { PromptOption } from '../types/game'

const interruptStore = useInterruptStore()
const snapshotStore = useSnapshotStore()
const interaction = useInteractionController()

const {
  prompt,
  myPlayerId,
  playerViews,
  selectedExtractIndices,
  selectedInlineCardIDs,
  playerOptionEntries,
  selectedPromptTargetOptionIndexes,
  nonPlayerOptions,
  canCancelPrompt,
  showPromptError,
  cancelPrompt,
  submitConfirm,
  submitOptionIndex,
  submitOptionIndexes,
  submitTargetSelection,
  submitSelectedCardIDs,
  submitRespondTake,
  submitRespondCounter,
  submitRespondDefend,
} = interaction
const myHand = computed(() => playerViews.value[myPlayerId.value]?.hand || [])
const myPromptSelectableCards = computed(() => {
  const player = playerViews.value[myPlayerId.value]
  const handCards = (player?.hand || []).map((card, index) => ({ card, index }))
  const blessingCards = (player?.field || [])
    .filter((fieldCard) => fieldCard?.mode === 'Cover' && fieldCard.effect === 'ElfBlessing' && !!fieldCard.card)
    .map((fieldCard, offset) => ({
      card: fieldCard.card!,
      index: handCards.length + offset,
    }))
  return [...handCards, ...blessingCards]
})

// 行动选择（攻击/法术/购买/提取/合成）不在这里显示，由 ActionPanel 承载
const isActionSelectionPrompt = computed(() => {
  if (!prompt.value) return false
  return prompt.value.presentation?.kind === 'action_hub'
})

const isVisible = computed(() =>
  prompt.value !== null && prompt.value.player_id === myPlayerId.value && !isActionSelectionPrompt.value
)

const autoResolvedPromptKey = ref('')

watch(() => prompt.value, () => {
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

const promptEffectHints = computed(() =>
  Array.isArray(prompt.value?.effect_hints)
    ? prompt.value.effect_hints.map((hint) => String(hint || '').trim()).filter((hint) => hint.length > 0)
    : []
)

const promptInteraction = computed(() => prompt.value?.interaction ?? null)

function promptSelectedCountForContract(): number {
  const contract = promptInteraction.value
  if (!contract || contract.submit_action !== 'select') return 0
  if (contract.selection_source === 'field' && contract.selection_value === 'option_index') {
    return interruptStore.selectedFieldOptionIndexes.length
  }
  if (contract.selection_source === 'hand' && contract.selection_value === 'card_id') {
    return isNonHandChooseCardsMultiMode.value
      ? selectedInlineCardIDs.value.length
      : interruptStore.selectedHandIndexes.length
  }
  if (contract.selection_source === 'target') {
    return interruptStore.selectedTargets.length
  }
  return 0
}

function canConfirmManualSelectContract(): boolean | null {
  const contract = promptInteraction.value
  if (!prompt.value || !contract || contract.submit_action !== 'select' || contract.confirm_mode !== 'manual') return null
  if (contract.selection_source === 'field' && contract.selection_value === 'option_index') {
    const cCount = promptSelectedCountForContract()
    return cCount >= prompt.value.min && cCount <= prompt.value.max
  }
  if (contract.selection_source === 'hand' && contract.selection_value === 'card_id') {
    if (isPlagueDeathTouchElementPrompt.value) {
      return resolvePlagueDeathTouchElementOptionIndex() !== null
    }
    const cCount = promptSelectedCountForContract()
    return cCount >= prompt.value.min && cCount <= prompt.value.max
  }
  if (contract.selection_source === 'target') {
    const tCount = promptSelectedCountForContract()
    if (contract.selection_value === 'option_index') {
      return (
        tCount >= prompt.value.min &&
        tCount <= prompt.value.max &&
        selectedPromptTargetOptionIndexes.value.length === tCount
      )
    }
    return tCount >= prompt.value.min && tCount <= prompt.value.max
  }
  if (contract.selection_source === 'option' && contract.selection_value === 'option_index') {
    const cCount = selectedExtractIndices.value.length
    return cCount >= prompt.value.min && cCount <= prompt.value.max
  }
  return null
}

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
  return hasCounterOrDefend.value && !!ids && ids.length > 0
})

const isConfirmType = computed(() => {
  if (!prompt.value) return false
  const kind = prompt.value.presentation?.kind
  return !!kind && kind !== 'card_picker' && kind !== 'target_picker' && kind !== 'action_hub'
})

const isExtractPrompt = computed(() => prompt.value?.presentation?.layout === 'extract')

// 圣疗 3 点治疗分配：每个目标独立 0..3 数字选择，当前前端允许总和不超过 3。
const isSaintHealAllocatePrompt = computed(() => prompt.value?.presentation?.layout === 'heal_allocate')
const saintHealAllocations = ref<number[]>([])
const SAINT_HEAL_TOTAL = 3

// 符文改造分配：战纹/魔纹 0..3 数字选择，要求总和=3。
const isRuneReforgeAllocatePrompt = computed(() => prompt.value?.presentation?.layout === 'rune_allocate')
const runeReforgeAllocations = ref<number[]>([])
const RUNE_REFORGE_TOTAL = 3

// 血腥祷言分配：两名队友逐行分配，要求总和等于本次 X。
const isBloodPrayerAllocatePrompt = computed(() => prompt.value?.presentation?.layout === 'blood_prayer_allocate')
const bloodPrayerAllocations = ref<number[]>([])

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
    if (isBloodPrayerAllocatePrompt.value && prompt.value) {
      bloodPrayerAllocations.value = prompt.value.options.map(() => 0)
    } else {
      bloodPrayerAllocations.value = []
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

const bloodPrayerTotal = computed(() => {
  if (!isBloodPrayerAllocatePrompt.value || !prompt.value) return 0
  const message = String(prompt.value.message || '')
  const match = message.match(/等于\s*(\d+)/)
  if (match?.[1]) return Number(match[1]) || 0
  return prompt.value.options?.length || 0
})

const bloodPrayerRemaining = computed(() => {
  const used = bloodPrayerAllocations.value.reduce((s, v) => s + (v || 0), 0)
  return bloodPrayerTotal.value - used
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

function setBloodPrayerAllocation(index: number, value: number) {
  if (!isBloodPrayerAllocatePrompt.value) return
  const current = bloodPrayerAllocations.value[index] || 0
  const otherSum = bloodPrayerAllocations.value.reduce((s, v, i) => s + (i === index ? 0 : (v || 0)), 0)
  const maxAllowed = Math.max(0, bloodPrayerTotal.value - otherSum)
  const clamped = Math.max(0, Math.min(value, Math.min(bloodPrayerTotal.value, maxAllowed)))
  if (clamped === current) return
  const next = bloodPrayerAllocations.value.slice()
  next[index] = clamped
  bloodPrayerAllocations.value = next
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

const canSubmitBloodPrayer = computed(() => {
  if (!isBloodPrayerAllocatePrompt.value) return false
  if (bloodPrayerAllocations.value.length !== 2) return false
  return bloodPrayerAllocations.value.reduce((s, v) => s + (v || 0), 0) === bloodPrayerTotal.value
})

function submitSaintHealAllocation() {
  if (!canSubmitSaintHeal.value) {
    showPromptError(`治疗分配无效（总和不能超过 ${SAINT_HEAL_TOTAL}）`)
    return
  }
  submitOptionIndexes([...saintHealAllocations.value])
}

function submitRuneReforgeAllocation() {
  if (!canSubmitRuneReforge.value) {
    showPromptError(`分配无效（战纹+魔纹之和必须等于 ${RUNE_REFORGE_TOTAL}）`)
    return
  }
  submitOptionIndexes([...runeReforgeAllocations.value])
}

function submitBloodPrayerAllocation() {
  if (!canSubmitBloodPrayer.value) {
    showPromptError(`分配无效（治疗点数之和必须等于 ${bloodPrayerTotal.value}）`)
    return
  }
  submitOptionIndexes([...bloodPrayerAllocations.value])
}









function toggleExtractOption(index: number) {
  interaction.toggleExtractOption(index, prompt.value?.max ?? 2)
}

function confirmExtractSelection() {
  const min = prompt.value?.min ?? 1
  const max = prompt.value?.max ?? 2
  const sel = selectedExtractIndices.value
  if (sel.length < min || sel.length > max) return
  interaction.submitOptionIndexes(sel)
}

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

function handleOptionClick(optionId: string) {
  if (optionId === 'counter_disabled') {
    showPromptError('此攻击无法应战')
    return
  }
  if (prompt.value?.presentation?.kind === 'skill_choice') {
    const idx = prompt.value.options.findIndex((o: { id: string }) => o.id === optionId)
    if (idx >= 0) {
      submitOptionIndex(idx)
    } else {
      cancelPrompt()
    }
    return
  }
  const structuredOptionKind = prompt.value?.presentation?.kind
  if (structuredOptionKind === 'branch_select' || structuredOptionKind === 'numeric') {
    const optionIndex = prompt.value?.options?.findIndex((o: { id: string }) => o.id === optionId) ?? -1
    if (optionIndex >= 0) {
      if (structuredOptionKind !== 'branch_select' && prompt.value?.presentation?.has_decline && optionIndex === (prompt.value.presentation.decline_index ?? 0)) {
        cancelPrompt()
        return
      }
      submitOptionIndex(optionIndex)
      return
    }
    if (optionId === 'skip' || optionId === 'cancel') {
      cancelPrompt()
      return
    }
  }
  if (optionId === 'skip' || optionId === 'cancel') {
    cancelPrompt()
    return
  }
  if (optionId === 'refuse') {
    // 魔爆冲击“不弃牌”是规则内显式选项，直接走 Cancel 语义。
    cancelPrompt()
    return
  }
  if (optionId === 'confirm') {
    submitConfirm()
    return
  }
  const optionIndex = prompt.value?.options?.findIndex((o: { id: string }) => o.id === optionId) ?? -1
  if (prompt.value?.presentation?.kind !== 'branch_select' && prompt.value?.presentation?.has_decline && optionIndex === (prompt.value.presentation.decline_index ?? 0)) {
    cancelPrompt()
    return
  }
  // 魔弹融合等确认选项：yes=0, no=1
  if (optionId === 'yes' || optionId === 'no') {
    submitOptionIndex(optionId === 'yes' ? 0 : 1)
    return
  }
  // 魔弹掌控方向选择：normal=0, reverse=1
  if (optionId === 'normal' || optionId === 'reverse') {
    submitOptionIndex(optionId === 'normal' ? 0 : 1)
    return
  }
  const responseKind = promptOptionResponseKind({ id: optionId })
  if (responseKind === 'take') {
    submitRespondTake()
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
    if (!submitRespondCounter(isMagicMissilePrompt.value)) return
    return
  }
  if (responseKind === 'defend') {
    if (interruptStore.selectedHandIndexes.length === 0) {
      showPromptError('请先选择一张【圣光】进行防御（圣盾需提前放置）')
      return
    }
    if (!submitRespondDefend()) return
    return
  }
  if (isNonHandChooseCardsMultiMode.value && isNonHandChooseCardOption(optionId)) {
    toggleInlineCardOption(optionId)
    return
  }
  {
    const optionIndex = prompt.value?.options?.findIndex((o: { id: string }) => o.id === optionId) ?? -1
    if (optionIndex >= 0) {
      submitOptionIndex(optionIndex)
    } else {
      const index = parseInt(optionId, 10)
      if (!Number.isNaN(index)) {
        submitOptionIndex(index)
      } else {
        showPromptError('当前选项缺少可提交的 option index，请刷新后重试')
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
  const contractResult = canConfirmManualSelectContract()
  if (contractResult !== null) return contractResult
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
      : isFieldCoverSelectionPrompt()
        ? interruptStore.selectedFieldOptionIndexes.length
      : interruptStore.selectedHandIndexes.length
    return cCount >= prompt.value.min && cCount <= prompt.value.max
  }
  return true
})

function confirmPromptAction() {
  if (!canConfirmPrompt.value) return
  const interactionContract = promptInteraction.value

  if (interactionContract?.submit_action === 'select' && interactionContract.confirm_mode === 'manual') {
    if (interactionContract.selection_source === 'field' && interactionContract.selection_value === 'option_index') {
      const indexes = interruptStore.selectedFieldOptionIndexes
      if (indexes.length > 0) {
        submitOptionIndexes(indexes)
      }
      return
    }
    if (interactionContract.selection_source === 'hand' && interactionContract.selection_value === 'card_id') {
      if (isPlagueDeathTouchElementPrompt.value) {
        const optionIndex = resolvePlagueDeathTouchElementOptionIndex()
        if (optionIndex === null) {
          showPromptError('请先在手牌区选择可用于死亡之触的同系牌')
          return
        }
        submitOptionIndex(optionIndex)
        return
      }
      if (isNonHandChooseCardsMultiMode.value) {
        submitSelectedCardIDs(selectedInlineCardIDs.value)
        return
      }
      const cardIDs = selectedPromptHandCardIDs(interruptStore.selectedHandIndexes)
      if (cardIDs.length === interruptStore.selectedHandIndexes.length) {
        submitSelectedCardIDs(cardIDs)
        return
      }
      showPromptError('当前卡牌选择缺少 card_id，请刷新后重试')
      return
    }
    if (interactionContract.selection_source === 'target') {
      if (interactionContract.selection_value === 'option_index') {
        const targetOptionIndexes = selectedPromptTargetOptionIndexes.value
        if (targetOptionIndexes.length !== interruptStore.selectedTargets.length) {
          showPromptError('请选择有效目标')
          return
        }
        submitOptionIndexes(targetOptionIndexes)
        return
      }
      submitTargetSelection()
      return
    }
  }

  if (isFieldCoverSelectionPrompt()) {
    const indexes = interruptStore.selectedFieldOptionIndexes
    if (indexes.length > 0) {
      submitOptionIndexes(indexes)
    }
    return
  }

  if (isPlagueDeathTouchElementPrompt.value) {
    const optionIndex = resolvePlagueDeathTouchElementOptionIndex()
    if (optionIndex === null) {
      showPromptError('请先在手牌区选择可用于死亡之触的同系牌')
      return
    }
    submitOptionIndex(optionIndex)
    return
  }

  if (promptRequiresManualTargetConfirm.value) {
    const targetOptionIndexes = selectedPromptTargetOptionIndexes.value
    if (targetOptionIndexes.length !== interruptStore.selectedTargets.length) {
      showPromptError('请选择有效的治疗目标')
      return
    }
    submitOptionIndexes(targetOptionIndexes)
    return
  }

  if (prompt.value?.presentation?.kind === 'target_picker' && interruptStore.selectedTargets.length > 0) {
    if (interruptStore.selectedTargets.length === 1) {
      const targetId = interruptStore.selectedTargets[0]
      if (!targetId) return
      submitTargetSelection()
    } else {
      submitTargetSelection()
    }
    return
  }

  if (isNonHandChooseCardsMultiMode.value) {
    if (selectedInlineCardIDs.value.length > 0 || prompt.value?.min === 0) {
      submitSelectedCardIDs(selectedInlineCardIDs.value)
    }
    return
  }

  const indices = interruptStore.selectedHandIndexes
  if (prompt.value?.presentation?.kind === 'card_picker' && indices.length === 0 && prompt.value.min === 0) {
    submitSelectedCardIDs([])
    return
  }
  if (indices.length > 0) {
    const cardIDs = selectedPromptHandCardIDs(indices)
    if (cardIDs.length === indices.length) {
      submitSelectedCardIDs(cardIDs)
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
  const entry = myPromptSelectableCards.value.find(({ card }) => String(card.id || '').trim() === cardID)
  return entry ? entry.index : null
}

function promptOptionForHandIndex(handIndex: number): PromptOption | null {
  if (!prompt.value?.options?.length) return null
  const selectedEntry = myPromptSelectableCards.value.find((entry) => entry.index === handIndex)
  const selectedCardID = String(selectedEntry?.card?.id || '').trim()
  if (!selectedCardID) return null
  for (const option of prompt.value.options) {
    const cardID = optionCardID(option)
    if (cardID && cardID === selectedCardID) return option
  }
  return null
}

function selectedPromptHandCardIDs(indices: number[]): string[] {
  if (!prompt.value || prompt.value.presentation?.kind !== 'card_picker') return []
  const cardSource = prompt.value.presentation?.card_source
  if (cardSource !== 'hand' && cardSource !== 'proxy') return []
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
  // card_picker with card_source=hand/proxy: selectable cards are matched by card_id in hand-like playable zones.
  if (p.kind === 'card_picker' && (p.card_source === 'hand' || p.card_source === 'proxy')) {
    return handIndexForPromptOption(option) !== null
  }
  if (p.kind === 'card_picker') return false
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

type SkillChoiceRendererButton = {
  id: string
  label: string
  disabled: boolean
  cancel: boolean
  imageSrc: string
  imageReady: boolean
  fallbackText: string
}

type FraudElementCardOption = {
  id: string
  title: string
  glyph: string
  tone: string
}

type ResponsePromptOption = {
  id: string
  buttonLabel: string
  disabled: boolean
  kind: Exclude<ResponseActionKind, null>
  imageSrc: string
  imageReady: boolean
  fallbackText: string
  enlarged: boolean
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

const isDirectionPrompt = computed(() => {
  const p = prompt.value
  if (!p || !p.options?.length) return false
  const choiceType = String(p.choice_type || '').trim()
  if (choiceType === 'mg_magic_bullet_direction' || choiceType === 'hb_radiant_cannon_side' || choiceType === 'ss_convert_color') return true
  if (choiceType) return false
  if (p.presentation?.kind !== 'branch_select') return false
  const optionIds = new Set(p.options.map((option) => String(option.id || '').trim().toLowerCase()))
  return optionIds.has('normal') && optionIds.has('reverse')
})

type DirectionPromptOption = {
  id: string
  label: string
  hint?: string
  description?: string
  disabled?: boolean
  tone?: string
  icon?: string
}

const directionPromptOptions = computed<DirectionPromptOption[]>(() => {
  if (!isDirectionPrompt.value || !prompt.value?.options?.length) return []
  return prompt.value.options.map((option, index) => {
    const label = String(option.label || option.button_label || '').trim()
    const hint = String(option.hint || '').trim()
    const disabled = !!(option as { disabled?: boolean }).disabled
    return {
      id: option.id,
      label: label || `选项 ${index + 1}`,
      hint: hint || undefined,
      description: hint || undefined,
      disabled,
      tone: option.id === 'reverse' ? 'prompt-direction--reverse' : 'prompt-direction--normal',
      icon: option.id === 'reverse' ? 'arrow-left' : 'arrow-right',
    }
  })
})

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
  if (isFieldCoverSelectionPrompt()) return true
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

const promptRendererKey = computed(() =>
  selectPromptRenderer({
    visible: isVisible.value,
    isActionHub: isActionSelectionPrompt.value,
    isExtractPrompt: isExtractPrompt.value,
    extractOptionCount: extractPromptOptions.value.length,
    isSkillChoicePrompt: isSkillChoicePrompt.value,
    isMultiSkillNameChoiceMode: isMultiSkillNameChoiceMode.value,
    skillPromptButtonCount: skillPromptButtons.value.length,
    skillBranchCount: skillBranchOptions.value.length,
    showTargetSelectionHintRow: showTargetSelectionHintRow.value,
    hasCounterOrDefend: hasCounterOrDefend.value,
    responseOptionCount: responsePromptOptions.value.length,
    inlinePrimaryButtonCount: inlinePrimaryButtons.value.length,
    promptNeedsCardConfirm: promptNeedsCardConfirm.value,
    canCancelPrompt: canCancelPrompt.value,
    showDecisionOverlay: showDecisionOverlay.value,
    isDirectionPrompt: isDirectionPrompt.value,
    directionOptionCount: directionPromptOptions.value.length,
    isFraudElementCardPickerPrompt: isFraudElementCardPickerPrompt.value,
    fraudElementOptionCount: fraudElementCardOptions.value.length,
    isSaintHealAllocatePrompt: isSaintHealAllocatePrompt.value,
    isRuneReforgeAllocatePrompt: isRuneReforgeAllocatePrompt.value,
    isBloodPrayerAllocatePrompt: isBloodPrayerAllocatePrompt.value,
  })
)

const useInlineSurface = computed(() => promptRendererUsesInlineSurface(promptRendererKey.value))

const cardConfirmHintText = computed(() => {
  if (isElfElementalShotPickPrompt.value) return '请从手牌区或扩展区选择法术牌/祝福牌并点击发动'
  if (isPlagueDeathTouchElementPrompt.value) return '请选择同系手牌并点击确认'
  if (isFieldCoverSelectionPrompt()) return '请选择扩展区盖牌并点击确认'
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

const cardPickerPromptMessage = computed(() =>
  showCardConfirmCancelRow.value ? cardConfirmPromptMessage.value : cardConfirmHintText.value
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
  (!!needsTargetSelection.value || !!needsCounterTargetSelection.value)
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
  if (isBloodPrayerAllocatePrompt.value) return []
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
    const shouldFilterDeclineOption =
      prompt.value?.presentation?.kind !== 'branch_select' &&
      prompt.value?.presentation?.has_decline
    const declineIndex = shouldFilterDeclineOption ? (prompt.value?.presentation?.decline_index ?? 0) : -1
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
  const rawLabel = String(option.label || option.buttonLabel || '').trim()
  let title = rawLabel || `技能 ${index + 1}`

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

const skillChoiceRendererButtons = computed<SkillChoiceRendererButton[]>(() =>
  skillPromptButtons.value.map((button) => ({
    id: button.id,
    label: button.label,
    disabled: button.disabled,
    cancel: button.cancel,
    imageSrc: skillButtonImageSrc(button),
    imageReady: isSkillButtonImageReady(button),
    fallbackText: skillButtonFallbackText(button),
  }))
)

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
  const actionableOptions = inlinePrimaryButtons.value.filter((opt) => opt.id !== 'skip' && opt.id !== 'cancel')
  const includeCancelLikeOption = actionableOptions.length === 1
  return inlinePrimaryButtons.value
    .filter((opt) => includeCancelLikeOption || (opt.id !== 'skip' && opt.id !== 'cancel'))
    .map((option, index) => {
      const rawLabel = String(option.label || '').trim()
      const isCancelLike = option.id === 'skip' || option.id === 'cancel'
      const title = isCancelLike ? '' : parseSkillTitle(option, index)
      const costMatch = rawLabel.match(/\[[^\]]+\]/)
      return {
        id: option.id,
        title,
        description: option.hint || (isCancelLike ? String(option.label || option.buttonLabel || '跳过').trim() : undefined),
        cost: costMatch ? costMatch[0] : undefined,
        disabled: !!option.disabled,
      }
    })
})

function onSkillChoiceRendererImageError(optionId: string) {
  const targetButton = skillPromptButtons.value.find((button) => button.id === optionId)
  if (!targetButton) return
  onSkillButtonImageError(targetButton)
}

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
  if (isBloodPrayerAllocatePrompt.value) return false  // → 血腥祷言分配专属弹窗
  if (needsCardSelection.value) return false           // → 卡牌选择流程（命中/防御/应战按钮留内联）
  // 简单确认弹框（是/否两个按钮）→ 也走 overlay 决策弹框，便于统一展示
  // 注意：已删除旧的 return false 逻辑，让 是/否 弹框也进入 overlay
  // 符合通用弹窗条件
  if (singleActivationCostConfirmOption.value) return true
  if (isDirectionPrompt.value) return false
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

const decisionOverlayCanCancel = computed(() =>
  canCancelPrompt.value && prompt.value?.presentation?.kind !== 'branch_select'
)

const decisionOverlayTitle = computed(() => {
  if (singleActivationCostConfirmOption.value) {
    return String(prompt.value?.message || '确认发动').trim()
  }
  return inlinePrimaryPromptMessage.value || '请选择'
})

const decisionOverlayOptions = computed(() => {
  if (decisionOverlayMode.value === 'yes-no') {
    return (prompt.value?.options || []).map((option) => ({
      id: option.id,
      label: option.label,
      buttonLabel: option.button_label,
      hint: option.hint,
      disabled: !!inlinePrimaryButtons.value.find((btn) => btn.id === option.id)?.disabled,
    }))
  }
  return inlinePrimaryButtons.value.map((option) => ({
    id: option.id,
    label: option.label,
    buttonLabel: option.buttonLabel,
    hint: option.hint,
    disabled: !!option.disabled,
  }))
})

const hasAnyInlineButton = computed(() => useInlineSurface.value)

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

const responsePromptOptions = computed<ResponsePromptOption[]>(() => {
  if (!hasCounterOrDefend.value) return []
  return inlinePrimaryButtons.value
    .map((option) => {
      const kind = promptOptionResponseKind({ id: option.id })
      if (!kind) return null
      return {
        id: option.id,
        buttonLabel: option.buttonLabel,
        disabled: !!option.disabled,
        kind,
        imageSrc: dockButtonImageSrc(option),
        imageReady: isDockButtonImageReady(option),
        fallbackText: dockButtonFallbackText(option),
        enlarged: shouldEnlargeResponseActionButton(option),
      }
    })
    .filter((option): option is ResponsePromptOption => option !== null)
})

function onResponsePromptImageError(optionId: string) {
  const option = inlinePrimaryButtons.value.find((candidate) => candidate.id === optionId)
  if (!option) return
  onDockButtonImageError(option)
}

type ExtractPromptOption = {
  id: string
  label: string
}

const extractPromptOptions = computed<ExtractPromptOption[]>(() => {
  if (!isExtractPrompt.value || !prompt.value?.options?.length) return []
  return prompt.value.options.map((option) => ({
    id: option.id,
    label: option.label,
  }))
})

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
          <ExtractPromptRenderer
            :visible="isExtractPrompt && extractPromptOptions.length > 0"
            :options="extractPromptOptions"
            :selected-indexes="selectedExtractIndices"
            :min="prompt?.min ?? 1"
            :max="prompt?.max ?? 2"
            :confirm-image-src="promptConfirmImageSrc()"
            :confirm-image-ready="isPromptConfirmImageReady()"
            confirm-fallback-text="确"
            @toggle="toggleExtractOption"
            @confirm="confirmExtractSelection"
            @confirm-image-error="onPromptConfirmImageError"
          />
        </template>

        <template v-else>
          <SkillChoicePromptRenderer
            v-if="isSkillChoicePrompt && skillPromptButtons.length > 0 && !isMultiSkillNameChoiceMode"
            :inline-visible="true"
            :overlay-visible="false"
            :title="skillPromptTitle"
            :buttons="skillChoiceRendererButtons"
            :branches="[]"
            @select="handleOptionClick"
            @image-error="onSkillChoiceRendererImageError"
          />

          <TargetPickerPromptRenderer
            v-else-if="promptRendererKey === 'target_picker'"
            :visible="showTargetSelectionHintRow"
            :message="targetSelectionPromptMessage"
            :show-confirm="promptRequiresManualTargetConfirm"
            :can-confirm="canConfirmPrompt"
            :confirm-image-src="promptConfirmImageSrc()"
            :confirm-image-ready="isPromptConfirmImageReady()"
            confirm-fallback-text="确"
            @confirm="confirmPromptAction"
            @confirm-image-error="onPromptConfirmImageError"
          />

          <ResponsePromptRenderer
            v-else-if="promptRendererKey === 'response'"
            :visible="hasCounterOrDefend"
            :hint="responseAttackElementHintText"
            :options="responsePromptOptions"
            @select="handleOptionClick"
            @image-error="onResponsePromptImageError"
          />

          <div v-else-if="promptRendererKey === 'inline_buttons'">
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

          <CardPickerPromptRenderer
            v-else-if="promptRendererKey === 'card_picker'"
            :visible="promptNeedsCardConfirm"
            :message="cardPickerPromptMessage"
            :effect-hints="promptEffectHints"
            :can-confirm="canConfirmPrompt"
            :show-cancel="showCardConfirmCancelRow"
            :confirm-image-src="promptConfirmImageSrc()"
            :confirm-image-ready="isPromptConfirmImageReady()"
            confirm-fallback-text="确"
            :cancel-image-src="dockButtonImageSrc(cancelDockButton)"
            :cancel-image-ready="isDockButtonImageReady(cancelDockButton)"
            :cancel-fallback-text="dockButtonFallbackText(cancelDockButton)"
            :cancel-title="isDockButtonImageStyle(cancelDockButton) ? cancelDockButton.buttonLabel : undefined"
            :cancel-aria-label="isDockButtonImageStyle(cancelDockButton) ? cancelDockButton.buttonLabel : undefined"
            @confirm="confirmPromptAction"
            @cancel="handleOptionClick(cancelDockButton.id)"
            @confirm-image-error="onPromptConfirmImageError"
            @cancel-image-error="onDockButtonImageError(cancelDockButton)"
          />
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

  <DirectionPromptRenderer
    :visible="isVisible && isDirectionPrompt && directionPromptOptions.length > 0"
    :title="prompt?.message || '请选择方向'"
    :options="directionPromptOptions"
    @select="handleOptionClick"
  />

  <FraudElementRenderer
    :visible="isFraudElementCardPickerPrompt && fraudElementCardOptions.length > 0"
    :title="prompt?.message || '请选择本次攻击系别'"
    :options="fraudElementCardOptions"
    @select="handleOptionClick"
  />

  <SkillChoicePromptRenderer
    :inline-visible="false"
    :overlay-visible="isMultiSkillNameChoiceMode && skillBranchOptions.length > 0"
    :title="skillPromptTitle"
    :buttons="[]"
    :branches="skillBranchOptions"
    @select="handleOptionClick"
    @image-error="onSkillChoiceRendererImageError"
  />

  <DecisionOverlayRenderer
    :visible="showDecisionOverlay"
    :title="decisionOverlayTitle"
    :mode="decisionOverlayMode"
    :options="decisionOverlayOptions"
    :activation-hint="singleActivationCostConfirmHintText"
    :activation-option-id="singleActivationCostConfirmOption?.id || ''"
    :activation-disabled="!!singleActivationCostConfirmOption?.disabled"
    :can-cancel="decisionOverlayCanCancel"
    :cancel-label="cancelDockButton.buttonLabel || '取消'"
    :cancel-option-id="cancelDockButton.id"
    @select="handleOptionClick"
    @cancel="handleOptionClick"
  />

  <AllocationOverlayRenderer
    :visible="isVisible && isSaintHealAllocatePrompt"
    :title="prompt?.message || '请分配治疗'"
    :rows="prompt?.options || []"
    :values="saintHealAllocations"
    :remaining="saintHealRemaining"
    :total="SAINT_HEAL_TOTAL"
    :can-submit="canSubmitSaintHeal"
    submit-label="确认分配"
    @change="setSaintHealAllocation"
    @submit="submitSaintHealAllocation"
  />

  <AllocationOverlayRenderer
    :visible="isVisible && isRuneReforgeAllocatePrompt"
    :title="prompt?.message || '请分配战纹/魔纹'"
    :rows="prompt?.options || []"
    :values="runeReforgeAllocations"
    :remaining="runeReforgeRemaining"
    :total="RUNE_REFORGE_TOTAL"
    :can-submit="canSubmitRuneReforge"
    submit-label="确认分配"
    @change="setRuneReforgeAllocation"
    @submit="submitRuneReforgeAllocation"
  />

  <AllocationOverlayRenderer
    :visible="isVisible && isBloodPrayerAllocatePrompt"
    :title="prompt?.message || '请分配治疗'"
    :rows="prompt?.options || []"
    :values="bloodPrayerAllocations"
    :remaining="bloodPrayerRemaining"
    :total="bloodPrayerTotal"
    :can-submit="canSubmitBloodPrayer"
    submit-label="确认分配"
    total-label="剩余待分配"
    @change="setBloodPrayerAllocation"
    @submit="submitBloodPrayerAllocation"
  />
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
